package storage

import (
	"context"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStorage implements local file system storage
type LocalStorage struct {
	basePath string
}

// NewLocalStorage creates a new local storage instance
func NewLocalStorage(basePath string) Storage {
	return &LocalStorage{basePath: basePath}
}

// Store saves content to local file system
func (l *LocalStorage) Store(ctx context.Context, path string, content io.Reader) error {
	fullPath := filepath.Join(l.basePath, path)
	
	// Create directory if it doesn't exist
	dir := filepath.Dir(fullPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}
	
	// Create file
	file, err := os.Create(fullPath)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()
	
	// Copy content to file
	if _, err := io.Copy(file, content); err != nil {
		return fmt.Errorf("failed to write content: %w", err)
	}
	
	return nil
}

// Retrieve gets content from local file system
func (l *LocalStorage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	fullPath := filepath.Join(l.basePath, path)
	
	file, err := os.Open(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file not found: %s", path)
		}
		return nil, fmt.Errorf("failed to open file: %w", err)
	}
	
	return file, nil
}

// Delete removes content from local file system
func (l *LocalStorage) Delete(ctx context.Context, path string) error {
	fullPath := filepath.Join(l.basePath, path)
	
	if err := os.Remove(fullPath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", path)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}
	
	return nil
}

// Exists checks if content exists in local file system
func (l *LocalStorage) Exists(ctx context.Context, path string) (bool, error) {
	fullPath := filepath.Join(l.basePath, path)
	
	_, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, fmt.Errorf("failed to check file existence: %w", err)
	}
	
	return true, nil
}

// List returns a list of paths with the given prefix
func (l *LocalStorage) List(ctx context.Context, prefix string) ([]string, error) {
	searchPath := filepath.Join(l.basePath, prefix)
	
	var paths []string
	err := filepath.Walk(searchPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		
		if !info.IsDir() {
			// Convert absolute path to relative path
			relPath, err := filepath.Rel(l.basePath, path)
			if err != nil {
				return err
			}
			paths = append(paths, strings.ReplaceAll(relPath, "\\", "/"))
		}
		
		return nil
	})
	
	if err != nil && !os.IsNotExist(err) {
		return nil, fmt.Errorf("failed to list files: %w", err)
	}
	
	return paths, nil
}

// GetSize returns the size of content in local file system
func (l *LocalStorage) GetSize(ctx context.Context, path string) (int64, error) {
	fullPath := filepath.Join(l.basePath, path)
	
	info, err := os.Stat(fullPath)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, fmt.Errorf("file not found: %s", path)
		}
		return 0, fmt.Errorf("failed to get file info: %w", err)
	}
	
	return info.Size(), nil
}

// GetChecksum returns the SHA256 checksum of content in local file system
func (l *LocalStorage) GetChecksum(ctx context.Context, path string) (string, error) {
	file, err := l.Retrieve(ctx, path)
	if err != nil {
		return "", err
	}
	defer file.Close()
	
	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", fmt.Errorf("failed to calculate checksum: %w", err)
	}
	
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}
