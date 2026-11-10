package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"time"

	"riders.agency/pulse/api/internal/platform/app"
)

func main() {
	root := repoRoot()
	modulesDir := filepath.Join(root, "apps/api/internal/modules")
	ownershipPath := filepath.Join(root, "docs/architecture/module-ownership.tsv")
	exceptionsPath := filepath.Join(root, "docs/architecture/module-ownership-exceptions.tsv")
	ownership, err := loadOwnership(ownershipPath)
	if err != nil {
		fail(err)
	}
	exceptions, err := loadExceptions(exceptionsPath)
	if err != nil {
		fail(err)
	}
	if err := checkOwnershipInventory(modulesDir, ownership); err != nil {
		fail(err)
	}
	if err := checkDuplicateRegistryKeys(ownership); err != nil {
		fail(err)
	}
	if err := checkAppCompositionImports(filepath.Join(root, "apps/api/internal/platform/app/http_modules.go")); err != nil {
		fail(err)
	}
	if err := checkBuildModuleNames(ownership); err != nil {
		fail(err)
	}
	if err := checkCrossModuleRepositoryImports(root, modulesDir, exceptions); err != nil {
		fail(err)
	}
	if err := checkRepositorySideEffectImports(root, modulesDir, exceptions); err != nil {
		fail(err)
	}
	if err := checkTransactionOwners(root, modulesDir, exceptions); err != nil {
		fail(err)
	}
	if err := checkSQLTableOwnership(root, modulesDir, ownership, exceptions); err != nil {
		fail(err)
	}
	fmt.Println("Module contract check passed")
}

type ownershipRow struct {
	Module          string
	HTTP            []string
	RuntimeActions  []string
	RuntimeTriggers []string
	OutboxEvents    []string
	PrimaryTables   []string
}

type ownershipException struct {
	Kind       string
	Source     string
	Target     string
	PathPrefix string
	Reason     string
	Owner      string
	Expires    string
}

func repoRoot() string {
	cwd, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	for {
		if _, err := os.Stat(filepath.Join(cwd, "go.work")); err == nil {
			return cwd
		}
		next := filepath.Dir(cwd)
		if next == cwd {
			fail(fmt.Errorf("go.work not found from %s", cwd))
		}
		cwd = next
	}
}

func checkOwnershipInventory(modulesDir string, ownership map[string]ownershipRow) error {
	actual, err := moduleDirectories(modulesDir)
	if err != nil {
		return err
	}
	owned := sortedOwnershipModules(ownership)
	if diff := setDiff(actual, owned); len(diff) > 0 {
		return fmt.Errorf("module ownership inventory missing modules: %s", strings.Join(diff, ", "))
	}
	if diff := setDiff(owned, actual); len(diff) > 0 {
		return fmt.Errorf("module ownership inventory has unknown modules: %s", strings.Join(diff, ", "))
	}
	return nil
}

func moduleDirectories(modulesDir string) ([]string, error) {
	entries, err := os.ReadDir(modulesDir)
	if err != nil {
		return nil, err
	}
	out := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() {
			out = append(out, entry.Name())
		}
	}
	sort.Strings(out)
	return out, nil
}

