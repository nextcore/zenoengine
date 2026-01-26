package slots

import (
	"context"
	"fmt"
	"reflect"
	"sync"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// TestStats tracks the results of the test execution
type TestStats struct {
	Total  int
	Passed int
	Failed int
	Errors []string
	mu     sync.Mutex
}

func (s *TestStats) AddPass() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Total++
	s.Passed++
}

func (s *TestStats) AddFail(name string, err error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Total++
	s.Failed++
	s.Errors = append(s.Errors, fmt.Sprintf("FAIL [%s]: %v", name, err))
}

type contextKey string

const statsKey contextKey = "testStats"

// WithTestStats injects the stats into the context
func WithTestStats(ctx context.Context, stats *TestStats) context.Context {
	return context.WithValue(ctx, statsKey, stats)
}

func RegisterTestSlots(eng *engine.Engine) {
	// SLOT: test
	eng.Register("test", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		testName := coerce.ToString(node.Value)
		if testName == "" {
			testName = "Unnamed Test"
		}

		fmt.Printf("RUN   %s...\n", testName)

		// Execute children (assertions)
		var err error
		for _, child := range node.Children {
			if e := eng.Execute(ctx, child, scope); e != nil {
				err = e
				break // Stop on first failure in a test case? Or continue? Usually stop.
			}
		}

		stats, ok := ctx.Value(statsKey).(*TestStats)
		if ok {
			if err != nil {
				fmt.Printf("FAIL  %s\n", testName)
				stats.AddFail(testName, err)
				// We return nil to allow other independent tests to run.
				// The error is recorded in stats.
				return nil
			} else {
				fmt.Printf("PASS  %s\n", testName)
				stats.AddPass()
			}
		}

		return nil
	}, engine.SlotMeta{Example: "test: 'My Test' { ... }"})

	// SLOT: assert.eq
	eng.Register("assert.eq", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Expect 2 args.
		// If used as property: assert.eq: $actual { expected: 10 }
		// If used as expression logic (not fully supported by parser yet, assuming children/attributes)

		var actual, expected interface{}

		// Style 1: assert.eq: $val, 10 (Not supported by standard parser easily without custom parsing)
		// Style 2: assert.eq: $val { expected: 10 }

		actual = resolveValue(node.Value, scope)

		for _, c := range node.Children {
			if c.Name == "expected" || c.Name == "val" {
				expected = parseNodeValue(c, scope)
			}
		}

		// Fallback: If node.Value acts as expected and we have actual in children? No, conventionally Value is the subject.

		// Comparison
		if !reflect.DeepEqual(actual, expected) {
			// Try coercing to string for loose comparison if types differ
			if coerce.ToString(actual) != coerce.ToString(expected) {
				return fmt.Errorf("expected %v, got %v", expected, actual)
			}
		}

		return nil
	}, engine.SlotMeta{Example: "assert.eq: $result { expected: 10 }"})

	// SLOT: assert.neq
	eng.Register("assert.neq", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var actual, expected interface{}
		actual = resolveValue(node.Value, scope)
		for _, c := range node.Children {
			if c.Name == "expected" || c.Name == "val" {
				expected = parseNodeValue(c, scope)
			}
		}

		if reflect.DeepEqual(actual, expected) {
			return fmt.Errorf("expected value to NOT be %v", actual)
		}
		// Loose check
		if coerce.ToString(actual) == coerce.ToString(expected) {
			return fmt.Errorf("expected value to NOT be %v", actual)
		}
		return nil
	}, engine.SlotMeta{})

	// SLOT: call
	// Usage: call: math.add { val: 10; val: 20; as: $res }
	eng.Register("call", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		slotName := coerce.ToString(node.Value)

		// Create a dynamic node to execute
		// We copy children to the new node to pass arguments
		callNode := &engine.Node{
			Name:     slotName,
			Children: node.Children,
		}

		return eng.Execute(ctx, callNode, scope)
	}, engine.SlotMeta{Example: "call: math.add { val: 1; as: $res }"})
}
