package apidoc

import (
	"encoding/json"
	"strings"
	"sync"
)

// Global Registry
var Registry = &APIRegistry{
	Routes: make(map[string]*RouteDoc),
}

type APIRegistry struct {
	mu     sync.RWMutex
	Routes map[string]*RouteDoc
	Title  string
	Desc   string
}

type RouteDoc struct {
	Method      string                 `json:"method"`
	Path        string                 `json:"path"`
	Summary     string                 `json:"summary"`
	Description string                 `json:"description"`
	Tags        []string               `json:"tags"`
	Params      []ParamDoc             `json:"parameters,omitempty"`
	RequestBody *RequestBodyDoc        `json:"requestBody,omitempty"`
	Responses   map[string]ResponseDoc `json:"responses"`
}

type ParamDoc struct {
	Name        string `json:"name"`
	In          string `json:"in"` // query, path, header
	Description string `json:"description,omitempty"`
	Required    bool   `json:"required"`
	Type        string `json:"type"` // string, integer
}

type RequestBodyDoc struct {
	Content map[string]MediaTypeDoc `json:"content"`
}

type MediaTypeDoc struct {
	Schema SchemaDoc `json:"schema"`
}

type SchemaDoc struct {
	Type       string              `json:"type"`
	Properties map[string]Property `json:"properties,omitempty"`
}

type Property struct {
	Type string `json:"type"`
}

type ResponseDoc struct {
	Description string `json:"description"`
}

func (r *APIRegistry) Register(method, path string, doc *RouteDoc) {
	r.mu.Lock()
	defer r.mu.Unlock()
	key := method + ":" + path
	r.Routes[key] = doc
}

// GenerateOpenAPI returns the full OpenAPI 3.0 JSON structure
func (r *APIRegistry) GenerateOpenAPI() map[string]interface{} {
	r.mu.RLock()
	defer r.mu.RUnlock()

	paths := make(map[string]map[string]interface{})

	for _, route := range r.Routes {
		pathItem, exists := paths[route.Path]
		if !exists {
			pathItem = make(map[string]interface{})
			paths[route.Path] = pathItem
		}

		method := strings.ToLower(route.Method)
		
		operation := map[string]interface{}{
			"summary":     route.Summary,
			"description": route.Description,
			"tags":        route.Tags,
			"responses":   route.Responses,
		}

		if len(route.Params) > 0 {
			operation["parameters"] = route.Params
		}
		
		if route.RequestBody != nil {
			operation["requestBody"] = route.RequestBody
		}

		pathItem[method] = operation
	}

	return map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]string{
			"title":       "ZenoEngine API",
			"version":     "1.0.0",
			"description": "Auto-generated API Documentation",
		},
		"paths": paths,
	}
}

// ToJSON returns the JSON bytes of the OpenAPI spec
func (r *APIRegistry) ToJSON() ([]byte, error) {
	spec := r.GenerateOpenAPI()
	return json.MarshalIndent(spec, "", "  ")
}
