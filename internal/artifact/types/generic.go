package types

import (
	"encoding/json"
	"fmt"
	"io"
	"path/filepath"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// GenericArtifact implements generic artifact handling
type GenericArtifact struct {
	metadata *artifact.Metadata
}

// NewGenericArtifact creates a new generic artifact
func NewGenericArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &GenericArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (g *GenericArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeGeneric
}

// GetArtifactMetadata returns artifact metadata
func (g *GenericArtifact) GetArtifactMetadata() *artifact.Metadata {
	return g.metadata
}

// GetPath returns the storage path for generic artifacts
func (g *GenericArtifact) GetPath() string {
	if g.metadata.Group != "" {
		return fmt.Sprintf("%s/%s/%s", g.metadata.Group, g.metadata.Name, g.metadata.Version)
	}
	return fmt.Sprintf("%s/%s", g.metadata.Name, g.metadata.Version)
}

// GetIndexPath returns the index path for generic artifacts
func (g *GenericArtifact) GetIndexPath() string {
	return "index.json"
}

// ValidatePath validates generic artifact path
func (g *GenericArtifact) ValidatePath(path string) error {
	// Generic artifacts can have any valid file path
	if strings.TrimSpace(path) == "" {
		return fmt.Errorf("path cannot be empty")
	}
	if strings.Contains(path, "..") {
		return fmt.Errorf("path cannot contain '..'")
	}
	return nil
}

// ParsePath parses generic artifact information from path
func (g *GenericArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := g.ValidatePath(path); err != nil {
		return nil, err
	}

	parts := strings.Split(path, "/")
	filename := parts[len(parts)-1]
	ext := filepath.Ext(filename)
	
	var name, version, group string
	
	if len(parts) >= 3 {
		group = parts[0]
		name = parts[1]
		version = parts[2]
	} else if len(parts) == 2 {
		name = parts[0]
		version = parts[1]
	} else {
		name = filename
		version = "latest"
	}

	return &artifact.ArtifactInfo{
		Name:    name,
		Version: version,
		Type:    artifact.ArtifactTypeGeneric,
		Path:    path,
		Metadata: map[string]string{
			"filename":  filename,
			"extension": ext,
			"group":     group,
		},
	}, nil
}

// GeneratePath creates a storage path for the artifact
func (g *GenericArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	group := info.Metadata["group"]
	if group != "" {
		return fmt.Sprintf("%s/%s/%s", group, info.Name, info.Version)
	}
	return fmt.Sprintf("%s/%s", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (g *GenericArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (g *GenericArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "generic-artifact",
	}, nil
}

// GenerateIndex generates generic artifact index
func (g *GenericArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type GenericIndex struct {
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

	index := GenericIndex{
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

// GetEndpoints returns generic artifact standard endpoints
func (g *GenericArtifact) GetEndpoints() []string {
	return []string{
		"GET /{path:.*}",
		"PUT /{path:.*}",
		"DELETE /{path:.*}",
		"GET /index.json",
	}
}
