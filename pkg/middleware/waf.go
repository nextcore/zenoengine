package middleware

import (
	"log/slog"
	"net/http"
	"os"
	"regexp"
	"strings"
)

var (
	// SQL Injection Patterns
	sqlPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(UNION|SELECT|INSERT|UPDATE|DELETE|DROP|ALTER|TRUNCATE)\s+(ALL|DISTINCT)?.*(FROM|INTO|TABLE|WHERE)`),
		regexp.MustCompile(`(?i)--|#|/\*`),
		regexp.MustCompile(`(?i)OR\s+.+=\s*.+`),
		regexp.MustCompile(`(?i)SLEEP\(\d+\)`),
	}

	// XSS Patterns
	xssPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)<script.*?>.*?</script.*?>`),
		regexp.MustCompile(`(?i)on\w+\s*=\s*".*?"`),
		regexp.MustCompile(`(?i)javascript\s*:\s*`),
		regexp.MustCompile(`(?i)<iframe.*?>`),
	}

	// Path Traversal
	traversalPatterns = []*regexp.Regexp{
		regexp.MustCompile(`\.\.\/`),
		regexp.MustCompile(`\.\.\\`),
		regexp.MustCompile(`(?i)/etc/passwd`),
	}

	// Malicious User Agents
	badUserAgents = []string{
		"sqlmap", "nikto", "dirbuster", "nmap", "gobuster",
	}
)

// WAF is a lightweight Web Application Firewall middleware
func WAF(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Only run if WAF_ENABLED=true
		if os.Getenv("WAF_ENABLED") != "true" {
			next.ServeHTTP(w, r)
			return
		}

		// 1. Check User Agent
		ua := strings.ToLower(r.UserAgent())
		for _, badUA := range badUserAgents {
			if strings.Contains(ua, badUA) {
				blockRequest(w, r, "Malicious User Agent Detected")
				return
			}
		}

		// 2. Check Query Parameters
		queryParams := r.URL.RawQuery
		if detectMaliciousContent(queryParams) {
			blockRequest(w, r, "Malicious Query Parameter Detected")
			return
		}

		// 3. Check Body (for form data / small payloads)
		// Note: We don't read the whole body for large files to avoid performance hit
		if r.ContentLength > 0 && r.ContentLength < 1024*1024 { // < 1MB
			// Peek logic or full read could be here, but for now we focus on URL/Headers
			// for performance reasons. In a real WAF, we might use a TeeReader.
		}

		next.ServeHTTP(w, r)
	})
}

func detectMaliciousContent(input string) bool {
	if input == "" {
		return false
	}

	// Check SQLi
	for _, p := range sqlPatterns {
		if p.MatchString(input) {
			return true
		}
	}

	// Check XSS
	for _, p := range xssPatterns {
		if p.MatchString(input) {
			return true
		}
	}

	// Check Traversal
	for _, p := range traversalPatterns {
		if p.MatchString(input) {
			return true
		}
	}

	return false
}

func blockRequest(w http.ResponseWriter, r *http.Request, reason string) {
	slog.Warn("🛡️ WAF BLOCKED REQUEST",
		"ip", r.RemoteAddr,
		"method", r.Method,
		"path", r.URL.Path,
		"reason", reason,
		"ua", r.UserAgent())

	w.WriteHeader(http.StatusForbidden)
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"success": false, "error": "Blocked by WAF", "reason": "` + reason + `"}`))
}
