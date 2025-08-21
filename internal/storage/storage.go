package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"path/filepath"
)

// Storage represents an abstract storage interface
type Storage interface {
	// Store saves content to the storage with the given path
	Store(ctx context.Context, path string, content io.Reader) error
	
	// Retrieve gets content from the storage
	Retrieve(ctx context.Context, path string) (io.ReadCloser, error)
	
	// Delete removes content from the storage
	Delete(ctx context.Context, path string) error
	
	// Exists checks if content exists at the given path
	Exists(ctx context.Context, path string) (bool, error)
	
	// List returns a list of paths with the given prefix
	List(ctx context.Context, prefix string) ([]string, error)
	
	// GetSize returns the size of content at the given path
	GetSize(ctx context.Context, path string) (int64, error)
	
	// GetChecksum returns the checksum of content at the given path
	GetChecksum(ctx context.Context, path string) (string, error)
}

// ShardedPath generates a sharded path based on hash
func ShardedPath(hash string) string {
	if len(hash) < 4 {
		return hash
	}
	return filepath.Join(hash[:2], hash[2:4], hash)
}

// GenerateHash generates SHA256 hash from content
func GenerateHash(content io.Reader) (string, error) {
	hasher := sha256.New()
	if _, err := io.Copy(hasher, content); err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// Config represents storage configuration
type Config struct {
	Type      string            `yaml:"type"`
	LocalPath string            `yaml:"local_path,omitempty"`
	Options   map[string]string `yaml:"options,omitempty"`
}

// Factory creates storage instances
type Factory interface {
	CreateStorage(config *Config) (Storage, error)
}
