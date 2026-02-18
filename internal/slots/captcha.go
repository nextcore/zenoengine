package slots

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"

	"github.com/dchest/captcha"
	"github.com/go-chi/chi/v5"
)

// RegisterCaptchaSlots mendaftarkan slot-slot captcha ke engine.
// Slot yang tersedia:
//   - captcha.new    : Buat captcha baru dan simpan ID ke scope
//   - captcha.verify : Verifikasi jawaban user
//   - captcha.image  : Tulis PNG captcha ke http.ResponseWriter
//   - captcha.serve  : Daftarkan route handler bawaan captcha ke router
func RegisterCaptchaSlots(eng *engine.Engine, r *chi.Mux) {

	// ─── captcha.new ────────────────────────────────────────────────────────
	// Membuat captcha baru dengan panjang default (6 digit) atau custom.
	// Menyimpan captcha ID ke scope.
	//
	// Contoh:
	//   captcha.new
	//     as: $captcha_id
	//
	//   captcha.new
	//     length: 4
	//     as: $captcha_id
	eng.Register("captcha.new", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		target := "captcha_id"
		length := captcha.DefaultLen

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "as":
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			case "length":
				if l, err := coerce.ToInt(val); err == nil && l > 0 {
					length = l
				}
			}
		}

		id := captcha.NewLen(length)
		scope.Set(target, id)
		return nil
	}, engine.SlotMeta{
		Description: "Membuat captcha baru dan menyimpan ID-nya ke scope.",
		Example: `captcha.new
  as: $captcha_id`,
	})

	// ─── captcha.verify ─────────────────────────────────────────────────────
	// Memverifikasi jawaban user terhadap captcha ID yang diberikan.
	// Menghapus captcha dari store setelah verifikasi (one-time use).
	//
	// Contoh:
	//   captcha.verify
	//     id: $captcha_id
	//     answer: $user_input
	//     as: $is_valid
	eng.Register("captcha.verify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id, answer string
		target := "captcha_valid"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "id":
				id = coerce.ToString(val)
			case "answer":
				answer = coerce.ToString(val)
			case "as":
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if id == "" {
			return fmt.Errorf("captcha.verify: 'id' is required")
		}
		if answer == "" {
			return fmt.Errorf("captcha.verify: 'answer' is required")
		}

		ok := captcha.VerifyString(id, answer)
		scope.Set(target, ok)
		return nil
	}, engine.SlotMeta{
		Description: "Memverifikasi jawaban user terhadap captcha ID. Menghapus captcha setelah verifikasi.",
		Example: `captcha.verify
  id: $captcha_id
  answer: $user_input
  as: $is_valid`,
	})

	// ─── captcha.image ──────────────────────────────────────────────────────
	// Menulis gambar PNG captcha langsung ke http.ResponseWriter.
	// Gunakan slot ini di dalam route handler untuk menampilkan captcha.
	//
	// Contoh:
	//   captcha.image
	//     id: $captcha_id
	//     width: 240
	//     height: 80
	eng.Register("captcha.image", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id string
		width := captcha.StdWidth
		height := captcha.StdHeight

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "id":
				id = coerce.ToString(val)
			case "width":
				if w, err := coerce.ToInt(val); err == nil && w > 0 {
					width = w
				}
			case "height":
				if h, err := coerce.ToInt(val); err == nil && h > 0 {
					height = h
				}
			}
		}

		if id == "" {
			return fmt.Errorf("captcha.image: 'id' is required")
		}

		// Tulis ke buffer terlebih dahulu untuk menangkap error
		var buf bytes.Buffer
		if err := captcha.WriteImage(&buf, id, width, height); err != nil {
			return fmt.Errorf("captcha.image: failed to write image: %w", err)
		}

		// Tulis ke ResponseWriter jika tersedia di context
		wVal := ctx.Value("httpResponseWriter")
		if wVal != nil {
			w := wVal.(http.ResponseWriter)
			w.Header().Set("Content-Type", "image/png")
			w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
			_, err := w.Write(buf.Bytes())
			return err
		}

		// Simpan bytes ke scope jika tidak ada ResponseWriter (misal: testing)
		scope.Set("captcha_image_bytes", buf.Bytes())
		return nil
	}, engine.SlotMeta{
		Description: "Menulis gambar PNG captcha ke http.ResponseWriter atau menyimpan bytes ke scope.",
		Example: `captcha.image
  id: $captcha_id
  width: 240
  height: 80`,
	})

	// ─── captcha.serve ──────────────────────────────────────────────────────
	// Mendaftarkan route handler bawaan dchest/captcha ke router chi.
	// Handler ini melayani gambar dan audio captcha secara otomatis.
	//
	// URL pattern:
	//   GET /captcha/{id}.png  → gambar PNG
	//   GET /captcha/{id}.wav  → audio WAV
	//   GET /captcha/{id}.png?reload=1 → reload captcha
	//
	// Contoh:
	//   captcha.serve
	//     prefix: /captcha
	//
	// Setelah ini, di HTML gunakan:
	//   <img src="/captcha/{captcha_id}.png">
	eng.Register("captcha.serve", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		prefix := "/captcha"

		for _, c := range node.Children {
			if c.Name == "prefix" {
				prefix = coerce.ToString(parseNodeValue(c, scope))
			}
		}

		// Pastikan prefix diawali /
		if !strings.HasPrefix(prefix, "/") {
			prefix = "/" + prefix
		}

		// Daftarkan handler ke router
		handler := captcha.Server(captcha.StdWidth, captcha.StdHeight)
		r.Handle(prefix+"/*", http.StripPrefix(prefix, handler))

		return nil
	}, engine.SlotMeta{
		Description: "Mendaftarkan route handler captcha ke router. Melayani PNG dan WAV secara otomatis.",
		Example: `captcha.serve
  prefix: /captcha`,
	})
}
