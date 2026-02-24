# Database: Getting Started

## Introduction

ZenoEngine provides a clean and consistent interface to working with databases. It supports **SQLite**, **MySQL**, and **PostgreSQL** out of the box.

## Configuration

Database connections are configured in your `.env` file:

```env
DB_CONNECTION=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_DATABASE=my_database
DB_USERNAME=root
DB_PASSWORD=secret
```

For SQLite (perfect for local development or testing):

```env
DB_CONNECTION=sqlite
DB_DATABASE=./data/database.db
```

## Running Queries

You can run raw SQL queries using the `db.query` slot:

```zeno
db.query: 'SELECT * FROM users WHERE active = ?' {
    bind: { active: 1 }
    as: $users
}
```

## Multiple Connections

ZenoEngine supports multiple named database connections. After configuring them in your database config, you can specify which connection to use:

```zeno
db.connection: 'analytics'
db.table: 'events'
db.get { as: $events }
```

## Migrations

ZenoEngine supports database migrations to keep your database schema in sync with your application:

```zeno
db.migrate {
    create_table: 'users' {
        id: 'bigint primary key autoincrement'
        name: 'varchar(255)'
        email: 'varchar(255) unique'
        created_at: 'timestamp'
    }
}
```
