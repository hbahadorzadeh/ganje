package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// ConanArtifact implements Conan artifact handling
type ConanArtifact struct {
	metadata *artifact.Metadata
}

// NewConanArtifact creates a new Conan artifact
func NewConanArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &ConanArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (c *ConanArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeConan
}

// GetArtifactMetadata returns artifact metadata
func (c *ConanArtifact) GetArtifactMetadata() *artifact.Metadata {
	return c.metadata
}

// GetPath returns the storage path for Conan artifacts
func (c *ConanArtifact) GetPath() string {
	if c.metadata.Group != "" {
		return fmt.Sprintf("conan/%s/%s/%s", c.metadata.Group, c.metadata.Name, c.metadata.Version)
	}
	return fmt.Sprintf("conan/%s/%s", c.metadata.Name, c.metadata.Version)
}

// GetIndexPath returns the index path for Conan artifacts
func (c *ConanArtifact) GetIndexPath() string {
	return "conan/index.json"
}

// ValidatePath validates Conan artifact path
func (c *ConanArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^conan/[a-zA-Z0-9._/-]+$`,
		`^conan/index\.json$`,
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
	return fmt.Errorf("invalid Conan artifact path: %s", path)
}

// ParsePath parses Conan artifact information from path
func (c *ConanArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := c.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.HasPrefix(path, "conan/") && !strings.HasSuffix(path, "index.json") {
		pathParts := strings.TrimPrefix(path, "conan/")
		parts := strings.Split(pathParts, "/")
		
		var name, version, group string
		if len(parts) >= 3 {
			group = parts[0]
			name = parts[1]
			version = parts[2]
		} else if len(parts) == 2 {
			name = parts[0]
			version = parts[1]
		} else {
			return nil, fmt.Errorf("invalid Conan path structure")
		}
		
		return &artifact.ArtifactInfo{
			Name:    name,
			Version: version,
			Type:    artifact.ArtifactTypeConan,
			Path:    path,
			Metadata: map[string]string{
				"group": group,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported Conan path type")
}

// GeneratePath creates a storage path for the artifact
func (c *ConanArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	group := info.Metadata["group"]
	if group != "" {
		return fmt.Sprintf("conan/%s/%s/%s", group, info.Name, info.Version)
	}
	return fmt.Sprintf("conan/%s/%s", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (c *ConanArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (c *ConanArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "conan-artifact",
	}, nil
}

// GenerateIndex generates Conan artifact index
func (c *ConanArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type ConanIndex struct {
		Artifacts []struct {
			Name     string            `json:"name"`
			Version  string            `json:"version"`
			Group    string            `json:"group,omitempty"`
			Path     string            `json:"path"`
			Size     int64             `json:"size"`
			Checksum string            `json:"checksum"`
			Props    map[string]string `json:"properties,omitempty"`
		} `json:"artifacts"`
	}

	index := ConanIndex{
		Artifacts: make([]struct {
			Name     string            `json:"name"`
			Version  string            `json:"version"`
			Group    string            `json:"group,omitempty"`
			Path     string            `json:"path"`
			Size     int64             `json:"size"`
			Checksum string            `json:"checksum"`
			Props    map[string]string `json:"properties,omitempty"`
		}, 0, len(artifacts)),
	}

	for _, art := range artifacts {
		group := art.Metadata["group"]
		artifact := struct {
			Name     string            `json:"name"`
			Version  string            `json:"version"`
			Group    string            `json:"group,omitempty"`
			Path     string            `json:"path"`
			Size     int64             `json:"size"`
			Checksum string            `json:"checksum"`
			Props    map[string]string `json:"properties,omitempty"`
		}{
			Name:     art.Name,
			Version:  art.Version,
			Group:    group,
			Path:     art.Path,
			Size:     art.Size,
			Checksum: art.Checksum,
			Props:    art.Metadata,
		}
		index.Artifacts = append(index.Artifacts, artifact)
	}

	return json.MarshalIndent(index, "", "  ")
}

// GetEndpoints returns Conan standard endpoints
func (c *ConanArtifact) GetEndpoints() []string {
	return []string{
		"GET /conan/{path:.*}",
		"PUT /conan/{path:.*}",
		"DELETE /conan/{path:.*}",
		"GET /conan/index.json",
	}
}
