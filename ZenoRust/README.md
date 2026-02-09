# ZenoRust

ZenoRust is a high-performance, async implementation of the ZenoEngine execution runtime, written in Rust (2024 Edition). It is designed to be a drop-in replacement for the core backend engine, offering faster execution, smaller binaries, and memory safety.

## Features

- **Async Core:** Fully asynchronous runtime built on `Tokio` and `Axum`.
- **Web Server:** Integrated high-performance HTTP server with dynamic routing.
- **Database:** Support for SQLite, MySQL, and PostgreSQL via `SQLx` (Raw SQL).
- **ZenoBlade:** Full-featured templating engine compatible with Laravel Blade (`@if`, `@foreach`, `@extends`, `@include`).
- **Standard Library:** Comprehensive set of built-in functions for File I/O, JSON, Regex, Crypto, and more.
- **Middleware:** Built-in Security Headers, IP Blocking, and CORS support.
- **Modularity:** Support for `include()` to split scripts into modules.

## Getting Started

### Prerequisites

- Rust (Latest Stable or Nightly for 2024 edition features)
- SQLite (optional, strictly speaking, as it can run in-memory or file mode)

### Installation

1.  Navigate to the `ZenoRust` directory.
2.  Install dependencies:
    ```bash
    cargo build
    ```

### Usage

**Run in CLI Mode:**
Execute a ZenoLang script directly:
```bash
cargo run -- source/your_script.zl
```

**Run in Server Mode:**
Start the web server (defaults to port 3000):
```bash
cargo run -- server
```
The server will route requests to `source/index.zl` by default.

### Configuration

Create a `.env` file in the root:
```env
APP_ENV=development
DATABASE_URL=sqlite://zeno.db?mode=rwc
PORT=3000
BLOCKED_IPS=10.0.0.1,192.168.1.50
```

## Language Reference

### Syntax
ZenoLang in ZenoRust supports:
- **Variables:** `let x = 10;`
- **Functions:** `fn add(a, b) { return a + b; }`
- **Control Flow:** `if (x > 5) { ... } else { ... }`
- **Anonymous Functions:** `router_get("/path", fn() { ... });`

### Built-in Functions

#### Web & Routing
- `router_get(path, handler)`: Register GET route.
- `router_post(path, handler)`: Register POST route.
- `request`: Global map containing `method`, `path`, `query`, `body`.
- `response_json(obj)`: Send JSON response.
- `response_status(code)`: Set HTTP status code.
- `fetch(url, options)`: Make HTTP requests.

#### Database
- `db_query(sql, params)`: Execute SQL query and return rows.
- `db_execute(sql, params)`: Execute SQL statement (INSERT/UPDATE/DELETE).

#### Templating
- `view(template_path, data)`: Render a ZenoBlade template.
- Directives: `{{ $val }}`, `@if`, `@foreach`, `@extends`, `@section`, `@yield`, `@include`.

#### Utilities
- `json_parse(str)`, `json_stringify(obj)`
- `file_read(path)`, `file_write(path, content)`, `file_delete(path)`
- `base64_encode(str)`, `hex_encode(str)`
- `hash_sha256(str)`, `uuid_v4()`, `random_int(min, max)`
- `password_hash(str, cost?)`, `password_verify(str, hash)`
- `is_email(str)`, `is_numeric(str)`
- `str_concat`, `str_replace`, `upper`, `len`
- `env_get(key)`

## Security

- **File Write Protection:** In production (`APP_ENV!=development`), writing to `.zl`, `.env`, or `.go` files is blocked.
- **SQL Injection:** Always use parameterized queries (`?` or `$1`) with `db_query`.
- **Middleware:** `BLOCKED_IPS` env var can block malicious actors. `Security Headers` are set by default.

## Feature Status
See [PROGRESS.md](PROGRESS.md) for a detailed breakdown.

## Migration Guide
If you are coming from the Go version of ZenoEngine, please read [MIGRATION.md](MIGRATION.md) for a syntax comparison and porting guide.
