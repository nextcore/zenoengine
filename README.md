<div align="center">

# ⚡ ZenoEngine

[![Go Version](https://img.shields.io/github/go-mod/go-version/nextcore/zenoengine?style=flat-square&color=00ADD8)](https://golang.org)
[![Latest Release](https://img.shields.io/github/v/release/nextcore/zenoengine?style=flat-square&color=6366F1)](https://github.com/nextcore/zenoengine/releases)
[![License](https://img.shields.io/github/license/nextcore/zenoengine?style=flat-square&color=475569)](LICENSE)

**The Developer Experience of Laravel. The Speed and Simplicity of Go.**

*ZenoEngine is a lightning-fast, production-ready execution engine for the **ZenoLang** programming language. It compiles your entire full-stack application into a single, high-performance binary.*

[📖 Read Documentation](DOCS/docs/index.md) &nbsp;•&nbsp; [🚀 Quick Start](#-quick-start) &nbsp;•&nbsp; [🤝 Contribute](ZENOLANG_STYLE_GUIDE.md)

</div>

---

## 🔥 Why ZenoEngine?

Traditional fullstack development requires configuring complex web servers, process managers, and language runtimes (Nginx, PHP-FPM, Node/PM2, Docker). **ZenoEngine changes everything.**

* **🚀 Compiled Go Speed**: Built on top of Go (`go 1.26+`), your routes, ORM database queries, and template rendering execute in microseconds with minimal RAM footprint.
* **🎨 Laravel-Parity DX**: Write clean, expressive logic using **ZenoLang** alongside a 1-to-1 port of the **Blade templating engine** (`@if`, `@foreach`, `@extends`, `$loop`, components, and more).
* **📦 Single-Binary Deployment**: Compile your entire application—including routes, views, database migrations, and assets—into a single executable binary. Just copy-paste and run.
* **🛡️ Secure by Default**: Built-in read-only database query guards (`SELECT`/`WITH`/`SHOW`/`EXPLAIN`), security obfuscation middleware (`middleware.spoof`), CSRF protection, mass-assignment guards, and automatic JWT handling.
* **🔄 Hot Reload**: Modify your templates and logic files and see changes instantly without restarting the server or breaking development flow.

---

## 💻 Code at a Glance

### 1. The Logic (`src/main.zl`)
Write clean, readable, brace-based backend logic for your routing and database queries:

```zeno
// Define a route and execute safe database query
http.get: '/users' {
    db.select: 'SELECT id, name, email FROM users WHERE is_active = 1' {
        as: $users
    }
    
    return: view: 'users.index' { users: $users }
}
```

### 2. The View (`views/users/index.blade.zl`)
Utilize a familiar, powerful Laravel-like template engine directly in your web views:

```html
@extends('layouts.app')

@section('content')
    <div class="container">
        <h1>Active Users</h1>
        <ul class="user-list">
            @foreach($users as $user)
                <li>
                    <strong>{{ $user.name }}</strong> ({{ $user.email }})
                    @if($loop.first) <span class="badge">Newest</span> @endif
                </li>
            @endforeach
        </ul>
    </div>
@endsection
```

---

## ✨ Features Out of the Box

* **Eloquent-inspired ORM & Schema Builder**: Relationships (`hasOne`, `hasMany`, `belongsTo`, `belongsToMany`), eager loading, and mass-assignment protection.
* **Safe Standard Database Slots (`db.select` & `db.execute`)**: Strict read-only query enforcement on `db.select`, variable prefix trimming (`as: $var`), and non-empty query validations.
* **Security Obfuscation (`middleware.spoof`)**: Masks server headers (`X-Powered-By: PHP/8.3.0`) and injects dummy session cookies to trick automated vulnerability scanners.
* **API Protection (`middleware.api_key`)**: Built-in verification for `X-API-KEY` headers, query parameters, or Bearer tokens.
* **Image Upload & Conversion (`upload.webp`)**: Auto-converts uploaded images to WebP format and cleans up previous file uploads.
* **HTML Sanitizer (`sanitize`)**: XSS prevention for user-generated HTML content.
* **Captcha Engine (`captcha.id` & `captcha.verify`)**: Native Captcha generator and verification slots.
* **Inertia.js & SPA Support**: Complete official Inertia.js protocol implementation (`inertia.render`, `inertia.share`, `inertia.location`) and SPA static hosting fallback.
* **Automated API Documentation**: Automatically generates interactive Swagger/OpenAPI documentation from route definitions.

---

## 🚀 Quick Start

### 1. Download the CLI
Download the latest executable binary from the Releases page for your operating system.

On Linux/macOS:
```bash
chmod +x zeno
mv zeno /usr/local/bin/
```

### 2. Run the Server
Start your application by passing the main ZenoLang entry point file directly to the binary:
```bash
zeno src/main.zl
```

Your app is now running at `http://localhost:3000` (or the port specified in your `.env` file)!

---

## 📚 Ecosystem & Libraries

* **Core Runtime**: [github.com/nextcore/zeno-go](https://github.com/nextcore/zeno-go)
* **Web Core Library**: [github.com/nextcore/zenoweb-core](https://github.com/nextcore/zenoweb-core)
* **VS Code Extension**: [vscode-zenolang](vscode-zenolang/)

---

## 🔧 Modular Embedding in Your Go App (Advanced)

If you are building a custom hybrid application in Go, you can selectively import only the modules you need using the Functional Options pattern:

```go
import (
    "github.com/nextcore/zeno-go/pkg/engine"
    "github.com/nextcore/zenoengine/pkg/app"
    "github.com/go-chi/chi/v5"
)

func main() {
    eng := engine.NewEngine()
    r := chi.NewRouter()
    
    // Option A: Register all slots
    app.RegisterAllSlots(eng, r, dbMgr, queue, nil)

    // Option B: Selective / Modular registration
    app.RegisterSlots(eng,
        app.WithCore(),
        app.WithWeb(r),
        app.WithDB(dbMgr),
    )
}
```

---

## 📜 License

ZenoEngine is open-source software licensed under the [Apache 2.0 License](LICENSE).
