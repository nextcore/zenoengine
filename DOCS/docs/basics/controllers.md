# Controllers

In ZenoEngine, controller logic is written directly inside route handlers or can be abstracted into separate `.zl` files and included via `import`.

## Inline Controllers

```zeno
http.get: '/users/{id}' {
    orm.model: 'users'
    orm.find: $id { as: $user }
    view: 'users.show' {
        user: $user
    }
}
```

## CRUD Controller Pattern

```zeno
// routes/web.zl

// List
http.get: '/users' {
    orm.model: 'users'
    db.get { as: $users }
    view: 'users.index' { users: $users }
}

// Show
http.get: '/users/{id}' {
    orm.model: 'users'
    orm.find: $id { as: $user }
    view: 'users.show' { user: $user }
}

// Store
http.post: '/users' {
    validate {
        name: 'required'
        email: 'required,email,unique:users'
        password: 'required,min:8,confirmed'
    }
    orm.model: 'users'
    orm.save: $form
    redirect: '/users'
}

// Update
http.put: '/users/{id}' {
    orm.model: 'users'
    orm.find: $id { as: $user }
    orm.save: $form
    redirect: '/users/' + $id
}

// Destroy
http.delete: '/users/{id}' {
    orm.model: 'users'
    orm.find: $id { as: $user }
    orm.delete: $user
    redirect: '/users'
}
```
