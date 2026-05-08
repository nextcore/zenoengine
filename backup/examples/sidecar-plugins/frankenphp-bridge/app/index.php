<?php
/**
 * Contoh PHP script sederhana untuk testing FrankenPHP bridge.
 * Jalankan via: php.run dengan parameter script: "app/index.php"
 */

// Baca Zeno scope dari environment variable (cara aman)
$zenoScope = json_decode($_SERVER['ZENO_SCOPE'] ?? getenv('ZENO_SCOPE') ?: '{}', true) ?? [];

$name = $zenoScope['name'] ?? 'World';
$version = PHP_VERSION;

echo "Hello, {$name}! PHP {$version} berjalan via FrankenPHP bridge.\n";
echo "Timestamp: " . date('Y-m-d H:i:s') . "\n";

// Contoh: akses scope data
if (!empty($zenoScope)) {
    echo "Zeno Scope: " . json_encode($zenoScope, JSON_PRETTY_PRINT) . "\n";
}
