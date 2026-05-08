<?php
/**
 * Laravel Octane/Worker Entry Point for FrankenPHP Sidecar.
 *
 * Script ini digunakan bersama dengan mode WORKER FrankenPHP.
 * Set env var: FRANKENPHP_CONFIG="worker app/laravel_octane.php"
 */

use Illuminate\Contracts\Http\Kernel;
use Illuminate\Http\Request;

// 1. Lokasi Autoload (Sesuaikan Path!)
$laravelPath = getenv('LARAVEL_PATH') ?: __DIR__ . '/../../../../laravel-app';

if (!file_exists($laravelPath . '/vendor/autoload.php')) {
    fwrite(STDERR, "Laravel not found at $laravelPath.\n");
    exit(1);
}

require $laravelPath . '/vendor/autoload.php';

// 2. Boot Laravel Application (Sekali saja!)
$app = require $laravelPath . '/bootstrap/app.php';
$kernel = $app->make(Kernel::class);

fwrite(STDERR, "✅ Laravel booted in WORKER mode via FrankenPHP bridge.\n");

// 3. Definition Handler
$handler = function () use ($kernel, $app) {
    // Inject Scope Zeno dari Header X-Zeno-Scope
    // (Dikirim oleh main.go proxyRequest)
    $zenoScope = [];
    if (isset($_SERVER['HTTP_X_ZENO_SCOPE'])) {
        $zenoScope = json_decode($_SERVER['HTTP_X_ZENO_SCOPE'], true) ?? [];
    }
    
    // FrankenPHP sudah menyiapkan superglobals ($_GET, $_SERVER, dll)
    // Kita capture request dari globals tersebut
    $request = Request::capture();
    
    // Inject scope ke attributes
    $request->attributes->add(['zeno_scope' => $zenoScope]);

    // Handle Request
    $response = $kernel->handle($request);

    // Send Response
    $response->send();

    // Terminate (cleanup logic laravel)
    $kernel->terminate($request, $response);
};

// 4. Loop Requests
// frankenphp_handle_request adalah fungsi built-in dari FrankenPHP Runtime
$maxRequests = (int)($_SERVER['MAX_REQUESTS'] ?? 500);

for ($nbRequests = 0; !$maxRequests || $nbRequests < $maxRequests; ++$nbRequests) {
    // Tunggu request masuk
    $keepRunning = frankenphp_handle_request($handler);

    // Garbage collection
    gc_collect_cycles();

    if (!$keepRunning) break;
}
