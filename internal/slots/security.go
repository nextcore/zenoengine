package slots

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"

	"github.com/gorilla/csrf"
	"golang.org/x/crypto/bcrypt"
)

func RegisterSecuritySlots(eng *engine.Engine) {

	// 1. CRYPTO.HASH
	eng.Register("crypto.hash", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Value utama (shorthand): crypto.hash: $password
		input := coerce.ToString(resolveValue(node.Value, scope))
		target := "hash_result"

		for _, c := range node.Children {
			if c.Name == "text" || c.Name == "val" {
				input = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if input == "" {
			return fmt.Errorf("crypto.hash: input text is required")
		}

		bytes, err := bcrypt.GenerateFromPassword([]byte(input), 10)
		if err != nil {
			return err
		}

		scope.Set(target, string(bytes))
		return nil
	}, engine.SlotMeta{
		Description: "Hash a plain-text password using bcrypt (cost: 10).",
		Example:     "crypto.hash: $pass\n  as: $hashed",
		Inputs: map[string]engine.InputMeta{
			"(value)": {Description: "Shorthand input for the password string to hash", Required: false, Type: "string"},
			"text":    {Description: "Alternative parameter to provide the password string", Required: false, Type: "string"},
			"val":     {Description: "Alternative parameter to provide the password string", Required: false, Type: "string"},
			"as":      {Description: "Variable name to store the generated hash result (Default: 'hash_result')", Required: false, Type: "string"},
		},
	})

	// 2. CRYPTO.VERIFY
	eng.Register("crypto.verify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var hash, text string
		target := "verify_result"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "hash" {
				hash = coerce.ToString(val)
			}
			if c.Name == "text" {
				text = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		err := bcrypt.CompareHashAndPassword([]byte(hash), []byte(text))
		isValid := (err == nil)
		scope.Set(target, isValid)
		return nil
	}, engine.SlotMeta{
		Description: "Verify a plain-text password against a bcrypt hash.",
		Example:     "crypto.verify\n  hash: $h\n  text: $p\n  as: $is_valid",
		Inputs: map[string]engine.InputMeta{
			"hash": {Description: "The bcrypt hash string to compare against", Required: true, Type: "string"},
			"text": {Description: "The plain-text password to verify", Required: true, Type: "string"},
			"as":   {Description: "Variable name to store the boolean result (Default: 'verify_result')", Required: false, Type: "string"},
		},
	})

	// 3. SEC.CSRF_TOKEN
	eng.Register("sec.csrf_token", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		reqVal := ctx.Value("httpRequest")
		if reqVal == nil {
			return fmt.Errorf("httpRequest not found in context")
		}
		r := reqVal.(*http.Request)
		token := csrf.Token(r)

		target := "csrf_token"
		if node.Value != nil {
			target = strings.TrimPrefix(coerce.ToString(node.Value), "$")
		}

		scope.Set(target, token)
		return nil
	}, engine.SlotMeta{
		Description: "Retrieve the CSRF token for the current HTTP request context.",
		Example:     "sec.csrf_token: $token",
		Inputs: map[string]engine.InputMeta{
			"(value)": {Description: "Variable name to store the CSRF token (Default: 'csrf_token')", Required: false, Type: "string"},
		},
	})

	// ==========================================
	// ALIASES FOR TEMPLATE COMPATIBILITY
	// ==========================================
	eng.Register("hash.make", eng.Registry["crypto.hash"], eng.Docs["crypto.hash"])
	eng.Register("hash.verify", eng.Registry["crypto.verify"], eng.Docs["crypto.verify"])
}
