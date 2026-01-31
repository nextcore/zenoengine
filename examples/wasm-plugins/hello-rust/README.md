# Hello Rust WASM Plugin

A simple WASM plugin for ZenoEngine written in Rust.

## Features

- **rust.greet** - Greet someone with a personalized message
- **rust.calculate** - Perform basic arithmetic operations

## Building

### Prerequisites
- Rust 1.70+ with `wasm32-wasi` target

### Install WASI target (if not already installed)
```bash
rustup target add wasm32-wasip1
```

### Build
```bash
cargo build --target wasm32-wasip1 --release
```

The compiled WASM file will be at:
```
target/wasm32-wasip1/release/hello_rust_plugin.wasm
```

### Copy to plugins directory
```bash
# Create plugin directory
mkdir -p ../../../plugins/hello-rust

# Copy WASM binary
cp target/wasm32-wasip1/release/hello_rust_plugin.wasm ../../../plugins/hello-rust/hello.wasm

# Copy manifest
cp manifest.yaml ../../../plugins/hello-rust/
```

## Usage

### 1. Enable plugins in .env
```bash
ZENO_PLUGINS_ENABLED=true
```

### 2. Use in ZenoLang

```zenolang
# Greet
rust.greet {
    name: "ZenoEngine"
}
log.info: $message  # "Hello from Rust, ZenoEngine! 🦀"

# Calculate
rust.calculate {
    a: 10
    b: 5
    op: "add"
}
log.info: $result  # 15
```

## Binary Size

Rust WASM plugins are typically **smaller** than Go/TinyGo equivalents:
- Unoptimized: ~200KB
- Optimized (with `opt-level="z"` and `strip`): ~50-80KB
- With `wasm-opt`: ~30-50KB

## Performance

Rust WASM plugins generally have:
- **Faster execution** than interpreted languages
- **Lower memory usage** than Go plugins
- **Near-native performance** for compute-heavy tasks
