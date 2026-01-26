# Dokumentasi Lengkap ZenoLang & ZenoEngine (Bahasa Indonesia)
**Panduan Resmi: Edisi Lengkap**

---

## 1. Pengantar
ZenoLang adalah bahasa konfigurasi berbasis slot yang dirancang untuk pengembangan backend yang **cepat**, **deklaratif**, dan **mudah dibaca**. ZenoLang berjalan di atas **ZenoEngine**, sebuah runtime Go performa tinggi.

Filosofi utama:
- **Human Friendly:** Kode harus bisa dibaca seperti instruksi bahasa Inggris sederhana.
- **Slot-Based:** Semua perintah adalah "slot" (mirip fungsi) yang menerima argumen (key-value) dan children.
- **Batteries Included:** Dilengkapi fitur built-in untuk Database, Auth, HTTP, File System, hingga Background Jobs.

---

## 2. Sintaks Inti & Kontrol Flow
Bagian ini membahas perintah dasar untuk logika pemrograman.

### 2.1 Variabel & Log (`var`, `log`)
Menyimpan data dan debugging.

```javascript
// Membuat variabel
var: $nama_user {
  val: "Budi Santoso"
}

// Alias legacy (masih didukung)
scope.set: $umur {
  val: 25
}

// Logging ke terminal
log: "User " + $nama_user + " berumur " + $umur
```

### 2.2 Kondisional (`if`)
Melakukan percabangan logika. Mendukung operator: `==`, `!=`, `>`, `<`, `>=`, `<=`.

```javascript
if: $umur >= 17 {
  then: {
    log: "Dewasa"
  }
  else: {
    log: "Belum Dewasa"
  }
}
```

### 2.3 Perulangan (`while`, `loop`, `for`, `forelse`)
ZenoLang mendukung beberapa cara untuk melakukan perulangan.

**a. While / Loop (Kondisional)**
Perulangan yang berjalan selama kondisi benar. `while` dan `loop` adalah alias yang identik.

```javascript
var: $i { val: 1 }

while: "$i <= 5" {
  do: {
    log: "Iterasi ke-" + $i
    math.calc: $i + 1 { as: $i }
    cast.to_int: $i { as: $i } // Pastikan tetap integer
  }
}
```

**b. For (Iterasi List / Foreach)**
Melakukan iterasi pada array atau list hasil database.

```javascript
// $users adalah list dari database
for: $users {
  as: $user         // Variabel untuk item saat ini (default: $item)
  do: {
    log: "Nama: " + $user.name
  }
}
```

**c. For (C-Style)**
ZenoLang juga mendukung format perulangan ala bahasa C.

```javascript
for: "$i=1; $i<=3; $i++" {
  do: {
    log: "Iterasi ke-" + $i
  }
}
```

**d. Forelse (List dengan Fallback)**
Mirip `for`, tetapi memiliki blok `empty:` jika data kosong.

```javascript
forelse: $users {
  as: $user
  do: {
    log: $user.name
  }
  empty: {
    log: "Tidak ada user."
  }
}
```

**e. Break & Continue (Kondisional)**
Gunakan `break` untuk berhenti paksa, dan `continue` untuk lanjut ke iterasi berikutnya. Sekarang mendukung pengecekan kondisi langsung!

```javascript
for: $items {
  do: {
    // Berhenti jika ID sudah 5
    break: "$item.id == 5"
    
    // Lewati jika status 'draft'
    continue: "$item.status == 'draft'"
    
    log: $item.title
  }
}
```

### 2.4 Logika & Kondisional Lanjutan
ZenoLang menyediakan slot core untuk pengecekan data yang lebih elegan tanpa harus menggunakan `if` yang panjang.

**a. Isset & Empty**
Mengecek apakah variabel ada atau kosong.

```javascript
isset: $user {
  do: { log: "User didefinisikan" }
}

empty: $cart {
  do: { log: "Keranjang belanja kosong" }
}
```

**b. Unless**
Kebalikan dari `if`. Menjalankan blok jika kondisi **salah**.

```javascript
unless: $is_admin {
  do: { log: "Akses ditolak!" }
}
```

**c. Switch Case**
Percabangan nilai yang lebih rapi daripada `if-else` bertingkat.

```javascript
switch: $status {
  case: "pending"  { do: { log: "Sedang diproses" } }
  case: "approved" { do: { log: "Disetujui" } }
  default:         { do: { log: "Status tidak dikenal" } }
}
```

### 2.5 Kontrol Akses & Izin (`auth`, `guest`, `can`)
Slot khusus untuk menangani status login dan izin user (RBAC).

```javascript
auth: {
  do: { log: "Hanya untuk user login" }
}

guest: {
  do: { log: "Silakan login terlebih dahulu" }
}

can: "edit-post" {
  resource: $post
  do: { log: "Anda boleh mengedit post ini" }
}
```

**d. Auth Check & User**
Mengecek status login atau mengambil data user dalam script ZenoLang.

```javascript
// Cek boolean
auth.check: { as: $sudah_login }

// Ambil data user
auth.user: { as: $user }
log: "Halo " + $user.username
```

### 2.6 Debugging Tools (`dd`, `dump`)
Alat bantu ala Laravel untuk melihat isi variabel dengan cepat.

- `dump:` Menampilkan isi variabel ke console tanpa berhenti.
- `dd:` (Dump and Die) Menampilkan isi variabel dan menghentikan script saat itu juga.

```javascript
dump: $user
dd: $data_kritis
```

### 2.7 Error Handling (`try-catch`)
Menangani error agar aplikasi tidak crash. ZenoEngine v1.1 telah menyempurnakan pesan error menjadi lebih **spesifik** dan **actionable** (misal: memberi tahu jika atribut salah ketik atau blok `do:` ketinggalan).

```javascript
try: {
  do: {
    // Kode yang mungkin error
    db.execute: "DELETE FROM users"
  }
  catch: {
    // Pesan error kini sangat detail, contoh:
    // "validation error: unknown attribute 'tablee' for slot 'db.select'. Allowed attributes: table, where, ..."
    log: "Terjadi kesalahan: " + $error
  }
}
```


### 2.8 Definisi Fungsi Global (`fn`, `call`) 🆕
ZenoEngine kini mendukung definisi fungsi global yang reusable.

```javascript
// 1. Definisi Fungsi (Global Scope)
fn: hitung_diskon {
  // Logic fungsi
  math.calc: $total * 0.1 { as: $diskon }
  log: "Diskon dihitung: " + $diskon
}

// 2. Panggil Fungsi
var: $total { val: 100000 }
call: hitung_diskon
// Output: Diskon dihitung: 10000
```

### 2.9 Utilitas Lainnya
- **Concurrency Control:** `sleep: 1000` (Jeda milidetik)
- **Timeout:** `ctx.timeout: "5s" { do: { ... } }` (Batasi waktu eksekusi)
- **Include Script:** `include: "src/modules/other.zl"` (Modularisasi kode)

---

## 3. HTTP & Web Server
ZenoEngine memiliki server HTTP built-in yang powerful.

### 3.1 Routing
Routing didefinisikan di level root file.

```javascript
http.get: /hello {
  do: {
    http.ok: { message: "Hello World" }
  }
}

// Route dengan Parameter (Menggunakan syntax Chi Router: {param})
// PENTING: Gunakan tanda kurung kurawal {}, bukan titik dua :
http.get: "/users/{id}" {
  do: {
    // Variabel $id otomatis di-inject ke scope
    http.ok: { message: "User ID: " + $id }
  }
}

http.post: /submit {
  do: { ... }
}

// Grouping Route
http.group: /api/v1 {
  do: {
    include: "routes/v1.zl"
  }
}
```

### 3.2 Mengambil Data Request
- **Query Param:** `http.query: "page" { as: $page }`
- **Form Data:** `http.form: "email" { as: $email }`
- **File Upload:** Lihat bagian *Upload File*.

### 3.3 Response Helpers (Lengkap)
Helper ini otomatis membungkus response dalam JSON standar: `{ "success": boolean, ... }`.

| Slot | HTTP Status | Kegunaan |
| :--- | :--- | :--- |
| `http.ok` | 200 | Sukses umum |
| `http.created` | 201 | Berhasil membuat resource |
| `http.accepted` | 202 | Request diterima (processing) |
| `http.no_content`| 204 | Sukses tanpa body (kosong) |
| `http.bad_request`| 400 | Error input client |
| `http.unauthorized`| 401 | Belum login |
| `http.forbidden` | 403 | Tidak punya hak akses |
| `http.not_found` | 404 | Data tidak ditemukan |
| `http.validation_error`| 422 | Gagal validasi input |
| `http.server_error`| 500 | Error sistem internal |

**Contoh Response Manual:**
```javascript
http.response: 200 {
  type: "text/html" // Default: application/json
  body: "<h1>Custom HTML</h1>"
}
```

### 3.4 Cookies & Redirect
```javascript
// Set Cookie
cookie.set: {
  name: "session_token"
  val: "xyz123"
  age: 3600 // Detik
}

// Redirect
http.redirect: "/login"
```

---

