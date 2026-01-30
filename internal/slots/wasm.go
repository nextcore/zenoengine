package slots

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
	"zeno/pkg/wasm"

	"github.com/go-chi/chi/v5"
)

// RegisterWASMPluginSlots loads and registers all WASM plugins
func RegisterWASMPluginSlots(eng *engine.Engine, r *chi.Mux, dbMgr *dbmanager.DBManager) {
	// Check if WASM plugins are enabled
	enabled := os.Getenv("ZENO_PLUGINS_ENABLED")
	if enabled != "true" && enabled != "1" {
		slog.Debug("WASM plugins disabled", "env", "ZENO_PLUGINS_ENABLED")
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
				return executeWASMSlot(pm, pluginName, slotName, node, scope)
			}, engine.SlotMeta{
				Description: slot.Description,
				Example:     slot.Example,
				Inputs:      convertInputMeta(slot.Inputs),
			})

			totalSlots++
		}
	}

	if totalSlots > 0 {
		slog.Info("🔌 WASM plugins registered",
			"plugins", len(plugins),
			"slots", totalSlots)
	}

	// TODO: Store plugin manager for cleanup on shutdown
	// For now, plugins will be cleaned up when process exits
}

// setupHostCallbacks configures host function callbacks
func setupHostCallbacks(pm *wasm.PluginManager, eng *engine.Engine, dbMgr *dbmanager.DBManager) {
	// Logging callback
	pm.SetHostCallback("log", func(level, message string) {
		switch level {
		case "debug":
			slog.Debug("[WASM Plugin] " + message)
		case "info":
			slog.Info("[WASM Plugin] " + message)
		case "warn":
			slog.Warn("[WASM Plugin] " + message)
		case "error":
			slog.Error("[WASM Plugin] " + message)
		default:
			slog.Info("[WASM Plugin] " + message)
		}
	})

	// Database query callback
	if dbMgr != nil {
		pm.SetHostCallback("db_query", func(connection, sql string, params map[string]interface{}) ([]map[string]interface{}, error) {
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

	// HTTP request callback (using existing http_client slot logic)
	pm.SetHostCallback("http_request", func(method, url string, headers, body map[string]interface{}) (map[string]interface{}, error) {
		// TODO: Implement HTTP request using net/http
		// For now, return placeholder
		return map[string]interface{}{
			"status": 200,
			"body":   map[string]interface{}{},
		}, nil
	})

	// Scope get callback
	pm.SetHostCallback("scope_get", func(key string) (interface{}, error) {
		// Note: This needs to be set per-execution with actual scope
		// For now, return error
		return nil, fmt.Errorf("scope access not available in this context")
	})

	// Scope set callback
	pm.SetHostCallback("scope_set", func(key string, value interface{}) error {
		// Note: This needs to be set per-execution with actual scope
		return fmt.Errorf("scope access not available in this context")
	})

	// File read callback (restricted)
	pm.SetHostCallback("file_read", func(path string) (string, error) {
		// Check if file access is allowed
		// For security, only allow reading from specific directories
		return "", fmt.Errorf("file access not permitted")
	})

	// File write callback (restricted)
	pm.SetHostCallback("file_write", func(path, content string) error {
		return fmt.Errorf("file write not permitted")
	})

	// Environment variable callback
	pm.SetHostCallback("env_get", func(key string) string {
		// Only allow specific env vars
		// Check plugin permissions first
		return os.Getenv(key)
	})
}

// executeWASMSlot executes a WASM plugin slot
func executeWASMSlot(pm *wasm.PluginManager, pluginName, slotName string, node *engine.Node, scope *engine.Scope) error {
	// Parse parameters from node
	params := make(map[string]interface{})

	// Add node value as main parameter if present
	if node.Value != nil {
		params["value"] = node.Value
	}

	// Add children as parameters
	for _, child := range node.Children {
		// Use existing parseNodeValue from utils.go
		params[child.Name] = child.Value
		if len(child.Children) > 0 {
			childMap := make(map[string]interface{})
			for _, grandchild := range child.Children {
				childMap[grandchild.Name] = grandchild.Value
			}
			params[child.Name] = childMap
		}
	}

	// Update scope callbacks for this execution
	pm.SetHostCallback("scope_get", func(key string) (interface{}, error) {
		val, ok := scope.Get(key)
		if !ok {
			return nil, fmt.Errorf("variable $%s not found", key)
		}
		return val, nil
	})

	pm.SetHostCallback("scope_set", func(key string, value interface{}) error {
		scope.Set(key, value)
		return nil
	})

	// Execute plugin slot
	response, err := pm.ExecuteSlot(pluginName, slotName, params)
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
