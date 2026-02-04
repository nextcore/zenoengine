# 🐘 ZenoEngine PHP-Native Bridge (Rust Edition)

Selamat datang di dokumentasi resmi **ZenoEngine PHP-Native Bridge**. Plugin ini memungkinkan Anda menjalankan script PHP dan framework **Laravel** dengan performa **100% Native** langsung dari ZenoLang.

## 🚀 Overview
Plugin ini menggunakan arsitektur **Sidecar**. ZenoEngine (Go) bertindak sebagai orkestrator yang menjalankan bridge (Rust) sebagai proses terpisah. Komunikasi dilakukan melalui protokol **JSON-RPC** via *Standard Input/Output (StdIn/StdOut)*.

### Mengapa menggunakan Rust?
- **Keamanan Memori & Performa**: Rust memberikan jaminan keamanan memori tanpa garbage collector.
- **Embedded PHP**: Rust melakukan *static linking* terhadap `libphp`, sehingga interpreter PHP tertanam langsung di dalam satu file binary.
- **Portable (Zero Installation)**: Pengguna tidak perlu menginstal PHP di sistem target.

---

## ✅ Kompatibilitas

| Fitur | Status | Keterangan |
| :--- | :--- | :--- |
| **Sistem Operasi** | Windows, Linux, macOS | Mendukung arsitektur x86_64 dan ARM64. |
| **Versi PHP** | 8.1 - 8.3+ | Ditentukan oleh library `php-dev` atau PHP SDK yang digunakan saat kompilasi (`build.rs`). |
| **Framework** | Laravel 10/11+ | Mendukung penuh Artisan, ORM Eloquent, dan Service Container. |

---

## 🛠️ Status & Kapabilitas (Implementasi Saat Ini)

Plugin ini telah mendukung fitur-fitur enterprise berikut:

1. **Request Lifecycle Isolation**: ✅ **AKTIF**.
   Bridge menggunakan `php_request_startup` dan `php_request_shutdown` di setiap loop request. Ini menjamin **State Framework (Laravel Container) bersih** setiap kali dipanggil, mencegah memory leak atau data leak antar user.

2. **Output Capture (Hybrid Strategy)**: ✅ **AKTIF**.
   Output PHP (`echo`, `print`) tidak langsung dibuang ke StdOut (karena akan merusak protokol JSON-RPC).
   *   **Mekanisme**: Menggunakan `ob_start()` di PHP dan menyimpan hasilnya ke **Temporary File** (`/tmp` atau `%TEMP%`).
   *   **Keuntungan**: Stabil dan mampu menangani output binary besar.
   *   **Catatan**: Sedikit overhead I/O disk dibanding in-memory buffer, namun lebih aman untuk implementasi awal.

3. **Managed DB Pooling (Protocol Ready)**: ⚠️ **PARTIAL**.
   Bridge sudah memiliki slot `php.db_proxy` yang siap meneruskan request query dari PHP ke ZenoEngine (Go). Namun, Anda perlu mengimplementasikan driver database di sisi aplikasi PHP/Laravel Anda untuk memanfaatkan fitur ini (lihat panduan *DB Proxy*).

4. **Inject Superglobals**: ✅ **AKTIF**.
   Variabel global seperti `$_SERVER['REQUEST_URI']`, `$_SERVER['REQUEST_METHOD']`, dan `$_SERVER['ZENO_SCOPE']` otomatis diisi berdasarkan payload dari Zeno.

---

## 📦 Instalasi & Kompilasi

### Prasyarat
- **Rust Toolchain** (cargo, rustc)
- **PHP Development Libraries** (`libphp` / PHP SDK)

### 1. Konfigurasi `build.rs`
Sesuaikan path ke library PHP Anda di file `build.rs` sebelum kompilasi (terutama di Windows).

### 2. Kompilasi Bridge
Jalankan perintah berikut di root folder plugin:

```bash
cargo build --release
```
*Hasil binary ada di: `target/release/php-native-bridge` (Linux/Mac) atau `php-native-bridge.exe` (Windows)*

### 3. Pemasangan Plugin
1. Pastikan `manifest.yaml` menunjuk ke binary hasil kompilasi:
   ```yaml
   binary: ./target/release/php-native-bridge
   ```
2. Aktifkan plugin di ZenoEngine (`ZENO_PLUGINS_ENABLED=true`).

---

## 💻 Penggunaan di ZenoLang

### 1. Menjalankan Script PHP
```javascript
php.run: {
    code: "echo 'Hello from Rust-Embedded PHP!';"
    as: $result
}
log: $result.output
```

### 2. Integrasi Laravel Artisan
Bridge otomatis mendeteksi file `artisan` di direktori kerja.

```javascript
// Menjalankan command artisan
php.laravel: "migrate --force" { as: $m }
log: "Migration Output: " + $m.output
```

### 3. Mengakses Data Zeno di PHP
Variabel dari Zeno otomatis disuntikkan ke superglobal `$_SERVER['ZENO_SCOPE']`.

```php
// Di dalam script PHP
$scope = json_decode($_SERVER['ZENO_SCOPE'] ?? '{}', true);
$userId = $scope['user_id'] ?? 0;
```

---

## 🏗️ Legacy (Zig)
Versi awal plugin ini dibangun menggunakan **Zig**. Source code Zig masih disimpan di folder `legacy_backup/` sebagai referensi.
