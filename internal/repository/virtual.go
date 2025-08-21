package repository

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// VirtualRepository implements virtual repository functionality
type VirtualRepository struct {
	name         string
	artifactType artifact.ArtifactType
	upstreams    []Repository
	storage      storage.Storage
	factory      artifact.Factory
	db           *database.DB
}

// NewVirtualRepository creates a new virtual repository
func NewVirtualRepository(name string, artifactType artifact.ArtifactType, upstreams []Repository, storage storage.Storage, factory artifact.Factory, db *database.DB) Repository {
	return &VirtualRepository{
		name:         name,
		artifactType: artifactType,
		upstreams:    upstreams,
		storage:      storage,
		factory:      factory,
		db:           db,
	}
}

// GetName returns repository name
func (v *VirtualRepository) GetName() string {
	return v.name
}

// GetType returns repository type
func (v *VirtualRepository) GetType() Type {
	return Virtual
}

// GetArtifactType returns the artifact type
func (v *VirtualRepository) GetArtifactType() artifact.ArtifactType {
	return v.artifactType
}

// Pull retrieves an artifact from upstream repositories in priority order
func (v *VirtualRepository) Pull(ctx context.Context, path string) (io.ReadCloser, *artifact.Metadata, error) {
	var lastErr error
	
	// Try each upstream repository in order
	for _, upstream := range v.upstreams {
		content, metadata, err := upstream.Pull(ctx, path)
		if err != nil {
			lastErr = err
			continue
		}
		
		// Successfully found artifact
		return content, metadata, nil
	}
	
	if lastErr != nil {
		return nil, nil, fmt.Errorf("artifact not found in any upstream repository: %w", lastErr)
	}
	
	return nil, nil, fmt.Errorf("artifact not found: %s", path)
}

// Push stores an artifact to the first local upstream repository
func (v *VirtualRepository) Push(ctx context.Context, path string, content io.Reader, metadata *artifact.Metadata) error {
	// Find first local repository in upstreams
	for _, upstream := range v.upstreams {
		if upstream.GetType() == Local {
			return upstream.Push(ctx, path, content, metadata)
		}
	}
	
	return fmt.Errorf("no local repository found in virtual repository upstreams")
}

// Delete removes an artifact from all upstream repositories
func (v *VirtualRepository) Delete(ctx context.Context, path string) error {
	var errors []error
	
	for _, upstream := range v.upstreams {
		if upstream.GetType() == Local {
			if err := upstream.Delete(ctx, path); err != nil {
				errors = append(errors, err)
			}
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to delete from some repositories: %v", errors)
	}
	
	return nil
}

// List returns aggregated list of artifacts from all upstreams
func (v *VirtualRepository) List(ctx context.Context, prefix string) ([]string, error) {
	pathMap := make(map[string]bool)
	
	for _, upstream := range v.upstreams {
		paths, err := upstream.List(ctx, prefix)
		if err != nil {
			continue // Skip failed upstreams
		}
		
		for _, path := range paths {
			pathMap[path] = true
		}
	}
	
	// Convert map to slice
	var result []string
	for path := range pathMap {
		result = append(result, path)
	}
	
	return result, nil
}

// GetIndex returns aggregated index from all upstreams
func (v *VirtualRepository) GetIndex(ctx context.Context, indexType string) (io.ReadCloser, error) {
	// For virtual repositories, we need to aggregate artifacts from all upstreams
	// and generate a combined index
	
	var allArtifacts []artifact.Artifact
	
	for _, upstream := range v.upstreams {
		// Get artifacts from each upstream
		if upstream.GetType() == Local {
			// For local repositories, get from database
			artifacts, err := v.db.GetArtifactsByRepository(ctx, upstream.GetName())
			if err != nil {
				continue
			}
			
			for range artifacts {
				tempArt, err := v.factory.CreateArtifact(v.artifactType)
				if err != nil {
					continue
				}
				
				allArtifacts = append(allArtifacts, tempArt)
			}
		}
	}
	
	if len(allArtifacts) == 0 {
		return nil, fmt.Errorf("no artifacts found in virtual repository")
	}
	
	// Generate combined index
	// Convert artifacts to ArtifactInfo list
	artifactInfoList := make([]*artifact.ArtifactInfo, len(allArtifacts))
	for i, art := range allArtifacts {
		info, err := art.ParsePath("")
		if err != nil {
			continue
		}
		artifactInfoList[i] = info
	}
	indexData, err := allArtifacts[0].GenerateIndex(artifactInfoList)
	if err != nil {
		return nil, fmt.Errorf("failed to generate virtual index: %w", err)
	}
	
	return io.NopCloser(strings.NewReader(string(indexData))), nil
}

// InvalidateCache invalidates cache in all upstream repositories
func (v *VirtualRepository) InvalidateCache(ctx context.Context, path string) error {
	var errors []error
	
	for _, upstream := range v.upstreams {
		if err := upstream.InvalidateCache(ctx, path); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to invalidate cache in some repositories: %v", errors)
	}
	
	return nil
}

// RebuildIndex rebuilds index for all upstream repositories
func (v *VirtualRepository) RebuildIndex(ctx context.Context) error {
	var errors []error
	
	for _, upstream := range v.upstreams {
		if err := upstream.RebuildIndex(ctx); err != nil {
			errors = append(errors, err)
		}
	}
	
	if len(errors) > 0 {
		return fmt.Errorf("failed to rebuild index in some repositories: %v", errors)
	}
	
	return nil
}

// GetStatistics returns aggregated statistics from all upstreams
func (v *VirtualRepository) GetStatistics(ctx context.Context) (*Statistics, error) {
	var totalStats Statistics
	
	for _, upstream := range v.upstreams {
		stats, err := upstream.GetStatistics(ctx)
		if err != nil {
			continue
		}
		
		totalStats.TotalArtifacts += stats.TotalArtifacts
		totalStats.TotalSize += stats.TotalSize
		totalStats.PullCount += stats.PullCount
		totalStats.PushCount += stats.PushCount
	}
	
	return &totalStats, nil
}
