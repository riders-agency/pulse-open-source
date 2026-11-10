package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

type inventoryPattern struct {
	source string
	re     *regexp.Regexp
}

func main() {
	inventoryPath := flag.String("inventory", "../../contracts/openapi/endpoints.inventory.tsv", "endpoint inventory TSV")
	webRoot := flag.String("web-root", "../web/src", "frontend source root")
	flag.Parse()

	patterns, err := readPatterns(*inventoryPath)
	if err != nil {
		fail("read inventory: %v", err)
	}
	paths, err := frontendPaths(*webRoot)
	if err != nil {
		fail("scan frontend: %v", err)
	}

	var missing []string
	for path := range paths {
		if !matchesAny(path, patterns) {
			missing = append(missing, path)
		}
	}
	sort.Strings(missing)
	for _, path := range missing {
		fmt.Printf("VIOLATION frontend_api_path_missing_from_inventory path=%q\n", path)
	}
	if len(missing) > 0 {
		os.Exit(1)
	}
	fmt.Printf("Frontend API path inventory check passed: paths=%d\n", len(paths))
}

func readPatterns(path string) ([]inventoryPattern, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	seen := map[string]struct{}{}
	var patterns []inventoryPattern
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid inventory row %q", line)
		}
		path := strings.TrimSpace(parts[1])
		if _, ok := seen[path]; ok {
			continue
		}
		seen[path] = struct{}{}
		patterns = append(patterns, inventoryPattern{source: path, re: inventoryPathRegexp(path)})
	}
	return patterns, scanner.Err()
}

func inventoryPathRegexp(path string) *regexp.Regexp {
	var out strings.Builder
	out.WriteString("^")
	for i := 0; i < len(path); {
		if path[i] == '{' {
			end := strings.IndexByte(path[i:], '}')
			if end >= 0 {
				out.WriteString(`[^/]+`)
				i += end + 1
				continue
			}
		}
		out.WriteString(regexp.QuoteMeta(string(path[i])))
		i++
	}
	out.WriteString("$")
	return regexp.MustCompile(out.String())
}

func frontendPaths(root string) (map[string]struct{}, error) {
	paths := map[string]struct{}{}
	err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			name := entry.Name()
			if name == "node_modules" || name == ".next" || name == "test-results" || name == "playwright-report" || name == "generated" || name == "__generated__" {
				return filepath.SkipDir
			}
			return nil
		}
		if !isFrontendSource(path) {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, candidate := range extractAPIPaths(string(raw)) {
			paths[candidate] = struct{}{}
		}
		return nil
	})
	return paths, err
}

func isFrontendSource(path string) bool {
	switch filepath.Ext(path) {
	case ".ts", ".tsx", ".js", ".jsx":
		return true
	default:
		return false
	}
}

var templateExpr = regexp.MustCompile(`\$\{[^}]+\}`)

func extractAPIPaths(raw string) []string {
	var out []string
	raw = strings.NewReplacer(`\n`, "\n", `\r`, "\r", `\t`, "\t").Replace(raw)
	for _, token := range strings.FieldsFunc(raw, func(r rune) bool {
		return r == '"' || r == '\'' || r == '`' || r == '<' || r == '>' || r == ' ' || r == '\n' || r == '\r' || r == '\t'
	}) {
		if !strings.Contains(token, "/api/v1") {
			continue
		}
		if strings.HasPrefix(token, "http://") || strings.HasPrefix(token, "https://") {
			if !strings.Contains(token, "pulse.riders.agency/api/v1") {
				continue
			}
		}
		idx := strings.Index(token, "/api/v1")
		if idx < 0 {
			continue
		}
		path := token[idx:]
		if open := strings.Index(path, "${"); open >= 0 && !strings.Contains(path[open:], "}") {
			path = path[:open]
		}
		path = strings.Split(path, "?")[0]
		path = strings.Split(path, "#")[0]
		path = templateExpr.ReplaceAllString(path, "x")
		if open := strings.Index(path, "${"); open >= 0 {
			path = path[:open]
		}
		path = strings.TrimRight(path, ".,);]")
		path = strings.TrimRight(path, "/")
		if strings.HasPrefix(path, "/api/v1") && len(path) > len("/api/v1") {
			out = append(out, path)
		}
	}
	return out
}

func matchesAny(path string, patterns []inventoryPattern) bool {
	for _, pattern := range patterns {
		if pattern.re.MatchString(path) {
			return true
		}
	}
	return false
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
