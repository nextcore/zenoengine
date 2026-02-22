# Zeno Go - The Scriptable Systems Framework

Zeno is a next-generation web framework that combines the **Developer Experience (DX) of Laravel** with the **Performance of Go**. It introduces a new paradigm: **Scriptable Systems**.

You write high-level, expressive business logic in **ZenoLang (.zl)** — a secure, dynamic scripting language — while the engine runs on a robust, multi-threaded Go runtime. For production, Zeno can **AOT Compile** your scripts into a single native binary.

![Zeno Engine](https://placehold.co/800x200?text=Zeno+Engine+v1.0)

## 🚀 Key Features

### 1. Laravel-like Developer Experience
Zeno provides a familiar home for PHP/Laravel developers, but with modern twists.

*   **Fluent Query Builder:** Write database queries that look like config but act like code.
    ```yaml
    db.table: "users" {
      where: { active: true }
      order_by: "created_at DESC"
      limit: 10
      get: $users
    }
    ```
*   **Blade Components:** Powerful frontend templating with `<x-tag>` syntax and `$attributes` bag.
    ```html
    <x-card title="Dashboard">
        Welcome, {{ $user.name }}!
    </x-card>
    ```
*   **Artisan-style CLI:**
    *   `zeno make:model User` - Scaffold migrations and schemas.
    *   `zeno route:list` - View all registered routes.
    *   `zeno migrate` - Run database migrations.
*   **Middleware Groups:** Secure your routes easily.
    ```yaml
    http.group: "/admin" {
      middleware: ["auth", "admin"]
      do: { ... }
    }
    ```

### 2. Enterprise Performance
*   **AOT Compiler (`zeno build`):** Compiles your `.zl` scripts into a **native Go binary**. Zero parsing overhead in production.
*   **WASM Plugins:** Extend Zeno with Rust, C++, or Go (TinyGo) via WebAssembly.
    *   `zeno plugin:new my-plugin --lang=rust`
*   **High Concurrency:** Built on Go's `net/http` and `chi` router. Non-blocking I/O by default.

### 3. World-Class Tooling
*   **Zeno Studio (LSP):** Context-aware autocomplete for VS Code.
    *   Try `zeno lsp:completion` to see the magic.
*   **Self-Documenting:**
    *   `zeno docs` generates API references automatically from the engine source.
*   **Testing Framework:**
    *   `zeno test` runs isolated test scripts with assertions like `assert.eq`.

## 📦 Installation

```bash
# Clone and Build (Prototype)
git clone https://github.com/zeno-lang/zeno-go
cd zeno-go
go build -o zeno cmd/zeno/zeno.go
```

## 🐳 Deployment (Docker)

Zeno is production-ready with a built-in Dockerfile.

```bash
# Build Image
docker build -t zeno-app .

# Run Container
docker run -p 3000:3000 -e APP_ENV=production zeno-app
```

## 🏁 Getting Started

1.  **Create a new project:**
    (Currently manual, or use `zeno-docs` as template)

2.  **Run the server:**
    ```bash
    ./zeno
    ```

3.  **Explore the Documentation Site:**
    The project includes a self-hosted docs site. Run it to learn more!
    ```bash
    ./zeno run zeno-docs/main.zl
    ```

## 🛠️ CLI Reference

| Command | Description |
| :--- | :--- |
| `zeno run <file>` | Execute a Zeno script. |
| `zeno build <file>` | Compile script to native binary (AOT). |
| `zeno docs` | Generate API documentation (JSON/Markdown). |
| `zeno route:list` | List all registered HTTP routes. |
| `zeno make:model <Name>` | Scaffold a new database migration. |
| `zeno migrate` | Run pending migrations. |
| `zeno test` | Run tests in `tests/` directory. |
| `zeno install <url>` | Install a package from Git. |
| `zeno plugin:new` | Scaffold a WASM plugin. |

## 🗺️ Roadmap & Vision

See [ROADMAP.md](ROADMAP.md) for the detailed 3-Phase Strategy (Foundation, Expansion, Domination).
See [VISION.md](VISION.md) for the philosophy behind "The Next Laravel".

## 📄 License

MIT License.
