#!/bin/bash
echo "🏗️  Building Example: Currency Converter..."

# Go back to ZenoWasm root to build
cd ../../

# Build WASM optimized
GOOS=js GOARCH=wasm go build -ldflags="-s -w" -o examples/currency-converter/public/zeno.wasm main.go

if [ $? -eq 0 ]; then
    echo "✅ WASM Built successfully at examples/currency-converter/public/zeno.wasm"

    # Ensure wasm_exec.js is there
    cp public/wasm_exec.js examples/currency-converter/public/

    echo "📱 Ready for Capacitor! Run 'npm install && npx cap run android' inside the example folder."
else
    echo "❌ Build failed"
    exit 1
fi
