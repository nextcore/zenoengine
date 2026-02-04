package slots

import (
	"context"
	"fmt"
	"net/smtp"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestMailSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterMailSlots(eng)

	// Mock SendMail
	origSendMail := SendMailFunc
	defer func() { SendMailFunc = origSendMail }()

	var sentAddr string
	var sentFrom string
	var sentTo []string
	var sentMsg []byte

	SendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
		sentAddr = addr
		sentFrom = from
		sentTo = to
		sentMsg = msg
		return nil
	}

	t.Run("mail.send success", func(t *testing.T) {
		scope := engine.NewScope(nil)
		scope.Set("email", "client@example.com")

		node := &engine.Node{
			Name: "mail.send",
			Value: "$email",
			Children: []*engine.Node{
				{Name: "host", Value: "smtp.mailtrap.io"},
				{Name: "port", Value: 2525},
				{Name: "user", Value: "u"},
				{Name: "pass", Value: "p"},
				{Name: "subject", Value: "Test"},
				{Name: "body", Value: "Hello"},
				{Name: "as", Value: "$ok"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		okVal, _ := scope.Get("ok")
		assert.Equal(t, true, okVal)

		assert.Equal(t, "smtp.mailtrap.io:2525", sentAddr)
		assert.Equal(t, "u", sentFrom)
		assert.Equal(t, []string{"client@example.com"}, sentTo)
		assert.Contains(t, string(sentMsg), "Subject: Test")
	})

	t.Run("mail.send missing args", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "mail.send",
			Children: []*engine.Node{
				{Name: "host", Value: "smtp.test"},
				// missing user, to
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "required")
	})

	t.Run("mail.send failure", func(t *testing.T) {
		SendMailFunc = func(addr string, a smtp.Auth, from string, to []string, msg []byte) error {
			return fmt.Errorf("connection refused")
		}

		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "mail.send",
			Children: []*engine.Node{
				{Name: "to", Value: "x"},
				{Name: "host", Value: "x"},
				{Name: "user", Value: "x"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		// Should return error OR set success false?
		// Implementation returns error: "return err"
		assert.Error(t, err)
	})
}
