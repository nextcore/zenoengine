<?php
/**
 * Worker mode script — PHP di-boot sekali, handle banyak request.
 * Cocok untuk Laravel/Symfony dengan FrankenPHP worker mode.
 *
 * Catatan: Worker mode memerlukan frankenphp_handle_request() yang
 * hanya tersedia saat berjalan dalam konteks FrankenPHP worker.
 */

// Boot aplikasi sekali
$booted = false;
$requestCount = 0;

// Simulasi boot aplikasi (misal: Laravel)
function bootApp(): array {
    return [
        'name'    => 'ZenoLang PHP Worker',
        'version' => PHP_VERSION,
        'booted'  => true,
    ];
}

$app = bootApp();
error_log("[Worker] App booted: " . json_encode($app));

// Cek apakah berjalan dalam FrankenPHP worker context
if (function_exists('frankenphp_handle_request')) {
    // Worker mode: loop handle request
    $handler = function() use (&$app, &$requestCount) {
        $requestCount++;
        $scope = json_decode($_SERVER['ZENO_SCOPE'] ?? '{}', true) ?? [];

        echo json_encode([
            'worker_mode' => true,
            'request_count' => $requestCount,
            'app' => $app,
            'scope' => $scope,
        ]);
    };

    $maxRequests = (int)($_SERVER['MAX_REQUESTS'] ?? 0);
    for ($n = 0; !$maxRequests || $n < $maxRequests; $n++) {
        if (!frankenphp_handle_request($handler)) {
            break;
        }
        gc_collect_cycles();
    }
} else {
    // CLI mode: jalankan sekali (untuk testing)
    $scope = json_decode(getenv('ZENO_SCOPE') ?: '{}', true) ?? [];
    echo json_encode([
        'worker_mode'   => false,
        'request_count' => 1,
        'app'           => $app,
        'scope'         => $scope,
        'note'          => 'Running in CLI mode (not FrankenPHP worker context)',
    ]);
    echo "\n";
}
