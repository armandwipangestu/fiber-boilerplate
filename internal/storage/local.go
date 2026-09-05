package storage

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/armandwipangestu/fiber-boilerplate/internal/config"
)

// LocalProvider stores objects under a folder on disk and serves them from the
// /uploads route, mirroring the NestJS boilerplate fallback storage.
type LocalProvider struct {
	root string
	url  string
}

// NewLocalProvider builds the local fallback provider, creating the storage
// directory when missing.
func NewLocalProvider(cfg config.Config) Storage {
	root := cfg.StoragePath
	if root == "" {
		root = "storage"
	}

	urlBase := cfg.PublicURL
	if urlBase == "" {
		urlBase = "http://localhost:8080"
	}

	_ = os.MkdirAll(root, 0o755)

	return &LocalProvider{root: root, url: strings.TrimRight(urlBase, "/")}
}

// Upload writes the payload to folder/<uuid><ext> and returns the key.
func (p *LocalProvider) Upload(ctx context.Context, opts UploadOptions) (string, error) {
	key := buildKey(opts.Folder, opts.Filename)

	folder := filepath.Join(p.root, filepath.Dir(key))
	if err := os.MkdirAll(folder, 0o755); err != nil {
		return "", fmt.Errorf("create storage folder: %w", err)
	}

	target := filepath.Join(p.root, key)
	if err := os.WriteFile(target, opts.Data, 0o644); err != nil {
		return "", fmt.Errorf("write storage file: %w", err)
	}
	return key, nil
}

// Delete removes the object; a missing file is not an error.
func (p *LocalProvider) Delete(ctx context.Context, key string) error {
	target := filepath.Join(p.root, key)
	if err := os.Remove(target); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil
		}
		return fmt.Errorf("delete storage file: %w", err)
	}
	return nil
}

// GetURL builds a URL served by the app's /uploads static route.
func (p *LocalProvider) GetURL(key string) string {
	return fmt.Sprintf("%s/uploads/%s", p.url, key)
}
