# WASM Plugin Interface Specification

## Overview

ZenoEngine WASM plugins communicate with the host via a JSON-based protocol over WebAssembly linear memory.

## Plugin Exports (Guest → Host)

Plugins **MUST** export the following functions:

### 1. `plugin_init() -> ptr`

Returns plugin metadata as JSON.

**Signature:**
```wat
(func (export "plugin_init") (result i32))
```

**Returns:** Pointer to JSON string in WASM memory

**JSON Format:**
```json
{
  "name": "plugin-name",
  "version": "1.0.0",
  "author": "Author Name",
  "description": "Plugin description",
  "license": "MIT",
  "homepage": "https://example.com"
}
```

### 2. `plugin_register_slots() -> ptr`

Returns array of slot definitions as JSON.

**Signature:**
```wat
(func (export "plugin_register_slots") (result i32))
```

**Returns:** Pointer to JSON array in WASM memory

**JSON Format:**
```json
[
  {
    "name": "plugin.slot_name",
    "description": "Slot description",
    "example": "plugin.slot_name: param_value",
    "inputs": {
      "param1": {
        "type": "string",
        "required": true,
        "description": "Parameter description"
      }
    }
  }
]
```

### 3. `plugin_execute(slot_name_ptr, slot_name_len, params_ptr, params_len) -> ptr`

Executes a slot with given parameters.

**Signature:**
```wat
(func (export "plugin_execute")
  (param $slot_name_ptr i32)
  (param $slot_name_len i32)
  (param $params_ptr i32)
  (param $params_len i32)
  (result i32))
```

**Parameters:**
- `slot_name_ptr`: Pointer to slot name string
- `slot_name_len`: Length of slot name
- `params_ptr`: Pointer to parameters JSON
- `params_len`: Length of parameters JSON

**Returns:** Pointer to response JSON in WASM memory

**Request JSON Format:**
```json
{
  "slot_name": "plugin.slot_name",
  "parameters": {
    "param1": "value1",
    "param2": 123
  },
  "context": {}
}
```

**Response JSON Format:**
```json
{
  "success": true,
  "data": {
    "result": "value"
  },
  "error": null
}
```

Or on error:
```json
{
  "success": false,
  "data": null,
  "error": "Error message"
}
```

### 4. `plugin_cleanup()`

Cleanup resources before unload.

**Signature:**
```wat
(func (export "plugin_cleanup"))
```

### 5. `alloc(size) -> ptr`

Allocate memory for host-to-guest data transfer.

**Signature:**
```wat
(func (export "alloc") (param $size i32) (result i32))
```

**Parameters:**
- `size`: Number of bytes to allocate

**Returns:** Pointer to allocated memory

## Host Exports (Host → Guest)

The host provides the following functions that plugins can import:

### 1. `host_log(level_ptr, level_len, msg_ptr, msg_len)`

Log a message.

**Signature:**
```wat
(import "env" "host_log"
  (func $host_log
    (param $level_ptr i32)
    (param $level_len i32)
    (param $msg_ptr i32)
    (param $msg_len i32)))
```

**Parameters:**
- `level`: "info", "warn", "error", "debug"
- `message`: Log message

**Example:**
```go
hostLog("info", "Hello from plugin!")
```

### 2. `host_db_query(request_ptr, request_len) -> response_ptr`

Execute a database query.

**Signature:**
```wat
(import "env" "host_db_query"
  (func $host_db_query
    (param $request_ptr i32)
    (param $request_len i32)
    (result i32)))
```

**Request JSON:**
```json
{
  "connection": "default",
  "sql": "SELECT * FROM users WHERE id = ?",
  "params": {
    "id": 123
  }
}
```

**Response JSON:**
```json
{
  "success": true,
  "data": {
    "rows": [
      {"id": 123, "name": "John"}
    ]
  }
}
```

### 3. `host_http_request(request_ptr, request_len) -> response_ptr`

Make an HTTP request.

**Signature:**
```wat
(import "env" "host_http_request"
  (func $host_http_request
    (param $request_ptr i32)
    (param $request_len i32)
    (result i32)))
```

