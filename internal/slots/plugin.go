package slots

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/wasm"

	"github.com/go-chi/chi/v5"
)

// Global plugin manager for cleanup
var globalPluginManager *wasm.PluginManager

// RegisterPluginSlots loads and registers all plugins (WASM & Sidecar)
func RegisterPluginSlots(eng *engine.Engine, r *chi.Mux, dbMgr *dbmanager.DBManager) {
	// Check if plugins are enabled
	enabled := os.Getenv("ZENO_PLUGINS_ENABLED")
	if enabled != "true" && enabled != "1" {
		slog.Debug("Plugins disabled", "env", "ZENO_PLUGINS_ENABLED")
		return
	}

	// Get plugin directory
	pluginDir := os.Getenv("ZENO_PLUGIN_DIR")
	if pluginDir == "" {
		pluginDir = "./plugins"
	}

	// Create plugin manager
	ctx := context.Background()
	pm, err := wasm.NewPluginManager(ctx, pluginDir)
	if err != nil {
		slog.Error("Failed to create plugin manager", "error", err)
		return
	}

	// Set host callbacks
	setupHostCallbacks(pm, eng, dbMgr)

	// Load plugins from directory
	if err := pm.LoadPluginsFromDir(pluginDir); err != nil {
		slog.Error("Failed to load plugins", "error", err)
		return
	}

	// Register slots from all loaded plugins
	plugins := pm.ListPlugins()
	totalSlots := 0

	for _, plugin := range plugins {
		for _, slot := range plugin.Slots {
			// Register each slot
			slotName := slot.Name
			pluginName := plugin.Manifest.Name

			eng.Register(slotName, func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
				// Inject scope into context for host functions
				ctx = context.WithValue(ctx, "scope", scope)
				return executePluginSlot(ctx, pm, pluginName, slotName, node, scope)
			}, engine.SlotMeta{
				Description: slot.Description,
				Example:     slot.Example,
				Inputs:      convertInputMeta(slot.Inputs),
			})

			totalSlots++
		}
	}

	if totalSlots > 0 {
		slog.Info("🔌 Plugins registered",
			"plugins", len(plugins),
			"slots", totalSlots)
	}

	// Store plugin manager for cleanup
	globalPluginManager = pm

	if r == nil {
		return
	}

	// Register admin API for hot reload
	r.Post("/api/admin/plugins/reload", func(w http.ResponseWriter, req *http.Request) {
		// Check for specific plugin name query param
		pluginName := req.URL.Query().Get("name")
		
		var err error
		if pluginName != "" {
			err = pm.ReloadPlugin(pluginName)
		} else {
			// Reload all if no name specified
			errs := pm.ReloadAllPlugins()
			if len(errs) > 0 {
				// Aggregate errors
				errStr := []string{}
				for p, e := range errs {
					errStr = append(errStr, fmt.Sprintf("%s: %v", p, e))
				}
				err = fmt.Errorf("reload failed for: %s", strings.Join(errStr, ", "))
			}
		}

		w.Header().Set("Content-Type", "application/json")
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(map[string]string{"status": "reloaded"})
	})
}

// CleanupPlugins gracefully shuts down all plugins
// This should be called during application shutdown
func CleanupPlugins() {
	if globalPluginManager != nil {
		slog.Info("🔌 Cleaning up plugins...")
		if err := globalPluginManager.Close(); err != nil {
			slog.Error("Failed to cleanup plugins", "error", err)
		} else {
			slog.Info("✅ Plugins cleaned up")
		}
		globalPluginManager = nil
	}
}

