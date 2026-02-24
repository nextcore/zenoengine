# Database Seeding

ZenoEngine includes the ability to seed your database with data using seed classes. All seed files are stored in the `database/seeders` directory.

## Writing Seeders

A seeder is a simple ZenoLang script using the `db.seed` slot:

```zeno
db.seed {
    db.table: 'users'
    db.insert {
        name: 'Alice'
        email: 'alice@example.com'
        password: 'hashed_password'
    }

    db.table: 'users'
    db.insert {
        name: 'Bob'
        email: 'bob@example.com'
        password: 'hashed_password'
    }
}
```

## Running Seeders

You can run the database seeder using the Zeno CLI:

```bash
zeno db:seed
```

## Calling Additional Seeders

You may wish to use the `db.seed` block to control which seeder classes are executed:

```zeno
db.seed {
    // ZenoEngine will execute all registered seeders
}
```
