# 🤖 Fitur Enterprise Otomatis (Auto-Cure PHP)

ZenoEngine v1.3+ hadir dengan filosofi **"Zero-Config Enterprise"**, di mana limitasi tradisional PHP diatasi secara otomatis oleh arsitektur **Native Bridge (Rust)**.

---

## 1. Auto-Healing (Self-Recovery)
PHP seringkali berhenti mendadak karena *memory exhaustion* atau *fatal error*. ZenoEngine memantau kesehatan proses Sidecar secara *real-time*.

*   **Cara Kerja**: Jika bridge PHP crash, ZenoEngine akan mendeteksinya dan melakukan restart otomatis dengan strategi *exponential backoff*.
*   **Keuntungan**: Aplikasi Anda tetap online meskipun ada script PHP yang tidak stabil.

## 2. Automatic State Persistence (v1.3 Default)
Secara default, Sidecar berjalan dalam mode **Managed Stateful**.

*   **Otomatisasi**: Interpreter PHP tetap hidup di memori. Namun, berkat implementasi `Request Lifecycle` di bridge Rust (`php_request_startup` / `shutdown`), state request (Global Variables) di-reset otomatis setiap kali request selesai.
*   **Performa**: Menghilangkan overhead inisialisasi awal engine PHP (Module Load, Extension Init), namun tetap menjamin kebersihan memori antar request (seperti PHP-FPM).

## 3. Global Session & Scope Sync
ZenoEngine secara otomatis menyinkronkan data antara scope ZenoLang dan PHP.

*   **Deep Injection**: Variabel `$user`, `$cart`, atau `$session` di ZenoLang otomatis tersedia di PHP melalui `$_SERVER['ZENO_SCOPE']`.
*   **Bi-directional**: Data dikirim dalam format JSON yang aman dan efisien.

## 4. Unified Error & Output Stream
Kesalahan dan Output dari PHP ditangani secara khusus agar tidak merusak protokol komunikasi.

*   **Output Capture (Temp File Strategy)**: Bridge menggunakan mekanisme `ob_start()` dan file sementara untuk menangkap output `echo` dari PHP. Ini memastikan output biner atau teks besar dapat dikirim balik ke Zeno dengan aman tanpa tercampur log debug di StdOut.
*   **Panic Protection**: Eksekusi PHP dibungkus dalam blok try-catch di level script wrapper untuk menangkap Exception yang tidak tertangani.

---

## 5. Ringkasan Otomatisasi

| Limitasi PHP | Status di ZenoEngine | Mekanisme Otomatis |
| :--- | :--- | :--- |
| **Crashes** | ✅ **Auto-Healed** | Process Watchdog & Restart |
| **Stateless** | ✅ **Persistent** | Embedded SAPI (Persistent process) |
| **Request Isolation** | ✅ **Safe** | `php_request_shutdown` Loop |
| **Sync Data** | ✅ **Synced** | Automatic Scope Injection (`$_SERVER['ZENO_SCOPE']`) |
| **Output** | ✅ **Buffered** | Temp File Output Capture |

---
*Dengan fitur-fitur ini, ZenoEngine mengubah PHP menjadi runtime enterprise yang tangguh dan modern.*
