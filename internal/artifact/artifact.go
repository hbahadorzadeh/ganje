package artifact

import (
	"context"
	"io"
	"time"
)

// ArtifactType represents the type of artifact
type ArtifactType string

const (
	ArtifactTypeMaven     ArtifactType = "maven"
	ArtifactTypePyPI      ArtifactType = "pypi"
	ArtifactTypeHelm      ArtifactType = "helm"
	ArtifactTypeDocker    ArtifactType = "docker"
	ArtifactTypeNPM       ArtifactType = "npm"
	ArtifactTypeGolang    ArtifactType = "golang"
	ArtifactTypeAnsible   ArtifactType = "ansible"
	ArtifactTypeTerraform ArtifactType = "terraform"
	ArtifactTypeGeneric   ArtifactType = "generic"
	ArtifactTypeCargo     ArtifactType = "cargo"
	ArtifactTypeConan     ArtifactType = "conan"
	ArtifactTypeNuGet     ArtifactType = "nuget"
	ArtifactTypeRubyGems  ArtifactType = "rubygems"
)

// ArtifactInfo represents metadata about an artifact
type ArtifactInfo struct {
	Name         string            `json:"name"`
	Version      string            `json:"version"`
	Type         ArtifactType      `json:"type"`
	Path         string            `json:"path"`
	Size         int64             `json:"size"`
	Checksum     string            `json:"checksum"`
	UploadTime   time.Time         `json:"upload_time"`
	Metadata     map[string]string `json:"metadata"`
	Dependencies []string          `json:"dependencies,omitempty"`
}

// Artifact interface defines the contract for all artifact types
type Artifact interface {
	// GetType returns the artifact type
	GetType() ArtifactType

	// ParsePath extracts artifact information from a path
	ParsePath(path string) (*ArtifactInfo, error)

	// GeneratePath creates a storage path for the artifact
	GeneratePath(info *ArtifactInfo) string

	// ValidateArtifact validates the artifact content
	ValidateArtifact(content io.Reader) error

	// GetMetadata extracts metadata from artifact content
	GetMetadata(content io.Reader) (map[string]string, error)

	// GenerateIndex creates an index for multiple artifacts
	GenerateIndex(artifacts []*ArtifactInfo) ([]byte, error)

	// GetEndpoints returns the HTTP endpoints for this artifact type
	GetEndpoints() []string
}

// Metadata represents artifact metadata
type Metadata struct {
	Name        string            `json:"name"`
	Version     string            `json:"version"`
	Group       string            `json:"group,omitempty"`
	Description string            `json:"description,omitempty"`
	Size        int64             `json:"size"`
	Checksum    string            `json:"checksum"`
	CreatedAt   time.Time         `json:"created_at"`
	UpdatedAt   time.Time         `json:"updated_at"`
	PullCount   int64             `json:"pull_count"`
	PushCount   int64             `json:"push_count"`
	Properties  map[string]string `json:"properties,omitempty"`
}

// RepositoryConfig represents a repository configuration
type RepositoryConfig struct {
	Name         string            `json:"name"`
	Type         string            `json:"type"`
	ArtifactType ArtifactType      `json:"artifact_type"`
	URL          string            `json:"url,omitempty"`
	Upstreams    []string          `json:"upstreams,omitempty"`
	Properties   map[string]string `json:"properties,omitempty"`
}

// Statistics represents repository statistics
type Statistics struct {
	TotalArtifacts int64 `json:"total_artifacts"`
	TotalSize      int64 `json:"total_size"`
	PullCount      int64 `json:"pull_count"`
	PushCount      int64 `json:"push_count"`
}

// Repository represents an artifact repository
type Repository interface {
	// GetName returns repository name
	GetName() string
	
	// GetType returns repository type (local, remote, virtual)
	GetType() string
	
	// Pull retrieves an artifact
	Pull(ctx context.Context, path string) (io.ReadCloser, *ArtifactInfo, error)
	
	// Push stores an artifact
	Push(ctx context.Context, path string, content io.Reader, metadata *ArtifactInfo) error
	
	// Delete removes an artifact
	Delete(ctx context.Context, path string) error
	
	// List returns list of artifacts
	List(ctx context.Context, prefix string) ([]*ArtifactInfo, error)
	
	// GetIndex returns index for the repository
	GetIndex(ctx context.Context) (io.ReadCloser, error)
	
	// InvalidateCache invalidates cached artifacts
	InvalidateCache(ctx context.Context, path string) error
	
	// RebuildIndex rebuilds the repository index
	RebuildIndex(ctx context.Context) error
	
	// GetStatistics returns repository statistics
	GetStatistics(ctx context.Context) (*Statistics, error)
}

// Factory creates artifacts based on type
type Factory interface {
	CreateArtifact(artifactType ArtifactType) (Artifact, error)
	GetSupportedTypes() []ArtifactType
}
