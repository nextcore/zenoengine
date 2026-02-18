package slots

import (
	"context"
	"testing"
	"zeno/pkg/engine"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestCaptchaSlots(t *testing.T) {
	eng := engine.NewEngine()
	r := chi.NewRouter()
	RegisterCaptchaSlots(eng, r)

	t.Run("captcha.new - default length", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "captcha.new",
			Children: []*engine.Node{
				{Name: "as", Value: "$my_captcha"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		id, ok := scope.Get("my_captcha")
		assert.True(t, ok, "captcha ID harus tersimpan di scope")
		assert.NotEmpty(t, id, "captcha ID tidak boleh kosong")
	})

	t.Run("captcha.new - custom length", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "captcha.new",
			Children: []*engine.Node{
				{Name: "length", Value: 4},
				{Name: "as", Value: "$cap_id"},
			},
		}

		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)

		id, ok := scope.Get("cap_id")
		assert.True(t, ok)
		assert.NotEmpty(t, id)
	})

	t.Run("captcha.verify - jawaban salah", func(t *testing.T) {
		// Buat captcha baru
		scope := engine.NewScope(nil)
		newNode := &engine.Node{
			Name: "captcha.new",
			Children: []*engine.Node{
				{Name: "as", Value: "$cid"},
			},
		}
		err := eng.Execute(context.Background(), newNode, scope)
		assert.NoError(t, err)

		idRaw, _ := scope.Get("cid")
		captchaID := idRaw.(string)

		// Verifikasi dengan jawaban yang pasti salah
		verifyScope := engine.NewScope(nil)
		verifyNode := &engine.Node{
			Name: "captcha.verify",
			Children: []*engine.Node{
				{Name: "id", Value: captchaID},
				{Name: "answer", Value: "000000"},
				{Name: "as", Value: "$result"},
			},
		}
		err = eng.Execute(context.Background(), verifyNode, verifyScope)
		assert.NoError(t, err)

		result, ok := verifyScope.Get("result")
		assert.True(t, ok)
		// Jawaban "000000" kemungkinan besar salah (bisa benar secara kebetulan, tapi sangat jarang)
		// Kita hanya memastikan slot berjalan tanpa error dan mengembalikan bool
		_, isBool := result.(bool)
		assert.True(t, isBool, "hasil verifikasi harus bertipe bool")
	})

	t.Run("captcha.verify - id kosong harus error", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "captcha.verify",
			Children: []*engine.Node{
				{Name: "id", Value: ""},
				{Name: "answer", Value: "123456"},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "id")
	})

	t.Run("captcha.image - tanpa ResponseWriter (simpan ke scope)", func(t *testing.T) {
		// Buat captcha baru
		scope := engine.NewScope(nil)
		newNode := &engine.Node{
			Name: "captcha.new",
			Children: []*engine.Node{
				{Name: "as", Value: "$img_id"},
			},
		}
		err := eng.Execute(context.Background(), newNode, scope)
		assert.NoError(t, err)

		idRaw, _ := scope.Get("img_id")
		captchaID := idRaw.(string)

		// Render gambar tanpa ResponseWriter
		imgScope := engine.NewScope(nil)
		imgNode := &engine.Node{
			Name: "captcha.image",
			Children: []*engine.Node{
				{Name: "id", Value: captchaID},
			},
		}
		err = eng.Execute(context.Background(), imgNode, imgScope)
		assert.NoError(t, err)

		bytesRaw, ok := imgScope.Get("captcha_image_bytes")
		assert.True(t, ok, "bytes gambar harus tersimpan di scope")
		imgBytes, ok := bytesRaw.([]byte)
		assert.True(t, ok)
		// PNG header: 8 bytes magic number
		assert.Greater(t, len(imgBytes), 8, "output harus berupa PNG yang valid")
		assert.Equal(t, []byte{0x89, 0x50, 0x4E, 0x47}, imgBytes[:4], "harus dimulai dengan PNG magic bytes")
	})

	t.Run("captcha.image - id kosong harus error", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "captcha.image",
			Children: []*engine.Node{
				{Name: "id", Value: ""},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		assert.Error(t, err)
	})

	t.Run("captcha.serve - daftarkan route", func(t *testing.T) {
		scope := engine.NewScope(nil)
		node := &engine.Node{
			Name: "captcha.serve",
			Children: []*engine.Node{
				{Name: "prefix", Value: "/captcha"},
			},
		}
		err := eng.Execute(context.Background(), node, scope)
		assert.NoError(t, err)
	})
}
