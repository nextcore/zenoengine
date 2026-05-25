# Standard Library API Reference

This document is auto-generated from the ZenoEngine source code. It contains the complete reference for all built-in ZenoLang slots.

## Array

### `array.join`

No description available.

**Example:**
```zeno
array.join: $tags
  sep: ', '
  as: $tag_str
```

---

### `array.pop`

No description available.

**Example:**
```zeno
array.pop: $stack
  as: $item
```

---

### `array.push`

Menambahkan elemen baru ke akhir array.

**Example:**
```zeno
array.push: $my_list
  val: 'New Item'
```

---

## Arrays

### `arrays.length`

Mengambil jumlah elemen dalam sebuah array atau list.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | **Yes** | Variabel penyimpan hasil |

**Example:**
```zeno
arrays.length: $users
  as: $count
```

---

## Aspnet

### `aspnet.hash`

Hash a plain-text password using the legacy ASP.NET Identity V3 PBKDF2/HMAC-SHA256 format.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | The plain-text password to hash |
| `as` | `string` | No | Variable name to store the generated hash result (Default: 'hash_result') |
| `iterations` | `int` | No | Iteration count (Default: 10000) |
| `password` | `string` | No | The plain-text password to hash |

**Example:**
```zeno
aspnet.hash: $input_pass
  iterations: 10000
  as: $db_hash
```

---

### `aspnet.login`

Authenticate user using legacy ASP.NET Core Identity AspNetUsers table schema and PBKDF2 hashing.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable to store the JWT token (Default: 'token') |
| `db` | `string` | No | Database connection name (Default: 'default') |
| `expires_in` | `int` | No | Token expiration time in seconds (Default: 86400) |
| `fields` | `list` | No | List of custom database columns to retrieve from AspNetUsers |
| `password` | `string` | **Yes** | Plain-text password |
| `secret` | `string` | No | JWT secret key for signing |
| `user_as` | `string` | No | Variable to store the user data map (Default: 'user') |
| `username` | `string` | **Yes** | Username or Email address of the user |

**Example:**
```zeno
aspnet.login:
  username: $input_user
  password: $input_pass
  fields: ['TenantId', 'FullName']
  as: $token
  user_as: $user
```

---

### `aspnet.validate_password`

Validate a password against configurable ASP.NET Core Identity password policies.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | The plain-text password to validate |
| `as` | `string` | No | Variable name to store validation status (Default: 'is_valid') |
| `errors_as` | `string` | No | Variable name to store list of error messages (Default: 'password_errors') |
| `password` | `string` | No | The plain-text password to validate |
| `require_digit` | `bool` | No | Require at least one digit (Default: true) |
| `require_lowercase` | `bool` | No | Require at least one lowercase letter (Default: true) |
| `require_non_alphanumeric` | `bool` | No | Require at least one special character (Default: true) |
| `require_uppercase` | `bool` | No | Require at least one uppercase letter (Default: true) |
| `required_length` | `int` | No | Minimum length of the password (Default: 6) |
| `required_unique_chars` | `int` | No | Minimum number of unique characters (Default: 1) |

**Example:**
```zeno
aspnet.validate_password: $password
  required_length: 8
  require_uppercase: true
  errors_as: $pw_errors
```

---

### `aspnet.verify`

Verify a plain-text password against an ASP.NET Identity V3 PBKDF2/HMAC-SHA256 hash.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable name to store the boolean result (Default: 'verify_result') |
| `hash` | `string` | **Yes** | The ASP.NET Identity V3 hash string (base64 encoded) |
| `password` | `string` | **Yes** | The plain-text password to verify |

**Example:**
```zeno
aspnet.verify
  hash: $db_hash
  password: $input_pass
  as: $is_valid
```

---

## Auth

### `auth.check`

Check if user is logged in (returns boolean).

---

### `auth.login`

Verify user credentials and return a JWT token.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variable to store token |
| `col_pass` | `any` | No | Password column (Default: 'password') |
| `col_user` | `any` | No | Email/Username column (Default: 'email') |
| `db` | `any` | No | Database connection name (Default: 'default') |
| `email` | `any` | No | Alias for username |
| `password` | `any` | **Yes** | Password |
| `secret` | `any` | No | JWT Secret key |
| `table` | `any` | No | User table name (Default: 'users') |
| `username` | `any` | No | Email or Username |

**Example:**
```zeno
auth.login
  username: $user
  password: $pass
  as: $token
```

---

### `auth.middleware`

Protect routes with JWT verification. Supports multi-tenant with subdomain detection.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `redirect` | `any` | No | Login URL for redirect on failure |
| `secret` | `any` | No | JWT Secret key |
| `set_auth_object` | `any` | No | Set $auth object with user_id, email, etc |
| `tenant_db_lookup` | `any` | No | Enable tenant validation from system DB |
| `tenant_header` | `any` | No | Header name for tenant ID (fallback to subdomain) |

**Example:**
```zeno
auth.middleware {
  do: {
     log: 'Hello Admin'
  }
}

// Multi-tenant:
auth.middleware {
  tenant_header: "X-Tenant-ID"
  tenant_db_lookup: true
  set_auth_object: true
  do: { ... }
}
```

---

### `auth.user`

Retrieve user data from current session.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variable to store user data |

**Example:**
```zeno
auth.user: $current_user
```

---

## Blade

### `blade.render_string`

Renders a blade template string and saves HTML to scope

---

## Cache

### `cache.forget`

No description available.

