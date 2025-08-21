package types

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// GoModuleArtifact implements Go module artifact handling
type GoModuleArtifact struct {
	metadata *artifact.Metadata
}

// NewGoModuleArtifact creates a new Go module artifact
func NewGoModuleArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &GoModuleArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (g *GoModuleArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeGolang
}

// GetArtifactMetadata returns artifact metadata
func (g *GoModuleArtifact) GetArtifactMetadata() *artifact.Metadata {
	return g.metadata
}

// GetPath returns the storage path for Go modules
func (g *GoModuleArtifact) GetPath() string {
	return fmt.Sprintf("%s/@v/%s.zip", g.metadata.Name, g.metadata.Version)
}

// GetIndexPath returns the index path for Go modules
func (g *GoModuleArtifact) GetIndexPath() string {
	return fmt.Sprintf("%s/@v/list", g.metadata.Name)
}

// ValidatePath validates Go module path
func (g *GoModuleArtifact) ValidatePath(path string) error {
	// Go module proxy paths
	patterns := []string{
		`^[a-zA-Z0-9._/-]+/@v/v[0-9]+\.[0-9]+\.[0-9]+.*\.zip$`,
		`^[a-zA-Z0-9._/-]+/@v/v[0-9]+\.[0-9]+\.[0-9]+.*\.info$`,
		`^[a-zA-Z0-9._/-]+/@v/v[0-9]+\.[0-9]+\.[0-9]+.*\.mod$`,
		`^[a-zA-Z0-9._/-]+/@v/list$`,
		`^[a-zA-Z0-9._/-]+/@latest$`,
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, path)
		if err != nil {
			return err
		}
		if matched {
			return nil
		}
	}
	return fmt.Errorf("invalid Go module path: %s", path)
}

// ParsePath parses Go module information from path
func (g *GoModuleArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := g.ValidatePath(path); err != nil {
		return nil, err
	}

	parts := strings.Split(path, "/@v/")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid Go module path structure")
	}

	moduleName := parts[0]
	versionPart := parts[1]

	if strings.HasSuffix(versionPart, ".zip") {
		version := strings.TrimSuffix(versionPart, ".zip")
		return &artifact.ArtifactInfo{
			Name:    moduleName,
			Version: version,
			Type:    artifact.ArtifactTypeGolang,
			Path:    path,
			Metadata: map[string]string{
				"type": "module",
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported Go module path type")
}

// GeneratePath creates a storage path for the artifact
func (g *GoModuleArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("%s/@v/%s.zip", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (g *GoModuleArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (g *GoModuleArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "go-module",
		"format": "zip",
	}, nil
}

// GenerateIndex generates Go module version list
func (g *GoModuleArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	versions := make([]string, 0, len(artifacts))
	for _, art := range artifacts {
		versions = append(versions, art.Version)
	}

	return []byte(strings.Join(versions, "\n")), nil
}

// GetEndpoints returns Go module proxy standard endpoints
func (g *GoModuleArtifact) GetEndpoints() []string {
	return []string{
		"GET /{module}/@v/list",
		"GET /{module}/@v/{version}.info",
		"GET /{module}/@v/{version}.mod",
		"GET /{module}/@v/{version}.zip",
		"GET /{module}/@latest",
	}
}
