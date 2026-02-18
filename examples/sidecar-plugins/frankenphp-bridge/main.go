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
//   - php.run    : Jalankan PHP script file
//   - php.health : Cek status bridge
package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"os/signal"
	"strings"
	"syscall"

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

	// Inject REQUEST_URI (escaped)
	if req, ok := params["request"].(map[string]interface{}); ok {
		if uri, ok := req["uri"].(string); ok {
			escaped := strings.ReplaceAll(uri, "'", `\'`)
			sb.WriteString(fmt.Sprintf("$_SERVER['REQUEST_URI'] = '%s';", escaped))
		}
		if method, ok := req["method"].(string); ok {
			escaped := strings.ReplaceAll(method, "'", `\'`)
			sb.WriteString(fmt.Sprintf("$_SERVER['REQUEST_METHOD'] = '%s';", escaped))
		}
	}

	return sb.String()
}

// executePHPCode menjalankan PHP code string dan mengembalikan output
func executePHPCode(code string, params map[string]interface{}) (string, int, error) {
	bootstrap := buildBootstrap(params)

	// Gabungkan bootstrap + user code
	fullCode := bootstrap + "\n?>\n" + code

	// Buat fake HTTP request untuk FrankenPHP
	req, err := http.NewRequest("GET", "http://localhost/eval", nil)
	if err != nil {
		return "", 500, err
	}

	// Buat response recorder untuk capture output
	rec := httptest.NewRecorder()

	// Buat FrankenPHP request context
	fpReq, err := frankenphp.NewRequestWithContext(req,
		frankenphp.WithRequestDocumentRoot(".", false),
	)
	if err != nil {
		return "", 500, err
	}

	// Override: jalankan code string langsung
	// FrankenPHP akan serve via ServeHTTP, tapi kita perlu inject code
	// Gunakan ExecutePHPCode untuk code string
	_ = rec
	_ = fpReq

	// Capture output via ExecutePHPCode (tidak butuh HTTP request)
	// Redirect stdout sementara
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

// executePHPScript menjalankan PHP script file dan mengembalikan output
func executePHPScript(script string, params map[string]interface{}) (string, int, error) {
	// Verifikasi file ada
	if _, err := os.Stat(script); os.IsNotExist(err) {
		return "", 404, fmt.Errorf("script not found: %s", script)
	}

	// Inject scope via environment variable (lebih aman dari superglobal injection)
	if scope, ok := params["scope"]; ok {
		scopeJSON, err := json.Marshal(scope)
		if err == nil {
			os.Setenv("ZENO_SCOPE", string(scopeJSON))
			defer os.Unsetenv("ZENO_SCOPE")
		}
	}

	// Capture stdout
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

// ─── Slot Handlers ──────────────────────────────────────────────────────────

func handlePluginInit(id string) Response {
	return Response{
		// Legacy format (tanpa type/id) untuk kompatibilitas dengan manager.go
		Success: true,
		Data: map[string]interface{}{
			"name":        "frankenphp-bridge",
			"version":     "1.0.0",
			"description": "PHP sidecar bridge menggunakan FrankenPHP (tanpa Caddy)",
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
					"description": "Jalankan PHP script file",
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
		return Response{
			Type:    "guest_response",
			ID:      id,
			Success: false,
			Error:   err.Error(),
			Data: map[string]interface{}{
				"output": output,
				"status": status,
			},
		}
	}

	return successResponse(id, map[string]interface{}{
		"output": output,
		"status": status,
	})
}

func handlePHPRun(id string, params map[string]interface{}) Response {
	script, ok := params["script"].(string)
	if !ok || script == "" {
		return errorResponse(id, "php.run: parameter 'script' (string) diperlukan")
	}

	output, status, err := executePHPScript(script, params)
	if err != nil {
		return Response{
			Type:    "guest_response",
			ID:      id,
			Success: false,
			Error:   err.Error(),
			Data: map[string]interface{}{
				"output": output,
				"status": status,
			},
		}
	}

	return successResponse(id, map[string]interface{}{
		"output": output,
		"status": status,
	})
}

func handlePHPHealth(id string) Response {
	return successResponse(id, map[string]interface{}{
		"status":  "healthy",
		"backend": "frankenphp",
		"version": "1.0.0",
	})
}

// ─── Main Loop ──────────────────────────────────────────────────────────────

func main() {
	// Setup logging ke stderr (tidak mengganggu stdout JSON-RPC)
	logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	// Init FrankenPHP (tanpa Caddy, tanpa HTTP server)
	slog.Info("🐘 Initializing FrankenPHP...")
	if err := frankenphp.Init(); err != nil {
		slog.Error("❌ Failed to initialize FrankenPHP", "error", err)
		os.Exit(1)
	}
	defer frankenphp.Shutdown()
	slog.Info("✅ FrankenPHP initialized")

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-sigCh
		slog.Info("🛑 Shutting down FrankenPHP bridge...")
		frankenphp.Shutdown()
		os.Exit(0)
	}()

	// JSON-RPC loop via stdin/stdout
	slog.Info("🚀 FrankenPHP bridge ready, listening on stdin...")

	stdout := os.Stdout
	scanner := bufio.NewScanner(os.Stdin)

	// Perbesar buffer untuk payload besar (misal: script output besar)
	const maxBuf = 10 * 1024 * 1024 // 10MB
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
		case "php.health":
			resp = handlePHPHealth(req.ID)
		default:
			// Abaikan host_response (balasan dari ZenoEngine untuk host_call kita)
			if req.Type == "host_response" {
				continue
			}
			resp = errorResponse(req.ID, fmt.Sprintf("unknown slot: %s", req.SlotName))
		}

		writeResponse(stdout, resp)
	}

	if err := scanner.Err(); err != nil {
		slog.Error("❌ stdin read error", "error", err)
		os.Exit(1)
	}
}