**Example:**
```zeno
cache.forget: 'homepage_stats'
```

---

### `cache.get`

Mengambil data cache. Always returns default value (cache disabled).

**Example:**
```zeno
cache.get
  key: "homepage_stats"
  default: 0
  as: $stats
```

---

### `cache.put`

Menyimpan data sementara (Cache). Currently disabled.

**Example:**
```zeno
cache.put
  key: "homepage_stats"
  val: stats_data
  ttl: "30m"
```

---

## Captcha

### `captcha.image`

Menulis gambar PNG captcha ke http.ResponseWriter atau menyimpan bytes ke scope.

**Example:**
```zeno
captcha.image
  id: $captcha_id
  width: 240
  height: 80
```

---

### `captcha.new`

Membuat captcha baru dan menyimpan ID-nya ke scope.

**Example:**
```zeno
captcha.new
  as: $captcha_id
```

---

### `captcha.serve`

Mendaftarkan route handler captcha ke router. Melayani PNG dan WAV secara otomatis.

**Example:**
```zeno
captcha.serve
  prefix: /captcha
```

---

### `captcha.verify`

Memverifikasi jawaban user terhadap captcha ID. Menghapus captcha setelah verifikasi.

**Example:**
```zeno
captcha.verify
  id: $captcha_id
  answer: $user_input
  as: $is_valid
```

---

## Cast

### `cast.to_int`

Mengubah variabel menjadi Integer.

**Example:**
```zeno
cast.to_int: $id { as: $id_int }
```

---

## Collections

### `collections.get`

No description available.

**Example:**
```zeno
collections.get: $list { index: 0; as: $item }
```

---

## Cookie

### `cookie.set`

No description available.

**Example:**
```zeno
cookie.set
  name: 'token'
  val: $token
```

---

## Core / General

### `__native_write`

Internal write for native blade

---

### `__native_write_safe`

Internal safe write for native blade

---

### `auth`

Execute block if user is logged in.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |

---

### `break`

Force stop. Supports conditional: `break: $i == 5`

---

### `call`

Memanggil fungsi yang didefinisikan dengan fn.

**Example:**
```zeno
call: hitung_gaji
```

---

### `can`

Execute block if user has specific permission (ability).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |
| `resource` | `any` | No | Object to check permission for |

---

### `cannot`

Menjalankan blok jika user TIDAK memiliki izin (ability).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Blok kode yang dijalankan |
| `resource` | `any` | No | Objek yang dicek izinnya |

---

### `coalesce`

Mengembalikan nilai default jika input bernilai null.

**Example:**
```zeno
coalesce: $user.name { default: 'Guest'; as: $name }
```

---

### `continue`

Continue to next iteration. Supports conditional: `continue: $i % 2 == 0`

---

### `dd`

Dump and Die. Display variable content and stop script immediately.

---

### `down`

Execute child blocks only during the rollback ('down') migration process.

**Example:**
```zeno
down {
  db.drop_table: 'users'
}
```

---

### `dump`

Dump variable to console without stopping execution.

---

### `empty`

Execute block if variable is empty (null, '', or empty array).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |

---

### `error`

Menampilkan pesan error validasi untuk field tertentu.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Blok kode yang dijalankan |

---

### `fn`

Mendefinisikan fungsi (menyimpan blok kode untuk dipanggil nanti).

**Example:**
```zeno
fn: hitung_gaji {
  ...
}
```

---

### `for`

Iterate (loop) over a list or array.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `__native_write` | `any` | No | Internal Blade attribute |
| `as` | `any` | No | Variable name for current element (Default: 'item') |
| `do` | `any` | No | Code block to repeat |

**Required Blocks:** `do`

**Example:**
```zeno
for: $list
  as: $item
  do: ...
```

---

### `foreach`

No description available.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `__native_write` | `any` | No | Internal Blade attribute |
| `as` | `any` | No | Variable name for current element (Default: 'item') |
| `do` | `any` | No | Code block to repeat |

**Example:**
```zeno
foreach: $list { as: $item ... }
```

---

### `forelse`

Perulangan list dengan blok cadangan jika list kosong.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `__native_write` | `any` | No | Internal Blade attribute |
| `as` | `any` | No | Alias variabel item |
| `do` | `any` | No | Blok yang diulang |
| `empty` | `any` | No | Blok jika data kosong (Legacy) |
| `forelse_empty` | `any` | No | Blok jika data kosong |

---

### `guest`

Execute block if user is NOT logged in (guest).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |

---

### `if`

Kondisional if-then-else. Support: ==, !=, >, <, >=, <=

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `else` | `any` | No | Blok kode jika kondisi salah |
| `then` | `any` | No | Blok kode jika kondisi benar |

**Required Blocks:** `then`

---

### `include`

No description available.

---

### `is_null`

No description available.

---

### `isset`

Execute block if variable is set/defined.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |

---

### `json`

Mengeluarkan nilai sebagai JSON langsung ke HTTP response.

---

### `len`

No description available.

**Example:**
```zeno
len: $my_list { as: $count }
```

---

### `log`

No description available.

**Example:**
```zeno
log: $user.name
```

---

### `loop`

While loop

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |

**Required Blocks:** `do`

---

### `return`

Halt execution of the current block/handler.

---

### `schema`

Memvalidasi tipe data variabel yang sudah ada.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `type` | `string` | **Yes** | Tipe data yang diharapkan |

**Example:**
```zeno
schema: $user_id { type: 'int' }
```

---

### `sleep`

Menghentikan eksekusi selama beberapa milidetik.

