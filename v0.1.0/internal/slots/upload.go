package slots

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings" // [WAJIB] Jangan lupa import strings
	"time"
	"zeno/pkg/engine"
	"zeno/pkg/utils/coerce"
)

func RegisterUploadSlots(eng *engine.Engine) {
	// ==========================================
	// SLOT: HTTP.UPLOAD
	// ==========================================
	eng.Register("http.upload", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		r, ok := ctx.Value("httpRequest").(*http.Request)
		if !ok {
			return fmt.Errorf("http.upload: request context not found")
		}

		// 1. Ambil Parameter
		field := "file"
		destDir := "public/uploads"
		targetVar := "uploaded_file"

		for _, c := range node.Children {
			if c.Name == "field" {
				field = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "dest" {
				destDir = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				// [FIX UTAMA] Bersihkan awalan $ agar variable tersimpan dengan benar
				targetVar = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		// 2. Ambil File dari Form
		file, header, err := r.FormFile(field)
		if err != nil {
			// Jika user tidak upload file (saat edit), biarkan kosong
			scope.Set(targetVar, "")
			return nil
		}
		defer file.Close()

		// 3. Buat Folder Tujuan (jika belum ada)
		if _, err := os.Stat(destDir); os.IsNotExist(err) {
			os.MkdirAll(destDir, 0755)
		}

		// 4. Generate Nama File Unik (timestamp_filename)
		// Bersihkan nama file dari spasi agar URL aman
		cleanName := strings.ReplaceAll(header.Filename, " ", "_")
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), cleanName)
		dstPath := filepath.Join(destDir, filename)

		// 5. Simpan File
		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("http.upload: failed to create file: %v", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return fmt.Errorf("http.upload: failed to save file: %v", err)
		}

		// 6. Return HANYA nama file ke variable (agar sesuai logika DB)
		scope.Set(targetVar, filename)
		return nil
	}, engine.SlotMeta{Example: "http.upload:\n  field: image\n  as: $new_file"})
}