## 4. Database (SQL)
ZenoEngine mendukung **MySQL**, **PostgreSQL**, **SQLite**, dan **SQL Server**.

### 4.1 Query Builder (Direkomendasikan)
Cara aman dan mudah untuk berinteraksi dengan DB tanpa SQL mentah.

**a. Mengambil Data (`db.get`, `db.first`, `db.count`)**
```javascript
db.table: users
db.where: { col: status, val: "active" }
db.where: { col: age, op: ">=", val: 18 } // Support op: >, <, >=, <=, LIKE
db.order_by: "created_at DESC"
db.limit: 10
db.offset: 0

// Eksekusi:
db.get: { as: $users }       // List of Maps
// atau
db.first: { as: $user }      // Single Map
// atau
db.count: { as: $total }     // Integer
```

**b. Manipulasi Data (`db.insert`, `db.update`, `db.delete`)**
```javascript
// Insert
db.table: users
db.insert: {
  name: "Budi"
  email: "budi@test.com"
}
// Hasil auto-increment ID ada di $db_last_id

// Update
db.table: users
db.where: { col: id, val: $id }
db.update: {
  status: "banned"
}

// Delete
db.table: users
db.where: { col: id, val: $id }
db.delete: { as: $affected_rows }
```

### 4.2 Raw SQL
Untuk query kompleks yang tidak bisa ditangani Query Builder.

```javascript
// Select
db.select: "SELECT * FROM users WHERE id = ?" {
  val: $id
  as: $result
  first: true // Jika ingin satu baris saja
}

// Execute (Insert/Update/Delete/DDL)
db.execute: "UPDATE users SET accessed_at = NOW()"
```

### 4.3 Database Transaction (ACID)
Menjamin integritas data dengan membungkus beberapa operasi dalam satu transaksi atomic. Jika satu operasi gagal, **semua perubahan dibatalkan (Rollback)**. Jika sukses semua, baru disimpan permanen (Commit).

```javascript
db.transaction: {
  do: {
    // 1. Kurangi Saldo Pengirim
    db.table: accounts
    db.where: { col: id, val: $sender_id }
    db.update: { balance: $new_sender_balance }

    // 2. Tambah Saldo Penerima
    db.table: accounts
    db.where: { col: id, val: $receiver_id }
    db.update: { balance: $new_receiver_balance }
    
    // 3. Catat Log Transaksi
    db.table: transactions
    db.insert: {
      amount: $amount
      from: $sender_id
      to: $receiver_id
    }
}
// Jika ada error di dalam blok 'do', otomatis Rollback.
```

### 4.4 Multi-Database Support (Database Agnostic) 🆕
ZenoEngine v1.2 sekarang **database agnostic**! Anda bisa dengan mudah beralih antara MySQL, SQLite, PostgreSQL, atau SQL Server tanpa mengubah kode aplikasi.

**a. Konfigurasi Database di `.env`**
```env
# Pilih driver database (mysql, sqlite, postgres, sqlserver)
DB_DRIVER=mysql

# Untuk MySQL
DB_HOST=127.0.0.1:3306
DB_USER=root
DB_PASS=password
DB_NAME=my_database

# Untuk PostgreSQL
# DB_DRIVER=postgres
# DB_HOST=localhost:5432
# DB_USER=postgres
# DB_PASS=password
# DB_NAME=mydb

# Untuk SQLite (cukup path file)
# DB_DRIVER=sqlite
# DB_NAME=./database.db

# Untuk SQL Server
# DB_DRIVER=sqlserver
# DB_HOST=localhost:1433
# DB_USER=sa
# DB_PASS=YourPassword123
# DB_NAME=MyDatabase

# Connection Pool (Opsional)
DB_MAX_OPEN_CONNS=100
DB_MAX_IDLE_CONNS=10
```

**b. Multiple Database Connections**
ZenoEngine mendukung koneksi ke beberapa database sekaligus! Cukup tambahkan prefix `DB_<NAMA>_` di `.env`:

```env
# Database Utama (default)
DB_DRIVER=mysql
DB_HOST=127.0.0.1:3306
DB_NAME=app_main

# Database Warehouse (untuk analytics)
DB_WAREHOUSE_HOST=192.168.1.100:3306
DB_WAREHOUSE_USER=analyst
DB_WAREHOUSE_PASS=secret
DB_WAREHOUSE_NAME=warehouse

# Database Internal (SQLite untuk queue/cache)
# Otomatis dibuat oleh ZenoEngine
```

**c. Menggunakan Database Tertentu**
Semua slot query builder mendukung parameter `db:` untuk memilih koneksi:

```javascript
// Query ke database default
db.table: users
db.get: { as: $users }

// Query ke database warehouse
db.table: analytics
  db: warehouse
db.get: { as: $stats }

// Query ke internal SQLite
db.table: jobs
  db: internal
db.count: { as: $pending_jobs }
```

**d. Kompatibilitas SQL Dialect**
ZenoEngine secara otomatis menangani perbedaan SQL dialect:

| Fitur | MySQL | SQLite | PostgreSQL | SQL Server |
|:------|:------|:-------|:-----------|:-----------|
| Identifier Quote | `` `table` `` | `"table"` | `"table"` | `[table]` |
| Placeholder | `?` | `?` | `$1, $2` | `@p1, @p2` |
| Auto Increment | `AUTO_INCREMENT` | `AUTOINCREMENT` | `SERIAL` | `IDENTITY` |
| LIMIT/OFFSET | `LIMIT x,y` | `LIMIT x OFFSET y` | `LIMIT x OFFSET y` | `OFFSET x ROWS FETCH NEXT y` |

**Anda tidak perlu khawatir tentang perbedaan ini!** Query Builder otomatis menyesuaikan syntax berdasarkan driver yang dipilih.

**e. Best Practices**
- Gunakan **Query Builder** (`db.table`, `db.insert`, dll) untuk portabilitas maksimal
- Hindari raw SQL (`db.execute`) jika ingin mudah beralih database
- Untuk query kompleks yang butuh raw SQL, pertimbangkan membuat versi per-dialect di migration

---

## 5. Validasi Input
Validasi data yang masuk dengan mudah.

```javascript
validate: $form_data {
  rules: {
    username: "required|min:5"
    email: "required|email"
    age: "numeric|min:18|max:100"
  }
  as: $errors
}

if: $errors_any { // Helper flag otomatis
  then: {
    http.validation_error: { errors: $errors }
  }
}
```

---

## 6. Autentikasi (Auth)
Built-in JWT Authentication sistem.

### 6.1 Login
Memverifikasi username/password dengan database hash (Bcrypt).
```javascript
auth.login: {
  username: $input_email
  password: $input_password
  table: "users"        // Default
  col_user: "email"     // Default
  col_pass: "password"  // Default
  secret: "RAHASIA_APP"
  as: $token            // JWT Token string
}
// Jika gagal, otomatis melempar error yang bisa di-catch
```

### 6.2 Middleware (Proteksi Route)
```javascript
auth.middleware: {
  secret: "RAHASIA_APP"
  do: {
    // Kode di sini hanya jalan jika token valid
    auth.user: { as: $current_user } // Ambil data user dari token
    
    http.ok: { message: "Secret Data", user: $current_user }
  }
}
```

---

## 7. Filesystem & Upload

### 7.1 Upload File
Menangani `multipart/form-data` dengan otomatis.
```javascript
http.upload: {
  field: "avatar"         // Nama field form
  dest: "public/uploads"  // Folder tujuan
  as: $filename           // Nama file tersimpan (e.g. 176231_profil.jpg)
}
```

### 7.2 Manipulasi File (IO)
```javascript
// Tulis File
io.file.write: {
  path: "logs/app.log"
  content: "Server started..."
  mode: 0644
}

// Baca File
io.file.read: "config.json" { as: $content }

// Hapus File
io.file.delete: "temp/old.tmp"

// Buat Folder
io.dir.create: "public/assets/new_folder"
```

### 7.3 Pengolahan Gambar
```javascript
// Cek Info
image.info: "uploads/foto.jpg" { as: $info }
// $info.width, $info.height

// Resize (Placeholder implementation)
image.resize: {
  source: "input.jpg"
  dest: "output_thumb.jpg"
  width: 150
}
```

---

## 8. Utilitas Sistem & Matematika

### 8.1 String & Text
```javascript
// Gabung String
strings.concat: {
  val: "Halo "
  val: $firstName
  as: $msg
}

// Slugify (Judul ke URL)
text.slugify: "Berita Hari Ini 2024" { as: $slug }
// Hasil: "berita-hari-ini-2024"

// Sanitize (Anti XSS)
text.sanitize: "<script>alert('hack')</script>Halo" { as: $clean }
// Hasil: "Halo"
```

### 8.2 Matematika
```javascript
// Math Dasar (Float)
math.calc: ceil($harga * 1.1) { as: $harga_pajak }
// Fungsi tersedia: ceil, floor, round, abs, max, min, sqrt, pow

// Keuangan (Decimal Precision - Anti floating point error)
money.calc: ($subtotal - $diskon) + $pajak { as: $total_fix }
```