**Request JSON:**
```json
{
  "method": "POST",
  "url": "https://api.example.com/endpoint",
  "headers": {
    "Content-Type": "application/json"
  },
  "body": {
    "key": "value"
  }
}
```

**Response JSON:**
```json
{
  "success": true,
  "data": {
    "status": 200,
    "body": {"result": "success"}
  }
}
```

### 4. `host_scope_get(key_ptr, key_len) -> response_ptr`

Get a variable from ZenoLang scope.

**Signature:**
```wat
(import "env" "host_scope_get"
  (func $host_scope_get
    (param $key_ptr i32)
    (param $key_len i32)
    (result i32)))
```

**Response JSON:**
```json
{
  "success": true,
  "data": {
    "value": "variable value"
  }
}
```

### 5. `host_scope_set(request_ptr, request_len) -> success`

Set a variable in ZenoLang scope.

**Signature:**
```wat
(import "env" "host_scope_set"
  (func $host_scope_set
    (param $request_ptr i32)
    (param $request_len i32)
    (result i32)))
```

**Request JSON:**
```json
{
  "key": "variable_name",
  "value": "variable value"
}
```

**Returns:** 1 for success, 0 for failure

### 6. `host_file_read(path_ptr, path_len) -> response_ptr`

Read a file (if permitted).

**Signature:**
```wat
(import "env" "host_file_read"
  (func $host_file_read
    (param $path_ptr i32)
    (param $path_len i32)
    (result i32)))
```

**Response JSON:**
```json
{
  "success": true,
  "data": {
    "content": "file contents"
  }
}
```

### 7. `host_file_write(request_ptr, request_len) -> success`

Write a file (if permitted).

**Signature:**
```wat
(import "env" "host_file_write"
  (func $host_file_write
    (param $request_ptr i32)
    (param $request_len i32)
    (result i32)))
```

**Request JSON:**
```json
{
  "path": "/path/to/file",
  "content": "file contents"
}
```

**Returns:** 1 for success, 0 for failure

### 8. `host_env_get(key_ptr, key_len) -> value_ptr`

Get an environment variable (if permitted).

**Signature:**
```wat
(import "env" "host_env_get"
  (func $host_env_get
    (param $key_ptr i32)
    (param $key_len i32)
    (result i32)))
```

**Returns:** Pointer to value string

## Example: Hello World Plugin (C# / .NET)

Using **NativeAOT** for WASI support in .NET 8 or 9:

```csharp
using System;
using System.Runtime.InteropServices;
using System.Text.Json;
using System.Text;

public class Plugin
{
    [UnmanagedCallersOnly(EntryPoint = "plugin_init")]
    public static int PluginInit()
    {
        var metadata = new { name = "hello-csharp", version = "1.0.0" };
        return AllocJSON(metadata);
    }

    [UnmanagedCallersOnly(EntryPoint = "plugin_register_slots")]
    public static int PluginRegisterSlots()
    {
        var slots = new[] {
            new { name = "csharp.greet", description = "Greet from C#" }
        };
        return AllocJSON(slots);
    }

    [UnmanagedCallersOnly(EntryPoint = "plugin_execute")]
    public static unsafe int PluginExecute(int namePtr, int nameLen, int paramsPtr, int paramsLen)
    {
        string slotName = PtrToString(namePtr, nameLen);
        
        if (slotName == "csharp.greet") {
            var response = new { 
                success = true, 
                data = new { message = "Hello from C# WASM!" } 
            };
            return AllocJSON(response);
        }

        return AllocJSON(new { success = false, error = "unknown slot" });
    }

    [UnmanagedCallersOnly(EntryPoint = "alloc")]
    public static int Alloc(int size)
    {
        IntPtr ptr = Marshal.AllocHGlobal(size);
        return (int)ptr;
    }

    private static int AllocJSON(object val)
    {
        byte[] json = JsonSerializer.SerializeToUtf8Bytes(val);
        int ptr = Alloc(json.Length + 1);
        unsafe {
            Marshal.Copy(json, 0, (IntPtr)ptr, json.Length);
            ((byte*)ptr)[json.Length] = 0; // Null terminator
        }
        return ptr;
    }

    private static string PtrToString(int ptr, int len)
    {
        return Marshal.PtrToStringAnsi((IntPtr)ptr, len);
    }
}
```

