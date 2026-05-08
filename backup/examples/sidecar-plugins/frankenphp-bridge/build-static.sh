#!/bin/bash
set -e

# Versi PHP yang akan digunakan (sesuaikan dengan release FrankenPHP)
PHP_VERSION="8.4"
OS="linux"
ARCH="x86_64"

# Direktori untuk menyimpan static lib
LIB_DIR="./libphp"
mkdir -p "$LIB_DIR"

echo "🐘 Downloading static PHP library..."

# Download static library dari FrankenPHP releases (ini URL contoh, perlu disesuaikan dengan real URL jika berubah)
# FrankenPHP menyediakan static build script, tapi kita coba ambil artifact jika ada, atau gunakan docker.
# TAPI, cara paling reliable tanpa docker adalah menggunakan script compile-php-static dari FrankenPHP atau download pre-built lib.

# KARENA URL static lib tidak selalu stabil/langsung direct link,
# kita akan coba pendekatan hybrid:
# 1. Cek jika user punya docker -> gunakan docker build (paling aman)
# 2. Jika tidak, beri instruksi manual.

if command -v docker &> /dev/null; then
    echo "🐳 Docker detected. Building using Docker (static build)..."
    
    # Buat Dockerfile sementara untuk build
    cat > Dockerfile.build <<EOF
FROM dunglas/frankenphp:static-builder-php8.4

WORKDIR /go/src/app
COPY . .

# Build static binary
RUN CGO_ENABLED=1 \
    go build \
    -tags "cgo netgo osusergo static_build nowatcher" \
    -ldflags "-w -s -extldflags '-static'" \
    -o frankenphp-bridge .
EOF

    # Jalankan build
    docker build -t frankenphp-bridge-builder -f Dockerfile.build .
    
    # Copy binary keluar
    id=$(docker create frankenphp-bridge-builder)
    docker cp $id:/go/src/app/frankenphp-bridge ./frankenphp-bridge
    docker rm -v $id
    rm Dockerfile.build
    
    echo "✅ Build success! Binary: ./frankenphp-bridge"
    exit 0
else
    echo "❌ Docker not found."
    echo "To build FrankenPHP bridge without system PHP installed, you need Docker."
    echo "Please install Docker or install php-dev locally (apt install php-dev libphp-embed)."
    exit 1
fi
