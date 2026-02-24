# Mutators & Attribute Casting

## Introduction

In ZenoEngine, you can transform how values are stored and retrieved from models. This is similar to Laravel Eloquent's Mutators and Casting.

## Casting Attributes

You can define how specific fields should be treated using the `cast` directive inside `orm.model`:

```zeno
orm.model: 'users' {
    fillable: 'name,email,settings'
    cast: 'settings' { type: 'json' }
    cast: 'is_admin' { type: 'boolean' }
    cast: 'balance' { type: 'float' }
}
```

After casting, values are automatically converted when retrieved:

```zeno
orm.model: 'users'
orm.find: 1 { as: $user }

// $user.is_admin is a boolean (true/false), not "1"/"0"
// $user.settings is a decoded JSON map
// $user.balance is a float
```

## Supported Cast Types

| Type | Description |
| --- | --- |
| `string` | Cast to string |
| `integer` | Cast to integer |
| `float` | Cast to floating point number |
| `boolean` | Cast to true/false |
| `json` | Automatically encode/decode as JSON |
| `array` | Alias for json |
| `datetime` | Cast to a formatted date/time string |

## Hidden Attributes

You can hide sensitive attributes from being returned in JSON responses using the `hidden` directive:

```zeno
orm.model: 'users' {
    fillable: 'name,email'
    hidden: 'password,remember_token'
}
```

When `$user` is serialized to JSON (e.g. via `json: $user`), the `password` field will be automatically excluded from the response.

::: tip Laravel Parity
This feature mirrors Eloquent's `$hidden` array property on models, providing the same protection against accidentally exposing sensitive data in API responses.
:::
