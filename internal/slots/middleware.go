package slots

import (
	"context"
	"fmt"
	"zeno/pkg/engine"
	"zeno/pkg/middleware"
	"zeno/pkg/utils/coerce"
)

func RegisterMiddlewareSlots(eng *engine.Engine) {
	// MIDDLEWARE.DEFINE
	eng.Register("middleware.define", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		name := coerce.ToString(resolveValue(node.Value, scope))
		if name == "" {
			return fmt.Errorf("middleware.define: name is required")
		}

		// The body of the middleware is the child 'do' block or just children
		// We store the whole node (or a synthetic node containing children)

		var logicNode *engine.Node

		// Check for 'do' block
		for _, c := range node.Children {
			if c.Name == "do" {
				logicNode = c
				break
			}
		}

		// If no 'do', assume all children are logic
		if logicNode == nil {
			logicNode = &engine.Node{
				Name: "do",
				Children: node.Children,
			}
		}

		middleware.Registry.Register(name, logicNode)
		fmt.Printf("   🛡️ [MIDDLEWARE] Defined custom middleware: %s\n", name)

		return nil
	}, engine.SlotMeta{
		Description: "Define a custom middleware logic.",
		Example: `middleware.define: "admin" {
  if: $auth.role != 'admin' {
    abort: 403
  }
}`,
		Group: "HTTP",
	})
}
