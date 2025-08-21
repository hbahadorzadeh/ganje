package repository

import (
	"context"
	"io"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// Type represents repository type
type Type string

const (
	Local   Type = "local"
	Remote  Type = "remote"
	Virtual Type = "virtual"
)

// Repository represents an artifact repository
type Repository interface {
	// GetName returns repository name
	GetName() string
	
	// GetType returns repository type
	GetType() Type
	
	// GetArtifactType returns the artifact type this repository handles
	GetArtifactType() artifact.ArtifactType
	
	// Pull retrieves an artifact
	Pull(ctx context.Context, path string) (io.ReadCloser, *artifact.Metadata, error)
	
	// Push stores an artifact
	Push(ctx context.Context, path string, content io.Reader, metadata *artifact.Metadata) error
	
	// Delete removes an artifact
	Delete(ctx context.Context, path string) error
	
	// List returns list of artifacts
	List(ctx context.Context, prefix string) ([]string, error)
	
	// GetIndex returns index for the repository
	GetIndex(ctx context.Context, indexType string) (io.ReadCloser, error)
	
	// InvalidateCache invalidates cached artifacts
	InvalidateCache(ctx context.Context, path string) error
	
	// RebuildIndex rebuilds the repository index
	RebuildIndex(ctx context.Context) error
	
	// GetStatistics returns repository statistics
	GetStatistics(ctx context.Context) (*Statistics, error)
}

// Statistics represents repository statistics
type Statistics struct {
	TotalArtifacts int64 `json:"total_artifacts"`
	TotalSize      int64 `json:"total_size"`
	PullCount      int64 `json:"pull_count"`
	PushCount      int64 `json:"push_count"`
}

// Config represents repository configuration
type Config struct {
	Name         string            `yaml:"name"`
	Type         string            `yaml:"type"`
	ArtifactType string            `yaml:"artifact_type"`
	URL          string            `yaml:"url,omitempty"`
	Upstream     []string          `yaml:"upstream,omitempty"`
	Options      map[string]string `yaml:"options,omitempty"`
}

// Manager manages repositories
type Manager interface {
	// GetRepository returns a repository by name
	GetRepository(name string) (Repository, error)
	
	// ListRepositories returns all repositories
	ListRepositories() []Repository
	
	// CreateRepository creates a new repository
	CreateRepository(config *Config) (Repository, error)
	
	// DeleteRepository deletes a repository
	DeleteRepository(name string) error
}

// Factory creates repository instances
type Factory interface {
	CreateRepository(config *Config, storage storage.Storage, artifactFactory artifact.Factory) (Repository, error)
}
