package slots

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"

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

	// ─── captcha.id / captcha.new ───────────────────────────────────────────
	handlerCaptchaID := func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var target string
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

		if target != "" {
			scope.Set(target, id)
		} else {
			// Write raw string ID to HTTP ResponseWriter if available
			wVal := ctx.Value("httpResponseWriter")
			if wVal == nil {
				wVal = ctx.Value("httpWriter")
			}
			if wVal != nil {
				if w, ok := wVal.(http.ResponseWriter); ok {
					w.Write([]byte(id))
				}
			} else {
				scope.Set("captcha_id", id)
			}
		}
		return nil
	}

	eng.Register("captcha.id", handlerCaptchaID, engine.SlotMeta{
		Description: "Membuat captcha ID baru dan mencetaknya atau menyimpannya ke scope.",
		Example:     "captcha.id\n  as: $captcha_id",
	})

	// ─── captcha.verify ─────────────────────────────────────────────────────
	eng.Register("captcha.verify", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var id, input, elseRedirect, target string

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "id":
				id = coerce.ToString(val)
			case "input":
				input = coerce.ToString(val)
			case "else_redirect":
				elseRedirect = coerce.ToString(val)
			case "as":
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if id == "" {
			return fmt.Errorf("captcha.verify: 'id' is required")
		}
		if input == "" {
			return fmt.Errorf("captcha.verify: 'input' is required")
		}

		isValid := captcha.VerifyString(id, input)

		if target != "" {
			scope.Set(target, isValid)
		} else {
			scope.Set("captcha_valid", isValid)
		}

		if !isValid && elseRedirect != "" {
			wVal := ctx.Value("httpResponseWriter")
			if wVal == nil {
				wVal = ctx.Value("httpWriter")
			}
			rVal := ctx.Value("httpRequest")

			if wVal != nil && rVal != nil {
				if w, okW := wVal.(http.ResponseWriter); okW {
					if r, okR := rVal.(*http.Request); okR {
						http.Redirect(w, r, elseRedirect, http.StatusSeeOther)
						return fmt.Errorf("captcha verification failed, redirecting to %s", elseRedirect)
					}
				}
			}
		}
		return nil
	}, engine.SlotMeta{
		Description: "Memverifikasi jawaban user terhadap captcha ID dan melakukan aksi opsional jika gagal.",
		Example: `captcha.verify
  id: $captcha_id
  input: $user_input
  else_redirect: /login?error=captcha
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

		if r == nil {
			fmt.Printf("   ⚠️  [CAPTCHA] Skip captcha.serve: router is nil (worker mode?)\n")
			return nil
		}

		// Daftarkan handler ke router
		handler := captcha.Server(captcha.StdWidth, captcha.StdHeight)
		r.Handle(prefix+"/*", http.StripPrefix(prefix, handler))

		fmt.Printf("   ➕ [CAPTCHA] Serving at %s/*\n", prefix)
		return nil
	}, engine.SlotMeta{
		Description: "Mendaftarkan route handler captcha ke router. Melayani PNG dan WAV secara otomatis.",
		Example: `captcha.serve
  prefix: /captcha`,
	})
}
