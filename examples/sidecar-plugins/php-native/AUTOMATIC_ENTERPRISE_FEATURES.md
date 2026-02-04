# 🤖 Fitur Enterprise Otomatis (Auto-Cure PHP)

ZenoEngine v1.3+ hadir dengan filosofi **"Zero-Config Enterprise"**, di mana limitasi tradisional PHP diatasi secara otomatis oleh arsitektur **Native Bridge**.

---

## 1. Auto-Healing (Self-Recovery)
PHP seringkali berhenti mendadak karena *memory exhaustion* atau *fatal error*. ZenoEngine memantau kesehatan proses Sidecar secara *real-time*.

*   **Cara Kerja**: Jika bridge PHP crash, ZenoEngine akan mendeteksinya dan melakukan restart otomatis dengan strategi *exponential backoff*.
*   **Keuntungan**: Aplikasi Anda tetap online meskipun ada script PHP yang tidak stabil.

## 2. Automatic State Persistence (v1.3 Default)
Secara default, Sidecar berjalan dalam mode **Managed Stateful**.

*   **Otomatisasi**: Interpreter PHP tetap hidup di memori. Namun, berkat implementasi `Request Lifecycle` di bridge Rust, state request (Global Variables) di-reset otomatis setiap kali request selesai.
*   **Performa**: Menghilangkan overhead inisialisasi awal engine PHP, namun tetap menjamin kebersihan memori antar request (seperti PHP-FPM).

## 3. Global Session & Scope Sync
ZenoEngine secara otomatis menyinkronkan data antara scope ZenoLang dan PHP.

*   **Deep Injection**: Variabel `$user`, `$cart`, atau `$session` di ZenoLang otomatis tersedia di PHP melalui `$_SERVER['ZENO_SCOPE']`.
*   **Bi-directional**: Data dikirim dalam format JSON yang aman dan efisien.

## 4. Unified Error Stream (AI-Native)
Kesalahan yang terjadi di PHP kini diproses oleh Zeno Diagnostic System.

*   **Structured Logs**: Output dari `stderr` sidecar ditangkap oleh ZenoEngine.
*   **Panic Protection**: Bridge Rust membungkus eksekusi PHP dalam blok `try-catch` (dan output buffering) untuk mencegah output error liar merusak protokol komunikasi.

---

## 5. Ringkasan Otomatisasi

| Limitasi PHP | Status di ZenoEngine | Mekanisme Otomatis |
| :--- | :--- | :--- |
| **Crashes** | ✅ **Auto-Healed** | Process Watchdog & Restart |
| **Stateless** | ✅ **Persistent** | Embedded SAPI |
| **Slow DB** | ✅ **Pooled** | Go DB Proxy (Default) |
| **Sync Data** | ✅ **Synced** | Automatic Scope Injection |
| **Request Isolation** | ✅ **Safe** | `php_request_shutdown` Loop |

---
*Dengan fitur-fitur ini, ZenoEngine mengubah PHP menjadi runtime enterprise yang tangguh dan modern.*
