package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/google/uuid"
)

const UploadDir = "uploads"

// Local stores uploaded files on disk under uploads/.
type Local struct {
	Root string
}

func NewLocal(root string) (*Local, error) {
	if root == "" {
		root = UploadDir
	}
	if err := os.MkdirAll(root, 0o755); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &Local{Root: root}, nil
}

// Save stores a multipart file and returns the web-relative path (e.g. /uploads/abc.jpg).
func (l *Local) Save(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	name := uuid.New().String() + strings.ToLower(ext)
	destPath := filepath.Join(l.Root, name)

	dest, err := os.Create(destPath)
	if err != nil {
		return "", fmt.Errorf("create file: %w", err)
	}
	defer dest.Close()

	if _, err := io.Copy(dest, file); err != nil {
		return "", fmt.Errorf("write file: %w", err)
	}

	return "/" + filepath.ToSlash(filepath.Join(l.Root, name)), nil
}

// Dir returns the filesystem path for static file serving.
func (l *Local) Dir() string {
	return l.Root
}
