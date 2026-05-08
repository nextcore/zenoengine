<?php
/**
 * Laravel Handler Script (One-Shot Mode)
 *
 * Script ini digunakan untuk menjalankan request Laravel per-panggilan (mirip CGI/Lambda).
 * Booting Laravel terjadi setiap kali script dijalankan.
 *
 * Cara pakai di ZenoLang:
 * plugin.call: frankenphp-bridge
 *   slot: php.run
 *   script: "examples/sidecar-plugins/frankenphp-bridge/app/laravel_handler.php"
 *   scope:
 *     request:
 *       uri: "/api/users"
 *       method: "GET"
 *       query: { "page": 1 }
 *       body: { "name": "Zeno" }
 *   as: $result
 */

use Illuminate\Contracts\Http\Kernel;
use Illuminate\Http\Request;

// 1. Lokasi Autoload (Sesuaikan Path!)
// Asumsi: script ini ada di examples/sidecar-plugins/frankenphp-bridge/app/
// Dan project laravel ada di parent folder atau diatur path-nya.
// Ubah path ini ke lokasi root project Laravel Anda.
$laravelPath = getenv('LARAVEL_PATH') ?: __DIR__ . '/../../../../laravel-app';

if (!file_exists($laravelPath . '/vendor/autoload.php')) {
    echo json_encode(['error' => "Laravel not found at $laravelPath. Please edit laravel_handler.php or set LARAVEL_PATH env."]);
    exit(1);
}

require $laravelPath . '/vendor/autoload.php';

// 2. Boot Laravel
$app = require $laravelPath . '/bootstrap/app.php';
$kernel = $app->make(Kernel::class);

// 3. Construct Request dari Zeno Scope
// Zeno mengirim scope via env var ZENO_SCOPE (JSON string)
$zenoScope = json_decode($_SERVER['ZENO_SCOPE'] ?? '{}', true) ?: [];
$requestData = $zenoScope['request'] ?? [];

// Mock Superglobals (agar Laravel membacanya)
$uri = $requestData['uri'] ?? '/';
$method = strtoupper($requestData['method'] ?? 'GET');
$query = $requestData['query'] ?? [];
$body = $requestData['body'] ?? [];

// Buat Request (Symfony Request)
$request = Request::create(
    $uri,
    $method,
    $query, // GET parameters
    [], // Cookies
    [], // Files
    $_SERVER, // Server vars
    json_encode($body) // Content (Raw Body) if JSON
);

// Inject Header Content-Type jika ada body
if (!empty($body)) {
    $request->headers->set('Content-Type', 'application/json');
}

// Tambahkan Zeno Scope ke attributes request
$request->attributes->add(['zeno_scope' => $zenoScope]);

// 4. Handle Request
$response = $kernel->handle($request);

// 5. Output Response
// Kita tangkap content response dan kirim ke stdout utk Zeno
$content = $response->getContent();
$status = $response->getStatusCode();
$headers = $response->headers->all();

// Format output JSON untuk kemudahan parsing di Zeno
// (Opsional: Zeno menangkap raw output, tapi JSON lebih rapi)
echo json_encode([
    'status' => $status,
    'headers' => $headers,
    'body' => $content
]);

// 6. Terminate
$kernel->terminate($request, $response);
