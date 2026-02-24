# Query Builder

## Introduction

ZenoEngine's database query builder provides a convenient, fluent interface to creating and running database queries. It can be used to perform most database operations in your application and works perfectly with all database systems supported by ZenoEngine.

The query builder uses PDO parameter binding to protect your application against SQL injection attacks. There is no need to clean or sanitize strings passed to the query builder as query bindings.

## Retrieving Results

### Retrieving All Rows From a Table

```zeno
db.table: 'users'
db.get { as: $users }
```

### Retrieving A Single Row

```zeno
db.table: 'users'
db.where {
    col: 'name'
    val: 'John'
}
db.first { as: $user }
```

## Where Clauses

### Basic Where Clauses

```zeno
db.table: 'users'
db.where {
    col: 'votes'
    op: '>'
    val: 100
}
db.get { as: $users }
```

### Or Where Clauses

```zeno
db.table: 'users'
db.where {
    col: 'votes'
    op: '>'
    val: 100
}
db.or_where {
    col: 'name'
    val: 'Dayle'
}
db.get { as: $result }
```

### Where Between

```zeno
db.table: 'orders'
db.where_between {
    col: 'price'
    val: [100, 500]
}
db.get { as: $orders }
```

### Where In

```zeno
db.table: 'users'
db.where {
    col: 'id'
    op: 'IN'
    val: [1, 2, 3]
}
db.get { as: $users }
```

## Ordering, Grouping, Limit & Offset

```zeno
db.table: 'users'
db.order_by {
    col: 'name'
    dir: 'desc'
}
db.limit: 10
db.offset: 20
db.get { as: $users }
```

## Aggregates

The query builder also provides a variety of methods for retrieving aggregate values like `count`, `max`, `min`, `avg`, and `sum`.

```zeno
db.table: 'orders'
db.count { as: $total }

db.table: 'orders'
db.max {
    col: 'price'
    as: $maxPrice
}

db.table: 'orders'
db.sum {
    col: 'price'
    as: $totalRevenue
}
```

## Joins

### Inner Join Clause

```zeno
db.table: 'users'
db.join {
    table: 'contacts'
    on: ['users.id', '=', 'contacts.user_id']
}
db.get { as: $result }
```

### Left Join Clause

```zeno
db.table: 'users'
db.left_join {
    table: 'posts'
    on: ['users.id', '=', 'posts.user_id']
}
db.get { as: $result }
```

## Insert Statements

```zeno
db.table: 'users'
db.insert {
    name: 'Alice'
    email: 'alice@example.com'
}
```

## Update Statements

```zeno
db.table: 'users'
db.where {
    col: 'id'
    val: 1
}
db.update {
    name: 'Alice Updated'
}
```

## Delete Statements

```zeno
db.table: 'users'
db.where {
    col: 'id'
    val: 1
}
db.delete
```

## Pagination

ZenoEngine makes it easy to paginate results, returning both the results and metadata automatically.

```zeno
db.table: 'users'
db.paginate {
    per_page: 15
    page: 1
    as: $paginator
}
// $paginator.data = list of items
// $paginator.total = total records
// $paginator.last_page = number of pages
// $paginator.current_page = current page
```

## Checking Existence

```zeno
db.table: 'users'
db.where {
    col: 'email'
    val: 'alice@example.com'
}
db.exists { as: $hasUser }
```
