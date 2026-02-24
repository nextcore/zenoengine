# Database Migrations

## Introduction

Migrations are like version control for your database, allowing your team to define and share the application's database schema definition. ZenoEngine uses the `db.migrate` slot to manage schema changes.

## Generating Migrations

Create a new migration file in `database/migrations/`:

```text
database/
└── migrations/
    ├── 001_create_users_table.zl
    ├── 002_create_posts_table.zl
    └── 003_add_team_id_to_users.zl
```

## Migration Structure

A basic migration to create a table:

```zeno
// database/migrations/001_create_users_table.zl
db.migrate {
    create_table: 'users' {
        id: 'bigint primary key autoincrement'
        name: 'varchar(255) not null'
        email: 'varchar(255) not null unique'
        password: 'varchar(255) not null'
        created_at: 'timestamp default current_timestamp'
        updated_at: 'timestamp default current_timestamp'
    }
}
```

## Running Migrations

Run all pending migrations:

```bash
zeno migrate
```

Roll back the last batch:

```bash
zeno migrate:rollback
```

Reset all migrations:

```bash
zeno migrate:reset
```

## Common Column Types

| Type | Description |
| --- | --- |
| `bigint primary key autoincrement` | Auto-incrementing primary key |
| `varchar(255)` | String up to 255 characters |
| `text` | Long string / blob text |
| `integer` | Integer |
| `boolean` | True/false |
| `timestamp` | Date and time |
| `decimal(8,2)` | Decimal number (for prices) |

## Adding Columns to Existing Tables

```zeno
// database/migrations/003_add_team_id_to_users.zl
db.migrate {
    alter_table: 'users' {
        add: 'team_id integer default null'
    }
}
```
