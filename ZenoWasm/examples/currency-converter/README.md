# Zeno Money (Mobile App Example)

Contoh aplikasi Konversi Mata Uang menggunakan **ZenoWasm** yang dibungkus menjadi APK Android menggunakan **Capacitor**.

## Prasyarat
- Node.js & NPM
- Go 1.21+
- Android Studio (untuk build APK)

## Cara Build (Web)

1.  Jalankan build script untuk mengkompilasi engine WASM:
    ```bash
    ./build_example.sh
    ```
2.  Buka `public/index.html` di browser (via local server) untuk testing.

## Cara Build (Android APK)

1.  Instal dependensi Capacitor:
    ```bash
    npm install
    ```

2.  Inisialisasi Android project:
    ```bash
    npx cap add android
    ```

3.  Sinkronisasi file web ke native:
    ```bash
    npx cap sync
    ```

4.  Buka di Android Studio & Run:
    ```bash
    npx cap open android
    ```
    Atau build langsung:
    ```bash
    npx cap run android
    ```

## Struktur
- `public/`: Berisi web assets (`index.html`, `zeno.wasm`, `wasm_exec.js`).
- `capacitor.config.json`: Konfigurasi ID aplikasi dan nama.
