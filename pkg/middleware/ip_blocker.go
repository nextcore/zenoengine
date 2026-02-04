package middleware

import (
	"bufio"
	"log/slog"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
)

// IPBlockList manages blocked IPs safely
type IPBlockList struct {
	mu      sync.RWMutex
	blocked map[string]bool
}

var GlobalBlockList = &IPBlockList{
	blocked: make(map[string]bool),
}

// LoadBlockList initializes the blocklist from Env and File
func (b *IPBlockList) Load() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// 1. Load from Env
	envIPs := os.Getenv("BLOCKED_IPS")
	if envIPs != "" {
		for _, ip := range strings.Split(envIPs, ",") {
			ip = strings.TrimSpace(ip)
			if ip != "" {
				b.blocked[ip] = true
			}
		}
	}

	// 2. Load from File (data/ip_blocklist.txt)
	file, err := os.Open("data/ip_blocklist.txt")
	if err == nil {
		defer file.Close()
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			ip := strings.TrimSpace(scanner.Text())
			if ip != "" && !strings.HasPrefix(ip, "#") {
				b.blocked[ip] = true
			}
		}
	}
}

// IsBlocked checks if an IP is in the blocklist
func (b *IPBlockList) IsBlocked(ip string) bool {
	b.mu.RLock()
	defer b.mu.RUnlock()
	return b.blocked[ip]
}

// Add blocks an IP dynamically
func (b *IPBlockList) Add(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	b.blocked[ip] = true
	slog.Warn("🚫 IP Blocked Dynamically", "ip", ip)
}

// Remove unblocks an IP
func (b *IPBlockList) Remove(ip string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	delete(b.blocked, ip)
	slog.Info("✅ IP Unblocked", "ip", ip)
}

// IPBlocker Middleware
func IPBlocker(next http.Handler) http.Handler {
	// Initialize once
	GlobalBlockList.Load()

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Get real IP (handle proxies if needed, but for now RemoteAddr)
		// Note: In production behind Cloudflare/Nginx, you need a helper to extract X-Forwarded-For
		// But Zeno's HostDispatcher suggests it might be direct or proxied.
		// For security, blindly trusting headers is bad unless configured.
		// We'll use RemoteAddr (Host:Port) for now and strip port.

		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr // Fallback if no port
		}

		if GlobalBlockList.IsBlocked(ip) {
			slog.Warn("🚫 Request Blocked (Blacklisted IP)", "ip", ip, "path", r.URL.Path)
			w.WriteHeader(http.StatusForbidden)
			w.Write([]byte(`{"success":false,"error":"Access Denied (IP Blocked)"}`))
			return
		}

		next.ServeHTTP(w, r)
	})
}
