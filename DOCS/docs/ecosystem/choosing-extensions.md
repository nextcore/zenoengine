# Memilih Mekanisme Ekstensi

ZenoEngine menyediakan tiga cara berbeda untuk memperluas fungsionalitas aplikasi Anda menggunakan bahasa pemrograman lain.

| Fitur | **WASM Plugin** | **Native Sidecar** | **Container Bridge** |
| :--- | :--- | :--- | :--- |
| **Lokasi** | Internal (Sandbox) | Internal (Proses Terpisah) | Eksternal (Network) |
| **Komunikasi** | Shared Memory (Memory-pipe) | Standard I/O (JSON Stream) | HTTP / RPC |
| **Performa** | Sangat Cepat (Near-native) | Cepat (No Network Lag) | Normal (Latensi Jaringan) |
| **Lifecycle** | Sandbox VM | Dikelola ZenoEngine | Dikelola Docker |
| **Isolasi** | Ekstrim (WebAssembly) | OS Process Isolation | Container Isolation |
| **Bahasa** | Rust, C, Go, AssemblyScript | Apa saja (Executable) | Apa saja (Rest API) |

## Panduan Pengambilan Keputusan

### Gunakan WASM Plugins Jika:
- Performa adalah segalanya (misal: pengolahan citra per frame).
- Membutuhkan isolasi keamanan tingkat tinggi.
- Kode Anda bisa di-compile ke WebAssembly.

### Gunakan Native Sidecar Jika:
- Membutuhkan pustaka asli yang tidak tersedia di WASM atau Rust (misal: Python `pandas` atau PHP `gd`).
- Ingin kemudahan instalasi (cukup taruh binari, tanpa setup Docker).
- Membutuhkan interaksi dua arah (sidecar memanggil database Zeno).

### Gunakan Container Bridge Jika:
- Layanan aplikasi sudah berjalan di Docker/Kubernetes.
- Membutuhkan skalabilitas horizontal di server yang berbeda.
- Ingin ketahanan ekstrem dengan *Circuit Breaker* dan *Retry* otomatis.
