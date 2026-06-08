package slots

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"github.com/nextcore/zeno-go/pkg/engine"
	"github.com/nextcore/zeno-go/pkg/utils/coerce"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// StorageDriver defines the contract for all file storage drivers.
type StorageDriver interface {
	Put(ctx context.Context, path string, content []byte) error
	PutFromFile(ctx context.Context, path string, srcPath string) error
	Exists(ctx context.Context, path string) (bool, error)
	Delete(ctx context.Context, path string) error
}

// LocalStorageDriver handles storage on the local filesystem.
type LocalStorageDriver struct {
	storageRoot string
}

func NewLocalStorageDriver() *LocalStorageDriver {
	dir := os.Getenv("STORAGE_DIR")
	if dir == "" {
		dir = filepath.Join("public", "storage")
	}
	return &LocalStorageDriver{storageRoot: dir}
}

func (l *LocalStorageDriver) getSafePath(relPath string) (string, error) {
	cleanRel := filepath.Clean(relPath)
	if strings.HasPrefix(cleanRel, "..") || filepath.IsAbs(cleanRel) {
		return "", fmt.Errorf("storage: path traversal attempt detected")
	}
	return filepath.Join(l.storageRoot, cleanRel), nil
}

func (l *LocalStorageDriver) Put(ctx context.Context, path string, content []byte) error {
	fullPath, err := l.getSafePath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("storage.put: failed to create directories: %v", err)
	}
	return os.WriteFile(fullPath, content, 0644)
}

func (l *LocalStorageDriver) PutFromFile(ctx context.Context, path string, srcPath string) error {
	fullPath, err := l.getSafePath(path)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return fmt.Errorf("storage.put: failed to create directories: %v", err)
	}

	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("storage.put: failed to open source file: %v", err)
	}
	defer srcFile.Close()

	dstFile, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("storage.put: failed to create destination file: %v", err)
	}
	defer dstFile.Close()

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		return fmt.Errorf("storage.put: failed to copy file: %v", err)
	}
	return nil
}

func (l *LocalStorageDriver) Exists(ctx context.Context, path string) (bool, error) {
	fullPath, err := l.getSafePath(path)
	if err != nil {
		return false, err
	}
	info, err := os.Stat(fullPath)
	if err != nil {
		return false, nil
	}
	return !info.IsDir(), nil
}

