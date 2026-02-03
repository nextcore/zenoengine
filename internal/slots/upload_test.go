package slots

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestUploadSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterUploadSlots(eng)

	// Setup temp dir for uploads
	tempDir, err := os.MkdirTemp("", "uploads")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	t.Run("http.upload success", func(t *testing.T) {
		// Create multipart body
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("avatar", "test.png")
		assert.NoError(t, err)
		part.Write([]byte("fake image content"))
		writer.Close()

		req := httptest.NewRequest("POST", "/upload", body)
		req.Header.Set("Content-Type", writer.FormDataContentType())

		ctx := context.WithValue(context.Background(), "httpRequest", req)
		scope := engine.NewScope(nil)

		node := &engine.Node{
			Name: "http.upload",
			Children: []*engine.Node{
				{Name: "field", Value: "avatar"},
				{Name: "dest", Value: tempDir},
				{Name: "as", Value: "$filename"},
			},
		}

		err = eng.Execute(ctx, node, scope)
		assert.NoError(t, err)

		filenameVal, ok := scope.Get("filename")
		assert.True(t, ok)
		filename := filenameVal.(string)
		assert.NotEmpty(t, filename)
		assert.Contains(t, filename, "test.png")

		// Verify file exists
		content, err := os.ReadFile(filepath.Join(tempDir, filename))
		assert.NoError(t, err)
		assert.Equal(t, "fake image content", string(content))
	})

	t.Run("http.upload no file", func(t *testing.T) {
		req := httptest.NewRequest("POST", "/upload", nil)
		ctx := context.WithValue(context.Background(), "httpRequest", req)
		scope := engine.NewScope(nil)

		node := &engine.Node{
			Name: "http.upload",
			Children: []*engine.Node{
				{Name: "field", Value: "avatar"},
				{Name: "as", Value: "$filename"},
			},
		}

		err := eng.Execute(ctx, node, scope)
		assert.NoError(t, err)

		filenameVal, ok := scope.Get("filename")
		assert.True(t, ok)
		assert.Equal(t, "", filenameVal)
	})
}
