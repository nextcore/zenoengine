# ZenoWasm

Run ZenoEngine in the browser via WebAssembly to power Single Page Applications (SPA) with Zeno Blade templates.

## Build

Prerequisite: Go 1.21+

```bash
GOOS=js GOARCH=wasm go build -o public/zeno.wasm main.go
```

## Run

Serve the `public` directory. You can use any static file server.

```bash
# Example with Python
python3 -m http.server -d public 8080
```

Open `http://localhost:8080` in your browser.

## API

The following functions are exposed to the global JavaScript scope:

- `zenoRegisterTemplate(name: string, content: string)`: Register a Blade template string.
- `zenoRender(templateName: string, dataJson: string|object)`: Render a template with data. Returns HTML string.
- `zenoRenderString(templateContent: string, dataJson: string|object)`: Render a raw template string directly.

## Development

The code in `adapter/` is ported from `internal/slots/` of the main ZenoEngine repo, adapted for WASM environment (Virtual FS, no HTTP server).