### 8.3 Lain-lain
```javascript
// Environment Variable
system.env: "DB_HOST" { as: $host }

// Casting Tipe Data
cast.to_int: "123" { as: $number }

// Null Safety
is_null: $var { as: $cek } // true/false
coalesce: $input_user { default: "Anonim"; as: $username }

// Panjang Array
arrays.length: $list_data { as: $jumlah }

### 8.4 Date & Time 📅
Slot khusus untuk manipulasi waktu, sangat berguna untuk fitur "Publish At" atau scheduling.

**a. Ambil Waktu Sekarang (`date.now`)**
Otomatis mendapatkan string terformat DAN objek `time.Time` (suffix `_obj`).

```javascript
date.now: { 
  as: $skarang 
  format: "Human" 
}
// Hasil: $skarang (string), $skarang_obj (object untuk manipulasi)
```

**b. Format Tanggal (`date.format`)**
Mendukung Alias, Layout Go, dan **Format Custom (C#/PHP Style)**.

```javascript
// Menggunakan format custom yyyy-MM-dd
date.format: $user.created_at {
  layout: "dd MMMM yyyy HH:mm"
  as: $tgl_cantik
}

// Alias tersedia: Human, RFC3339, date, time, full
date.format: $obj_waktu { layout: "Human"; as: $h }
```

**c. Manipulasi Waktu (`date.add`, `date.parse`)**
Menambah durasi atau mengubah string menjadi objek.

```javascript
// Tambah 2 hari kedepan
date.add: $skarang_obj { 
  duration: "48h" 
  as: $lusa 
}

// Parse string manual
date.parse: "2025-12-25" { as: $xmas_obj }
```

### 8.5 Keamanan & Kriptografi 🔐
Kumpulan slot untuk menangani hashing dan keamanan.

**a. Hashing Password (`crypto.hash`, `crypto.verify`)**
Menggunakan algoritma Bcrypt yang aman.

```javascript
// Hash Password
crypto.hash: "rahasia123" { as: $hashed }

// Verifikasi Password
crypto.verify: {
  hash: $hashed
  text: "rahasia123"
  as: $is_valid // true/false
}
```

**b. ASP.NET Core Identity (`crypto.verify_aspnet`)** 🆕
Mendukung verifikasi password hash legacy dari ASP.NET Core Identity (V2/V3). Berguna untuk migrasi sistem lama ke ZenoEngine.

```javascript
crypto.verify_aspnet: {
  hash: $db_hash_aspnet
  password: $input_password
  as: $valid
}
```

**c. CSRF Protection (`sec.csrf_token`)**
Mengambil token CSRF untuk form HTML.

```javascript
sec.csrf_token: { as: $token }
```

---

## 9. Email & Background Jobs

### 9.1 Kirim Email
```javascript
mail.send: $user_email {
  subject: "Selamat Bergabung"
  body: "<h1>Halo!</h1>"
  host: "smtp.mailtrap.io"
  port: 587
  user: "smtp_user"
  pass: "smtp_pass"
  as: $is_success
}
```

### 9.2 Background Workers (Redis)
ZenoEngine mendukung antrean tugas berbasis Redis yang kuat.

```javascript
// 1. Konfigurasi Worker (Wajib di main.zl atau boot)
// Menentukan antrean apa saja yang akan diproses oleh instance ini
worker.config: ["high", "default", "low"]

// 2. Masukkan Job ke Antrean
job.enqueue: {
  queue: "default"         // Opsional, default: "default"
  payload: {
    task: "send_email"
    email: "user@test.com"
    user_id: 123
  }
}
```

---

## 10. Templating (Blade)
Untuk merender HTML dinamis, gunakan file `.blade.zl` di folder `views/`. Sintaks mirip Laravel Blade.

**Contoh View:**
`view.blade: "dashboard.blade.zl" { data: $users }`

**Sintaks Blade:**
- `{{ $variabel }}` : Output (escaped)
- `{!! $html !!}` : Output Raw
- `@if(...) @else @endif`
- `@foreach($items as $item) ... @endforeach`
- `@extends('layout')`
- `@section('content') ... @endsection`
- `@include('partials.header')`

---

**Selesai.** Dokumentasi ini mencakup seluruh fitur standar ZenoEngine v1.0.

---

## 11. Logika Lanjutan (Blade Logic & Control Flow)
ZenoEngine menyediakan logika kontrol flow canggih yang bisa digunakan di dalam `.blade.zl` maupun `.zl` biasa.

### 11.1 Loop & Percabangan
```javascript
// Switch Case
logic.switch: $status {
  case: "pending" { do: { log: "Menunggu" } }
  case: "success" { do: { log: "Berhasil" } }
  default: { do: { log: "Unknown" } }
}

// Foreach dengan Else (jika kosong)
logic.forelse: $items {
  as: $item
  do: {
    log: $item.name
  }
  empty: {
    log: "Tidak ada data"
  }
}

// While Loop (Kini tersedia langsung sebagai core slot)
while: "$i < 10" {
  do: {
    log: $i
    math.calc: $i + 1 { as: $i }
  }
}

// C-Style For Loop
logic.for: "$i = 0; $i < 5; $i++" {
  do: { log: $i }
}

// Loop Control
logic.break
logic.continue
```

### 11.2 Pengecekan Nilai (Isset, Empty, Unless)
```javascript
// Cek variabel ada
logic.isset: $user {
  do: { log: "User ada" }
}

// Cek kosong (null, "", 0, [])
logic.empty: $cart {
  do: { log: "Keranjang kosong" }
}

// Kebalikan dari If (Jalankan jika FALSE)
logic.unless: $is_login {
  do: { log: "Silakan Login" }
}
```

### 11.3 Auth Helpers
Helper cepat untuk cek status login di view/logic.

```javascript
// Jika Login
logic.auth: {
  do: { log: "Halo User" }
}

// Jika Tamu (Belum Login)
logic.guest: {
  do: { log: "Halo Tamu" }
}

// Authorization (Gate/Policy)
logic.can: "edit_post" {
  resource: $post
  do: { log: "Bisa Edit" }
}

logic.cannot: "delete_post" {
  resource: $post
  do: { log: "Tidak Bisa Hapus" }
}
```

---

## 12. Fitur Tambahan (JSON, Network, SSE, Cache)

### 12.1 JSON Manipulation
```javascript
// Parse JSON String ke Object
json.parse: $json_string { as: $obj }

// Object ke JSON String
json.stringify: $obj { as: $json_string }

// Output JSON langsung ke HTTP Response
logic.json: $data
```

### 12.2 HTTP Client (Fetch)
Melakukan request ke API luar.
```javascript
http.fetch: "https://api.example.com/data" {
  method: "POST"
  body: $payload // Auto JSON
  as: $response
}
```

### 12.3 Realtime (Server-Sent Events)
Stream data realtime ke browser.
```javascript
sse.stream: {
  // Kirim satu event
  sse.send: {
    event: "welcome"
    data: "Halo!"
  }
  
  // Loop stream (misal data ticker)
  sse.loop: {
    interval: 1000 // ms
    do: {
      sse.send: { data: $harga_saham }
    }
  }
}
```

### 12.4 Cache (Redis Compatible API)
*Catatan: Saat ini adapter Redis belum aktif, slot ini untuk kompatibilitas masa depan.*
```javascript
cache.put: {
  key: "stats"
  val: $data
  ttl: "30m"
}

cache.get: {
  key: "stats"
  default: 0
  as: $result
}
```

---

## 13. Testing Framework
ZenoLang memiliki framework testing bawaan untuk unit testing logika Anda.

```javascript
test: "Cek Penjumlahan" {
  math.calc: 1 + 1 { as: $hasil }
  
  assert.eq: $hasil { expected: 2 }
}

test: "Cek Validasi Email" {
  validate: { email: "bukan-email" } { rules: { email: "email" } as: $err }
  
  assert.neq: $err { expected: nil }
}
```

---

## 14. Blade Components & Stack
Fitur advanced untuk templating `.blade.zl`.

```javascript
// Render Component (views/components/alert.blade.zl)
view.component: "alert" {
  type: "error"
  slot: { "Ada Masalah!" } // Default slot
}

// Push ke Stack
view.push: "scripts" {
  do: {
    <script src="app.js"></script>
  }
}

// Render Stack (di layout)
view.stack: "scripts"
```

---

---

## 15. Contoh Aplikasi Lengkap (Studi Kasus: Manajemen Produk)
Untuk memberikan gambaran utuh bagaimana semua komponen bekerja bersama, berikut adalah kode lengkap untuk modul "Manajemen Produk".


### 15.1 Database Migration (`migrations/001_create_products.zl`)
**💡 Best Practice:** Gunakan logika kondisional untuk mendukung multiple database dialect!

```javascript
// migrations/001_create_products.zl
log: "Migrasi: Membuat tabel products"

// Deteksi driver dari environment
system.env: "DB_DRIVER" { as: $driver }

