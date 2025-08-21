package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// AnsibleArtifact implements Ansible Galaxy artifact handling
type AnsibleArtifact struct {
	metadata *artifact.Metadata
}

// NewAnsibleArtifact creates a new Ansible artifact
func NewAnsibleArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &AnsibleArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (a *AnsibleArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeAnsible
}

// GetArtifactMetadata returns artifact metadata
func (a *AnsibleArtifact) GetArtifactMetadata() *artifact.Metadata {
	return a.metadata
}

// GetPath returns the storage path for Ansible collections
func (a *AnsibleArtifact) GetPath() string {
	return fmt.Sprintf("download/%s-%s-%s.tar.gz", a.metadata.Group, a.metadata.Name, a.metadata.Version)
}

// GetIndexPath returns the index path for Ansible Galaxy
func (a *AnsibleArtifact) GetIndexPath() string {
	return fmt.Sprintf("api/v2/collections/%s/%s/", a.metadata.Group, a.metadata.Name)
}

// ValidatePath validates Ansible collection path
func (a *AnsibleArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^download/[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.tar\.gz$`,
		`^api/v2/collections/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/$`,
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
	return fmt.Errorf("invalid Ansible collection path: %s", path)
}

// ParsePath parses Ansible collection information from path
func (a *AnsibleArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := a.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.HasPrefix(path, "download/") && strings.HasSuffix(path, ".tar.gz") {
		filename := strings.TrimPrefix(path, "download/")
		filename = strings.TrimSuffix(filename, ".tar.gz")
		
		parts := strings.Split(filename, "-")
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid Ansible collection filename format")
		}
		
		namespace := parts[0]
		name := parts[1]
		version := strings.Join(parts[2:], "-")
		
		return &artifact.ArtifactInfo{
			Name:    name,
			Version: version,
			Type:    artifact.ArtifactTypeAnsible,
			Path:    path,
			Metadata: map[string]string{
				"namespace": namespace,
				"filename": filename + ".tar.gz",
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported Ansible path type")
}

// GeneratePath creates a storage path for the artifact
func (a *AnsibleArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	namespace := info.Metadata["namespace"]
	if namespace == "" {
		namespace = "default"
	}
	return fmt.Sprintf("download/%s-%s-%s.tar.gz", namespace, info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (a *AnsibleArtifact) ValidateArtifact(content io.Reader) error {
	// Basic validation - check if content is readable
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (a *AnsibleArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "ansible-collection",
		"format": "tar.gz",
	}, nil
}

// GenerateIndex generates Ansible Galaxy collection metadata
func (a *AnsibleArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type AnsibleVersion struct {
		Version string `json:"version"`
		Href    string `json:"href"`
	}

	type AnsibleCollection struct {
		Namespace   string           `json:"namespace"`
		Name        string           `json:"name"`
		Description string           `json:"description"`
		Versions    []AnsibleVersion `json:"versions"`
	}

	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	first := artifacts[0]
	namespace := first.Metadata["namespace"]
	if namespace == "" {
		namespace = "default"
	}
	collection := AnsibleCollection{
		Namespace:   namespace,
		Name:        first.Name,
		Description: "Ansible Galaxy Collection",
		Versions:    make([]AnsibleVersion, 0, len(artifacts)),
	}

	for _, art := range artifacts {
		artNamespace := art.Metadata["namespace"]
		if artNamespace == "" {
			artNamespace = "default"
		}
		version := AnsibleVersion{
			Version: art.Version,
			Href:    fmt.Sprintf("/download/%s-%s-%s.tar.gz", artNamespace, art.Name, art.Version),
		}
		collection.Versions = append(collection.Versions, version)
	}

	return json.MarshalIndent(collection, "", "  ")
}

// GetEndpoints returns Ansible Galaxy standard endpoints
func (a *AnsibleArtifact) GetEndpoints() []string {
	return []string{
		"GET /api/v2/collections/{namespace}/{name}/",
		"GET /download/{namespace}-{name}-{version}.tar.gz",
		"POST /api/v2/collections/",
	}
}
