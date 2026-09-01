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
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"
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

	// ==========================================
	// SLOT: UPLOAD.WEBP
	// ==========================================
	eng.Register("upload.webp", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		rVal := ctx.Value("httpRequest")
		if rVal == nil {
			return fmt.Errorf("upload.webp: httpRequest context not found")
		}
		r, ok := rVal.(*http.Request)
		if !ok {
			return fmt.Errorf("upload.webp: invalid httpRequest context")
		}

		field := "file"
		subDir := ""
		oldFile := ""
		targetVar := "uploaded_file"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			switch c.Name {
			case "field":
				field = coerce.ToString(val)
			case "dir", "dest":
				subDir = coerce.ToString(val)
			case "old":
				oldFile = coerce.ToString(val)
			case "as":
				targetVar = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		file, header, err := r.FormFile(field)
		if err != nil {
			if err == http.ErrMissingFile {
				scope.Set(targetVar, oldFile)
				return nil
			}
			scope.Set(targetVar, oldFile)
			return nil
		}
		defer file.Close()

		uploadDir := filepath.Join("assets", "uploads", subDir)
		if subDir == "" {
			uploadDir = "public/uploads"
		}

		if err := os.MkdirAll(uploadDir, 0755); err != nil {
			return fmt.Errorf("upload.webp: failed to create directory: %v", err)
		}

		cleanName := strings.ReplaceAll(header.Filename, " ", "_")
		filename := fmt.Sprintf("%d_%s", time.Now().Unix(), cleanName)
		dstPath := filepath.Join(uploadDir, filename)

		dst, err := os.Create(dstPath)
		if err != nil {
			return fmt.Errorf("upload.webp: failed to create destination file: %v", err)
		}
		defer dst.Close()

		if _, err := io.Copy(dst, file); err != nil {
			return fmt.Errorf("upload.webp: failed to save file: %v", err)
		}

		if oldFile != "" {
			oldPath := filepath.Join(uploadDir, oldFile)
			os.Remove(oldPath)
		}

		scope.Set(targetVar, filename)
		return nil
	}, engine.SlotMeta{
		Description: "Memproses unggahan gambar dan menyimpan ke direktori upload.",
		Example:     "upload.webp:\n  field: gambar\n  dir: berita\n  old: $old_gambar\n  as: $new_gambar",
	})
}
