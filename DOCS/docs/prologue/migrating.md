# Migrating from Laravel

ZenoEngine was designed with Laravel developers in mind. If you know Laravel, you're already halfway to knowing ZenoEngine. This guide maps common Laravel patterns to their ZenoEngine equivalents.

## Routing

| Laravel | ZenoEngine |
| --- | --- |
| `Route::get('/users', ...)` | `http.get: '/users' { ... }` |
| `Route::post('/users', ...)` | `http.post: '/users' { ... }` |
| `Route::middleware('auth')` | `middleware: 'auth'` inside `http.group` |
| `Route::prefix('/api')` | `http.group: '/api' { ... }` |
| `Route::domain('api.myapp.com')` | `http.host: 'api.myapp.com'` |

## Eloquent ORM

| Laravel | ZenoEngine |
| --- | --- |
| `User::all()` | `orm.model: 'users'` → `db.get { as: $users }` |
| `User::find($id)` | `orm.find: $id { as: $user }` |
| `User::create($data)` | `orm.model: 'users'` → `orm.save: $data` |
| `$user->save()` | `orm.save: $user` |
| `$user->delete()` | `orm.delete: $user` |
| `User::with('posts')` | `orm.with: 'posts' { ... }` |

## Query Builder

| Laravel | ZenoEngine |
| --- | --- |
| `DB::table('users')->get()` | `db.table: 'users'` → `db.get { as: $users }` |
| `->where('votes', '>', 100)` | `db.where { col:'votes' op:'>' val:100 }` |
| `->orWhere('name', 'Dayle')` | `db.or_where { col:'name' val:'Dayle' }` |
| `->whereBetween('price', [1, 100])` | `db.where_between { col:'price' val:[1,100] }` |
| `->join('table', 'a', '=', 'b')` | `db.join { table:'t' on:['a','=','b'] }` |
| `->count()` | `db.count { as: $count }` |
| `->paginate(15)` | `db.paginate { per_page:15 as: $paginator }` |

## Blade Templating

| Laravel | ZenoEngine |
| --- | --- |
| `{{ $name }}` | `{{ $name }}` ✅ Identical |
| `{!! $html !!}` | `{!! $html !!}` ✅ Identical |
| `@if / @endif` | `@if / @endif` ✅ Identical |
| `@foreach / @endforeach` | `@foreach / @endforeach` ✅ Identical |
| `@csrf` | `@csrf` ✅ Identical |
| `@method('PUT')` | `@method('PUT')` ✅ Identical |
| `@error('field')` | `@error('field')` ✅ Identical |
| `@extends / @section / @yield` | `@extends / @section / @yield` ✅ Identical |
| `@include('partial')` | `@include('partial')` ✅ Identical |
| `$loop->first`, `$loop->last` | `$loop->first`, `$loop->last` ✅ Identical |

## Validation

| Laravel | ZenoEngine |
| --- | --- |
| `$request->validate(['email' => 'required\|email'])` | `validate { email: 'required,email' }` |
| `$errors->has('email')` | `@error('email')` in Blade |

## What is Different?

There are a few things that work differently in ZenoEngine by design:

1. **No Classes or PHP**: ZenoEngine uses ZenoLang, a declarative scripting language instead of PHP classes. There are no `class User extends Model` definitions. Instead, models are configured via slots (`orm.model: 'users' { ... }`).

2. **No Composer**: There is no package manager like Composer. ZenoEngine's ecosystem comes built-in via its slots system. WASM plugins can be used to extend functionality.

3. **Single Binary**: Your entire application compiles to a single Go binary. There is no PHP runtime, no PHP-FPM, and no Nginx configuration needed.

4. **Performance**: ZenoEngine is orders of magnitude faster than a traditional Laravel setup, handling thousands of concurrent requests with minimal memory usage.
