# ZenoLang Maximum Potential Demo

**Task Management & Team Collaboration Platform**

A comprehensive demonstration of ZenoLang's capabilities, showcasing modern web development features including authentication, real-time updates, file uploads, background jobs, and RESTful APIs.

---

## 🚀 Features Demonstrated

### ✅ **Authentication & Authorization**
- JWT-based authentication
- Secure password hashing (Bcrypt)
- Role-based access control (Admin/Member)
- Session management with cookies
- Protected routes with middleware

### ✅ **Database Operations**
- SQLite database (portable)
- Database migrations
- Query Builder (safe, SQL-injection proof)
- Database transactions (ACID compliance)
- Foreign key relationships
- Auto-seeding demo data

### ✅ **Task Management (CRUD)**
- Create, Read, Update, Delete operations
- Advanced filtering and search
- Pagination support
- Status tracking (Pending/In Progress/Completed)
- Priority levels (Low/Medium/High)
- Due date management
- Task assignment to users/teams

### ✅ **File Upload & Management**
- Avatar upload during registration
- Task attachment uploads
- File storage in organized directories
- Secure file handling

### ✅ **Real-time Features**
- Server-Sent Events (SSE) for notifications
- Live notification stream
- Real-time dashboard updates
- Unread notification counter

### ✅ **Team Collaboration**
- Team creation and management
- Team ownership and permissions
- Task assignment to teams
- Member management

### ✅ **RESTful API**
- Full CRUD API for tasks
- API for teams
- JWT authentication for API
- Pagination support
- Proper HTTP status codes
- JSON responses

### ✅ **Modern UI/UX**
- Responsive design (mobile-friendly)
- Modern CSS with variables
- Card-based layouts
- Interactive modals
- Form validation
- Beautiful color schemes
- Smooth animations

### ✅ **Advanced ZenoLang Features**
- Helper functions (`fn` slot)
- Error handling (`try-catch`)
- Conditional logic (`if-else`, `switch`)
- Loops (`for`, `forelse`, `while`)
- Validation with rules
- Date/time manipulation
- String operations
- Null safety (`coalesce`)
- Debugging tools (`log`, `dump`)

---

## 📦 Installation & Setup

### Database Configuration

This tutorial uses a **separate SQLite database** (`tutorial_max.db`) to avoid conflicts with your main ZenoEngine database.

The database configuration is already set in `.env`:
```env
DB_TUTORIAL_DRIVER=sqlite
DB_TUTORIAL_NAME=./tutorial_max.db
```

All tables use the `tutorial_` prefix:
- `tutorial_users`
- `tutorial_teams`
- `tutorial_tasks`
- `tutorial_notifications`

### 1. **Include in Main Application**

Add this line to your `/src/main.zl`:

```javascript
include: src/tutorial/max/main.zl
```

### 2. **Start the Server**

```bash
cd /home/max/Documents/PROJ/ZenoEngine
go run cmd/zeno/zeno.go
```

The tutorial database (`tutorial_max.db`) will be created automatically on first run.

### 3. **Access the Application**

Open your browser and navigate to:
```
http://localhost:8080/tutorial/max
```

---

## 🔐 Demo Credentials

The application comes with pre-seeded demo data:

| Role | Email | Password |
|------|-------|----------|
| **Admin** | admin@demo.com | password123 |
| **User** | john@demo.com | password123 |
| **User** | jane@demo.com | password123 |

---

## 📚 API Documentation

### Authentication

All API endpoints require JWT authentication. Include the token in the Authorization header:

```bash
Authorization: Bearer <your_jwt_token>
```

### Endpoints

#### **Tasks API**

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/tasks` | List all tasks (with pagination) |
| POST | `/api/v1/tasks` | Create a new task |
| GET | `/api/v1/tasks/:id` | Get a specific task |
| PUT | `/api/v1/tasks/:id` | Update a task |
| DELETE | `/api/v1/tasks/:id` | Delete a task |

**Example: Create Task**
```bash
curl -X POST http://localhost:8080/api/v1/tasks \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "New Task",
    "description": "Task description",
    "priority": "high",
    "status": "pending"
  }'
```

#### **Teams API**

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/teams` | List all teams |
| POST | `/api/v1/teams` | Create a new team |

---

## 🗂️ Project Structure

