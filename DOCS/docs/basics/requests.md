# Requests

## Accessing the Request

To access the current HTTP request, you have access to `$request` inside any route handler. The request object contains the method, URL, headers, query params, form data, and JSON body.

```zeno
http.get: '/users' {
    // $request.method = "GET"
    // $request.url = full URL string
    // $request.path = "/users"
    // $request.query = map of query string params
    log: $request.method
}
```

## Request Input

### Query String

```zeno
http.get: '/search' {
    // Access ?q=hello via $request.query
    var: $term { val: $request.query.q }
}
```

### Capturing All Form Data

If you want to capture the entire form data as a map (object), use `http.form` without a specific key.

```zeno
http.post: '/register' {
    do: {
        http.form: { as: $formData }
        
        db.table: "users"
        db.insert: {
            name: $formData.name,
            email: $formData.email,
            password: $formData.password
        }
    }
}
```

### Form Data (POST)

```zeno
http.post: '/login' {
    // $form contains submitted form fields
    var: $email { val: $form.email }
    var: $password { val: $form.password }
}
```

### JSON Body

```zeno
http.post: '/api/users' {
    // $request.body contains the parsed JSON
    var: $name { val: $request.body.name }
}
```

## Route Parameters

Route parameters (e.g., `{id}`) are automatically injected as scope variables:

```zeno
http.get: '/users/{id}' {
    // $id is automatically available
    orm.find: $id { as: $user }
}
```
