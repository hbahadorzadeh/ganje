package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// RubyGemsArtifact implements RubyGems artifact handling
type RubyGemsArtifact struct {
	metadata *artifact.Metadata
}

// NewRubyGemsArtifact creates a new RubyGems artifact
func NewRubyGemsArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &RubyGemsArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (r *RubyGemsArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeRubyGems
}

// GetArtifactMetadata returns artifact metadata
func (r *RubyGemsArtifact) GetArtifactMetadata() *artifact.Metadata {
	return r.metadata
}

// GetPath returns the storage path for RubyGems
func (r *RubyGemsArtifact) GetPath() string {
	return fmt.Sprintf("gems/%s-%s.gem", r.metadata.Name, r.metadata.Version)
}

// GetIndexPath returns the index path for RubyGems
func (r *RubyGemsArtifact) GetIndexPath() string {
	return "specs.4.8.gz"
}

// ValidatePath validates RubyGems path
func (r *RubyGemsArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^gems/[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.gem$`,
		`^specs\.4\.8\.gz$`,
		`^latest_specs\.4\.8\.gz$`,
		`^prerelease_specs\.4\.8\.gz$`,
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
	return fmt.Errorf("invalid RubyGems path: %s", path)
}

// ParsePath parses RubyGems information from path
func (r *RubyGemsArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := r.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.HasPrefix(path, "gems/") && strings.HasSuffix(path, ".gem") {
		filename := strings.TrimPrefix(path, "gems/")
		gemName := strings.TrimSuffix(filename, ".gem")
		
		// Find last dash to separate name and version
		lastDash := strings.LastIndex(gemName, "-")
		if lastDash == -1 {
			return nil, fmt.Errorf("invalid gem filename format")
		}
		
		name := gemName[:lastDash]
		version := gemName[lastDash+1:]
		
		return &artifact.ArtifactInfo{
			Name:    name,
			Version: version,
			Type:    artifact.ArtifactTypeRubyGems,
			Path:    path,
			Metadata: map[string]string{
				"filename": filename,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported RubyGems path type")
}

// GeneratePath creates a storage path for the artifact
func (r *RubyGemsArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("gems/%s-%s.gem", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (r *RubyGemsArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (r *RubyGemsArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "ruby-gem",
		"format": "gem",
	}, nil
}

// GenerateIndex generates RubyGems specs index
func (r *RubyGemsArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type GemSpec struct {
		Name     string `json:"name"`
		Version  string `json:"version"`
		Platform string `json:"platform"`
	}

	specs := make([][]interface{}, 0, len(artifacts))
	for _, art := range artifacts {
		spec := []interface{}{
			art.Name,
			art.Version,
			"ruby",
		}
		specs = append(specs, spec)
	}

	return json.Marshal(specs)
}

// GetEndpoints returns RubyGems standard endpoints
func (r *RubyGemsArtifact) GetEndpoints() []string {
	return []string{
		"GET /specs.4.8.gz",
		"GET /latest_specs.4.8.gz",
		"GET /prerelease_specs.4.8.gz",
		"GET /gems/{name}-{version}.gem",
		"POST /api/v1/gems",
	}
}
