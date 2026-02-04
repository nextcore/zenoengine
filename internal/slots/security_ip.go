package slots

import (
	"context"
	"fmt"
	"net"
	"net/http"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/middleware"
	"zeno/pkg/utils/coerce"
)

func RegisterIPSecuritySlots(eng *engine.Engine) {

	// 1. SEC.BLOCK_IP
	eng.Register("sec.block_ip", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var ip string

		// Value can be direct: sec.block_ip: "192.168.1.1"
		if node.Value != nil {
			val := resolveValue(node.Value, scope)
			ip = coerce.ToString(val)
		}

		// Or via children: sec.block_ip: { ip: "..." }
		for _, c := range node.Children {
			if c.Name == "ip" {
				val := parseNodeValue(c, scope)
				ip = coerce.ToString(val)
			}
		}

		// If no IP provided, try to block current request IP
		if ip == "" {
			reqVal := ctx.Value("httpRequest")
			if reqVal != nil {
				r := reqVal.(*http.Request)
				host, _, err := net.SplitHostPort(r.RemoteAddr)
				if err == nil {
					ip = host
				} else {
					ip = r.RemoteAddr
				}
			}
		}

		if ip == "" {
			return fmt.Errorf("sec.block_ip: IP address required")
		}

		// Block the IP
		middleware.GlobalBlockList.Add(ip)
		return nil
	}, engine.SlotMeta{
		Description: "Add an IP address to the global blocklist.",
		Example:     "sec.block_ip: $request.ip",
	})

	// 2. SEC.UNBLOCK_IP
	eng.Register("sec.unblock_ip", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var ip string
		if node.Value != nil {
			val := resolveValue(node.Value, scope)
			ip = coerce.ToString(val)
		}

		if ip == "" {
			for _, c := range node.Children {
				if c.Name == "ip" {
					ip = coerce.ToString(parseNodeValue(c, scope))
				}
			}
		}

		if ip == "" {
			return fmt.Errorf("sec.unblock_ip: IP address required")
		}

		middleware.GlobalBlockList.Remove(ip)
		return nil
	}, engine.SlotMeta{Example: "sec.unblock_ip: '1.2.3.4'"})

	// 3. SEC.IS_BLOCKED
	eng.Register("sec.is_blocked", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var ip string
		target := "is_blocked"

		if node.Value != nil {
			val := resolveValue(node.Value, scope)
			ip = coerce.ToString(val)
		}

		for _, c := range node.Children {
			if c.Name == "ip" {
				ip = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if ip == "" {
			return fmt.Errorf("sec.is_blocked: IP address required")
		}

		blocked := middleware.GlobalBlockList.IsBlocked(ip)
		scope.Set(target, blocked)
		return nil
	}, engine.SlotMeta{Example: "sec.is_blocked: $ip\n  as: $banned"})
}