// setupHostCallbacks configures host function callbacks
func setupHostCallbacks(pm *wasm.PluginManager, eng *engine.Engine, dbMgr *dbmanager.DBManager) {
	// Logging callback
	pm.SetHostCallback("log", func(ctx context.Context, level, message string) {
		switch level {
		case "debug":
			slog.Debug("[Plugin] " + message)
		case "info":
			slog.Info("[Plugin] " + message)
		case "warn":
			slog.Warn("[Plugin] " + message)
		case "error":
			slog.Error("[Plugin] " + message)
		default:
			slog.Info("[Plugin] " + message)
		}
	})

	// Database query callback
	if dbMgr != nil {
		pm.SetHostCallback("db_query", func(ctx context.Context, connection, sql string, params map[string]interface{}) ([]map[string]interface{}, error) {
			// Get database connection
			db := dbMgr.GetConnection(connection)
			if db == nil {
				return nil, fmt.Errorf("database connection not found: %s", connection)
			}

			// Execute query
			rows, err := db.Query(sql)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			// Get column names
			columns, err := rows.Columns()
			if err != nil {
				return nil, err
			}

			// Scan results
			var results []map[string]interface{}
			for rows.Next() {
				// Create slice for scanning
				values := make([]interface{}, len(columns))
				valuePtrs := make([]interface{}, len(columns))
				for i := range values {
					valuePtrs[i] = &values[i]
				}

				if err := rows.Scan(valuePtrs...); err != nil {
					return nil, err
				}

				// Convert to map
				row := make(map[string]interface{})
				for i, col := range columns {
					val := values[i]
					// Convert []byte to string
					if b, ok := val.([]byte); ok {
						row[col] = string(b)
					} else {
						row[col] = val
					}
				}
				results = append(results, row)
			}

			return results, nil
		})
	}

	// HTTP request callback
	pm.SetHostCallback("http_request", func(ctx context.Context, method, url string, headers, body map[string]interface{}) (map[string]interface{}, error) {
		// Create HTTP client with timeout
		client := &http.Client{
			Timeout: 30 * time.Second,
		}

		// Prepare request body
		var reqBody io.Reader
		if body != nil && len(body) > 0 {
			jsonBody, err := json.Marshal(body)
			if err != nil {
				return nil, fmt.Errorf("failed to marshal request body: %w", err)
			}
			reqBody = bytes.NewBuffer(jsonBody)
		}

		// Create HTTP request
		req, err := http.NewRequest(strings.ToUpper(method), url, reqBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		// Set headers
		if headers != nil {
			for key, value := range headers {
				if strValue, ok := value.(string); ok {
					req.Header.Set(key, strValue)
				}
			}
		}

		// Set default Content-Type if not provided and body exists
		if reqBody != nil && req.Header.Get("Content-Type") == "" {
			req.Header.Set("Content-Type", "application/json")
		}

		// Execute request
		resp, err := client.Do(req)
		if err != nil {
			return nil, fmt.Errorf("request failed: %w", err)
		}
		defer resp.Body.Close()

		// Read response body
		respBody, err := io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to read response: %w", err)
		}

		// Try to parse as JSON, fallback to string
		var parsedBody interface{}
		if err := json.Unmarshal(respBody, &parsedBody); err != nil {
			// Not JSON, return as string
			parsedBody = string(respBody)
		}

		// Build response
		response := map[string]interface{}{
			"status":  resp.StatusCode,
			"body":    parsedBody,
			"headers": make(map[string]string),
		}

		// Copy response headers
		for key, values := range resp.Header {
			if len(values) > 0 {
				response["headers"].(map[string]string)[key] = values[0]
			}
		}

		return response, nil
	})

	// Scope get callback
	pm.SetHostCallback("scope_get", func(ctx context.Context, key string) (interface{}, error) {
		scope, ok := ctx.Value("scope").(*engine.Scope)
		if !ok || scope == nil {
			return nil, fmt.Errorf("scope access not available in this context")
		}

		val, found := scope.Get(key)
		if !found {
			return nil, fmt.Errorf("variable $%s not found", key)
		}
		return val, nil
	})

	// Scope set callback
	pm.SetHostCallback("scope_set", func(ctx context.Context, key string, value interface{}) error {
		scope, ok := ctx.Value("scope").(*engine.Scope)
		if !ok || scope == nil {
			return fmt.Errorf("scope access not available in this context")
		}

		scope.Set(key, value)
		return nil
	})

	// File read callback
	pm.SetHostCallback("file_read", func(ctx context.Context, path string) (string, error) {
		// Clean and validate path
		cleanPath := filepath.Clean(path)
		
		// Check if path is absolute or tries to escape
		if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
			return "", fmt.Errorf("absolute paths and parent directory access not allowed")
		}
		
		// Read file
		content, err := os.ReadFile(cleanPath)
		if err != nil {
			return "", fmt.Errorf("failed to read file: %w", err)
		}
		
		return string(content), nil
	})

	// File write callback
	pm.SetHostCallback("file_write", func(ctx context.Context, path, content string) error {
		// Clean and validate path
		cleanPath := filepath.Clean(path)
		
		// Check if path is absolute or tries to escape
		if filepath.IsAbs(cleanPath) || strings.Contains(cleanPath, "..") {
			return fmt.Errorf("absolute paths and parent directory access not allowed")
		}
		
		// Write file
		if err := os.WriteFile(cleanPath, []byte(content), 0644); err != nil {
			return fmt.Errorf("failed to write file: %w", err)
		}
		
		return nil
	})

	// Environment variable callback
	pm.SetHostCallback("env_get", func(ctx context.Context, key string) string {
		return os.Getenv(key)
	})
}

