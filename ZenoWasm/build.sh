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

# 3. Compress (Gzip)
echo "📦 Compressing (gzip)..."
gzip -k -f -9 public/zeno.wasm
GZ_SIZE=$(du -h public/zeno.wasm.gz | cut -f1)
echo "✅ Gzipped Size: $GZ_SIZE"

# 4. Compress (Brotli) - If available
if command -v brotli &> /dev/null; then
    echo "📦 Compressing (brotli)..."
    brotli -f -q 11 -o public/zeno.wasm.br public/zeno.wasm
    BR_SIZE=$(du -h public/zeno.wasm.br | cut -f1)
    echo "✅ Brotli Size: $BR_SIZE"
else
    echo "⚠️  Brotli not installed. Skipping .br generation."
    echo "   (Install 'brotli' for even smaller production builds)"
fi

echo "🚀 Ready to deploy in ./public/"
