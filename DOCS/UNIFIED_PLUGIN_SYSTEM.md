# Sidecar & WASM Plugin Guide

## Overview
ZenoEngine supports a unified plugin architecture where slots can be implemented either as sandboxed WASM modules or as high-performance Native Sidecar processes. Both types follow the same communication protocol and manifest structure.

## 1. Plugin Manifest Structure (`manifest.yaml`)

```yaml
name: php-high-speed
version: 1.2.0
type: sidecar          # Options: wasm (default), sidecar
binary: php-worker.exe # For WASM: .wasm file | For Sidecar: Native executable

# Sidecar-specific configuration
sidecar:
  protocol: json-rpc   # Options: json-rpc (stdin/stdout)
  auto_start: true     # Automatically start process on engine load
  keep_alive: true     # Restart if crashed
  max_retries: 3       # Max restart attempts

permissions:
  network: ["*"]
  filesystem: ["./storage"]
```

## 2. Communication Protocol (Shared)
Both WASM and Sidecar plugins use the existing JSON protocol defined in `WASM_PLUGIN_SPEC.md`.

### Execution Request (Host -> Guest)
```json
{
  "slot_name": "php.run",
  "parameters": {
    "script": "index.php",
    "data": { "user_id": 1 }
  }
}
```

### Execution Response (Guest -> Host)
```json
{
  "success": true,
  "data": { "status": "ok" },
  "error": null
}
```

## 3. Sidecar Architecture
The Plugin Manager handles `type: sidecar` plugins transparently:

1.  **Process Lifecycle**: Uses Go's `os/exec` to start the sidecar process on demand or at engine startup.
2.  **IPC (Inter-Process Communication)**:
    -   **StdIn/StdOut**: Primary channel for JSON-RPC. Highly portable across Windows, Linux, and macOS.
3.  **Health Monitoring**: Monitors the sidecar process PID. If it crashes, the manager automatically restarts it based on `max_retries` config.
4.  **Resource Cleanup**: Ensures all sidecar processes are killed when ZenoEngine shuts down.

## 4. "Plug-and-Play" Features
-   **Universal Slot Registration**: You can write ZenoLang code without knowing if the slot is WASM or Sidecar.
-   **Zero-Config Binary selection**: The `manifest.yaml` points to the correct binary for the platform.
-   **Lean Core**: Native binaries are kept in the plugin folder, keeping the ZenoEngine core purely Go.

## 5. Ecosystem Example: .NET 8/9 Sidecar
Because Sidecar plugins use standard OS processes, you can use languages like **C# / .NET**, **Python**, **Node.js**, or **PHP** to build high-performance plugins.

### Sample C# Sidecar (`Program.cs`)
```csharp
using System.Text.Json;

while (Console.ReadLine() is { } line) {
    var request = JsonDocument.Parse(line);
    var slotName = request.RootElement.GetProperty("slot_name").GetString();

    object result = slotName switch {
        "dotnet.greet" => new { message = "Hello from .NET NativeAOT!" },
        _ => new { error = "Unknown slot" }
    };

    Console.WriteLine(JsonSerializer.Serialize(new {
        success = true,
        data = result
    }));
}
```

### Manifest for .NET Plugin
```yaml
name: dotnet-utils
type: sidecar
binary: ./DotnetPlugin.exe # Compiled with NativeAOT
sidecar:
  protocol: json-rpc
```

For a working PHP example, see `examples/sidecar-plugins/php-native/`.