// executePluginSlot executes a plugin slot (WASM or Sidecar)
func executePluginSlot(ctx context.Context, pm *wasm.PluginManager, pluginName, slotName string, node *engine.Node, scope *engine.Scope) error {
	// Inject plugin name into context for permission checking in host functions
	ctx = context.WithValue(ctx, "pluginName", pluginName)

	// Parse parameters from node
	params := make(map[string]interface{})

	// Add node value as main parameter if present
	if node.Value != nil {
		params["value"] = node.Value
	}

	// [NEW] DEEP SCOPE INJECTION
	// Inject current scope into parameters for better PHP context
	params["_zeno_scope"] = scope.ToMap()

	// Add children as parameters
	for _, child := range node.Children {
		var value interface{}
		
		if child.Value != nil {
			// Try to preserve numeric types
			valStr := fmt.Sprintf("%v", child.Value)
			
			// Try parsing as float
			if f, err := strconv.ParseFloat(valStr, 64); err == nil {
				value = f
			} else if b, err := strconv.ParseBool(valStr); err == nil {
				// Try parsing as bool
				value = b
			} else {
				// Keep as string
				value = child.Value
			}
		} else if len(child.Children) > 0 {
			// Handle nested objects
			value = parseNodeValue(child, scope)
		}
		
		if value != nil {
			params[child.Name] = value
		}
	}

	// [NEW] ASYNC EXECUTION SUPPORT
	isAsync, _ := params["async"].(bool)
	if isAsync {
		go func() {
			_, _ = pm.ExecuteSlot(ctx, pluginName, slotName, params)
			// Note: result is ignored for fire-and-forget async
		}()
		return nil
	}

	// Execute plugin slot (Synchronous)
	response, err := pm.ExecuteSlot(ctx, pluginName, slotName, params)
	if err != nil {
		return fmt.Errorf("plugin execution failed: %w", err)
	}

	if !response.Success {
		return fmt.Errorf("plugin error: %s", response.Error)
	}

	// Store response data in scope
	if response.Data != nil {
		for key, value := range response.Data {
			scope.Set(key, value)
		}
	}

	return nil
}

// convertInputMeta converts WASM InputMeta to engine.InputMeta
func convertInputMeta(inputs map[string]wasm.InputMeta) map[string]engine.InputMeta {
	result := make(map[string]engine.InputMeta)
	for key, input := range inputs {
		result[key] = engine.InputMeta{
			Type:        input.Type,
			Required:    input.Required,
			Description: input.Description,
		}
	}
	return result
}
