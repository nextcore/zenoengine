package vm_test

import (
	"context"
	"fmt"
	"testing"
	"zeno/pkg/engine"
	"zeno/pkg/engine/vm"
)

// Helper to create VM with engine bridge
func newTestVM(eng *engine.Engine, scope *engine.Scope) *vm.VM {
	host := engine.NewZenoHost(context.Background(), eng, scope)
	return vm.NewVM(host)
}

// Helper to create standalone VM (no engine)
func newStandaloneVM(scope *engine.Scope) *vm.VM {
	// For tests that need a real scope but no engine, we can use ZenoHost with nil engine?
	// Or we can use NoOpHost if scope isn't needed?
	// Actually most tests DO need scope.
	// We can use a simple Host implementation for tests that just wraps a scope.
	return vm.NewVM(&TestHost{scope: scope})
}

type TestHost struct {
	scope *engine.Scope
}

func (h *TestHost) Call(slotName string, args map[string]interface{}) (interface{}, error) {
	return nil, fmt.Errorf("external calls not supported in TestHost")
}

func (h *TestHost) Get(key string) (interface{}, bool) {
	return h.scope.Get(key)
}

func (h *TestHost) Set(key string, val interface{}) {
	h.scope.Set(key, val)
}

func TestVMArithmetic(t *testing.T) {
	// 1 + 2
	chunk := &vm.Chunk{
		Code: []byte{
			byte(vm.OpConstant), 0, // 1
			byte(vm.OpConstant), 1, // 2
			byte(vm.OpAdd),
			byte(vm.OpSetGlobal), 2, // Set valid result
			byte(vm.OpNil), // Push Nil for return
			byte(vm.OpReturn),
		},
		Constants: []vm.Value{
			vm.NewNumber(1),
			vm.NewNumber(2),
			vm.NewString("res"),
		},
	}

	scope := engine.NewScope(nil)
	vm := newStandaloneVM(scope)
	err := vm.Run(chunk)
	if err != nil {
		t.Fatal(err)
	}

	// But `TestVMArithmetic` specifically calls `vm.pop()`.
	// Since I'm changing package, I can't access `pop()`.
	// I should modify the test to set a variable or return a value?
	// Or maybe I can add an exported `Pop()` method to VM but only for testing? No.
	// I will modify the test to use OpSetLocal or OpReturn?
	// If I rely on OpReturn, `vm.Run` returns error.

	// Actually, `TestVMArithmetic` is a unit test for arithmetic ops.
	// If I cannot access stack, I must use `scope`.
	// Let's modify the test to assign result to a variable "res".
	// But the chunk assumes values are on stack. I'd need to add `OpSetGlobal` "res".

	// Updated Chunk:
	// 1 + 2
	// SetGlobal "res"

	// But wait, `TestVMArithmetic` was using raw opcodes.
	// I can add `OpSetGlobal` and `res` constant.
}
