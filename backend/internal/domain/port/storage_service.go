package port

import (
	"context"
	"io"
)

type StorageService interface {
	Upload(ctx context.Context, path string, content io.Reader, contentType string) (string, error)
	Download(ctx context.Context, path string) (io.ReadCloser, error)
	Delete(ctx context.Context, path string) error
	GetURL(ctx context.Context, path string) (string, error)
}
