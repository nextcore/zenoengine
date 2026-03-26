# Container Bridge (Docker RPC)

ZenoEngine dirancang sebagai *"High-Concurrency Web Orchestrator"* yang berjalan **sangat cepat untuk I/O**, namun terkadang Anda membutuhkan logika berat (Machine Learning, konversi video, kalkulasi kompleks) yang paling cocok dikerjakan oleh bahasa lain seperti Python, PHP, atau Node.js.

**Container Bridge** adalah jembatan native yang memungkinkan ZenoLang berkomunikasi dengan *container mana pun* via HTTP/JSON — seolah memanggil fungsi lokal. Dengan ini, ZenoEngine mendapatkan akses ke seluruh ekosistem Python (TensorFlow, Pandas), PHP (Composer), Node.js (npm), dan lainnya tanpa meninggalkan satu binary pun.

---

## Arsitektur

```
[ Client ]
    │  HTTP
    ▼
[ ZenoEngine ]  ←── Circuit Breaker ──→  [ Service A: Python AI ]
    │                                        [ Service B: PHP Payroll ]
    └── Weighted Load Balancer ─────────→  [ Service C: Node.js PDF ]
                                             [ Service D: Go Encoder ]
```

ZenoEngine bertindak sebagai **API Gateway** dan **Service Orchestrator** sekaligus — mengelola health check, retry, circuit breaker, dan load balancing secara otomatis.

---

## Cara Kerja Container Bridge

Setiap container yang ingin dihubungkan ke ZenoEngine harus **menyediakan HTTP API minimal**:

| Endpoint | Method | Wajib? | Keterangan |
|:---|:---|:---|:---|
| `/health` | `HEAD` atau `GET` | **Ya** | Digunakan oleh health checker otomatis |
| `/your/endpoint` | `POST` | Ya | Endpoint bisnis yang dipanggil oleh `docker.call` |

ZenoEngine akan mengirim `Content-Type: application/json` dan mengharapkan respons JSON.

---

## Membuat "ZenoEngine-Ready" Container

### 🐍 Python (FastAPI)

```python
# main.py
from fastapi import FastAPI, Request
import uvicorn

app = FastAPI()

@app.head("/health")
@app.get("/health")
async def health():
    return {"status": "ok"}

@app.post("/api/analyze")
async def analyze(request: Request):
    body = await request.json()
    text = body.get("text", "")
    # Logika ML/AI Anda di sini
    result = {"sentiment": "positive", "score": 0.95, "input": text}
    return result

if __name__ == "__main__":
    uvicorn.run(app, host="0.0.0.0", port=8000)
```

```dockerfile
# Dockerfile
FROM python:3.11-slim
WORKDIR /app
COPY requirements.txt .
RUN pip install fastapi uvicorn
COPY . .
EXPOSE 8000
CMD ["python", "main.py"]
```

```yaml
# docker-compose.yml
services:
  ai_service:
    build: ./ai_service
    ports:
      - "8000:8000"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 30s
      timeout: 5s
      retries: 3
```

```zeno
# Panggil dari ZenoLang
docker.call: "ai_service" {
    port: 8000
    endpoint: "/api/analyze"
    payload: { text: $input_text }
    as: $result
}
log.info: "Sentimen: " + $result.sentiment
```

---

### 🐘 PHP (Laravel / Slim)

```php
<?php
// index.php — Slim Framework
use Psr\Http\Message\ResponseInterface as Response;
use Psr\Http\Message\ServerRequestInterface as Request;
use Slim\Factory\AppFactory;

require __DIR__ . '/vendor/autoload.php';

$app = AppFactory::create();

// Health Check — WAJIB untuk ZenoEngine
$app->get('/health', function (Request $req, Response $res) {
    $res->getBody()->write(json_encode(['status' => 'ok']));
    return $res->withHeader('Content-Type', 'application/json');
});
$app->map(['HEAD'], '/health', function (Request $req, Response $res) {
    return $res;
});

// Endpoint Bisnis
$app->post('/api/hitung-gaji', function (Request $req, Response $res) {
    $data = json_decode($req->getBody(), true);
    $gaji_bersih = $data['gaji_pokok'] - ($data['gaji_pokok'] * 0.05);
    $res->getBody()->write(json_encode([
        'success' => true,
        'gaji_bersih' => $gaji_bersih
    ]));
    return $res->withHeader('Content-Type', 'application/json');
});

$app->run();
```

```dockerfile
# Dockerfile
FROM php:8.2-cli
WORKDIR /app
RUN apt-get update && apt-get install -y curl
COPY . .
RUN curl -sS https://getcomposer.org/installer | php
RUN php composer.phar install
EXPOSE 8080
CMD ["php", "-S", "0.0.0.0:8080", "index.php"]
```

