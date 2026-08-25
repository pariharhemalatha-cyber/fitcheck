package storage

import (
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"github.com/ashokparihar/fitcheck/internal/config"
	"github.com/google/uuid"
)

// Storage handles clothing photo uploads (local disk or Supabase).
type Storage interface {
	Save(file multipart.File, header *multipart.FileHeader) (storagePath string, err error)
	LocalPath(storagePath string) (string, error)
	PublicURL(storagePath string) string
	IsRemote() bool
}

// New returns Supabase storage on Vercel/production, local disk otherwise.
func New(cfg *config.Config) (Storage, error) {
	if cfg.UseSupabaseStorage() {
		return NewSupabase(cfg.SupabaseURL, cfg.SupabaseServiceRoleKey, "closet")
	}
	root := UploadDir
	if cfg.IsVercel {
		root = "/tmp/uploads"
	}
	return NewLocal(root)
}

func PublicSrc(storagePath string, s Storage) string {
	if storagePath == "" {
		return ""
	}
	url := s.PublicURL(storagePath)
	if strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") {
		return url
	}
	if strings.HasPrefix(url, "/") {
		return url
	}
	return "/uploads/" + strings.TrimPrefix(url, "/uploads/")
}

// SaveToTemp copies bytes to /tmp for AI vision analysis.
func SaveToTemp(r io.Reader, ext string) (string, error) {
	if ext == "" {
		ext = ".jpg"
	}
	path := filepath.Join(os.TempDir(), uuid.New().String()+ext)
	f, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer f.Close()
	if _, err := io.Copy(f, r); err != nil {
		return "", err
	}
	return path, nil
}

// LocalDirStorage wraps Local for static file serving.
type LocalDirStorage struct {
	*Local
}

func NewLocalDir(cfg *config.Config) (*Local, error) {
	if cfg.UseSupabaseStorage() {
		return nil, fmt.Errorf("remote storage active")
	}
	return NewLocal(UploadDir)
}
