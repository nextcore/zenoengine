# Mini-Zeno

Ini adalah versi minimalis (stripped-down) dari ZenoEngine. Dibuat sebagai *sub-project* untuk menunjukkan cara kerja inti *ZenoLang* tanpa kompleksitas engine utama.

## Apa yang ada di sini?
- **Lexer & Parser Sederhana:** Membaca file `.zl` dan mengubahnya menjadi *Abstract Syntax Tree* (AST).
- **Executor & Registry:** Memetakan *slot* ZenoLang (seperti `log:` atau `http.get:`) ke fungsi Go.
- **HTTP Router Sederhana:** Menjalankan server web jika ada rute yang didaftarkan.

## Cara Menjalankan

1. Pastikan Anda berada di direktori akar repositori atau di dalam folder `mini-zeno`.
2. Jalankan engine dengan file contoh:
   ```bash
   go run main.go example.zl
   ```

## Contoh (`example.zl`)

```zeno
log: "Starting Minimalist ZenoEngine..."

http.get: "/hello" {
    log: "Someone visited /hello"
    body: "Hello from Minimalist ZenoEngine!"
}

http.get: "/about" {
    body: "This is a minimalist version of the complex ZenoEngine."
}
```

## Memahami Konsep

Versi minimalis ini membantu memahami konsep inti ZenoLang:
1. **Pohon (Tree):** Semua sintaks adalah pohon slot. `http.get` memiliki *child* `log` dan `body`.
2. **Registry:** Engine utama mendaftarkan berbagai fungsi. Saat parser melihat `log`, ia memanggil fungsi Go yang telah didaftarkan dengan nama `"log"`.
3. **Eksekusi:** File dibaca satu kali pada saat inisialisasi, meregistrasi rute. Ketika pengguna mengakses `/hello`, *children* dari blok `/hello` dieksekusi.
