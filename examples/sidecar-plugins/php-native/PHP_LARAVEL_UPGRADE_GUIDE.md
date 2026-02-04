# 🐘 Panduan Upgrade PHP & Laravel (Rust Bridge)

Panduan ini menjelaskan cara memperbarui versi PHP pada **Rust Bridge** dan mengintegrasikan versi **Laravel** terbaru ke dalam ekosistem ZenoLang.

---

## 1. Upgrade Versi PHP pada Rust Bridge

ZenoEngine PHP-Native Bridge bekerja dengan cara melakukan *static linking* terhadap `libphp` (PHP Embedded SAPI) menggunakan crate `cc` di Rust.

### A. Persiapan Library PHP (Headers & Binaries)
Untuk menggunakan PHP 8.3 atau 8.4+, Anda membutuhkan file header (`.h`) dan library (`.lib` atau `.a`).

1.  **Windows**:
    - Download **PHP SDK** atau binary "Thread Safe" (TS) dari [windows.php.net](https://windows.php.net/download/).
    - Pastikan Anda mengambil paket `devel` yang berisi `php8ts.lib` dan folder `include`.
2.  **Linux**:
    - Install paket development: `sudo apt install php8.3-dev` (atau versi terbaru).
    - Pastikan `php-config` tersedia di PATH.

### B. Konfigurasi `build.rs`
Edit file `build.rs` di root project plugin untuk mengarahkan linker ke library PHP baru.

**Contoh Konfigurasi Windows:**
```rust
#[cfg(target_os = "windows")]
{
    // Arahkan ke folder SDK PHP yang baru didownload
    println!("cargo:rustc-link-search=native=C:/php-sdk-8.3/lib");
    println!("cargo:rustc-link-lib=static=php8");
}
```

### C. Build Ulang
Lakukan kompilasi ulang dengan Cargo. Rust akan otomatis me-link library baru.

```bash
cargo clean
cargo build --release
```

---

## 2. Integrasi Laravel Terbaru

Setelah Bridge di-update, Anda bisa menginstal Laravel versi terbaru.

### A. Konfigurasi Database Proxy
Agar Laravel menggunakan **Connection Pool** milik ZenoEngine:

1.  **Install Zeno-Laravel Provider**: Gunakan driver database khusus yang berkomunikasi via JSON-RPC.
2.  **Edit `.env` Laravel**:
    ```env
    DB_CONNECTION=zeno_proxy
    ZENO_BRIDGE_ENABLED=true
    ```

### B. Request Lifecycle
Rust Bridge v1.3+ menggunakan siklus `php_request_startup` dan `php_request_shutdown`. Ini menjamin kompatibilitas 100% dengan Laravel Service Container. Tidak ada konfigurasi khusus yang dibutuhkan di sisi Laravel; bridge akan mereset state otomatis setiap request.

---

## 3. Strategi Bundling

Untuk membuat aplikasi portable (User tidak perlu install PHP):

1.  **Static Linking**: Pastikan Anda me-link `libphp` secara statis di `build.rs` (default).
2.  **Runtime DLL (Windows)**: Di Windows, meskipun static linking, terkadang file `php8.dll` tetap dibutuhkan di folder yang sama dengan binary `.exe` kecuali Anda melakukan kompilasi PHP custom static build. Sertakan DLL ini dalam paket distribusi Anda.

---
*Dokumentasi ini diperbarui untuk ZenoEngine Rust Bridge.*
