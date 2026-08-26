package storage

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/google/uuid"
)

// Supabase stores files in a Supabase Storage bucket.
type Supabase struct {
	baseURL    string
	serviceKey string
	bucket     string
	httpClient *http.Client
}

func NewSupabase(baseURL, serviceKey, bucket string) (*Supabase, error) {
	if baseURL == "" || serviceKey == "" {
		return nil, fmt.Errorf("supabase url and service role key required")
	}
	return &Supabase{
		baseURL:    strings.TrimRight(baseURL, "/"),
		serviceKey: serviceKey,
		bucket:     bucket,
		httpClient: &http.Client{Timeout: 60 * time.Second},
	}, nil
}

func (s *Supabase) Save(file multipart.File, header *multipart.FileHeader) (string, error) {
	defer file.Close()

	ext := filepath.Ext(header.Filename)
	if ext == "" {
		ext = ".jpg"
	}
	objectPath := fmt.Sprintf("local/%s%s", uuid.New().String(), strings.ToLower(ext))

	data, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("read upload: %w", err)
	}

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "image/jpeg"
	}

	url := fmt.Sprintf("%s/storage/v1/object/%s/%s", s.baseURL, s.bucket, objectPath)
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("x-upsert", "true")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("upload to supabase: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("supabase upload %d: %s", resp.StatusCode, string(body))
	}

	return s.PublicURL(objectPath), nil
}

func (s *Supabase) LocalPath(storagePath string) (string, error) {
	url := storagePath
	if !strings.HasPrefix(url, "http") {
		url = s.PublicURL(storagePath)
	}

	resp, err := s.httpClient.Get(url)
	if err != nil {
		return "", fmt.Errorf("download image: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		return "", fmt.Errorf("download image status %d", resp.StatusCode)
	}

	ext := filepath.Ext(url)
	if ext == "" || len(ext) > 5 {
		ext = ".jpg"
	}
	return SaveToTemp(resp.Body, ext)
}

func (s *Supabase) PublicURL(storagePath string) string {
	if strings.HasPrefix(storagePath, "http://") || strings.HasPrefix(storagePath, "https://") {
		return storagePath
	}
	path := strings.TrimPrefix(storagePath, "/uploads/")
	path = strings.TrimPrefix(path, "/")
	return fmt.Sprintf("%s/storage/v1/object/public/%s/%s", s.baseURL, s.bucket, path)
}

func (s *Supabase) IsRemote() bool { return true }

// EnsureBucket creates the bucket if missing (best-effort on deploy).
func (s *Supabase) EnsureBucket() error {
	url := fmt.Sprintf("%s/storage/v1/bucket", s.baseURL)
	body, _ := json.Marshal(map[string]any{
		"id":     s.bucket,
		"name":   s.bucket,
		"public": true,
	})
	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.serviceKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	// 409 = already exists
	if resp.StatusCode >= 400 && resp.StatusCode != 409 {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("create bucket %d: %s", resp.StatusCode, string(b))
	}
	return nil
}
