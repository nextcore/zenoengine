# ZenoLang Style Guide

This guide outlines the recommended coding standards and conventions for development with **ZenoLang** and **ZenoEngine**.

## 📁 File Conventions

- **Logic Files**: Use the `.zl` extension for all ZenoLang logic files (e.g., `src/main.zl`, `src/routes.zl`).
- **Template Files**: Use the `.blade.zl` extension for Blade templating files (e.g., `views/welcome.blade.zl`).
- **Configuration**: Sensitive configuration should be kept in a `.env` file and accessed via the `os.env` or `config` slots.

## ✍️ Syntax & Structure

### Braces and Colon
ZenoLang uses a clean, brace-based syntax for blocks. Slots typically use a colon `:` to accept a primary value.

```zeno
// Good
http.get: '/hello' {
    return: 'World'
}

// Avoid (unless no value is needed)
http.get {
    path: '/hello'
    return: 'World'
}
```

### Variable Naming & Assignment
Variables in ZenoLang are always prefixed with a dollar sign `$`. Use the `var` slot for assignments.

```zeno
// Preferred (Modern)
var: $name { val: 'Zeno' }

// Legacy (Supported)
scope.set: $name { val: 'Zeno' }

// Concise shorthand (Best for simple values)
$count: 10
$is_active: true
```

Use `snake_case` for variable names and slot attributes.

## 🍪 Sessions & State

- Use `session.set` and `session.get` for persistent user state.
- Always use `session.flash` for one-time messages (e.g. form feedback).
- Use `session.regenerate` after successful user authentication.

```zeno
session.set: "user_role" { val: "admin" }
session.flash: "message" { val: "Welcome back!" }
```

## 🛣️ Routing Conventions

- Keep routes organized in `src/main.zl` or separate files included via `load`.
- Use `http.group` for shared middleware or path prefixes.
- Use `http.host` for multi-tenant/subdomain routing.

```zeno
http.group: '/api/v1' {
    middleware: 'auth'
    do: {
        http.get: '/profile' { ... }
    }
}
```

## 🗄️ Database & ORM

- Use `orm.model` for Active Record style interactions.
- Define `fillable` properties to prevent mass assignment vulnerabilities.
- Prefer `db.query` for complex fluent queries.

```zeno
orm.model: 'users' {
    fillable: 'name,email,password'
    
    orm.hasMany: 'Post' { as: 'posts' }
}
```

## 🎨 View Layer

- Use `{{ $var }}` for escaped output and `{!! $var !!}` for raw HTML.
- Use `@csrf` in every form that makes a `POST`, `PUT`, `PATCH`, or `DELETE` request.
- Leverage `@extends` and `@section` for layout management.

## 🛠️ Metaprogramming

- Use `meta.eval` sparingly for dynamic execution.
- Use `meta.template` for powerful code generation tasks.

---

*This guide is a living document. Follow these conventions to ensure your ZenoEngine applications are maintainable and secure.*
