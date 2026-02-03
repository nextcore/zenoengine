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

## ⚠️ Limitasi

1. **Stateless Execution**: Secara default, variabel PHP tidak bertahan di memori antar panggilan slot kecuali Anda menggunakan *Persistent State* di level bridge (mirip FrankenPHP worker).
2. **Blocking Nature**: Slot `php.run` bersifat sinkron. ZenoEngine akan menunggu hasil eksekusi PHP selesai sebelum lanjut ke baris berikutnya (kecuali dijalankan di dalam worker asinkron).
3. **Resource Isolation**: Meskipun berjalan di proses terpisah, plugin ini tidak memiliki sandbox seketat WASM. Plugin memiliki izin akses penuh ke filesystem sesuai user yang menjalankan ZenoEngine.
4. **Multithreading**: PHP pada dasarnya *single-threaded*. Pemanggilan paralel dari ZenoLang akan ditangani secara antrean oleh satu proses bridge, atau Anda harus menyalakan beberapa instance bridge.

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

## 💻 Contoh Penggunaan Lanjutan

### 1. Integrasi Laravel Artisan
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
1. PHP tidak membuka koneksi database sendiri (tidak butuh PDO/MySQLi).
2. PHP mengirim request query ke ZenoEngine via JSON-RPC.
3. ZenoEngine menggunakan **Go Connection Pool** yang sangat efisien untuk mengeksekusi query.
4. Hasil dikembalikan ke PHP.

**Manfaat**: 1000 request PHP hanya membutuhkan sedikit koneksi database yang terus digunakan kembali (*reused*), meningkatkan skalabilitas aplikasi hingga 10x lipat.

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
*Dokumentasi ini diperbarui untuk ZenoEngine v0.5.0.*
