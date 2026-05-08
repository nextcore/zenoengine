//go:build cgo

// FrankenPHP Sidecar Bridge untuk ZenoEngine
//
// Binary ini berjalan sebagai child process dari ZenoEngine.
// Komunikasi via stdin/stdout menggunakan protokol JSON-RPC.
//
// Build:
//
//	CGO_ENABLED=1 go build -o frankenphp-bridge .
//
// Slot yang tersedia:
//   - php.eval   : Jalankan PHP code string
//   - php.run    : Jalankan PHP script file (One-Shot CLI)
//   - php.request: Forward request ke internal HTTP server (Worker Mode)
//   - php.health : Cek status bridge
package main

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/dunglas/frankenphp"
)

// ─── Protokol JSON-RPC ──────────────────────────────────────────────────────

type Request struct {
	Type       string                 `json:"type"`
	ID         string                 `json:"id"`
	SlotName   string                 `json:"slot_name"`
	Parameters map[string]interface{} `json:"parameters"`
}

type Response struct {
	Type    string                 `json:"type"`
	ID      string                 `json:"id"`
	Success bool                   `json:"success"`
	Data    map[string]interface{} `json:"data,omitempty"`
	Error   string                 `json:"error,omitempty"`
}

// ─── Global State ───────────────────────────────────────────────────────────

var (
	httpClient *http.Client
	socketPath = "/tmp/zeno-php-" + fmt.Sprint(os.Getpid()) + ".sock"
	tcpAddr    string
	server     *http.Server
)

// ─── Helpers ────────────────────────────────────────────────────────────────

func writeResponse(w io.Writer, resp Response) {
	data, _ := json.Marshal(resp)
	fmt.Fprintln(w, string(data))
}

func errorResponse(id, msg string) Response {
	return Response{
		Type:    "guest_response",
		ID:      id,
		Success: false,
		Error:   msg,
	}
}

func successResponse(id string, data map[string]interface{}) Response {
	return Response{
		Type:    "guest_response",
		ID:      id,
		Success: true,
		Data:    data,
	}
}

// ─── PHP Execution ──────────────────────────────────────────────────────────

// buildBootstrap membangun PHP bootstrap code untuk inject scope dari Zeno
func buildBootstrap(params map[string]interface{}) string {
	var sb strings.Builder
	sb.WriteString("<?php ")

	// Inject ZENO_SCOPE ke $_SERVER
	if scope, ok := params["scope"]; ok {
		scopeJSON, err := json.Marshal(scope)
		if err == nil {
			// Escape single quotes untuk keamanan
			escaped := strings.ReplaceAll(string(scopeJSON), `\`, `\\`)
			escaped = strings.ReplaceAll(escaped, "'", `\'`)
			sb.WriteString(fmt.Sprintf("$_SERVER['ZENO_SCOPE'] = json_decode('%s', true);", escaped))
		}
	}

	return sb.String()
}

// executePHPCode menjalankan PHP code string dan mengembalikan output
func executePHPCode(code string, params map[string]interface{}) (string, int, error) {
	bootstrap := buildBootstrap(params)
	fullCode := bootstrap + "\n?>\n" + code

	// Capture output via ExecutePHPCode (tidak butuh HTTP request)
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	exitCode := frankenphp.ExecutePHPCode(fullCode)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	if exitCode != 0 {
		return buf.String(), exitCode, fmt.Errorf("PHP exited with code %d", exitCode)
	}

	return buf.String(), 200, nil
}

// executePHPScript menjalankan PHP script file dan mengembalikan output (CLI Mode)
func executePHPScript(script string, params map[string]interface{}) (string, int, error) {
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return "", 404, fmt.Errorf("script not found: %s", script)
	}

	if scope, ok := params["scope"]; ok {
		scopeJSON, err := json.Marshal(scope)
		if err == nil {
			os.Setenv("ZENO_SCOPE", string(scopeJSON))
			defer os.Unsetenv("ZENO_SCOPE")
		}
	}

	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	args := []string{script}
	exitCode := frankenphp.ExecuteScriptCLI(script, args)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	io.Copy(&buf, r)
	r.Close()

	if exitCode != 0 {
		return buf.String(), exitCode, fmt.Errorf("PHP script exited with code %d", exitCode)
	}

	return buf.String(), 200, nil
}

// ─── Internal Server (Worker Mode) ──────────────────────────────────────────