// Buat tabel sesuai dialect
if: $driver == "sqlite" {
  then: {
    db.execute: "CREATE TABLE IF NOT EXISTS products (
        id INTEGER PRIMARY KEY AUTOINCREMENT,
        name TEXT NOT NULL,
        price REAL NOT NULL,
        stock INTEGER DEFAULT 0,
        created_at DATETIME DEFAULT CURRENT_TIMESTAMP
    )"
  }
  else: {
    // MySQL / PostgreSQL
    db.execute: "CREATE TABLE IF NOT EXISTS products (
        id INT AUTO_INCREMENT PRIMARY KEY,
        name VARCHAR(255) NOT NULL,
        price DECIMAL(10,2) NOT NULL,
        stock INT DEFAULT 0,
        created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
    )"
  }
}

log: "✅ Tabel products berhasil dibuat"
```

> **Catatan:** Jika menggunakan **Query Builder** (`db.table`, `db.insert`), Anda tidak perlu khawatir tentang perbedaan dialect karena ZenoEngine otomatis menanganinya!

### 15.2 Routing & Controller (`main.zl`)
```javascript
// main.zl
http.group: /products {
  
  // 1. TAMPILKAN LIST PRODUK
  http.get: / {
    do: {
      db.table: products
      db.order_by: "id DESC"
      db.get: { as: $products }
      
      view.blade: "products/index.blade.zl" {
        items: $products
        title: "Daftar Produk"
      }
    }
  }

  // 2. FORM TAMBAH (UI)
  http.get: /create {
    do: {
      view.blade: "products/create.blade.zl"
    }
  }

  // 3. PROSES SIMPAN (LOGIC)
  http.post: /store {
    do: {
      // Validasi Input
      validate: $form {
        rules: {
          name: "required|min:3"
          price: "required|numeric"
          stock: "numeric"
        }
        as: $errors
      }

      if: $errors_any {
        then: {
          // Kembali ke form dengan error
          // (Di implementasi nyata, kirim error via session/flash)
          http.validation_error: { errors: $errors }
        }
        else: {
          // Simpan ke DB
          db.table: products
          db.insert: {
            name: $form.name
            price: $form.price
            stock: $form.stock
          }

          // Redirect Sukses
          http.redirect: "/products"
        }
      }
    }
  }
}
```

### 15.3 View Template (`views/products/index.blade.zl`)
```html
<!-- views/products/index.blade.zl -->
@extends('layouts.app')

@section('content')
  <h1>{{ $title }}</h1>
  
  <a href="/products/create" class="btn">Tambah Produk</a>
  
  <table>
    <thead>
      <tr>
        <th>ID</th>
        <th>Nama</th>
        <th>Harga</th>
        <th>Stok</th>
      </tr>
    </thead>
    <tbody>
      @foreach($items as $item)
        <tr>
          <td>{{ $item.id }}</td>
          <td>{{ $item.name }}</td>
          <td>Rp {{ $item.price }}</td>
          <td>
            @if($item.stock > 0)
              <span class="badge-ok">Tersedia ({{ $item.stock }})</span>
            @else
              <span class="badge-err">Habis</span>
            @endif
          </td>
        </tr>
      @endforeach
    </tbody>
  </table>
@endsection
```


Contoh di atas mencakup **Routing**, **Database**, **Validasi**, **Logika Kondisional**, dan **Templating** dalam satu alur kerja yang nyata.

---

## 16. Contoh Sistem Autentikasi Lengkap
Berikut adalah blueprint lengkap untuk membuat sistem Register, Login, dan Protected Dashboard.

### 16.1 Migrasi User (`migrations/002_create_users.zl`)
```javascript
db.execute: "CREATE TABLE IF NOT EXISTS users (
    id INT AUTO_INCREMENT PRIMARY KEY,
    name VARCHAR(100),
    email VARCHAR(100) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)"
```

### 16.2 Auth Routing & Logic (`routes/auth.zl`)

**A. Form & Proses Register**
```javascript
http.get: /register {
  do: { view.blade: "auth/register.blade.zl" }
}

http.post: /register {
  do: {
    // 1. Validasi
    validate: $form {
      rules: {
        name: "required"
        email: "required|email"
        password: "required|min:6"
      }
      as: $errors
    }
    
    if: $errors_any {
      then: { http.validation_error: { errors: $errors } }
      else: {
        // 2. Hash Password
        crypto.hash: $form.password { as: $hash }

        // 3. Simpan User
        db.table: users
        db.insert: {
          name: $form.name
          email: $form.email
          password: $hash
        }
        
        http.redirect: "/login"
      }
    }
  }
}
```

**B. Form & Proses Login**
```javascript
http.get: /login {
  do: { view.blade: "auth/login.blade.zl" }
}

http.post: /login {
  do: {
    // 1. Cek Kredensial & Dapat Token JWT
    try: {
      do: {
        auth.login: {
          email:    $form.email
          password: $form.password
          secret:   "MY_SECRET_KEY" // Sebaiknya dari system.env
          as:       $token
        }
        
        // 2. Simpan Token di Cookie
        cookie.set: {
            name: "token"
            val:  $token
            age:  86400 // 1 Hari
        }
        
        http.redirect: "/dashboard"
      }
      catch: {
        // Login Gagal
        view.blade: "auth/login.blade.zl" {
            error: "Email atau password salah"
        }
      }
    }
  }
}
```

**C. Logout**
```javascript
http.post: /logout {
  do: {
    // Hapus Cookie
    cookie.set: {
      name: "token"
      val: ""
      age: -1
    }
    http.redirect: "/login"
  }
}
```

### 16.3 Protected Routes (`main.zl`)
Mengamankan halaman dashboard agar hanya bisa diakses user yang login.

```javascript
http.group: /dashboard {
    // Middleware Guard
    do: {
        auth.middleware: {
            secret: "MY_SECRET_KEY"
            do: {
                // Di dalam blok ini, user aman & terautentikasi
                
                // Ambil Data Profil
                auth.user: { as: $me }
                
                // Render Dashboard
                http.get: / {
                    do: {
                        view.blade: "dashboard/index.blade.zl" {
                            user: $me
                        }
                    }
                }
            }
        }
    }
}
```

### 16.4 View Helper
Di Blade view, Anda bisa menggunakan `@if` untuk mengecek login status (jika variabel `$user` di-pass atau menggunakan helper logic).

```html
<!-- views/layouts/nav.blade.zl -->
<nav>
    <a href="/">Home</a>
    
    @isset($user)
        <span>Halo, {{ $user.name }}</span>
        <form action="/logout" method="POST">
            <button>Logout</button>
        </form>
    @else
        <a href="/login">Login</a>
        <a href="/register">Register</a>
    @endisset

</nav>
```

---

## 17. Contoh REST API Lengkap & Auto-Docs
ZenoEngine dirancang untuk membangun REST API yang rapi dan terdokumentasi secara otomatis (Swagger/OpenAPI).

### 17.1 Struktur API (`routes/api_v1.zl`)
Pastikan setiap route memiliki metadata (`summary`, `tags`) agar muncul di dokumentasi.

```javascript
http.group: /api/v1 {
  
  // ==========================================
  // GET /api/v1/articles
  // ==========================================
  http.get: /articles {
    summary: "Ambil Daftar Artikel"
    desc: "Mengembalikan daftar artikel dengan pagination."
    tags: ["Artikel"]
    
    query: {
      page: "Nomor halaman (Default: 1)|int"
      limit: "Jumlah per halaman (Default: 10)|int"
    }

    do: {
      // Logic Pagination
      http.query: "page" { as: $page }
      cast.to_int: $page { as: $p }
      
      // Query DB
      db.table: articles
      db.limit: 10
      db.order_by: "id DESC"
      db.get: { as: $data }
      
      // Response JSON Standar
      http.ok: {
        data: $data
        meta: { page: $p, limit: 10 }
      }
    }
  }

  // ==========================================
  // POST /api/v1/articles
  // ==========================================
  http.post: /articles {
    summary: "Buat Artikel Baru"
    tags: ["Artikel"]
    
    // Dokumentasi Body Request
    body: {
      title: "Judul Artikel|required"
      content: "Isi Konten|required"
    }

    do: {
      auth.middleware: {
        secret: "MY_SECRET"
        do: {
           // Proses Simpan
           db.table: articles
           db.insert: {
             title: $form.title
             content: $form.content
             author_id: $session.user_id
           }
           
           http.created: {
             message: "Artikel berhasil dibuat"
             id: $db_last_id
           }
        }
      }
    }
  }

  // ==========================================
  // GET /api/v1/articles/{id}
  // ==========================================
  http.get: /articles/{id} {
    summary: "Detail Artikel"
    tags: ["Artikel"]
    
    // Path Parameter
    params: {
      id: "ID Artikel|int"
    }

    do: {
      db.table: articles
      db.where: { col: id, val: $params.id }
      db.first: { as: $article }
      
      if: $article_found {
        then: { http.ok: { data: $article } }
        else: { http.not_found: { message: "Artikel tidak ditemukan" } }
      }
    }
  }

  // ==========================================
  // GET /api/v1/stats (RAW SQL EXAMPLE)
  // ==========================================
  http.get: /stats {
    summary: "Statistik Artikel"
    tags: ["Laporan"]
    query: {
      min_views: "Minimal Views|int"
    }

    do: {
      // Default Value
      coalesce: $request.min_views { default: 0; as: $min_v }

      // Raw Query dengan Parameter (?) untuk mencegah SQL Injection
      db.select: "
        SELECT author_id, COUNT(*) as total, SUM(views) as views 
        FROM articles 
        WHERE views >= ? 
        GROUP BY author_id 
        ORDER BY views DESC
      " {
        val: $min_v
        as: $stats
      }

      http.ok: { data: $stats }
    }
  }
}
```

### 17.2 Mengakses Dokumentasi API
Setelah server berjalan, dokumentasi otomatis tersedia di:

1.  **Swagger UI (Visual):**
    Buka browser dan akses: `http://localhost:3000/api/docs`
    *(Anda bisa mencoba endpoint langsung dari halaman ini)*

