package repository

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// RemoteRepository implements remote repository functionality with caching
type RemoteRepository struct {
	name         string
	artifactType artifact.ArtifactType
	upstreamURL  string
	storage      storage.Storage
	factory      artifact.Factory
	db           *database.DB
	httpClient   *http.Client
}

// NewRemoteRepository creates a new remote repository
func NewRemoteRepository(name string, artifactType artifact.ArtifactType, upstreamURL string, storage storage.Storage, factory artifact.Factory, db *database.DB) Repository {
	return &RemoteRepository{
		name:         name,
		artifactType: artifactType,
		upstreamURL:  upstreamURL,
		storage:      storage,
		factory:      factory,
		db:           db,
		httpClient:   &http.Client{Timeout: 30 * time.Second},
	}
}

// GetName returns repository name
func (r *RemoteRepository) GetName() string {
	return r.name
}

// GetType returns repository type
func (r *RemoteRepository) GetType() Type {
	return Remote
}

// GetArtifactType returns the artifact type
func (r *RemoteRepository) GetArtifactType() artifact.ArtifactType {
	return r.artifactType
}

// Pull retrieves an artifact from remote repository with caching
func (r *RemoteRepository) Pull(ctx context.Context, path string) (io.ReadCloser, *artifact.Metadata, error) {
	// Check if artifact exists in cache
	repo, err := r.db.GetRepository(ctx, r.name)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get repository info: %w", err)
	}

	cacheEntry, err := r.db.GetCacheEntry(ctx, repo.ID, path)
	if err == nil && time.Now().Before(cacheEntry.ExpiresAt) {
		// Return from cache
		content, err := r.storage.Retrieve(ctx, cacheEntry.LocalPath)
		if err == nil {
			metadata := &artifact.Metadata{
				Size:     cacheEntry.Size,
				Checksum: cacheEntry.Checksum,
			}
			return content, metadata, nil
		}
	}

	// Fetch from upstream
	url := fmt.Sprintf("%s/%s", r.upstreamURL, path)
	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to fetch from upstream: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, nil, fmt.Errorf("upstream returned status %d", resp.StatusCode)
	}

	// Store in cache
	cachePath := fmt.Sprintf("cache/%s/%s", r.name, path)
	if err := r.storage.Store(ctx, cachePath, resp.Body); err != nil {
		return nil, nil, fmt.Errorf("failed to cache artifact: %w", err)
	}

	// Get size and checksum
	size, err := r.storage.GetSize(ctx, cachePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get cached artifact size: %w", err)
	}

	checksum, err := r.storage.GetChecksum(ctx, cachePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get cached artifact checksum: %w", err)
	}

	// Save cache entry
	cacheEntry = &database.CacheEntry{
		RepositoryID: repo.ID,
		Path:         path,
		LocalPath:    cachePath,
		Size:         size,
		Checksum:     checksum,
		ExpiresAt:    time.Now().Add(24 * time.Hour), // Cache for 24 hours
	}
	r.db.SaveCacheEntry(ctx, cacheEntry)

	// Return cached content
	content, err := r.storage.Retrieve(ctx, cachePath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve cached artifact: %w", err)
	}

	metadata := &artifact.Metadata{
		Size:     size,
		Checksum: checksum,
	}

	return content, metadata, nil
}

// Push is not supported for remote repositories
func (r *RemoteRepository) Push(ctx context.Context, path string, content io.Reader, metadata *artifact.Metadata) error {
	return fmt.Errorf("push operation not supported for remote repository")
}

// Delete is not supported for remote repositories
func (r *RemoteRepository) Delete(ctx context.Context, path string) error {
	return fmt.Errorf("delete operation not supported for remote repository")
}

// List returns list of artifacts (limited support)
func (r *RemoteRepository) List(ctx context.Context, prefix string) ([]string, error) {
	// For remote repositories, we can only list cached artifacts
	return r.storage.List(ctx, fmt.Sprintf("cache/%s/%s", r.name, prefix))
}

// GetIndex returns index from upstream
func (r *RemoteRepository) GetIndex(ctx context.Context, indexType string) (io.ReadCloser, error) {
	// Use a default index path
	indexPath := "index.json"
	url := fmt.Sprintf("%s/%s", r.upstreamURL, indexPath)

	resp, err := r.httpClient.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch index from upstream: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		resp.Body.Close()
		return nil, fmt.Errorf("upstream returned status %d for index", resp.StatusCode)
	}

	return resp.Body, nil
}

// InvalidateCache invalidates cached artifacts
func (r *RemoteRepository) InvalidateCache(ctx context.Context, path string) error {
	repo, err := r.db.GetRepository(ctx, r.name)
	if err != nil {
		return fmt.Errorf("failed to get repository info: %w", err)
	}

	cacheEntry, err := r.db.GetCacheEntry(ctx, repo.ID, path)
	if err != nil {
		return nil // Cache entry doesn't exist
	}

	// Delete from storage
	if err := r.storage.Delete(ctx, cacheEntry.LocalPath); err != nil {
		return fmt.Errorf("failed to delete cached artifact: %w", err)
	}

	// Delete cache entry
	return r.db.DeleteCacheEntry(ctx, repo.ID, path)
}

// RebuildIndex rebuilds the repository index (no-op for remote)
func (r *RemoteRepository) RebuildIndex(ctx context.Context) error {
	// Remote repositories don't maintain their own index
	return nil
}

// GetStatistics returns repository statistics
func (r *RemoteRepository) GetStatistics(ctx context.Context) (*Statistics, error) {
	dbStats, err := r.db.GetRepositoryStatistics(ctx, r.name)
	if err != nil {
		return nil, err
	}
	
	return &Statistics{
		TotalArtifacts: dbStats.TotalArtifacts,
		TotalSize:      dbStats.TotalSize,
		PullCount:      dbStats.PullCount,
		PushCount:      dbStats.PushCount,
	}, nil
}
