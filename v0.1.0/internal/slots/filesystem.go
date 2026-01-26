package slots

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterFileSystemSlots(eng *engine.Engine) {

	// 1. IO.FILE.WRITE
	eng.Register("io.file.write", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var path, content string
		var mode int = 0644

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "path" {
				path = coerce.ToString(val)
			}
			if c.Name == "content" {
				content = coerce.ToString(val)
			}
			if c.Name == "mode" {
				if m, err := coerce.ToInt(val); err == nil {
					mode = m
				}
			}
		}

		if path == "" {
			return fmt.Errorf("io.file.write: path is required")
		}

		cleanPath := filepath.Clean(filepath.FromSlash(path))
		if err := os.MkdirAll(filepath.Dir(cleanPath), 0755); err != nil {
			return err
		}
		return os.WriteFile(cleanPath, []byte(content), os.FileMode(mode))
	}, engine.SlotMeta{})

	// 2. IO.FILE.READ
	eng.Register("io.file.read", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var path, target string
		target = "file_content"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "path" {
				path = coerce.ToString(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if node.Value != nil {
			val, err := MustResolveValue(node.Value, scope, coerce.ToString(node.Value))
			if err != nil {
				return err
			}
			path = coerce.ToString(val)
		}

		if path == "" {
			return fmt.Errorf("io.file.read: path is required")
		}

		cleanPath := filepath.Clean(filepath.FromSlash(path))
		content, err := os.ReadFile(cleanPath)
		if err != nil {
			return err
		}

		scope.Set(target, string(content))
		return nil
	}, engine.SlotMeta{Example: "io.file.read: $path\n  as: $content"})

	// 3. IO.DIR.CREATE
	eng.Register("io.dir.create", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		path := coerce.ToString(resolveValue(node.Value, scope))
		if path == "" {
			for _, c := range node.Children {
				if c.Name == "path" {
					path = coerce.ToString(parseNodeValue(c, scope))
				}
			}
		}
		if path == "" {
			return fmt.Errorf("io.dir.create: path is required")
		}
		return os.MkdirAll(filepath.Clean(filepath.FromSlash(path)), 0755)
	}, engine.SlotMeta{})

	// 4. IO.FILE.DELETE
	eng.Register("io.file.delete", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var pathRaw string
		if node.Value != nil {
			val, err := MustResolveValue(node.Value, scope, coerce.ToString(node.Value))
			if err != nil {
				return err
			}
			pathRaw = coerce.ToString(val)
		}

		if pathRaw == "" {
			for _, c := range node.Children {
				if c.Name == "path" {
					pathRaw = coerce.ToString(parseNodeValue(c, scope))
				}
			}
		}

		if pathRaw == "" {
			return fmt.Errorf("io.file.delete: path is required")
		}

		cleanPath := filepath.Clean(filepath.FromSlash(pathRaw))
		fmt.Printf("🗑️ [FILESYSTEM] Deleting: '%s'\n", cleanPath)

		err := os.Remove(cleanPath)
		if err != nil {
			if os.IsNotExist(err) {
				return nil
			}
			return fmt.Errorf("failed to delete file: %v", err)
		}
		return nil
	}, engine.SlotMeta{Example: "io.file.delete: $path"})
}