2.  **OpenAPI Spec (JSON):**
    Akses: `http://localhost:3000/api/docs/json`
    *(Bisa di-import ke Postman atau Insomnia)*

---

## 18. Excel Export & Template 🆕
ZenoEngine menyediakan fitur powerful untuk menghasilkan file Excel `.xlsx` berdasarkan template yang sudah ada. Fitur ini sangat berguna untuk membuat laporan (Invoice, PO, Laporan Bulanan) dengan desain kompleks tanpa harus coding formatting dari nol.

### 18.1 Penggunaan Dasar (`excel.from_template`)
Slot ini membaca file template dan mengisi data ke sel tertentu, lalu mengirimkannya sebagai download ke browser.

```javascript
http.get: /download-report {
  do: {
    excel.from_template: "templates/report.xlsx" {
      filename: "Laporan_2024.xlsx"
      
      // Mengisi Data ke Sel Spesifik
      cell: "B2" { val: "Laporan Penjualan" }
      cell: "B3" { val: "Desember 2024" }
      
      // Batch Data (Key-Value)
      data: {
        C5: 5000000
        C6: "Lunas"
      }
    }
  }
}
```

### 18.2 Template Markers (Dynamic Rows)
Anda bisa menaruh placeholder `{{ variable }}` langsung di dalam file Excel template. ZenoEngine akan otomatis menggantinya dengan data.

**Fitur Marker:**
- **Scalar Replacement**: Mengganti `{{ title }}` dengan string.
- **List Expansion (Insert Rows)**: Jika data berupa list (array of objects), ZenoEngine otomatis menduplikasi baris ke bawah sesuai jumlah data.

**Contoh:**
Jika di Excel A5 ada `{{ items.name }}` dan B5 ada `{{ items.price }}`.

```javascript
excel.from_template: "templates/invoice.xlsx" {
  data: {
    // Scalar
    customer: "PT. Maju Jaya"
    date: "2024-12-30"
    
    // List (Otomatis Expand Rows)
    // Gunakan @json jika data ditulis manual di sini
    items: @json('[
      {"name": "Produk A", "price": 10000},
      {"name": "Produk B", "price": 25000}
    ]')
    // Atau gunakan variabel dari database: items: $db_products
  }
}
```

---

## 19. HTTP Client (Advanced) 🆕
Untuk berinteraksi dengan API pihak ketiga (Payment Gateway, Layanan Eksternal), gunakan slot `http.request` yang lebih canggih daripada `http.fetch`.

### 19.1 Fitur Utama
- **Method Bebas**: GET, POST, PUT, DELETE, PATCH, dll.
- **Custom Headers**: Mendukung Bearer Token, API Key, dll.
- **JSON Body Support**: Otomatis konversi Map/List ke JSON.
- **Structured Response**: Mengembalikan objek status, body, headers.

### 19.2 Contoh Penggunaan (POST Request)
```javascript
http.request: "https://api.payment.com/v1/charge" {
  method: "POST"
  timeout: 10 // detik
  
  headers: {
    "Authorization": "Bearer SK-12345"
    "Content-Type": "application/json"
  }
  
  // Body otomatis di-jsonify
  body: {
    amount: 50000
    currency: "IDR"
    order_id: "INV-001"
  }
  
  as: $response
}

// Cek Hasil
if: $response.status == 200 {
  then: {
    log: "Payment Sukses: " + $response.body.transaction_id
  }
  else: {
    log: "Gagal: " + $response.body.error_message
  }
}
```

---
---

## 20. Session & State Management 🆕

**Pertanyaan Umum: Apakah ZenoEngine mendukung Session (`$_SESSION`)?**

ZenoEngine didesain dengan arsitektur **Stateless** untuk performa maksimal dan kemudahan scaling. Artinya, **TIDAK ADA** penyimpanan session di sisi server (seperti file session PHP atau default in-memory session).

Sebagai gantinya, ZenoEngine menggunakan standar modern berbasis **JWT (JSON Web Token)** dan **Cookies**.

### 20.1 Mengakses Data User (`$session`)
Saat Anda menggunakan `auth.middleware`, ZenoEngine memverifikasi token dan menyimpannya ke variabel `$session`. Ini bukan session database, melainkan data *claims* yang ada di dalam Token JWT.

```javascript
http.get: /dashboard {
  middleware: "auth"
  do: {
    // $session otomatis tersedia dari token
    log: "User ID: " + $session.user_id
    log: "Email: " + $session.email
    
    view.blade: "dashboard.blade.zl" {
      user: $session // Pass ke view
    }
  }
}
```

### 20.2 Flash Messages (Notifikasi Redirect)
Bagaimana cara menampilkan pesan "Data Berhasil Disimpan" setelah redirect? Karena tidak ada session flash, gunakan pola **Query Parameter** atau **Cookie**.

**Solusi 1: Query Parameter (Paling Sederhana)**
Cocok untuk pesan sukses sederhana.

*Controller (Simpan & Redirect):*
```javascript
http.redirect: "/posts?msg=Data berhasil disimpan"
```

*View (`posts.blade.zl`):*
```html
@if($query.msg)
  <div class="alert alert-success">
    {{ $query.msg }}
  </div>
@endif
```
*(Catatan: Variabel `$query` otomatis tersedia di View jika dipassing, atau akses via global helper request jika tersedia).*

**Solusi 2: Cookie (Data Sensitif/Kompleks)**
Set cookie sesaat sebelum redirect.

```javascript
cookie.set: {
  name: "flash_msg"
  val: "Login Berhasil"
  age: 5 // Kedaluwarsa dalam 5 detik
}
http.redirect: "/dashboard"
```

---
---

## 21. Arsitektur Polyglot & Masa Depan 🚀

**Tanya: Bisakah ZenoEngine berjalan di atas Runtime lain (.NET, Rust, NodeJS)?**

**Jawab: SANGAT BISA.**

ZenoLang adalah **Spesifikasi Bahasa**, sedangkan **ZenoEngine (Go)** adalah implementasi runtime-nya.

Jika kelak tersedia **ZenoEngine.NET** atau **ZenoEngine.Rust** yang mematuhi standar syntax `http.get`, `db.query`, dll., maka:

1.  **Portabilitas Kode 100%**: File script `.zl` yang Anda tulis hari ini di versi Go, bisa di-copy paste mentah-mentah ke server yang menjalankan runtime Rust atau .NET tanpa perubahan satu baris pun.
2.  **Microservice Heterogen**: Dalam satu cluster Kubernetes, Anda bisa memiliki:
    -   `Service A` (User Auth) berjalan di **ZenoEngine (Go)** untuk konkurensi tinggi.
    -   `Service B` (Image Processing) berjalan di **ZenoEngine (Rust)** untuk performa CPU mentah.
    -   `Service C` (Enterprise Logic) berjalan di **ZenoEngine (.NET Core / Modern .NET)** untuk performa tinggi & cross-platform di Linux/Docker.
    
Semua service tersebut berkomunikasi via HTTP/REST standar, dan developer hanya perlu menguasai **satu bahasa** (ZenoLang) untuk menulis logika di semua service tersebut.

Ini mencapai visi **"Write Once, Run Anywhere, On Any Optimized Runtime"**.

---
---

## 22. Arsitektur Microservice 🏗️

ZenoEngine sangat ideal untuk membangun sistem terdistribusi (Microservices) karena sifatnya yang **Stateless**, **Ringan (High Performance)**, dan **Modular**.

### 22.1 Pola Komunikasi Antar Service
Dalam arsitektur microservice, setiap layanan berjalan terpisah (misal di container Docker berbeda) dan berkomunikasi melalui HTTP REST API.

**Skenario:**
- `Service Auth` (Port 3001): Menangani user & token.
- `Service Product` (Port 3002): Menangani data produk.
- `Service Order` (Port 3003): Menangani pesanan.

**Contoh: Service Order memverifikasi User ke Service Auth**
Saat User membuat pesanan, Service Order perlu tahu siapa user ini. Alih-alih akses DB user langsung (anti-pattern), ia memanggil API Auth.

