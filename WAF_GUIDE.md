# 🛡️ ZenoEngine Built-in WAF (Web Application Firewall)

ZenoEngine v1.7+ sudah dilengkapi dengan **Lightweight WAF** bawaan untuk melindungi aplikasi Anda dari serangan web yang paling umum langsung di layer engine.

---

## 1. Fitur Utama WAF

WAF ZenoEngine bekerja dengan cara memantau setiap request yang masuk dan mencocokkannya dengan pola serangan yang dikenal (Signature-based protection):

1.  **SQL Injection Protection**: Mendeteksi pola keyword SQL berbahaya seperti `UNION SELECT`, `OR 1=1`, `DROP TABLE`, dan penggunaan komentar `--` atau `/*` dalam query parameter.
2.  **XSS (Cross-Site Scripting) Protection**: Memblokir upaya penyuntikan tag `<script>`, atribut event HTML (`onerror`, `onload`), dan skema `javascript:`.
3.  **Path Traversal Protection**: Mencegah akses ilegal ke file sistem di luar direktori publik dengan memblokir pola `../` atau `..\`.
4.  **Malicious User-Agent Blocking**: Secara otomatis memblokir tool pemindaian kerentanan populer seperti `sqlmap`, `nikto`, `dirbuster`, dan `nmap`.

---

## 2. Cara Mengaktifkan

WAF tidak aktif secara default untuk memberikan fleksibilitas saat pengembangan. Untuk mengaktifkannya di produksi, tambahkan variabel ini di file `.env`:

```env
# Aktifkan WAF (true/false)
WAF_ENABLED=true
```

---

## 3. Monitoring & Log

Setiap kali WAF memblokir request, ZenoEngine akan mencatat kejadian tersebut di log `stderr` (atau log file Anda) dengan format terstruktur:

```json
{
  "level": "WARN",
  "msg": "🛡️ WAF BLOCKED REQUEST",
  "ip": "192.168.1.5:54321",
  "method": "GET",
  "path": "/users",
  "reason": "Malicious Query Parameter Detected",
  "ua": "sqlmap/1.4.11#stable"
}
```

Request yang diblokir akan menerima response **403 Forbidden** dengan body JSON:
```json
{
  "success": false,
  "error": "Blocked by WAF",
  "reason": "..."
}
```

---

## 4. Tips & False Positives

WAF menggunakan algoritma pencocokan pola yang ketat. Jika aplikasi Anda memang perlu menerima data yang mirip dengan pola SQL (misal: aplikasi belajar SQL), Anda dapat menonaktifkan WAF atau menggunakan port internal yang tidak terlindungi.

**Rekomendasi Keamanan:**
*   Aktifkan WAF bersamaan dengan `RATE_LIMIT_REQUESTS`.
*   Gunakan `AUTO_HTTPS=true` agar payload request terenkripsi sebelum diperiksa oleh WAF.

---
*ZenoEngine: High Performance, High Security.*
