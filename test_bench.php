<?php
$start = microtime(true);
$sum = 0;

// Equivalen 10.000 iterasi
for ($i = 0; $i < 10000; $i++) {
    $sum++;
}

$elapsed = microtime(true) - $start;
echo "PHP Selesai. Hasil: " . $sum . "\n";
echo "Waktu Dieksekusi: " . number_format($elapsed * 1000, 4) . " ms\n";
