package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"syscall/js"
	"time"

	"zeno-wasm/adapter"
	"zeno-wasm/embed"
	"zeno/pkg/engine"
)

var eng *engine.Engine

func main() {
	// Initialize Engine
	eng = engine.NewEngine()

	// Register Slots
	adapter.RegisterUtilSlots(eng)
	adapter.RegisterLogicSlots(eng)
	adapter.RegisterBladeSlots(eng)

	// Expose functions to JS
	js.Global().Set("zenoRegisterTemplate", js.FuncOf(registerTemplate))
	js.Global().Set("zenoRender", js.FuncOf(render))
	js.Global().Set("zenoRenderString", js.FuncOf(renderString))

	// Inject Datastar
	injectDatastar()

	fmt.Println("ZenoWasm Initialized 🚀")

	// Prevent exit
	select {}
}

func injectDatastar() {
	doc := js.Global().Get("document")
	if doc.IsNull() || doc.IsUndefined() {
		return // Not in browser env
	}

	// Check if already loaded
	// Simple check: see if window.datastar exists or similar, or just check for our script id
	if !js.Global().Get("Datastar").IsUndefined() {
		return
	}

	head := doc.Call("querySelector", "head")
	script := doc.Call("createElement", "script")
	script.Set("type", "module")
	script.Set("textContent", embed.DatastarSource)
	script.Set("id", "datastar-embedded")

	head.Call("appendChild", script)
	fmt.Println("Datastar library injected successfully.")
}

func registerTemplate(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return "error: missing arguments (name, content)"
	}
	name := args[0].String()
	content := args[1].String()

	adapter.RegisterTemplate(name, content)
	return nil
}

func render(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return "error: missing arguments (templateName, data)"
	}
	templateName := args[0].String()

	// Handle Data (String or Object)
	dataArg := args[1]
	var dataJson string

	if dataArg.Type() == js.TypeObject {
		// Convert JS Object to JSON string via JSON.stringify
		jsonObj := js.Global().Get("JSON")
		dataJson = jsonObj.Call("stringify", dataArg).String()
	} else {
		dataJson = dataArg.String()
	}

	// Parse Data
	var data map[string]interface{}
	if err := json.Unmarshal([]byte(dataJson), &data); err != nil {
		return fmt.Sprintf("error parsing json: %v", err)
	}

	// Create Scope
	scope := engine.NewScope(nil)
	for k, v := range data {
		scope.Set(k, v)
	}

	// Create Context with Buffer
	var buf bytes.Buffer
	ctx := context.Background()
	ctx = context.WithValue(ctx, "httpWriter", &buf) // Use &buf (pointer to Buffer which implements io.Writer)
	ctx = context.WithValue(ctx, "engine", eng)

	// Execute View Slot directly?
	// We need to construct a Node that calls "view.blade" with value templateName
	root := &engine.Node{
		Name:  "view.blade",
		Value: templateName,
	}

	if err := eng.Execute(ctx, root, scope); err != nil {
		return fmt.Sprintf("error executing: %v", err)
	}

	return buf.String()
}

func renderString(this js.Value, args []js.Value) interface{} {
	if len(args) < 2 {
		return "error: missing arguments (templateContent, data)"
	}
	templateContent := args[0].String()

	// Handle Data
	dataArg := args[1]

	// Register as temporary template
	tempName := fmt.Sprintf("temp_%d", time.Now().UnixNano())
	adapter.RegisterTemplate(tempName, templateContent)

	// Reuse render logic
	// We construct args for render: [tempName, dataArg]
	return render(this, []js.Value{js.ValueOf(tempName), dataArg})
}
