# FrankenPHP Bridge — ZenoEngine Sidecar

Plugin sidecar Go yang menggunakan **FrankenPHP** sebagai PHP runtime, tanpa Caddy, tanpa port HTTP. Berkomunikasi dengan ZenoEngine via stdin/stdout JSON-RPC.

## Arsitektur

```
ZenoEngine ◄══ stdin/stdout JSON-RPC ══► frankenphp-bridge
                                              └── FrankenPHP (CGO + libphp)
                                                    └── PHP interpreter
```

## Prasyarat Build

### Linux
```bash
# Ubuntu/Debian
sudo apt install php-dev libphp-embed

# Atau via static build FrankenPHP
# Lihat: https://frankenphp.dev/docs/compile/
```

### macOS
```bash
brew install php
```

### Windows
Gunakan WSL2 dengan Ubuntu.

## Build

### Menggunakan Docker (Recommended)
Cara ini tidak memerlukan instalasi PHP di host machine.

```bash
cd examples/sidecar-plugins/frankenphp-bridge
chmod +x build-static.sh
./build-static.sh
```

### Build Manual (Linux/Mac)
Memerlukan `php-dev` atau `libphp` terinstall.

```bash
# Ubuntu/Debian
sudo apt install php-dev libphp-embed

# Build
cd examples/sidecar-plugins/frankenphp-bridge
CGO_ENABLED=1 go build -tags=nowatcher -o frankenphp-bridge .
```

## Slot yang Tersedia

### `php.eval`
Jalankan PHP code string langsung.

```
# ZenoLang
plugin.call: frankenphp-bridge
  slot: php.eval
  code: "echo 'Hello from PHP ' . phpversion();"
  as: $result

log: $result.output
```

### `php.run`
Jalankan PHP script file.

```
# ZenoLang
plugin.call: frankenphp-bridge
  slot: php.run
  script: "app/index.php"
  scope:
    name: "Max"
    user_id: 42
  as: $result

log: $result.output
```

### `php.extensions`
List loaded PHP extensions to verify the environment.

```
# ZenoLang
plugin.call: frankenphp-bridge
  slot: php.extensions
  as: $exts

log: "Loaded extensions: " + $exts.count
log: $exts.extensions
```

### `php.health`
Cek status bridge.

```
# ZenoLang
plugin.call: frankenphp-bridge
  slot: php.health
  as: $health

log: $health.status
```

## Ekstensi PHP
Image `static-builder` secara default menyertakan **sebagian besar ekstensi populer** (seperti PDO, MySQL, PGSQL, Redis, GD, Intl, BCMath, OpCache, dll). Anda bisa menggunakan slot `php.extensions` untuk melihat daftarnya.

Jika memerlukan ekstensi khusus yang tidak ada, Anda perlu memodifikasi `build-static.sh` untuk melakukan custom build library `libphp`.

## Laravel Support

Anda bisa menjalankan aplikasi Laravel melalui sidecar ini menggunakan mode **One-Shot (CLI)**. Ini mirip dengan menjalankan request via CGI atau Lambda function (booting per request).

### Contoh Script Handler
Lihat `app/laravel_handler.php` sebagai contoh entry point.

### Cara Pakai di ZenoLang

```zeno
# Panggil Laravel Handler
plugin.call: frankenphp-bridge
  slot: php.run
  script: "path/to/laravel_handler.php"
  scope:
    request:
      uri: "/api/users"
      method: "GET"
      query: { page: 1 }
  as: $response

# Parse Output JSON dari Handler
$jsonResponse = json_decode($response.output)
log: "Status: " + $jsonResponse.status
log: "Body: " + $jsonResponse.body
```

> **Catatan Performa**: Mode ini melakukan booting framework Laravel setiap kali request dijalankan.

## Laravel Octane / Worker Mode (High Performance)

Untuk performa maksimal (setara Laravel Octane), gunakan mode **Worker**. Dalam mode ini, aplikasi booting sekali saja dan tetap berjalan di memori.

### 1. Siapkan Worker Script
Gunakan `app/laravel_octane.php` sebagai entry point. Pastikan path ke Laravel project benar.

### 2. Konfigurasi Sidecar
Set environment variable `FRANKENPHP_CONFIG` di `manifest.yaml` atau saat menjalankan ZenoEngine.

```yaml
# manifest.yaml
sidecar:
  env:
    FRANKENPHP_CONFIG: "worker ./app/laravel_octane.php"
    MAX_REQUESTS: "500"
    
    # [Optional] Listen di TCP Port (bukan Unix Socket)
    # Berguna jika ingin debug atau akses dari luar container/host
    FRANKENPHP_PORT: "8000" 
    FRANKENPHP_HOST: "127.0.0.1" # Default: 127.0.0.1
```

### 3. Panggil via `php.request`
Gunakan slot khusus `php.request` yang akan mem-proxy request ke worker internal (Unix Socket atau TCP).

```zeno
# Panggil Worker (Ultra Fast)
plugin.call: frankenphp-bridge
  slot: php.request
  scope:
    request:
      uri: "/api/users"
      method: "GET"
  as: $response

log: "Status: " + $response.status
log: "Body: " + $response.body
```

Mode ini jauh lebih cepat (sub-millisecond latency untuk hello world) dibandingkan `php.run` (30ms+).

## Test Manual

```bash
# Build dulu
CGO_ENABLED=1 go build -o frankenphp-bridge .

# Test php.eval
echo '{"id":"1","slot_name":"php.eval","parameters":{"code":"echo phpversion();"}}' \
  | ./frankenphp-bridge

# Test php.run
echo '{"id":"2","slot_name":"php.run","parameters":{"script":"app/index.php","scope":{"name":"Max"}}}' \
  | ./frankenphp-bridge

# Test php.health
echo '{"id":"3","slot_name":"php.health","parameters":{}}' \
  | ./frankenphp-bridge
```

## Keunggulan vs Rust Bridge

| Fitur | Rust Bridge (lama) | FrankenPHP Bridge (ini) |
|---|---|---|
| Output capture | ❌ Bug (base64) | ✅ Native |
| Laravel support | ⚠️ Terbatas | ✅ Official |
| Worker mode | ❌ Tidak ada | ✅ Ada |
| Maintenance | ❌ Manual | ✅ Komunitas PHP |
| Build complexity | ❌ Rust + libphp FFI | ✅ Go + CGO |
