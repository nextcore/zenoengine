# 🧩 Panduan Menambah Ekstensi PHP (Custom Build)

Karena ZenoEngine PHP Bridge menggunakan **Static Linking**, Anda tidak bisa sekadar mengaktifkan ekstensi via `php.ini` (`extension=gd.so`). Semua ekstensi harus dikompilasi dan disatukan ke dalam file binary `.exe` bridge.

Panduan ini menjelaskan cara menambahkan ekstensi (seperti GD, BCMath, Redis, dll) ke dalam build Rust Anda.

---

## 1. Konsep Static Extension

Pada PHP Embed SAPI, ekstensi harus tersedia saat proses linking (penyambungan kode) oleh compiler Rust.

1.  **Library File**: Anda butuh file `.lib` (Windows) atau `.a` (Linux) dari ekstensi tersebut.
2.  **Rust Linker**: `build.rs` harus diberitahu untuk mengambil library tersebut.

---

## 2. Persiapan (Windows)

Jika Anda menggunakan PHP SDK (Binary Tools) dari windows.php.net:

1.  Buka folder SDK PHP Anda (misal `C:\php-sdk`).
2.  Cari folder `lib` atau `ext`.
3.  Pastikan ada file library untuk ekstensi yang diinginkan, misalnya:
    - `php_gd.lib`
    - `php_mbstring.lib`
    - `php_openssl.lib`

---

## 3. Modifikasi `build.rs`

Buka file `examples/sidecar-plugins/php-native/build.rs`. Tambahkan instruksi linking untuk setiap ekstensi.

```rust
fn main() {
    #[cfg(target_os = "windows")]
    {
        // Path ke library PHP SDK
        println!("cargo:rustc-link-search=native=C:/php-sdk/lib");

        // Link Core PHP
        println!("cargo:rustc-link-lib=static=php8");

        // --- TAMBAHAN EKSTENSI ---
        // Link GD (Grafik)
        println!("cargo:rustc-link-lib=static=php_gd");

        // Link MBString (Wajib untuk Laravel)
        println!("cargo:rustc-link-lib=static=php_mbstring");

        // Link OpenSSL
        println!("cargo:rustc-link-lib=static=php_openssl");
        // Catatan: OpenSSL mungkin butuh libcrypto.lib dan libssl.lib tambahan dari OpenSSL SDK
    }

    // ... konfigurasi linux ...
}
```

---

## 4. Modifikasi Source Code (Opsional)

Untuk beberapa ekstensi yang tidak otomatis dimuat (*static build*), Anda mungkin perlu memanggil fungsi inisialisasi modul secara manual di `src/php.rs` sebelum `php_embed_init`, meskipun pada build modern biasanya linker sudah menangani ini via simbol.

Jika Anda mengalami error "Extension not found" padahal sudah di-link, pastikan Anda menggunakan build PHP yang dikonfigurasi dengan `--enable-all-static` atau setara.

---

## 5. Build Ulang

Jalankan perintah build untuk menghasilkan binary baru yang sudah "gemuk" berisi ekstensi tersebut.

```bash
cargo clean
cargo build --release
```

Sekarang binary `php-native-bridge.exe` Anda sudah memiliki kemampuan grafis (GD) atau enkripsi (OpenSSL) secara *native* tanpa perlu file DLL eksternal!
