## 2026-02-03 - Unimplemented WAF Body Check
**Vulnerability:** The WAF middleware had an explicit check for request body size but the implementation block for inspecting the body content was empty/commented out, allowing malicious payloads in POST requests to bypass the WAF.
**Learning:** Placeholder code or "TODO" comments in security-critical paths can create a false sense of security if not tracked or implemented immediately.
**Prevention:** Ensure all security features documented or scaffolded are fully implemented or explicitly marked as "DISABLED/UNIMPLEMENTED" in a way that is visible during security audits. Use linter rules to flag empty blocks in security middleware.

## 2026-02-03 - Path Traversal Oracle in SPA Static Hosting
**Vulnerability:** The `http.static` slot in SPA mode (`spa: true`) leaked file existence information outside the root directory. It used `os.Stat` on the resolved path (which could traverse out of root via `..`) *before* checking if the file was inside the root. This created an Oracle: existing external files returned 404 (via `http.FileServer` rejection), while non-existing files returned 200 (serving `index.html`).
**Learning:** Even if the final file server (`http.FileServer`) is secure against traversal, preliminary checks (like `os.Stat`) performed on unsafe paths can introduce side channels or information disclosure vulnerabilities.
**Prevention:** Always validate that a resolved path is within the expected root directory (using `filepath.Rel` or prefix checks) *before* performing any filesystem operations (like `os.Stat` or `os.Open`) on it.

## 2026-02-03 - Unrestricted File Write (RCE Risk)
**Vulnerability:** The `io.file.write` slot allowed writing to any file extension, including `.zl` (ZenoLang source) and `.go` files. This could allow an attacker with filesystem write access (e.g., via an upload feature utilizing this slot) to modify the application's source code, leading to Remote Code Execution (RCE).
**Learning:** General-purpose filesystem APIs in an interpreted language engine must have strict boundaries. Allowing self-modification of source code is almost always a critical vulnerability.
**Prevention:** Implement a blocklist (or allowlist) of file extensions for filesystem write operations. Explicitly forbid writing to source code extensions (`.zl`, `.go`), configuration files (`.env`), and version control directories (`.git`) in production environments. Exception made for `APP_ENV=development` to support tooling like source generators.

## 2026-02-03 - Missing Anti-Bot Mechanism
**Vulnerability:** The application lacked proactive defenses against automated bots, scraping, and brute-force attacks beyond basic rate limiting.
**Learning:** Defense in depth requires distinguishing between human users and automated scripts. IP-based blocking is often insufficient as bots rotate IPs.
**Prevention:** Implemented a "JS Challenge Interstitial" (inspired by SafeLine/Cloudflare). This middleware serves a lightweight HTML page requiring JavaScript execution to solve a challenge before accessing the site. This filters out dumb bots (curl, python requests, simple scrapers) while remaining transparent to legitimate browsers. The feature is toggleable via `BOT_DEFENSE_ENABLED`.

## 2026-02-04 - Missing IP Reputation System
**Gap:** While rate limiting helps, there was no mechanism to permanently or dynamically block specific malicious IPs (e.g., confirmed attackers, botnets) at the application edge.
**Prevention:** Implemented `IPBlocker` middleware and corresponding ZenoLang slots (`sec.block_ip`, `sec.unblock_ip`). This allows both configuration-based blocking (via Env/File) and dynamic runtime blocking (e.g., a login controller banning an IP after N failed attempts).
