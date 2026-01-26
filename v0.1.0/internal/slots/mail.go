package slots

import (
	"context"
	"fmt"
	"net/smtp"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterMailSlots(eng *engine.Engine) {

	// MAIL.SEND (SMTP)
	eng.Register("mail.send", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var host, user, pass, to, subject, body string
		var port int = 587 // Default SMTP port
		target := "mail_sent"

		// Support Shorthand: mail.send: "client@example.com" (atau $email)
		if node.Value != nil {
			to = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)

			if c.Name == "host" {
				host = coerce.ToString(val)
			}
			if c.Name == "port" {
				if p, err := coerce.ToInt(val); err == nil {
					port = p
				}
			}
			if c.Name == "user" {
				user = coerce.ToString(val)
			}
			if c.Name == "pass" {
				pass = coerce.ToString(val)
			}
			if c.Name == "to" {
				to = coerce.ToString(val)
			}
			if c.Name == "subject" {
				subject = coerce.ToString(val)
			}
			if c.Name == "body" {
				body = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if host == "" || user == "" || to == "" {
			return fmt.Errorf("mail.send: host, user, and to are required")
		}

		// Setup Authentication
		auth := smtp.PlainAuth("", user, pass, host)

		// Setup Message (Simple text/html)
		msg := []byte(fmt.Sprintf("To: %s\r\n"+
			"Subject: %s\r\n"+
			"Content-Type: text/html; charset=UTF-8\r\n"+
			"\r\n"+
			"%s\r\n", to, subject, body))

		addr := fmt.Sprintf("%s:%d", host, port)

		// Send Email
		err := smtp.SendMail(addr, auth, user, []string{to}, msg)
		if err != nil {
			scope.Set(target, false)
			// Coba tangani error common
			if strings.Contains(err.Error(), "authentication failed") {
				return fmt.Errorf("mail.send: authentication failed, check user/pass")
			}
			return err
		}

		scope.Set(target, true)
		return nil
	}, engine.SlotMeta{
		Description: "Mengirim email via SMTP.",
		Example: `mail.send: $client_email
  host: "smtp.gmail.com"
  port: 587
  user: $smtp_user
  pass: $smtp_pass
  subject: "Invoice"
  body: $html_content
  as: $is_sent`,
	})
}
