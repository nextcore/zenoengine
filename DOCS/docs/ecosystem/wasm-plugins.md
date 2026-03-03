# WASM Plugin Architecture

ZenoLang is an exceptionally fast interpreter, but there are times when your application requires CPU-intensive computations (such as image processing, cryptography, or heavy data crunching). For these scenarios, ZenoEngine supports a state-of-the-art **WebAssembly (WASM)** Plugin architecture.

## Zero CGO Engine

ZenoEngine uses [wazero](https://wazero.io), a zero CGO WebAssembly runtime. This means you do not need C-toolchains (like GCC/Clang) installed on your server to run or compile the ZenoEngine binary itself. The engine remains 100% portable across Windows, Linux, and macOS while still executing high-performance compiled plugins.

## Building Plugins

You can write your plugins in any language that compiles to WebAssembly (`.wasm`), including:
- **Rust**
- **C/C++**
- **Go** (via TinyGo)
- **AssemblyScript**

Currently, plugins need to be placed securely in the `plugins/` directory of your ZenoEngine installation alongside a `manifest.yaml` configuration file.

### Sandboxing & Permissions

Security is paramount when executing third-party binaries. ZenoEngine employs strict sandboxing. The `manifest.yaml` file defines exactly what Host Functions your WASM code is allowed to access:

```yaml
id: "image_processor_v1"
name: "Image Processor Mutator"
version: "1.0.0"
entrypoint: "process.wasm"
permissions:
  network: false
  filesystem: true
  database: false
  scope: true
```

If a WASM plugin attempts to execute a database query but lacks the `database: true` permission in its manifest, the execution will panic instantly.

## Host Functions

WASM plugins in ZenoEngine aren't just blind calculators. Through **Host Functions**, the engine extends bindings so your Rust/C code can interact with the engine environment:

- `host_db_query`: Execute raw SQL queries using the active internal database connection.
- `host_http_request`: Perform outgoing external API calls.
- `host_scope_set` / `host_scope_get`: Read and write variables dynamically to the active ZenoLang HTTP Request scope!

*Further examples and SDK scaffolding for writing ZenoEngine WASM plugins in Go/Rust will be provided in upcoming developer tools releases.*
