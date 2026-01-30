// Simple test WASM module for testing runtime
// Build with: tinygo build -o test.wasm -target=wasi test.go

package main

//export add
func add(a, b int32) int32 {
	return a + b
}

//export multiply
func multiply(a, b int32) int32 {
	return a * b
}

//export alloc
func alloc(size int32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}

func main() {
	// Required for WASI
}