**Main Value Type:** `int`

**Example:**
```zeno
sleep: 1000
```

---

### `switch`

Conditional branching (Switch Case).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `case` | `any` | No | Case value to check |
| `default` | `any` | No | Default block if no case matches |

---

### `try`

Handle errors using a try-catch block.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variable name for error message (Default: 'error') |
| `catch` | `any` | No | Error handling code block |
| `do` | `any` | No | Main code block to execute |

**Example:**
```zeno
try {
  do: { ... }
  catch: { ... }
}
```

---

### `unless`

Reverse of IF. Execute block if condition is FALSE.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |

---

### `up`

Execute child blocks only during the forward ('up') migration process.

**Example:**
```zeno
up {
  db.create_table: 'users' { ... }
}
```

---

### `validate`

No description available.

**Example:**
```zeno
validate: $form
  rules:
    email: "required|email|unique:users,email"
    password: "required|confirmed|min:8"
    role: "in:admin,user"
  as: $errs
  as_safe: $valid_data
```

---

### `var`

Standard variable definition/assignment slot.

**Example:**
```zeno
var: $user { val: $data }
```

---

### `while`

While loop

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |

**Required Blocks:** `do`

---

## Crypto

### `crypto.hash`

Hash a plain-text password using bcrypt (cost: 10).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | Shorthand input for the password string to hash |
| `as` | `string` | No | Variable name to store the generated hash result (Default: 'hash_result') |
| `text` | `string` | No | Alternative parameter to provide the password string |
| `val` | `string` | No | Alternative parameter to provide the password string |

**Example:**
```zeno
crypto.hash: $pass
  as: $hashed
```

---

### `crypto.verify`

Verify a plain-text password against a bcrypt hash.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable name to store the boolean result (Default: 'verify_result') |
| `hash` | `string` | **Yes** | The bcrypt hash string to compare against |
| `text` | `string` | **Yes** | The plain-text password to verify |

**Example:**
```zeno
crypto.verify
  hash: $h
  text: $p
  as: $is_valid
```

---

## Ctx

### `ctx.timeout`

Limit execution time of a code block.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `do` | `any` | No | Code block to execute |
| `duration` | `any` | No | Timeout duration (e.g., '5s', '1m') |

**Example:**
```zeno
ctx.timeout: '5s' {
  do: { ... }
}
```

---

## Date

### `date.add`

Menambah durasi ke tanggal.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `add` | `any` | No | Alias untuk duration |
| `as` | `any` | No | Variabel penyimpan hasil |
| `date` | `any` | No | Objek tanggal sumber |
| `duration` | `any` | No | Durasi (1h, 30m, 10s) |
| `val` | `any` | No | Alias untuk date |

**Example:**
```zeno
date.add: $now { duration: '2h'; as: $future }
```

---

### `date.format`

Memformat objek tanggal atau string tanggal.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variabel penyimpan hasil |
| `date` | `any` | No | Alias untuk val |
| `format` | `any` | No | Alias untuk layout |
| `layout` | `any` | No | Format tujuan |
| `val` | `any` | No | Objek atau string tanggal |

**Example:**
```zeno
date.format: $created_at { layout: 'Human'; as: $tgl }
```

---

### `date.now`

Mengambil waktu saat ini.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variabel penyimpan hasil string |
| `format` | `any` | No | Alias untuk layout |
| `layout` | `any` | No | Format tanggal (RFC3339, Human, dll) |

**Example:**
```zeno
date.now: { as: $skarang }
```

---

### `date.parse`

Mengubah string menjadi objek tanggal.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variabel penyimpan hasil |
| `format` | `any` | No | Alias untuk layout |
| `input` | `any` | No | String tanggal |
| `layout` | `any` | No | Format sumber |
| `val` | `any` | No | Alias untuk input |

**Example:**
```zeno
date.parse: '2023-12-25' { as: $tgl_obj }
```

---

## Db

### `db.avg`

Calculate the AVG (average) of a specific column based on the query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The column name to average |
| `as` | `string` | **Yes** | Variable name to store the result |

**Example:**
```zeno
db.avg: 'rating'
  as: $average
```

---

### `db.boolean`

Add a boolean column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.boolean: 'column_name'
```

---

### `db.columns`

Specify the column(s) to retrieve in the query. Can be a single string or an array of strings.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `any` | No | A single column name or array of column names |

**Example:**
```zeno
db.columns: ['id', 'name']
```

---

### `db.count`

Count the number of rows based on the current query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | **Yes** | Variable name to store result |

**Example:**
```zeno
db.count
  as: $total
```

---

### `db.create_table`

Create a new database table using a fluent schema building definition.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the table to create |
| `db` | `string` | No | The database connection name (Default: 'default') |

**Main Value Type:** `string`

**Example:**
```zeno
db.create_table: 'posts' {
  db.id: 'id'
  db.string: 'title'
  db.text: 'body'
}
```

---

### `db.date`

Add a date column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.date: 'column_name'
```

---

### `db.decimal`

Add a decimal column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `precision` | `int` | No | Total number of digits (Default: 10) |
| `scale` | `int` | No | Number of digits to the right of decimal point (Default: 2) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.decimal: 'price' {
  precision: 12
  scale: 4
}
```

---

### `db.delete`

Perform a DELETE database operation based on query where constraints.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable to store the count of deleted rows |

**Example:**
```zeno
db.delete
  as: $count
