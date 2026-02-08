# Contoh Aplikasi: Konversi Mata Uang 💸

Contoh aplikasi **ZenoWasm** sederhana untuk menghitung konversi kurs (USD <-> IDR) secara realtime. Aplikasi ini mendemonstrasikan:
1.  **Reaktivitas**: Input dua arah (Datastar).
2.  **Logic**: Perhitungan uang presisi tinggi (`money.calc`).
3.  **Network**: Fetch data API kurs (`http.fetch`).
4.  **Mobile Ready**: Bisa dibungkus jadi APK Android dengan **Capacitor**.

---

## 🚀 Cara Menjalankan (Tanpa Install Go)

Anda hanya perlu mendownload engine ZenoWasm yang sudah jadi.

### 1. Siapkan Folder
Pastikan Anda berada di dalam folder `currency-converter`.

### 2. Download Engine
Unduh 2 file inti ke dalam folder `public/`:

*   **zeno.wasm** (Engine): [Download Disini](https://github.com/nextcore/zenoengine/raw/main/ZenoWasm/public/zeno.wasm.gz)
    *   *Catatan: File ini terkompresi (.gz). Harap ekstrak menjadi `zeno.wasm`.*
*   **wasm_exec.js** (Loader): [Download Disini](https://github.com/nextcore/zenoengine/raw/main/ZenoWasm/public/wasm_exec.js)

**Cara Cepat via Terminal (Linux/Mac/Git Bash):**
```bash
cd public
curl -L -O https://github.com/nextcore/zenoengine/raw/main/ZenoWasm/public/zeno.wasm.gz
curl -L -O https://github.com/nextcore/zenoengine/raw/main/ZenoWasm/public/wasm_exec.js
gzip -d zeno.wasm.gz
```

### 3. Jalankan Server
Jalankan web server statis di folder `public/`.

```bash
# Menggunakan Python
python3 -m http.server -d public 8080

# Atau Node.js (http-server)
npx http-server public
```

Buka browser di `http://localhost:8080`.

---

## 📱 Cara Membuat APK Android (Opsional)

Jika Anda ingin mengubah website ini menjadi aplikasi Android native.

### Prasyarat
- Node.js & NPM
- Android Studio

### Langkah-Langkah
1.  **Instal Capacitor**:
    ```bash
    npm install
    ```

2.  **Setup Android**:
    ```bash
    npx cap add android
    ```

3.  **Jalankan di Emulator/HP**:
    ```bash
    npx cap run android
    ```

---

## 🛠️ Build Manual (Untuk Developer Go)
Jika Anda ingin mengkompilasi ulang engine dari source code (misal: habis edit `main.go`).

```bash
# Dari root repo
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o ZenoWasm/examples/currency-converter/public/zeno.wasm ZenoWasm/main.go
```
