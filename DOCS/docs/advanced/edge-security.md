# Edge Security (WAF & Bot Defense)

Unlike traditional frameworks that require an overhanging reverse proxy like NGINX or Cloudflare for application security, ZenoEngine is designed with edge-computing capabilities baked directly into its core router.

## Web Application Firewall (WAF)

ZenoEngine features an ultra-lightweight, built-in Web Application Firewall. It analyzes incoming HTTP requests in real-time before they even reach your ZenoLang code or the database layer.

The WAF protects against common malicious footprints:
- **SQL Injection (SQLi)** checks in Query Strings and Form payloads.
- **Cross-Site Scripting (XSS)** detection.
- **Path Traversal** attacks.
- **Malicious Bot/Scanner Signatures** (e.g., detecting tools like `sqlmap` or `nikto` via User-Agent).

### Configuration

You can enable or disable the WAF easily from your `.env` file:

```env
# Enable WAF at the router level
WAF_ENABLED=true
```

When enabled, malicious payload attempts are immediately dropped with a `403 Forbidden` early in the lifecycle.

## Bot Defense (JS Challenge)

To combat aggressive Layer 7 DDoS attacks, web scrapers, and automated spam nets, ZenoEngine ships with a Cloudflare-style "I Am Under Attack" mode.

### How it Works
When enabled, the engine intercepts the very first visit from a new client IP. Instead of passing the request to your application, it serves a lightweight `503 Service Unavailable` HTML page containing a JavaScript challenge (a loading spinner). The client's browser must evaluate this JavaScript to compute a mathematical payload and submit it back to the server. 

If the browser succeeds, ZenoEngine drops a secure `zeno_bot_token` cookie and allows the request through. Dumb bots and simple `curl` scripts that cannot execute JavaScript will be permanently stuck at the challenge page.

### Configuration

```env
# Enable DDoS Protection / JS Challenge Mode
BOT_DEFENSE_ENABLED=true

# Secret key used to sign the bot validation tokens
BOT_TOKEN_SECRET="your_highly_secure_random_string"
```

> [!IMPORTANT]
> Since this feature requires JavaScript execution and cookie handling, it should **NOT** be enabled on domains strictly serving API endpoints intended for programmatic consumers (like mobile apps). Use this feature for human-facing web applications.
