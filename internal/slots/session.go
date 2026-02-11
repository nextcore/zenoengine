package slots

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// FlashSessionKeyPrefix is the prefix for flash cookies
const FlashSessionKeyPrefix = "_flash_"

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

		// URL Encode to be safe in cookie
		cookieVal := url.QueryEscape(string(jsonBytes))

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

		// If we can't write (read-only context?), we can still read cookie but can't delete it.
		// Detailed logic: Read cookie, then Set-Cookie MaxAge=-1 to delete it.

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
			// Not found
			scope.Set(target, nil)
			return nil
		}

		// Decode value
		jsonStr, err := url.QueryUnescape(cookie.Value)
		if err != nil {
			scope.Set(target, nil)
			return nil
		}

		var val interface{}
		if err := json.Unmarshal([]byte(jsonStr), &val); err != nil {
			// If not valid JSON, maybe raw string?
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

	// 3. SESSION.KEEP - Keep flash data for another request (Not implemented yet - re-set cookie?)
}
