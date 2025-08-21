package database

import "context"

// DatabaseInterface defines the interface for database operations
type DatabaseInterface interface {
	GetRepository(ctx context.Context, name string) (*Repository, error)
	SaveRepository(ctx context.Context, repo *Repository) error
	UpdateRepository(ctx context.Context, name string, updates map[string]interface{}) error
	DeleteRepository(ctx context.Context, name string) error
	ListRepositories(ctx context.Context) ([]*Repository, error)
	GetArtifactsByRepository(ctx context.Context, repositoryName string) ([]*ArtifactInfo, error)
	SaveArtifact(ctx context.Context, artifact *ArtifactInfo) error
	GetArtifact(ctx context.Context, repositoryName, name, version string) (*ArtifactInfo, error)
	GetArtifactByPath(ctx context.Context, repositoryName, path string) (*ArtifactInfo, error)
	DeleteArtifactByPath(ctx context.Context, repositoryName, path string) error
	IncrementPullCount(ctx context.Context, artifactID uint) error
	GetRepositoryStatistics(ctx context.Context, repositoryName string) (*Statistics, error)
	LogAccess(ctx context.Context, log *AccessLog) error
}


// Ensure DB implements DatabaseInterface
var _ DatabaseInterface = (*DB)(nil)
