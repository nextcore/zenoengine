package slots

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"github.com/nextcore/zeno-go/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestMailSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterMailSlots(eng)

	// Clean up storage dir before and after tests
	mockMailDir := filepath.Join("storage", "logs", "mail")
	os.RemoveAll(mockMailDir)
	defer os.RemoveAll("storage")

	t.Run("mail.send mock mode (success)", func(t *testing.T) {
		// SMTP_HOST is not set, so it should run in Mock Mode
		os.Unsetenv("SMTP_HOST")

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "mail.send",
			Children: []*engine.Node{
				{Name: "to", Value: "recipient@example.com"},
				{Name: "subject", Value: "Hello Test"},
				{Name: "body", Value: "This is a body"},
				{Name: "html", Value: "<h1>This is html</h1>"},
				{Name: "as", Value: "$sent"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		sent, ok := scope.Get("sent")
		assert.True(t, ok)
		assert.True(t, sent.(bool))

		// Check if mock file was created
		files, err := os.ReadDir(mockMailDir)
		assert.NoError(t, err)
		assert.Len(t, files, 1)

		content, err := os.ReadFile(filepath.Join(mockMailDir, files[0].Name()))
		assert.NoError(t, err)
		assert.Contains(t, string(content), "To: recipient@example.com")
		assert.Contains(t, string(content), "Subject: Hello Test")
		assert.Contains(t, string(content), "This is a body")
		assert.Contains(t, string(content), "<h1>This is html</h1>")
	})

	t.Run("mail.send missing recipient", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "mail.send",
			Children: []*engine.Node{
				{Name: "subject", Value: "No recipient"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required attribute 'to'")
	})

	t.Run("mail.send missing subject", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "mail.send",
			Children: []*engine.Node{
				{Name: "to", Value: "test@example.com"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "missing required attribute 'subject'")
	})
}