func startInternalServer() {
	if server != nil {
		slog.Info("⚠️ Internal server already running")
		return
	}

	// Check for TCP Port config
	port := os.Getenv("FRANKENPHP_PORT")
	host := os.Getenv("FRANKENPHP_HOST")
	if host == "" {
		host = "127.0.0.1"
	}

	var listener net.Listener
	var err error

	if port != "" {
		// TCP Mode
		tcpAddr = net.JoinHostPort(host, port)
		slog.Info("🔌 Configured to use TCP port", "addr", tcpAddr)
		listener, err = net.Listen("tcp", tcpAddr)
		if err != nil {
			slog.Error("❌ Failed to listen on TCP", "addr", tcpAddr, "error", err)
			return
		}
	} else {
		// Unix Socket Mode (Default)
		os.Remove(socketPath)
		slog.Info("🔌 Configured to use Unix Socket", "path", socketPath)
		listener, err = net.Listen("unix", socketPath)
		if err != nil {
			slog.Error("❌ Failed to listen on socket", "path", socketPath, "error", err)
			return
		}
	}

	// Server config
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		frankenphp.ServeHTTP(w, r)
	})

	server = &http.Server{
		Handler: handler,
	}

	// Setup HTTP Client
	if port != "" {
		// TCP Client
		httpClient = &http.Client{
			Timeout: 30 * time.Second,
		}
	} else {
		// Unix Socket Client
		httpClient = &http.Client{
			Transport: &http.Transport{
				DialContext: func(ctx context.Context, _, _ string) (net.Conn, error) {
					return net.Dial("unix", socketPath)
				},
			},
			Timeout: 30 * time.Second,
		}
	}

	go func() {
		addr := socketPath
		if tcpAddr != "" {
			addr = tcpAddr
		}
		slog.Info("🚀 Internal HTTP server starting on " + addr)
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			slog.Error("❌ Internal server failed", "error", err)
		}
	}()
}

// proxyRequest meneruskan request Zeno ke internal server
func proxyRequest(params map[string]interface{}) (map[string]interface{}, error) {
	if httpClient == nil {
		startInternalServer()
		time.Sleep(100 * time.Millisecond)
	}

	// Parse request parameters
	reqData, _ := params["request"].(map[string]interface{})
	uri, _ := reqData["uri"].(string)
	method, _ := reqData["method"].(string)
	if uri == "" {
		uri = "/"
	}
	if method == "" {
		method = "GET"
	}

	var bodyReader io.Reader
	if body, ok := reqData["body"].(string); ok {
		bodyReader = strings.NewReader(body)
	} else if bodyMap, ok := reqData["body"].(map[string]interface{}); ok {
		jsonBody, _ := json.Marshal(bodyMap)
		bodyReader = bytes.NewReader(jsonBody)
	}

	// Tentukan Target URL
	targetURL := "http://unix" + uri // Default dummy host for unix
	if tcpAddr != "" {
		targetURL = "http://" + tcpAddr + uri
	}

	// Buat HTTP request ke internal server
	// URL host dummy, yang penting path
	req, err := http.NewRequest(method, targetURL, bodyReader)
	if err != nil {
		return nil, err
	}

	// Inject Headers
	if headers, ok := reqData["headers"].(map[string]interface{}); ok {
		for k, v := range headers {
			if val, ok := v.(string); ok {
				req.Header.Set(k, val)
			}
		}
	}

	// Default Content-Type jika body ada
	if bodyReader != nil && req.Header.Get("Content-Type") == "" {
		req.Header.Set("Content-Type", "application/json")
	}

	// Inject Scope via Header khusus (FrankenPHP bisa baca env/server vars, tapi lewat HTTP header paling mudah diparsing di PHP)
	if scope, ok := params["scope"]; ok {
		scopeJSON, _ := json.Marshal(scope)
		req.Header.Set("X-Zeno-Scope", string(scopeJSON))
	}

	// Eksekusi Request
	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Baca Response
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	respHeaders := make(map[string]string)
	for k, v := range resp.Header {
		respHeaders[k] = strings.Join(v, ", ")
	}

	return map[string]interface{}{
		"status":  resp.StatusCode,
		"headers": respHeaders,
		"body":    string(respBody),
	}, nil
}

// ─── Slot Handlers ──────────────────────────────────────────────────────────

func handlePluginInit(id string) Response {
	return Response{
		Success: true,
		Data: map[string]interface{}{
			"name":        "frankenphp-bridge",
			"version":     "1.1.0", // Bumped version for Worker support
			"description": "PHP sidecar bridge menggunakan FrankenPHP (Worker Supported)",
		},
	}
}

