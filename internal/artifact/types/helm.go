package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// HelmArtifact implements Helm chart artifact handling
type HelmArtifact struct {
	metadata *artifact.Metadata
}

// NewHelmArtifact creates a new Helm artifact
func NewHelmArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &HelmArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (h *HelmArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeHelm
}

// GetArtifactMetadata returns artifact metadata
func (h *HelmArtifact) GetArtifactMetadata() *artifact.Metadata {
	return h.metadata
}

// GetPath returns the storage path for Helm charts
func (h *HelmArtifact) GetPath() string {
	return fmt.Sprintf("%s-%s.tgz", h.metadata.Name, h.metadata.Version)
}

// GetIndexPath returns the index path for Helm repository
func (h *HelmArtifact) GetIndexPath() string {
	return "index.yaml"
}

// ValidatePath validates Helm chart path
func (h *HelmArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.tgz$`,
		`^index\.yaml$`,
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
	return fmt.Errorf("invalid Helm chart path: %s", path)
}

// ParsePath parses Helm chart information from path
func (h *HelmArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := h.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.HasSuffix(path, ".tgz") {
		filename := strings.TrimSuffix(path, ".tgz")
		parts := strings.Split(filename, "-")
		if len(parts) < 2 {
			return nil, fmt.Errorf("invalid Helm chart filename format")
		}
		
		// Last part is version, everything before is name
		version := parts[len(parts)-1]
		name := strings.Join(parts[:len(parts)-1], "-")
		
		return &artifact.ArtifactInfo{
			Name:    name,
			Version: version,
			Type:    artifact.ArtifactTypeHelm,
			Path:    path,
			Metadata: map[string]string{
				"filename": path,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported Helm path type")
}

// GeneratePath creates a storage path for the artifact
func (h *HelmArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("%s-%s.tgz", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (h *HelmArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (h *HelmArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "helm-chart",
		"format": "tgz",
	}, nil
}

// GenerateIndex generates Helm repository index.yaml
func (h *HelmArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type HelmChart struct {
		APIVersion  string `yaml:"apiVersion"`
		Name        string `yaml:"name"`
		Version     string `yaml:"version"`
		Description string `yaml:"description,omitempty"`
		Created     string `yaml:"created"`
		Digest      string `yaml:"digest"`
		URLs        []string `yaml:"urls"`
	}

	type HelmIndex struct {
		APIVersion string                       `yaml:"apiVersion"`
		Generated  string                       `yaml:"generated"`
		Entries    map[string][]HelmChart      `yaml:"entries"`
	}

	index := HelmIndex{
		APIVersion: "v1",
		Generated:  "2023-01-01T00:00:00Z",
		Entries:    make(map[string][]HelmChart),
	}

	for _, art := range artifacts {
		chart := HelmChart{
			APIVersion: "v2",
			Name:       art.Name,
			Version:    art.Version,
			Created:    art.UploadTime.Format("2006-01-02T15:04:05Z"),
			Digest:     art.Checksum,
			URLs:       []string{art.Path},
		}

		if index.Entries[art.Name] == nil {
			index.Entries[art.Name] = []HelmChart{}
		}
		index.Entries[art.Name] = append(index.Entries[art.Name], chart)
	}

	// Convert to YAML (using JSON as intermediate for simplicity)
	jsonData, err := json.Marshal(index)
	if err != nil {
		return nil, err
	}

	return jsonData, nil
}

// GetEndpoints returns Helm repository standard endpoints
func (h *HelmArtifact) GetEndpoints() []string {
	return []string{
		"GET /index.yaml",
		"GET /{chart}-{version}.tgz",
		"POST /api/charts",
		"DELETE /api/charts/{name}/{version}",
	}
}
