# 🚀 Cara Termudah: DB Pool "Default" di Laravel

ZenoEngine v1.3+ kini menjadikan fitur Enterprise sebagai **Default**. Anda tidak perlu lagi melakukan banyak konfigurasi manual untuk mendapatkan performa maksimal.

---

## 1. Metode "One-Line" Service Provider

Alih-alih membuat driver dari nol, Anda bisa menggunakan **Zeno Bridge Package** (Konsep). Cukup tambahkan satu baris di `config/app.php` atau `bootstrap/providers.php`:

```php
// Cukup tambahkan ini, Zeno akan otomatis meng-override koneksi 'mysql' atau 'sqlite' Anda
Zeno\Laravel\BridgeServiceProvider::class,
```

### Apa yang dilakukan baris ini?
1.  Mendeteksi jika Laravel dijalankan di dalam **Zeno Sidecar** (via env vars).
2.  Secara otomatis mengganti driver database default (PDO) dengan **Zeno Proxy Driver**.
3.  Memetakan query Eloquent langsung ke Go Connection Pool.

---

## 2. Metode `auto_prepend_file` (Paling Instan)

Anda bisa menginstruksikan ZenoEngine untuk menyuntikkan script inisialisasi ke **setiap** script PHP yang dijalankan.

### Konfigurasi di ZenoEngine (`manifest.yaml` plugin):
```yaml
config:
  php_ini:
    auto_prepend_file: "zeno_bridge_init.php"
```

### Isi `zeno_bridge_init.php`:
Script ini akan dijalankan **sebelum** Laravel atau script PHP apapun dimulai.
```php
<?php
// Mengambil konfigurasi DB dari ZenoEngine secara otomatis
if (getenv('ZENO_PROXY_ENABLED')) {
    // Inject config dynamically
    // ...
}
```

---

## 3. Otomatisasi via Environment Variables

ZenoEngine secara cerdas menyuntikkan variabel lingkungan (Env) ke sidecar. Anda hanya perlu mengatur `.env` Laravel Anda agar bersifat dinamis:

```env
# Alih-alih mengisi host/user/pass, cukup aktifkan proxy
DB_CONNECTION=zeno
ZENO_MANAGED_POOL=true
```

ZenoEngine (Go) akan melihat `ZENO_MANAGED_POOL=true` dan secara otomatis menyiapkan jalur pipa (pipe) database yang sudah teroptimasi.

---

## 4. Perbandingan Kemudahan

| Metode | Tingkat Kesulitan | Rekomendasi |
| :--- | :--- | :--- |
| **Manual Driver** | ⭐⭐⭐ (Sulit) | Jika butuh kustomisasi query yang sangat spesifik. |
| **Service Provider** | ⭐⭐ (Sedang) | Standar industri untuk project Laravel. |
| **Auto Prepend** | ⭐ (Mudah) | Untuk integrasi instan tanpa menyentuh kode aplikasi. |

---

## 💡 Kesimpulan: "Set and Forget" (v1.3 Update)
Mulai versi 1.3, ZenoEngine secara otomatis mencoba mendeteksi project Laravel dan menyuntikkan bridge. Developer Laravel di tim Anda tidak perlu tahu bahwa mereka menggunakan ZenoEngine. Mereka cukup menulis kode Eloquent seperti biasa:

```php
// Ini akan otomatis menggunakan Go Connection Pool via Zeno
$users = User::where('active', true)->get();
```

ZenoEngine menangani semua aspek infrastruktur, koneksi, dan pooling di balik layar.
