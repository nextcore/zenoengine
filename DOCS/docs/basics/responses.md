# Responses

## Creating Responses

### JSON Responses

ZenoEngine automatically serializes any data returned as a `json` response. This is perfect for building APIs:

```zeno
http.get: '/api/users' {
    orm.model: 'users'
    db.get { as: $users }
    json: $users
}
```

### Redirect Responses

```zeno
http.post: '/users' {
    orm.model: 'users'
    orm.save: $form
    redirect: '/users'
}
```

### View Responses

Render a Blade template as an HTML response:

```zeno
http.get: '/welcome' {
    view: 'welcome' {
        name: 'World'
        title: 'Welcome'
    }
}
```

### Plain Text / HTML

```zeno
http.get: '/ping' {
    return: 'pong'
}
```

## Setting HTTP Status Codes

```zeno
http.post: '/api/users' {
    // ... create user ...
    json: $user {
        status: 201
    }
}
```

## Response Headers

```zeno
http.get: '/api/data' {
    header: 'Content-Type' { val: 'application/json' }
    header: 'X-Custom-Header' { val: 'value' }
    json: $data
}
```