```zeno
# Panggil dari ZenoLang
docker.nodes: "payroll_service" {
    nodes: "10.0.1.10:8080, 10.0.1.11:8080"
    weight: 100
    check: "/health"
}

docker.call: "payroll_service" {
    endpoint: "/api/hitung-gaji"
    payload: { gaji_pokok: $employee.salary }
    retry: 3
    circuit_breaker: true
    as: $result
}
```

---

### 🟨 Node.js (Express)

```javascript
// server.js
const express = require('express');
const app = express();
app.use(express.json());

// Health Check — WAJIB
app.head('/health', (req, res) => res.sendStatus(200));
app.get('/health', (req, res) => res.json({ status: 'ok' }));

// Generate PDF
app.post('/api/generate-pdf', async (req, res) => {
    const { title, content } = req.body;
    // Logika PDF (misal: pakai puppeteer)
    const pdfBuffer = await generatePDF(title, content);
    res.json({
        success: true,
        url: '/files/' + savedFilename,
        size: pdfBuffer.length
    });
});

app.listen(3001, '0.0.0.0', () => {
    console.log('PDF Service running on :3001');
});
```

```dockerfile
FROM node:20-slim
WORKDIR /app
COPY package*.json ./
RUN npm ci --only=production
COPY . .
EXPOSE 3001
CMD ["node", "server.js"]
```

```zeno
docker.call: "pdf_service" {
    port: 3001
    endpoint: "/api/generate-pdf"
    payload: { title: $doc.title content: $doc.body }
    timeout: 30000
    as: $pdf
}
http.json: { url: $pdf.url }
```

---

### 🐹 Go Service

```go
// main.go
package main

import (
    "encoding/json"
    "net/http"
)

func healthHandler(w http.ResponseWriter, r *http.Request) {
    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

func encodeHandler(w http.ResponseWriter, r *http.Request) {
    var body map[string]interface{}
    json.NewDecoder(r.Body).Decode(&body)
    // Logika encoding video, dsb.
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]interface{}{
        "success": true,
        "output":  "encoded_" + body["filename"].(string),
    })
}

func main() {
    http.HandleFunc("/health", healthHandler)
    http.HandleFunc("/api/encode", encodeHandler)
    http.ListenAndServe(":9090", nil)
}
```

---

## Docker Compose: Arsitektur Lengkap

Berikut adalah contoh `docker-compose.yml` untuk deployment ZenoEngine bersama berbagai service:

```yaml
# docker-compose.yml
version: '3.9'

services:
  # ─── ZenoEngine (API Gateway & Orchestrator) ───
  zenoengine:
    image: your-org/zenoengine:latest
    ports:
      - "3000:3000"
    environment:
      APP_PORT: ":3000"
      APP_ENV: production
      DB_DRIVER: postgres
      DB_HOST: db
      DB_USER: zeno
      DB_PASS: secret
      DB_NAME: zenodb
    depends_on:
      - db
      - ai_service
      - pdf_service
    restart: unless-stopped

  # ─── Python AI Service ───
  ai_service:
    build: ./services/ai
    expose:
      - "8000"
    restart: unless-stopped
    deploy:
      replicas: 2   # 2 instance → Zeno otomatis load balance!
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8000/health"]
      interval: 10s

  # ─── Node.js PDF Service ───
  pdf_service:
    build: ./services/pdf
    expose:
      - "3001"
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:3001/health"]
      interval: 10s

  # ─── PHP Legacy/Payroll Service ───
  payroll_service:
    build: ./services/payroll
    expose:
      - "8080"
    restart: unless-stopped

  # ─── Database ───
  db:
    image: postgres:15
    environment:
      POSTGRES_USER: zeno
      POSTGRES_PASSWORD: secret
      POSTGRES_DB: zenodb
    volumes:
      - pgdata:/var/lib/postgresql/data

volumes:
  pgdata:
```

---

## API Reference Lengkap

### `docker.nodes` — Registrasi Service Pool

```zeno
docker.nodes: "ai_service" {
    nodes: "ai_service:8000"  # Nama container Docker langsung
    weight: 100               # Bobot load balancing (default: 50)
    check: "/health"          # Endpoint health check (default: /health)
    ttl: 0                    # Time-to-live dalam detik (0 = permanen)
}
```

