package main

import (
	"bufio"
	"flag"
	"fmt"
	"net/http"
	"os"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"

	"riders.agency/pulse/api/internal/platform/app"
	"riders.agency/pulse/api/internal/platform/httpserver"
)

type routeSet map[string]struct{}

func main() {
	inventoryPath := flag.String("inventory", "../../contracts/openapi/endpoints.inventory.tsv", "endpoint inventory TSV")
	flag.Parse()

	inventory, err := readInventory(*inventoryPath)
	if err != nil {
		fail("read inventory: %v", err)
	}
	runtime, err := runtimeRoutes()
	if err != nil {
		fail("walk runtime routes: %v", err)
	}

	missing := difference(runtime, inventory)
	extra := difference(inventory, runtime)
	if len(missing) == 0 && len(extra) == 0 {
		fmt.Printf("Route inventory check passed: routes=%d\n", len(runtime))
		return
	}

	for _, route := range missing {
		fmt.Printf("VIOLATION runtime_route_missing_from_inventory route=%q\n", route)
	}
	for _, route := range extra {
		fmt.Printf("VIOLATION inventory_route_missing_from_runtime route=%q\n", route)
	}
	os.Exit(1)
}

func readInventory(path string) (routeSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	routes := routeSet{}
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
		method := strings.ToUpper(strings.TrimSpace(parts[0]))
		path := normalizePath(strings.TrimSpace(parts[1]))
		routes[method+"\t"+path] = struct{}{}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	return routes, nil
}

func runtimeRoutes() (routeSet, error) {
	handler := httpserver.NewRouter(httpserver.Dependencies{}, app.BuildModules())
	routes, ok := handler.(chi.Routes)
	if !ok {
		return nil, fmt.Errorf("router does not expose chi routes")
	}

	result := routeSet{}
	err := chi.Walk(routes, func(method string, route string, _ http.Handler, _ ...func(http.Handler) http.Handler) error {
		method = strings.ToUpper(strings.TrimSpace(method))
		if method == "" || method == http.MethodHead || method == http.MethodOptions {
			return nil
		}
		path := normalizePath(route)
		if path == "/hooks" {
			if method == http.MethodPost {
				result[http.MethodPost+"\t/hooks/{hook_key}/{endpoint_key}"] = struct{}{}
			}
			return nil
		}
		result[method+"\t"+path] = struct{}{}
		return nil
	})
	return result, err
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
	path = strings.ReplaceAll(path, "/*", "")
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	for strings.Contains(path, "//") {
		path = strings.ReplaceAll(path, "//", "/")
	}
	if len(path) > 1 {
		path = strings.TrimSuffix(path, "/")
	}
	return path
}

func difference(left, right routeSet) []string {
	out := make([]string, 0)
	for route := range left {
		if _, ok := right[route]; !ok {
			out = append(out, route)
		}
	}
	sort.Strings(out)
	return out
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
