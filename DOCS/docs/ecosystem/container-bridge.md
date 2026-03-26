# Container Bridge (Docker RPC)

ZenoEngine dirancang secara fundamental sebagai *"High-Concurrency Web Orchestrator"*. Ia melibas lalulintas I/O dengan luar biasa cepat, namun terkadang Anda memiliki skenario matematis berat (seperti kalkulasi perpajakan raksasa, *Machine Learning*, atau konversi Video) yang lebih cocok diselesaikan oleh bahasa pemrograman fungsional seperti Python, Java, atau PHP.

Untuk menunjang arsitektur **Microservices** ini, sistem *Container Bridge Native Slot* memungkinkan skrip ZenoLang (`.zl`) untuk berkomunikasi dengan Docker Container eksternal seakan memanggil fungsi lokal.

---

## 🚀 Mengapa Menggunakan Container Bridge?

1. **Skalabilitas Tak Terbatas:** Anda dapat merespon *traffic webhook* secara *real-time* dengan ZenoEngine, dan mengirimkannya diam-diam ke lusinan Container *Worker* Node/Python (Sidecar Pattern) secara merata.
2. **Isolasi Memori (Keamanan Server):** Merender file Excel raksasa sering membuat server *Out of Memory*. Jika Container `node_excel_printer` hancur meledak, proses ZenoEngine Anda akan tetap sehat bugar dan tidak terganggu sama sekali.
3. **Multi-Bahasa (Polyglot):** Memanfaatkan C-Extensions miliaran baris milik Python (TensorFlow/Pandas) tanpa perlu menulisnya menggunakan API primitif HTTP secara manual. ZenoEngine yang merangkum *header*, sesi HTTP, dan penjahitan JSON.

---

## 🛠️ API Reference

### 1. `docker.health`
Memeriksa status layanan target kontainer sebelum mengirimkan muatan *(payload)* HTTP. Berperan ganda sebagai pengecek *Circuit Breaker*.

```zeno
docker.health: "python_worker_ai" {
    port: 8000
    as: $status_kesehatan
}

log.info: $status_kesehatan
```

**Hasil Objek:**
```json
{
  "status": "healthy", 
  "host": "python_worker_ai", 
  "port": "8000"
} 
// Jika mati, mengembalikan `{"status": "unhealthy", "error": "HTTP 502"}`
```

---

## 2. Load Balancing & Service Discovery (`docker.nodes`)

ZenoEngine memiliki sistem *Service Discovery* bawaan. Anda bisa mendaftarkan node secara statis atau dinamis.

### Registrasi Statis (di ZenoLang)
```zeno
docker.nodes: "payment_service" {
    nodes: "10.0.0.1:8080, 10.0.0.2:8080"
    weight: 200 # Node ini lebih kuat
    check: "/health"
}
```

### Registrasi Dinamis (Auto-Join API)
Node eksternal (Python, Node, Go) bisa mendaftarkan dirinya sendiri ke ZenoEngine tanpa perlu menyentuh konfigurasi Zeno:

```bash
curl -X POST http://zeno-host:3000/api/zeno/register \
  -d '{
    "service": "ai_service",
    "host": "10.0.5.20",
    "port": 5000,
    "weight": 100,
    "ttl": 60
  }'
```

- **Weight**: Mengatur pembagian beban (Weighted Round Robin).
- **TTL**: Jika node tidak mengirim registrasi ulang dalam X detik, Zeno akan menghapusnya otomatis (*Self-Healing*).

---

## 3. `docker.call`
Sang Panglima Utama. Mengirim RPC *(Remote Procedure Call)* otomatis bertipe JSON ke layanan mikro manapun terlepas dari bahasa perakitannya.

```zeno
docker.call: "php_legacy_payroll" {
    endpoint: "/api/hitung-gaji"
    method: "POST"
    port: 80
    payload: {
        id_karyawan: 432
        data_absensi: $bulan_maret
    }
    timeout: 30000        // Menunggu sampai 30 detik
    retry: 3              // [BARU] Coba lagi 3x jika gagal koneksi
    circuit_breaker: true // [BARU] Aktifkan fail-fast jika servis mati total
    as: $hitung
}

if: $hitung.success == true {
     log: "Sukses Menghitung. Total gaji:"
     log: $hitung.data.gaji_bersih
} else {
     log.error: "Peringatan! Service bermasalah. Pesan: " + $hitung.error
     if: $hitung.circuit_blocked {
         log.warn: "Circuit Breaker aktif! Menghentikan sementara permintaan."
     }
}
```

**Atribut Ketahanan (Resilience):**
*   `retry` (Integer): Jumlah percobaan ulang otomatis jika terjadi kegagalan jaringan atau *timeout*.
*   `circuit_breaker` (Boolean): Jika bernilai `true`, Zeno akan memantau kesehatan layanan tersebut secara otomatis. Jika terjadi 5 kali kegagalan berturut-turut, *circuit* akan terbuka selama 30 detik dan langsung memblokir semua permintaan berikutnya (`fail-fast`) untuk mencegah penumpukan beban pada server.

**Hasil Objek (Balasan):**
*   `success` (Boolean): `true` jika HTTP Code 200-299.
*   `code` (Integer): Kode HTTP asli.
*   `error` (String): Pesan kesalahan koneksi atau "Circuit Breaker: Open".
*   `circuit_blocked` (Boolean): Bernilai `true` jika permintaan dibatalkan oleh *Circuit Breaker*.
*   `data` (Object/Array): Hasil parsing JSON otomatis.
*   `raw` (String): Isi balasan utuh.

---

> Sekilas tentang ZenoEngine Ecosystem: Gunakan WebAssembly (WASM) *Plugins* jika algoritma berat tersebut harus dieksekusi **dalam milidetik hitungan memori yang sama**, dan gunakan *Container Bridge* jika stabilitas serta batas sumber daya bahasa (Pihak Ke-3) harus ditegakkan.