```

---

### `db.doesnt_exist`

Check if no rows exist based on the current query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | **Yes** | Variable name to store the boolean result |

**Example:**
```zeno
db.doesnt_exist
  as: $is_empty
```

---

### `db.drop_table`

Drop a database table if it exists.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the table to drop |
| `db` | `string` | No | The database connection name (Default: 'default') |

**Main Value Type:** `string`

**Example:**
```zeno
db.drop_table: 'users'
```

---

### `db.execute`

Execute a raw SQL query (INSERT, UPDATE, DELETE, etc.).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bind` | `any` | No | Bind parameters container |
| `db` | `string` | No | Database connection name |

**Main Value Type:** `string`

**Example:**
```zeno
db.execute: 'UPDATE users SET x=1'
```

---

### `db.exists`

Check if at least one row exists based on the current query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | **Yes** | Variable name to store the boolean result |

**Example:**
```zeno
db.exists
  as: $has_users
```

---

### `db.first`

Retrieve the first row from the database based on the current query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | **Yes** | Variable name to store result |

**Example:**
```zeno
db.first
  as: $user
```

---

### `db.foreign`

Define a foreign key constraint for a column inside a db.create_table block.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The local column name to apply the foreign key to |
| `on` | `string` | No | The referenced parent column name (Default: 'id') |
| `on_delete` | `string` | No | Action on parent record deletion (e.g., CASCADE, SET NULL) |
| `references` | `string` | **Yes** | The parent table name reference |

**Main Value Type:** `string`

**Example:**
```zeno
db.foreign: 'user_id' {
  references: 'users'
  on: 'id'
  on_delete: 'CASCADE'
}
```

---

### `db.get`

Retrieve multiple rows from the database based on the current query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | **Yes** | Variable name to store results |

**Example:**
```zeno
db.get
  as: $users
```

---

### `db.group_by`

Add a GROUP BY clause to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | Column name to group by |

**Example:**
```zeno
db.group_by: 'status'
```

---

### `db.having`

Add a HAVING clause filter to the query (typically used with GROUP BY).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `col` | `string` | **Yes** | Column or aggregate field name |
| `op` | `string` | No | Comparison operator (Default: '>') |
| `val` | `any` | **Yes** | Value to compare against |

**Example:**
```zeno
db.having {
  col: 'count'
  op: '>'
  val: 5
}
```

---

### `db.hook`

Register lifecycle hooks for a database table (before/after insert, update, delete, save).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `after_delete` | `any` | No | Code block executed after a DELETE |
| `after_insert` | `any` | No | Code block executed after an INSERT |
| `after_save` | `any` | No | Code block executed after INSERT or UPDATE |
| `after_update` | `any` | No | Code block executed after an UPDATE |
| `before_delete` | `any` | No | Code block executed before a DELETE |
| `before_insert` | `any` | No | Code block executed before an INSERT |
| `before_save` | `any` | No | Code block executed before INSERT or UPDATE |
| `before_update` | `any` | No | Code block executed before an UPDATE |

**Example:**
```zeno
db.hook: 'posts' {
  before_insert: {
    var: $data.slug slug($data.title)
  }
  after_save: {
    cache.forget: "posts_list"
  }
  after_update: {
    db.insert: activity_log { action: "updated" table: "posts" }
  }
}
```

---

### `db.id`

Add an auto-incrementing primary key column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.id: 'id'
```

---

### `db.insert`

Perform an INSERT database operation. Insert data specified in the children block.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `*(any)` | `any` | No | Column name and the value to insert |

**Example:**
```zeno
db.insert
  name: 'John Doe'
  email: 'john@example.com'
```

---

### `db.integer`

Add a integer column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.integer: 'column_name'
```

---

### `db.join`

Perform an INNER JOIN operation with another table.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `on` | `list` | **Yes** | Array representing ['left_col', 'operator', 'right_col'] |
| `table` | `string` | **Yes** | The table to join |

**Example:**
```zeno
db.join {
  table: 'posts'
  on: ['users.id', '=', 'posts.user_id']
}
```

---

### `db.json`

Add a json column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.json: 'column_name'
```

---

### `db.last`

Retrieve the last row (ordered by 'id DESC') from the database.

**Example:**
```zeno
db.last
  as: $user
```

---

### `db.left_join`

Perform a LEFT OUTER JOIN operation with another table.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `on` | `list` | **Yes** | Array representing ['left_col', 'operator', 'right_col'] |
| `table` | `string` | **Yes** | The table to join |

**Example:**
```zeno
db.left_join {
  table: 'posts'
  on: ['users.id', '=', 'posts.user_id']
}
```

---

### `db.limit`

Set a LIMIT on the number of rows retrieved in the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `int` | **Yes** | Maximum number of rows to retrieve |

**Example:**
```zeno
db.limit: 10
```

---

### `db.max`

Retrieve the MAX (maximum value) of a specific column based on the query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The column name to find the maximum of |
| `as` | `string` | **Yes** | Variable name to store the result |

**Example:**
```zeno
db.max: 'age'
  as: $oldest
```

---

### `db.min`

Retrieve the MIN (minimum value) of a specific column based on the query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The column name to find the minimum of |
| `as` | `string` | **Yes** | Variable name to store the result |

**Example:**
```zeno
db.min: 'age'
  as: $youngest
```

---

### `db.offset`

Set an OFFSET to skip a number of rows in the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `int` | **Yes** | Number of rows to skip |

**Example:**
```zeno
db.offset: 20
```

---

### `db.or_where`

Add an OR WHERE filter constraint to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `col` | `string` | **Yes** | Column name |
| `op` | `string` | No | Comparison operator (Default: '=') |
| `val` | `any` | **Yes** | Filter value |

**Example:**
```zeno
db.or_where
  col: role
  val: 'admin'
