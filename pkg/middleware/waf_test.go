package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWAF_BodyCheck(t *testing.T) {
	// Enable WAF
	os.Setenv("WAF_ENABLED", "true")
	defer os.Unsetenv("WAF_ENABLED")

	tests := []struct {
		name           string
		body           string
		expectedStatus int
	}{
		{
			name:           "Safe Body",
			body:           `{"key": "value"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "SQL Injection in Body",
			body:           `{"key": "UNION SELECT * FROM users"}`,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "XSS in Body",
			body:           `{"key": "<script>alert(1)</script>"}`,
			expectedStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()

			handler := WAF(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusOK)
			}))

			handler.ServeHTTP(rec, req)

			assert.Equal(t, tt.expectedStatus, rec.Code, "Case: %s", tt.name)
		})
	}
}
