# Mail & Notifications

Sending automatic transactional emails (welcome emails, password resets, receipts) is essential for modern web applications. ZenoEngine comes with a built-in, lightweight SMTP mail dispatcher module ready to send both plain text and HTML emails out of the box.

## Configuration

Before you can dispatch emails, configure your SMTP server connection in the `.env` file. Do not hardcode credentials in your `.zl` scripts.

```env
# Mailer Settings
MAIL_HOST=smtp.mailtrap.io
MAIL_PORT=2525
MAIL_USERNAME=your_username
MAIL_PASSWORD=your_password
MAIL_FROM_ADDRESS="hello@zenoengine.com"
MAIL_FROM_NAME="ZenoEngine Team"
```

## Sending Emails

To send an email, use the `mail.send` slot. It accepts several configuration parameters including the recipient array, subject line, and the body content.

```zeno
// In your controller or logic block

mail.send: {
    to: ['user@example.com', 'admin@example.com']
    subject: "Welcome to ZenoEngine!"
    body: "Thank you for registering. Your account is now active."
    as: $mailResult
    error: $mailError
}

if: $mailError != null {
    // Handle SMTP failure
    log.error: "Failed to send welcome email: " + $mailError
    http.response: { status: 500, json: { error: "Email delivery failed" } }
    return
}

http.json: { success: true, message: "Email dispatched successfully!" }
```

## HTML Emails

While plain text is sufficient for simple notifications, you will often want to send rich HTML emails. The `mail.send` slot automatically detects if the `body` string contains HTML tags and sends the email appropriately with the `text/html` Content-Type.

For best results, you should render your email content using the Blade engine before passing it to the mailer:

```zeno
// 1. Render the blade template (resources/views/emails/welcome.blade.zl)
view.render: 'emails/welcome' {
    user_name: "John Doe",
    verify_link: "https://myapp.com/verify/12345"
    as: $htmlContent
}

// 2. Dispatch the rendered HTML
mail.send: {
    to: ['john@example.com']
    subject: "Welcome Aboard, John!"
    body: $htmlContent
}
```

```html
<!-- resources/views/emails/welcome.blade.zl -->
<!DOCTYPE html>
<html>
<body>
    <h1>Welcome, {{ $user_name }}!</h1>
    <p>We are thrilled to have you here.</p>
    <a href="{{ $verify_link }}" style="padding: 10px; background: blue; color: white;">
        Verify Your Account
    </a>
</body>
</html>
```

## Attachments

The ZenoEngine mailer will support file attachments in an upcoming release. For now, it is specialized in rapid, direct notification dispatches.
