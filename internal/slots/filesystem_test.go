package slots

import (
	"context"
	"os"
	"testing"
	"zeno/pkg/engine"
)

func TestFileSystemSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterFileSystemSlots(eng)

	t.Run("write_and_read", func(t *testing.T) {
		scope := engine.NewScope(nil)
		path := "test_temp_file.txt"
		content := "hello filesystem"

		// Write
		writeNode := &engine.Node{
			Name: "io.file.write",
			Children: []*engine.Node{
				{Name: "path", Value: path},
				{Name: "content", Value: content},
			},
		}
		if err := eng.Execute(context.Background(), writeNode, scope); err != nil {
			t.Fatalf("io.file.write failed: %v", err)
		}
		defer os.Remove(path)

		// Read
		readNode := &engine.Node{
			Name: "io.file.read",
			Children: []*engine.Node{
				{Name: "path", Value: path},
				{Name: "as", Value: "$result"},
			},
		}
		if err := eng.Execute(context.Background(), readNode, scope); err != nil {
			t.Fatalf("io.file.read failed: %v", err)
		}

		val, _ := scope.Get("result")
		if val != content {
			t.Errorf("Expected '%s', got %v", content, val)
		}
	})

	t.Run("delete", func(t *testing.T) {
		scope := engine.NewScope(nil)
		path := "test_delete_file.txt"
		os.WriteFile(path, []byte("test"), 0644)

		node := &engine.Node{
			Name:  "io.file.delete",
			Value: path,
		}
		if err := eng.Execute(context.Background(), node, scope); err != nil {
			t.Fatalf("io.file.delete failed: %v", err)
		}

		if _, err := os.Stat(path); !os.IsNotExist(err) {
			t.Errorf("File should have been deleted")
		}
	})
}
