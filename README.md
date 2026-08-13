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

* **🚀 Compiled Go Speed**: Built on top of Go, your routes, ORM database queries, and template rendering execute in microseconds with minimal RAM footprint.
* **🎨 Laravel-Parity DX**: Write clean, expressive logic using **ZenoLang** alongside a 1-to-1 port of the **Blade templating engine** (`@if`, `@foreach`, `@extends`, `$loop`, components, and more).
* **📦 Single-Binary Deployment**: Compile your entire application—including routes, views, database migrations, and assets—into a single executable binary. Just copy-paste and run.
* **🛡️ Secure by Default**: Built-in protection against common security risks, including CSRF protection, mass-assignment guards, and automatic cryptographically secure JWT configuration.
* **🔄 Hot Reload**: Modify your templates and logic files and see changes instantly without restarting the server or breaking development flow.

---

## 💻 Code at a Glance

### 1. The Logic (`src/main.zl`)
Write clean, readable, brace-based backend logic for your routing and database queries:

```zeno
// Define a route and query database
http.get: '/users' {
    $users: db.query: 'users' {
        select: 'id, name, email'
        where: 'is_active = ?' { val: true }
        limit: 10
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

* **Eloquent-inspired ORM**: Relationships (`hasOne`, `hasMany`, `belongsTo`, `belongsToMany`), eager loading (resolving N+1 issues in exactly 2 queries), and mass-assignment protection.
* **Robust Routing**: Clean URL routing, route grouping, and middleware support out-of-the-box.
* **Integrated Mail & Storage**: Configurable SMTP/mock mailers and standard file storage slots (`put`, `delete`, `exists`).
* **Automated API Documentation**: Automatically generate interactive Swagger/OpenAPI documentation from route definitions and comments.

---

## 🚀 Quick Start

### 1. Download the CLI
Download the latest executable binary from the [Releases](https://github.com/nextcore/zenoengine/releases) page for your operating system.

On Linux/macOS:
```bash
chmod +x zeno
mv zeno /usr/local/bin/
```

### 2. Create a New Project
Scaffold a fully configured MVC or Modular starting boilerplate:
```bash
zeno new my-app
```

### 3. Run the Development Server
Navigate to the directory and start development with automatic hot reloading:
```bash
cd my-app
zeno run src/main.zl
```

Your app is now running at `http://localhost:8080`!

---

## 📚 Ecosystem & Libraries

* **Core Runtime**: [github.com/nextcore/zeno-go](https://github.com/nextcore/zeno-go)
* **VS Code Extension**: [vscode-zenolang](vscode-zenolang/)

---

## 📜 License

ZenoEngine is open-source software licensed under the [Apache 2.0 License](LICENSE).
