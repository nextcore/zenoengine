# Eager Loading

## Introduction

When accessing Eloquent relationships, the relationship data is **lazy loaded** by default. This means the relationship data is not actually loaded until you access the property. However, this creates the classic **N+1 query problem**.

## The N+1 Problem

Imagine you have a post with authors:

```zeno
// ❌ BAD: This causes N+1 queries!
// 1 query for posts, then 1 query PER POST for the user
orm.model: 'posts'
db.get { as: $posts }

@foreach($posts as $post)
    // This triggers a new DB query for EACH post!
    print: $post.user.name
@endforeach
```

## Solving N+1 with `orm.with`

ZenoEngine solves this natively using `orm.with`, which executes exactly **2 queries total** regardless of the number of parent records:

```zeno
// ✅ GOOD: Only 2 SQL queries total, no matter how many posts

// Step 1: Define model with relationship
orm.model: 'posts' {
    orm.belongsTo: 'users' {
        as: 'author'
        foreign_key: 'user_id'
    }
}

// Step 2: Fetch all posts
orm.model: 'posts'
db.get { as: $posts }

// Step 3: Eager load authors in a SINGLE query
orm.model: 'posts'
orm.with: 'author' {
    set: $posts { val: $posts }
}

// Now each post has $post.author with no extra queries!
@foreach($posts as $post)
    print: $post.author.name
@endforeach
```

## How it Works Internally

`orm.with` follows a 3-step process identical to Laravel's eager loading:

1. **Collect Keys**: Scan all parent records and collect unique foreign key values (e.g., all `user_id` values from `$posts`).
2. **Single Batch Query**: Execute one optimized `WHERE id IN (...)` query to fetch all related records at once.
3. **Map In Memory**: Match and attach related records back to the correct parent objects in memory using a dictionary (O(1) lookup).

This is the same algorithm Laravel uses internally in `Eloquent::with()`.

::: info Query Count Comparison
| Approach | Queries for 1000 posts |
| --- | --- |
| Lazy loading | 1001 queries |
| `orm.with` Eager Loading | **2 queries** |
:::
