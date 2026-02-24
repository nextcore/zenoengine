# Routing

ZenoEngine's routing system is designed to feel identical to Laravel, with the same expressive syntax and powerful features.

## Basic Routing

The most basic ZenoEngine routes accept a URI and a closure block, providing a very simple and expressive method of defining routes.

```zeno
http.get: '/welcome' {
    return: 'Hello World'
}
```

### Available Router Methods

The router allows you to register routes that respond to any HTTP verb:

```zeno
http.get: '/users' { ... }
http.post: '/users' { ... }
http.put: '/users/{id}' { ... }
http.patch: '/users/{id}' { ... }
http.delete: '/users/{id}' { ... }
```

## Route Parameters

Sometimes you will need to capture segments of the URI within your route. For example, you may need to capture a user's ID from the URL. You may do so by defining route parameters:

```zeno
http.get: '/users/{id}' {
    return: 'User ' + $id
}
```

Parameters are natively injected into the ZenoLang scope, so you can access `{id}` simply as `$id`.

## Route Groups

Route groups allow you to share route attributes, such as middleware, across a large number of routes without needing to define those attributes on each individual route.

```zeno
http.group: '/admin' {
    middleware: 'auth'
    do: {
        http.get: '/dashboard' { 
            // Accessible at /admin/dashboard
        }
    }
}
```

## Subdomain Routing

Route groups may also be used to handle subdomain routing.

```zeno
http.host: 'api.myapp.com'
do: {
    http.get: '/users' {
        // ...
    }
}
```
