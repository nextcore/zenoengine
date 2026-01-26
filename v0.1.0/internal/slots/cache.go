package slots

import (
	"context"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

// Cache slots - Currently disabled (Redis removed)
// All cache operations are no-ops for compatibility

func RegisterCacheSlots(eng *engine.Engine, rdb interface{}) {

	// CACHE.PUT
	eng.Register("cache.put", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Cache disabled - Redis removed
		// This is a no-op for compatibility
		return nil
	}, engine.SlotMeta{
		Description: "Menyimpan data sementara (Cache). Currently disabled.",
		Example: `cache.put
  key: "homepage_stats"
  val: stats_data
  ttl: "30m"`,
	})

	// CACHE.GET
	eng.Register("cache.get", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var target string
		var defaultVal interface{}
		target = "cache_value"

		for _, c := range node.Children {
			rawVal := parseNodeValue(c, scope)
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
			if c.Name == "default" {
				defaultVal = rawVal
			}
		}

		// Always return default value (cache miss)
		scope.Set(target, defaultVal)
		return nil
	}, engine.SlotMeta{
		Description: "Mengambil data cache. Always returns default value (cache disabled).",
		Example: `cache.get
  key: "homepage_stats"
  default: 0
  as: $stats`,
	})

	// CACHE.FORGET
	eng.Register("cache.forget", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		// Cache disabled - no-op
		return nil
	}, engine.SlotMeta{Example: "cache.forget: 'homepage_stats'"})
}
