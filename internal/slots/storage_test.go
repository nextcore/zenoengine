package slots

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"github.com/nextcore/zeno-go/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestStorageSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterStorageSlots(eng)

	// Setup custom STORAGE_DIR for testing
	testDir := filepath.Join(".", "test_storage")
	os.Setenv("STORAGE_DIR", testDir)
	defer func() {
		os.Unsetenv("STORAGE_DIR")
		os.RemoveAll(testDir)
	}()

	t.Run("storage.put and exists and delete", func(t *testing.T) {
		scope := engine.NewScope(nil)

		// 1. Put raw content
		nodePut := &engine.Node{
			Name: "storage.put",
			Children: []*engine.Node{
				{Name: "content", Value: "hello storage"},
				{Name: "path", Value: "docs/hello.txt"},
				{Name: "as", Value: "$saved_path"},
			},
		}

		err := eng.Execute(context.Background(), nodePut, scope)
		assert.NoError(t, err)

		savedPath, ok := scope.Get("saved_path")
		assert.True(t, ok)
		assert.Equal(t, "docs/hello.txt", savedPath.(string))

		// Verify file actually written to filesystem
		fullPath := filepath.Join(testDir, "docs", "hello.txt")
		contentBytes, err := os.ReadFile(fullPath)
		assert.NoError(t, err)
		assert.Equal(t, "hello storage", string(contentBytes))

		// 2. Check exists
		scopeExists := engine.NewScope(nil)
		nodeExists := &engine.Node{
			Name: "storage.exists",
			Children: []*engine.Node{
				{Name: "path", Value: "docs/hello.txt"},
				{Name: "as", Value: "$exists"},
			},
		}

		err = eng.Execute(context.Background(), nodeExists, scopeExists)
		assert.NoError(t, err)

		exists, ok := scopeExists.Get("exists")
		assert.True(t, ok)
		assert.True(t, exists.(bool))

		// 3. Delete file
		scopeDelete := engine.NewScope(nil)
		nodeDelete := &engine.Node{
			Name: "storage.delete",
			Children: []*engine.Node{
				{Name: "path", Value: "docs/hello.txt"},
				{Name: "as", Value: "$deleted"},
			},
		}

		err = eng.Execute(context.Background(), nodeDelete, scopeDelete)
		assert.NoError(t, err)

		deleted, ok := scopeDelete.Get("deleted")
		assert.True(t, ok)
		assert.True(t, deleted.(bool))

		// Verify file is gone
		_, err = os.Stat(fullPath)
		assert.True(t, os.IsNotExist(err))
	})

	t.Run("storage.put copy file path", func(t *testing.T) {
		scope := engine.NewScope(nil)

		// Create dummy file to copy from
		srcFile := filepath.Join(testDir, "temp_src.txt")
		err := os.MkdirAll(testDir, 0755)
		assert.NoError(t, err)
		err = os.WriteFile(srcFile, []byte("source data"), 0644)
		assert.NoError(t, err)
		defer os.Remove(srcFile)

		nodePut := &engine.Node{
			Name: "storage.put",
			Children: []*engine.Node{
				{Name: "content", Value: srcFile},
				{Name: "path", Value: "copied/target.txt"},
				{Name: "is_file_path", Value: true},
				{Name: "as", Value: "$saved_path"},
			},
		}

		err = eng.Execute(context.Background(), nodePut, scope)
		assert.NoError(t, err)

		targetPath := filepath.Join(testDir, "copied", "target.txt")
		contentBytes, err := os.ReadFile(targetPath)
		assert.NoError(t, err)
		assert.Equal(t, "source data", string(contentBytes))
	})

	t.Run("storage path traversal prevention", func(t *testing.T) {
		scope := engine.NewScope(nil)

		nodePut := &engine.Node{
			Name: "storage.put",
			Children: []*engine.Node{
				{Name: "content", Value: "malicious"},
				{Name: "path", Value: "../../../etc/passwd"},
			},
		}

		err := eng.Execute(context.Background(), nodePut, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "traversal attempt")
	})

	t.Run("s3 driver initialization fallback", func(t *testing.T) {
		os.Setenv("STORAGE_DISK", "s3")
		defer os.Unsetenv("STORAGE_DISK")

		os.Unsetenv("S3_ENDPOINT")

		driver, err := resolveDriver()
		assert.NoError(t, err)

		_, isLocal := driver.(*LocalStorageDriver)
		assert.True(t, isLocal, "should fall back to local storage driver when s3 configuration is incomplete")
	})
}
