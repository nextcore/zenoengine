package wasm

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"github.com/tetratelabs/wazero/api"
)

// HostFunctions provides functions that WASM plugins can call
type HostFunctions struct {
	runtime *Runtime

	// Callbacks for host operations
	OnLog         func(level string, message string)
	OnDBQuery     func(connection, sql string, params map[string]interface{}) ([]map[string]interface{}, error)
	OnHTTPRequest func(method, url string, headers, body map[string]interface{}) (map[string]interface{}, error)
	OnScopeGet    func(key string) (interface{}, error)
	OnScopeSet    func(key string, value interface{}) error
	OnFileRead    func(path string) (string, error)
	OnFileWrite   func(path, content string) error
	OnEnvGet      func(key string) string
}

// NewHostFunctions creates host functions for WASM plugins
func NewHostFunctions(runtime *Runtime) *HostFunctions {
	return &HostFunctions{
		runtime: runtime,
		// Default implementations (can be overridden)
		OnLog: func(level, message string) {
			slog.Info("[WASM Plugin]", "level", level, "message", message)
		},
		OnEnvGet: func(key string) string {
			return "" // Default: no env access
		},
	}
}

// RegisterHostFunctions registers all host functions with the WASM runtime
func (hf *HostFunctions) RegisterHostFunctions(ctx context.Context, moduleName string) error {
	// Create host module builder
	hostBuilder := hf.runtime.runtime.NewHostModuleBuilder("env")

	// Register log function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostLog).
		Export("host_log")

	// Register db_query function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostDBQuery).
		Export("host_db_query")

	// Register http_request function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostHTTPRequest).
		Export("host_http_request")

	// Register scope_get function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostScopeGet).
		Export("host_scope_get")

	// Register scope_set function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostScopeSet).
		Export("host_scope_set")

	// Register file_read function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostFileRead).
		Export("host_file_read")

	// Register file_write function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostFileWrite).
		Export("host_file_write")

	// Register env_get function
	hostBuilder.NewFunctionBuilder().
		WithFunc(hf.hostEnvGet).
		Export("host_env_get")

	// Instantiate host module
	_, err := hostBuilder.Instantiate(ctx)
	if err != nil {
		return fmt.Errorf("failed to instantiate host module: %w", err)
	}

	return nil
}

// hostLog handles log calls from WASM
// Signature: host_log(level_ptr: i32, level_len: i32, msg_ptr: i32, msg_len: i32)
func (hf *HostFunctions) hostLog(ctx context.Context, m api.Module, levelPtr, levelLen, msgPtr, msgLen uint32) {
	// Read level string
	levelBytes, ok := m.Memory().Read(levelPtr, levelLen)
	if !ok {
		slog.Error("Failed to read log level from WASM memory")
		return
	}
	level := string(levelBytes)

	// Read message string
	msgBytes, ok := m.Memory().Read(msgPtr, msgLen)
	if !ok {
		slog.Error("Failed to read log message from WASM memory")
		return
	}
	message := string(msgBytes)

	// Call callback
	if hf.OnLog != nil {
		hf.OnLog(level, message)
	}
}

// hostDBQuery handles database query calls from WASM
// Signature: host_db_query(request_ptr: i32, request_len: i32) -> response_ptr: i32
func (hf *HostFunctions) hostDBQuery(ctx context.Context, m api.Module, requestPtr, requestLen uint32) uint32 {
	// Read request JSON
	requestBytes, ok := m.Memory().Read(requestPtr, requestLen)
	if !ok {
		return hf.allocateErrorResponse(m, "failed to read request")
	}

	// Parse request
	var request struct {
		Connection string                 `json:"connection"`
		SQL        string                 `json:"sql"`
		Params     map[string]interface{} `json:"params"`
	}
	if err := json.Unmarshal(requestBytes, &request); err != nil {
		return hf.allocateErrorResponse(m, fmt.Sprintf("invalid request JSON: %v", err))
	}

	// Call callback
	if hf.OnDBQuery == nil {
		return hf.allocateErrorResponse(m, "database access not available")
	}

	rows, err := hf.OnDBQuery(request.Connection, request.SQL, request.Params)
	if err != nil {
		return hf.allocateErrorResponse(m, err.Error())
	}

	// Serialize response
	response := HostResponse{
		Success: true,
		Data: map[string]interface{}{
			"rows": rows,
		},
	}

	return hf.allocateJSONResponse(m, response)
}

// hostHTTPRequest handles HTTP request calls from WASM
// Signature: host_http_request(request_ptr: i32, request_len: i32) -> response_ptr: i32
func (hf *HostFunctions) hostHTTPRequest(ctx context.Context, m api.Module, requestPtr, requestLen uint32) uint32 {
	// Read request JSON
	requestBytes, ok := m.Memory().Read(requestPtr, requestLen)
	if !ok {
		return hf.allocateErrorResponse(m, "failed to read request")
	}

	// Parse request
	var request struct {
		Method  string                 `json:"method"`
		URL     string                 `json:"url"`
		Headers map[string]interface{} `json:"headers"`
		Body    map[string]interface{} `json:"body"`
	}
	if err := json.Unmarshal(requestBytes, &request); err != nil {
		return hf.allocateErrorResponse(m, fmt.Sprintf("invalid request JSON: %v", err))
	}

	// Call callback
	if hf.OnHTTPRequest == nil {
		return hf.allocateErrorResponse(m, "HTTP access not available")
	}

	result, err := hf.OnHTTPRequest(request.Method, request.URL, request.Headers, request.Body)
	if err != nil {
		return hf.allocateErrorResponse(m, err.Error())
	}

	// Serialize response
	response := HostResponse{
		Success: true,
		Data:    result,
	}

	return hf.allocateJSONResponse(m, response)
}

