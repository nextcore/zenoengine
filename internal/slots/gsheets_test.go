package slots

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"google.golang.org/api/option"
	"google.golang.org/api/sheets/v4"
)

func TestGSheetSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterGSheetSlots(eng)

	// Mock Google API Server
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")

		// Respond to Get Values
		if r.Method == "GET" {
			// Response structure for ValueRange
			// https://developers.google.com/sheets/api/reference/rest/v4/spreadsheets.values/get
			resp := map[string]interface{}{
				"range": "Sheet1!A1:B2",
				"majorDimension": "ROWS",
				"values": [][]interface{}{
					{"Name", "Age"},
					{"Alice", 30},
				},
			}
			json.NewEncoder(w).Encode(resp)
			return
		}

		// Respond to Append/Update/Clear
		if r.Method == "POST" || r.Method == "PUT" {
			resp := map[string]interface{}{
				"spreadsheetId": "123",
				"updatedRange": "Sheet1!A1",
			}
			json.NewEncoder(w).Encode(resp)
			return
		}
	}))
	defer ts.Close()

	// Mock Service Creator
	origFunc := GetSheetsServiceFunc
	defer func() { GetSheetsServiceFunc = origFunc }()

	GetSheetsServiceFunc = func(ctx context.Context, creds interface{}) (*sheets.Service, error) {
		// Create service pointing to mock server
		return sheets.NewService(ctx,
			option.WithEndpoint(ts.URL),
			option.WithHTTPClient(ts.Client()),
			option.WithoutAuthentication(), // Important for testing
		)
	}

	t.Run("gsheet.get", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "gsheet.get",
			Children: []*engine.Node{
				{Name: "id", Value: "123"},
				{Name: "range", Value: "Sheet1!A1:B2"},
				{Name: "credentials", Value: "fake.json"}, // Required by slot check
				{Name: "as", Value: "$data"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		dataRaw, ok := scope.Get("data")
		assert.True(t, ok)
		data := dataRaw.([][]interface{})
		assert.Equal(t, 2, len(data))
		assert.Equal(t, "Name", data[0][0])
		assert.Equal(t, "Alice", data[1][0])
	})
}
