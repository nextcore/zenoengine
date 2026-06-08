package slots

import (
	"context"
	"fmt"
	"net/smtp"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"
)

func RegisterMailSlots(eng *engine.Engine) {

	// 1. MAIL.SEND
	eng.Register("mail.send", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var to []string
		var subject, body, html string
		target := "mail_status"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "to" {
				if str, ok := val.(string); ok {
					to = []string{str}
				} else if list, ok := val.([]interface{}); ok {
					for _, item := range list {
						to = append(to, coerce.ToString(item))
					}
				} else if strList, ok := val.([]string); ok {
					to = strList
				}
			}
			if c.Name == "subject" {
				subject = coerce.ToString(val)
			}
			if c.Name == "body" {
				body = coerce.ToString(val)
			}
			if c.Name == "html" {
				html = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if len(to) == 0 {
			return fmt.Errorf("mail.send: to recipient is required")
		}
		if subject == "" {
			return fmt.Errorf("mail.send: subject is required")
		}

		// Read SMTP Configuration (supporting both SMTP_* and MAIL_* prefixes)
		smtpHost := os.Getenv("SMTP_HOST")
		if smtpHost == "" {
			smtpHost = os.Getenv("MAIL_HOST")
		}
		smtpPort := os.Getenv("SMTP_PORT")
		if smtpPort == "" {
			smtpPort = os.Getenv("MAIL_PORT")
		}
		smtpUser := os.Getenv("SMTP_USERNAME")
		if smtpUser == "" {
			smtpUser = os.Getenv("MAIL_USERNAME")
		}
		smtpPass := os.Getenv("SMTP_PASSWORD")
		if smtpPass == "" {
			smtpPass = os.Getenv("MAIL_PASSWORD")
		}
		fromAddr := os.Getenv("SMTP_FROM_ADDRESS")
		if fromAddr == "" {
			fromAddr = os.Getenv("MAIL_FROM_ADDRESS")
		}
		fromName := os.Getenv("SMTP_FROM_NAME")
		if fromName == "" {
			fromName = os.Getenv("MAIL_FROM_NAME")
		}

		if fromAddr == "" {
			fromAddr = "noreply@zenoengine.local"
		}

		// Compose Email Message (MIME format)
		boundary := "ZenoBoundary1234567890"
		msg := ""
		if fromName != "" {
			msg += fmt.Sprintf("From: %s <%s>\r\n", fromName, fromAddr)
		} else {
			msg += fmt.Sprintf("From: %s\r\n", fromAddr)
		}
		msg += fmt.Sprintf("To: %s\r\n", strings.Join(to, ", "))
		msg += fmt.Sprintf("Subject: %s\r\n", subject)
		msg += "MIME-Version: 1.0\r\n"
		
		if html != "" && body != "" {
			// Multipart alternative (both Plain Text and HTML)
			msg += fmt.Sprintf("Content-Type: multipart/alternative; boundary=%s\r\n\r\n", boundary)
			
			// Plain text part
			msg += fmt.Sprintf("--%s\r\n", boundary)
			msg += "Content-Type: text/plain; charset=utf-8\r\n\r\n"
			msg += body + "\r\n"
			
			// HTML part
			msg += fmt.Sprintf("--%s\r\n", boundary)
			msg += "Content-Type: text/html; charset=utf-8\r\n\r\n"
			msg += html + "\r\n"
			
			msg += fmt.Sprintf("--%s--\r\n", boundary)
		} else if html != "" {
			// HTML only
			msg += "Content-Type: text/html; charset=utf-8\r\n\r\n"
			msg += html + "\r\n"
		} else {
			// Plain text only
			msg += "Content-Type: text/plain; charset=utf-8\r\n\r\n"
			msg += body + "\r\n"
		}

		// Check if SMTP_HOST is provided. If not, run in Mock Mode.
		if smtpHost == "" {
			// Mock Mode
			mockDir := filepath.Join("storage", "logs", "mail")
			os.MkdirAll(mockDir, 0755)
			mockFile := filepath.Join(mockDir, fmt.Sprintf("%d_mail.txt", time.Now().UnixNano()))
			
			mockContent := fmt.Sprintf("=== MOCK EMAIL SENT ===\nSMTP Config: [Host: %s, Port: %s, User: %s]\n\n%s\n=======================", smtpHost, smtpPort, smtpUser, msg)
			_ = os.WriteFile(mockFile, []byte(mockContent), 0644)
			
			// Log to output
			fmt.Printf("[MAIL MOCK] Simulated sending email to %s. Saved to %s\n", strings.Join(to, ", "), mockFile)
			scope.Set(target, true)
			return nil
		}

		// SMTP Authentication
		var auth smtp.Auth
		if smtpUser != "" || smtpPass != "" {
			auth = smtp.PlainAuth("", smtpUser, smtpPass, smtpHost)
		}

		// Send Email
		addr := fmt.Sprintf("%s:%s", smtpHost, smtpPort)
		err := smtp.SendMail(addr, auth, fromAddr, to, []byte(msg))
		if err != nil {
			scope.Set(target, false)
			return fmt.Errorf("mail.send: failed to send email via SMTP: %v", err)
		}

		scope.Set(target, true)
		return nil
	}, engine.SlotMeta{
		Description: "Send email natively via SMTP or in Mock Mode if no SMTP_HOST is configured.",
		Example:     "mail.send:\n  to: 'user@example.com'\n  subject: 'Welcome'\n  body: 'Hello User'\n  as: $sent",
		Inputs: map[string]engine.InputMeta{
			"to":      {Description: "Recipient email address (string or list of strings)", Required: true, Type: "string/list"},
			"subject": {Description: "Subject line of the email", Required: true, Type: "string"},
			"body":    {Description: "Plain text email body", Required: false, Type: "string"},
			"html":    {Description: "HTML email body", Required: false, Type: "string"},
			"as":      {Description: "Variable name to store the boolean send status (Default: 'mail_status')", Required: false, Type: "string"},
		},
	})
}
