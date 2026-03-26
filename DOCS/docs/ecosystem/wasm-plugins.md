# Plugins & Sidecar Architecture

ZenoEngine dirancang untuk menjadi **Polyglot**. Meskipun ZenoLang sangat cepat, ada kalanya Anda membutuhkan komputasi berat atau pustaka dari bahasa lain. Ekosistem Zeno menyediakan dua mekanisme utama untuk ekstensi lokal: **WASM Plugins** dan **Native Sidecars**.

## 1. WebAssembly (WASM) Plugins

Gunakan WASM jika Anda ingin kode tambahan berjalan dengan performa mendekati native di dalam memori ZenoEngine namun tetap terisolasi dalam *sandbox*.

### Zero CGO Engine
ZenoEngine menggunakan [wazero](https://wazero.io), mesin WebAssembly tanpa CGO. Ini berarti Zeno tetap *portable* (single binary) tanpa membutuhkan toolchain C di server produksi.

### Build & Deploy
Anda bisa menulis plugin dalam bahasa apapun yang mendukung komputasi ke `.wasm`:
- **Rust**, **Go** (TinyGo), **AssemblyScript**, **Zig**, **C/C++**.

Tempatkan file `.wasm` di folder `plugins/` bersama `manifest.yaml`.

---

## 2. Native Managed Sidecars

Ini adalah fitur tingkat lanjut untuk menjalankan proses binari asli (native executable) yang dikelola langsung oleh ZenoEngine.

### Cara Kerja
ZenoEngine akan melakukan `spawn` proses tersebut dan berkomunikasi menggunakan **Standard I/O (JSON RPC)**. Ini sangat berguna jika Anda ingin menggunakan fitur bahasa yang tidak bisa di-compile ke WASM (misal: Python scripts atau PHP CLI).

### Fitur Utama Sidecar:
- **Lifecycle Management**: Zeno otomatis menjalankan sidecar saat startup dan mematikannya saat shutdown.
- **Auto-Healing**: Jika proses sidecar *crash*, Zeno akan merestart-nya secara otomatis hingga 5 kali.
- **Bi-directional RPC**: Sidecar bisa memanggil fungsi internal Zeno (database, scope, log) melalui protokol JSON khusus.

---

## Manifest Configuration (`manifest.yaml`)

Baik WASM maupun Native Sidecar dikonfigurasi melalui manifes yang sama:

```yaml
id: "my_extension"
name: "Power Processor"
version: "1.0.0"
type: "wasm" # Atau "sidecar"
binary: "processor.wasm" # Atau "binary.exe"
permissions:
  network: false
  filesystem: true
  database: true
  scope: true
```

## Host Functions (The Bridge)

Ekstensi Anda tidak berjalan "buta". Zeno menyediakan **Host Functions** agar kode Anda bisa berinteraksi dengan lingkungan:
- `db_query`: Eksekusi SQL langsung.
- `scope_get` / `scope_set`: Mengambil/mengubah variabel di skrip ZenoLang `.zl`.
- `http_request`: Melakukan *outgoing call*.
- `log`: Mengirim log ke terminal utama ZenoEngine.

> [!TIP]
> Baca panduan [Memilih Ekstensi](./choosing-extensions.md) untuk menentukan mana yang terbaik bagi proyek Anda.
