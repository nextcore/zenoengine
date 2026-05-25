# Mail & Notifications

Sending transactional emails (welcome emails, password resets, receipts) is essential for modern web applications. ZenoEngine comes with a built-in, lightweight SMTP mail dispatcher module ready to send both plain text and HTML emails out of the box.

## Configuration

Before you can dispatch emails, configure your SMTP server connection in the `.env` file. You can use either the `SMTP_` or the `MAIL_` prefixes.

```env
# Mailer Settings (Prefix Option A)
SMTP_HOST=smtp.mailtrap.io
SMTP_PORT=2525
SMTP_USERNAME=your_username
SMTP_PASSWORD=your_password
SMTP_FROM_ADDRESS="hello@zenoengine.com"
SMTP_FROM_NAME="ZenoEngine Team"

# Prefix Option B (Fully Supported Fallback)
# MAIL_HOST=smtp.mailtrap.io
# MAIL_PORT=2525
# ...
```

---

## Mock Mode (Development Fallback)

If `SMTP_HOST` (or `MAIL_HOST`) is not configured or empty, the mail dispatcher automatically runs in **Mock Mode**:
1. It logs the email dispatch parameters to the system console.
2. It saves the composed email content as a `.txt` file under the `./storage/logs/mail/` directory.

This allows developers to test registration flows or notifications locally without needing an active SMTP server or risking sending real emails accidentally.

---

## Sending Emails

To send an email, use the `mail.send` slot. It accepts the recipient(s), subject line, and body/HTML content.

```zl
mail.send {
  to: 'user@example.com' # Or an array: ['user1@mail.com', 'user2@mail.com']
  subject: 'Welcome to ZenoEngine!'
  body: 'Thank you for registering. Your account is now active.'
  as: $is_sent
}
```

### Parameter Reference

* **`to`** (string or list, **Required**): Recipient email address(es).
* **`subject`** (string, **Required**): Subject line of the email.
* **`body`** (string, Optional): Plain text content.
* **`html`** (string, Optional): HTML formatted email content.
* **`as`** (string, Optional): Variable to store boolean success status (Defaults to `$mail_status`).

---

## HTML & Multipart Emails

You can send plain text, HTML, or both (as a multipart/alternative MIME message).

### Sending HTML-only Emails

```zl
mail.send {
  to: 'john@example.com'
  subject: 'Verification Link'
  html: '<p>Please click <a href="#">here</a> to verify.</p>'
}
```

### Sending Multipart Emails (Plain Text + HTML Fallback)

Providing both `body` and `html` sends a multipart email, allowing email clients that don't support HTML to render the plain text version safely:

```zl
mail.send {
  to: 'john@example.com'
  subject: 'Multipart Update'
  body: 'Please visit our website for the update.'
  html: '<h1>Updates!</h1><p>Please visit our <a href="#">website</a>.</p>'
}
```

### Rendering via Blade templates

For rich transactional templates, render using the Blade view engine first, then pass the HTML variable:

```zl
view.render: 'emails/welcome' {
  user_name: 'John Doe'
  verify_link: 'https://myapp.com/verify/12345'
  as: $welcome_html
}

mail.send {
  to: 'john@example.com'
  subject: 'Welcome Aboard!'
  html: $welcome_html
}
```
