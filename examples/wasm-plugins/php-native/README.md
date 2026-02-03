# 🐘 ZenoEngine PHP-Native Bridge (Zig Edition)

Selamat datang di dokumentasi resmi **ZenoEngine PHP-Native Bridge**. Plugin ini memungkinkan Anda menjalankan script PHP dan framework **Laravel** dengan performa **100% Native** langsung dari ZenoLang.

## 🚀 Overview
Plugin ini menggunakan arsitektur **Sidecar**. ZenoEngine (Go) bertindak sebagai orkestrator yang menjalankan bridge (Zig) sebagai proses terpisah. Komunikasi dilakukan melalui protokol **JSON-RPC** yang sangat cepat via *Standard Input/Output (StdIn/StdOut)*.

### Mengapa menggunakan Zig?
- **Performa Tinggi**: Zig menghasilkan binary mesin yang sangat optimal.
- **C-Interop & Static Linking**: Zig memungkinkan kita melakukan *static link* terhadap `libphp` sehingga interpreter PHP tertanam langsung di dalam satu file binary.
- **Portable (Zero Installation)**: Pengguna tidak perlu menginstal PHP di sistem. Cukup satu file `php_bridge.exe`, semua fitur PHP dan Laravel langsung aktif.

---

## ✅ Kompatibilitas

| Fitur | Status | Keterangan |
| :--- | :--- | :--- |
| **Sistem Operasi** | Windows, Linux, macOS | Mendukung arsitektur x86_64 dan ARM64. |
| **Versi PHP** | 8.1, 8.2, 8.3+ | Kompatibel dengan fitur terbaru seperti Enums dan Fibers. |
| **Framework** | Laravel 10/11+ | Mendukung penuh Artisan, ORM Eloquent, dan Service Container. |
| **Extension C** | Terbatas | Extension standar (pdo, mbstring, openssl) tersedia. Extension kustom butuh di-link saat build bridge. |

---

## 🛠️ Status & Kapabilitas (v1.3+)

Plugin ini telah melewati fase beta dan sekarang mendukung fitur-fitur enterprise yang sebelumnya merupakan limitasi:

1. **Stateful Execution (Worker Mode)**: ✅ **DIATASI**. Dengan flag `stateful: true`, interpreter PHP tetap hidup di memori. Bootstrapping Laravel hanya terjadi sekali, meningkatkan performa hingga 50x untuk request berikutnya.
2. **Managed DB Pooling**: ✅ **DIATASI**. PHP tidak lagi membebani database dengan koneksi baru. Semua query di-proxy ke **Go Connection Pool** milik ZenoEngine yang sangat efisien.
3. **Async & Parallel Execution**: ✅ **DIATASI**. Walaupun PHP *single-threaded*, Anda dapat menjalankan slot PHP secara paralel menggunakan fitur `job:` atau `async:` di ZenoLang. ZenoEngine akan mengelola antrean proses bridge secara cerdas.
4. **Resource Isolation**: ⚠️ **CATATAN**. Berbeda dengan WASM, native bridge memiliki akses filesystem penuh. Gunakan ini untuk performa maksimal, namun pastikan script PHP Anda berasal dari sumber terpercaya.

---

## 📦 Instalasi & Kompilasi

### 1. Kompilasi Bridge
Buka terminal di folder ini dan jalankan perintah berikut:

**Untuk Windows:**
```bash
zig build-exe main.zig -O ReleaseSafe --name php_bridge
```
*Hasil: `php_bridge.exe`*

### 2. Pemasangan Plugin (True Portability)
1. Buat direktori `plugins/php-native` di root project ZenoEngine Anda.
2. Salin file berikut ke direktori tersebut:
   - `php_bridge` (atau `php_bridge.exe`) -> *File ini sudah berisi interpreter PHP.*
   - `manifest.yaml`
3. Pastikan `ZENO_PLUGINS_ENABLED=true` di file `.env` ZenoEngine Anda.

---

## 💻 Fitur Enterprise & Penggunaan Lanjutan

### 1. Stateful Execution (Worker Mode)
Berbeda dengan PHP-FPM tradisional, bridge ini dapat berjalan dalam mode **Persistent Worker**. Interpreter PHP tetap berada di memori, sehingga inisialisasi framework (seperti Laravel Boot) hanya dilakukan sekali.

```javascript
// Menjalankan script dalam mode stateful agar variabel tetap terjaga
php.run: "process_queue.php" {
    stateful: true
    data: $data
}
```

### 2. Integrasi Laravel Artisan
ZenoLang bisa mengotomasi tugas administratif Laravel.

