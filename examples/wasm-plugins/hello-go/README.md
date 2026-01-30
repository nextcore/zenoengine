# Hello World WASM Plugin

Simple example plugin demonstrating ZenoEngine WASM plugin system.

## Features

- **hello.greet** - Greet someone with a personalized message
- **hello.log** - Log messages using host logging

## Building

### Requirements

- [TinyGo](https://tinygo.org/) 0.30.0 or later

### Build Command

```bash
tinygo build -o hello.wasm -target=wasi main.go
```

## Installation

1. Build the plugin (see above)
2. Copy the entire directory to ZenoEngine plugins folder:
   ```bash
   cp -r hello-go/ /path/to/zeno/plugins/
   ```

## Usage

### In ZenoLang Scripts

```zenolang
# Greet someone
hello.greet {
    name: "Alice"
}
# Result: Sets $message = "Hello, Alice! 👋"

# Log a message
hello.log {
    message: "Plugin is working!"
}
```

### Expected Output

```
[WASM Plugin] level=info message="Greeting: Hello, Alice! 👋"
[WASM Plugin] level=info message="[Plugin] Plugin is working!"
```

## File Structure

```
hello-go/
├── main.go          # Plugin source code
├── manifest.yaml    # Plugin configuration
├── hello.wasm       # Compiled WASM binary (after build)
└── README.md        # This file
```

## Development

### Testing Locally

```bash
# Build
tinygo build -o hello.wasm -target=wasi main.go

# Test with ZenoEngine
zeno run test.zl
```

### Debugging

The plugin uses `host_log()` for debugging. Check ZenoEngine logs for output.

## License

MIT
