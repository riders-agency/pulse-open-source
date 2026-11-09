package httpserver

import (
	_ "embed"
	"encoding/json"
	"net/http"
	"sort"
	"strings"

	"github.com/go-chi/chi/v5"
	yaml "gopkg.in/yaml.v3"
)

//go:embed openapi_pulse_stub.yaml
var openAPISlice []byte

//go:embed doc_inventory.tsv
var docInventory []byte

func mountPulseDocs(ar chi.Router) {
	ar.Get("/docs/openapi.yaml", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/yaml; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(openAPISlice)
	})
	ar.Get("/docs/openapi.json", func(w http.ResponseWriter, _ *http.Request) {
		var doc any
		if err := yaml.Unmarshal(openAPISlice, &doc); err != nil {
			http.Error(w, "invalid embedded openapi", http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(doc)
		if err != nil {
			http.Error(w, "encode", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	})
	ar.Get("/docs/endpoints", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/tab-separated-values; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(docInventory)
	})
	ar.Get("/docs/ai-tools.json", func(w http.ResponseWriter, _ *http.Request) {
		tools, err := buildAITools()
		if err != nil {
			http.Error(w, "invalid embedded openapi", http.StatusInternalServerError)
			return
		}
		b, err := json.Marshal(map[string]any{"tools": tools})
		if err != nil {
			http.Error(w, "encode", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write(b)
	})
	ar.Get("/docs/testing/postman", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		_, _ = w.Write([]byte(`{"info":{"name":"Riders Pulse — import OpenAPI JSON from /api/v1/docs/openapi.json","schema":"https://schema.getpostman.com/json/collection/v2.1.0/collection.json"},"item":[]}`))
	})
}

func MountPulseDocs(ar chi.Router) {
	mountPulseDocs(ar)
}

type aiToolDefinition struct {
	Name        string `json:"name"`
	Description string `json:"description"`
	Method      string `json:"method"`
	Path        string `json:"path"`
	InputSchema any    `json:"input_schema"`
	Auth        any    `json:"auth"`
	Examples    any    `json:"examples,omitempty"`
}

func buildAITools() ([]aiToolDefinition, error) {
	var doc map[string]any
	if err := yaml.Unmarshal(openAPISlice, &doc); err != nil {
		return nil, err
	}

	inventory := parseInventory(docInventory)
	paths := asMap(doc["paths"])
	tools := make([]aiToolDefinition, 0, len(paths))
	for path, rawPathItem := range paths {
		for method, rawOperation := range asMap(rawPathItem) {
			method = strings.ToUpper(method)
			if !isHTTPMethod(method) {
				continue
			}
			operation := asMap(rawOperation)
			name := stringValue(operation["operationId"])
			if name == "" {
				name = inventory[method+"\t"+path]
			}
			if name == "" {
				name = fallbackOperationID(method, path)
			}

			description := stringValue(operation["summary"])
			if description == "" {
				description = stringValue(operation["description"])
			}
			if description == "" {
				description = method + " " + path
			}

			auth := operation["security"]
			if auth == nil {
				auth = doc["security"]
			}
			if auth == nil {
				auth = map[string]any{"type": "session_or_bearer"}
			}

			tools = append(tools, aiToolDefinition{
				Name:        name,
				Description: description,
				Method:      method,
				Path:        path,
				InputSchema: inputSchemaFor(path, operation),
				Auth:        auth,
				Examples:    examplesFor(operation),
			})
		}
	}

	sort.Slice(tools, func(i, j int) bool {
		if tools[i].Path == tools[j].Path {
			return tools[i].Method < tools[j].Method
		}
		return tools[i].Path < tools[j].Path
	})
	return tools, nil
}

func parseInventory(raw []byte) map[string]string {
	result := make(map[string]string)
	for _, line := range strings.Split(string(raw), "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}
		result[strings.ToUpper(parts[0])+"\t"+parts[1]] = parts[3]
	}
	return result
}

func inputSchemaFor(path string, operation map[string]any) map[string]any {
	properties := map[string]any{}
	required := make([]string, 0)

	for _, name := range pathParameters(path) {
		properties[name] = map[string]any{"type": "string", "x-in": "path"}
		required = append(required, name)
	}

	for _, rawParam := range asSlice(operation["parameters"]) {
		param := asMap(rawParam)
		name := stringValue(param["name"])
		if name == "" {
			continue
		}
		schema := asMap(param["schema"])
		if len(schema) == 0 {
			schema = map[string]any{"type": "string"}
		}
		if where := stringValue(param["in"]); where != "" {
			schema["x-in"] = where
		}
		properties[name] = schema
		if boolValue(param["required"]) {
			required = append(required, name)
		}
	}

	if requestBody := asMap(operation["requestBody"]); len(requestBody) > 0 {
		bodySchema := requestBodyJSONSchema(requestBody)
		if len(bodySchema) > 0 {
			properties["body"] = bodySchema
			if boolValue(requestBody["required"]) {
				required = append(required, "body")
			}
		}
	}

	schema := map[string]any{
		"type":       "object",
		"properties": properties,
	}
	if len(required) > 0 {
		schema["required"] = required
	}
	return schema
}

func pathParameters(path string) []string {
	var out []string
	rest := path
	for {
		start := strings.Index(rest, "{")
		if start < 0 {
			return out
		}
		rest = rest[start+1:]
		end := strings.Index(rest, "}")
		if end < 0 {
			return out
		}
		name := strings.TrimSpace(rest[:end])
		if name != "" {
			out = append(out, name)
		}
		rest = rest[end+1:]
	}
}

func requestBodyJSONSchema(requestBody map[string]any) map[string]any {
	content := asMap(requestBody["content"])
	if media := asMap(content["application/json"]); len(media) > 0 {
		return asMap(media["schema"])
	}
	for _, rawMedia := range content {
		media := asMap(rawMedia)
		if schema := asMap(media["schema"]); len(schema) > 0 {
			return schema
		}
	}
	return nil
}

func examplesFor(operation map[string]any) any {
	requestBody := asMap(operation["requestBody"])
	content := asMap(requestBody["content"])
	for _, rawMedia := range content {
		media := asMap(rawMedia)
		if examples := media["examples"]; examples != nil {
			return examples
		}
		if example := media["example"]; example != nil {
			return example
		}
	}
	return nil
}

func isHTTPMethod(method string) bool {
	switch method {
	case http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete:
		return true
	default:
		return false
	}
}

func fallbackOperationID(method, path string) string {
	replacer := strings.NewReplacer("/", " ", "{", " ", "}", " ", "-", " ", "_", " ")
	parts := strings.Fields(replacer.Replace(path))
	if len(parts) == 0 {
		return strings.ToLower(method)
	}
	for i, part := range parts {
		parts[i] = strings.ToUpper(part[:1]) + part[1:]
	}
	return strings.ToLower(method) + strings.Join(parts, "")
}

func asMap(value any) map[string]any {
	if typed, ok := value.(map[string]any); ok {
		return typed
	}
	return nil
}

func asSlice(value any) []any {
	if typed, ok := value.([]any); ok {
		return typed
	}
	return nil
}

func stringValue(value any) string {
	if typed, ok := value.(string); ok {
		return typed
	}
	return ""
}

func boolValue(value any) bool {
	if typed, ok := value.(bool); ok {
		return typed
	}
	return false
}
