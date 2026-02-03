# Robotics WASM Plugin Example

This example demonstrates how to move core functionality (like robotics control) into a high-performance WASM plugin to keep the ZenoEngine core lean.

## Files
- `main.rs`: Rust source code for the plugin.
- `manifest.yaml`: Plugin configuration and permissions.

## How to Build
1. Install Rust and the WASM target:
   ```bash
   rustup target add wasm32-wasi
   ```
2. Compile the plugin:
   ```bash
   rustc --target wasm32-wasi -O main.rs -o robotics.wasm
   ```
3. Copy `robotics.wasm` and `manifest.yaml` to your ZenoEngine `plugins/robotics/` directory.

## Usage in ZenoLang
Ensure `ZENO_PLUGINS_ENABLED=true` in your `.env`.

```zenolang
robot.sense: "distance" { as: $d }
if: $d < 50 {
    robot.act: "stop"
}
```
