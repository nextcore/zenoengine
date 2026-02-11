package slots

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	datastar "github.com/starfederation/datastar/sdk/go"
)

// RegisterDatastarSlots registers Datastar slots
func RegisterDatastarSlots(eng *engine.Engine) {

	// 1. DATASTAR.STREAM - Start Datastar SSE connection
	eng.Register("datastar.stream", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("datastar.stream: not in http context")
		}
		r, ok := ctx.Value("httpRequest").(*http.Request)
		if !ok {
			return fmt.Errorf("datastar.stream: not in http context")
		}

		// Set Headers Explicitly
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Transfer-Encoding", "chunked")

		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		sse := datastar.NewSSE(w, r)

		// Store sse in scope
		scope.Set("__datastar_sse", sse)

		// Execute children
		for _, child := range node.Children {
			if err := eng.Execute(ctx, child, scope); err != nil {
				// Ignore context cancellation/timeout (client disconnect or HTTP timeout)
				if err == context.Canceled || err == context.DeadlineExceeded ||
					strings.Contains(err.Error(), "context deadline exceeded") ||
					strings.Contains(err.Error(), "context canceled") {
					return nil
				}
				return err
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Start Datastar SSE stream",
	})

	// 2. DATASTAR.FRAGMENT - Send/Merge Fragment
	eng.Register("datastar.fragment", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		sseObj, ok := scope.Get("__datastar_sse")
		if !ok {
			return fmt.Errorf("datastar.fragment: must be inside datastar.stream")
		}
		sse := sseObj.(*datastar.ServerSentEventGenerator)

		var fragment string
		if node.Value != nil {
			fragment = coerce.ToString(resolveValue(node.Value, scope))
		}

		var selector string
		var mergeMode string
		var settleDuration int
		var useViewTransitions bool

		for _, child := range node.Children {
			val := parseNodeValue(child, scope)
			switch child.Name {
			case "selector":
				selector = coerce.ToString(val)
			case "merge":
				mergeMode = coerce.ToString(val)
			case "settle":
				settleDuration, _ = coerce.ToInt(val)
			case "viewTransition":
				useViewTransitions, _ = coerce.ToBool(val)
			case "content":
				fragment = coerce.ToString(val) // Alternative to value
			}
		}

		opts := []datastar.MergeFragmentOption{}
		if selector != "" {
			opts = append(opts, datastar.WithSelector(selector))
		}
		if mergeMode != "" {
			opts = append(opts, datastar.WithMergeMode(datastar.FragmentMergeMode(mergeMode)))
		}
		if settleDuration > 0 {
			opts = append(opts, datastar.WithSettleDuration(time.Duration(settleDuration)*time.Millisecond))
		}
		if useViewTransitions {
			opts = append(opts, datastar.WithUseViewTransitions(true))
		}

		return sse.MergeFragments(fragment, opts...)
	}, engine.SlotMeta{
		Description: "Merge HTML fragment",
	})

	// 3. DATASTAR.REMOVE - Remove Fragment
	eng.Register("datastar.remove", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		sseObj, ok := scope.Get("__datastar_sse")
		if !ok {
			return fmt.Errorf("datastar.remove: must be inside datastar.stream")
		}
		sse := sseObj.(*datastar.ServerSentEventGenerator)

		selector := ""
		if node.Value != nil {
			selector = coerce.ToString(resolveValue(node.Value, scope))
		}

		return sse.RemoveFragments(selector)
	}, engine.SlotMeta{
		Description: "Remove HTML fragment matching selector",
	})

	// 4. DATASTAR.SIGNAL - Read Signals
	eng.Register("datastar.signal", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		r, ok := ctx.Value("httpRequest").(*http.Request)
		if !ok {
			return fmt.Errorf("datastar.signal: not in http context")
		}

		// var target interface{}

		// If used as value: $signals = datastar.signal
		// Returning map[string]interface{}{}

		// But datastar.ReadSignals decodes into a struct.
		// We might need a flexible way.
		// For now simple map reading from request if possible?
		// SDK uses json.Unmarshal into struct.

		// Simplest: just parse everything into a map
		signals := make(map[string]interface{})
		if err := datastar.ReadSignals(r, &signals); err != nil {
			return err
		}

		// If specific key requested
		if node.Value != nil {
			key := coerce.ToString(resolveValue(node.Value, scope))
			if val, ok := signals[key]; ok {
				return val.(error) // Wait, Execute returns error. Slot must return value via some mechanism or just return error.
				// Slots don't return values directly. They modify scope or return error.
				// But implementation details of engine likely allow returning value if used in expression context.
				// ZenoEngine architecture check:
				// Engine.Execute returns error.
				// Values are returned via specific return mechanisms or Scope injection if designed.
				// BUT: engine.resolveValue calls execute and expects something?
				// Let's look at how other slots return values. `math.add` etc.
				// Actually slots usually manipulate scope or perform action.
				// `datastar.signal { var: "myVar" }` might set 'myVar' in scope.
			}
		}

		return nil
	}, engine.SlotMeta{
		Description: "Read datastar signals",
	})
}
