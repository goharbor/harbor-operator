package storage

import (
	"context"
	"fmt"

	minv6 "github.com/minio/minio-go/v6"
)

// Minio defines related operations of minio.
type Minio interface {
	IsBucketExists(ctx context.Context, bucket string) (bool, error)
	CreateBucket(ctx context.Context, bucket string) error
}

// MinioEndpoint contains the related access info of a minio server.
type MinioEndpoint struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	// Optional
	Location string
	UseSSL   bool
}

// MinioClient is an implementation of Minio.
type MinioClient struct {
	client   *minv6.Client
	location string
}

// NewMinioClient constructs a new minio client.
func NewMinioClient(endpoint *MinioEndpoint) (Minio, error) {
	client, err := minv6.New(
		endpoint.Endpoint,
		endpoint.AccessKeyID,
		endpoint.SecretAccessKey,
		endpoint.UseSSL,
	)
	if err != nil {
		return nil, fmt.Errorf("create minv6 client error: %w", err)
	}

	return &MinioClient{
		client:   client,
		location: endpoint.Location,
	}, nil
}

// IsBucketExists checks if the bucket existing.
func (m MinioClient) IsBucketExists(ctx context.Context, bucket string) (bool, error) {
	return m.client.BucketExistsWithContext(ctx, bucket)
}

// CreateBucket creates a bucket.
func (m MinioClient) CreateBucket(ctx context.Context, bucket string) error {
	return m.client.MakeBucketWithContext(ctx, bucket, m.location)
}
