package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestValidatorSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterValidatorSlots(eng, nil)

	t.Run("validate required field success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		data := map[string]interface{}{
			"username": "john_doe",
		}
		scope.Set("input", data)

		node := &engine.Node{
			Name:  "validator.validate",
			Value: "$input",
			Children: []*engine.Node{
				{
					Name: "rules",
					Value: map[string]interface{}{
						"username": "required",
					},
				},
				{Name: "as", Value: "$errors"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		require.NoError(t, err)

		errs, _ := scope.Get("errors")
		assert.Nil(t, errs, "Errors should be nil on success")
	})

	t.Run("validate required field failure", func(t *testing.T) {
		scope := engine.NewScope(nil)
		data := map[string]interface{}{
			"username": "",
		}
		scope.Set("input", data)

		node := &engine.Node{
			Name:  "validator.validate",
			Value: "$input",
			Children: []*engine.Node{
				{
					Name: "rules",
					Value: map[string]interface{}{
						"username": "required",
					},
				},
				{Name: "as", Value: "$errors"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		require.NoError(t, err)

		errsVal, _ := scope.Get("errors")
		errs := errsVal.(map[string]string)
		assert.Contains(t, errs, "username")
		assert.Equal(t, "username is required", errs["username"])
	})

	t.Run("validate email format", func(t *testing.T) {
		scope := engine.NewScope(nil)
		data := map[string]interface{}{
			"email_ok":  "test@example.com",
			"email_bad": "not-an-email",
		}
		scope.Set("input", data)

		node := &engine.Node{
			Name:  "validator.validate",
			Value: "$input",
			Children: []*engine.Node{
				{
					Name: "rules",
					Value: map[string]interface{}{
						"email_ok":  "email",
						"email_bad": "email",
					},
				},
				{Name: "as", Value: "$errors"},
			},
		}

		eng.Execute(context.Background(), node, scope)
		errsVal, _ := scope.Get("errors")
		errs := errsVal.(map[string]string)

		assert.NotContains(t, errs, "email_ok")
		assert.Contains(t, errs, "email_bad")
	})

	t.Run("validate min max rules", func(t *testing.T) {
		scope := engine.NewScope(nil)
		data := map[string]interface{}{
			"age_low":   "10",
			"age_high":  "100",
			"txt_short": "hi",
			"txt_long":  "hello world",
		}
		scope.Set("input", data)

		node := &engine.Node{
			Name:  "validator.validate",
			Value: "$input",
			Children: []*engine.Node{
				{
					Name: "rules",
					Value: map[string]interface{}{
						"age_low":   "numeric|min:18",
						"age_high":  "numeric|max:50",
						"txt_short": "min:5",
						"txt_long":  "max:5",
					},
				},
				{Name: "as", Value: "$errors"},
			},
		}

		eng.Execute(context.Background(), node, scope)
		errsVal, _ := scope.Get("errors")
		errs := errsVal.(map[string]string)

		assert.Contains(t, errs, "age_low")
		assert.Contains(t, errs, "age_high")
		assert.Contains(t, errs, "txt_short")
		assert.Contains(t, errs, "txt_long")
	})

	t.Run("validate strict filtering (whitelist)", func(t *testing.T) {
		scope := engine.NewScope(nil)
		data := map[string]interface{}{
			"username": "admin",
			"email":    "admin@example.com",
			"is_admin": true,    // Should be filtered out
			"role":     "super", // Should be filtered out
		}
		scope.Set("input", data)

		node := &engine.Node{
			Name:  "validator.validate",
			Value: "$input",
			Children: []*engine.Node{
				{
					Name: "rules",
					Value: map[string]interface{}{
						"username": "required",
						"email":    "required|email",
					},
				},
				{Name: "as", Value: "$errors"},
				{Name: "as_safe", Value: "$clean_data"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		require.NoError(t, err)

		// Check Errors (Should be nil)
		errs, _ := scope.Get("errors")
		assert.Nil(t, errs)

		// Check Safe Data
		cleanDataVal, ok := scope.Get("clean_data")
		require.True(t, ok, "clean_data should be set")

		cleanData := cleanDataVal.(map[string]interface{})
		assert.Equal(t, "admin", cleanData["username"])
		assert.Equal(t, "admin@example.com", cleanData["email"])

		// Ensure dangerous fields are removed
		_, hasAdmin := cleanData["is_admin"]
		assert.False(t, hasAdmin, "is_admin should be removed")
		_, hasRole := cleanData["role"]
		assert.False(t, hasRole, "role should be removed")
	})
}
