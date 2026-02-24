# Requests

## Accessing the Request

To access the current HTTP request, you have access to `$request` inside any route handler. The request object contains the method, URL, headers, query params, form data, and JSON body.

```zeno
http.get: '/users' {
    // $request.method = "GET"
    // $request.url = full URL string
    // $request.path = "/users"
    // $request.query = map of query string params
    print: $request.method
}
```

## Request Input

### Query String

```zeno
http.get: '/search' {
    // Access ?q=hello via $request.query
    set: $term { val: $request.query.q }
}
```

### Form Data (POST)

```zeno
http.post: '/login' {
    // $form contains submitted form fields
    set: $email { val: $form.email }
    set: $password { val: $form.password }
}
```

### JSON Body

```zeno
http.post: '/api/users' {
    // $request.body contains the parsed JSON
    set: $name { val: $request.body.name }
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