func (l *LocalStorageDriver) Delete(ctx context.Context, path string) error {
	fullPath, err := l.getSafePath(path)
	if err != nil {
		return err
	}
	err = os.Remove(fullPath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// S3StorageDriver handles storage on AWS S3 or S3-compatible cloud storage (MinIO, R2, Spaces).
type S3StorageDriver struct {
	client *minio.Client
	bucket string
}

func NewS3StorageDriver() (*S3StorageDriver, error) {
	endpoint := os.Getenv("S3_ENDPOINT")
	accessKey := os.Getenv("S3_ACCESS_KEY")
	secretKey := os.Getenv("S3_SECRET_KEY")
	bucket := os.Getenv("S3_BUCKET")
	useSSL, _ := coerce.ToBool(os.Getenv("S3_SSL"))
	region := os.Getenv("S3_REGION")

	if endpoint == "" || accessKey == "" || secretKey == "" || bucket == "" {
		return nil, fmt.Errorf("storage: missing required S3 credentials in environment")
	}

	opts := &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	}

	client, err := minio.New(endpoint, opts)
	if err != nil {
		return nil, fmt.Errorf("storage: failed to initialize S3 client: %v", err)
	}

	return &S3StorageDriver{client: client, bucket: bucket}, nil
}

func (s *S3StorageDriver) Put(ctx context.Context, path string, content []byte) error {
	reader := bytes.NewReader(content)
	_, err := s.client.PutObject(ctx, s.bucket, path, reader, int64(len(content)), minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage.put (S3): %v", err)
	}
	return nil
}

func (s *S3StorageDriver) PutFromFile(ctx context.Context, path string, srcPath string) error {
	_, err := s.client.FPutObject(ctx, s.bucket, path, srcPath, minio.PutObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage.put (S3): %v", err)
	}
	return nil
}

func (s *S3StorageDriver) Exists(ctx context.Context, path string) (bool, error) {
	_, err := s.client.StatObject(ctx, s.bucket, path, minio.StatObjectOptions{})
	if err != nil {
		return false, nil
	}
	return true, nil
}

func (s *S3StorageDriver) Delete(ctx context.Context, path string) error {
	err := s.client.RemoveObject(ctx, s.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("storage.delete (S3): %v", err)
	}
	return nil
}

// resolveDriver selects the driver based on STORAGE_DISK env var with local fallback.
func resolveDriver() (StorageDriver, error) {
	disk := os.Getenv("STORAGE_DISK")
	if strings.ToLower(disk) == "s3" {
		drv, err := NewS3StorageDriver()
		if err != nil {
			fmt.Printf("[STORAGE WARNING] S3 init failed: %v. Falling back to local storage.\n", err)
			return NewLocalStorageDriver(), nil
		}
		return drv, nil
	}
	return NewLocalStorageDriver(), nil
}

func RegisterStorageSlots(eng *engine.Engine) {

	// 1. STORAGE.PUT
	eng.Register("storage.put", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var content interface{}
		var path string
		isFilePath := false
		target := "storage_path"

		for _, c := range node.Children {
			val := parseNodeValue(c, scope)
			if c.Name == "content" || c.Name == "val" || c.Name == "value" {
				content = val
			}
			if c.Name == "path" {
				path = coerce.ToString(val)
			}
			if c.Name == "is_file_path" {
				isFilePath, _ = coerce.ToBool(val)
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if path == "" {
			return fmt.Errorf("storage.put: path is required")
		}
		if content == nil {
			return fmt.Errorf("storage.put: content is required")
		}

		driver, err := resolveDriver()
		if err != nil {
			return err
		}

		// Security Path traversal check on local driver
		if localDrv, ok := driver.(*LocalStorageDriver); ok {
			if _, err := localDrv.getSafePath(path); err != nil {
				return err
			}
		}

		if isFilePath {
			srcPath := coerce.ToString(content)
			if err := driver.PutFromFile(ctx, path, srcPath); err != nil {
				return err
			}
		} else {
			var data []byte
			switch v := content.(type) {
			case []byte:
				data = v
			case string:
				data = []byte(v)
			default:
				data = []byte(coerce.ToString(v))
			}

			if err := driver.Put(ctx, path, data); err != nil {
				return err
			}
		}

		scope.Set(target, path)
		return nil
	}, engine.SlotMeta{
		Description: "Save file content or copy an existing local file to the storage system.",
		Example:     "storage.put:\n  content: $uploaded_temp_path\n  path: 'avatars/1.jpg'\n  is_file_path: true\n  as: $file_url",
		Inputs: map[string]engine.InputMeta{
			"content":      {Description: "File content (string/bytes) or local source filepath to copy", Required: true, Type: "string/bytes"},
			"path":         {Description: "Target relative path inside storage (e.g. 'images/user.png')", Required: true, Type: "string"},
			"is_file_path": {Description: "Whether the content should be treated as a filepath to copy from (Default: false)", Required: false, Type: "bool"},
			"as":           {Description: "Variable name to store the stored file path (Default: 'storage_path')", Required: false, Type: "string"},
		},
	})

	// 2. STORAGE.DELETE
	eng.Register("storage.delete", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var path string
		target := "storage_deleted"

		if node.Value != nil {
			path = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			if c.Name == "path" {
				path = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if path == "" {
			return fmt.Errorf("storage.delete: path is required")
		}

		driver, err := resolveDriver()
		if err != nil {
			return err
		}

		// Security Path traversal check on local driver
		if localDrv, ok := driver.(*LocalStorageDriver); ok {
			if _, err := localDrv.getSafePath(path); err != nil {
				return err
			}
		}

		err = driver.Delete(ctx, path)
		if err != nil {
			scope.Set(target, false)
			return nil
		}

		scope.Set(target, true)
		return nil
	}, engine.SlotMeta{
		Description: "Delete a file from the storage system.",
		Example:     "storage.delete: 'avatars/1.jpg' { as: $deleted }",
		Inputs: map[string]engine.InputMeta{
			"path": {Description: "Relative path of the file to delete", Required: true, Type: "string"},
			"as":   {Description: "Variable name to store delete status (Default: 'storage_deleted')", Required: false, Type: "string"},
		},
	})

	// 3. STORAGE.EXISTS
	eng.Register("storage.exists", func(ctx context.Context, node *engine.Node, scope *engine.Scope) error {
		var path string
		target := "storage_exists"

		if node.Value != nil {
			path = coerce.ToString(resolveValue(node.Value, scope))
		}

		for _, c := range node.Children {
			if c.Name == "path" {
				path = coerce.ToString(parseNodeValue(c, scope))
			}
			if c.Name == "as" {
				target = strings.TrimPrefix(coerce.ToString(c.Value), "$")
			}
		}

		if path == "" {
			return fmt.Errorf("storage.exists: path is required")
		}

		driver, err := resolveDriver()
		if err != nil {
			return err
		}

		// Security Path traversal check on local driver
		if localDrv, ok := driver.(*LocalStorageDriver); ok {
			if _, err := localDrv.getSafePath(path); err != nil {
				return err
			}
		}

		exists, err := driver.Exists(ctx, path)
		if err != nil {
			scope.Set(target, false)
			return nil
		}

		scope.Set(target, exists)
		return nil
	}, engine.SlotMeta{
		Description: "Check if a file exists in the storage system.",
		Example:     "storage.exists: 'avatars/1.jpg' { as: $exists }",
		Inputs: map[string]engine.InputMeta{
			"path": {Description: "Relative path of the file to check", Required: true, Type: "string"},
			"as":   {Description: "Variable name to store exists status (Default: 'storage_exists')", Required: false, Type: "string"},
		},
	})
}
