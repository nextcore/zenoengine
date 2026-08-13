package slots

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strings"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"
)

// FlashSessionKeyPrefix is the prefix for flash cookies
const FlashSessionKeyPrefix = "_flash_"

// Helper to get APP_KEY for signing
func getAppKey() []byte {
	key := os.Getenv("APP_KEY")
	if key == "" {
		key = "zenoengine_default_secret_key_change_me_in_production"
	}
	return []byte(key)
}

// Generate HMAC signature for value
func signCookieValue(value string) string {
	mac := hmac.New(sha256.New, getAppKey())
	mac.Write([]byte(value))
	sig := hex.EncodeToString(mac.Sum(nil))
	return value + "|" + sig
}

// Verify HMAC signature of cookie value, returns clean value or empty string if invalid
func verifyCookieValue(signedValue string) string {
	parts := strings.SplitN(signedValue, "|", 2)
	if len(parts) != 2 {
		return ""
	}
	val := parts[0]
	sig := parts[1]

	mac := hmac.New(sha256.New, getAppKey())
	mac.Write([]byte(val))
	expectedSig := hex.EncodeToString(mac.Sum(nil))

	if hmac.Equal([]byte(sig), []byte(expectedSig)) {
		return val
	}
	return ""
}

// RegisterSessionSlots registers session related slots
func RegisterSessionSlots(eng *engine.Engine) {

	// 1. SESSION.FLASH - Store data for next request (via Cookie)
	eng.Register("session.flash", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("session.flash: not in http context")
		}

		var key string
		var val interface{}

		// Parse arguments
		if node.Value != nil {
			val = resolveValue(node.Value, scope)
		}

		for _, c := range node.Children {
			if c.Name == "key" {
				key = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "val" || c.Name == "value" {
				val = parseNodeValue(c, scope)
			}
		}

		if key == "" {
			return fmt.Errorf("session.flash: key is required")
		}

		// Encode value to JSON string
		jsonBytes, err := json.Marshal(val)
		if err != nil {
			return fmt.Errorf("session.flash: failed to marshal value: %v", err)
		}

		// Sign cookie value to prevent tampering
		signedVal := signCookieValue(string(jsonBytes))
		cookieVal := url.QueryEscape(signedVal)

		// Set Cookie (Short lived, e.g. 5 minutes to allow redirect)
		http.SetCookie(w, &http.Cookie{
			Name:     FlashSessionKeyPrefix + key,
			Value:    cookieVal,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   300,
		})

		return nil
	}, engine.SlotMeta{
		Description: "Flash data to the session (cookie) for the next request.",
		Example:     "session.flash: { key: 'error', val: 'Invalid credentials' }",
	})

	// 2. SESSION.GET_FLASH - Retrieve and delete flash data
	eng.Register("session.get_flash", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return nil
		}
		r := reqVal.(*http.Request)
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)

		var key string
		target := "flash_data"

		// Parse arguments
		if node.Value != nil {
			key = coerce.ToString(resolveValue(node.Value, scope))
			target = key // Default target same as key name if shorthand used
		}

		for _, c := range node.Children {
			if c.Name == "key" {
				key = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if key == "" {
			return fmt.Errorf("session.get_flash: key is required")
		}

		cookieName := FlashSessionKeyPrefix + key
		cookie, err := r.Cookie(cookieName)

		if err != nil || cookie.Value == "" {
			scope.Set(target, nil)
			return nil
		}

		// Decode value
		escapedStr, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			scope.Set(target, nil)
			return nil
		}

		// Verify signature
		jsonStr := verifyCookieValue(escapedStr)
		if jsonStr == "" {
			scope.Set(target, nil)
			return nil
		}

		var val interface{}
		if err := json.Unmarshal([]byte(jsonStr), &val); err != nil {
			val = jsonStr
		}

		scope.Set(target, val)

		// Delete Cookie (Flash is read-once)
		if ok {
			http.SetCookie(w, &http.Cookie{
				Name:     cookieName,
				Value:    "",
				Path:     "/",
				HttpOnly: true,
				MaxAge:   -1,
			})
		}

		return nil
	}, engine.SlotMeta{
		Description: "Retrieve flash data and remove it from session.",
		Example:     "session.get_flash: 'error' { as: $error_msg }",
	})

	// 3. SESSION.SET
	eng.Register("session.set", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		if !ok {
			return fmt.Errorf("session.set: missing context")
		}
		key := coerce.ToString(resolveValue(node.Value, scope))
		var val interface{}
		for _, c := range node.Children {
			if c.Name == "val" || c.Name == "value" {
				val = parseNodeValue(c, scope)
			}
		}

		jsonBytes, _ := json.Marshal(val)
		signedVal := signCookieValue(string(jsonBytes))
		cookieVal := url.QueryEscape(signedVal)

		http.SetCookie(w, &http.Cookie{
			Name:     "_session_" + key,
			Value:    cookieVal,
			Path:     "/",
			HttpOnly: true,
			MaxAge:   86400 * 7, // 1 week
		})
		return nil
	}, engine.SlotMeta{Description: "Set session data."})

	// 4. SESSION.GET
	eng.Register("session.get", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return nil
		}
		r := reqVal.(*http.Request)
		key := coerce.ToString(resolveValue(node.Value, scope))
		target := key
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		cookie, err := r.Cookie("_session_" + key)
		if err != nil {
			scope.Set(target, nil)
			return nil
		}

		escapedStr, _ := url.QueryUnescape(cookie.Value)
		jsonStr := verifyCookieValue(escapedStr)
		if jsonStr == "" {
			// Signature failed, treat as nil/unauthorized session
			scope.Set(target, nil)
			return nil
		}

		var val interface{}
		json.Unmarshal([]byte(jsonStr), &val)
		scope.Set(target, val)
		return nil
	}, engine.SlotMeta{Description: "Get session data."})

	// 5. SESSION.DESTROY
	eng.Register("session.destroy", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		w, ok := ctx.Value("httpWriter").(http.ResponseWriter)
		r, ok2 := ctx.Value("httpRequest").(*http.Request)
		if !ok || !ok2 {
			return nil
		}

		// Clear all session and flash cookies
		for _, cookie := range r.Cookies() {
			if strings.HasPrefix(cookie.Name, "_session_") || strings.HasPrefix(cookie.Name, FlashSessionKeyPrefix) {
				http.SetCookie(w, &http.Cookie{
					Name:   cookie.Name,
					Value:  "",
					Path:   "/",
					MaxAge: -1,
				})
			}
		}
		return nil
	}, engine.SlotMeta{Description: "Destroy all session data."})

	// 6. SESSION.REGENERATE
	eng.Register("session.regenerate", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		return nil
	}, engine.SlotMeta{Description: "Regenerate session ID (Security)."})
}
