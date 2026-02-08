package adapter

import (
	"context"
	"encoding/json"
	"strings"
	"syscall/js"

	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// Auth Constants
const (
	AuthTokenKey = "zeno_auth_token"
	AuthUserKey  = "zeno_auth_user"
)

func RegisterAuthSlots(eng *engine.Engine) {
	// AUTH.LOGIN: Persist user session
	eng.Register("auth.login", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var user interface{}
		var token string

		for _, c := range node.Children {
			if c.Name == "user" {
				user = ResolveValue(c.Value, scope)
			} else if c.Name == "token" {
				token = coerce.ToString(ResolveValue(c.Value, scope))
			}
		}

		if token != "" {
			js.Global().Get("localStorage").Call("setItem", AuthTokenKey, token)
		}

		if user != nil {
			userJson, _ := json.Marshal(user)
			js.Global().Get("localStorage").Call("setItem", AuthUserKey, string(userJson))
		}

		return nil
	}, engine.SlotMeta{Description: "Login user and persist session"})

	// AUTH.LOGOUT: Clear session
	eng.Register("auth.logout", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		js.Global().Get("localStorage").Call("removeItem", AuthTokenKey)
		js.Global().Get("localStorage").Call("removeItem", AuthUserKey)
		return nil
	}, engine.SlotMeta{Description: "Logout user and clear session"})

	// AUTH.USER: Get current user
	eng.Register("auth.user", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		target := "user"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		userStr := js.Global().Get("localStorage").Call("getItem", AuthUserKey)
		if userStr.IsNull() {
			scope.Set(target, nil)
			return nil
		}

		var userData interface{}
		json.Unmarshal([]byte(userStr.String()), &userData)
		scope.Set(target, userData)
		return nil
	}, engine.SlotMeta{Description: "Retrieve logged in user data"})

	// AUTH.CHECK: Boolean check
	eng.Register("auth.check", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		target := "is_auth"
		for _, c := range node.Children {
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		token := js.Global().Get("localStorage").Call("getItem", AuthTokenKey)
		isAuth := !token.IsNull() && token.String() != ""

		scope.Set(target, isAuth)
		return nil
	}, engine.SlotMeta{Description: "Check if user is logged in"})
}

// Helper for Router Middleware to check auth synchronously
func CheckAuth() bool {
	token := js.Global().Get("localStorage").Call("getItem", AuthTokenKey)
	return !token.IsNull() && token.String() != ""
}
