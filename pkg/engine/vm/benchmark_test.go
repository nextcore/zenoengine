package vm

import (
	"context"
	"testing"
	"zeno/pkg/engine"
)

// BenchmarkASTWalker measures the current tree-walking interpreter
func BenchmarkASTWalker(b *testing.B) {
	eng := engine.NewEngine()
	node := &engine.Node{
		Name:  "$x",
		Value: "10 + 20",
	}
	// Note: Current Zeno's Execute logic for "$x: 10 + 20"
	// would typically involve resolving shorthand values.

	scope := engine.NewScope(nil)
	ctx := context.Background()

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		_ = eng.Execute(ctx, node, scope)
	}
}

// BenchmarkZenoVM measures the new bytecode VM
func BenchmarkZenoVM(b *testing.B) {
	// Compile the same operation: $x: 10 + 20
	node := &engine.Node{
		Name:  "$x",
		Value: "10 + 20",
	}
	compiler := NewCompiler()
	chunk, _ := compiler.Compile(node)

	scope := engine.NewScope(nil)
	vm := newStandaloneVM(scope)

	b.ResetTimer()
	b.ReportAllocs()
	for i := 0; i < b.N; i++ {
		// In a real scenario, compilation happens once,
		// so we only benchmark the Run step.
		_ = vm.Run(chunk)
	}
}
