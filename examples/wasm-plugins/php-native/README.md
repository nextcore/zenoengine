# 🐘 ZenoEngine PHP-Native Bridge (Zig Edition)

Selamat datang di dokumentasi resmi **ZenoEngine PHP-Native Bridge**. Plugin ini memungkinkan Anda menjalankan script PHP dan framework **Laravel** dengan performa **100% Native** langsung dari ZenoLang.

## 🚀 Overview
Plugin ini menggunakan arsitektur **Sidecar**. ZenoEngine (Go) bertindak sebagai orkestrator yang menjalankan bridge (Zig) sebagai proses terpisah. Komunikasi dilakukan melalui protokol **JSON-RPC** yang sangat cepat via *Standard Input/Output (StdIn/StdOut)*.

### Mengapa menggunakan Zig?
- **Performa Tinggi**: Zig menghasilkan binary mesin yang sangat optimal.
- **C-Interop**: Memudahkan pemanggilan `libphp` secara langsung di masa depan.
- **Portable**: Memungkinkan kompilasi untuk Windows (.exe) tanpa dependensi runtime tambahan.

---

## 🛠️ Prasyarat
Sebelum memulai, pastikan Anda memiliki:
1. **ZenoEngine** terpasang.
2. **Zig Compiler** (v0.13.0 atau terbaru) terpasang di path sistem.
3. **PHP CLI** terpasang (untuk menjalankan script PHP).

---

## 📦 Instalasi & Kompilasi

### 1. Kompilasi Bridge
Buka terminal di folder ini dan jalankan perintah berikut:

**Untuk Windows:**
```bash
zig build-exe main.zig -O ReleaseSafe --name php_bridge
```
*Hasil: `php_bridge.exe`*

**Untuk Linux/macOS:**
```bash
zig build-exe main.zig -O ReleaseSafe --name php_bridge
```
*Hasil: `php_bridge`*

### 2. Pemasangan Plugin
1. Buat direktori `plugins/php-native` di root project ZenoEngine Anda.
2. Salin file berikut ke direktori tersebut:
   - `php_bridge` (atau `php_bridge.exe`)
   - `manifest.yaml`
3. Pastikan `ZENO_PLUGINS_ENABLED=true` di file `.env` ZenoEngine Anda.

---

## ⚙️ Konfigurasi (manifest.yaml)

File `manifest.yaml` mengatur bagaimana ZenoEngine mengenali dan menjalankan plugin ini.

```yaml
name: php-native
version: 1.0.0
type: sidecar          # Menggunakan tipe sidecar (bukan wasm)
binary: ./php_bridge   # Path ke file binary hasil compile zig

sidecar:
  protocol: json-rpc   # Protokol komunikasi
  auto_start: true     # Nyalakan otomatis saat ZenoEngine start
  keep_alive: true     # Restart otomatis jika proses mati

permissions:
  filesystem: ["*"]    # Izin akses file untuk PHP
  network: ["*"]       # Izin akses jaringan
```

---

## 💻 Penggunaan di ZenoLang

Setelah terpasang, Anda akan mendapatkan akses ke slot `php.*`.

### 1. Menjalankan Perintah Laravel Artisan
ZenoLang bisa memerintah Laravel untuk melakukan tugas-tugas administratif.

```javascript
// main.zl
php.laravel: "migrate --force" {
    as: $result
}
log: $result.output
```

### 2. Menjalankan Script PHP Kustom
Anda bisa mengirim data dari ZenoLang ke PHP untuk diproses secara intensif.

```javascript
// main.zl
$data: {
    user_id: 42
    action: "generate_report"
}

php.run: "process.php" {
    params: $data
    as: $php_output
}

log: "Status: " + $php_output.status
```

---

## 🏗️ Arsitektur Detail

### Alur Eksekusi:
1. **ZenoEngine** membaca `main.zl`.
2. Saat menemui slot `php.*`, ZenoEngine mengirim pesan JSON ke **Sidecar Bridge** (Zig).
3. **Bridge** menerima pesan, memanggil interpreter **PHP**, dan menangkap hasilnya.
4. **Bridge** mengirimkan balik hasil eksekusi dalam format JSON ke ZenoEngine.
5. ZenoEngine memasukkan hasil tersebut ke dalam variabel (**Scope**) ZenoLang.

### Protokol JSON-RPC:
**Request (Zeno -> Zig):**
```json
{
  "slot_name": "php.run",
  "parameters": { "script": "test.php" }
}
```

**Response (Zig -> Zeno):**
```json
{
  "success": true,
  "data": { "output": "Hello from PHP!", "status": 200 }
}
```

---

## ❓ Troubleshooting
- **Error: Sidecar process not found**: Pastikan path `binary` di `manifest.yaml` sudah benar dan file sudah di-compile.
- **Permission Denied**: Di Linux/macOS, pastikan file binary memiliki izin eksekusi (`chmod +x php_bridge`).
- **PHP Not Found**: Pastikan binary PHP dapat diakses oleh proses bridge.

---
*Dokumentasi ini dibuat secara otomatis oleh ZenoEngine Assistant.*