**Build:**
```bash
dotnet publish -c Release -r wasi-wasm
```

## Memory Management

### String Passing

All strings are passed as `(ptr, length)` pairs:
- `ptr`: Pointer to first byte
- `length`: Number of bytes

### Allocation

- **Host → Guest**: Host calls guest's `alloc(size)` function
- **Guest → Host**: Guest allocates and returns pointer

### Deallocation

- WASM garbage collection handles deallocation
- Or plugin can export `free(ptr)` for manual management

## Example: Hello World Plugin (Go)

```go
package main

import (
	"encoding/json"
	"unsafe"
)

//export plugin_init
func plugin_init() int32 {
	metadata := map[string]string{
		"name":    "hello",
		"version": "1.0.0",
	}
	return allocJSON(metadata)
}

//export plugin_register_slots
func plugin_register_slots() int32 {
	slots := []map[string]interface{}{
		{
			"name":        "hello.greet",
			"description": "Greet someone",
		},
	}
	return allocJSON(slots)
}

//export plugin_execute
func plugin_execute(slotNamePtr, slotNameLen, paramsPtr, paramsLen int32) int32 {
	slotName := ptrToString(slotNamePtr, slotNameLen)
	
	if slotName == "hello.greet" {
		response := map[string]interface{}{
			"success": true,
			"data": map[string]string{
				"message": "Hello from WASM!",
			},
		}
		return allocJSON(response)
	}
	
	return allocJSON(map[string]interface{}{
		"success": false,
		"error":   "unknown slot",
	})
}

//export plugin_cleanup
func plugin_cleanup() {}

//export alloc
func alloc(size int32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}

func allocJSON(v interface{}) int32 {
	data, _ := json.Marshal(v)
	ptr := alloc(int32(len(data)))
	copy(unsafe.Slice(ptr, len(data)), data)
	return int32(uintptr(unsafe.Pointer(ptr)))
}

func ptrToString(ptr, length int32) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), length)
}

func main() {}
```

**Build:**
```bash
tinygo build -o hello.wasm -target=wasi main.go
```

## Plugin Manifest

Every plugin must have a `manifest.yaml`:

```yaml
name: hello
version: 1.0.0
author: Your Name
description: Hello World plugin
license: MIT
binary: hello.wasm

permissions:
  scope:
    - read
    - write
```

## Security

### Sandboxing

- Plugins run in isolated WASM sandbox
- No direct access to host resources
- All operations go through host functions

### Permissions

Plugins must declare required permissions:
- `network`: URL patterns
- `env`: Environment variables
- `filesystem`: File paths
- `database`: Database names
- `scope`: read/write access

### Resource Limits

Host can enforce:
- Maximum memory usage
- Maximum execution time
- CPU usage limits

## Best Practices

1. **Keep plugins small** - WASM works best with focused functionality
2. **Use TinyGo** - Produces smaller WASM binaries than standard Go
3. **Minimize allocations** - Reduce GC pressure
4. **Handle errors gracefully** - Always return proper error responses
5. **Document slots** - Provide clear descriptions and examples
6. **Version carefully** - Use semantic versioning
7. **Test thoroughly** - Test with host before distribution

## Debugging

### Logging

Use `host_log()` for debugging:
```go
hostLog("debug", "Variable value: " + value)
```

### Error Handling

Always return structured errors:
```json
{
  "success": false,
  "error": "Detailed error message with context"
}
```

## Distribution

### Plugin Package Structure

```
my-plugin/
├── manifest.yaml
├── plugin.wasm
├── README.md
└── LICENSE
```

### Installation

Users install by copying to plugins directory:
```bash
cp -r my-plugin/ /path/to/zeno/plugins/
```

Or via CLI (future):
```bash
zeno plugin install my-plugin
```

## Version Compatibility

Plugins should declare minimum ZenoEngine version:

```yaml
requires:
  zeno: ">=0.3.0"
```

Host will reject incompatible plugins.
