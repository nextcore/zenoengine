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

Retrieve current logged-in user data.

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

## Column

### `column.boolean`

No description available.

---

### `column.datetime`

No description available.

---

### `column.id`

No description available.

---

### `column.integer`

No description available.

---

### `column.string`

No description available.

---

### `column.text`

No description available.

---

### `column.timestamps`

No description available.

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

Membuat atau mengubah nilai variabel dalam scope saat ini.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `type` | `string` | No | Tipe data (opsional) |
| `val` | `any` | No | Nilai variabel |
| `value` | `any` | No | Alias untuk val |

**Example:**
```zeno
var: $count
  val: 10
  type: 'int'
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

No description available.

**Example:**
```zeno
crypto.hash: $pass
  as: $hashed
```

---

### `crypto.verify`

No description available.

**Example:**
```zeno
crypto.verify
  hash: $h
  text: $p
```

---

### `crypto.verify_aspnet`

No description available.

**Example:**
```zeno
crypto.verify_aspnet
  hash: $db_hash
  password: $input_pass
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

No description available.

**Example:**
```zeno
db.avg: 'rating'
  as: $average
```

---

### `db.columns`

No description available.

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

### `db.delete`

No description available.

**Example:**
```zeno
db.delete
  as: $count
```

---

### `db.doesnt_exist`

Check if no rows exist based on the current query state.

**Example:**
```zeno
db.doesnt_exist
  as: $is_empty
```

---

### `db.execute`

Execute a raw SQL query (INSERT, UPDATE, DELETE, etc.).

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `bind` | `any` | No | Bind parameters container |
| `db` | `string` | No | Database connection name |
| `params` | `list` | No | List of bind values |
| `val` | `any` | No | Single bind value |

**Main Value Type:** `string`

**Example:**
```zeno
db.execute: 'UPDATE users SET x=1'
```

---

### `db.exists`

Check if at least one row exists based on the current query state.

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

No description available.

**Example:**
```zeno
db.group_by: 'status'
```

---

### `db.having`

No description available.

**Example:**
```zeno
db.having {
  col: count
  op: '>'
  val: 5
}
```

---

### `db.insert`

No description available.

**Example:**
```zeno
db.insert
  name: $name
```

---

### `db.join`

No description available.

**Example:**
```zeno
db.join {
  table: posts
  on: ['users.id', '=', 'posts.user_id']
}
```

---

### `db.left_join`

No description available.

**Example:**
```zeno
db.left_join ...
```

---

### `db.limit`

No description available.

**Example:**
```zeno
db.limit: $limit
```

---

### `db.max`

No description available.

**Example:**
```zeno
db.max: 'age'
  as: $oldest
```

---

### `db.min`

No description available.

**Example:**
```zeno
db.min: 'age'
  as: $youngest
```

---

### `db.offset`

No description available.

**Example:**
```zeno
db.offset: $offset
```

---

### `db.or_where`

Add an OR WHERE filter to the query.

**Example:**
```zeno
db.or_where
  col: role
  val: 'admin'
```

---

### `db.order_by`

No description available.

**Example:**
```zeno
db.order_by: 'id DESC'
```

---

### `db.paginate`

Retrieve rows paginated with metadata.

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

No description available.

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
| `params` | `list` | No | List of bind values |
| `val` | `any` | No | Single bind value |

**Main Value Type:** `string`

**Example:**
```zeno
db.select: 'SELECT * FROM users'
  as: $users
```

---

### `db.sum`

No description available.

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

### `db.update`

No description available.

**Example:**
```zeno
db.update
  status: 'active'
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

Add a WHERE BETWEEN filter to the query.

**Example:**
```zeno
db.where_between
  col: age
  val: [18, 30]
```

---

### `db.where_in`

No description available.

**Example:**
```zeno
db.where_in {
  col: id
  val: [1, 2, 3]
}
```

---

### `db.where_not_between`

Add a WHERE NOT BETWEEN filter to the query.

**Example:**
```zeno
db.where_not_between
  col: age
  val: [18, 30]
```

---

### `db.where_not_in`

No description available.

**Example:**
```zeno
db.where_not_in {
  col: status
  val: ['archived', 'deleted']
}
```

---

### `db.where_not_null`

No description available.

**Example:**
```zeno
db.where_not_null: created_at
```

---

### `db.where_null`

No description available.

**Example:**
```zeno
db.where_null: deleted_at
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

## Excel

### `excel.from_template`

Generate Excel from template with marker support

**Example:**
```zeno
excel.from_template: 'template.xlsx'
  data:
    title: "Report"
    users: $user_list
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

### `http.fetch`

No description available.

**Example:**
```zeno
http.fetch: $api_url
  as: $response
```

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

## Image

### `image.info`

No description available.

**Example:**
```zeno
image.info
  path: 'uploads/foto.jpg'
```

---

### `image.resize`

Mengubah ukuran atau format gambar (Placeholder implementasi).

**Example:**
```zeno
image.resize
  source: "input.jpg"
  dest: "output_thumb.jpg"
  width: 100
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

Mengirim email via SMTP.

**Example:**
```zeno
mail.send: $client_email
  host: "smtp.gmail.com"
  port: 587
  user: $smtp_user
  pass: $smtp_pass
  subject: "Invoice"
  body: $html_content
  as: $is_sent
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

## Pdf

### `pdf.download`

Generates a PDF from an HTML string and directly triggers a file download to the client's browser.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `filename` | `string` | **Yes** | The name of the file to be downloaded (e.g. 'report.pdf'). |
| `html` | `string` | **Yes** | The HTML structure to render. |
| `orientation` | `string` | No | Page orientation (Portrait/Landscape). |
| `page_size` | `string` | No | Page size (A4, Letter). |

**Example:**
```zeno
pdf.download:
  html: $invoice_html
  filename: 'Invoice-101.pdf'
```

---

### `pdf.generate`

Convert HTML string to PDF bytes using wkhtmltopdf.

**Inputs:**

| Name | Type | Required | Description |
|------|------|----------|-------------|
| `as` | `any` | **Yes** | Variable to store the resulting PDF byte array. |
| `error` | `any` | No | Variable to store any conversion errors. |
| `html` | `string` | **Yes** | The raw HTML string to convert. |
| `orientation` | `string` | No | Page orientation: 'Portrait' or 'Landscape' (Default: Portrait). |
| `page_size` | `string` | No | Page size dimension: 'A4', 'Letter', etc (Default: A4). |

**Example:**
```zeno
pdf.generate:
  html: $html_body
  orientation: 'Landscape'
  as: $pdf_bytes
  error: $pdf_error
```

---

## Schema

### `schema.create`

Create a new database table using fluent schema builder.

**Example:**
```zeno
schema.create: 'users' {
  column.id: 'id'
  column.string: 'name'
}
```

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

No description available.

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

### `session.flash`

Flash data to the session (cookie) for the next request.

**Example:**
```zeno
session.flash: { key: 'error', val: 'Invalid credentials' }
```

---

### `session.get_flash`

Retrieve flash data and remove it from session.

**Example:**
```zeno
session.get_flash: 'error' { as: $error_msg }
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