func handleRegisterSlots(id string) Response {
	return Response{
		Success: true,
		Data: map[string]interface{}{
			"slots": []map[string]interface{}{
				{
					"name":        "php.eval",
					"description": "Jalankan PHP code string langsung",
				},
				{
					"name":        "php.run",
					"description": "Jalankan PHP script file (One-Shot)",
				},
				{
					"name":        "php.request",
					"description": "Forward request ke worker (High Performance)",
				},
				{
					"name":        "php.extensions",
					"description": "List loaded PHP extensions",
				},
				{
					"name":        "php.health",
					"description": "Cek status FrankenPHP bridge",
				},
			},
		},
	}
}

func handlePHPEval(id string, params map[string]interface{}) Response {
	code, ok := params["code"].(string)
	if !ok || code == "" {
		return errorResponse(id, "php.eval: parameter 'code' (string) diperlukan")
	}
	output, status, err := executePHPCode(code, params)
	if err != nil {
		return Response{Type: "guest_response", ID: id, Success: false, Error: err.Error(), Data: map[string]interface{}{"output": output, "status": status}}
	}
	return successResponse(id, map[string]interface{}{"output": output, "status": status})
}

func handlePHPRun(id string, params map[string]interface{}) Response {
	script, ok := params["script"].(string)
	if !ok || script == "" {
		return errorResponse(id, "php.run: parameter 'script' (string) diperlukan")
	}
	output, status, err := executePHPScript(script, params)
	if err != nil {
		return Response{Type: "guest_response", ID: id, Success: false, Error: err.Error(), Data: map[string]interface{}{"output": output, "status": status}}
	}
	return successResponse(id, map[string]interface{}{"output": output, "status": status})
}

func handlePHPRequest(id string, params map[string]interface{}) Response {
	respData, err := proxyRequest(params)
	if err != nil {
		return errorResponse(id, fmt.Sprintf("request failed: %v", err))
	}
	return successResponse(id, respData)
}

func handlePHPExtensions(id string) Response {
	code := `echo json_encode(get_loaded_extensions());`
	output, status, err := executePHPCode(code, map[string]interface{}{})
	if err != nil || status != 0 {
		return errorResponse(id, fmt.Sprintf("failed to get extensions: %v", err))
	}
	var extensions []string
	json.Unmarshal([]byte(output), &extensions)
	return successResponse(id, map[string]interface{}{"extensions": extensions, "count": len(extensions)})
}

func handlePHPHealth(id string) Response {
	mode := "cli"
	if server != nil {
		mode = "worker"
	}
	return successResponse(id, map[string]interface{}{
		"status":  "healthy",
		"backend": "frankenphp",
		"mode":    mode,
		"version": "1.1.0",
	})
}

// ─── Main Loop ──────────────────────────────────────────────────────────────

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelInfo}))
	slog.SetDefault(logger)

	slog.Info("🐘 Initializing FrankenPHP...")
	// Init FrankenPHP (will pick up FRANKENPHP_CONFIG env var for worker mode)
	if err := frankenphp.Init(); err != nil {
		slog.Error("❌ Failed to initialize FrankenPHP", "error", err)
		os.Exit(1)
	}
	defer frankenphp.Shutdown()

	// Check if Worker config is present, then auto-start server
	if os.Getenv("FRANKENPHP_CONFIG") != "" {
		slog.Info("Worker config detected, starting internal server...")
		startInternalServer()
	}

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("🛑 Shutting down FrankenPHP bridge...")
		if server != nil {
			server.Close()
			os.Remove(socketPath)
		}
		frankenphp.Shutdown()
		os.Exit(0)
	}()

	slog.Info(fmt.Sprintf("🚀 FrankenPHP bridge ready (PID: %d)", os.Getpid()))

	scanner := bufio.NewScanner(os.Stdin)
	const maxBuf = 10 * 1024 * 1024
	buf := make([]byte, maxBuf)
	scanner.Buffer(buf, maxBuf)

	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		var req Request
		if err := json.Unmarshal([]byte(line), &req); err != nil {
			slog.Error("⚠️ JSON parse error", "error", err)
			continue
		}

		var resp Response
		switch req.SlotName {
		case "plugin_init":
			resp = handlePluginInit(req.ID)
		case "plugin_register_slots":
			resp = handleRegisterSlots(req.ID)
		case "php.eval":
			resp = handlePHPEval(req.ID, req.Parameters)
		case "php.run":
			resp = handlePHPRun(req.ID, req.Parameters)
		case "php.request":
			resp = handlePHPRequest(req.ID, req.Parameters)
		case "php.extensions":
			resp = handlePHPExtensions(req.ID)
		case "php.health":
			resp = handlePHPHealth(req.ID)
		default:
			if req.Type == "host_response" {
				continue
			}
			resp = errorResponse(req.ID, fmt.Sprintf("unknown slot: %s", req.SlotName))
		}
		writeResponse(os.Stdout, resp)
	}
}