*Code di Service Order (`order.zl`):*
```javascript
http.post: /orders {
  do: {
    // 1. Ambil Token dari Header
    http.header: "Authorization" { as: $token }
    
    // 2. Validasi ke Service Auth (Inter-service communication)
    http.request: "http://service-auth:3001/validate-token" {
      method: "POST"
      body: { token: $token }
      as: $auth_res
    }
    
    if: $auth_res.status != 200 {
      then: { http.unauthorized: { message: "Token tidak valid" } }
    }
    
    // 3. Proses Pesanan (User Valid)
    $user_id : $auth_res.body.user_id
    // ... logic insert order ...
  }
}
```

### 22.2 Konfigurasi Database Terisolasi
Setiap microservice sebaiknya memiliki database sendiri. Gunakan `.env` yang berbeda untuk setiap service.

- **Service Auth**: `DB_NAME=auth_db`
- **Service Product**: `DB_NAME=product_db`

### 22.3 API Gateway
Anda bisa menggunakan satu instance ZenoEngine sebagai **API Gateway** yang meneruskan request ke service di belakangnya menggunakan `http.request` atau `http.redirect` (tergantung pola).

```javascript
// Gateway Code
http.group: /api/products {
  do: {
     // Proxy ke Service Product
     http.request: "http://service-product:3002" + $request.path {
       method: $request.method
       body: $request.body
       as: $res
     }
     http.response: $res.status { body: $res.body }
  }
}
```

---
---

## 23. Multi-Tenancy (Database Per Tenant) 🏢

ZenoEngine mendukung arsitektur SaaS Multi-Tenant di mana setiap tenant (pelanggan) memiliki database terpisah, namun dilayani oleh satu aplikasi (codebase) yang sama.

### 23.1 Konfigurasi Database Tenant (`.env`)
Definisikan koneksi database untuk setiap tenant di file lingkungan.

```env
# Database System (untuk validasi tenant & metadata)
DB_DRIVER=mysql
DB_HOST=127.0.0.1
DB_PORT=3306
DB_NAME=saas_system
DB_USER=root
DB_PASS=secret

# Tenant ABC (MySQL)
DB_TENANT_ABC_DRIVER=mysql
DB_TENANT_ABC_HOST=127.0.0.1
DB_TENANT_ABC_PORT=3306
DB_TENANT_ABC_NAME=tenant_abc_db
DB_TENANT_ABC_USER=root
DB_TENANT_ABC_PASS=secret

# Tenant XYZ (PostgreSQL)
DB_TENANT_XYZ_DRIVER=postgres
DB_TENANT_XYZ_HOST=127.0.0.1
DB_TENANT_XYZ_PORT=5432
DB_TENANT_XYZ_NAME=tenant_xyz_db
DB_TENANT_XYZ_USER=postgres
DB_TENANT_XYZ_PASS=secret

# Tenant 123 (SQLite - untuk development/testing)
DB_TENANT_123_DRIVER=sqlite
DB_TENANT_123_NAME=./data/tenant_123.db
```

### 23.2 Database System - Tabel Tenants
Buat tabel di database sistem untuk menyimpan metadata tenant.

**Migration: `migrations/001_create_tenants.zl`**
```javascript
db.execute: "CREATE TABLE IF NOT EXISTS tenants (
    id INT AUTO_INCREMENT PRIMARY KEY,
    tenant_code VARCHAR(50) UNIQUE NOT NULL,
    tenant_name VARCHAR(100) NOT NULL,
    db_connection VARCHAR(50) NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
)"

// Seed Data Tenant
db.table: tenants
db.insert: {
  tenant_code: "abc"
  tenant_name: "PT. ABC Corporation"
  db_connection: "tenant_abc"
  is_active: true
}

db.table: tenants
db.insert: {
  tenant_code: "xyz"
  tenant_name: "XYZ Industries"
  db_connection: "tenant_xyz"
  is_active: true
}
```

### 23.3 Helper untuk Identifikasi Tenant

Buat helper module yang dapat digunakan kembali untuk mengidentifikasi tenant dari berbagai sumber.

**File: `src/helpers/tenant.zl`**
```javascript
// ============================================
// TENANT IDENTIFICATION HELPER
// ============================================
// Helper ini mengidentifikasi tenant dari:
// 1. HTTP Header "X-Tenant-ID" (prioritas tertinggi)
// 2. Subdomain (misal: abc.myapp.com → "abc")
// 3. Query Parameter ?tenant=abc (untuk testing)
// 4. Default tenant jika tidak ditemukan
// ============================================

// STRATEGI 1: Dari HTTP Header
http.header: "X-Tenant-ID" { as: $tenant_from_header }

// STRATEGI 2: Dari Subdomain
http.header: "Host" { as: $host }

// Ekstrak subdomain (bagian pertama sebelum titik)
// Contoh: "abc.myapp.com" → ["abc", "myapp", "com"]
string.split: $host { delimiter: "."; as: $host_parts }

// Ambil subdomain (elemen pertama)
// Jika hanya 2 bagian (myapp.com), maka tidak ada subdomain
if: $host_parts.length > 2 {
  then: {
    $tenant_from_subdomain : $host_parts.0
  }
}

// STRATEGI 3: Dari Query Parameter (untuk testing/debugging)
http.query: "tenant" { as: $tenant_from_query }

// PRIORITAS: Header > Subdomain > Query > Default
coalesce: $tenant_from_header { 
  default: $tenant_from_subdomain; 
  as: $tenant_id_temp1 
}

coalesce: $tenant_id_temp1 { 
  default: $tenant_from_query; 
  as: $tenant_id_temp2 
}

coalesce: $tenant_id_temp2 { 
  default: "default"; 
  as: $tenant_code 
}

// ============================================
// VALIDASI TENANT
// ============================================
// Cek apakah tenant valid di database sistem

db.table: tenants {
  db: "system"  // Gunakan koneksi database sistem
}
db.where: { col: "tenant_code", val: $tenant_code }
db.where: { col: "is_active", val: true }
db.first: { as: $tenant_info }

if: $tenant_info_found {
  then: {
    // Tenant valid - set variabel global
    $tenant_id : $tenant_info.tenant_code
    $tenant_name : $tenant_info.tenant_name
    $tenant_db : $tenant_info.db_connection
    
    log: "✓ Tenant identified: " + $tenant_id + " (" + $tenant_name + ")"
  }
  else: {
    // Tenant tidak ditemukan atau tidak aktif
    log: "✗ Invalid tenant: " + $tenant_code
    
    http.error: {
      status: 400
      message: "Tenant tidak valid atau tidak aktif"
      tenant_code: $tenant_code
    }
  }
}
```

### 23.4 Implementasi di Routes

**A. Menggunakan Helper di Route Group**

**File: `src/main.zl`**
```javascript
// ============================================
// MULTI-TENANT APPLICATION ROUTES
// ============================================

http.group: /app {
  do: {
    // Include tenant helper - ini akan set $tenant_id dan $tenant_db
    sys.include: "helpers/tenant.zl"
    
    // Setelah include, variabel berikut tersedia:
    // - $tenant_id: kode tenant (misal: "abc")
    // - $tenant_name: nama tenant (misal: "PT. ABC Corporation")
    // - $tenant_db: nama koneksi database (misal: "tenant_abc")
    
    // ========================================
    // DASHBOARD
    // ========================================
    http.get: /dashboard {
      do: {
        // Query ke database tenant yang teridentifikasi
        db.table: users {
          db: $tenant_db  // Dynamic database connection
        }
        db.count: { as: $total_users }
        
        db.table: products {
          db: $tenant_db
        }
        db.count: { as: $total_products }
        
        view.blade: "dashboard.blade.zl" {
          tenant_name: $tenant_name
          total_users: $total_users
          total_products: $total_products
        }
      }
    }
    
    // ========================================
    // USERS MANAGEMENT
    // ========================================
    http.get: /users {
      do: {
        db.table: users {
          db: $tenant_db
        }
        db.order_by: "created_at DESC"
        db.get: { as: $users }
        
        view.blade: "users/index.blade.zl" {
          users: $users
          tenant: $tenant_name
        }
      }
    }
    
    http.post: /users {
      do: {
        // Validasi
        validate: $form {
          rules: {
            name: "required|min:3"
            email: "required|email"
          }
          as: $errors
        }
        
        if: $errors_any {
          then: {
            http.validation_error: { errors: $errors }
          }
          else: {
            // Insert ke database tenant
            db.table: users {
              db: $tenant_db
            }
            db.insert: {
              name: $form.name
              email: $form.email
              tenant_id: $tenant_id  // Simpan referensi tenant
            }
            
            http.redirect: "/app/users?msg=User berhasil ditambahkan"
          }
        }
      }
    }
    
    // ========================================
    // PRODUCTS (API Example)
    // ========================================
    http.get: /api/products {
      do: {
        db.table: products {
          db: $tenant_db
        }
        db.get: { as: $products }
        
        http.ok: {
          tenant: $tenant_id
          data: $products
        }
      }
    }
  }
}
```

**B. Tenant-Specific API Endpoints**

