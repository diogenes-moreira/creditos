package storage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

// LocalStorage implements StorageService using the local filesystem.
type LocalStorage struct {
	basePath string
}

func NewLocalStorage(basePath string) *LocalStorage {
	_ = os.MkdirAll(basePath, 0755)
	return &LocalStorage{basePath: basePath}
}

func (s *LocalStorage) Upload(_ context.Context, path string, content io.Reader, contentType string) (string, error) {
	_ = contentType
	fullPath := filepath.Join(s.basePath, path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return "", fmt.Errorf("failed to create directory: %w", err)
	}
	f, err := os.Create(fullPath)
	if err != nil {
		return "", fmt.Errorf("failed to create file: %w", err)
	}
	defer f.Close()
	if _, err := io.Copy(f, content); err != nil {
		return "", fmt.Errorf("failed to write file: %w", err)
	}
	return "/storage/" + path, nil
}

func (s *LocalStorage) Download(_ context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(s.basePath, path)
	f, err := os.Open(fullPath)
	if err != nil {
		return nil, fmt.Errorf("file not found: %w", err)
	}
	return f, nil
}

func (s *LocalStorage) Delete(_ context.Context, path string) error {
	fullPath := filepath.Join(s.basePath, path)
	return os.Remove(fullPath)
}

func (s *LocalStorage) GetURL(_ context.Context, path string) (string, error) {
	return "/storage/" + path, nil
}
