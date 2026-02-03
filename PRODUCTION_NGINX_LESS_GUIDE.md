# 🌐 Panduan Deployment Tanpa Nginx (Front-Facing ZenoEngine)

ZenoEngine dirancang menggunakan stack **Go standard library (`net/http`)** yang sangat performan dan aman. Mulai v1.4, ZenoEngine sudah sanggup melayani trafik internet secara langsung tanpa perlu reverse proxy seperti Nginx.

---

## 🛡️ Fitur Keamanan Bawaan

ZenoEngine sudah dilengkapi dengan "Armor" produksi:

1.  **Rate Limiting**: Melindungi dari serangan Brute Force dan DDoS ringan.
    - Konfigurasi via `.env`: `RATE_LIMIT_REQUESTS=100`, `RATE_LIMIT_WINDOW=60`.
2.  **CSRF Protection**: Melindungi form dari serangan Cross-Site Request Forgery secara otomatis.
3.  **Security Headers (Helmet-like)**:
    - `X-Frame-Options: DENY` (Anti Clickjacking).
    - `X-Content-Type-Options: nosniff`.
    - `Strict-Transport-Security` (HSTS) otomatis aktif di `APP_ENV=production`.
    - **CSP (Content Security Policy)** yang dapat dikonfigurasi.
4.  **Gzip Compression**: Mengurangi ukuran payload hingga 70% untuk akses yang lebih cepat.
5.  **CORS**: Pengaturan domain yang diizinkan secara granular.

---

## 🔒 Konfigurasi HTTPS (TLS)

ZenoEngine mendukung SSL/TLS secara native. Anda hanya perlu menyediakan file sertifikat:

### 1. Siapkan Sertifikat (Let's Encrypt / Cloudflare)
Pastikan Anda memiliki file `.crt` (atau `.pem`) dan `.key`.

### 2. Update `.env`
```env
APP_ENV=production
APP_PORT=:443

# Path ke sertifikat SSL
SSL_CERT_PATH=/etc/letsencrypt/live/domain.com/fullchain.pem
SSL_KEY_PATH=/etc/letsencrypt/live/domain.com/privkey.pem
```

ZenoEngine akan mendeteksi path tersebut dan otomatis menjalankan server via **HTTPS**.

---

## 🚀 Mengapa Tanpa Nginx?

1.  **Lower Latency**: Menghilangkan satu hop network tambahan (Nginx -> Zeno).
2.  **Simplified Stack**: Cukup satu binary ZenoEngine untuk menjalankan seluruh aplikasi + server.
3.  **Low Resource**: Go server sangat ringan dalam penggunaan RAM dibandingkan menjalankan Nginx + PHP-FPM secara bersamaan.
4.  **Native Websockets**: ZenoEngine mendukung websocket (untuk Live Reload atau Fitur Real-time) tanpa perlu konfigurasi proxy yang rumit.

---

## 💡 Rekomendasi Deployment

Meskipun sanggup berjalan sendiri, kami merekomendasikan:

*   **Systemd**: Gunakan systemd unit untuk memastikan ZenoEngine restart otomatis jika server reboot.
*   **Port 80 to 443**: Jika Anda menggunakan port 443, pastikan binary Zeno memiliki izin (`setcap 'cap_net_bind_service=+ep' zeno`) atau dijalankan di belakang Load Balancer Cloud (seperti AWS ALB atau Cloudflare).
*   **Cloudflare**: Gunakan Cloudflare di depan ZenoEngine untuk perlindungan DDoS tingkat lanjut dan caching edge.

---
*ZenoEngine: Modern, Secure, and Production-Ready.*
