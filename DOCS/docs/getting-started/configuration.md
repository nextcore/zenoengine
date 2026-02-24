# Configuration

## Introduction

All of the configuration for the ZenoEngine framework is read from environment variables, typically defined in a `.env` file in the root of your project. This approach is identical to Laravel's configuration system and keeps sensitive credentials out of your version control.

## The `.env` File

A fresh ZenoEngine installation includes a `.env.example` file. Copy it to `.env`:

```bash
cp .env.example .env
```

A typical `.env` file looks like this:

```env
APP_NAME=MyApp
APP_ENV=local
APP_DEBUG=true
APP_URL=http://localhost:3000

DB_CONNECTION=sqlite
DB_DATABASE=./data/database.db

JWT_SECRET=your-very-secret-key

ZENO_REQUEST_TIMEOUT=30s
```

## Key Configuration Values

| Variable | Description | Default |
| --- | --- | --- |
| `APP_NAME` | Your application name | `ZenoApp` |
| `APP_ENV` | Environment (`local`, `production`) | `local` |
| `APP_PORT` | Port the server listens on | `3000` |
| `DB_CONNECTION` | Database driver (`sqlite`, `mysql`, `postgres`) | `sqlite` |
| `DB_HOST` | Database host | `127.0.0.1` |
| `DB_PORT` | Database port | `3306` |
| `DB_DATABASE` | Database name or file path | `./data/database.db` |
| `DB_USERNAME` | Database username | — |
| `DB_PASSWORD` | Database password | — |
| `JWT_SECRET` | Secret key for JWT token signing | — |
| `ZENO_REQUEST_TIMEOUT` | Per-request timeout limit | `30s` |

## Accessing Configuration in ZenoLang

You can read environment variables in your `.zl` scripts using the `env` slot:

```zeno
env: 'APP_NAME' { as: $appName }
print: $appName
```
