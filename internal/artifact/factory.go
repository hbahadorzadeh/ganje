package artifact

import (
	"io"
	"path/filepath"
)

// BasicArtifact provides a simple implementation of the Artifact interface
type BasicArtifact struct {
	artifactType ArtifactType
}

// GetType returns the artifact type
func (b *BasicArtifact) GetType() ArtifactType {
	return b.artifactType
}

// ParsePath extracts artifact information from a path
func (b *BasicArtifact) ParsePath(path string) (*ArtifactInfo, error) {
	return &ArtifactInfo{
		Name:    filepath.Base(path),
		Version: "1.0.0",
		Type:    b.artifactType,
		Path:    path,
	}, nil
}

// GeneratePath creates a storage path for the artifact
func (b *BasicArtifact) GeneratePath(info *ArtifactInfo) string {
	return filepath.Join(string(info.Type), info.Name, info.Version, filepath.Base(info.Path))
}

// ValidateArtifact validates the artifact content
func (b *BasicArtifact) ValidateArtifact(content io.Reader) error {
	return nil // Basic validation - always passes
}

// GetMetadata extracts metadata from artifact content
func (b *BasicArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": string(b.artifactType),
	}, nil
}

// GenerateIndex creates an index for multiple artifacts
func (b *BasicArtifact) GenerateIndex(artifacts []*ArtifactInfo) ([]byte, error) {
	return []byte("{}"), nil // Basic empty JSON index
}

// GetEndpoints returns the HTTP endpoints for this artifact type
func (b *BasicArtifact) GetEndpoints() []string {
	return []string{
		"/{repo}/*path",
	}
}

// DefaultFactory implements the Factory interface
type DefaultFactory struct{}

// NewFactory creates a new artifact factory
func NewFactory() Factory {
	return &DefaultFactory{}
}

// CreateArtifact creates an artifact instance based on type
func (f *DefaultFactory) CreateArtifact(artifactType ArtifactType) (Artifact, error) {
	return &BasicArtifact{artifactType: artifactType}, nil
}

// GetSupportedTypes returns all supported artifact types
func (f *DefaultFactory) GetSupportedTypes() []ArtifactType {
	return []ArtifactType{
		ArtifactTypeMaven, ArtifactTypePyPI, ArtifactTypeHelm, ArtifactTypeDocker, ArtifactTypeNPM, ArtifactTypeGolang,
		ArtifactTypeAnsible, ArtifactTypeTerraform, ArtifactTypeGeneric, ArtifactTypeCargo,
		ArtifactTypeNuGet, ArtifactTypeRubyGems, ArtifactTypeBazel,
	}
}
