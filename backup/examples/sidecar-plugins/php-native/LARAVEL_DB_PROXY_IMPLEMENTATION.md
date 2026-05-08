# ⚡ Detail Implementasi Laravel Connection Pool via Zeno Proxy

Panduan ini menjelaskan secara teknis konsep bagaimana Laravel dapat menggunakan **Go Connection Pool** milik ZenoEngine.

---

## 1. Konsep Arsitektur
1.  **Laravel** memanggil Query Builder / Eloquent.
2.  **Custom Driver** di Laravel mencegat query tersebut.
3.  Query dikirim ke **Native Bridge** (Rust) via JSON-RPC.
4.  **Native Bridge** (via slot `php.db_proxy`) meneruskan pesan `host_call` dengan fungsi `db_query` ke **ZenoEngine (Go)**.
5.  **ZenoEngine** mengeksekusi query menggunakan connection pool-nya dan mengembalikan hasilnya.

---

## 2. Implementasi Custom Driver di Laravel (Tugas Anda)

Saat ini, ZenoEngine menyediakan **Protokol** komunikasi (`php.db_proxy`). Anda sebagai developer perlu mengimplementasikan driver database di sisi Laravel (Client) untuk menggunakan protokol ini.

### Contoh Mock Driver
Berikut adalah contoh bagaimana implementasi driver `ZenoProxyConnection` akan terlihat di sisi PHP:

```php
<?php

namespace App\Database;

use Illuminate\Database\Connection;

class ZenoProxyConnection extends Connection
{
    // ... setup standard connection ...

    protected function run($query, $bindings, \Closure $callback)
    {
        // 1. Serialize Query
        $payload = [
            'sql' => $query,
            'params' => $bindings
        ];

        // 2. Kirim ke Bridge (Mekanisme ini butuh library client JSON-RPC PHP)
        // Contoh konseptual:
        // $response = ZenoRPC::call('php.db_proxy', $payload);

        // 3. Return Result
        // return $response['data'];

        throw new \Exception("Implementasi Driver PHP belum tersedia. Silakan buat wrapper JSON-RPC.");
    }
}
```

> **Catatan:** `php.db_proxy` di bridge Rust saat ini sudah siap menerima request dan melakukan `host_call` ke ZenoEngine. Namun, mekanisme pengiriman pesan dari PHP (user-land) ke Rust (parent process) perlu dibangun, misalnya dengan menulis ke *Named Pipe* khusus atau menggunakan fungsi internal `zeno_rpc_call()` jika nanti diimplementasikan di level ekstensi C.

---

## 3. Roadmap Masa Depan

Kami berencana merilis paket `zeno-laravel-driver` resmi yang akan menangani semua kerumitan ini secara otomatis. Untuk saat ini, fitur DB Pooling tersedia sebagai **API Low-Level** bagi power user yang ingin membangun integrasi kustom.
