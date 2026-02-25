# ZenoEngine MVC Boilerplate

Welcome to your new ZenoEngine application! This project uses a classic **Model-View-Controller (MVC)** architecture inspired by frameworks like Laravel.

## 🚀 Features Included
- **Authentication**: Pre-built Login, Register, and Logout flows.
- **User Management**: Admin dashboard to view all registered users.
- **Middleware**: Ready-to-use `auth` and `admin` protection.
- **Database**: Pre-configured SQLite database with a `users` migration.

## 📁 Directory Structure
```text
project/
├── app/
│   ├── controllers/   # Logical handlers for routes
│   └── middleware/    # Route filters (e.g., auth, admin)
├── database/
│   └── migrations/    # Database schema definitions
├── public/            # Static files (CSS, JS, images)
├── resources/
│   └── views/         # ZenoBlade HTML templates
├── routes/
│   ├── api.zl         # JSON API routes
│   └── web.zl         # HTML Web routes
└── src/
    └── main.zl        # ZenoLang application entry point
```

## 🏃 Getting Started
1. Start the server:
   ```bash
   ./zeno
   ```
2. Open your browser and visit: `http://localhost:3000`

### Default Admin Account
The database is pre-seeded with an admin account:
- **Email**: `admin@zeno`
- **Password**: `password123`
