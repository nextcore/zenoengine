# Contoh Aplikasi: CRUD Dashboard 🛍️

Aplikasi Dashboard Admin lengkap menggunakan **ZenoWasm**.

## Fitur
1.  **Autentikasi**: Login menggunakan `https://dummyjson.com/auth/login`.
2.  **Proteksi**: Middleware `auth` melindungi halaman admin.
3.  **CRUD**: Create, Read, Update, Delete data produk.
4.  **Layout**: Sidebar + Navbar Layout pattern dengan `@extends`.

## Cara Menjalankan

### 1. Download Engine
Pastikan Anda memiliki `zeno.wasm` (diekstrak dari .gz) dan `wasm_exec.js` di folder `public/`.

Jika belum ada, copy dari contoh sebelumnya atau download:
```bash
cd public
curl -L -O https://github.com/nextcore/zenoengine/raw/main/ZenoWasm/public/zeno.wasm.gz
curl -L -O https://github.com/nextcore/zenoengine/raw/main/ZenoWasm/public/wasm_exec.js
gzip -d zeno.wasm.gz
```

### 2. Jalankan Server
```bash
python3 -m http.server -d public 8081
```
Buka `http://localhost:8081`.

### 3. Login Demo
Gunakan kredensial default DummyJSON:
- **Username**: `emilys`
- **Password**: `emilyspass`
