# Directory Structure

## Introduction

ZenoEngine is **"structure-agnostic"** — it does not enforce any particular directory layout. Instead, it provides a flexible set of primitives (`include:`, `view.root:`, `http.static:`) that allow you to build your project in any style you prefer.

## The One Rule

The only thing ZenoEngine requires is a single entry point:

```text
src/main.zl        ← Always the starting point
```

Everything else is completely up to you.

## Structure Styles

### 1. Laravel Style (Recommended for single apps)

Familiar to any Laravel developer. Great for teams migrating from PHP.

```text
src/main.zl
routes/
├── web.zl          # Web routes (HTML responses)
└── api.zl          # API routes (JSON responses)
resources/
└── views/          # Blade template files
database/
├── migrations/
└── seeders/
public/             # CSS, JS, images
```

```zeno
// src/main.zl
include: routes/web.zl
include: routes/api.zl
```

---

### 2. Multi-App / Modular (Recommended for hosting multiple apps)

ZenoEngine can host **multiple independent applications** in a single instance. Each app lives in its own directory and can use any internal style.

```text
src/
└── main.zl              ← Loads all apps

apps/
├── blog/                ← App 1
│   ├── routes/
│   │   ├── web.zl
│   │   └── api.zl
│   ├── resources/
│   │   └── views/       ← Blog's own Blade templates
│   └── public/          ← Blog's own CSS/JS
│
├── shop/                ← App 2
│   ├── routes/
│   │   ├── web.zl
│   │   └── api.zl
│   ├── resources/
│   │   └── views/       ← Shop's own Blade templates
│   └── public/          ← Shop's own CSS/JS
│
└── shared/
    └── public/          ← Shared assets (fonts, icons, etc.)
```

```zeno
// src/main.zl
http.static: 'shared/public' {
    path: '/assets'
}

include: apps/blog/routes/web.zl
include: apps/shop/routes/web.zl
```

```zeno
// apps/blog/routes/web.zl
view.root: 'apps/blog/resources/views'   // ← isolate views per app

http.static: 'apps/blog/public' {
    path: '/blog/assets'
}

http.get: '/blog' {
    view: 'index'   // → apps/blog/resources/views/index.blade.zl
}
```

---

### 3. Domain-Driven Design (DDD)

```text
src/main.zl
domain/
├── users/
│   ├── routes.zl
│   └── views/
├── products/
│   ├── routes.zl
│   └── views/
└── orders/
    ├── routes.zl
    └── views/
```

---

### 4. Feature-Based (Angular / Next.js-style)

```text
src/main.zl
features/
├── auth/
│   ├── login.zl
│   ├── register.zl
│   └── views/
├── dashboard/
│   ├── index.zl
│   └── views/
└── reports/
    ├── index.zl
    └── views/
```

---

### 5. Micro-App (One file per concern)

```text
src/main.zl
apps/
├── landing.zl          ← entire landing page in one file
├── api-v1.zl           ← all v1 API endpoints
└── webhooks.zl         ← webhook handlers
```

---

### 6. Mixed Styles (The real world)

Different apps inside the same engine can each use a different style. There is no conflict.

```zeno
// src/main.zl

// App A: Laravel-style (team of PHP devs)
include: apps/blog/routes/web.zl

// App B: DDD-style (Go team)
include: domain/shop/routes.zl

// App C: Single file
include: apps/landing.zl

// Shared virtual host
http.host: 'api.mysite.com' { do: {
    include: apps/api/routes.zl
}}
```

## The Three Isolation Primitives

| Primitive | Purpose | How |
| --- | --- | --- |
| `include:` | Load any `.zl` file from any path | Free-form, no restrictions |
| `view.root:` | Set Blade view directory per app | Declared at top of each app's route file |
| `http.static:{path:}` | Serve static files at a unique URL path | Each app specifies its own `root` and `path` |

## Multi-App Isolation Summary

```
src/main.zl
└── include: apps/blog/routes/web.zl   ← blog views → apps/blog/resources/views/
    include: apps/shop/routes/web.zl   ← shop views → apps/shop/resources/views/
    include: apps/admin/main.zl        ← admin views → apps/admin/views/
```

Each app is **completely independent** — its own routes, views, and static assets. Yet they all run inside **one process, one binary**.
