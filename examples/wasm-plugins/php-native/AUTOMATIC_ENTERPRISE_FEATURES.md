# 🤖 Fitur Enterprise Otomatis (Auto-Cure PHP)

ZenoEngine v1.3+ hadir dengan filosofi **"Zero-Config Enterprise"**, di mana limitasi tradisional PHP diatasi secara otomatis oleh engine tanpa perlu konfigurasi tambahan.

---

## 1. Auto-Healing (Self-Recovery)
PHP seringkali berhenti mendadak karena *memory exhaustion* atau *fatal error*. ZenoEngine kini memantau kesehatan proses Sidecar secara *real-time*.

*   **Cara Kerja**: Jika bridge PHP crash, ZenoEngine akan mendeteksinya dan melakukan restart otomatis dengan strategi *exponential backoff*.
*   **Keuntungan**: Aplikasi Anda tetap online meskipun ada script PHP yang tidak stabil.

## 2. Automatic State Persistence (v1.3 Default)
Secara default, Sidecar kini berjalan dalam mode **Managed Stateful**.

*   **Otomatisasi**: Variabel statis dan inisialisasi framework (seperti Service Container Laravel) tetap terjaga di memori antar panggilan slot.
*   **Performa**: Menghilangkan overhead bootstrapping PHP (~50-100ms) pada setiap request.

## 3. Global Session & Scope Sync
ZenoEngine secara otomatis menyinkronkan data antara scope ZenoLang dan PHP.

*   **Deep Injection**: Variabel `$user`, `$cart`, atau `$session` di ZenoLang otomatis tersedia di dalam superglobal PHP (`$_SESSION` atau `$_ZENO`) melalui filter bridge otomatis.
*   **Bi-directional**: Perubahan data di PHP dapat dikirim balik ke ZenoLang secara instan.

## 4. Unified Error Stream (AI-Native)
Kesalahan yang terjadi di PHP kini diproses oleh Zeno Diagnostic System.

*   **Structured Logs**: `Fatal Error` atau `Warning` dari PHP ditangkap dari StdErr sidecar, diparsing, dan ditampilkan sebagai **Structured Diagnostic JSON** di ZenoEngine.
*   **AI Debugging**: Karena formatnya JSON, AI Agent dapat langsung menganalisa error PHP tersebut dan menyarankan perbaikan kode.

---

## 5. Ringkasan Otomatisasi

| Limitasi PHP | Status di ZenoEngine | Mekanisme Otomatis |
| :--- | :--- | :--- |
| **Crashes** | ✅ **Auto-Healed** | Process Watchdog & Restart |
| **Stateless** | ✅ **Persistent** | Managed Stateful Worker |
| **Slow DB** | ✅ **Pooled** | Go DB Proxy (Default) |
| **Sync Data** | ✅ **Synced** | Automatic Scope Injection |
| **Async** | ✅ **Supported** | Zeno `async:` slot handling |

---
*Dengan fitur-fitur ini, ZenoEngine mengubah PHP menjadi runtime enterprise yang tangguh dan modern.*
