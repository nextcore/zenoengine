# ORM Relationships

## Introduction

Database tables are often related to one another. ZenoEngine makes managing and working with these relationships easy, supporting the four relationship types you know from Laravel Eloquent.

## One To One (hasOne)

A one-to-one relationship is a very basic type of database relationship. Define it using `orm.hasOne` inside your model block:

```zeno
orm.model: 'users' {
    orm.hasOne: 'profiles' {
        as: 'profile'
        foreign_key: 'user_id'
        local_key: 'id'
    }
}
```

Then eager load it:

```zeno
orm.model: 'users'
db.get { as: $users }

orm.model: 'users'
orm.with: 'profile' {
    set: $users { val: $users }
}

// Each user now has a $user.profile object attached
```

## One To Many (hasMany)

One-to-many relationships are used when a single parent model owns many child models. For example, a user may have many posts:

```zeno
orm.model: 'users' {
    orm.hasMany: 'posts' {
        as: 'posts'
        foreign_key: 'user_id'
        local_key: 'id'
    }
}
```

## Many To One (belongsTo)

The inverse of `hasMany` is `belongsTo`. A post belongs to a user:

```zeno
orm.model: 'posts' {
    orm.belongsTo: 'users' {
        as: 'author'
        foreign_key: 'user_id'
    }
}
```

## Many To Many (belongsToMany)

Many-to-many relations involve an intermediary "pivot" table. For example, a User may have many Roles, and Roles may be shared by many Users:

```zeno
orm.model: 'users' {
    orm.belongsToMany: 'roles' {
        as: 'roles'
        table: 'role_user'          // Pivot table name
        foreign_pivot_key: 'user_id'
        related_pivot_key: 'role_id'
    }
}
```

Then eager load:

```zeno
orm.model: 'users'
db.get { as: $users }

orm.model: 'users'
orm.with: 'roles' {
    set: $users { val: $users }
}

// Each user now has $user.roles = [...roles array...]
```

::: tip N+1 Prevention
ZenoEngine's eager loading always resolves any relationship in exactly **2 SQL queries** total, regardless of how many parent records you have loaded — identical to Laravel's `with()` behavior.
:::
