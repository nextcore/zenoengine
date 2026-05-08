# Database Migrations

## Introduction

Migrations are like version control for your database, allowing your team to define and share the application's database schema definition. ZenoEngine executes `.zl` files in the `migrations/` directory to manage schema changes.

## Generating Migrations

Create a new migration file in `migrations/` directory at the root of your project:

```text
migrations/
├── 001_create_users_table.zl
├── 002_create_posts_table.zl
└── 003_add_team_id_to_users.zl
```

Files are executed in alphabetical order.

## Migration Structure

ZenoEngine supports two styles of migrations: **Schema Builder** (Recommended) and **Raw SQL**.

### 1. Schema Builder (Recommended)

The Schema Builder provides a database-agnostic way to create tables using ZenoLang syntax. You can wrap your operations in `up` and `down` blocks to support rollbacks.

```zeno
// migrations/001_create_users_table.zl
up {
    db.create_table: 'users' {
        db.id: 'id'
        db.string: 'name' { limit: 100 }
        db.string: 'email' { unique: true }
        db.timestamp: 'created_at'
    }
}

down {
    db.drop_table: 'users'
}
```

If you don't use `up` and `down` blocks, the file will be executed top-to-bottom during migration (equivalent to `up`), but it won't support automatic rollback.

#### Supported Column Types:
- `db.id`: Auto-incrementing primary key.
- `db.string`: String/VARCHAR column. Supports `limit` child (default 255).
- `db.integer`: Integer column.
- `db.timestamp`: Datetime/Timestamp column.
- `db.boolean`: Boolean column.
- `db.text`: Long text column.
- `db.decimal`: Decimal/Numeric column. Supports `precision` and `scale` children.
- `db.date`: Date column.
- `db.json`: JSON column (becomes JSONB in Postgres, JSON in MySQL, TEXT in SQLite).

#### Supported Properties:
- `limit`: Set length (for string).
- `unique`: Set unique constraint (true/false).
- `nullable`: Set nullable (true/false).
- `precision`: Set precision (for decimal).
- `scale`: Set scale (for decimal).

#### Foreign Keys:
You can define foreign keys inside `db.create_table`:

```zeno
db.create_table: 'posts' {
    db.id: 'id'
    db.integer: 'user_id'
    db.foreign: 'user_id' {
        references: 'users'
        on: 'id'
        on_delete: 'CASCADE' // Optional
    }
}
```

### 2. Raw SQL Style

If you need to execute complex SQL queries or specific database features, you can use raw SQL via the `db.execute` slot.

```zeno
// migrations/002_create_posts_table_raw.zl
db.execute: "CREATE TABLE posts (id INTEGER PRIMARY KEY, title TEXT, content TEXT)"
```

## Running Migrations

Run all pending migrations in the default `migrations/` directory:

```bash
./zeno migrate
```

Run migrations from a specific folder (e.g., for a module):

```bash
./zeno migrate modules/blog/migrations
```

## Rolling Back Migrations

To rollback the last batch of migrations:

```bash
./zeno migrate:rollback
```

This will execute the `down` blocks of the migrations in the last batch in reverse order.
