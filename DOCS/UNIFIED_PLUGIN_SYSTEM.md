# Unified Plugin System Design (WASM & Sidecar)

## Overview
ZenoEngine will support a unified plugin architecture where slots can be implemented either as sandboxed WASM modules or as high-performance Native Sidecar processes. Both types will follow the same communication protocol and manifest structure.

## 1. Updated Manifest Structure (`manifest.yaml`)

```yaml
name: php-high-speed
version: 1.2.0
type: sidecar          # Options: wasm (default), sidecar
binary: php-worker.exe # For WASM: .wasm file | For Sidecar: Native executable

# Sidecar-specific configuration
sidecar:
  protocol: json-rpc   # Options: json-rpc (stdin/stdout), fastcgi, http
  auto_start: true
  keep_alive: true
  max_retries: 3

permissions:
  network: ["*"]
  filesystem: ["./storage"]
```

## 2. Communication Protocol (Shared)
Both WASM and Sidecar plugins will use the existing JSON protocol defined in `WASM_PLUGIN_SPEC.md`.

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

## 3. Sidecar Manager Architecture
The `PluginManager` will be extended to handle `type: sidecar`:

1.  **Process Lifecycle**: Uses Go's `os/exec` to start the sidecar process on demand or at engine startup.
2.  **IPC (Inter-Process Communication)**:
    -   **StdIn/StdOut**: Primary channel for JSON-RPC. Highly portable across Windows and Linux.
    -   **Shared Memory (Future)**: For zero-copy data transfer in extreme performance scenarios.
3.  **Health Monitoring**: Monitors the sidecar process PID. If it crashes, the manager can automatically restart it.
4.  **Resource Cleanup**: Ensures all sidecar processes are killed when ZenoEngine shuts down.

## 4. Why this is "Plug-and-Play"?
-   **Universal Slot Registration**: The user writes ZenoLang code without knowing if the slot is WASM or Sidecar.
-   **Zero-Config Binary selection**: The `manifest.yaml` can eventually support OS-specific binaries:
    ```yaml
    binaries:
      windows: ./bin/win/worker.exe
      linux: ./bin/linux/worker
      wasm: ./bin/universal.wasm
    ```
-   **Lean Core**: Native binaries are kept in the plugin folder, keeping the ZenoEngine core purely Go.

## 5. Ecosystem Example: .NET 8/9 Sidecar
Because Sidecar plugins use standard OS processes, you can use **C# / .NET** to build high-performance plugins.

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

## 6. Implementation Path
1.  **Refactor `LoadedPlugin`**: Add a `Driver` interface (WASMDriver, SidecarDriver).
2.  **Implement `SidecarDriver`**: Handles process spawning and pipe-based communication.
3.  **Update `PluginManager.LoadPlugin`**: Branch logic based on `type` field in manifest.