// hostScopeGet handles scope variable get calls from WASM
// Signature: host_scope_get(key_ptr: i32, key_len: i32) -> response_ptr: i32
func (hf *HostFunctions) hostScopeGet(ctx context.Context, m api.Module, keyPtr, keyLen uint32) uint32 {
	// Read key
	keyBytes, ok := m.Memory().Read(keyPtr, keyLen)
	if !ok {
		return hf.allocateErrorResponse(m, "failed to read key")
	}
	key := string(keyBytes)

	// Call callback
	if hf.OnScopeGet == nil {
		return hf.allocateErrorResponse(m, "scope access not available")
	}

	value, err := hf.OnScopeGet(key)
	if err != nil {
		return hf.allocateErrorResponse(m, err.Error())
	}

	// Serialize response
	response := HostResponse{
		Success: true,
		Data: map[string]interface{}{
			"value": value,
		},
	}

	return hf.allocateJSONResponse(m, response)
}

// hostScopeSet handles scope variable set calls from WASM
// Signature: host_scope_set(request_ptr: i32, request_len: i32) -> success: i32
func (hf *HostFunctions) hostScopeSet(ctx context.Context, m api.Module, requestPtr, requestLen uint32) uint32 {
	// Read request JSON
	requestBytes, ok := m.Memory().Read(requestPtr, requestLen)
	if !ok {
		return 0 // failure
	}

	// Parse request
	var request struct {
		Key   string      `json:"key"`
		Value interface{} `json:"value"`
	}
	if err := json.Unmarshal(requestBytes, &request); err != nil {
		return 0 // failure
	}

	// Call callback
	if hf.OnScopeSet == nil {
		return 0 // failure
	}

	if err := hf.OnScopeSet(request.Key, request.Value); err != nil {
		return 0 // failure
	}

	return 1 // success
}

// hostFileRead handles file read calls from WASM
// Signature: host_file_read(path_ptr: i32, path_len: i32) -> response_ptr: i32
func (hf *HostFunctions) hostFileRead(ctx context.Context, m api.Module, pathPtr, pathLen uint32) uint32 {
	// Read path
	pathBytes, ok := m.Memory().Read(pathPtr, pathLen)
	if !ok {
		return hf.allocateErrorResponse(m, "failed to read path")
	}
	path := string(pathBytes)

	// Call callback
	if hf.OnFileRead == nil {
		return hf.allocateErrorResponse(m, "file access not available")
	}

	content, err := hf.OnFileRead(path)
	if err != nil {
		return hf.allocateErrorResponse(m, err.Error())
	}

	// Serialize response
	response := HostResponse{
		Success: true,
		Data: map[string]interface{}{
			"content": content,
		},
	}

	return hf.allocateJSONResponse(m, response)
}

// hostFileWrite handles file write calls from WASM
// Signature: host_file_write(request_ptr: i32, request_len: i32) -> success: i32
func (hf *HostFunctions) hostFileWrite(ctx context.Context, m api.Module, requestPtr, requestLen uint32) uint32 {
	// Read request JSON
	requestBytes, ok := m.Memory().Read(requestPtr, requestLen)
	if !ok {
		return 0 // failure
	}

	// Parse request
	var request struct {
		Path    string `json:"path"`
		Content string `json:"content"`
	}
	if err := json.Unmarshal(requestBytes, &request); err != nil {
		return 0 // failure
	}

	// Call callback
	if hf.OnFileWrite == nil {
		return 0 // failure
	}

	if err := hf.OnFileWrite(request.Path, request.Content); err != nil {
		return 0 // failure
	}

	return 1 // success
}

// hostEnvGet handles environment variable get calls from WASM
// Signature: host_env_get(key_ptr: i32, key_len: i32) -> value_ptr: i32
func (hf *HostFunctions) hostEnvGet(ctx context.Context, m api.Module, keyPtr, keyLen uint32) uint32 {
	// Read key
	keyBytes, ok := m.Memory().Read(keyPtr, keyLen)
	if !ok {
		return hf.allocateString(m, "")
	}
	key := string(keyBytes)

	// Call callback
	value := ""
	if hf.OnEnvGet != nil {
		value = hf.OnEnvGet(key)
	}

	return hf.allocateString(m, value)
}

// Helper: allocate error response in WASM memory
func (hf *HostFunctions) allocateErrorResponse(m api.Module, errorMsg string) uint32 {
	response := HostResponse{
		Success: false,
		Error:   errorMsg,
	}
	return hf.allocateJSONResponse(m, response)
}

// Helper: allocate JSON response in WASM memory
func (hf *HostFunctions) allocateJSONResponse(m api.Module, response interface{}) uint32 {
	jsonBytes, err := json.Marshal(response)
	if err != nil {
		slog.Error("Failed to marshal response", "error", err)
		return 0
	}

	return hf.allocateString(m, string(jsonBytes))
}

// Helper: allocate string in WASM memory
func (hf *HostFunctions) allocateString(m api.Module, str string) uint32 {
	bytes := []byte(str)
	length := uint32(len(bytes))

	// Call alloc function in WASM module
	allocFn := m.ExportedFunction("alloc")
	if allocFn == nil {
		slog.Error("WASM module does not export 'alloc' function")
		return 0
	}

	results, err := allocFn.Call(context.Background(), uint64(length))
	if err != nil {
		slog.Error("Failed to allocate memory in WASM", "error", err)
		return 0
	}

	ptr := uint32(results[0])

	// Write string to allocated memory
	if !m.Memory().Write(ptr, bytes) {
		slog.Error("Failed to write to WASM memory", "ptr", ptr)
		return 0
	}

	return ptr
}
