# Complete ZenoLang & ZenoEngine Documentation
**Official Guide: Full Edition (v1.2)**

---

## 1. Introduction
ZenoLang is a slot-based configuration language designed for **fast**, **declarative**, and **readable** backend development. ZenoLang runs on **ZenoEngine**, a high-performance Go runtime.

Core Philosophy:
- **Human Friendly:** Code should read like simple English instructions.
- **Slot-Based:** All commands are "slots". Format: `slot: value { children }`.
- **Batteries Included:** Built-in features for Database, Auth, HTTP, File System, Background Jobs, and more.

---

## 2. Core Syntax & Control Flow

### 2.1 Variables & Logs
```javascript
// Assignment (Modern)
var: $name { val: "John" }

// Legacy Assignment
scope.set: $age { val: 25 }

// String Interpolation
log: "User " + $name + " is " + $age
```

### 2.2 Conditionals
```javascript
if: $age >= 18 {
  then: { log: "Adult" }
  else: { log: "Minor" }
}

// Switch Case
switch: $role {
  case: "admin" { log: "Welcome Admin" }
  case: "editor" { log: "Welcome Editor" }
  default: { log: "Welcome User" }
}
```

### 2.3 Loops
```javascript
// Foreach
for: $items {
  as: $item
  do: {
    log: $item.name
    
    // Conditional Break/Continue
    break: "$item.id == 5"
    continue: "$item.active == false"
  }
}

// While
while: "$count < 5" {
  do: {
    log: $count
    math.calc: $count + 1 { as: $count }
  }
}
```

### 2.4 Error Handling & Functions
```javascript
// Try-Catch
try: {
  do: {
    db.execute: "INVALID SQL"
  }
  catch: {
    log: "Error: " + $error
  }
}

// Global Functions
fn: calculate_tax {
  math.calc: $amount * 0.1 { as: $tax }
}

call: calculate_tax
```

### 2.5 Type Safety & Schema
ZenoEngine provides both static and runtime type validation to ensure code reliability.

#### Static Analysis (CLI)
Use the `check` command to validate your code without running it. It catches unknown slots, missing attributes, and literal type mismatches.
```bash
zeno check path/to/script.zl
```

#### Explicit Type Locking (Schema)
Use the `schema` slot to lock a variable to a specific type.
```javascript
schema: $user_id { type: "int" }
```

#### Typed Variables
You can enforce types when defining variables using the `var` slot.
```javascript
var: $count {
  val: 10
  type: "int"
}
```

Supported Types: `string`, `int`, `bool`, `float`, `decimal`, `list`, `map`, `any`.

#### When does validation occur?
1.  **Static Analysis (`zeno check`)**: Before execution. Validates literal values (e.g., `sleep: "ten"`) and structure. Catches typos early.
2.  **Runtime Validation**: During execution. Validates dynamic data (e.g., HTTP input or variable values) when assigned or used in a `schema` locked variable.

#### Special Type: `decimal`
The `decimal` type is designed for high-precision financial data. While internally it uses string-based storage to prevent floating-point errors, ZenoEngine validates it as a valid numeric decimal. Always use `decimal` for money, prices, and taxes in combination with `money.calc`.

```javascript
// Reliable money handling
var: $price { val: "150.50", type: "decimal" }
schema: $tax { type: "decimal" }
money.calc: $price * $tax { as: $total }
```

---

## 3. HTTP & Routing
ZenoEngine uses **Chi Router**. 

### 3.1 Routing Syntax (CRITICAL)
**Always use `{curly_braces}` for parameters.** Variables are auto-injected.

```javascript
// Route with ID parameter
http.get: "/users/{id}" {
  do: {
    // $id is available here automatically
    http.ok: { id: $id }
  }
}

// Route Groups
http.group: "/api/v1" {
  middleware: "auth" // Optional: inherit middleware
  do: {
    http.get: "/profile" { ... }
  }
}
```

### 3.2 Request Data
```javascript
// Query Params (?page=1)
http.query: "page" { as: $page }

// Form/Body Data
http.form: "email" { as: $email }

// Headers
var: $auth_header { val: $request.headers.Authorization }
```

### 3.3 Responses & Redirects
```javascript
http.ok: { message: "Success" }
http.created: { id: 123 }
http.not_found: { error: "Missing" }
http.server_error: { error: "Crash" }

// Redirect
http.redirect: "/dashboard"
```

---

## 4. Database (SQL)
Supports MySQL, PostgreSQL, SQLite, SQL Server. Driver agnostic.

### 4.1 Query Builder (Recommended)
```javascript
db.table: users

// Filters
db.where: { col: status, val: "active" }
db.where: { col: age, op: ">", val: 18 }
db.where_in: { col: role, val: ["admin", "staff"] }
db.where_null: "deleted_at"

// Ordering & Limits
db.order_by: "created_at DESC"
db.limit: 10
db.offset: 0

// Execution
db.get: { as: $users }       // List
db.first: { as: $user }      // Single (or null)
db.count: { as: $total }     // Int

// Writes
db.insert: { name: "John", email: "test@test.com" } // Returns $db_last_id
db.update: { status: "inactive" }
db.delete: { as: $count }
```

