package minio

import (
	"context"
	"fmt"
	"io"
	"path/filepath"
	"time"

	"github.com/gabriel-vasile/mimetype"
	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/nhassl3/IpBuild-backend/internal/domain"
)

// MaxFileSize is the maximum allowed size of an uploaded file (10 MB).
const MaxFileSize = 10 << 20

var allowedContentTypes = map[string]string{
	"text/plain":         ".txt",
	"text/rtf":           ".rtf",
	"application/rtf":    ".rtf",
	"application/pdf":    ".pdf",
	"application/msword": ".doc",
	"application/vnd.openxmlformats-officedocument.wordprocessingml.document": ".docx",
}

type ByteStorage interface {
	Upload(
		ctx context.Context,
		objectName, contentType string,
		reader io.Reader,
		size int64,
	) (string, error)
	Delete(ctx context.Context, objectName string) error
	// PresignedURL returns a temporary, signed URL granting read access to the
	// object for the given duration.
	PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error)
}

// MinIO S3 Compatibility storage of the files
type MinIO struct {
	client *minio.Client
	bucket string
}

func NewMinIO(ctx context.Context, endpoint, accessKey, secretKey, token, bucket string, useSSL bool) (*MinIO, error) {
	minioClient, err := minio.New(endpoint, &minio.Options{
		Creds:      credentials.NewStaticV4(accessKey, secretKey, token),
		Secure:     useSSL,
		MaxRetries: 5,
	})
	if err != nil {
		return nil, fmt.Errorf("minio: create client: %w", err)
	}

	exists, err := minioClient.BucketExists(ctx, bucket)
	if err != nil {
		return nil, fmt.Errorf("minio: check bucket %q: %w", bucket, err)
	}
	if !exists {
		if err := minioClient.MakeBucket(ctx, bucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio: create bucket %q: %w", bucket, err)
		}
	}

	// The bucket stays private: objects (resumes contain PII) are never exposed
	// publicly. Read access is granted on demand via presigned URLs.
	return &MinIO{client: minioClient, bucket: bucket}, nil
}

func (m *MinIO) Upload(
	ctx context.Context,
	objectName, contentType string,
	reader io.Reader,
	size int64,
) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", fmt.Errorf("minio: upload to %q error: %w", objectName, err)
	}
	return objectName, nil
}

func (m *MinIO) Delete(ctx context.Context, objectName string) error {
	if err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
		return fmt.Errorf("minio: remove object %q error: %w", objectName, err)
	}
	return nil
}

func (m *MinIO) PresignedURL(ctx context.Context, objectName string, expiry time.Duration) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, m.bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("minio: presign object %q error: %w", objectName, err)
	}
	return u.String(), nil
}

// ResolveContentType detects the real content type from the file bytes (the
// client-provided Content-Type is not trusted) and validates size and type
// against the allowlist. It returns the detected content type on success.
func ResolveContentType(data []byte) (string, error) {
	if len(data) > MaxFileSize {
		return "", domain.ErrFileTooLarge
	}
	contentType := mimetype.Detect(data).String()
	if _, ok := allowedContentTypes[contentType]; !ok {
		return "", domain.ErrInvalidContentType
	}
	return contentType, nil
}

// GenerateObjectName creates a unique object path for file storage.
func GenerateObjectName(prefix, owner, contentType string) string {
	ext := allowedContentTypes[contentType]
	return filepath.Join(prefix, owner, uuid.NewString()+ext)
}
