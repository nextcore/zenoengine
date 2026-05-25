# API Documentation (apidoc)

The `apidoc` package is a self-contained, thread-safe, and zero-dependency Go module designed to programmatically generate OpenAPI 3.0.0 compliant specifications and serve Swagger UI documentation. 

While it is used natively by ZenoEngine to generate documentation for ZenoLang routes, it is designed to be easily embedded and used in any pure Go project.

---

## Installation

If you are using ZenoEngine as a dependency, you can import the package directly:

```go
import "zeno/pkg/apidoc"
```

If you are in another project, you can copy the `pkg/apidoc` directory into your project structure, or import it directly if within the same Go workspace/module.

---

## Basic Usage

The package exposes a global `apidoc.Registry` where you can register your API endpoints programmatically.

### 1. Registering Routes

To document a route, define a `RouteDoc` struct and register it to the global registry:

```go
package main

import (
	"zeno/pkg/apidoc"
)

func registerUserEndpoints() {
	apidoc.Registry.Register("GET", "/api/v1/users", &apidoc.RouteDoc{
		Summary:     "Get all users",
		Description: "Fetch a list of active users from the database.",
		Tags:        []string{"Users"},
		Params: []apidoc.ParamDoc{
			{
				Name:        "limit",
				In:          "query",
				Description: "Maximum number of users to return",
				Required:    false,
				Type:        "integer",
			},
		},
		Responses: map[string]apidoc.ResponseDoc{
			"200": {
				Description: "A JSON array of user objects",
			},
		},
	})
}
```

### 2. Generating OpenAPI JSON

You can generate the complete OpenAPI 3.0.0 specification using `ToJSON()`:

```go
jsonBytes, err := apidoc.Registry.ToJSON()
if err != nil {
    // Handle error
}
```

---

## Serving Swagger UI

The `apidoc` package includes a built-in `SwaggerUIHandler` helper function that serves the interactive Swagger UI interface directly using a lightweight CDN-based HTML template.

Here is a complete, working example using Go's standard library `net/http`:

```go
package main

import (
	"net/http"
	"zeno/pkg/apidoc"
)

func main() {
	// Register an example endpoint
	apidoc.Registry.Register("GET", "/api/v1/ping", &apidoc.RouteDoc{
		Summary:     "Ping Check",
		Description: "Returns pong if the server is healthy",
		Tags:        []string{"System"},
		Responses: map[string]apidoc.ResponseDoc{
			"200": {Description: "Server is healthy"},
		},
	})

	// 1. Serve the OpenAPI specification JSON
	http.HandleFunc("/swagger.json", func(w http.ResponseWriter, r *http.Request) {
		jsonBytes, err := apidoc.Registry.ToJSON()
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.Write(jsonBytes)
	})

	// 2. Serve the interactive Swagger UI page (Tinggal Pakai!)
	http.HandleFunc("/docs", apidoc.SwaggerUIHandler("/swagger.json"))

	// Start standard Go web server
	http.ListenAndServe(":8080", nil)
}
```

Now, navigating to `http://localhost:8080/docs` in your browser will display the fully interactive Swagger UI documentation for your Go project.

---

## API Reference

### Struct: `RouteDoc`

Defines the structure of a single API route/endpoint documentation block.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Summary` | `string` | A short summary of what the operation does. |
| `Description` | `string` | A verbose explanation of the operation behavior. |
| `Tags` | `[]string` | A list of tags for API grouping. |
| `Params` | `[]ParamDoc` | URL path, query string, or header parameters. |
| `RequestBody` | `*RequestBodyDoc` | The request body definition (optional). |
| `Responses` | `map[string]ResponseDoc` | HTTP status codes mapped to their expected responses. |

### Struct: `ParamDoc`

Defines a path, query, or header parameter.

| Field | Type | Description |
| :--- | :--- | :--- |
| `Name` | `string` | The name of the parameter (case-sensitive). |
| `In` | `string` | The location of the parameter: `"path"`, `"query"`, or `"header"`. |
| `Description` | `string` | A brief description of the parameter. |
| `Required` | `bool` | Whether this parameter is mandatory. |
| `Type` | `string` | The data type (e.g. `"string"`, `"integer"`, `"boolean"`). |
