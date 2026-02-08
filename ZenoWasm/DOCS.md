# Panduan Developer ZenoWasm 🚀

**ZenoWasm** adalah framework **Single Page Application (SPA)** unik yang memungkinkan Anda menjalankan logika backend **ZenoLang** dan templating **Blade** langsung di dalam browser menggunakan **WebAssembly (WASM)**.

Ini berarti Anda bisa membangun aplikasi web yang dinamis, reaktif, dan offline-ready tanpa perlu berpindah konteks ke JavaScript framework yang kompleks (seperti React/Vue/Angular), sambil tetap menggunakan sintaks yang familiar bagi developer backend (Go/PHP/Laravel).

---

## 🏁 Memulai (Getting Started)

### Prasyarat
- **Go 1.21+** terinstal di sistem Anda.
- Browser modern (Chrome/Firefox/Edge) yang mendukung WebAssembly.

### 1. Instalasi & Build
ZenoWasm adalah project *standalone*. Anda perlu mengkompilasi kode sumber ZenoWasm menjadi file binary `.wasm`.

```bash
# Masuk ke direktori ZenoWasm
cd ZenoWasm

# Jalankan build script (Linux/Mac)
./build.sh

# Atau build manual
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o public/zeno.wasm main.go
```

Hasil build akan ada di folder `public/zeno.wasm` (sekitar 3.5MB setelah gzip).

### 2. Struktur Proyek
Untuk menjalankan aplikasi, Anda hanya butuh 3 file statis:

```text
public/
├── index.html      # Titik masuk aplikasi (HTML + JS Loader)
├── zeno.wasm       # Engine Zeno (Binary)
└── wasm_exec.js    # Go WASM Loader (bawaan Go)
```

### 3. Menjalankan Server Dev
Anda bisa menggunakan server statis apa saja (Python, Nginx, Caddy).

```bash
# Contoh dengan Python
python3 -m http.server -d public 8080
```
Buka browser di `http://localhost:8080`.

---

## 🎨 Templating (Zeno Blade)

ZenoWasm menggunakan engine **Blade** yang sama persis dengan ZenoEngine di backend. Bedanya, template tidak dibaca dari disk, melainkan didaftarkan ke **Virtual Filesystem** di browser.

### Mendaftarkan Template
Gunakan fungsi JS `zenoRegisterTemplate` di `index.html`:

```javascript
zenoRegisterTemplate("home", `
    <h1>Halo, {{ $nama }}!</h1>
    @if($isAdmin)
        <button>Admin Panel</button>
    @endif
`);
```

### Layouts & Sections
Mendukung pewarisan template penuh.

**Template `layout`:**
```html
<nav>My App</nav>
<main>
    @yield('content')
</main>
```

**Template `dashboard`:**
```html
@extends('layout')
@section('content')
    <h2>Dashboard</h2>
    <p>Selamat datang!</p>
@endsection
```

---

## 🧭 Routing (SPA)

ZenoWasm memiliki router client-side bawaan yang terintegrasi dengan **History API** browser. Ini memungkinkan navigasi antar halaman tanpa reload.

### Inisialisasi Router
Definisikan rute menggunakan sintaks ZenoLang:

```javascript
zenoInitRouter(`
    # Rute Sederhana
    router.get: '/' { view: 'home' }
    router.get: '/about' { view: 'about' }

    # Rute dengan Logika
    router.get: '/profile' {
        do: {
            auth.user: { as: $u }
            view: 'profile'
        }
    }
`);
```

### Middleware (Route Guard)
Melindungi halaman tertentu (misal: hanya user login).

```javascript
router.group: {
    middleware: 'auth'
    do: {
        router.get: '/dashboard' { view: 'dashboard' }
        router.get: '/settings' { view: 'settings' }
    }
}
```
*Catatan: Middleware `auth` akan mengecek apakah ada token login yang tersimpan.*

---

## ⚡ Reaktivitas (Datastar)

ZenoWasm terintegrasi dengan library **Datastar** (tertanam di dalam) untuk interaksi UI reaktif tanpa menulis JavaScript.

### State & Binding
Gunakan atribut `data-*` untuk menghubungkan variabel UI.

```html
<!-- Inisialisasi State -->
<div data-signals="{ count: 0 }">

    <!-- Tampilkan Teks (Reaktif) -->
    <span data-text="$count">0</span>

    <!-- Event Click -->
    <button data-on-click="$count++">Tambah</button>

    <!-- Input Binding (Two-Way) -->
    <input type="text" data-bind-count>
</div>
```

---

## 📚 API Reference (Standard Library)

ZenoWasm menyediakan slot khusus (Native Functions) yang dioptimalkan untuk browser.

### 🌐 Network (HTTP Fetch)
Mengambil data dari API eksternal secara asinkron.

```zeno
http.fetch: 'https://api.example.com/data' {
    method: 'GET'
    then: {
        as: $response
        # Update UI dengan data baru (harus trigger render ulang atau update signal)
        js.call: 'window.updateData' { args: [$response] }
    }
    catch: {
        as: $error
        js.log: $error
    }
}
```

### 💰 Finance (Money & Math)
Perhitungan uang presisi tinggi (Decimal) untuk menghindari error floating-point.

```zeno
# Hitung Diskon
money.calc: ($harga * $qty) - $diskon { as: $total }

# Format Rupiah
money.format: $total {
    symbol: 'Rp '
    precision: 2
    as: $str_total
}
```

### 🔐 Auth & Storage
Manajemen sesi pengguna (disimpan di LocalStorage).

*   `auth.login`: Simpan token/user. `{ token: '...', user: {...} }`
*   `auth.logout`: Hapus sesi.
*   `auth.user`: Ambil data user login `{ as: $user }`.
*   `auth.check`: Cek status login `{ as: $isLoggedIn }`.
*   `storage.set`, `storage.get`, `storage.remove`: Akses raw LocalStorage.

### 🔌 JS Interop
Berinteraksi langsung dengan JavaScript browser.

```zeno
# Panggil Fungsi JS Global
js.call: 'alert' { args: ['Halo!'] }

# Set Properti DOM/Window
js.set: 'document.title' { val: 'Judul Baru' }

# Ambil Nilai JS
js.get: 'window.innerWidth' { as: $width }

# Log ke Console
js.log: 'Pesan debug'
```

---

## 🚀 Tips Performa

1.  **Gunakan Layout**: Jangan render ulang seluruh halaman jika hanya konten yang berubah. Gunakan `@extends`.
2.  **Datastar untuk Interaksi**: Jangan gunakan `zenoNavigate` untuk interaksi kecil (seperti counter atau toggle). Gunakan Datastar `data-signals` karena jauh lebih cepat (hanya update DOM node terkait).
3.  **Lazy Load Data**: Render kerangka halaman (skeleton) dulu, lalu panggil `http.fetch` di dalam rute untuk mengisi data.

Selamat berkarya dengan **ZenoWasm**! 🎉
