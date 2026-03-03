# Inertia.js SPA Integration

ZenoEngine provides first-party support for **[Inertia.js](https://inertiajs.com)**, allowing you to build fully modern, single-page React, Vue, or Svelte applications without the complexity of building separate API endpoints or managing client-side routing.

With Inertia, you build your frontend just like a classic server-rendered app, but it is delivered via an SPA architecture that feels lightning fast.

## Setting Up Inertia

To return an Inertia response, use the `inertia.render` slot instead of `http.view` or `http.json`.

```zeno
// src/main.zl

http.get: '/dashboard' {
    do: {
        // Fetch users from DB
        db.query: 'users'
        db.get: { as: $users }
        
        // Render a Vue/React component named 'Dashboard'
        inertia.render: 'Dashboard' {
            props: {
                users: $users,
                title: "Admin Dashboard"
            }
        }
    }
}
```

When accessed via a browser directly, ZenoEngine will automatically wrap the component in the root Blade layout (`index.blade.zl` or equivalent) and inject the page data as a `data-page` JSON object.

When accessed via an Inertia XHR request (e.g., clicking an `<Link>` component on the frontend), ZenoEngine returns pure JSON containing just the requested props.

## Shared Data

Sometimes you need to share certain data (like the currently authenticated user, or global flash messages) across **all** Inertia views, without manually passing it to every `inertia.render` call.

Use the `inertia.share` slot, typically inside a global middleware:

```zeno
// src/middleware/inertia_shared.zl

http.middleware: 'inertia.shared' {
    do: {
        // Example: Share the logged-in user if available from JWT
        var: $user { val: null }
        if: $authUser != null {
            var: $user { val: { id: $authUser.id, name: $authUser.name } }
        }
        
        inertia.share: {
            auth: {
                user: $user
            },
            flash: {
                success: $session.success,
                error: $session.error
            }
        }
        
        http.next: true
    }
}
```

## Root Template

The root HTML template requires a specific setup to mount the JavaScript framework. ZenoEngine usually looks for `resources/views/app.blade.zl`.

Inside this template, use the `@inertia` and `@inertiaHead` directives.

```html
<!-- resources/views/app.blade.zl -->
<!DOCTYPE html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1.0, maximum-scale=1.0" />
    
    @inertiaHead
    
    <!-- Your compiled Vite/Webpack assets here -->
    <script type="module" src="/dist/main.js" defer></script>
  </head>
  <body>
    @inertia
  </body>
</html>
```

## Redirects

Inertia requires specialized `409 Conflict` redirects in certain lifecycle scenarios (like `PUT`/`PATCH` responses). ZenoEngine's built-in `http.redirect` slot automatically detects if the incoming request is an Inertia request (via the `X-Inertia` header) and formats the redirect response correctly.

```zeno
http.post: '/users' {
    do: {
        // Save logic...
        
        // This automatically issues a 302 or an Inertia 303/409 as needed
        http.redirect: '/users' {
            flash: { success: "User created" }
        }
    }
}
```
