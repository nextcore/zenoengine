# ZenoRust Migration Guide

This guide details the differences between the original **ZenoEngine (Go)** and the new **ZenoRust**. While ZenoRust aims for high feature parity, there are syntactical changes and architectural improvements you should be aware of when migrating scripts.

## 1. Syntax & Built-in Functions

The most significant change is the shift from "Dot Notation" (e.g., `http.get`) to **Snake Case** (e.g., `http_get`) for standard library functions. This flat namespace improves parsing performance and simplicity.

| Category | ZenoGo (Original) | ZenoRust (New) | Notes |
| :--- | :--- | :--- | :--- |
| **String** | `len(s)` | `len(s)` | Same |
| | `upper(s)` | `upper(s)` | Same |
| | `str(v)` | `str(v)` | Same |
| | `str.replace(s, a, b)` | `str_replace(s, a, b)` | |
| | `str.concat(...)` | `str_concat(...)` | |
| **Array** | `push(arr, v)` | `push(arr, v)` | Same |
| | `arr.join(arr, sep)` | `arr_join(arr, sep)` | |
| | `arr.pop(arr)` | `arr_pop(arr)` | |
| **Map** | `map.keys(m)` | `map_keys(m)` | |
| **HTTP** | `http.get(url)` | `http_get(url)` | |
| | `http.post(url, body)` | `http_post(url, body)` | |
| | `fetch(url, opts)` | `fetch(url, opts)` | Same |
| | `response.json(obj)` | `response_json(obj)` | |
| | `response.status(code)` | `response_status(code)` | |
| **JSON** | `json.parse(s)` | `json_parse(s)` | |
| | `json.stringify(v)` | `json_stringify(v)` | |
| **File I/O** | `file.read(path)` | `file_read(path)` | |
| | `file.write(path, data)`| `file_write(path, data)`| Protected in Prod |
| | `file.delete(path)` | `file_delete(path)` | |
| | `dir.create(path)` | `dir_create(path)` | |
| **Time** | `time.now()` | `time_now()` | Returns ISO 8601 |
| | `time.format(t, fmt)` | `time_format(t, fmt)` | |
| **Utils** | `env.get(key)` | `env_get(key)` | |
| | `coalesce(...)` | `coalesce(...)` | Same |
| | `include(path)` | `include(path)` | Same |
| **Crypto** | `hash.sha256(s)` | `hash_sha256(s)` | |
| | `uuid.v4()` | `uuid_v4()` | |
| | `base64.encode(s)` | `base64_encode(s)` | |
| **Regex** | `regex.match(p, s)` | `regex_match(p, s)` | |

## 2. Database (SQL & Query Builder)

ZenoRust introduces a functional Query Builder API that is **database agnostic**, resolving placeholders (`?` vs `$1`) automatically.

### Raw SQL
*   **Go:** `db.query("SELECT * FROM users WHERE id = ?", 1)`
*   **Rust:** `db_query("SELECT * FROM users WHERE id = ?", [1])` (Note: params passed as array)

### Query Builder
*   **Go:** `db.table("users").where("id", 1).get()`
*   **Rust:**
    ```javascript
    let q = db_table("users");
    qb_where(q, "id", "=", 1);
    let users = qb_get(q);
    ```
    *Note: ZenoRust uses functional style `qb_*(builder)` instead of method chaining `builder.*()`.*

## 3. Metaprogramming & Reflection

ZenoRust adds powerful metaprogramming capabilities not fully present in the original engine.

*   **`eval(code_string)`**: Executes a string of ZenoLang code in the *current scope*.
*   **`call_function(name, args...)`**: Calls a function dynamically by name.

```javascript
let func_name = "my_function";
call_function(func_name, arg1, arg2);
```

## 4. Plugins (WASM & Sidecar)

The plugin system is fully implemented with 100% parity.

*   **Sidecar:** JSON-RPC over Stdin/Stdout.
*   **WASM:** Supports loading `.wasm` files.
    *   **Advanced:** Includes a "String ABI" that automatically handles memory allocation to pass Strings/JSON to the WASM guest and read results back.

## 5. Templating (ZenoBlade)

The **ZenoBlade** engine is 100% compatible with the original specification.
*   `{{ variable }}` for interpolation.
*   `@if`, `@else`, `@foreach` for control flow.
*   `@include('path')` for partials.
*   `@extends('layout')`, `@section('name')`, `@yield('name')` for inheritance.

## 6. Known Limitations

*   **Strict Types in DB:** When using SQLite, boolean columns must be `INTEGER` (0/1) for correct mapping.
*   **Method Chaining:** ZenoRust does not currently support `object.method()` syntax for arbitrary objects; use functional equivalents (e.g. `qb_where(q)` instead of `q.where()`).
