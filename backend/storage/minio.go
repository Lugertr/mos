package storage

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/url"
	"path"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinioConfig struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	Bucket          string // default bucket
	Region          string
	Prefix          string // optional key prefix like "documents/"
}

type MinioStorage struct {
	client *minio.Client
	cfg    MinioConfig
}

func NewMinioStorage(cfg MinioConfig) (*MinioStorage, error) {
	mc, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
		Region: cfg.Region,
	})
	if err != nil {
		return nil, err
	}
	return &MinioStorage{client: mc, cfg: cfg}, nil
}

func (m *MinioStorage) UploadStream(ctx context.Context, filename string, r io.Reader, size int64, contentType string) (FileUploadResult, error) {
	name := path.Base(filename)
	now := time.Now().UTC().Format("20060102-150405")
	key := strings.Trim(m.cfg.Prefix, "/") + "/" + now + "-" + name
	key = strings.TrimPrefix(key, "/")

	// hash while streaming
	h := sha256.New()
	tee := io.TeeReader(r, h)

	opts := minio.PutObjectOptions{ContentType: contentType}
	info, err := m.client.PutObject(ctx, m.cfg.Bucket, key, tee, size, opts)
	if err != nil {
		return FileUploadResult{}, err
	}

	sha := hex.EncodeToString(h.Sum(nil))
	return FileUploadResult{
		Provider: "minio",
		Bucket:   m.cfg.Bucket,
		Key:      key,
		Mime:     contentType,
		Size:     info.Size,
		Sha256:   sha,
	}, nil
}

func (m *MinioStorage) SignedURL(ctx context.Context, bucket, key string, expirySeconds int) (string, error) {
	if bucket == "" {
		bucket = m.cfg.Bucket
	}
	// empty params
	values := url.Values{}
	expiry := time.Duration(expirySeconds) * time.Second
	u, err := m.client.PresignedGetObject(ctx, bucket, key, expiry, values)
	if err != nil {
		return "", err
	}
	return u.String(), nil
}
