package main

import (
	"fmt"

	"log"
	"net/http"
	"os"
	"strings"
)

// --- AST Definition ---

type Node struct {
	Name     string
	Value    string
	Children []*Node
}

// --- Minimal Lexer & Parser ---
// We will do a very simplified parsing:
// It expects lines in the format: `name: value {` or `name: {` or `name: value` or `}`
// It ignores comments.

func parse(input string) (*Node, error) {
	lines := strings.Split(input, "\n")
	root := &Node{Name: "root"}
	stack := []*Node{root}

	for _, line := range lines {
		line = strings.TrimSpace(line)

		// Ignore comments and empty lines
		if line == "" || strings.HasPrefix(line, "//") {
			continue
		}

		// Handle block end
		if line == "}" {
			if len(stack) > 1 {
				stack = stack[:len(stack)-1]
			}
			continue
		}

		// Handle block start or single slot
		hasOpenBrace := strings.HasSuffix(line, "{")
		if hasOpenBrace {
			line = strings.TrimSuffix(line, "{")
			line = strings.TrimSpace(line)
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) == 0 || parts[0] == "" {
			continue
		}

		name := strings.TrimSpace(parts[0])
		value := ""
		if len(parts) > 1 {
			value = strings.TrimSpace(parts[1])
			// Strip quotes if any
			value = strings.TrimPrefix(value, "\"")
			value = strings.TrimSuffix(value, "\"")
		}

		node := &Node{Name: name, Value: value}
		parent := stack[len(stack)-1]
		parent.Children = append(parent.Children, node)

		if hasOpenBrace {
			stack = append(stack, node)
		}
	}

	return root, nil
}

// --- Minimal Engine/Executor ---

type HandlerFunc func(node *Node, w http.ResponseWriter, r *http.Request)

type Engine struct {
	Registry map[string]HandlerFunc
}

func NewEngine() *Engine {
	return &Engine{
		Registry: make(map[string]HandlerFunc),
	}
}

func (e *Engine) Register(name string, handler HandlerFunc) {
	e.Registry[name] = handler
}

func (e *Engine) Execute(node *Node, w http.ResponseWriter, r *http.Request) {
	for _, child := range node.Children {
		if handler, ok := e.Registry[child.Name]; ok {
			handler(child, w, r)
		} else {
			fmt.Printf("Warning: Slot '%s' not found in registry\n", child.Name)
		}
	}
}

// --- Main ---

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: go run main.go <script.zl>")
		return
	}

	content, err := os.ReadFile(os.Args[1])
	if err != nil {
		log.Fatalf("Failed to read file: %v", err)
	}

	ast, err := parse(string(content))
	if err != nil {
		log.Fatalf("Parse error: %v", err)
	}

	engine := NewEngine()

	// Register basic slots
	engine.Register("log", func(node *Node, w http.ResponseWriter, r *http.Request) {
		fmt.Println("LOG:", node.Value)
	})

	// Simple router registry
	routes := make(map[string]*Node)

	engine.Register("http.get", func(node *Node, w http.ResponseWriter, r *http.Request) {
		if w == nil { // Initialization phase
			path := node.Value
			routes[path] = node
			fmt.Println("Registered route:", path)
		}
	})

	// Initialization pass
	engine.Execute(ast, nil, nil)

	// Start HTTP Server
	if len(routes) > 0 {
		http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "GET" {
				return
			}
			if routeNode, ok := routes[r.URL.Path]; ok {
				// We need a sub-executor to handle the block inside http.get
				for _, child := range routeNode.Children {
					if child.Name == "body" {
						fmt.Fprint(w, child.Value)
					} else if handler, ok := engine.Registry[child.Name]; ok {
						handler(child, w, r)
					}
				}
			} else {
				http.NotFound(w, r)
			}
		})

		fmt.Println("Minimal Engine listening on :8080...")
		log.Fatal(http.ListenAndServe(":8080", nil))
	} else {
		fmt.Println("Execution finished. No routes registered.")
	}
}
