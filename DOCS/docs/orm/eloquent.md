# Eloquent ORM

## Introduction

ZenoEngine includes an Eloquent-inspired ORM that provides a beautiful, simple ActiveRecord implementation for working with your database. Each database table has a corresponding "Model" that is used to interact with that table.

## Defining Models

To get started, let's create an Eloquent model. Models are defined using the `orm.model` slot.

```zeno
orm.model: 'users' {
    fillable: 'name,email,password'
}
```

### Mass Assignment

Notice that in the example above, we passed a `fillable` property to our model definition. The `$fillable` property serves as a "white list" of attributes that should be mass assignable.

```zeno
// 'is_admin' will be safely ignored if it is not in the fillable array
orm.model: 'users'
orm.save: $request.body
```

## Retrieving Models

Once you have created a model and its associated database table, you are ready to start retrieving data from your database.

```zeno
orm.model: 'users'
db.get { as: $users }

@foreach($users as $user)
    print: $user.name
@endforeach
```

## Relationships

Database tables are often related to one another. For example, a blog post may have many comments, or an order could be related to the user who placed it. ZenoEngine makes managing and working with these relationships easy.

### One To Many

A one-to-many relationship is used to define relationships where a single model is the parent to one or more child models.

```zeno
orm.model: 'users' {
    orm.hasMany: 'posts' {
        as: 'posts'
        foreign_key: 'user_id'
        local_key: 'id'
    }
}
```

### Many To Many

Many-to-many relations are slightly more complicated than `hasOne` and `hasMany` relationships. An example of such a relationship is a user with many roles, where the roles are also shared by other users. 

```zeno
orm.model: 'users' {
    orm.belongsToMany: 'roles' {
        as: 'roles'
        table: 'role_user'
        foreign_pivot_key: 'user_id'
        related_pivot_key: 'role_id'
    }
}
```

## Eager Loading

When accessing Eloquent relationships as properties, the relationship data is "lazy loaded". This means the relationship data is not actually loaded until you first access the property. However, ZenoEngine can "eager load" relationships at the time you query the parent model to prevent the N+1 query problem!

```zeno
// Select all users and eager load their related posts
orm.model: 'users'
db.get { as: $users }

orm.model: 'users'
orm.with: 'posts' {
    set: $users { val: $users }
}
```

ZenoEngine's ORM solves the N+1 problem identically to Laravel by executing exactly 2 queries instead of `N+1` queries.
