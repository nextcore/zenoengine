# Blade Templates

## Introduction

Blade is the simple, yet powerful templating engine provided with ZenoEngine. Unlike some other templating engines, Blade does not restrict you from using plain ZenoLang code in your templates. In fact, all Blade templates are compiled into native Go code and cached until they are modified.

Blade template files use the `.blade.zl` file extension and are typically stored in the `resources/views` directory.

## Displaying Data

You may display data that is passed to your Blade views by wrapping the variable in `{{ }}` braces. By default, `{{ }}` statements are automatically sent through PHP's `htmlspecialchars` equivalent to prevent XSS attacks:

```blade
Hello, {{ $name }}.
```

### Displaying Unescaped Data

If you do **not** want your data to be escaped, you may use the `{!! !!}` syntax:

```blade
Hello, {!! $content !!}.
```

## Blade Directives

In addition to template inheritance and displaying data, Blade also provides convenient shortcuts for common PHP control structures, such as conditional statements and loops.

### If Statements

```blade
@if ($user->isAdmin())
    <p>Welcome, Admin!</p>
@elseif ($user->isMember())
    <p>Welcome, Member!</p>
@else
    <p>Welcome, Guest!</p>
@endif
```

### Loops

Blade provides simple directives for working with PHP's looping structures.

```blade
@foreach ($users as $user)
    <li>{{ $user->name }}</li>
@endforeach
```

#### The `$loop` Variable

When looping, a `$loop` variable will be available inside your loop. This variable provides access to some useful bits of information such as the current loop index and whether this is the first or last iteration through the loop:

```blade
@foreach ($users as $user)
    @if ($loop->first)
        This is the first iteration.
    @endif

    @if ($loop->last)
        This is the last iteration.
    @endif

    <p>{{ $loop->index }}: {{ $user->name }}</p>
@endforeach
```

| Property | Description |
| --- | --- |
| `$loop->index` | The index of the current loop iteration (starts at 0). |
| `$loop->iteration` | The current loop iteration (starts at 1). |
| `$loop->first` | Whether this is the first iteration. |
| `$loop->last` | Whether this is the last iteration. |
| `$loop->count` | The total number of items in the collection being iterated. |

## Forms

### CSRF Field

Anytime you define an HTML form in your application, you should include a hidden CSRF token field in the form so that the CSRF protection middleware can validate the request. You may use the `@csrf` Blade directive to generate the token field:

```blade
<form method="POST" action="/profile">
    @csrf
    ...
</form>
```

### Method Fields

Since HTML forms can't make `PUT`, `PATCH`, or `DELETE` requests, you will need to add a hidden `_method` field to spoof these HTTP verbs. The `@method` Blade directive can create this field for you:

```blade
<form action="/post/my-post" method="POST">
    @method('PUT')
    @csrf
    ...
</form>
```

## Validation Errors

The `@error` directive may be used to quickly check if validation error messages exist for a given attribute:

```blade
<input type="email" name="email">

@error('email')
    <div class="alert alert-danger">{{ $message }}</div>
@enderror
```

## Layouts

### Defining A Layout

```blade
<!-- resources/views/layouts/app.blade.zl -->
<html>
<head>
    <title>App - @yield('title')</title>
</head>
<body>
    @yield('content')
</body>
</html>
```

### Extending A Layout

```blade
@extends('layouts.app')

@section('title', 'Page Title')

@section('content')
    <p>This is my body content.</p>
@endsection
```

## Including Sub-Views

Blade's `@include` directive allows you to include a Blade view from within another view:

```blade
<div>
    @include('shared.errors')

    <form>
        <!-- Form Contents -->
    </form>
</div>
```

## Per-App View Roots (Multi-App)

When running multiple applications inside a single ZenoEngine instance, each app needs its own view directory. Use `view.root:` to set the Blade view root directory per app.

```zeno
// Declare at the top of your app's route file
view.root: 'apps/blog/resources/views'

http.get: '/blog' {
    // Automatically resolves to: apps/blog/resources/views/index.blade.zl
    view: 'index'
}
```

`view.root:` applies to **all** Blade operations in the current context:

| Directive | Resolution |
| --- | --- |
| `view: 'index'` | `{view.root}/index.blade.zl` |
| `@extends('layouts.app')` | `{view.root}/layouts/app.blade.zl` |
| `@include('partials.nav')` | `{view.root}/partials/nav.blade.zl` |
| `@component('alert')` | `{view.root}/components/alert.blade.zl` |

If `view.root:` is not set, it defaults to `views/`.

::: tip Multi-App Architecture
See the [Multi-App Architecture](/ecosystem/multi-app) guide for a complete example of hosting multiple apps with isolated views, routes, and static assets.
:::

