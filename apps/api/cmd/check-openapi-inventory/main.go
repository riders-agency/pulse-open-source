package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"strings"

	yaml "gopkg.in/yaml.v3"
)

type inventoryOperation struct {
	Method        string
	Path          string
	OperationID   string
	RuntimeStatus string
}

type operationSet map[string]inventoryOperation

func main() {
	inventoryPath := flag.String("inventory", "../../contracts/openapi/endpoints.inventory.tsv", "endpoint inventory TSV")
	openapiPath := flag.String("openapi", "../../contracts/openapi/dist/openapi.yaml", "bundled OpenAPI YAML")
	flag.Parse()

	inventory, err := readInventory(*inventoryPath)
	if err != nil {
		fail("read inventory: %v", err)
	}
	contract, err := readOpenAPI(*openapiPath)
	if err != nil {
		fail("read openapi: %v", err)
	}

	var violations int
	for key, op := range inventory {
		contractOp, ok := contract[key]
		if !ok {
			fmt.Printf("VIOLATION openapi_missing_inventory_operation method=%s path=%s operation_id=%s\n", op.Method, op.Path, op.OperationID)
			violations++
			continue
		}
		if contractOp.OperationID != op.OperationID {
			fmt.Printf("VIOLATION openapi_operation_id_drift method=%s path=%s inventory=%s openapi=%s\n", op.Method, op.Path, op.OperationID, contractOp.OperationID)
			violations++
		}
	}
	for key, op := range contract {
		if _, ok := inventory[key]; !ok {
			if op.RuntimeStatus == "preview" {
				continue
			}
			fmt.Printf("VIOLATION openapi_operation_missing_from_inventory method=%s path=%s operation_id=%s\n", op.Method, op.Path, op.OperationID)
			violations++
		}
	}
	if violations > 0 {
		os.Exit(1)
	}
	fmt.Printf("OpenAPI inventory check passed: operations=%d\n", len(contract))
}

func readInventory(path string) (operationSet, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	ops := operationSet{}
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid inventory row %q", line)
		}
		op := inventoryOperation{
			Method:      strings.ToUpper(strings.TrimSpace(parts[0])),
			Path:        normalizePath(parts[1]),
			OperationID: strings.TrimSpace(parts[3]),
		}
		if op.Method == "" || op.Path == "" || op.OperationID == "" {
			return nil, fmt.Errorf("invalid inventory row %q", line)
		}
		ops[op.key()] = op
	}
	return ops, scanner.Err()
}

func readOpenAPI(path string) (operationSet, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var doc map[string]any
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil, err
	}
	paths := asMap(doc["paths"])
	ops := operationSet{}
	for rawPath, rawPathItem := range paths {
		pathItem := asMap(rawPathItem)
		for rawMethod, rawOperation := range pathItem {
			method := strings.ToUpper(strings.TrimSpace(rawMethod))
			if !isHTTPMethod(method) {
				continue
			}
			operation := asMap(rawOperation)
			operationID := stringValue(operation["operationId"])
			if operationID == "" {
				return nil, fmt.Errorf("missing operationId method=%s path=%s", method, rawPath)
			}
			op := inventoryOperation{
				Method:        method,
				Path:          normalizePath(rawPath),
				OperationID:   operationID,
				RuntimeStatus: strings.ToLower(stringValue(operation["x-runtime-status"])),
			}
			ops[op.key()] = op
		}
	}
	return ops, nil
}

func (op inventoryOperation) key() string {
	return op.Method + "\t" + op.Path
}

func normalizePath(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return "/"
	}
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

func asMap(value any) map[string]any {
	if out, ok := value.(map[string]any); ok {
		return out
	}
	return map[string]any{}
}

func stringValue(value any) string {
	if out, ok := value.(string); ok {
		return strings.TrimSpace(out)
	}
	return ""
}

func isHTTPMethod(method string) bool {
	switch method {
	case "GET", "POST", "PUT", "PATCH", "DELETE":
		return true
	default:
		return false
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, format+"\n", args...)
	os.Exit(1)
}
