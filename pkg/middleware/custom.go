package middleware

import (
	"context"
	"net/http"
	"sync"
	"zeno/pkg/engine"
)

// CustomMiddlewareRegistry stores user-defined middleware nodes
type CustomRegistry struct {
	mu          sync.RWMutex
	definitions map[string]*engine.Node
}

var Registry = &CustomRegistry{
	definitions: make(map[string]*engine.Node),
}

func (r *CustomRegistry) Register(name string, node *engine.Node) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.definitions[name] = node
}

func (r *CustomRegistry) Get(name string) *engine.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.definitions[name]
}

// ZenoMiddleware returns a Chi-compatible middleware that executes a Zeno node.
func ZenoMiddleware(eng *engine.Engine, node *engine.Node) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Create a temporary scope for middleware
			// We need to inject request data similar to router handler?
			// The router already injects request context via chi context middleware if placed correctly?
			// No, standard router handler does injection. Middleware runs BEFORE handler.
			// So middleware logic needs to access Request.

			// However, Zeno Scope is created inside the final handler (createHandler).
			// Middleware doesn't have access to THAT scope yet because it doesn't exist.

			// We must create a scope for the middleware execution.
			// And ideally pass variables down?
			// Chi context allows passing data.

			// For MVP: Middleware runs in its own scope.
			// It can read Request (via context helpers if we implement them, or standard http slots).
			// Slots like `http.header` usage?

			// We need to pass `eng` to execute.

			// Inject httpWriter and httpRequest to context for slots to work
			ctx := r.Context()
			ctx = context.WithValue(ctx, "httpWriter", w)
			ctx = context.WithValue(ctx, "httpRequest", r)

			// Execute Middleware Node
			// Scope is ephemeral
			scope := engine.NewScope(nil)

			err := eng.Execute(ctx, node, scope)
			if err != nil {
				// If middleware fails/aborts, we stop the chain.
				// Slots like `abort` return error.
				// We assume error means "Stop".
				// Logging?
				// fmt.Printf("Middleware Error: %v\n", err)
				return
			}

			// If success, call next
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