```

---

### `db.order_by`

Add an ORDER BY sorting clause to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | Sorting expression (e.g. 'created_at DESC') |

**Example:**
```zeno
db.order_by: 'id DESC'
```

---

### `db.paginate`

Retrieve rows paginated with metadata.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | **Yes** | Variable name to store the paginator object containing data and meta |
| `page` | `int` | No | Current page number (Default: 1) |
| `per_page` | `int` | No | Number of rows per page (Default: 15) |

**Example:**
```zeno
db.paginate
  page: 1
  per_page: 20
  as: $users_paginator
```

---

### `db.pluck`

Retrieve a single column's values as a flat array.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | The column name to pluck |
| `as` | `string` | **Yes** | Variable name to store the array result |
| `col` | `string` | No | Alias for column name |

**Example:**
```zeno
db.pluck: 'id'
  as: $user_ids
```

---

### `db.query`

Alias for db.select

---

### `db.right_join`

Perform a RIGHT OUTER JOIN operation with another table.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `on` | `list` | **Yes** | Array representing ['left_col', 'operator', 'right_col'] |
| `table` | `string` | **Yes** | The table to join |

**Example:**
```zeno
db.right_join {
  table: 'posts'
  on: ['users.id', '=', 'posts.user_id']
}
```

---

### `db.seed`

Execute database seeders.

---

### `db.select`

Perform a SELECT query and retrieve multiple rows.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable to store results |
| `bind` | `any` | No | Bind parameters container |
| `db` | `string` | No | Database connection name |
| `first` | `bool` | No | Return only the first row as a map (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.select: 'SELECT * FROM users'
  as: $users
```

---

### `db.string`

Add a string column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `length` | `int` | No | Alias for limit |
| `limit` | `int` | No | Maximum length of the string (Default: 255) |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.string: 'name' {
  limit: 100
  unique: true
}
```

---

### `db.sum`

Calculate the SUM of a specific column based on the query state.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The column name to sum |
| `as` | `string` | **Yes** | Variable name to store the result |

**Example:**
```zeno
db.sum: 'price'
  as: $total_price
```

---

### `db.table`

Set the table to be used for subsequent database operations.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `db` | `any` | No | Database connection name (Default: 'default') |
| `name` | `any` | No | Table name (Optional if specified in main value) |

**Example:**
```zeno
db.table: 'users'
```

---

### `db.text`

Add a text column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.text: 'column_name'
```

---

### `db.timestamp`

Add a timestamp column to the table schema.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The name of the column |
| `nullable` | `bool` | No | Whether the column allows NULL values (Default: true) |
| `unique` | `bool` | No | Whether the column values must be unique (Default: false) |

**Main Value Type:** `string`

**Example:**
```zeno
db.timestamp: 'column_name'
```

---

### `db.update`

Perform an UPDATE database operation. Update columns specified in the children block based on query where constraints.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `*(any)` | `any` | No | Column name and the new value to update |

**Example:**
```zeno
db.update
  status: 'active'
  updated_at: 'NOW()'
```

---

### `db.where`

Add a WHERE filter to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `col` | `any` | No | Column name |
| `op` | `any` | No | Operator (Default: '=') |
| `val` | `any` | No | Filter value |

**Example:**
```zeno
db.where
  col: id
  val: $user_id
```

---

### `db.where_between`

Add a WHERE BETWEEN constraint to filter a range of values.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `col` | `string` | **Yes** | Column name |
| `val` | `list` | **Yes** | Array representing the lower and upper bounds: [min, max] |

**Example:**
```zeno
db.where_between
  col: age
  val: [18, 30]
```

---

### `db.where_in`

Add an 'AND WHERE IN' filter constraint to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `col` | `string` | **Yes** | Column name |
| `val` | `list` | **Yes** | Array or slice of allowed values |

**Example:**
```zeno
db.where_in {
  col: 'status'
  val: ['active', 'pending']
}
```

---

### `db.where_not_between`

Add a WHERE NOT BETWEEN constraint to exclude a range of values.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `col` | `string` | **Yes** | Column name |
| `val` | `list` | **Yes** | Array representing the lower and upper bounds: [min, max] |

**Example:**
```zeno
db.where_not_between
  col: age
  val: [18, 30]
```

---

### `db.where_not_in`

Add an 'AND WHERE NOT IN' filter constraint to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `col` | `string` | **Yes** | Column name |
| `val` | `list` | **Yes** | Array or slice of values to exclude |

**Example:**
```zeno
db.where_not_in {
  col: 'role'
  val: ['admin', 'moderator']
}
```

---

### `db.where_not_null`

Add a 'WHERE column IS NOT NULL' constraint to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The column name to check |

**Example:**
```zeno
db.where_not_null: 'created_at'
```

---

### `db.where_null`

Add a 'WHERE column IS NULL' constraint to the query.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | **Yes** | The column name to check |

**Example:**
```zeno
db.where_null: 'deleted_at'
```

---

## Docker

### `docker.call`

Call an external docker container microservice with resilience (retry & circuit breaker).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variable to store result |
| `circuit_breaker` | `any` | No | Enable circuit breaker protection (default false) |
| `endpoint` | `any` | No | HTTP Path (default /) |
| `method` | `any` | No | HTTP Method (default POST) |
| `payload` | `any` | No | JSON array/object to send |
| `port` | `any` | No | Port (default 80) |
| `retry` | `any` | No | Number of retries on failure (default 0) |
| `timeout` | `any` | No | Timeout in ms (default 15000) |

