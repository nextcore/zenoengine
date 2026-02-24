package slots

import (
	"path/filepath"
	"testing"
	"time"
	"zeno/pkg/engine"
)

func TestEnsureBladeExt(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"view", "view.blade.zl"},
		{"view.blade.zl", "view.blade.zl"},
		{"dir/view", "dir/view.blade.zl"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ensureBladeExt(tt.input)
			if result != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

func TestClearBladeCache(t *testing.T) {
	// Add dummy entry to cache
	testPath := filepath.Join("views", "dummy.blade.zl")
	bladeCache.Store(testPath, &cachedTemplate{
		ast:     &engine.Node{},
		modTime: time.Now(),
	})

	// Verify it's there
	if _, ok := bladeCache.Load(testPath); !ok {
		t.Fatal("expected cache entry to exist")
	}

	// Clear cache
	ClearBladeCache()

	// Verify it's gone
	if _, ok := bladeCache.Load(testPath); ok {
		t.Fatal("expected cache entry to be deleted")
	}
}