```
/src/tutorial/max/
├── main.zl                    # Entry point
├── migrations/                # Database migrations
│   ├── 001_users.zl
│   ├── 002_teams.zl
│   ├── 003_tasks.zl
│   └── 004_notifications.zl
├── seeders/                   # Demo data
│   └── demo_data.zl
├── utils/                     # Helper functions
│   └── helpers.zl
├── modules/
│   ├── auth/                  # Authentication
│   │   └── routes.zl
│   ├── tasks/                 # Task management
│   │   ├── list.zl
│   │   ├── create.zl
│   │   ├── edit.zl
│   │   ├── delete.zl
│   │   └── complete.zl
│   ├── teams/                 # Team management
│   │   └── routes.zl
│   └── realtime/              # Real-time features
│       ├── notifications.zl
│       └── dashboard.zl
└── api/v1/                    # RESTful API
    ├── tasks.zl
    └── teams.zl

/views/tutorial/max/
├── layout.blade.zl            # Master layout
├── dashboard.blade.zl         # Dashboard
├── auth/
│   ├── login.blade.zl
│   └── register.blade.zl
├── tasks/
│   ├── index.blade.zl
│   ├── create.blade.zl
│   └── edit.blade.zl
└── teams/
    └── index.blade.zl

/public/tutorial/max/
├── css/
│   └── app.css               # Modern CSS
└── js/
    └── app.js                # JavaScript
```

---

## 🎯 Key Learning Points

### 1. **Database Transactions**
See how transactions ensure data integrity in `seeders/demo_data.zl` and `modules/tasks/create.zl`.

### 2. **Authentication Middleware**
Check `modules/auth/routes.zl` for JWT implementation and `modules/tasks/list.zl` for middleware usage.

### 3. **File Uploads**
See `modules/auth/routes.zl` (avatar) and `modules/tasks/create.zl` (attachments) for file handling.

### 4. **Real-time with SSE**
Explore `modules/realtime/notifications.zl` for Server-Sent Events implementation.

### 5. **Query Builder**
All database operations use the safe Query Builder instead of raw SQL.

### 6. **Validation**
Every form has comprehensive validation - see any create/update route.

### 7. **Helper Functions**
Reusable functions in `utils/helpers.zl` demonstrate the `fn` slot.

---

## 🔧 Customization

### Change Database
The demo uses SQLite by default. To use MySQL or PostgreSQL, update `.env`:

```env
DB_DRIVER=mysql
DB_HOST=127.0.0.1:3306
DB_USER=root
DB_PASS=password
DB_NAME=zenolang_demo
```

### Add More Features
- Extend the API with more endpoints
- Add more real-time features
- Implement background jobs
- Add email notifications
- Create more complex queries

---

## 🐛 Troubleshooting

### Database Not Created
Make sure the migrations run on first start. Check server logs for migration messages.

### Login Not Working
Ensure the database is seeded with demo data. Check for "Demo data seeded successfully" in logs.

### File Uploads Failing
Verify that `uploads/avatars/` and `uploads/attachments/` directories exist and are writable.

---

## 📖 ZenoLang Features Used

- ✅ `http.get`, `http.post`, `http.put`, `http.delete` - Routing
- ✅ `auth.middleware`, `auth.login`, `auth.user` - Authentication
- ✅ `db.table`, `db.insert`, `db.update`, `db.delete`, `db.get` - Database
- ✅ `db.transaction` - Transactions
- ✅ `validate` - Input validation
- ✅ `http.upload` - File uploads
- ✅ `crypto.hash` - Password hashing
- ✅ `sse.stream`, `sse.send` - Real-time
- ✅ `fn`, `call` - Functions
- ✅ `if`, `for`, `forelse`, `switch` - Control flow
- ✅ `try-catch` - Error handling
- ✅ `view.blade` - Templating
- ✅ `json.stringify` - JSON operations
- ✅ `date.now`, `date.format` - Date/time
- ✅ `coalesce` - Null safety

---

## 🎨 UI Screenshots

The application features:
- Modern, clean design with Inter font
- Responsive layout (works on mobile)
- Color-coded priority and status badges
- Interactive cards with hover effects
- Professional forms with validation
- Beautiful statistics dashboard
- Real-time notification system

---

## 📝 License

This is a demonstration project for ZenoLang. Feel free to use it as a starting point for your own projects!

---

## 🤝 Contributing

This demo showcases ZenoLang's capabilities. To learn more about ZenoLang, check the main documentation.

---

**Built with ⚡ ZenoLang** - The fast, declarative backend framework