**Example:**
```zeno
docker.call: 'php_worker' {
  endpoint: '/calculate'
  payload: { data: 1 }
  retry: 3
  circuit_breaker: true
  as: $res
}
```

---

### `docker.health`

Check health of a docker sidecar HTTP service.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variable to store result |
| `host` | `any` | No | Hostname of the container |
| `port` | `any` | No | Port of the container |

**Example:**
```zeno
docker.health: 'php_worker' {
  port: 8000
  as: $h
}
```

---

### `docker.nodes`

Register a pool of nodes for a service name (Load Balancing).

**Example:**
```zeno
docker.nodes: 'payment_service' {
  nodes: ['10.0.0.1', '10.0.0.2']
  check: '/status'
}
```

---

## Engine

### `engine.slots`

Returns documentation metadata for all registered ZenoLang slots.

**Example:**
```zeno
engine.slots: { as: $docs }
```

---

## Hash

### `hash.make`

Hash a plain-text password using bcrypt (cost: 10).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | Shorthand input for the password string to hash |
| `as` | `string` | No | Variable name to store the generated hash result (Default: 'hash_result') |
| `text` | `string` | No | Alternative parameter to provide the password string |
| `val` | `string` | No | Alternative parameter to provide the password string |

**Example:**
```zeno
crypto.hash: $pass
  as: $hashed
```

---

### `hash.verify`

Verify a plain-text password against a bcrypt hash.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable name to store the boolean result (Default: 'verify_result') |
| `hash` | `string` | **Yes** | The bcrypt hash string to compare against |
| `text` | `string` | **Yes** | The plain-text password to verify |

**Example:**
```zeno
crypto.verify
  hash: $h
  text: $p
  as: $is_valid
```

---

## Http

### `http.accepted`

Send 202 Accepted response

**Example:**
```zeno
http.accepted: {
  message: "Request accepted"
}
```

---

### `http.bad_request`

Send 400 Bad Request response

**Example:**
```zeno
http.bad_request: {
  message: "Invalid parameters"
  errors: $errors
}
```

---

### `http.body`

No description available.

**Example:**
```zeno
http.body { as: $raw }
```

---

### `http.created`

Send 201 Created response

**Example:**
```zeno
http.created: {
  message: "Resource created"
  id: $db_last_id
}
```

---

### `http.delete`

No description available.

---

### `http.forbidden`

Send 403 Forbidden response

**Example:**
```zeno
http.forbidden: {
  message: "Access denied"
}
```

---

### `http.form`

No description available.

**Example:**
```zeno
http.form: 'email'
  as: $email
```

---

### `http.get`

No description available.

---

### `http.group`

No description available.

---

### `http.header`

No description available.

**Example:**
```zeno
http.header: 'X-Tenant-ID'
  as: $tenant_id
```

---

### `http.host`

No description available.

**Example:**
```zeno
http.host: { as: $host }
```

---

### `http.json_body`

No description available.

**Example:**
```zeno
http.json_body { as: $data }
```

---

### `http.middleware`

Mendefinisikan middleware kustom menggunakan ZenoLang.

**Example:**
```zeno
http.middleware: 'auth' {
  do: {
    session.get: 'user_id' { as: $uid }
    if: $uid == null { then: { http.redirect: '/login' } }
  }
}
```

---

### `http.next`

Melanjutkan ke handler berikutnya dalam rantai middleware.

---

### `http.no_content`

Send 204 No Content response

**Example:**
```zeno
http.no_content
```

---

### `http.not_found`

Send 404 Not Found response

**Example:**
```zeno
http.not_found: {
  message: "Resource not found"
}
```

---

### `http.ok`

Send 200 OK response with auto JSON wrapping

**Example:**
```zeno
http.ok: {
  data: $posts
}
```

---

### `http.patch`

No description available.

---

### `http.post`

No description available.

---

### `http.proxy`

Meneruskan request ke backend service lain (Reverse Proxy).

**Example:**
```zeno
http.proxy: "http://localhost:8080"
  path: "/api"
```

---

### `http.put`

No description available.

---

### `http.query`

No description available.

**Example:**
```zeno
http.query: 'page'
  as: $page_param
```

---

### `http.redirect`

No description available.

**Example:**
```zeno
http.redirect: '/home'
```

---

### `http.request`

No description available.

**Example:**
```zeno
http.request: 'https://api.com'
  method: 'POST'
  body: $data
  as: $res
```

---

### `http.response`

No description available.

**Example:**
```zeno
http.response: 200
  body: $data
```

---

### `http.routes`

Mengambil daftar semua rute HTTP yang terdaftar di engine.

**Example:**
```zeno
http.routes: { as: $routes }
```

---

### `http.server_error`

Send 500 Internal Server Error response

**Example:**
```zeno
http.server_error: {
  message: "Internal error"
  error: $error
}
```

---

### `http.static`

Hosting aplikasi SPA (React/Vue) atau Static Site.

**Example:**
```zeno
http.static: "./dist"
  path: "/"
  spa: true
```

---

### `http.unauthorized`

Send 401 Unauthorized response

**Example:**
```zeno
http.unauthorized: {
  message: "Authentication required"
}
```

---

### `http.upload`

No description available.

**Example:**
```zeno
http.upload:
  field: image
  as: $new_file
```