func loadOwnership(path string) (map[string]ownershipRow, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	out := map[string]ownershipRow{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 6 {
			return nil, fmt.Errorf("module ownership row must have 6 tab-separated fields: %q", line)
		}
		name := strings.TrimSpace(fields[0])
		if name == "" || strings.Contains(name, "/") {
			return nil, fmt.Errorf("invalid module ownership name %q", name)
		}
		if _, ok := out[name]; ok {
			return nil, fmt.Errorf("duplicate module ownership row %q", name)
		}
		out[name] = ownershipRow{
			Module:          name,
			HTTP:            splitList(fields[1]),
			RuntimeActions:  splitList(fields[2]),
			RuntimeTriggers: splitList(fields[3]),
			OutboxEvents:    splitList(fields[4]),
			PrimaryTables:   splitList(fields[5]),
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func sortedOwnershipModules(ownership map[string]ownershipRow) []string {
	out := make([]string, 0, len(ownership))
	for name := range ownership {
		out = append(out, name)
	}
	sort.Strings(out)
	return out
}

func loadExceptions(path string) ([]ownershipException, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var out []ownershipException
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fields := strings.Split(line, "\t")
		if len(fields) != 7 {
			return nil, fmt.Errorf("module ownership exception row must have 7 tab-separated fields: %q", line)
		}
		ex := ownershipException{
			Kind:       strings.TrimSpace(fields[0]),
			Source:     strings.TrimSpace(fields[1]),
			Target:     strings.TrimSpace(fields[2]),
			PathPrefix: strings.TrimSpace(fields[3]),
			Reason:     strings.TrimSpace(fields[4]),
			Owner:      strings.TrimSpace(fields[5]),
			Expires:    strings.TrimSpace(fields[6]),
		}
		if ex.Kind == "" || ex.Source == "" || ex.Target == "" || ex.PathPrefix == "" || ex.Reason == "" || ex.Owner == "" || ex.Expires == "" {
			return nil, fmt.Errorf("module ownership exception has empty required field: %q", line)
		}
		expiry, err := time.Parse("2006-01-02", ex.Expires)
		if err != nil {
			return nil, fmt.Errorf("module ownership exception %s %s->%s has invalid expires date %q", ex.Kind, ex.Source, ex.Target, ex.Expires)
		}
		if time.Now().After(expiry.Add(24 * time.Hour)) {
			return nil, fmt.Errorf("module ownership exception expired: %s %s->%s expired %s", ex.Kind, ex.Source, ex.Target, ex.Expires)
		}
		out = append(out, ex)
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return out, nil
}

func splitList(value string) []string {
	value = strings.TrimSpace(value)
	if value == "" || value == "-" {
		return nil
	}
	parts := strings.Split(value, ",")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part != "" && part != "-" {
			out = append(out, part)
		}
	}
	return out
}

func checkDuplicateRegistryKeys(ownership map[string]ownershipRow) error {
	for _, group := range []struct {
		name string
		keys func(ownershipRow) []string
	}{
		{name: "http route", keys: func(row ownershipRow) []string { return row.HTTP }},
		{name: "runtime action", keys: func(row ownershipRow) []string { return row.RuntimeActions }},
		{name: "runtime trigger", keys: func(row ownershipRow) []string { return row.RuntimeTriggers }},
		{name: "outbox event", keys: func(row ownershipRow) []string { return row.OutboxEvents }},
	} {
		seen := map[string]string{}
		for _, module := range sortedOwnershipModules(ownership) {
			for _, key := range group.keys(ownership[module]) {
				if previous := seen[key]; previous != "" {
					return fmt.Errorf("duplicate %s key %q in modules %s and %s", group.name, key, previous, module)
				}
				seen[key] = module
			}
		}
	}
	return nil
}

func checkAppCompositionImports(path string) error {
	forbidden := regexp.MustCompile(`riders\.agency/pulse/api/internal/modules/[^"]+/(transport|repository)(/|")`)
	var violations []string
	raw, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	for _, line := range strings.Split(string(raw), "\n") {
		if forbidden.MatchString(line) {
			violations = append(violations, filepath.ToSlash(path)+": "+strings.TrimSpace(line))
		}
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		return fmt.Errorf("platform/app must register root modules, not transport/repository packages:\n%s", strings.Join(violations, "\n"))
	}
	return nil
}

func checkBuildModuleNames(ownership map[string]ownershipRow) error {
	seen := map[string]struct{}{}
	for _, name := range app.BuildModules().Names() {
		name = strings.TrimSpace(name)
		if name == "" {
			return fmt.Errorf("BuildModules contains module with empty name")
		}
		prefix := strings.SplitN(name, ".", 2)[0]
		if prefix != "api" {
			if _, ok := ownership[prefix]; !ok {
				return fmt.Errorf("BuildModules contains %q but module ownership has no %q row", name, prefix)
			}
		}
		if _, ok := seen[name]; ok {
			return fmt.Errorf("BuildModules contains duplicate module name %q", name)
		}
		seen[name] = struct{}{}
	}
	return nil
}

func checkCrossModuleRepositoryImports(root, modulesDir string, exceptions []ownershipException) error {
	re := regexp.MustCompile(`riders\.agency/pulse/api/internal/modules/([^/]+)/repository`)
	var violations []string
	err := filepath.WalkDir(modulesDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		source := sourceModule(modulesDir, path)
		if source == "" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for _, match := range re.FindAllStringSubmatch(string(raw), -1) {
			target := match[1]
			if target == source {
				continue
			}
			rel := filepath.ToSlash(relativePath(root, path))
			if hasException(exceptions, "repository_import", source, target, rel) {
				continue
			}
			violations = append(violations, fmt.Sprintf("%s imports %s repository from %s", source, target, rel))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		return fmt.Errorf("cross-module repository imports require explicit module-ownership-exceptions.tsv entries:\n%s", strings.Join(violations, "\n"))
	}
	return nil
}

func checkRepositorySideEffectImports(root, modulesDir string, exceptions []ownershipException) error {
	forbidden := map[string]string{
		`"net/http"`: "net/http",
		`"riders.agency/pulse/api/internal/platform/httptransport"`: "httptransport",
		`"riders.agency/pulse/api/internal/platform/outbox"`:        "outbox",
	}
	var violations []string
	err := filepath.WalkDir(modulesDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(relativePath(root, path))
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") || !strings.Contains(rel, "/repository/") {
			return nil
		}
		source := sourceModule(modulesDir, path)
		if source == "" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		for literal, target := range forbidden {
			if !strings.Contains(string(raw), literal) {
				continue
			}
			if hasException(exceptions, "repository_side_effect_import", source, target, rel) {
				continue
			}
			violations = append(violations, fmt.Sprintf("%s repository imports %s from %s", source, target, rel))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		return fmt.Errorf("repository side-effect imports require explicit module-ownership-exceptions.tsv entries:\n%s", strings.Join(violations, "\n"))
	}
	return nil
}

func checkTransactionOwners(root, modulesDir string, exceptions []ownershipException) error {
	var violations []string
	err := filepath.WalkDir(modulesDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(relativePath(root, path))
		if entry.IsDir() || filepath.Ext(path) != ".go" || strings.HasSuffix(path, "_test.go") {
			return nil
		}
		layer := ""
		switch {
		case strings.Contains(rel, "/repository/"):
			layer = "repository"
		case strings.Contains(rel, "/transport/"):
			layer = "transport"
		default:
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(raw), ".Begin(") {
			return nil
		}
		source := sourceModule(modulesDir, path)
		if !hasException(exceptions, "transaction_owner", source, layer, rel) {
			violations = append(violations, fmt.Sprintf("%s %s owns transaction in %s", source, layer, rel))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		return fmt.Errorf("repository/transport transaction owners require explicit module-ownership-exceptions.tsv entries:\n%s", strings.Join(violations, "\n"))
	}
	return nil
}

func checkSQLTableOwnership(root, modulesDir string, ownership map[string]ownershipRow, exceptions []ownershipException) error {
	readContracts, err := loadReadContracts(root, ownership)
	if err != nil {
		return err
	}
	tableOwners := map[string]string{}
	for module, row := range ownership {
		for _, table := range row.PrimaryTables {
			if table == "-" {
				continue
			}
			if previous := tableOwners[table]; previous != "" && previous != module {
				return fmt.Errorf("table %q is owned by both %s and %s", table, previous, module)
			}
			tableOwners[table] = module
		}
	}
	if len(tableOwners) == 0 {
		return nil
	}
	tablePattern := regexp.MustCompile(`(?i)\b(DELETE\s+FROM|FROM|JOIN|UPDATE|INTO)\s+([a-z_][a-z0-9_]*(?:\.[a-z_][a-z0-9_]*)?)`)
	var violations []string
	err = filepath.WalkDir(modulesDir, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() || filepath.Ext(path) != ".go" {
			return nil
		}
		source := sourceModule(modulesDir, path)
		if source == "" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		rel := filepath.ToSlash(relativePath(root, path))
		for _, match := range tablePattern.FindAllStringSubmatch(string(raw), -1) {
			table := strings.ToLower(strings.TrimSpace(match[2]))
			if strings.HasPrefix(table, "pulse_read.") {
				if err := checkReadReference(readContracts, source, strings.ToUpper(match[1]), table); err != nil {
					violations = append(violations, rel+": "+err.Error())
				}
				continue
			}
			table = strings.TrimPrefix(table, "public.")
			target := tableOwners[table]
			if target == "" || target == source {
				continue
			}
			if hasException(exceptions, "sql_table", source, target, rel) {
				continue
			}
			violations = append(violations, fmt.Sprintf("%s references %s table %q from %s", source, target, table, rel))
		}
		return nil
	})
	if err != nil {
		return err
	}
	if len(violations) > 0 {
		sort.Strings(violations)
		return fmt.Errorf("cross-module SQL table references require explicit module-ownership-exceptions.tsv entries:\n%s", strings.Join(violations, "\n"))
	}
	return nil
}

func sourceModule(modulesDir, path string) string {
	rel, err := filepath.Rel(modulesDir, path)
	if err != nil || strings.HasPrefix(rel, "..") {
		return ""
	}
	parts := strings.Split(filepath.ToSlash(rel), "/")
	if len(parts) == 0 {
		return ""
	}
	return parts[0]
}

func hasException(exceptions []ownershipException, kind, source, target, relPath string) bool {
	for _, ex := range exceptions {
		if ex.Kind == kind && ex.Source == source && ex.Target == target && strings.HasPrefix(relPath, filepath.ToSlash(ex.PathPrefix)) {
			return true
		}
	}
	return false
}

func relativePath(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return rel
}

func setDiff(left, right []string) []string {
	rightSet := map[string]struct{}{}
	for _, item := range right {
		rightSet[item] = struct{}{}
	}
	var out []string
	for _, item := range left {
		if _, ok := rightSet[item]; !ok {
			out = append(out, item)
		}
	}
	sort.Strings(out)
	return out
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, err)
	os.Exit(1)
}
