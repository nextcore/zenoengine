# Pagination

ZenoEngine's `db.paginate` slot returns a complete pagination object, similar to Laravel's `LengthAwarePaginator`.

## Basic Usage

```zeno
db.table: 'users'
db.paginate {
    per_page: 15
    page: 1
    as: $paginator
}
```

## Paginator Object

The `$paginator` object has the following structure:

| Property | Description |
| --- | --- |
| `$paginator.data` | The array of items for the current page. |
| `$paginator.total` | The total number of matching records. |
| `$paginator.per_page` | Number of items per page. |
| `$paginator.current_page` | The current page number. |
| `$paginator.last_page` | The total number of pages. |
| `$paginator.from` | The starting index of items on this page. |
| `$paginator.to` | The ending index of items on this page. |

## In a Controller + View

```zeno
http.get: '/users' {
    db.table: 'users'
    db.paginate {
        per_page: 15
        page: $request.query.page
        as: $paginator
    }
    view: 'users.index' {
        paginator: $paginator
    }
}
```

```blade
{{-- resources/views/users/index.blade.zl --}}
@foreach ($paginator->data as $user)
    <div>{{ $user->name }}</div>
@endforeach

<div class="pagination">
    <span>Page {{ $paginator->current_page }} of {{ $paginator->last_page }}</span>
</div>
```
