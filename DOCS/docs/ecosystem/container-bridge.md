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

### 2. `docker.call`
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
    timeout: 30000 // Menunggu sampai 30 detik
    as: $hitung
}

if: $hitung.success == true {
     log: "Sukses Menghitung. Total gaji:"
     log: $hitung.data.gaji_bersih
} else {
     log.error: "Peringatan! Service PHP tertidur. Pesan: " + $hitung.error
}
```

**Hasil Objek (Balasan):**
*   `success` (Boolean): Bernilai `true` murni apabila HTTP Code yang diterima berada di rentang 200 hingga 299.
*   `code` (Integer): Kode asil angka HTTP (Contoh: `200`, `404`, `500`).
*   `error` (String): Pesan putus koneksi murni (hanya ada jika DNS tidak mengenali *host container* atau koneksi melebih batas waktu / *timeout*).
*   `data` (Object/Array): Otomatis mem-*parsing* teks balasan Container pekerja jika ia mengirim data bertipe *Valid JSON*.
*   `raw` (String): Mengambil isi utuh balasan sekalipun itu bukan JSON (Cth: teks HTML/XML).

---

> Sekilas tentang ZenoEngine Ecosystem: Gunakan WebAssembly (WASM) *Plugins* jika algoritma berat tersebut harus dieksekusi **dalam milidetik hitungan memori yang sama**, dan gunakan *Container Bridge* jika stabilitas serta batas sumber daya bahasa (Pihak Ke-3) harus ditegakkan.
