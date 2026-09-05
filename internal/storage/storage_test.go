package storage

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

func TestLocalProvider_RoundTrip(t *testing.T) {
	root := t.TempDir()
	cfg := config.Config{StoragePath: root, PublicURL: "http://localhost:8080"}
	provider := NewLocalProvider(cfg)

	key, err := provider.Upload(context.Background(), UploadOptions{
		Folder:      "avatars",
		Filename:    "me.png",
		ContentType: "image/png",
		Data:        []byte("png-bytes"),
	})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "avatars/"), "key=%s", key)
	assert.True(t, strings.HasSuffix(key, ".png"), "key=%s", key)

	data, err := os.ReadFile(filepath.Join(root, key))
	require.NoError(t, err)
	assert.Equal(t, "png-bytes", string(data))

	assert.Equal(t, "http://localhost:8080/uploads/"+key, provider.GetURL(key))

	require.NoError(t, provider.Delete(context.Background(), key))
	_, err = os.Stat(filepath.Join(root, key))
	assert.True(t, os.IsNotExist(err), "file should be removed")
}

func TestLocalProvider_DeleteMissingIsNoop(t *testing.T) {
	provider := NewLocalProvider(config.Config{StoragePath: t.TempDir()})
	assert.NoError(t, provider.Delete(context.Background(), "avatars/does-not-exist.png"))
}

func TestLocalProvider_DefaultRootAndURL(t *testing.T) {
	provider := NewLocalProvider(config.Config{})
	url := provider.GetURL("avatars/x.png")
	assert.Equal(t, "http://localhost:8080/uploads/avatars/x.png", url)
}

func TestLocalProvider_DefaultFolder(t *testing.T) {
	root := t.TempDir()
	provider := NewLocalProvider(config.Config{StoragePath: root, PublicURL: "http://x"})

	key, err := provider.Upload(context.Background(), UploadOptions{
		Filename: "no-folder.txt",
		Data:     []byte("hi"),
	})
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(key, "avatars/"), "default folder should be avatars, got %s", key)
}

func TestS3Provider_GetURL_CustomEndpoint(t *testing.T) {
	cfg := config.Config{
		S3AccessKeyID:     "minioadmin",
		S3SecretAccessKey: "minioadmin",
		S3Bucket:          "fiber-boilerplate",
		S3Endpoint:        "http://localhost:9000",
		S3Region:          "us-east-1",
	}
	provider, err := NewS3Provider(cfg)
	require.NoError(t, err) // bucket create fails gracefully when MinIO is down

	assert.Equal(t, "http://localhost:9000/fiber-boilerplate/avatars/a.png", provider.GetURL("avatars/a.png"))
}

func TestS3Provider_GetURL_AWSEndpoint(t *testing.T) {
	cfg := config.Config{
		S3AccessKeyID:     "ak",
		S3SecretAccessKey: "sk",
		S3Bucket:          "my-bucket",
		S3Region:          "eu-west-1",
	}
	provider, err := NewS3Provider(cfg)
	require.NoError(t, err)

	assert.Equal(t, "https://my-bucket.s3.eu-west-1.amazonaws.com/avatars/a.png", provider.GetURL("avatars/a.png"))
}

func TestS3Provider_RequiresCredentials(t *testing.T) {
	_, err := NewS3Provider(config.Config{S3Bucket: "b"})
	assert.Error(t, err)
}

func TestNew_PicksLocalWhenS3Missing(t *testing.T) {
	cfg := config.Config{StoragePath: t.TempDir()}
	storage := New(cfg, nil)
	assert.IsType(t, &LocalProvider{}, storage)
}

func TestNew_PicksS3WhenConfigured(t *testing.T) {
	cfg := config.Config{
		StoragePath:       t.TempDir(),
		S3AccessKeyID:     "ak",
		S3SecretAccessKey: "sk",
		S3Bucket:          "b",
	}
	storage := New(cfg, nil)
	assert.IsType(t, &S3Provider{}, storage)
}