---

### `http.validation_error`

Send 422 Validation Error response

**Example:**
```zeno
http.validation_error: {
  message: "Validation failed"
  errors: $errors
}
```

---

## Inertia

### `inertia.location`

Force a full page reload to a URL

**Example:**
```zeno
inertia.location:
  url: "/login"
```

---

### `inertia.render`

Render Inertia.js response

**Example:**
```zeno
inertia.render:
  component: "Dashboard"
  props: { user: $user }
```

---

### `inertia.share`

Share data across all Inertia requests

**Example:**
```zeno
inertia.share:
  auth: $auth
  flash: $flash
```

---

## Io

### `io.dir.create`

No description available.

---

### `io.file.delete`

No description available.

**Example:**
```zeno
io.file.delete: $path
```

---

### `io.file.read`

No description available.

**Example:**
```zeno
io.file.read: $path
  as: $content
```

---

### `io.file.write`

No description available.

---

## Job

### `job.enqueue`

Add a job to the background queue (Redis/DB).

**Example:**
```zeno
job.enqueue
  queue: "emails"
  payload:
    to: "budi@example.com"
    subject: "Welcome"
```

---

## Json

### `json.parse`

No description available.

**Example:**
```zeno
json.parse: $response_body
  as: $data
```

---

### `json.stringify`

No description available.

**Example:**
```zeno
json.stringify: $data
  as: $json_str
```

---

## Jwt

### `jwt.refresh`

Refresh JWT token with a new expiration.

**Example:**
```zeno
jwt.refresh: $old_token
  as: $new_token
```

---

### `jwt.sign`

Generate JWT token with custom claims

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variable to store token |
| `claims` | `any` | **Yes** | Token claims as map |
| `expires_in` | `any` | No | Expiry in seconds (default: 86400) |
| `secret` | `any` | **Yes** | JWT Secret key |

**Example:**
```zeno
jwt.sign:
  secret: env("JWT_SECRET")
  claims: { user_id: $user.id }
  expires_in: 86400
  as: $token
```

---

### `jwt.verify`

Explicitly verify a JWT token and retrieve its claims.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Resulting claims |
| `secret` | `any` | No | Secret Key |
| `token` | `any` | No | Token String |

**Example:**
```zeno
jwt.verify: $token
  secret: 'shhh'
  as: $user_data
```

---

## Logic

### `logic.compare`

Compare two values.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable to store result (Default: compare_result) |
| `op` | `string` | **Yes** | Comparison operator |
| `v1` | `any` | **Yes** | First value |
| `v2` | `any` | **Yes** | Second value |

**Example:**
```zeno
logic.compare
  v1: $age
  op: '>'
  v2: 17
```

---

## Mail

### `mail.send`

Send email natively via SMTP or in Mock Mode if no SMTP_HOST is configured.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable name to store the boolean send status (Default: 'mail_status') |
| `body` | `string` | No | Plain text email body |
| `html` | `string` | No | HTML email body |
| `subject` | `string` | **Yes** | Subject line of the email |
| `to` | `string/list` | **Yes** | Recipient email address (string or list of strings) |

**Example:**
```zeno
mail.send:
  to: 'user@example.com'
  subject: 'Welcome'
  body: 'Hello User'
  as: $sent
```

---

## Map

### `map.keys`

No description available.

**Example:**
```zeno
map.keys: $user
  as: $fields
```

---

### `map.set`

No description available.

**Example:**
```zeno
map.set: $user
  age: 30
```

---

## Math

### `math.calc`

Melakukan perhitungan matematika menggunakan ekspresi string.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variabel penyimpan hasil |
| `expr` | `string` | No | Alias untuk val |
| `val` | `string` | No | Ekspresi matematika (jika tidak via value utama) |

**Main Value Type:** `string`

**Example:**
```zeno
math.calc: ceil($total / 10)
  as: $pages
```

---

## Meta

### `meta.eval`

Evaluates a string as ZenoLang code dynamically.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | The ZenoLang code string to evaluate |

**Example:**
```zeno
meta.eval: "http.get: '/api'"
```

---

### `meta.parse`

Parses ZenoLang code into an AST Map (Code as Data).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | The ZenoLang code string to parse |
| `as` | `string` | No | Variable to store the AST map |

**Example:**
```zeno
$ast: meta.parse: "print: 'hello'"
```

---

### `meta.run`

Executes an AST Map as ZenoLang code.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `map` | No | The AST Map to execute |

**Example:**
```zeno
meta.run: $ast
```

---

### `meta.scope`

Returns all variables in the current scope as a map (Introspection).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable to store the scope map |

**Example:**
```zeno
$vars: meta.scope
```

---

### `meta.template`

Renders a Blade template into a string variable (useful for code generation).

**Example:**
```zeno
meta.template: 'codegen/route' { resource: 'users'; as: $code }
```

---

## Money

### `money.calc`

Melakukan perhitungan keuangan menggunakan Decimal untuk presisi tinggi.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variabel penyimpan hasil |
| `val` | `decimal` | No | Ekspresi keuangan |

**Main Value Type:** `decimal`

**Example:**
```zeno
money.calc: ($harga * $qty) - $diskon
  as: $total
```

---

## Mysql

### `mysql.execute`

Alias for db.execute

---

### `mysql.select`

Alias for db.select

---

## Orm

### `orm.belongsTo`

Define a many-to-one relationship.

---

### `orm.belongsToMany`

Define a many-to-many relationship.

---

### `orm.delete`

