# 🐘 Panduan Upgrade PHP & Laravel (ZenoEngine Sidecar)

Panduan ini menjelaskan cara memperbarui versi PHP pada **Zig Bridge** dan mengintegrasikan versi **Laravel** terbaru ke dalam ekosistem ZenoLang.

---

## 1. Upgrade Versi PHP pada Zig Bridge

ZenoEngine PHP-Native Bridge bekerja dengan cara melakukan *binding* terhadap `libphp` (PHP Embedded SAPI).

### A. Persiapan Library PHP (Headers & Binaries)
Untuk menggunakan PHP 8.3 atau 8.4+, Anda membutuhkan file header (`.h`) dan library (`.lib` atau `.so`).

1.  **Windows**:
    - Download **PHP SDK** atau binary "Thread Safe" dari [windows.php.net](https://windows.php.net/download/).
    - Pastikan Anda mengambil paket `devel` yang berisi `php8ts.lib` dan folder `include`.
2.  **Linux**:
    - Install paket development: `sudo apt install php8.3-dev` (atau versi terbaru).
    - Lokasi headers biasanya ada di `/usr/include/php/`.

### B. Konfigurasi `main.zig`
Perbarui `main.zig` untuk merujuk pada header PHP yang baru. Gunakan `@cInclude` untuk menghubungkan fungsi C dari PHP ke Zig.

```zig
const php = @cImport({
    @cDefine("ZEND_WIN32", "1"); // Jika di Windows
    @cDefine("PHP_WIN32", "1");  // Jika di Windows
    @cInclude("sapi/embed/php_embed.h");
});
```

### C. Build dengan Linker Flags
Saat melakukan build menggunakan Zig, Anda harus memberitahu compiler di mana lokasi library PHP terbaru tersebut.

**Build Command (Contoh PHP 8.3):**
```bash
zig build-exe main.zig \
  -I./include/php \
  -L./lib \
  -lphp8 \
  -O ReleaseSafe \
  -target x86_64-windows
```

---

## 2. Integrasi Laravel Terbaru ke ZenoLang

Setelah Bridge siap, Anda bisa menginstal Laravel terbaru di dalam direktori project Zeno Anda.

### A. Instalasi Laravel
Jalankan composer di dalam root atau subfolder plugins:
```bash
composer create-project laravel/laravel:^11.0 my-laravel-app
```

### B. Konfigurasi Database Proxy
Agar Laravel menggunakan **Connection Pool** milik ZenoEngine (bukan koneksi database langsung), Anda harus mengkonfigurasi Laravel untuk berkomunikasi via Bridge.

1.  **Install Zeno-Laravel Provider** (Opsional/Custom):
    Buat driver database kustom di Laravel yang mengirim query via `php.db_proxy` ke Zeno.
2.  **Edit `.env` Laravel**:
    ```env
    DB_CONNECTION=zeno_proxy
    ZENO_BRIDGE_ENABLED=true
    ```

### C. Mendaftarkan Slot Laravel Baru
Anda bisa menambah slot khusus untuk fitur Laravel tertentu di ZenoLang. Edit file `manifest.yaml` di plugin PHP Anda:

```yaml
slots:
  - name: "php.laravel"
    description: "Jalankan perintah Artisan"
  - name: "laravel.route"
    description: "Panggil route Laravel secara internal"
```

Lalu di ZenoLang (`main.zl`), Anda bisa memanggilnya:

```javascript
// Memanggil Controller Laravel 11 langsung
php.laravel: "call:controller" {
    class: "App\\Http\\Controllers\\UserController"
    method: "index"
    as: $users
}

log.info: "Total users dari Laravel: " + $users.count
```

---

## 3. Strategi "Zero-Installation" (Bundling)

Untuk membuat aplikasi Anda benar-benar portable (user tidak perlu install PHP/Laravel sendiri):

1.  **Static Linking**: Gunakan Zig untuk menautkan `libphp.a` secara statis ke dalam `php_bridge.exe`.
2.  **Asset Embedding**: Gunakan fitur `@embedFile` di Zig untuk memasukkan file-file *core* Laravel yang kritikal ke dalam satu binary bridge (atau bundling via Zip).
3.  **Zeno-Bundler**: Gunakan perintah masa depan `zeno build` untuk mengemas seluruh folder `vendor/` dan `public/` Laravel menjadi satu file eksekusi Zeno.

---

## 4. Tips Performa Enterprise

*   **OPcache**: Pastikan OPcache diaktifkan di dalam inisialisasi PHP di `main.zig` agar script PHP tidak di-parse berulang kali.
*   **Worker Mode**: Gunakan pola *loop* di Zig agar proses PHP tidak mati setelah satu request (mirip RoadRunner atau FrankenPHP).
*   **Shared Memory**: Gunakan slot `scope.set` untuk membagi state antara Go (Zeno) dan PHP secara *real-time* tanpa overhead JSON yang besar.

---
*Dokumentasi ini dibuat untuk mendukung ekosistem ZenoEngine v1.5+*