| Atribut | Tipe | Wajib | Default | Keterangan |
|:---|:---|:---|:---|:---|
| `nodes` | string | Ya | — | Satu atau lebih host:port, pisah koma |
| `weight` | int | Tidak | 50 | Bobot Weighted Round Robin |
| `check` | string | Tidak | `/health` | Path endpoint health check |
| `ttl` | int | Tidak | 0 | Detik sebelum node dianggap kedaluwarsa |

---

### `docker.call` — Memanggil Service

```zeno
docker.call: "service_name" {
    endpoint: "/api/path"    # Path endpoint di container
    method: "POST"           # GET, POST, PUT, DELETE (default: POST)
    port: 8000               # Port (akan diabaikan jika pakai docker.nodes)
    payload: {               # Data yang dikirim (otomatis jadi JSON)
        key: $value
    }
    headers: {               # Header tambahan (opsional)
        Authorization: "Bearer " + $token
    }
    timeout: 10000           # Millisecond (default: 10000 = 10 detik)
    retry: 3                 # Coba ulang jika gagal (default: 0)
    circuit_breaker: true    # Aktifkan circuit breaker (default: false)
    as: $result              # Simpan hasil ke variabel
}
```

**Respons (`$result`):**

| Field | Tipe | Keterangan |
|:---|:---|:---|
| `success` | bool | `true` jika HTTP 200–299 |
| `code` | int | HTTP status code |
| `data` | any | Hasil parsing JSON dari body |
| `raw` | string | Body respons mentah |
| `error` | string | Pesan error (jika gagal) |
| `circuit_blocked` | bool | `true` jika diblokir circuit breaker |

---

### Auto-Join API — Registrasi Dinamis

Service eksternal bisa mendaftarkan dirinya sendiri **tanpa konfigurasi ZenoEngine** sama sekali:

```bash
# Daftarkan node baru (dipanggil oleh service saat startup)
curl -X POST http://zeno-host:3000/api/zeno/register \
  -H "Content-Type: application/json" \
  -d '{
    "service": "ai_service",
    "host": "10.0.5.20",
    "port": 8000,
    "weight": 100,
    "ttl": 30
  }'
```

```json
// Respons sukses
{
  "status": "registered",
  "node": "10.0.5.20:8000",
  "expiry": "2026-03-26T12:30:00+07:00"
}
```

**Pola Heartbeat** — Service mengirim registrasi ulang sebelum TTL habis:

```python
# heartbeat.py — jalankan sebagai background thread
import requests, time, threading

ZENO_HOST = "http://zenoengine:3000"
TTL = 30  # harus lebih kecil dari TTL yang didaftarkan

def heartbeat():
    while True:
        try:
            requests.post(f"{ZENO_HOST}/api/zeno/register", json={
                "service": "ai_service",
                "host": "ai_service",
                "port": 8000,
                "weight": 100,
                "ttl": TTL
            })
        except:
            pass
        time.sleep(TTL - 5)  # kirim 5 detik sebelum expired

threading.Thread(target=heartbeat, daemon=True).start()
```

---

## Circuit Breaker

Aktifkan `circuit_breaker: true` pada `docker.call` untuk melindungi sistem dari efek domino.

| State | Kondisi | Perilaku |
|:---|:---|:---|
| **Closed** (Normal) | — | Semua request diteruskan |
| **Open** (Memblokir) | 5 kegagalan berturut-turut | Semua request langsung gagal selama 30 detik |
| **Half-Open** (Percobaan) | Setelah 30 detik | 1 request percobaan dikirim |

```zeno
docker.call: "payment_service" {
    endpoint: "/charge"
    payload: { amount: $total }
    circuit_breaker: true
    as: $result
}

if: $result.circuit_blocked == true {
    then: {
        http.json: { error: "Payment service sedang bermasalah, coba lagi nanti" }
        return
    }
}
```

---

## Tips & Best Practices

> [!TIP]
> Gunakan nama **service Docker Compose** sebagai hostname — misal `ai_service:8000` — sehingga Docker menangani DNS resolution otomatis.

> [!IMPORTANT]
> Selalu implementasikan `/health` endpoint di setiap container Anda. ZenoEngine melakukan health check setiap 10 detik dan akan **mengeluarkan node yang tidak sehat** dari pool secara otomatis.

> [!NOTE]
> Untuk service yang di-scale (`replicas: 2`), daftarkan semua IP/hostname instance ke `docker.nodes`. ZenoEngine akan mendistribusikan beban secara adil berdasarkan `weight` masing-masing.

> [!WARNING]
> Circuit breaker bersifat **in-memory** dan akan reset saat ZenoEngine di-restart. Untuk state yang persisten, pertimbangkan menggunakan health check yang lebih agresif dengan TTL pendek.
