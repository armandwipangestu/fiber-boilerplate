package storage

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

// S3Provider stores objects in an S3-compatible bucket.
type S3Provider struct {
	client   *s3.Client
	bucket   string
	region   string
	endpoint string
}

// NewS3Provider builds an S3 client. A custom endpoint (e.g. MinIO) uses path
// style addressing; when empty the AWS endpoint and virtual-host URLs are used.
// The bucket must exist (create it in MinIO or AWS beforehand).
func NewS3Provider(cfg config.Config) (Storage, error) {
	if cfg.S3AccessKeyID == "" || cfg.S3SecretAccessKey == "" || cfg.S3Bucket == "" {
		return nil, errors.New("S3 access key, secret key, and bucket are required")
	}

	endpoint := canonEndpoint(cfg.S3Endpoint)

	opts := s3.Options{
		Region:       cfg.S3Region,
		Credentials:  credentials.NewStaticCredentialsProvider(cfg.S3AccessKeyID, cfg.S3SecretAccessKey, ""),
		UsePathStyle: cfg.S3ForcePathStyle || endpoint != "",
	}
	if endpoint != "" {
		opts.BaseEndpoint = aws.String(endpoint)
	}
	client := s3.New(opts)

	return &S3Provider{client: client, bucket: cfg.S3Bucket, region: cfg.S3Region, endpoint: endpoint}, nil
}

// Upload writes the payload to folder/<uuid><ext> and returns the key.
func (s *S3Provider) Upload(ctx context.Context, opts UploadOptions) (string, error) {
	key := buildKey(opts.Folder, opts.Filename)
	contentType := opts.ContentType
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	_, err := s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(opts.Data),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("put object %s: %w", key, err)
	}
	return key, nil
}

// Delete removes the object; not-found is treated as success.
func (s *S3Provider) Delete(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("delete object %s: %w", key, err)
	}
	return nil
}

// GetURL builds the public URL: custom endpoint uses {endpoint}/{bucket}/{key},
// otherwise AWS virtual-hosted style.
func (s *S3Provider) GetURL(key string) string {
	if s.endpoint != "" {
		return fmt.Sprintf("%s/%s/%s", s.endpoint, s.bucket, key)
	}
	return fmt.Sprintf("https://%s.s3.%s.amazonaws.com/%s", s.bucket, s.region, key)
}
