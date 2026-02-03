## 2026-02-03 - Unimplemented WAF Body Check
**Vulnerability:** The WAF middleware had an explicit check for request body size but the implementation block for inspecting the body content was empty/commented out, allowing malicious payloads in POST requests to bypass the WAF.
**Learning:** Placeholder code or "TODO" comments in security-critical paths can create a false sense of security if not tracked or implemented immediately.
**Prevention:** Ensure all security features documented or scaffolded are fully implemented or explicitly marked as "DISABLED/UNIMPLEMENTED" in a way that is visible during security audits. Use linter rules to flag empty blocks in security middleware.
