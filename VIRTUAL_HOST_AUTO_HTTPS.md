# 🌐 Panduan Virtual Host & Automatic HTTPS (Caddy-Style)

ZenoEngine v1.4+ kini mendukung pengelolaan multi-domain (Virtual Host) dan pengamanan otomatis menggunakan **SSL/TLS dari Let's Encrypt** tanpa konfigurasi eksternal, mirip dengan cara kerja Caddy Server.

---

## ⚡ Performa Enterprise: Native Host-based Routing (O(1))

Mulai v1.8, ZenoEngine tidak lagi menggunakan pengecekan linear (O(n)) untuk Virtual Host. Kami telah mengimplementasikan **Native Host Map**.

*   **Skalabilitas Tinggi**: Waktu pencarian router tetap instan baik Anda memiliki 1 domain maupun 10.000 domain.
*   **Zero Overhead**: Dispatcher menggunakan lookup tabel hash yang sangat efisien di layer middleware awal.

---

## 1. Penggunaan Virtual Host (Domain & Subdomain)

Anda dapat mengelompokkan route berdasarkan domain atau subdomain menggunakan slot `http.host`.

### Contoh Konfigurasi di `src/main.zl`:

```javascript
// Konfigurasi untuk Subdomain API
http.host: "api.zeno.dev" {
    http.get: "/v1/status" {
        http.response: 200 {
            body: { status: "API is Running" }
        }
    }
}

// Konfigurasi untuk Domain Utama
http.host: "zeno.dev" {
    http.get: "/" {
        blade: "welcome.blade.zl"
    }
}

// Route default (jika host tidak cocok dengan di atas)
http.get: "/health" {
    http.response: 200 { body: "OK" }
}
```

---

## 2. Automatic HTTPS (Zero-Config SSL)

ZenoEngine dapat secara otomatis meminta, menginstal, dan memperbarui sertifikat SSL dari Let's Encrypt untuk semua domain yang terdaftar.

### Cara Mengaktifkan:
Tambahkan variabel berikut di file `.env` Anda:

```env
# Aktifkan fitur otomatis HTTPS
AUTO_HTTPS=true

# Domain utama aplikasi (Opsional jika sudah pakai http.host)
APP_DOMAIN=zeno.dev

# Port harus 443 untuk HTTPS
APP_PORT=:443
```

### Apa yang terjadi di balik layar?
1.  **ACME Challenge**: ZenoEngine akan membuka port `80` sementara untuk memverifikasi kepemilikan domain ke Let's Encrypt.
2.  **Certificate Storage**: Sertifikat akan disimpan dengan aman di folder `data/certs/`.
3.  **Auto-Renewal**: Sertifikat akan diperbarui secara otomatis sebelum masa berlakunya habis.

---

## 3. Reverse Proxy (Caddy-Style)

ZenoEngine v1.5+ kini mendukung fitur Reverse Proxy yang memungkinkan Anda meneruskan request ke layanan lain (seperti Node.js, Python, atau API eksternal).

### Contoh Penggunaan:

```javascript
// Meneruskan semua trafik di /api ke backend Node.js
http.proxy: "http://localhost:8080" {
    path: "/api"
}

// Meneruskan subdomain tertentu ke layanan lain
http.host: "dashboard.zeno.dev" {
    http.proxy: "http://localhost:3000"
}
```

### Keunggulan Proxy ZenoEngine:
1.  **Automatic Header Handling**: Secara otomatis mengatur header `X-Forwarded-For` dan sinkronisasi `Host`.
2.  **Zero-Configuration**: Langsung aktif sebagai slot ZenoLang.
3.  **Unified Middleware**: Anda bisa memasang middleware (seperti `auth`) sebelum request diteruskan ke backend proxy.

---

## 4. Hosting SPA & Static Site

ZenoEngine dapat bertindak sebagai file server berperforma tinggi untuk menghosting website statis atau aplikasi SPA modern (React, Vue, Angular).

### Contoh Static Site:
```javascript
http.host: "blog.zeno.dev" {
    http.static: "./my-blog-files"
}
```

### Contoh SPA (React/Vue) dengan Fallback:
Aplikasi SPA membutuhkan fitur fallback ke `index.html` agar routing sisi klien (client-side routing) berfungsi dengan benar saat halaman direfresh.

```javascript
http.host: "dashboard.zeno.dev" {
    http.static: "./dashboard/dist" {
        path: "/"
        spa: true
    }
}
```

---

## 5. Persiapan Deployment di Production

Untuk menggunakan fitur ini secara maksimal di server:

1.  **Akses Port Rendah**: Pastikan binary Zeno memiliki izin untuk menjalankan port 80 dan 443.
    ```bash
    sudo setcap 'cap_net_bind_service=+ep' ./zeno
    ```
2.  **DNS Config**: Pastikan domain/subdomain Anda sudah diarahkan (A Record) ke IP server Anda.
3.  **Firewall**: Buka port 80 (HTTP) dan 443 (HTTPS) pada firewall server Anda.

---

## 🚀 Keunggulan Dibandingkan Server Lain

| Fitur | ZenoEngine | Nginx + Certbot | Caddy |
| :--- | :--- | :--- | :--- |
| **Konfigurasi** | Di dalam kode ZenoLang | File Config Terpisah | Caddyfile |
| **SSL Otomatis** | ✅ Bawaan | ❌ Perlu Certbot | ✅ Bawaan |
| **Virtual Host** | ✅ O(1) Scalable | ❌ O(n) Linear | ✅ O(1) Scalable |
| **Latensi** | ⚡ Sangat Rendah | ⬇️ Medium (Hop+) | ⚡ Rendah |

---
*ZenoEngine: Build, Secure, and Scale in one place.*
