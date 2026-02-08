#!/bin/bash
echo "🏗️  Building ZenoWasm..."

# 1. Build WASM with optimizations (strip debug info)
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o public/zeno.wasm main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

# 2. Check Size
SIZE=$(du -h public/zeno.wasm | cut -f1)
echo "✅ Build success! Size: $SIZE"

# 3. Compress (Simulate Production)
echo "📦 Compressing (gzip)..."
gzip -k -f -9 public/zeno.wasm
GZ_SIZE=$(du -h public/zeno.wasm.gz | cut -f1)
echo "✅ Gzipped Size: $GZ_SIZE"

echo "🚀 Ready to deploy in ./public/"
