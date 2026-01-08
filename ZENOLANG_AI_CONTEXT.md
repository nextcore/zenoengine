# ZenoLang AI Context & Syntax Reference

> **SYSTEM PROMPT / CONTEXT FOR AI AGENTS**
> Use this file as the SOURCE OF TRUTH when writing ZenoLang code.
> ZenoLang is a slot-based configuration language running on ZenoEngine (Go).

---

## 1. Syntax Rules (STRICT)

1.  **Structure**: Tree-based structure using indentation (2 spaces) and braces `{}`.
2.  **Slots**: Format is `slot_name: value { children }`.
    *   Example: `log: "Hello"`
    *   Example: `if: $condition { ... }`
3.  **Variables**: Always prefixed with `$` (e.g., `$user_id`).
    *   **Assignment**: `var: $name { val: "value" }` (Preferred) or `scope.set`.
    *   **Access**: `$user.name` (Dot notation), `$items.0` (Index).
    *   **Interpolation**: `"Hello " + $name` (String concatenation).
4.  **Quoting**:
    *   Strings *should* be quoted if they contain spaces or special chars: `"valid string"`.
    *   Keys in maps DO NOT need quotes: `name: "Budi"`.
5.  **Comments**: Use `//` for single line comments.

---

## 2. API Reference (Slots)

### 2.1 Database (SQL)
**Driver Agnostic**: MySQL, PostgreSQL, SQLite, SQL Server.
**Schema**: Use `db.table` to set context.

| Slot | Signature / Children | Description |
| :--- | :--- | :--- |
| `db.table` | `value: "table"`, `db: "conn"` | Set active table/connection. |
| `db.columns`| `value: ["col1", "col2"]` | Select specific columns. |
| `db.get` | `as: $var` | Fetch multiple rows (List of Maps). |
| `db.first` | `as: $var` | Fetch single row (Map). Access check: `if: $var` or `$var_found`. |
| `db.count` | `as: $var` | Count rows (Int). |
| `db.insert` | Child keys as columns. | Insert data. Returns `$db_last_id` (except Postgres). |
| `db.update` | Child keys as columns. | Update data. Requires `db.where`. |
| `db.delete` | `as: $count` (optional) | Delete data. Requires `db.where`. |
| `db.where` | `col: "name"`, `val: $val`, `op: "="` | Add WHERE clause. Op defaults to `=`. Matches `LIKE`, `>`, etc. |
| `db.where_in` | `col: "id"`, `val: [1, 2]` | WHERE IN clause. |
| `db.where_not_in`| `col: "id"`, `val: [1, 2]` | WHERE NOT IN clause. |
| `db.where_null` | `value: "col_name"` | WHERE col IS NULL. |
| `db.where_not_null`| `value: "col_name"` | WHERE col IS NOT NULL. |
| `db.group_by` | `value: "col_name"` | GROUP BY clause. |
| `db.having` | `col: "c"`, `op: ">"`, `val: 1` | HAVING clause. |
| `db.order_by` | `value: "id DESC"` | ORDER BY clause. |
| `db.limit` | `value: $int` | LIMIT clause. |
| `db.offset` | `value: $int` | OFFSET clause. |
| `db.join` | `table: "t2"`, `on: ["t1.id", "=", "t2.fk"]` | INNER JOIN. |
| `db.left_join` | `table: "t2"`, `on: [...]` | LEFT JOIN. |
| `db.transaction`| `do: { ... }` | Atomic transaction. Auto-rollback on error. |
| `db.select` | `sql: "SQL"`, `bind: { val: $p1 }`, `as: $res` | Raw SQL Select. Uses `bind` for safe parameters. |
| `db.execute` | `sql: "SQL"`, `bind: { val: $p1 }` | Raw SQL Execute. Uses `bind` for safe parameters. |

### 2.2 HTTP Server & Client
**Router**: Uses `chi` under the hood. **MUST use `{id}` for parameters**.

| Slot | Signature / Children | Description |
| :--- | :--- | :--- |
| `http.get`, `post`...| `value: "/path/{id}"`, `do: {}`| Define Route. **Use `{variable}` for params**, NOT `:variable`. |
| `http.group` | `value: "/api"`, `do: {}` | Group routes. Supports middleware inheritance. |
| `http.query` | `value: "param"`, `as: $var` | Get Query Param (`?id=1`). |
| `http.form` | `value: "field"`, `as: $var` | Get Form/Multipart data. |
| `http.response` | `status: 200`, `data: $json` | Send JSON response. |
| `http.ok`, `created`| `value: { ... }` | Send 200/201 JSON response. |
| `http.not_found` | `value: { ... }` | Send 404 response. |
| `http.redirect` | `value: "/url"` | Redirect user. |
| `http.upload` | `field: "file"`, `dest: "path"`, `as: $name` | Handle File Upload. |

### 2.3 Logic & Flow Control

