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

## 📖 Syntax Dasar ZenoLang

Bagi Anda yang berasal dari PHP (Laravel) atau bahasa skrip lain, ZenoLang dirancang sangat familiar namun lebih ringkas.

### Variabel & Tipe Data
ZenoLang menggunakan tipe data dinamis.

```zeno
# Deklarasi Variabel
var: $nama { val: 'Budi' }
var: $umur { val: 25 }

# Shorthand (Direkomendasikan)
$kota: 'Jakarta'
$aktif: true

# Array & Map
$hobi: ['Coding', 'Gaming']
$profil: {
    role: 'admin'
    level: 99
}

# Akses Data
log: $hobi.0       # Output: Coding
log: $profil.role  # Output: admin
```

### Logika Percabangan (Control Flow)

```zeno
# IF - ELSE
if: $umur > 17 {
    then: {
        log: 'Dewasa'
    }
    else: {
        log: 'Belum Cukup Umur'
    }
}

# SWITCH
switch: $status {
    case: 'pending' { log: 'Menunggu...' }
    case: 'success' { log: 'Berhasil!' }
    default: { log: 'Status tidak dikenal' }
}

# NULL COALESCE (Default Value)
# Jika $input null, gunakan 'Default'
coalesce: $input { default: 'Default'; as: $hasil }
```

### Perulangan (Loops)

```zeno
# FOREACH (Iterasi Array/Map)
foreach: $items {
    as: $item
    do: {
        log: $item.name
    }
}

# FOR (C-Style)
for: '$i = 0; $i < 5; $i++' {
    do: {
        log: $i
    }
}

# WHILE
while: $count > 0 {
    do: {
        log: $count
        $count-- # Decrement
    }
}
```

### Error Handling

```zeno
try {
    do: {
        # Kode yang mungkin error
        http.fetch: 'https://api-error.com'
    }
    catch: {
        # Tangkap error di variabel $error
        as: $error
        js.call: 'alert' { args: ['Terjadi kesalahan: ' + $error] }
    }
}
```

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

#### Studi Kasus: Autentikasi dengan JWT
Berikut adalah pola standar untuk menangani Login di ZenoWasm:

1.  **Form Login** mengirim data ke Backend API.
2.  Backend mengembalikan **JWT Token**.
3.  ZenoWasm menyimpan token menggunakan `auth.login`.

```zeno
# Rute Aksi Login
router.get: '/do-login' {
    do: {
        # 1. Kirim Request ke API Backend
        http.fetch: 'https://api.myapp.com/login' {
            method: 'POST'
            body: {
                username: $username
                password: $password
            }
            then: {
                as: $resp

                # 2. Simpan Token & User (Jika sukses)
                if: $resp.token {
                    then: {
                        auth.login: {
                            token: $resp.token
                            user: $resp.user
                        }
                        # Redirect ke Dashboard
                        js.call: 'zenoNavigate' { args: ['/dashboard'] }
                    }
                    else: {
                        js.call: 'alert' { args: ['Login Gagal!'] }
                    }
                }
            }
        }
    }
}
```

Untuk request selanjutnya yang butuh Auth (Authenticated Request), sertakan token di header:

```zeno
auth.user: { as: $user } # Ambil user/token dari storage (jika tersimpan di object user)
storage.get: { key: 'zeno_auth_token'; as: $token } # Atau ambil raw token

http.fetch: 'https://api.myapp.com/protected-data' {
    headers: {
        'Authorization': 'Bearer ' + $token
    }
    # ... handle response
}
```

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