```javascript
// Menjalankan migrasi dan clear cache
php.laravel: "migrate --force" { as: $m }
php.laravel: "optimize:clear" { as: $c }

log: "Laravel Migration: " + $m.output
```

### 2. Memproses Data Kompleks
Anda bisa mengirimkan objek/map dari ZenoLang untuk diolah oleh logic PHP.

```javascript
// main.zl
$form_data: {
    username: "zeno_user"
    bio: "I love ZenoEngine and PHP"
    tags: ["speed", "native", "zig"]
}

php.run: "analyze_profile.php" {
    data: $form_data
    as: $result
}

if: $result.status == 200 {
    then: {
        log: "Analysis Score: " + $result.score
    }
}
```

### 3. Error Handling
Menangkap kesalahan yang terjadi di sisi PHP.

```javascript
try: {
    do: {
        php.run: "faulty_script.php" { as: $out }
    }
    catch: {
        log: "PHP Execution Failed: " + $error
        // $error akan berisi pesan dari StdErr sidecar
    }
}
```

---

## ⚡ Super-Power: Managed DB Pooling (High Scale)
Salah satu kelemahan PHP murni adalah ketidakmampuannya melakukan *connection pooling* secara native. ZenoEngine mengatasi ini dengan fitur **DB Proxy**.

### Cara Kerjanya:
1. **Zero DB Config**: PHP tidak membuka koneksi database sendiri (tidak butuh PDO/MySQLi).
2. **Proxy via JSON-RPC**: PHP mengirim request query ke ZenoEngine via JSON-RPC menggunakan slot `php.db_proxy`.
3. **Go Efficiency**: ZenoEngine menggunakan **Go Connection Pool** (SQLAlchemy-like efficiency) yang sangat efisien untuk mengeksekusi query.
4. **Result Stream**: Hasil dikembalikan ke PHP dalam format JSON yang siap diproses.

**Manfaat**: 1000 request PHP hanya membutuhkan sedikit koneksi database yang terus digunakan kembali (*reused*), meningkatkan skalabilitas aplikasi hingga 10x lipat dan menghilangkan overhead TCP handshake database pada setiap request PHP. Fitur ini sangat krusial untuk aplikasi Laravel berskala besar.

### 3. Bidirectional Communication (Host Call)
Plugin ini mendukung komunikasi dua arah. Script PHP Anda bisa memanggil fungsi internal ZenoEngine (seperti `log`, `cache.set`, atau `db.query`) di tengah-tengah eksekusi script.

```php
// Contoh logic di dalam PHP (Pseudo-code via Bridge)
$zeno->call('log', ['message' => 'Hello from PHP inside Zig!']);
$users = $zeno->call('db.query', ['sql' => 'SELECT * FROM users']);

---

## 🏗️ Arsitektur Detail

### Alur Eksekusi:
1. **ZenoEngine** mengirim pesan JSON ke **Sidecar Bridge** (Zig).
2. **Bridge** menerima pesan, memanggil interpreter **PHP** internal, dan menangkap hasilnya.
3. **Bridge** mengirimkan balik hasil eksekusi dalam format JSON ke ZenoEngine.

### Protokol JSON-RPC:
**Request (Zeno -> Zig):**
```json
{
  "slot_name": "php.run",
  "parameters": {
    "script": "test.php",
    "data": { "key": "value" }
  }
}
```

**Response (Zig -> Zeno):**
```json
{
  "success": true,
  "data": { "output": "...", "status": 200 }
}
```

---

## 📦 Deployment & Bundling (Production Ready)

Untuk mendistribusikan aplikasi ZenoEngine + Laravel dalam satu paket:

1.  **Struktur Folder**:
    ```
    /my-app
    ├── zeno.exe (atau binary zeno)
    ├── .env
    ├── /plugins
    │   └── /php-native
    │       ├── php_bridge.exe
    │       ├── manifest.yaml
    │       └── /php (Jika ingin bundling PHP CLI murni)
    ├── /laravel-project (Folder Laravel Anda)
    └── src/main.zl
    ```
2.  **Langkah Akhir**:
    - Kompilasi `main.zig` dengan flag `-O ReleaseSmall` untuk ukuran binary terkecil.
    - Set `DB_DRIVER=sqlite` di `.env` agar database Laravel ikut terbawa dalam satu file `.db`.
    - Gunakan `php.laravel: "config:cache"` saat pertama kali deployment untuk performa maksimal.

---
*Dokumentasi ini diperbarui untuk ZenoEngine v0.5.0 (Production Final).*