**File: `src/routes/api.zl`**
```javascript
// API dengan Multi-Tenant Support
http.group: /api/v1 {
  do: {
    // Identifikasi tenant untuk semua endpoint API
    sys.include: "helpers/tenant.zl"
    
    // GET /api/v1/stats - Statistik per tenant
    http.get: /stats {
      summary: "Statistik Tenant"
      tags: ["Analytics"]
      
      do: {
        // Query 1: Total Users
        db.table: users {
          db: $tenant_db
        }
        db.count: { as: $total_users }
        
        // Query 2: Total Orders
        db.table: orders {
          db: $tenant_db
        }
        db.count: { as: $total_orders }
        
        // Query 3: Revenue
        db.table: orders {
          db: $tenant_db
        }
        db.where: { col: "status", val: "paid" }
        db.select: "SUM(total_amount) as revenue" {
          as: $revenue_data
        }
        
        http.ok: {
          tenant: {
            id: $tenant_id
            name: $tenant_name
          }
          stats: {
            total_users: $total_users
            total_orders: $total_orders
            revenue: $revenue_data.0.revenue
          }
        }
      }
    }
  }
}
```

### 23.5 Alternatif: Identifikasi dari Subdomain Saja

Jika aplikasi Anda hanya menggunakan subdomain (tanpa header), buat helper yang lebih sederhana:

**File: `src/helpers/tenant_subdomain.zl`**
```javascript
// Identifikasi tenant HANYA dari subdomain
http.header: "Host" { as: $host }

// Ekstrak subdomain
string.split: $host { delimiter: "."; as: $parts }

// Validasi: harus ada minimal 3 bagian (subdomain.domain.tld)
if: $parts.length < 3 {
  then: {
    http.error: {
      status: 400
      message: "Akses harus melalui subdomain tenant (misal: abc.myapp.com)"
    }
  }
}

$tenant_code : $parts.0

// Validasi di database
db.table: tenants
db.where: { col: "tenant_code", val: $tenant_code }
db.where: { col: "is_active", val: true }
db.first: { as: $tenant_info }

if: $tenant_info_found {
  then: {
    $tenant_id : $tenant_info.tenant_code
    $tenant_db : $tenant_info.db_connection
  }
  else: {
    http.error: {
      status: 404
      message: "Tenant tidak ditemukan: " + $tenant_code
    }
  }
}
```

### 23.6 Testing Multi-Tenant Setup

**A. Testing dengan cURL (Header-based)**
```bash
# Tenant ABC
curl -H "X-Tenant-ID: abc" http://localhost:3000/app/dashboard

# Tenant XYZ
curl -H "X-Tenant-ID: xyz" http://localhost:3000/app/api/products
```

**B. Testing dengan Subdomain (Local Development)**

Edit `/etc/hosts` (Linux/Mac) atau `C:\Windows\System32\drivers\etc\hosts` (Windows):
```
127.0.0.1  abc.myapp.local
127.0.0.1  xyz.myapp.local
```

Akses di browser:
- `http://abc.myapp.local:3000/app/dashboard`
- `http://xyz.myapp.local:3000/app/users`

**C. Testing dengan Query Parameter**
```
http://localhost:3000/app/dashboard?tenant=abc
http://localhost:3000/app/users?tenant=xyz
```

### 23.7 Best Practices

1. **Isolasi Data**: Pastikan setiap query SELALU menggunakan `db: $tenant_db` untuk mencegah data leak antar tenant.

2. **Validasi Tenant**: Selalu validasi tenant di awal request untuk mencegah akses tidak sah.

3. **Logging**: Log setiap akses tenant untuk audit trail:
   ```javascript
   log: "Tenant: " + $tenant_id + " | User: " + $session.user_id + " | Action: view_users"
   ```

4. **Error Handling**: Berikan error yang jelas jika tenant tidak valid:
   ```javascript
   if: !$tenant_info_found {
     then: {
       http.error: {
         status: 400
         message: "Tenant tidak valid"
         code: "INVALID_TENANT"
       }
     }
   }
   ```

5. **Caching Tenant Info**: Untuk performa, cache informasi tenant di memory/Redis (jika tersedia).

6. **Migration per Tenant**: Jalankan migrasi untuk setiap tenant database:
   ```bash
   # Script untuk run migration di semua tenant
   zeno migrate --tenant=abc
   zeno migrate --tenant=xyz
   ```

### 23.8 Keuntungan Arsitektur Multi-Tenant ZenoEngine

✅ **Isolasi Data Penuh**: Setiap tenant memiliki database terpisah (data security maksimal)  
✅ **Skalabilitas**: Tenant besar bisa dipindah ke server database terpisah  
✅ **Fleksibilitas Database**: Tenant A bisa MySQL, Tenant B bisa PostgreSQL  
✅ **Satu Codebase**: Semua tenant menggunakan kode yang sama  
✅ **Dynamic Switching**: Tidak perlu restart server saat menambah tenant baru  
✅ **Backup Granular**: Backup per tenant, bukan satu database besar

---

## 24. Advanced Query Builder (Fitur Baru)

ZenoLang kini mendukung fitur Query Builder tingkat lanjut yang setara dengan framework modern seperti Laravel, memungkinkan Anda melakukan query kompleks tanpa Raw SQL.

### 24.1. Memilih Kolom Spesifik (`db.columns`)
Secara default `db.get` mengambil semua kolom (`SELECT *`). Gunakan `db.columns` untuk memilih kolom tertentu.

```javascript
/* Memilih kolom spesifik */
db.table: users
db.columns: ["id", "name", "email"]
db.get: { as: $users }
```

### 24.2. Joins (`db.join`, `db.left_join`)
Menggabungkan data dari beberapa tabel.

```javascript
/* INNER JOIN */
db.table: users
db.columns: ["users.name", "posts.title"]
db.join: {
    table: posts
    on: ["users.id", "=", "posts.user_id"]
}
db.get: { as: $results }

/* LEFT JOIN */
db.table: users
db.left_join: {
    table: orders
    on: ["users.id", "=", "orders.user_id"]
}
db.get: { as: $user_orders }
```

### 24.3. Filter Lanjutan (`db.where_in`, `db.where_not_in`, `db.where_null`)
Melakukan filter data dengan operator himpunan atau null check.

```javascript
/* WHERE IN */
db.table: products
db.where_in: {
    col: "category_id"
    val: [1, 5, 12]
}
db.get: { as: $products }

/* WHERE NULL */
db.table: users
db.where_null: "deleted_at"
db.get: { as: $active_users }
```

### 24.4. Aggregates & Grouping (`db.group_by`, `db.having`)
Melakukan pengelompokan data dan filter pada hasil agregasi.

```javascript
/* Menghitung total transaksi per user */
db.table: transactions
db.columns: ["user_id", "COUNT(*) as total_trx", "SUM(amount) as total_amount"]
db.group_by: "user_id"
db.having: {
    col: "total_amount"
    op: ">"
    val: 1000000
}
db.get: { as: $whales }
```

### 24.5. Keunggulan
- **Database Agnostic**: Query ini akan di-compile menjadi SQL yang sesuai dengan database yang digunakan (MySQL, PostgreSQL, SQL Server, SQLite).
- **Safe**: Parameter binding otomatis untuk mencegah SQL Injection.
- **Support Dot Notation**: Otomatis menangani quoting untuk kolom seperti `users.name` menjadi `` `users`.`name` ``.

---
**Selesai.** Anda kini memiliki panduan lengkap: Dasar, CRUD Web, Auth System, REST API, Excel Export, HTTP Client, Manajemen Session, Visi Polyglot, Arsitektur Microservice, Multi-Tenancy, dan **Advanced Query Builder**.

---

## 12. Export Excel (Excelize) 🆕
Fitur native untuk menghasilkan file Excel (.xlsx) menggunakan template. Sangat berguna untuk laporan.

### 12.1 Konsep Template
ZenoEngine menggunakan pendekatan **Template-Based**. Anda membuat file Excel biasa sebagai template, menandai sel dengan `{{ variable }}`, dan ZenoEngine akan mengisinya.

**Fitur:**
- **Scalar Replacement**: Mengganti `{{ nama }}` dengan nilai variabel.
- **List Expansion**: Jika variabel adalah array, baris excel akan otomatis diduplikasi (insert rows) sesuai jumlah data.
- **Image Insertion**: Menyisipkan gambar ke dalam sel.
- **Formula Injection**: Menulis rumus Excel dinamis.

### 12.2 Penggunaan (`excel.from_template`)

```javascript
excel.from_template: "templates/report_invoice.xlsx" {
  filename: "Invoice_2024.xlsx"
  sheet: "Sheet1"
  
  // 1. Data Mapping (Scalar & List)
  data: {
    // Scalar (Header Info)
    "invoice_no": "INV/001/2024"
    "customer": $user.name
    "date": $tanggal_hari_ini
    
    // List (Otomatis expand baris)
    // Pastikan di template ada row dengan {{ items.name }}, {{ items.qty }}, dst.
    "items": $list_belanja
  }
  
  // 2. Insert Gambar (Optional)
  images: {
    "A1": "public/logo.png"  // Koordinat sel : Path gambar
  }

  // 3. Formula Injection (Optional) 🆕
  formulas: {
    "E21": "SUM(E5:E20)"
    "F21": "AVERAGE(F5:F20)"
  }
  
  // 4. Override Cell Manual (Optional)
  cell: "E20" { val: "Total Bayar" }
}
```

