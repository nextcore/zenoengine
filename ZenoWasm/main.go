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
	adapter.RegisterRouterSlots(eng)

	// Expose functions to JS
	js.Global().Set("zenoRegisterTemplate", js.FuncOf(registerTemplate))
	js.Global().Set("zenoRender", js.FuncOf(render))
	js.Global().Set("zenoRenderString", js.FuncOf(renderString))

	// Router APIs
	js.Global().Set("zenoInitRouter", js.FuncOf(initRouter))
	js.Global().Set("zenoNavigate", js.FuncOf(navigate))

	// Inject Datastar
	injectDatastar()

	fmt.Println("ZenoWasm Initialized 🚀")

	// Prevent exit
	select {}
}

// initRouter registers routes defined in a Zeno script
func initRouter(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return "error: missing routes script content"
	}
	scriptContent := args[0].String()

	// Parse and Execute Script to Register Routes
	// We wrap it in a dummy node execution context
	ctx := context.Background()
	scope := engine.NewScope(nil)

	// Need to parse script content into AST first.
	// Since we don't have direct parser access here easily (it's in pkg/engine/parser.go but internal logic might need engine helpers),
	// we will rely on adapter.RegisterTemplate to store it as a virtual file, then load it.

	adapter.RegisterTemplate("routes.zl", scriptContent)

	// Create a node to include/execute this routes file
	// Actually, we can just use engine.LoadScript if we adapted it, but LoadScript reads from OS.
	// We need to use our virtual adapter.
	// Best hack: Use 'include' slot logic which we haven't fully exposed for virtual yet.
	// Simpler: Just parse the string directly if we can access Parser.
	// Since engine.LoadScript is for files, let's use a helper in adapter or main.
	// Wait, we can't easily parse raw string without creating a file.
	// Workaround: We execute a virtual "view.blade" which contains the script.
	// But view.blade parses as Blade (HTML mixed). Routes script is pure logic.

	// Solution: We assume the user passes a ZenoLang script.
	// We can use `engine.ParseString` if exported.

	program, err := engine.ParseString(scriptContent, "routes.zl")
	if err != nil {
		fmt.Println("Router Init Error:", err)
		return err.Error()
	}

	if err := eng.Execute(ctx, program, scope); err != nil {
		fmt.Println("Router Execute Error:", err)
		return err.Error()
	}

	fmt.Println("Routes Initialized.")

	// Trigger initial navigation
	path := js.Global().Get("window").Get("location").Get("pathname").String()
	navigate(js.Value{}, []js.Value{js.ValueOf(path)})

	return nil
}

func navigate(this js.Value, args []js.Value) interface{} {
	if len(args) < 1 {
		return nil
	}
	path := args[0].String()
	fmt.Println("Navigating to:", path)

	// Match Route
	route, params := adapter.MatchRoute(path)
	if route == nil {
		fmt.Println("No route found for:", path)
		// Render 404 if registered, or ignore
		return nil
	}

	// Create Context & Scope
	var buf bytes.Buffer
	ctx := context.Background()
	ctx = context.WithValue(ctx, "httpWriter", &buf)
	ctx = context.WithValue(ctx, "engine", eng)

	scope := engine.NewScope(nil)
	// Inject Params
	for k, v := range params {
		scope.Set(k, v)
	}

	// Execute Handler
	if err := eng.Execute(ctx, route.Handler, scope); err != nil {
		fmt.Println("Route Handler Error:", err)
		return err.Error()
	}

	// Update DOM
	html := buf.String()
	doc := js.Global().Get("document")
	app := doc.Call("getElementById", "app")
	if !app.IsNull() {
		app.Set("innerHTML", html)

		// Update Browser URL (History PushState)
		// Only push if it's a new navigation, not initial load
		// For simplicity here we assume simple forward nav.
		// Check if current path matches to avoid duplicates
		current := js.Global().Get("window").Get("location").Get("pathname").String()
		if current != path {
			js.Global().Get("history").Call("pushState", nil, "", path)
		}
	}

	return nil
}

func injectDatastar() {
	doc := js.Global().Get("document")
	if doc.IsNull() || doc.IsUndefined() {
		return // Not in browser env
	}

	if !js.Global().Get("Datastar").IsUndefined() {
		return
	}

	head := doc.Call("querySelector", "head")
	script := doc.Call("createElement", "script")
	script.Set("type", "module")
	script.Set("textContent", embed.DatastarSource)
	script.Set("id", "datastar-embedded")

	head.Call("appendChild", script)

	// Inject Client-Side Link Interceptor for Routing
	scriptRouter := doc.Call("createElement", "script")
	scriptRouter.Set("textContent", `
		document.addEventListener("click", (e) => {
			const link = e.target.closest("a");
			if (link && link.href && link.href.startsWith(window.location.origin)) {
				e.preventDefault();
				const path = new URL(link.href).pathname;
				zenoNavigate(path);
			}
		});

		window.addEventListener("popstate", () => {
			zenoNavigate(window.location.pathname);
		});
	`)
	head.Call("appendChild", scriptRouter)

	fmt.Println("Libraries injected successfully.")
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
