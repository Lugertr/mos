package storage

import (
	"context"
	"io"
)

type FileUploadResult struct {
	Provider string
	Bucket   string
	Key      string
	Mime     string
	Size     int64
	Sha256   string
}

type Storage interface {
	// UploadStream: size can be -1 if unknown, but prefer passing accurate size when available.
	UploadStream(ctx context.Context, filename string, r io.Reader, size int64, contentType string) (FileUploadResult, error)
	// SignedURL returns presigned GET URL valid for expirySeconds.
	SignedURL(ctx context.Context, bucket, key string, expirySeconds int) (string, error)
}
