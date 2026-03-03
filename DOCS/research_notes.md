# ZenoEngine Repository Summary

## Overview
**ZenoEngine** is a stable, production-ready execution engine for the **ZenoLang** programming language. Built natively in Go, it focuses on high performance, code generation (metaprogramming), and web application features. The language (`.zl` files) uses a clean brace-based syntax.

## Key Architecture
1. **Engine (`pkg/engine`)**: Contains the lexer, parser, node definitions, executor, and an arena/pool for memory and execution management. 
2. **Slots (`internal/slots`)**: Acts as the standard library bridging Go capabilities into ZenoLang. It has a massive set of bindings including routing (`chi`), database (`dbmanager` supporting MySQL, Postgres, SQLite, SQLServer), template engines (`Blade`, `Inertia`), and utilities (`auth`, `math`, `mail`, `jobs`).
3. **App Initialization (`internal/app/registry.go`)**: Registers all built-in slot bindings during engine startup.
4. **Source Code Execution**: Execution typically starts from ZenoLang script files like `src/main.zl`, which configures routes, middleware, and logic.

## Technical Details
- **Language**: Go 1.26
- **Routing**: `go-chi/chi`
- **Database**: Native drivers (MySQL, Postgres, SQLite, SQLServer) via a central `DBManager`.
- **Other Features**: JWT, Cors, Rate Limiting, CRON-like Worker queues, LiveReload, WASM integrations.

## Summary
 ZenoEngine is more than just a programming language evaluator; it is a full-fledged web framework powered by Go, executing ZenoLang scripts. It brings "Laravel-style" web development patterns (Blade, Inertia, active record/DB integration) into a dynamically evaluated scripting language managed by a highly concurrent Go backend.

## Security Deep-Dive
### CSRF (Cross-Site Request Forgery) Protection
CSRF diimplementasikan secara sangat mirip dengan Laravel, namun menggunakan library `github.com/gorilla/csrf` di level middleware Go:
1. **Middleware Global (`internal/app/router.go`)**: 
   - Perlindungan CSRF diaktifkan secara global untuk semua *route* melalui fungsi `csrf.Protect(...)`.
   - Middleware ini otomatis mengecualikan (bypass) titik akhir `/api/*` dan `/health`, karena API biasanya tidak mempertahankan *state* berbasis *cookie* (*stateless JWT*).
   - Ia menarik `CSRF_TOKEN` dan daftar _Trusted Origins_ langsung dari konfigurasi `.env`.
2. **Injeksi ke Template Blade (`internal/slots/blade.go`)**:
   - Jika *developer* menggunakan `view.blade` untuk me-render HTML, variabel `csrf_field` dan `csrf_token` otomatis dilempar ke dalam *Scope* (konteks *template*).
   - Terdapat transpiler otomatis. Jika *developer* menulis `@csrf` di dalam *file* `.blade.zl`, sistem akan otomatis mengubahnya menjadi injeksi elemen HTML input tipe *hidden* (`<input type="hidden" name="gorilla.csrf.Token" ...>`).
3. **ZenoLang Slots API (`internal/slots/security.go`)**:
   - Terdapat *slot* khusus bernama `sec.csrf_token` yang dapat dipanggil dari skrip `.zl` untuk mengambil token CSRF mentah (biasanya digunakan untuk merakit *header* saat berinteraksi via AJAX/Inertia).

## Nginx-Like Capabilities (Edge Computing Features)
ZenoEngine dirancang agar bisa di-*deploy* langsung menghadap internet *(Public-Facing)* tanpa mutlak memerlukan perlindungan server tradisional tambahan seperti NGINX atau Apache. Ia memiliki banyak kapabilitas mandiri yang mengejutkan, antara lain:

1. **Virtual Hosting / Server Blocks (`pkg/host` & `HostDispatcher`)**
   - Mendukung konsep *Multi-Tenancy* ekstrem di mana sebuah *instance* / aplikasi ZenoEngine yang sama dapat melayani *domain-domain* yang berbeda secara independen.
   - Pengecekan *Host* (misal: `api.domain.com` ke *router* A dan `app.domain.com` ke *router* B) tidak dilakukan secara linear, melainkan dengan pemetaan kamus *O(1) Map Lookup* sehingga tidak melambat walau ada ribuan profil *host* sekaligus.

2. **Web Application Firewall (WAF) (`pkg/middleware/waf.go`)**
   - Seperti *ModSecurity* khusus NGINX, ZenoEngine memiliki *layer* WAF ultra-ringan. 
   - Ia memindai *User-Agent* dari peramban-peramban *hacker* *(bot scan)* seperti `sqlmap`, `nikto`, `nmap` dan memblokirnya otomatis sebelum masuk ke tahap *routing*.
   - Menganalisis *Query String* dan pangkalan data formulir (hingga ukuran maks 1MB) secara *real-time* untuk mencari sidik jari pola **XSS**, **SQL Injection**, dan **Path Traversal**.

3. **Bot Defense / JS Challenge (`pkg/middleware/bot_defense.go`)**
   - Mirip sekali fungsionalitasnya dengan *"Cloudflare I Am Under Attack Mode"*. Jika dihidupkan, klien yang baru pertama kali mengunjungi sebuah situs ZenoEngine akan disajikan halaman HTML *loading spinner* 503 yang berisi logika *JavaScript*. 
   - Ini memecahkan *(mitigate)* serangan DDoS L7, lalu mengeset *Cookie Authorization (zeno_bot_token)*. Bot / alat *scraper* HTTP primitif (seperti `curl`) yang tidak bisa menjalankan eksekusi JS akan tertahan *(dropped)* di fase *middleware* ini.

4. **IP Blocker & Rate Limiter**
   - Sangat mirip arahan *nginx rate limiting* (`limit_req`) dan `deny IP`, dapat difungsikan pada level konfigurasi `.env`.
