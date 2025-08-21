package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// DockerArtifact implements Docker artifact handling
type DockerArtifact struct {
	metadata *artifact.Metadata
}

// NewDockerArtifact creates a new Docker artifact
func NewDockerArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &DockerArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (d *DockerArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeDocker
}

// GetArtifactMetadata returns artifact metadata
func (d *DockerArtifact) GetArtifactMetadata() *artifact.Metadata {
	return d.metadata
}

// GetPath returns the storage path for Docker images
func (d *DockerArtifact) GetPath() string {
	return fmt.Sprintf("v2/%s/manifests/%s", d.metadata.Name, d.metadata.Version)
}

// GetIndexPath returns the index path for Docker registry
func (d *DockerArtifact) GetIndexPath() string {
	return fmt.Sprintf("v2/%s/tags/list", d.metadata.Name)
}

// ValidatePath validates Docker image path
func (d *DockerArtifact) ValidatePath(path string) error {
	// Docker registry API v2 paths
	patterns := []string{
		`^v2/[a-zA-Z0-9._/-]+/manifests/[a-zA-Z0-9._-]+$`,
		`^v2/[a-zA-Z0-9._/-]+/blobs/sha256:[a-f0-9]{64}$`,
		`^v2/[a-zA-Z0-9._/-]+/tags/list$`,
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
	return fmt.Errorf("invalid Docker registry path: %s", path)
}

// ParsePath parses Docker image information from path
func (d *DockerArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := d.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.Contains(path, "/manifests/") {
		parts := strings.Split(path, "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid Docker manifest path")
		}
		
		name := strings.Join(parts[1:len(parts)-2], "/")
		tag := parts[len(parts)-1]
		
		return &artifact.ArtifactInfo{
			Name:    name,
			Version: tag,
			Type:    artifact.ArtifactTypeDocker,
			Path:    path,
			Metadata: map[string]string{
				"type": "manifest",
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported Docker path type")
}

// GeneratePath creates a storage path for the artifact
func (d *DockerArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("v2/%s/manifests/%s", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (d *DockerArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (d *DockerArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "docker-image",
		"format": "manifest",
	}, nil
}

// GenerateIndex generates Docker tags list
func (d *DockerArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type DockerTagsList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	first := artifacts[0]
	tagsList := DockerTagsList{
		Name: first.Name,
		Tags: make([]string, 0, len(artifacts)),
	}

	for _, art := range artifacts {
		tagsList.Tags = append(tagsList.Tags, art.Version)
	}

	return json.MarshalIndent(tagsList, "", "  ")
}

// GetEndpoints returns Docker registry standard endpoints
func (d *DockerArtifact) GetEndpoints() []string {
	return []string{
		"GET /v2/",
		"GET /v2/{name}/tags/list",
		"GET /v2/{name}/manifests/{reference}",
		"PUT /v2/{name}/manifests/{reference}",
		"GET /v2/{name}/blobs/{digest}",
		"PUT /v2/{name}/blobs/uploads/",
	}
}
