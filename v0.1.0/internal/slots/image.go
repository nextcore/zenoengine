package slots

import (
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
	// Library resize pihak ketiga (golang.org/x/image/draw) direkomendasikan,
	// tapi untuk "Zero-effort" start kita pakai basic logic atau library standar.
)

func RegisterImageSlots(eng *engine.Engine) {

	// IMAGE.INFO (Cek Ukuran)
	eng.Register("image.info", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		path := ""
		target := "image_info"

		for _, c := range node.Children {
			// Gunakan parseNodeValue dari utils.go
			if c.Name == "path" {
				path = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				// Bersihkan awalan $ agar konsisten
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if path == "" {
			return fmt.Errorf("image.info: path is required")
		}

		f, err := os.Open(path)
		if err != nil {
			return err
		}
		defer f.Close()

		// Decode Config (Hanya header, cepat)
		cfg, _, err := image.DecodeConfig(f)
		if err != nil {
			return fmt.Errorf("image.info: failed to decode image: %v", err)
		}

		info := map[string]interface{}{
			"width":  cfg.Width,
			"height": cfg.Height,
		}
		scope.Set(target, info)
		return nil
	}, engine.SlotMeta{Example: "image.info\n  path: 'uploads/foto.jpg'"})

	// IMAGE.RESIZE (Simple Implementation)
	eng.Register("image.resize", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var srcPath, destPath string
		var w, h int

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "source" {
				srcPath = coerce.ToString(val)
			}
			if c.Name == "dest" {
				destPath = coerce.ToString(val)
			}
			if c.Name == "width" {
				w, _ = coerce.ToInt(val)
			}
			if c.Name == "height" {
				h, _ = coerce.ToInt(val)
			}
		}

		// [FIX] Bungkam error "declared and not used"
		// Variabel w & h disiapkan untuk implementasi resize library masa depan.
		_ = w
		_ = h

		if srcPath == "" || destPath == "" {
			return fmt.Errorf("image.resize: source and dest paths required")
		}

		// Membuka file source
		file, err := os.Open(srcPath)
		if err != nil {
			return err
		}
		defer file.Close()

		img, format, err := image.Decode(file)
		if err != nil {
			return fmt.Errorf("image.resize: decode failed: %v", err)
		}

		// LOGIKA RESIZE (Disederhanakan / Placeholder)
		// Saat ini hanya melakukan re-encoding (compress) tanpa resize dimensi.
		// TODO: Gunakan 'github.com/nfnt/resize' untuk menggunakan w & h.

		out, err := os.Create(destPath)
		if err != nil {
			return err
		}
		defer out.Close()

		if format == "png" {
			return png.Encode(out, img)
		}
		// Default JPEG quality 75
		return jpeg.Encode(out, img, &jpeg.Options{Quality: 75})

	}, engine.SlotMeta{
		Description: "Mengubah ukuran atau format gambar (Placeholder implementasi).",
		Example: `image.resize
  source: "input.jpg"
  dest: "output_thumb.jpg"
  width: 100`,
	})
}
