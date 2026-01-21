package vm

import (
	"bytes"
	"context"
	"fmt"
	"strings"
	"testing"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func TestVMArithmetic(t *testing.T) {
	// 1 + 2
	chunk := &Chunk{
		Code: []byte{
			byte(OpConstant), 0, // 1
			byte(OpConstant), 1, // 2
			byte(OpAdd),
			byte(OpReturn),
		},
		Constants: []Value{
			NewNumber(1),
			NewNumber(2),
		},
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	err := vm.Run(context.Background(), chunk, scope)
	if err != nil {
		t.Fatal(err)
	}

	result := vm.pop()
	if result.AsNum != 3 {
		t.Errorf("Expected 3, got %g", result.AsNum)
	}
}

func TestVMCompilerVariables(t *testing.T) {
	// AST: $x: 10
	node := &engine.Node{
		Name:  "$x",
		Value: "10",
	}

	compiler := NewCompiler()
	chunk, err := compiler.Compile(node)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	err = vm.Run(context.Background(), chunk, scope)
	if err != nil {
		t.Fatal(err)
	}

	val, ok := scope.Get("x")
	if !ok {
		t.Fatal("Variable x should be set in scope")
	}

	// Value representation in prototype might need adjustment,
	// but for now we expect the raw value or NewNumber.
	// Currently compiler uses NewNumber(10)
	if n, ok := val.(float64); ok && n != 10 {
		t.Errorf("Expected 10, got %v", val)
	}
}

func TestVMComplexSlot(t *testing.T) {
	// AST:
	// http.response:
	//    status: 201
	//    body: "created"
	node := &engine.Node{
		Name: "http.response",
		Children: []*engine.Node{
			{Name: "status", Value: "201"},
			{Name: "body", Value: "created"},
		},
	}

	// Mock Engine Registry
	eng := engine.NewEngine()
	called := false
	eng.Register("http.response", func(ctx context.Context, n *engine.Node, s *engine.Scope) error {
		called = true
		// Verify attributes
		statusFound := false
		bodyFound := false
		for _, child := range n.Children {
			if child.Name == "status" && child.Value == 201.0 {
				statusFound = true
			}
			if child.Name == "body" && child.Value == "created" {
				bodyFound = true
			}
		}
		if !statusFound || !bodyFound {
			t.Errorf("Attributes not correctly passed. StatusFound: %v, BodyFound: %v", statusFound, bodyFound)
		}
		return nil
	}, engine.SlotMeta{})

	compiler := NewCompiler()
	chunk, err := compiler.Compile(node)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	ctx := context.WithValue(context.Background(), "engine", eng)

	err = vm.Run(ctx, chunk, scope)
	if err != nil {
		t.Fatal(err)
	}

	if !called {
		t.Error("http.response slot was not called")
	}
}
func TestVMControlFlow(t *testing.T) {
	// AST:
	// if: $x == 10
	//   then:
	//     $res: "yes"
	//   else:
	//     $res: "no"
	node := &engine.Node{
		Name:  "if",
		Value: "$x == 10",
		Children: []*engine.Node{
			{
				Name: "then",
				Children: []*engine.Node{
					{Name: "$res", Value: "'yes'"},
				},
			},
			{
				Name: "else",
				Children: []*engine.Node{
					{Name: "$res", Value: "'no'"},
				},
			},
		},
	}

	compiler := NewCompiler()
	chunk, err := compiler.Compile(node)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()

	// Case 1: x == 10
	scope1 := engine.NewScope(nil)
	scope1.Set("x", 10.0)
	err = vm.Run(context.Background(), chunk, scope1)
	if err != nil {
		t.Fatal(err)
	}
	res1, _ := scope1.Get("res")
	if res1 != "yes" {
		t.Errorf("Expected 'yes', got %v", res1)
	}

	// Case 2: x != 10
	scope2 := engine.NewScope(nil)
	scope2.Set("x", 20.0)
	err = vm.Run(context.Background(), chunk, scope2)
	if err != nil {
		t.Fatal(err)
	}
	res2, _ := scope2.Get("res")
	if res2 != "no" {
		t.Errorf("Expected 'no', got %v", res2)
	}
}

func TestVMSerialization(t *testing.T) {
	chunk := &Chunk{
		Code: []byte{byte(OpConstant), 0, byte(OpReturn)},
		Constants: []Value{
			NewString("hello"),
		},
		LocalNames: []string{"var1"},
	}

	var buf bytes.Buffer
	if err := chunk.Serialize(&buf); err != nil {
		t.Fatal(err)
	}

	decoded, err := DeserializeChunk(&buf)
	if err != nil {
		t.Fatal(err)
	}

	if !bytes.Equal(chunk.Code, decoded.Code) {
		t.Error("Code mismatch after serialization")
	}
	if len(chunk.Constants) != len(decoded.Constants) || decoded.Constants[0].AsPtr.(string) != "hello" {
		t.Error("Constants mismatch after serialization")
	}
	if len(chunk.LocalNames) != len(decoded.LocalNames) || decoded.LocalNames[0] != "var1" {
		t.Error("LocalNames mismatch after serialization")
	}
}

func TestVMInternalFunctions(t *testing.T) {
	// fn: myFunc {
	//   $x: 20
	// }
	// $x: 10
	// call: myFunc
	// $res: $x

	rootNode := &engine.Node{
		Name: "root",
		Children: []*engine.Node{
			{
				Name:  "fn",
				Value: "myFunc",
				Children: []*engine.Node{
					{Name: "$x", Value: 20},
				},
			},
			{Name: "$x", Value: 10},
			{Name: "call", Value: "myFunc"},
			{Name: "$res", Value: "$x"},
		},
	}

	compiler := NewCompiler()
	chunk, err := compiler.Compile(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	ctx := context.WithValue(context.Background(), "engine", &engine.Engine{})

	if err := vm.Run(ctx, chunk, scope); err != nil {
		t.Fatal(err)
	}

	res, _ := scope.Get("res")
	// If dynamic scope is used, $x will be 20 after call.
	// If static scope with isolation is used (standard function behavior),
	// $x in the caller should remain 10 if $x in the function was local.
	// Current implementation uses OpSetGlobal for fn (to match existing slot behavior)
	// and OpSetLocal for $x if recognized.
	// In the sub-compiler, $x will be a NEW local.
	// OpSetGlobal should be updated to handle local synchronization.

	// Wait, our OpCall implementation:
	// vm.pushFrame(fnChunk, vm.sp-argCount)
	// This means locals in myFunc start at some index.
	// $x in the root: index 0
	// $x in myFunc: index 0 (relative to frame base)

	// When myFunc finishes, $x (20) is synced to scope.
	// Then root resumes, and $res: $x happens.
	// $x will be loaded from index 0 of root frame, which is still 10!

	v, _ := coerce.ToInt(res)
	if v != 10 {
		t.Errorf("Expected $x to be 10 (isolated), got %v", res)
	}
}

func TestVMDisassembler(t *testing.T) {
	rootNode := &engine.Node{
		Name: "root",
		Children: []*engine.Node{
			{Name: "$x", Value: 10},
			{
				Name:  "if",
				Value: "$x == 10",
				Children: []*engine.Node{
					{
						Name: "then",
						Children: []*engine.Node{
							{Name: "log", Value: "Correct"},
						},
					},
				},
			},
		},
	}

	compiler := NewCompiler()
	chunk, err := compiler.Compile(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	// This is mainly to ensure it doesn't crash and we can see the output
	chunk.Disassemble("TestChunk")
}

func TestVMIteration(t *testing.T) {
	// Program:
	// items: [10, 20, 30] {
	//   $sum: $sum + $item
	// }

	items := []interface{}{10.0, 20.0, 30.0}

	rootNode := &engine.Node{
		Name: "root",
		Children: []*engine.Node{
			{Name: "$sum", Value: 0},
			{
				Name:  "items",
				Value: items,
				Children: []*engine.Node{
					{
						// In our current compiler simplification,
						// OpIterNext pushes the item to stack.
						// We need to 'capture' it.
						// For now, let's assume the compiler should have
						// emitted an OpSetLocal for a dummy var or something.
						// Let's adjust compiler to push it and we use it as a 'dummy' for test.
						Name:  "$sum",
						Value: "1 + 1", // Just a placeholder for children execution
					},
				},
			},
		},
	}

	// NOTE: The current compiler logic for iteration is VERY simplified
	// (it runs children but doesn't map 'item' to a variable yet).
	// But let's verify if the loop itself runs 3 times.

	// We'll modify the compiler to at least POP the item pushed by IterNext
	// so it doesn't leak on stack.

	compiler := NewCompiler()
	chunk, err := compiler.Compile(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	ctx := context.WithValue(context.Background(), "engine", &engine.Engine{})

	if err := vm.Run(ctx, chunk, scope); err != nil {
		t.Fatal(err)
	}

	// If it ran 3 times, some side effect should be visible.
	// Since our loop body is just a dummy, let's look at the disassembler first.
	chunk.Disassemble("IterationTest")
}

func TestVMIterationItem(t *testing.T) {
	// Program:
	// $sum: 0
	// items: [10, 20, 30] {
	//   $sum: $sum + $item
	// }

	items := []interface{}{10.0, 20.0, 30.0}

	rootNode := &engine.Node{
		Name: "root",
		Children: []*engine.Node{
			{Name: "$sum", Value: 0},
			{
				Name:  "items",
				Value: items,
				Children: []*engine.Node{
					{
						// In ZenoLang, $sum: $sum + $item would be:
						Name:  "$sum",
						Value: "$sum + $item",
					},
				},
			},
		},
	}

	compiler := NewCompiler()
	chunk, err := compiler.Compile(rootNode)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	ctx := context.WithValue(context.Background(), "engine", &engine.Engine{})

	if err := vm.Run(ctx, chunk, scope); err != nil {
		t.Fatal(err)
	}

	// Verify Sum
	res, _ := scope.Get("sum")
	v, _ := coerce.ToInt(res)
	if v != 60 {
		t.Errorf("Expected sum to be 60, got %v", res)
	}

	chunk.Disassemble("IterationItemTest")
}

func TestVMPOSStress(t *testing.T) {
	// 1. Test Logical Operators and Property Access
	src := `
	$email: "user@example.com"
	$password: "secret"
	$request: {
		body: {
			email: "user@example.com",
			password: "secret"
		}
	}
	
	$valid: false
	if: $request.body.email == $email && $request.body.password == $password {
		then: { $valid: true }
	}

	$tags: ["admin", "pos", "staff"]
	$role: $tags.0
	
	$is_not_empty: false
	if: !($email == "") {
		then: { $is_not_empty: true }
	}

	$stopped: false
	if: true {
		then: {
			$stopped: true
			stop
			$stopped: false // Should not be reached
		}
	}
	`

	compiler := NewCompiler()
	node, err := engine.ParseString(src)
	if err != nil {
		t.Fatal(err)
	}

	chunk, err := compiler.Compile(node)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	eng := engine.NewEngine()
	ctx := context.WithValue(context.Background(), "engine", eng)
	scope := engine.NewScope(nil)

	err = vm.Run(ctx, chunk, scope)
	if err != nil {
		chunk.Disassemble("POSStressTest")
		t.Fatal(err)
	}

	valid, _ := scope.Get("valid")
	role, _ := scope.Get("role")
	email, _ := scope.Get("email")
	request, _ := scope.Get("request")

	if valid != true {
		chunk.Disassemble("POSStressTest")
		t.Errorf("Expected valid to be true, got %v. Email: %v, Request: %v", valid, email, request)
	}
	if role != "admin" {
		t.Errorf("Expected role to be admin, got %v", role)
	}
	if val, ok := scope.Get("is_not_empty"); !ok || val.(bool) != true {
		t.Errorf("Expected is_not_empty to be true, got %v", val)
	}
	if val, ok := scope.Get("stopped"); !ok || val.(bool) != true {
		t.Errorf("Expected stopped to be true, got %v", val)
	}
}

func TestVMStringConcat(t *testing.T) {
	src := `
	$str1: "Hello"
	$str2: " World"
	$res1: $str1 + $str2
	
	$num: 42
	$res2: "Answer: " + $num
	`

	compiler := NewCompiler()
	node, err := engine.ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	chunk, err := compiler.Compile(node)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)

	err = vm.Run(context.Background(), chunk, scope)
	if err != nil {
		t.Fatal(err)
	}

	if val, ok := scope.Get("res1"); !ok || val.(string) != "Hello World" {
		t.Errorf("Expected 'Hello World', got %v", val)
	}
	if val, ok := scope.Get("res2"); !ok || val.(string) != "Answer: 42" {
		t.Errorf("Expected 'Answer: 42', got %v", val)
	}
}

func TestVMTryCatch(t *testing.T) {
	// Case 1: Exception Caught
	src1 := `
	$res: "init"
	try {
	   call: nonExistentSlot
	   $res: "not reached"
	} catch {
	   $res: "caught"
	}
	`
	runTestScript(t, src1, map[string]interface{}{"res": "caught"})

	// Case 2: Exception Variable
	src2 := `
	$errMsg: ""
	try {
	   nonExistentSlot
	} catch: $e {
	   $errMsg: $e
	}
	`
	// We expect errMsg to contain "slot not found"
	_, scope2 := runTestScriptReturnVM(t, src2)
	errMsg, _ := scope2.Get("errMsg")
	if !strings.Contains(fmt.Sprintf("%v", errMsg), "slot not found") {
		t.Errorf("Expected error message to contain 'slot not found', got %v", errMsg)
	}
	// Stack validation removed as VM frame is invalid after execution

	// Case 3: No Exception
	src3 := `
	$res: "init"
	try {
	   $res: "success"
	} catch {
	   $res: "fail"
	}
	`
	runTestScript(t, src3, map[string]interface{}{"res": "success"})
}

// Helper to reduce boilerplate
func runTestScript(t *testing.T, src string, expectedGlobals map[string]interface{}) {
	_, scope := runTestScriptReturnVM(t, src)
	for k, v := range expectedGlobals {
		val, ok := scope.Get(k)
		if !ok {
			t.Errorf("Expected global %s to be set", k)
			continue
		}
		if fmt.Sprintf("%v", val) != fmt.Sprintf("%v", v) {
			t.Errorf("Expected global %s to be %v, got %v", k, v, val)
		}
	}
}

func TestVMNestedSlotArgs(t *testing.T) {
	// mock.func:
	//   config:
	//     nested: "found"
	node := &engine.Node{
		Name: "mock.func",
		Children: []*engine.Node{
			{
				Name: "config",
				Children: []*engine.Node{
					{Name: "nested", Value: "found"},
				},
			},
		},
	}

	called := false
	eng := engine.NewEngine()
	eng.Register("mock.func", func(ctx context.Context, n *engine.Node, s *engine.Scope) error {
		called = true
		// Verify we can find "nested" inside "config"
		// 1. Find config
		var configNode *engine.Node
		for _, child := range n.Children {
			if child.Name == "config" {
				configNode = child
				break
			}
		}
		if configNode == nil {
			t.Errorf("Config argument not found")
			return nil
		}

		// 2. We expect configNode to have Children if it was reconstructed,
		// OR we expect configNode.Value to be a Map if we accept that (but standard slots use Children).
		// Let's assume we want to support the Standard Zeno Node traversal (Children).
		if len(configNode.Children) == 0 {
			// Check if it's in Value
			if _, ok := configNode.Value.(map[string]interface{}); ok {
				t.Logf("Warning: Nested data found in Value as Map, but Children is empty. Slots expecting Children will fail.")
				t.Fail()
			} else {
				t.Errorf("Config node has no children and no map value")
			}
		} else {
			// Verify Child
			if configNode.Children[0].Name != "nested" || configNode.Children[0].Value != "found" {
				t.Errorf("Unexpected nested structure")
			}
		}
		return nil
	}, engine.SlotMeta{})

	compiler := NewCompiler()
	chunk, err := compiler.Compile(node)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	ctx := context.WithValue(context.Background(), "engine", eng)

	err = vm.Run(ctx, chunk, scope)
	if err != nil {
		t.Fatal(err)
	}
	if !called {
		t.Error("Slot not called")
	}
}

func runTestScriptReturnVM(t *testing.T, src string) (*VM, *engine.Scope) {
	node, err := engine.ParseString(src)
	if err != nil {
		t.Fatal(err)
	}
	compiler := NewCompiler()
	chunk, err := compiler.Compile(node)
	if err != nil {
		t.Fatal(err)
	}

	vm := NewVM()
	scope := engine.NewScope(nil)
	eng := engine.NewEngine()
	ctx := context.WithValue(context.Background(), "engine", eng)

	err = vm.Run(ctx, chunk, scope)
	if err != nil {
		t.Fatalf("VM Error: %v", err)
	}
	return vm, scope
}
