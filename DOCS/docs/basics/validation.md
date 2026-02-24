# Validation

## Introduction

ZenoEngine provides several different approaches to validate your application's incoming data. Most commonly, you will use the built-in `validate` slot within your route handlers:

## Quick Reference

```zeno
http.post: '/users' {
    validate {
        name: 'required'
        email: 'required,email'
        password: 'required,min:8,confirmed'
    }

    // If validation fails, errors are redirected back automatically
    orm.model: 'users'
    orm.save: $form
}
```

## Available Validation Rules

| Rule | Description |
| --- | --- |
| `required` | The field must be present and non-empty. |
| `email` | The field must be a valid email address. |
| `min:N` | The field must have a minimum value or string length of N. |
| `max:N` | The field must have a maximum value or string length of N. |
| `confirmed` | Must have a matching `_confirmation` field (e.g. `password_confirmation`). |
| `unique:table` | The field must be unique in the given database table. |
| `exists:table` | The field must exist in the given database table. |
| `in:a,b,c` | The field must be one of the given values. |

## Displaying Validation Errors

You can display validation errors in your Blade templates using `@error`:

```blade
<form method="POST" action="/users">
    @csrf

    <div>
        <label>Email</label>
        <input type="email" name="email" value="{{ old('email') }}">

        @error('email')
            <span class="error">{{ $message }}</span>
        @enderror
    </div>

    <button type="submit">Register</button>
</form>
```
