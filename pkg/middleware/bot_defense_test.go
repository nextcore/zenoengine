package middleware

import (
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBotDefense(t *testing.T) {
	// Setup
	handler := BotDefense(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Protected Content"))
	}))

	t.Run("Disabled_By_Default", func(t *testing.T) {
		os.Unsetenv("BOT_DEFENSE_ENABLED")
		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Protected Content", rec.Body.String())
	})

	t.Run("Enabled_No_Token_Serves_Challenge", func(t *testing.T) {
		os.Setenv("BOT_DEFENSE_ENABLED", "true")
		defer os.Unsetenv("BOT_DEFENSE_ENABLED")

		req := httptest.NewRequest("GET", "/", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		// Should serve 503 Service Unavailable (interstitial)
		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
		assert.Contains(t, rec.Body.String(), "Security Check - ZenoEngine")
		assert.Contains(t, rec.Body.String(), "_zeno/challenge")
	})

	t.Run("Enabled_Static_Assets_Bypass", func(t *testing.T) {
		os.Setenv("BOT_DEFENSE_ENABLED", "true")
		defer os.Unsetenv("BOT_DEFENSE_ENABLED")

		req := httptest.NewRequest("GET", "/style.css", nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.Equal(t, "Protected Content", rec.Body.String())
	})

	t.Run("Challenge_Verification_Success", func(t *testing.T) {
		// Test the POST handler directly
		form := strings.NewReader("solution=zeno-shield-active&original_url=/dashboard")
		req := httptest.NewRequest("POST", "/_zeno/challenge", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		BotChallengeHandler(rec, req)

		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/dashboard", rec.Header().Get("Location"))

		// Check cookie
		cookies := rec.Result().Cookies()
		assert.NotEmpty(t, cookies)
		found := false
		for _, c := range cookies {
			if c.Name == "zeno_bot_token" {
				found = true
				break
			}
		}
		assert.True(t, found, "Cookie zeno_bot_token should be set")
	})

	t.Run("Challenge_Verification_OpenRedirect_Prevention", func(t *testing.T) {
		// Attempt Protocol Relative URL
		form := strings.NewReader("solution=zeno-shield-active&original_url=//malicious.com")
		req := httptest.NewRequest("POST", "/_zeno/challenge", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		BotChallengeHandler(rec, req)

		// Should default to root "/"
		assert.Equal(t, http.StatusFound, rec.Code)
		assert.Equal(t, "/", rec.Header().Get("Location"))

		// Attempt Absolute URL
		form2 := strings.NewReader("solution=zeno-shield-active&original_url=http://malicious.com")
		req2 := httptest.NewRequest("POST", "/_zeno/challenge", form2)
		req2.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec2 := httptest.NewRecorder()

		BotChallengeHandler(rec2, req2)
		assert.Equal(t, "/", rec2.Header().Get("Location"))
	})

	t.Run("Challenge_Verification_Fail", func(t *testing.T) {
		form := strings.NewReader("solution=WRONG&original_url=/")
		req := httptest.NewRequest("POST", "/_zeno/challenge", form)
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		rec := httptest.NewRecorder()

		BotChallengeHandler(rec, req)

		assert.Equal(t, http.StatusForbidden, rec.Code)
	})
}
