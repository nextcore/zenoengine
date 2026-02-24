# Middleware

## Introduction

Middleware provide a convenient mechanism for inspecting and filtering HTTP requests entering your application. All middleware are resolved out of the ZenoEngine service container.

## Applying Middleware to Routes

You may assign middleware to individual routes using the `middleware` attribute:

```zeno
http.get: '/profile' {
    middleware: 'auth'
    do: {
        // Only authenticated users...
    }
}
```

## Middleware Groups

Sometimes you may want to group several middleware together to make them easier to assign to routes. You can accomplish this using route groups:

```zeno
http.group: '/admin' {
    middleware: 'auth'
    do: {
        http.get: '/dashboard' { ... }
        http.get: '/settings' { ... }
    }
}
```

## Available Middleware

| Middleware | Description |
| --- | --- |
| `auth` | Requires a valid JWT token. Injects `$auth` into scope. |

### Accessing Auth Data

When the `auth` middleware is applied, authenticated user data is available via `$auth`:

```zeno
http.get: '/profile' {
    middleware: 'auth'
    do: {
        // $auth.id, $auth.email, $auth.role are available
        return: $auth.email
    }
}
```
