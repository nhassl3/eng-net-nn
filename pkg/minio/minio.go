package minio

import (
	"context"
	"fmt"
	"io"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type ByteStorage interface {
	Upload(
		ctx context.Context,
		objectName, contentType string,
		reader io.Reader,
		size int64,
	) (string, error)
	Delete(ctx context.Context, objectName string) error
	objectURL(objectName string) string
}

// MinIO S3 Compatibility storage of the files
type MinIO struct {
	client *minio.Client
	publicURL,
	endpoint,
	bucket string
	useSSL bool
}

func NewMinIO(ctx context.Context, publicURL, endpoint, accessKey, secretKey, token, bucket string, useSSL bool) (*MinIO, error) {
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

	policy := fmt.Sprintf(`{
			"Version": "2012-10-17",
			"Statement": [{
				"Effect": "Allow",
				"Principal": {"AWS": ["*"]},
				"Action": ["s3:GetObject"],
				"Resource": ["arn:aws:s3:::%s/*"]
			}]
		}`, bucket)
	if err := minioClient.SetBucketPolicy(ctx, bucket, policy); err != nil {
		return nil, fmt.Errorf("minio: set bucket policy: %w", err)
	}

	return &MinIO{client: minioClient, publicURL: publicURL, endpoint: endpoint, bucket: bucket, useSSL: useSSL}, nil
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
	return m.objectURL(objectName), nil
}

func (m *MinIO) Delete(ctx context.Context, objectName string) error {
	if err := m.client.RemoveObject(ctx, m.bucket, objectName, minio.RemoveObjectOptions{ForceDelete: true}); err != nil {
		return fmt.Errorf("minio: remove object %q error: %w", objectName, err)
	}
	return nil
}

func (m *MinIO) objectURL(objectName string) string {
	return fmt.Sprintf("%s/%s", m.bucket, objectName)
}
