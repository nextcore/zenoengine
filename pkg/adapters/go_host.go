package adapters

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"zeno/pkg/dbmanager"
	"zeno/pkg/engine"
)

// GoHostAdapter implements HostInterface using Go standard libraries.
type GoHostAdapter struct {
	dbMgr *dbmanager.DBManager
}

// NewGoHostAdapter creates a new adapter.
// dbMgr can be nil if DB access is not required or not available initially.
func NewGoHostAdapter(dbMgr *dbmanager.DBManager) *GoHostAdapter {
	return &GoHostAdapter{
		dbMgr: dbMgr,
	}
}

// --- System ---

func (h *GoHostAdapter) Log(level string, message string) {
	fmt.Printf("[%s] %s\n", level, message)
}

// --- Database ---

func (h *GoHostAdapter) DBQuery(ctx context.Context, dbName string, query string, args []interface{}) (engine.Rows, error) {
	if h.dbMgr == nil {
		return nil, fmt.Errorf("host: database manager not initialized")
	}

	db := h.dbMgr.GetConnection(dbName)
	if db == nil {
		if dbName == "" {
			dbName = "default"
			db = h.dbMgr.GetConnection(dbName)
		}
	}
	if db == nil {
		return nil, fmt.Errorf("host: database '%s' not found", dbName)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return rows, nil
}

func (h *GoHostAdapter) DBExecute(ctx context.Context, dbName string, query string, args []interface{}) (engine.Result, error) {
	if h.dbMgr == nil {
		return nil, fmt.Errorf("host: database manager not initialized")
	}

	db := h.dbMgr.GetConnection(dbName)
	if db == nil {
		if dbName == "" {
			dbName = "default"
			db = h.dbMgr.GetConnection(dbName)
		}
	}
	if db == nil {
		return nil, fmt.Errorf("host: database '%s' not found", dbName)
	}

	res, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	return res, nil
}

// --- HTTP ---

func (h *GoHostAdapter) HTTPSendResponse(ctx context.Context, status int, contentType string, body []byte) error {
	w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
	if !ok {
		return fmt.Errorf("host: not in http context")
	}
	fmt.Printf("DEBUG HOST: HTTPSendResponse status=%d type=%s bodyLen=%d body='%s'\n", status, contentType, len(body), string(body))
	w.Header().Set("Content-Type", contentType)
	w.WriteHeader(status)
	w.Write(body)
	return nil
}

func (h *GoHostAdapter) HTTPGetHeader(ctx context.Context, key string) string {
	r, ok := ctx.Value("httpRequest").(*http.Request)
	if !ok {
		return ""
	}
	return r.Header.Get(key)
}

func (h *GoHostAdapter) HTTPGetQuery(ctx context.Context, key string) string {
	r, ok := ctx.Value("httpRequest").(*http.Request)
	if !ok {
		return ""
	}
	return r.URL.Query().Get(key)
}

func (h *GoHostAdapter) HTTPGetBody(ctx context.Context) ([]byte, error) {
	r, ok := ctx.Value("httpRequest").(*http.Request)
	if !ok {
		return nil, fmt.Errorf("host: not in http context")
	}
	return io.ReadAll(r.Body)
}
