# ⚡ Detail Implementasi Laravel Connection Pool via Zeno Proxy

Panduan ini menjelaskan secara teknis bagaimana Laravel dapat menggunakan **Go Connection Pool** milik ZenoEngine alih-alih membuka koneksi database sendiri. Ini sangat berguna untuk skalabilitas tinggi karena Go menangani ribuan koneksi dengan sangat efisien.

---

## 1. Konsep Arsitektur
1.  **Laravel** memanggil Query Builder / Eloquent.
2.  **Custom Driver** di Laravel mencegat query tersebut.
3.  Query dikirim ke **Zig Bridge** (Sidecar) via JSON-RPC internal.
4.  **Zig Bridge** mengirim pesan `host_call` dengan fungsi `db_query` ke **ZenoEngine (Go)**.
5.  **ZenoEngine** mengeksekusi query menggunakan connection pool-nya dan mengembalikan hasilnya ke Laravel.

---

## 2. Implementasi Custom Driver di Laravel

Anda perlu membuat dua file utama di project Laravel Anda.

### A. Buat Connection Class (`App\Database\ZenoProxyConnection.php`)
Class ini bertugas meneruskan statement SQL ke bridge.

```php
<?php

namespace App\Database;

use Illuminate\Database\Connection;
use Illuminate\Database\Query\Processors\Processor;
use Illuminate\Database\Query\Grammars\Grammar;

class ZenoProxyConnection extends Connection
{
    protected function run($query, $bindings, \Closure $callback)
    {
        // Panggil fungsi global yang disediakan oleh Zig Bridge
        // Biasanya bridge akan menyuntikkan fungsi rpc() ke runtime PHP
        return $this->proxyToZeno($query, $bindings);
    }

    protected function proxyToZeno($query, $bindings)
    {
        // Contoh pemanggilan RPC ke ZenoEngine via Zig Sidecar
        // Pesan JSON dikirim ke StdOut yang ditangkap oleh Go
        $response = \Zeno::call('db_query', [
            'connection' => $this->getName(),
            'sql' => $query,
            'params' => $bindings
        ]);

        if (!$response['success']) {
            throw new \Exception("Zeno DB Proxy Error: " . $response['error']);
        }

        return $response['data']['rows'];
    }

    // Implementasi method abstract lainnya...
    public function getDefaultQueryGrammar() { return new Grammar; }
    public function getDefaultPostProcessor() { return new Processor; }
}
```

### B. Registrasi Driver di `AppServiceProvider.php`

```php
<?php

namespace App\Providers;

use App\Database\ZenoProxyConnection;
use Illuminate\Support\ServiceProvider;
use Illuminate\Support\Facades\DB;

class AppServiceProvider extends ServiceProvider
{
    public function boot()
    {
        DB::extend('zeno_proxy', function ($config, $name) {
            return new ZenoProxyConnection(null, $config['database'], $config['prefix'], $config);
        });
    }
}
```

---

## 3. Konfigurasi `config/database.php`

Tambahkan koneksi baru yang menggunakan driver `zeno_proxy`.

```php
'connections' => [
    'zeno' => [
        'driver' => 'zeno_proxy',
        'database' => env('DB_DATABASE', 'main'),
    ],
],
```

Ubah default connection di `.env`:
```env
DB_CONNECTION=zeno
```

---

## 4. Keuntungan Menggunakan Cara Ini

| Fitur | Laravel Standar | Laravel + Zeno Proxy |
| :--- | :--- | :--- |
| **Koneksi** | 1 Request = 1 Koneksi Baru | 1000 Request = Reusable Go Pool |
| **Memory** | Tinggi (overhead PDO) | Sangat Rendah (Lightweight JSON) |
| **Handshake** | Tiap request ada TCP handshake | Handshake hanya dilakukan Go sekali saja |
| **Keamanan** | Credential ada di file `.env` PHP | Credential hanya diketahui oleh Go (lebih aman) |

---

## 5. Sinkronisasi Scope (Deep Integration)

ZenoEngine juga mengirimkan `_zeno_scope` ke dalam script PHP. Anda dapat mengakses variabel ZenoLang langsung dari Laravel:

```php
// Di dalam Controller Laravel
$zenoUser = request()->header('X-Zeno-User-ID');
// Atau jika bridge menyuntikkan variabel:
$cartTotal = $_ZENO['cart_total'];

log_info("Laravel memproses data dari Zeno: " . $cartTotal);
```

---
*Teknologi ini memungkinkan ZenoEngine menjadi orkestrator yang sangat powerfull bagi framework besar seperti Laravel.*
