package repository

import (
	"fmt"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// DefaultFactory implements the Factory interface
type DefaultFactory struct{}

// NewDefaultFactory creates a new DefaultFactory
func NewDefaultFactory() Factory {
	return &DefaultFactory{}
}

// CreateRepository creates a repository based on the configuration
func (f *DefaultFactory) CreateRepository(config *Config, storage storage.Storage, artifactFactory artifact.Factory) (Repository, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	if config.Name == "" {
		return nil, fmt.Errorf("repository name is required")
	}

	if config.ArtifactType == "" {
		return nil, fmt.Errorf("artifact type is required")
	}

	artifactType := artifact.ArtifactType(config.ArtifactType)

	switch Type(config.Type) {
	case Local:
		return NewLocalRepository(config.Name, artifactType, storage, artifactFactory, nil), nil
	case Remote:
		if config.URL == "" {
			return nil, fmt.Errorf("URL is required for remote repositories")
		}
		return NewRemoteRepository(config.Name, artifactType, config.URL, storage, artifactFactory, nil), nil
	case Virtual:
		if len(config.Upstream) == 0 {
			return nil, fmt.Errorf("upstream repositories are required for virtual repositories")
		}
		// For virtual repositories, we need to pass actual Repository instances, not strings
		// In a real implementation, this would resolve the upstream repository names
		var upstreams []Repository
		return NewVirtualRepository(config.Name, artifactType, upstreams, storage, artifactFactory, nil), nil
	default:
		return nil, fmt.Errorf("unsupported repository type: %s", config.Type)
	}
}
