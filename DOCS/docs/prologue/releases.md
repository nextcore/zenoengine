# Release Notes

## ZenoEngine v1.0

### New Features

- **Blade Templating Engine**: Full Laravel-parity Blade transpiler with `@if`, `@foreach`, `@extends`, `@section`, `@yield`, `@include`, `@error`, `@csrf`, `@method`, and the `$loop` variable.
- **ORM & Relationships**: Eloquent-inspired ORM with `orm.model`, `orm.save`, `orm.find`, `orm.delete`, and all four relationship types: `hasOne`, `hasMany`, `belongsTo`, `belongsToMany`.
- **Eager Loading (N+1 Fix)**: `orm.with` resolves relationships in exactly 2 SQL queries regardless of dataset size.
- **Mass Assignment Protection**: Define `fillable` in `orm.model` to guard against unsafe mass assignment.

### Query Builder

- Full set of `WHERE` clauses: `db.where`, `db.or_where`, `db.where_between`, `db.where_null`, `db.where_in`.
- Joins: `db.join`, `db.left_join`, `db.right_join`.
- Aggregates: `db.count`, `db.sum`, `db.avg`, `db.min`, `db.max`.
- `db.pluck`, `db.exists`, `db.doesnt_exist`.
- `db.paginate` with full metadata (total, last_page, current_page).

### Routing

- `http.get`, `http.post`, `http.put`, `http.patch`, `http.delete`.
- Route groups with `http.group` and middleware support.
- Virtual host routing with `http.host`.
- Static file & SPA hosting with `http.static`.
- Reverse proxy with `http.proxy`.

### Developer Experience

- Auto-injected `$request`, `$form`, `$auth`, `$params` into route scope.
- Request timeout protection (configurable via `ZENO_REQUEST_TIMEOUT`).
- Automatic API documentation generation from route metadata.
- WASM plugin system for extending ZenoEngine without recompiling.
