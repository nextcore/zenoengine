# Zig PHP Bridge for ZenoEngine

This plugin allows ZenoEngine to run PHP and Laravel scripts with 100% native performance by using a Zig-compiled sidecar process.

## How to Build
1. Install Zig (0.13.0 or later).
2. Compile the bridge:
   ```bash
   zig build-exe main.zig -O ReleaseSafe --name php_bridge
   ```
3. Ensure you have PHP installed or bundled in this directory.
4. Copy `php_bridge` (or `php_bridge.exe` on Windows) and `manifest.yaml` to your ZenoEngine `plugins/php-native/` directory.

## Why Zig?
Zig provides the best C-interoperability, allowing us to link directly against `libphp` if needed, while producing a tiny, high-performance binary that works perfectly on Windows without dependencies.