### 4.2 Advanced Database
```javascript
// Joins
db.table: users
db.join: {
  table: posts
  on: ["users.id", "=", "posts.user_id"]
}

// Group By & Having
db.table: orders
db.group_by: "user_id"
db.having: { col: "total", op: ">", val: 1000 }
db.get: { as: $big_spenders }

// Transactions (Auto-Rollback)
db.transaction: {
  do: {
    db.table: accounts
    db.update: { balance: 0 }
    // ... more ops ...
  }
}
```

### 4.3 Raw SQL with Binding (Safe)
Use `db.select` and `db.execute` for complex queries. Always use `bind` to prevent SQL Injection.

```javascript
// Select with Parameters
db.select: {
  sql: "SELECT * FROM users WHERE email = ? AND status = ?"
  bind: {
    email: $email
    status: "active"
  }
  as: $users
}

// Execute with Parameters
db.execute: {
  sql: "UPDATE users SET role = ? WHERE id = ?"
  bind: {
    role: "admin"
    id: $user_id
  }
}
```

---

## 5. Input Validation
```javascript
validate: $form {
  rules: {
    username: "required|min:5"
    email: "required|email"
    age: "numeric|min:18"
  }
  as: $errors
}

if: $errors_any {
  then: { http.validation_error: { errors: $errors } }
}
```

---

## 6. Authentication & Security
Built-in JWT support.

```javascript
// Login
auth.login: {
  username: $form.email
  password: $form.password
  table: "users"        // Default
  col_user: "email"     // Default
  col_pass: "password"  // Default
  as: $token
}

// Route Protection
http.get: "/dashboard" {
  do: {
    auth.middleware: {
      do: {
        auth.user: { as: $current_user }
        http.ok: { msg: "Secret", user: $current_user }
      }
    }
  }
}

// Hashing
crypto.hash: $password { as: $hash }
crypto.verify: { hash: $hash, text: $input, as: $valid }
```

---

## 7. Filesystem & Utils

### 7.1 Files
```javascript
// Write
io.file.write: { path: "log.txt", content: "Data" }

// Read
io.file.read: "config.json" { as: $json_str }

// Delete
io.file.delete: "temp.txt"

// Make Dir
io.dir.create: "uploads/2024"
```

### 7.2 Math
```javascript
// Standard (Float)
math.calc: ceil($price * 1.1) { as: $total }

// Financial (Decimal - SAFE)
money.calc: ($amount - $fee) * 100 { as: $cents }
```

### 7.3 Background Jobs
```javascript
// Configure Queues (in main.zl)
worker.config: ["default", "emails"]

// Enqueue
job.enqueue: {
  queue: "emails"
  payload: { to: "user@test.com", subject: "Hi" }
}

### 6.2 Excel
Generate spreadsheets from templates. (See `slots.RegisterExcelSlots`)

### 6.3 Google Sheets
ZenoEngine provides built-in support for Google Sheets as a database-like data source.

#### `gsheet.get`
Fetch data from a spreadsheet.
- `id`: Spreadsheet ID.
- `range`: Cell range (e.g., `"Sheet1!A1:B10"`).
- `credentials`: path or JSON string of service account.
- `as`: Variable name for rows.

#### `gsheet.find`
Search for rows using header-based filtering.
- `range`: Cell range including headers.
- `where`: Map of criteria `{ "ColumnName": "Value" }`.
- `as`: Variable name for matching rows.

#### `gsheet.append`
Append data to the end of a sheet.
- `values`: List of lists (rows).

#### `gsheet.update`
Overwrite a specific range.

#### `gsheet.clear`
Remove data from a specific range.

---

### 7.4 Email
```javascript
mail.send: $email {
  subject: "Welcome"
  body: "<h1>Hi!</h1>"
  host: "smtp.example.com"
  port: 587
  user: $smtp_user
  pass: $smtp_pass
}
```

---

## 8. Metaprogramming
ZenoLang supports powerful metaprogramming capabilities, allowing code to treat other code as data.

### 8.1 Dynamic Execution
Use `meta.eval` to execute valid ZenoLang code strings at runtime.
```javascript
meta.eval: "log: 'Hello from meta'"
```

### 8.2 Scope Introspection
Use `meta.scope` to retrieve the current execution variables as a Map.
```javascript
var: $data { val: meta.scope }
```

### 8.3 Code Generation
You can use Blade templates to generate ZenoLang code and then execute it.
```javascript
meta.template: "codegen/route.blade.zl" {
  resource: "products"
  as: $code
}
meta.eval: $code
```