---

## 13. Modern Frontend (Inertia.js) 🆕
ZenoEngine memiliki dukungan first-party untuk **Inertia.js**. Ini memungkinkan Anda membangun *Single Page Application* (React, Vue, Svelte) tanpa membuat API terpisah (Monolith Modern).

### 13.1 Render Halaman (`inertia.render`)
Mengirim response Inertia ke frontend.

```javascript
http.get: /dashboard {
  do: {
    // Ambil data dari DB
    db.table: stats
    db.first: { as: $stats }
    
    // Render Component React/Vue
    inertia.render: {
      component: "Dashboard/Index"  // Nama file komponen di frontend
      props: {
        stats: $stats
        auth_user: $auth_user
      }
    }
  }
}
```

### 13.2 Shared Data (`inertia.share`)
Membagikan data ke SELURUH halaman Inertia (biasanya untuk User Auth, Flash Message, Setting Global). Panggil ini di Middleware.

```javascript
// Di middleware global
inertia.share: {
  auth: {
    user: $current_user
  }
  app_name: "Zeno App v1.0"
  flash: $flash_messages
}
```

### 13.3 Manual Location (`inertia.location`)
Memaksa full page reload (hard redirect) di sisi client Inertia. Berguna saat login/logout atau pindah domain.

```javascript
inertia.location: {
  url: "/login"
}
```

---

## 14. Lampiran: Referensi API Lengkap (Cheat Sheet) 📚

### 14.1 Database (SQL)
**Driver Agnostic**: Mendukung MySQL, PostgreSQL, SQLite, SQL Server.

| Slot | Signature / Children | Deskripsi |
| :--- | :--- | :--- |
| `db.table` | `value: "table"`, `db: "conn"` | Set tabel aktif/koneksi. |
| `db.columns`| `value: ["col1", "col2"]` | Pilih kolom spesifik. |
| `db.get` | `as: $var` | Ambil banyak baris (List of Maps). |
| `db.first` | `as: $var` | Ambil satu baris (Map). |
| `db.count` | `as: $var` | Hitung jumlah baris (Int). |
| `db.insert` | Child keys sebagai kolom. | Insert data. Return `$db_last_id`. |
| `db.update` | Child keys sebagai kolom. | Update data. Butuh `db.where`. |
| `db.delete` | `as: $count` (optional) | Hapus data. Butuh `db.where`. |
| `db.where` | `col: "name"`, `val: $val`, `op: "="` | Filter WHERE. Op default `=`. |
| `db.where_in` | `col: "id"`, `val: [1, 2]` | WHERE IN. |
| `db.where_not_in`| `col: "id"`, `val: [1, 2]` | WHERE NOT IN. |
| `db.where_null` | `value: "col_name"` | WHERE col IS NULL. |
| `db.where_not_null`| `value: "col_name"` | WHERE col IS NOT NULL. |
| `db.group_by` | `value: "col_name"` | GROUP BY. |
| `db.having` | `col: "c"`, `op: ">"`, `val: 1` | HAVING. |
| `db.order_by` | `value: "id DESC"` | ORDER BY. |
| `db.limit` | `value: $int` | LIMIT. |
| `db.offset` | `value: $int` | OFFSET. |
| `db.join` | `table: "t2"`, `on: ["t1.id", "=", "t2.fk"]` | INNER JOIN. |
| `db.left_join` | `table: "t2"`, `on: [...]` | LEFT JOIN. |
| `db.transaction`| `do: { ... }` | Transaksi Atomik. Auto-rollback jika error. |
| `db.select` | `value: "SQL"`, `val: $p1`, `as: $res` | Raw SQL Select (Aman). |
| `db.execute` | `value: "SQL"` | Raw SQL Execute (Aman). |

### 14.2 HTTP Server & Client
**PENTING**: Gunakan syntax `{id}` untuk parameter, JANGAN gunakan `:id`.

| Slot | Signature / Children | Deskripsi |
| :--- | :--- | :--- |
| `http.get`, `post`...| `value: "/path/{id}"`, `do: {}`| Definisi Route. Variabel `{id}` otomatis di-inject. |
| `http.group` | `value: "/api"`, `do: {}` | Grouping route. Bisa inherit middleware. |
| `http.query` | `value: "param"`, `as: $var` | Ambil Query Param (`?id=1`). |
| `http.form` | `value: "field"`, `as: $var` | Ambil Form/Multipart data. |
| `http.response` | `status: 200`, `data: $json` | Kirim JSON response custom. |
| `http.ok` | `value: { ... }` | Kirim 200 OK. |
| `http.created` | `value: { ... }` | Kirim 201 Created. |
| `http.not_found` | `value: { ... }` | Kirim 404 Not Found. |
| `http.redirect` | `value: "/url"` | Redirect user. |
| `http.upload` | `field: "file"`, `dest: "path"`, `as: $name` | Handle Upload File. |

### 14.3 Logika & Flow Control

| Slot | Signature / Children | Deskripsi |
| :--- | :--- | :--- |
| `if` | `value: "$a > $b"`, `then: {}`, `else: {}` | Kondisional. |
| `switch` | `value: $var`, `case: "val" {}`, `default: {}` | Switch Case. |
| `while` / `loop` | `value: "cond"`, `do: {}` | While Loop. |
| `for` / `foreach` | `value: $list`, `as: $item`, `do: {}` | Foreach Loop. |
| `break` | `value: "$i > 5"` (opsional) | Hentikan loop. |
| `continue` | `value: "$i % 2 == 0"` (opsional) | Lanjut iterasi berikutnya. |
| `return` | - | Hentikan eksekusi handler saat ini. |
| `try` | `do: {}`, `catch: {}` | Error handling. Pesan error di `$error`. |
| `ctx.timeout` | `value: "5s"`, `do: {}` | Batasi waktu eksekusi blok kode. |
| `fn` | `value: name`, `children...` | Definisi fungsi global. |
| `call` | `value: name` | Panggil fungsi global. |
| `logic.compare` | `v1: $a`, `op: "=="`, `v2: $b`, `as: $res`| Komparasi eksplisit. |
| `isset` | `value: $var`, `do: {}` | Eksekusi jika variabel ada. |
| `empty` | `value: $var`, `do: {}` | Eksekusi jika variabel kosong/null. |
| `unless` | `value: $bool`, `do: {}` | Eksekusi jika kondisi FALSE. |

### 14.4 Utils, Security & Filesystem

| Slot | Signature | Deskripsi |
| :--- | :--- | :--- |
| `log` | `value: "msg"` | Print ke console. |
| `var` | `val: $val` | Set variabel. |
| `sleep` | `value: ms` | Sleep N milidetik. |
| `coalesce` | `val: $a`, `default: "b"`, `as: $r` | Null coalescing. |
| `cast.to_int` | `val: $v`, `as: $i` | Ubah ke Integer. |
| `crypto.hash` | `val: $pass`, `as: $hash` | Hash Password (Bcrypt). |
| `crypto.verify` | `hash: $h`, `text: $p`, `as: $ok` | Verifikasi Password. |
| `sec.csrf_token`| `as: $token` | Ambil token CSRF. |
| `validator.validate`| `input: $map`, `rules: {...}`, `as: $err` | Validasi input. |
| `math.calc` | `expr: "ceil($a * 1.1)"`, `as: $res` | Matematika Float. |
| `money.calc` | `expr: "$a - $b"`, `as: $res` | Matematika Decimal (Keuangan). |
| `io.file.write` | `path: "f.txt"`, `content: "s"` | Tulis file. |
| `io.file.read` | `path: "f.txt"`, `as: $content` | Baca file. |
| `io.file.delete`| `path: "f.txt"` | Hapus file. |
| `io.dir.create` | `path: "dir"` | Buat direktori. |
| `image.info` | `path: "img.jpg"`, `as: $info` | Cek dimensi gambar. |
| `image.resize` | `source: "src"`, `dest: "dst"`, `width: 100` | Resize/Convert gambar. |

### 14.5 Auth & Jobs

| Slot | Signature | Deskripsi |
| :--- | :--- | :--- |
| `auth.login`| `username: $u`, `password: $p`, `as: $token` | Login & Issue JWT. |
| `auth.middleware`| `do: {}` | Proteksi route. Inject `$auth`. |
| `auth.user` | `as: $user` | Ambil user login saat ini. |
| `auth.check`| `as: $bool` | Cek status login (true/false). |
| `jwt.sign` | `claims: {...}`, `secret: "s"`, `as: $t` | Sign JWT manual. |
| `jwt.verify`| `token: $t`, `secret: "s"`, `as: $c` | Verify JWT manual. |
| `worker.config` | `value: ["queue1", "queue2"]` | Konfigurasi worker queue. |
| `job.enqueue` | `queue: "q"`, `payload: {...}` | Masukkan job ke antrean. |
| `mail.send` | `to: $e`, `subject: "s"`, `body: "b"`, `host: "smtp"` | Kirim email SMTP. |
