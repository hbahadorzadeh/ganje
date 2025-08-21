package repository

import (
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// LocalRepository implements local repository functionality
type LocalRepository struct {
	name         string
	artifactType artifact.ArtifactType
	storage      storage.Storage
	factory      artifact.Factory
	db           database.DatabaseInterface
}

// NewLocalRepository creates a new local repository
func NewLocalRepository(name string, artifactType artifact.ArtifactType, storage storage.Storage, factory artifact.Factory, db database.DatabaseInterface) Repository {
	return &LocalRepository{
		name:         name,
		artifactType: artifactType,
		storage:      storage,
		factory:      factory,
		db:           db,
	}
}

// GetName returns repository name
func (l *LocalRepository) GetName() string {
	return l.name
}

// GetType returns repository type
func (l *LocalRepository) GetType() Type {
	return Local
}

// GetArtifactType returns the artifact type
func (l *LocalRepository) GetArtifactType() artifact.ArtifactType {
	return l.artifactType
}

// Pull retrieves an artifact from local storage
func (l *LocalRepository) Pull(ctx context.Context, path string) (io.ReadCloser, *artifact.Metadata, error) {
	// Check if artifact exists in storage
	exists, err := l.storage.Exists(ctx, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to check artifact existence: %w", err)
	}
	if !exists {
		return nil, nil, fmt.Errorf("artifact not found: %s", path)
	}

	// Get artifact from storage
	content, err := l.storage.Retrieve(ctx, path)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to retrieve artifact: %w", err)
	}

	// Get metadata from database
	artifactInfo, err := l.db.GetArtifactByPath(ctx, l.name, path)
	if err != nil {
		content.Close()
		return nil, nil, fmt.Errorf("failed to get artifact metadata: %w", err)
	}

	// Update pull statistics
	if err := l.db.IncrementPullCount(ctx, artifactInfo.ID); err != nil {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to update pull statistics: %v\n", err)
	}

	metadata := &artifact.Metadata{
		Name:     artifactInfo.Name,
		Version:  artifactInfo.Version,
		Group:    artifactInfo.Group,
		Size:     artifactInfo.Size,
		Checksum: artifactInfo.Checksum,
	}

	return content, metadata, nil
}

// Push stores an artifact to local storage
func (l *LocalRepository) Push(ctx context.Context, path string, content io.Reader, metadata *artifact.Metadata) error {
	// Store artifact in storage
	if err := l.storage.Store(ctx, path, content); err != nil {
		return fmt.Errorf("failed to store artifact: %w", err)
	}

	// Get size and checksum from storage
	size, err := l.storage.GetSize(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to get artifact size: %w", err)
	}

	checksum, err := l.storage.GetChecksum(ctx, path)
	if err != nil {
		return fmt.Errorf("failed to get artifact checksum: %w", err)
	}

	// Save metadata to database
	if err := l.db.SaveArtifact(ctx, &database.ArtifactInfo{
		RepositoryID: 1, // TODO: Get actual repository ID
		Type:         string(l.artifactType),
		Name:         metadata.Name,
		Version:      metadata.Version,
		Path:         path,
		Size:         size,
		Checksum:     checksum,
		CreatedAt:    time.Now(),
		PushCount:    1,
	}); err != nil {
		return fmt.Errorf("failed to save artifact metadata: %w", err)
	}

	return nil
}

// Delete removes an artifact from local storage
func (l *LocalRepository) Delete(ctx context.Context, path string) error {
	// Delete from storage
	if err := l.storage.Delete(ctx, path); err != nil {
		return fmt.Errorf("failed to delete artifact from storage: %w", err)
	}

	// Delete from database
	if err := l.db.DeleteArtifactByPath(ctx, l.name, path); err != nil {
		return fmt.Errorf("failed to delete artifact metadata: %w", err)
	}

	return nil
}

// List returns list of artifacts
func (l *LocalRepository) List(ctx context.Context, prefix string) ([]string, error) {
	return l.storage.List(ctx, prefix)
}

// GetIndex returns index for the repository
func (l *LocalRepository) GetIndex(ctx context.Context, indexType string) (io.ReadCloser, error) {
	// Get all artifacts for this repository
	artifacts, err := l.db.GetArtifactsByRepository(ctx, l.name)
	if err != nil {
		return nil, fmt.Errorf("failed to get artifacts: %w", err)
	}

	// Convert to artifact interface
	var artifactList []artifact.Artifact
	for range artifacts {
		artifact, err := l.factory.CreateArtifact(l.artifactType)
		if err != nil {
			return nil, fmt.Errorf("failed to create artifact: %w", err)
		}
		
		artifactList = append(artifactList, artifact)
	}

	// Generate index using first artifact (they should all be same type)
	if len(artifactList) == 0 {
		return nil, fmt.Errorf("no artifacts found")
	}

	// Convert artifacts to ArtifactInfo list
	artifactInfoList := make([]*artifact.ArtifactInfo, len(artifactList))
	for i, art := range artifactList {
		info, err := art.ParsePath("")
		if err != nil {
			continue
		}
		artifactInfoList[i] = info
	}
	indexData, err := artifactList[0].GenerateIndex(artifactInfoList)
	if err != nil {
		return nil, fmt.Errorf("failed to generate index: %w", err)
	}

	return io.NopCloser(strings.NewReader(string(indexData))), nil
}

// InvalidateCache invalidates cached artifacts (no-op for local)
func (l *LocalRepository) InvalidateCache(ctx context.Context, path string) error {
	// Local repositories don't have cache
	return nil
}

// RebuildIndex rebuilds the repository index
func (l *LocalRepository) RebuildIndex(ctx context.Context) error {
	// For local repositories, we can rebuild index by scanning storage
	// and updating database records
	paths, err := l.storage.List(ctx, "")
	if err != nil {
		return fmt.Errorf("failed to list storage: %w", err)
	}

	for _, path := range paths {
		// Parse artifact metadata from path
		tempArt, err := l.factory.CreateArtifact(l.artifactType)
		if err != nil {
			continue
		}

		metadata, err := tempArt.ParsePath(path)
		if err != nil {
			continue
		}

		// Get size and checksum
		size, err := l.storage.GetSize(ctx, path)
		if err != nil {
			continue
		}

		checksum, err := l.storage.GetChecksum(ctx, path)
		if err != nil {
			continue
		}

		// Update database
		if err := l.db.SaveArtifact(ctx, &database.ArtifactInfo{
			RepositoryID: 1, // TODO: Get actual repository ID
			Type:         string(l.artifactType),
			Name:         metadata.Name,
			Version:      metadata.Version,
			Path:         path,
			Size:         size,
			Checksum:     checksum,
			CreatedAt:    time.Now(),
		}); err != nil {
			continue
		}
	}

	return nil
}

// GetStatistics returns repository statistics
func (l *LocalRepository) GetStatistics(ctx context.Context) (*Statistics, error) {
	dbStats, err := l.db.GetRepositoryStatistics(ctx, l.name)
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
