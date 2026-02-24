# Multi-App Architecture

## Overview

One of ZenoEngine's most powerful features is its ability to host **multiple independent applications** within a single engine instance and binary. Rather than running separate processes or servers, all apps share one runtime.

```
One binary → Multiple apps → Virtual hosting / path-based routing
```

## How It Works

ZenoEngine uses three independent primitives for multi-app isolation:

| Primitive | What It Does |
| --- | --- |
| [`include:`](#modules-with-include) | Loads any `.zl` file, enabling free-form modular organization |
| [`view.root:`](#per-app-blade-views) | Sets the Blade view root directory for the current app |
| [`http.static:`](#per-app-static-assets) | Serves static files from any directory at any URL path |

## Entry Point

Every ZenoEngine project has exactly one global entry point:

```zeno
// src/main.zl — bootstraps all apps
http.static: 'shared/public' {
    path: '/assets'
}

include: apps/blog/routes/web.zl
include: apps/shop/routes/web.zl
include: apps/admin/main.zl
```

## Modules with `include:`

The `include:` slot loads any ZenoLang file. There are no restrictions on the path, allowing complete freedom in organizing your code.

```zeno
// You can include from anywhere
include: apps/blog/routes/web.zl
include: domain/products/routes.zl
include: features/auth/login.zl
```

## Per-App Blade Views

By default, the `view:` slot looks for templates in the `views/` directory. Use `view.root:` to override this for each app:

```zeno
// apps/blog/routes/web.zl
view.root: 'apps/blog/resources/views'

http.get: '/blog' {
    view: 'index'
    // → apps/blog/resources/views/index.blade.zl ✅
}
```

```zeno
// apps/shop/routes/web.zl
view.root: 'apps/shop/resources/views'

http.get: '/shop' {
    view: 'catalog'
    // → apps/shop/resources/views/catalog.blade.zl ✅
}
```

::: info Scope Isolation
`view.root:` is scoped to the current execution context. The blog's `view.root` does not affect the shop's views, and vice versa.
:::

`view.root:` applies to all Blade operations in its scope: `view:`, `@extends`, `@include`, `@component`, and `meta.template`.

## Per-App Static Assets

Each app can serve its own CSS, JavaScript, and images from its own public directory:

```zeno
// apps/blog/routes/web.zl
http.static: 'apps/blog/public' {
    path: '/blog/assets'
}
```

```zeno
// apps/shop/routes/web.zl
http.static: 'apps/shop/public' {
    path: '/shop/assets'
}
```

Reference them in Blade templates:

```blade
<link rel="stylesheet" href="/blog/assets/css/app.css">
```

For assets shared across all apps, serve a shared directory from `src/main.zl`:

```zeno
// src/main.zl
http.static: 'shared/public' {
    path: '/assets'
}
```

## Virtual Hosting (Multi-Domain)

Use `http.host:` to route different domains to different apps within the same engine:

```zeno
http.host: 'blog.mysite.com' { do: {
    include: apps/blog/routes/web.zl
}}

http.host: 'shop.mysite.com' { do: {
    include: apps/shop/routes/web.zl
}}
```

## Complete Example

```text
src/
└── main.zl

apps/
├── blog/
│   ├── routes/
│   │   ├── web.zl        ← view.root + http.static + routes
│   │   └── api.zl
│   ├── resources/views/  ← Blade templates
│   └── public/           ← CSS, JS, images
│
├── shop/
│   ├── routes/
│   │   ├── web.zl
│   │   └── api.zl
│   ├── resources/views/
│   └── public/
│
└── shared/
    └── public/           ← Shared fonts, icons, global CSS
```

```zeno
// src/main.zl
http.static: 'shared/public' { path: '/assets' }

include: apps/blog/routes/web.zl
include: apps/shop/routes/web.zl
```

```zeno
// apps/blog/routes/web.zl
view.root: 'apps/blog/resources/views'

http.static: 'apps/blog/public' { path: '/blog/assets' }

http.get: '/blog' {
    view: 'index'
}
http.post: '/blog/posts' {
    validate { title: 'required' }
    orm.model: 'posts'
    orm.save: $form
    redirect: '/blog'
}
```