No description available.

---

### `orm.find`

Find a single record by primary key.

**Example:**
```zeno
orm.find: 1 { as: $user }
```

---

### `orm.hasMany`

Define a one-to-many relationship.

---

### `orm.hasOne`

Define a one-to-one relationship.

---

### `orm.model`

Define the active model/table for ORM operations.

**Example:**
```zeno
orm.model: 'users'
```

---

### `orm.save`

Save (Insert or Update) a model object.

**Example:**
```zeno
orm.save: $user
```

---

### `orm.with`

Eager load a relationship.

---

## Scope

### `scope.set`

Create a variable (Legacy alias for 'var').

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `key` | `string` | No | Variable name |
| `name` | `string` | No | Variable name (alias for key) |
| `val` | `any` | No | Variable value |
| `value` | `any` | No | Variable value (alias for val) |

**Example:**
```zeno
scope.set: $my_var
  val: 123
```

---

## Sec

### `sec.csrf_token`

Retrieve the CSRF token for the current HTTP request context.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `(value)` | `string` | No | Variable name to store the CSRF token (Default: 'csrf_token') |

**Example:**
```zeno
sec.csrf_token: $token
```

---

## Section

### `section.define`

Define a layout section

---

### `section.yield`

Yield a layout section

---

## Session

### `session.destroy`

Destroy all session data.

---

### `session.flash`

Flash data to the session (cookie) for the next request.

**Example:**
```zeno
session.flash: { key: 'error', val: 'Invalid credentials' }
```

---

### `session.get`

Get session data.

---

### `session.get_flash`

Retrieve flash data and remove it from session.

**Example:**
```zeno
session.get_flash: 'error' { as: $error_msg }
```

---

### `session.regenerate`

Regenerate session ID (Security).

---

### `session.set`

Set session data.

---

## Storage

### `storage.delete`

Delete a file from the storage system.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable name to store delete status (Default: 'storage_deleted') |
| `path` | `string` | **Yes** | Relative path of the file to delete |

**Example:**
```zeno
storage.delete: 'avatars/1.jpg' { as: $deleted }
```

---

### `storage.exists`

Check if a file exists in the storage system.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable name to store exists status (Default: 'storage_exists') |
| `path` | `string` | **Yes** | Relative path of the file to check |

**Example:**
```zeno
storage.exists: 'avatars/1.jpg' { as: $exists }
```

---

### `storage.put`

Save file content or copy an existing local file to the storage system.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `string` | No | Variable name to store the stored file path (Default: 'storage_path') |
| `content` | `string/bytes` | **Yes** | File content (string/bytes) or local source filepath to copy |
| `is_file_path` | `bool` | No | Whether the content should be treated as a filepath to copy from (Default: false) |
| `path` | `string` | **Yes** | Target relative path inside storage (e.g. 'images/user.png') |

**Example:**
```zeno
storage.put:
  content: $uploaded_temp_path
  path: 'avatars/1.jpg'
  is_file_path: true
  as: $file_url
```

---

## String

### `string.replace`

Mengganti substring dalam string dengan string lain.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variabel penyimpan hasil |
| `find` | `any` | **Yes** | Substring yang dicari |
| `limit` | `any` | No | Jumlah penggantian maksimum (-1 untuk semua) |
| `replace` | `any` | **Yes** | Substring pengganti |

**Example:**
```zeno
string.replace: $text
  find: 'old'
  replace: 'new'
  as: $result
```

---

## Strings

### `strings.concat`

Menggabungkan beberapa string menjadi satu secara fleksibel.

**Example:**
```zeno
strings.concat: 'Hello '
  val: $name
  as: $greeting
```

---

## System

### `system.args`

Mengambil argument command line yang dilewatkan ke script.

**Example:**
```zeno
system.args: { as: $my_args }
```

---

### `system.env`

No description available.

---

## Text

### `text.sanitize`

Membersihkan teks dari tag HTML berbahaya (XSS prevention).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variabel penyimpan hasil |
| `input` | `any` | No | Teks sumber |
| `val` | `any` | No | Alias untuk input |

**Example:**
```zeno
text.sanitize: $user_input
  as: $clean_input
```

---

### `text.slugify`

Mengubah teks menjadi format URL-friendly slug.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | No | Variabel penyimpan hasil |
| `text` | `any` | No | Teks sumber |
| `val` | `any` | No | Alias untuk text |

**Example:**
```zeno
text.slugify: 'Halo Dunia'
  as: $my_slug
```

---

## Time

### `time.sleep`

Pause execution for a duration.

**Example:**
```zeno
time.sleep: '1s'
```

---

## Validator

### `validator.validate`

No description available.

**Example:**
```zeno
validate: $form
  rules:
    email: "required|email|unique:users,email"
    password: "required|confirmed|min:8"
    role: "in:admin,user"
  as: $errs
  as_safe: $valid_data
```

---

## View

### `view.blade`

Render Blade natively using ZenoLang AST.

---

### `view.class`

Render Blade @class

---

### `view.component`

Render a Blade Component

---

### `view.extends`

Extend a layout

---

### `view.include`

Include a partial view

---

### `view.push`

Push content to stack

---

### `view.root`

Sets the base directory for Blade views for this app/module.

**Example:**
```zeno
view.root: 'apps/blog/resources/views'
```

---

### `view.stack`

Render stack content

---

## Worker

### `worker.config`

Configure worker queues.

**Example:**
```zeno
worker.config
  - "high_priority"
  - "default"
```

---

