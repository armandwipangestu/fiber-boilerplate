package storage

import (
	"context"
	"log/slog"
	"path"
	"strings"

	"github.com/google/uuid"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

// defaultFolder is used when no folder is provided for an upload.
const defaultFolder = "avatars"

// UploadOptions carries an upload payload and its desired location.
type UploadOptions struct {
	Folder      string
	Filename    string
	ContentType string
	Data        []byte
}

// Bucket is the storage abstraction. Implementations write to object storage
// (S3-compatible, e.g. MinIO) or to a local folder as a fallback.
type Storage interface {
	// Upload stores the payload and returns the storage key (folder/name).
	Upload(ctx context.Context, opts UploadOptions) (string, error)
	// Delete removes the object at key; a missing object is not an error.
	Delete(ctx context.Context, key string) error
	// GetURL returns a publicly accessible URL for the given key.
	GetURL(key string) string
}

// New selects the S3 provider when the required credentials are configured,
// otherwise falls back to local disk storage.
func New(cfg config.Config, logger *slog.Logger) Storage {
	if cfg.S3AccessKeyID != "" && cfg.S3SecretAccessKey != "" && cfg.S3Bucket != "" {
		provider, err := NewS3Provider(cfg)
		if err == nil {
			if logger != nil {
				logger.Info("storage initialized", "provider", "s3", "endpoint", cfg.S3Endpoint)
			}
			return provider
		}
		if logger != nil {
			logger.Warn("failed to initialize S3 storage, falling back to local", "error", err)
		}
	} else if logger != nil {
		logger.Warn("S3 config missing, using local storage", "path", cfg.StoragePath)
	}
	return NewLocalProvider(cfg)
}

// buildKey generates a unique object key under folder preserving the upload
// extension, mirroring the NestJS boilerplate (`avatars/<uuid><ext>`).
func buildKey(folder, filename string) string {
	if folder == "" {
		folder = defaultFolder
	}
	ext := path.Ext(filename)
	return folder + "/" + uuid.NewString() + ext
}

// canonEndpoint trims trailing slashes so URLs compose cleanly.
func canonEndpoint(raw string) string {
	return strings.TrimRight(raw, "/")
}
