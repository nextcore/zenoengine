package slots

import (
	"context"
	"image"
	"image/color"
	"image/png"
	"os"
	"path/filepath"
	"testing"
	"zeno/pkg/engine"

	"github.com/stretchr/testify/assert"
)

func TestImageSlots(t *testing.T) {
	eng := engine.NewEngine()
	RegisterImageSlots(eng)

	// Create a temporary image
	tempDir, err := os.MkdirTemp("", "images")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tempDir)

	imgPath := filepath.Join(tempDir, "test.png")

	// Generate a 10x10 PNG
	img := image.NewRGBA(image.Rect(0, 0, 10, 10))
	img.Set(0, 0, color.RGBA{255, 0, 0, 255})

	f, err := os.Create(imgPath)
	if err != nil {
		t.Fatal(err)
	}
	png.Encode(f, img)
	f.Close()

	t.Run("image.info", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "image.info",
			Children: []*engine.Node{
				{Name: "path", Value: imgPath},
				{Name: "as", Value: "$info"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		infoRaw, ok := scope.Get("info")
		assert.True(t, ok)
		info := infoRaw.(map[string]interface{})
		assert.Equal(t, 10, info["width"])
		assert.Equal(t, 10, info["height"])
	})

	t.Run("image.resize (copy)", func(t *testing.T) {
		destPath := filepath.Join(tempDir, "resized.png")
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "image.resize",
			Children: []*engine.Node{
				{Name: "source", Value: imgPath},
				{Name: "dest", Value: destPath},
				{Name: "width", Value: 5},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		// Verify dest exists
		_, err = os.Stat(destPath)
		assert.NoError(t, err)
	})
}
