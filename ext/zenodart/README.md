# ZenoDart

A Dart port of the [ZenoEngine](https://github.com/zeno-engine/zeno) core.

ZenoDart allows you to execute ZenoLang scripts (`.zl`) using a Dart runtime. It achieves high feature parity with the original Go/Rust implementation, supporting asynchronous execution, HTTP networking, file I/O, and more.

## Features

- **Core Runtime**: Lexer, Parser, AST, and Async Executor.
- **Variables & Scope**: Full support for `$var`, nested scopes, and dot-notation lookup.
- **Control Flow**: `if` (with `then`/`else`), `foreach` (List & Map).
- **Expressions**: String concatenation, equality checks, list literals.
- **Modules**:
  - `http`: `http.get`
  - `json`: `json.parse`, `json.stringify`
  - `file`: `file.read`, `file.write`, `file.delete`
  - `env`: `env.get`
  - `server`: `http.server` (using `shelf`)
- **Include**: Modularize code with `include: "path/to/script.zl"`.

## Installation

1.  Navigate to the `ext/zenodart` directory.
2.  Install dependencies:
    ```bash
    dart pub get
    ```

## Usage

Run a script:

```bash
dart run bin/zenodart.dart path/to/script.zl
```

If no script is provided, it looks for `main.zl` or `src/main.zl`.

### Options

- `--help` (`-h`): Show usage information.
- `--version` (`-v`): Show version.
- `--debug` (`-d`): Print the AST structure before execution.

## Example

```yaml
http.server: {
    port: 8080
    routes: {
        get: "/hello" {
            return: "Hello from ZenoDart!"
        }
    }
}
```
