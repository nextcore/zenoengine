// Hello World WASM Plugin for ZenoEngine
// Build: tinygo build -o hello.wasm -target=wasi main.go

package main

import (
	"encoding/json"
	"unsafe"
)

//export plugin_init
func plugin_init() int32 {
	metadata := map[string]string{
		"name":        "hello",
		"version":     "1.0.0",
		"author":      "ZenoEngine Team",
		"description": "Simple hello world plugin",
		"license":     "MIT",
	}
	return allocJSON(metadata)
}

//export plugin_register_slots
func plugin_register_slots() int32 {
	slots := []map[string]interface{}{
		{
			"name":        "hello.greet",
			"description": "Greet someone with a message",
			"example":     "hello.greet { name: 'World' }",
			"inputs": map[string]interface{}{
				"name": map[string]interface{}{
					"type":        "string",
					"required":    true,
					"description": "Name to greet",
				},
			},
		},
		{
			"name":        "hello.log",
			"description": "Log a message using host logging",
			"example":     "hello.log { message: 'Test' }",
		},
	}
	return allocJSON(slots)
}

//export plugin_execute
func plugin_execute(slotNamePtr, slotNameLen, paramsPtr, paramsLen int32) int32 {
	slotName := ptrToString(slotNamePtr, slotNameLen)
	paramsJSON := ptrToString(paramsPtr, paramsLen)

	// Parse parameters
	var request map[string]interface{}
	if err := json.Unmarshal([]byte(paramsJSON), &request); err != nil {
		return allocJSON(map[string]interface{}{
			"success": false,
			"error":   "invalid parameters: " + err.Error(),
		})
	}

	params, _ := request["parameters"].(map[string]interface{})

	switch slotName {
	case "hello.greet":
		return executeGreet(params)
	case "hello.log":
		return executeLog(params)
	default:
		return allocJSON(map[string]interface{}{
			"success": false,
			"error":   "unknown slot: " + slotName,
		})
	}
}

//export plugin_cleanup
func plugin_cleanup() {
	// Nothing to cleanup for this simple plugin
}

//export alloc
func alloc(size int32) *byte {
	buf := make([]byte, size)
	return &buf[0]
}

// executeGreet implements hello.greet slot
func executeGreet(params map[string]interface{}) int32 {
	name, ok := params["name"].(string)
	if !ok {
		name = "World"
	}

	message := "Hello, " + name + "! 👋"

	// Log using host function
	hostLog("info", "Greeting: "+message)

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"message": message,
		},
	}
	return allocJSON(response)
}

// executeLog implements hello.log slot
func executeLog(params map[string]interface{}) int32 {
	message, ok := params["message"].(string)
	if !ok {
		message = "No message provided"
	}

	// Call host logging function
	hostLog("info", "[Plugin] "+message)

	response := map[string]interface{}{
		"success": true,
		"data": map[string]interface{}{
			"logged": true,
		},
	}
	return allocJSON(response)
}

// Helper: allocate JSON in WASM memory
func allocJSON(v interface{}) int32 {
	data, _ := json.Marshal(v)
	ptr := alloc(int32(len(data) + 1)) // +1 for null terminator
	slice := unsafe.Slice(ptr, len(data)+1)
	copy(slice, data)
	slice[len(data)] = 0 // null terminator
	return int32(uintptr(unsafe.Pointer(ptr)))
}

// Helper: convert pointer to string
func ptrToString(ptr, length int32) string {
	return unsafe.String((*byte)(unsafe.Pointer(uintptr(ptr))), length)
}

// Host function: log
func hostLog(level, message string) {
	levelBytes := []byte(level)
	msgBytes := []byte(message)

	hostLogRaw(
		int32(uintptr(unsafe.Pointer(&levelBytes[0]))),
		int32(len(levelBytes)),
		int32(uintptr(unsafe.Pointer(&msgBytes[0]))),
		int32(len(msgBytes)),
	)
}

//go:wasm-module env
//export host_log
func hostLogRaw(levelPtr, levelLen, msgPtr, msgLen int32)

func main() {
	// Required for WASI
}