| Slot | Signature / Children | Description |
| :--- | :--- | :--- |
| `if` | `value: "$a > $b"`, `then: {}`, `else: {}` | Conditional. Ops: `==, !=, >, <, >=, <=`. |
| `switch` | `value: $var`, `case: "val" {}`, `default: {}` | Switch case. |
| `while` / `loop` | `value: "cond"`, `do: {}` | While loop. |
| `for` / `foreach` | `value: $list`, `as: $item`, `do: {}` | Foreach loop. Legacy: C-style `for: "$i=0; $i<10; $i++"`. |
| `break` | `value: "$i > 5"` (optional) | Break loop. Supports optional condition. |
| `continue` | `value: "$i % 2 == 0"` (optional) | Continue loop. Supports optional condition. |
| `return` | - | Halt execution of current handler. |
| `try` | `do: {}`, `catch: {}` | Error handling. Error msg in `$error`. |
| `ctx.timeout` | `value: "5s"`, `do: {}` | Limit execution time of block. |
| `fn` | `value: name`, `children...` | Define global function. |
| `call` | `value: name` | Call global function. |
| `scope.set` | `key: "name"`, `val: $val` | Legacy alias for `var`. |
| `logic.compare` | `v1: $a`, `op: "=="`, `v2: $b`, `as: $res`| Explicit comparison. |
| `isset` | `value: $var`, `do: {}` | Execute if variable exists. |
| `empty` | `value: $var`, `do: {}` | Execute if variable is empty/null. |
| `unless` | `value: $bool`, `do: {}` | Execute if condition is false. |

### 2.4 Utils, Security & Filesystem

| Slot | Signature | Description |
| :--- | :--- |
| `log` | `value: "msg"` | Print to console. |
| `var` | `val: $val` | Set variable. |
| `sleep` | `value: ms` | Sleep for N milliseconds. |
| `coalesce` | `val: $a`, `default: "b"`, `as: $r` | Null coalescing. |
| `is_null` | `val: $a`, `as: $bool` | Check if null. |
| `cast.to_int` | `val: $v`, `as: $i` | Cast to Integer. |
| `crypto.hash` | `val: $pass`, `as: $hash` | Bcrypt hash. |
| `crypto.verify` | `hash: $h`, `text: $p`, `as: $ok` | Verify Bcrypt. |
| `sec.csrf_token`| `as: $token` | Get CSRF token. |
| `validator.validate`| `input: $map`, `rules: {f: "required"}`, `as: $err` | Validate input. Rules: `required`, `email`, `numeric`, `min:X`, `max:X`. |
| `math.calc` | `expr: "ceil($a * 1.1)"`, `as: $res` | Float math. Vars auto-converted. |
| `money.calc` | `expr: "$a - $b"`, `as: $res` | Decimal math (Financial). |
| `io.file.write` | `path: "f.txt"`, `content: "s"`, `mode: 0644` | Write file. |
| `io.file.read` | `path: "f.txt"`, `as: $content` | Read file. |
| `io.file.delete`| `path: "f.txt"` | Delete file. |
| `io.dir.create` | `path: "dir"` | Create directory (mkdir -p). |
| `image.info` | `path: "img.jpg"`, `as: $info` | Get width/height. |
| `image.resize` | `source: "src"`, `dest: "dst"`, `width: 100` | Resize/Convert image. |

### 2.5 Auth & Jobs

| Slot | Signature | Description |
| :--- | :--- |
| `auth.login`| `username: $u`, `password: $p`, `table: "users"`, `as: $token` | Login & Issue JWT. |
| `auth.middleware`| `do: {}` | Protect route. Injects `$auth`. |
| `auth.user` | `as: $user` | Get current user. |
| `auth.check`| `as: $bool` | Check login status. |
| `jwt.sign` | `claims: {...}`, `secret: "s"`, `as: $t` | Manually sign JWT. |
| `jwt.verify`| `token: $t`, `secret: "s"`, `as: $c` | Manually verify JWT. |
| `worker.config` | `value: ["queue1", "queue2"]` | Configure worker queues. |
| `job.enqueue` | `queue: "q"`, `payload: {...}` | Push background job. |
| `mail.send` | `to: $e`, `subject: "s"`, `body: "b"`, `host: "smtp"`, `as: $ok` | Send SMTP email. |

---

## 3. High-Confidence Code Patterns

### 3.1 Standard CRUD (Create)
**Pattern**: Validation -> Check -> Insert -> Redirect/Respond.

```javascript
http.post: "/users/store" {
  do: {
    // 1. Validation
    validate: $form {
      rules: {
        name: "required|min:3"
        email: "required|email"
      }
      as: $errors
    }
    
    // 2. Error Handling
    if: $errors_any {
      then: { http.validation_error: { errors: $errors } }
      else: {
        // 3. Database Insert
        db.table: users
        db.insert: {
          name: $form.name
          email: $form.email
          created_at: date.now
        }
        
        // 4. Redirect (Standard UX)
        http.redirect: "/users"
      }
    }
  }
}
```

### 3.2 Parameterized Routes (CHI COMPLIANCE)
**Critical**: Use `{variable}` syntax. Variables are auto-injected.

```javascript
// CORRECT
http.get: "/users/{id}/edit" {
  do: {
    // $id is available automatically
    db.table: users
    db.where: { col: id, val: $id }
    db.first: { as: $user }
    
    // Check existence
    if: $user {
      then: { view.blade: "users.edit" { data: $user } }
      else: { http.not_found: { message: "User not found" } }
    }
  }
}

// INCORRECT (Do NOT use)
// http.get: "/users/:id/edit" 
```

### 3.3 Financial Calculations
**Critical**: Use `money.calc` for currency to avoid floating point errors.

```javascript
money.calc: ($price * $quantity) - $discount { as: $total }
```
