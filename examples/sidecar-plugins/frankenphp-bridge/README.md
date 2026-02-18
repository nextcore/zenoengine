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

### `php.health`
Cek status bridge.

```
# ZenoLang
plugin.call: frankenphp-bridge
  slot: php.health
  as: $health

log: $health.status
```

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
