# ZenoEngine Modular Boilerplate

Welcome to your new ZenoEngine application! This project uses a **Modular (Domain-Driven)** architecture. Instead of splitting files by type (controllers vs views), files are grouped by feature (e.g., `modules/auth`, `modules/users`). This makes the codebase highly scalable and easier to navigate for large applications.

## 🚀 Features Included
- **Authentication**: Pre-built Login, Register, and Logout flows inside `modules/auth`.
- **User Management**: Dashboards and Admin User Management inside `modules/users`.
- **Middleware**: Shared `auth` and `admin` logic inside `modules/core/middleware.zl`.
- **Database**: Pre-configured SQLite database with a `users` migration.

## 📁 Directory Structure
```text
project/
├── api/
│   └── v1/            # (Optional) REST API endpoints
├── database/
│   └── migrations/    # Database schema definitions
├── modules/           # Feature-based Domain Modules
│   ├── auth/          # Authentication routes & views
│   ├── core/          # Shared layout & middleware
│   └── users/         # User dashboard & management
├── public/            # Static files (CSS, JS, images)
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
