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

// GetIndexPath returns the collection index path per Galaxy NG API v3
// /api/v3/plugin/ansible/content/published/collections/index/{namespace}/{name}/
func (a *AnsibleArtifact) GetIndexPath() string {
	return fmt.Sprintf("api/v3/plugin/ansible/content/published/collections/index/%s/%s/", a.metadata.Group, a.metadata.Name)
}

// ValidatePath validates Ansible collection path
func (a *AnsibleArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^download/[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.tar\.gz$`,
		`^api/v3/plugin/ansible/content/published/collections/index/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/$`,
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

// GenerateIndex generates collection metadata per Galaxy NG API v3 `Get a specific collection`
// Example schema fields: href, namespace, name, deprecated, versions_url, highest_version { href, version }, created_at, updated_at
func (a *AnsibleArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type HighestVersion struct {
		Href    string `json:"href"`
		Version string `json:"version"`
	}
	type Collection struct {
		Href          string         `json:"href"`
		Namespace     string         `json:"namespace"`
		Name          string         `json:"name"`
		Deprecated    bool           `json:"deprecated"`
		VersionsURL   string         `json:"versions_url"`
		Highest       *HighestVersion `json:"highest_version"`
		CreatedAt     string         `json:"created_at"`
		UpdatedAt     string         `json:"updated_at"`
	}

	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	first := artifacts[0]
	namespace := first.Metadata["namespace"]
	if namespace == "" {
		namespace = a.metadata.Group
	}
	if namespace == "" {
		namespace = "default"
	}
	name := first.Name
	base := fmt.Sprintf("/api/v3/plugin/ansible/content/published/collections/index/%s/%s/", namespace, name)

	// choose highest version lexicographically for determinism (simplified)
	highest := ""
	for _, art := range artifacts {
		if art.Name != name {
			continue
		}
		if highest == "" || art.Version > highest {
			highest = art.Version
		}
	}

	coll := Collection{
		Href:        base,
		Namespace:   namespace,
		Name:        name,
		Deprecated:  false,
		VersionsURL: base + "versions/",
		CreatedAt:   "",
		UpdatedAt:   "",
	}
	if highest != "" {
		coll.Highest = &HighestVersion{
			Href:    base + "versions/" + highest + "/",
			Version: highest,
		}
	}
	return json.Marshal(coll)
}

// GetEndpoints returns Ansible Galaxy standard endpoints
func (a *AnsibleArtifact) GetEndpoints() []string {
	return []string{
		"GET /api/v3/plugin/ansible/content/published/collections/index/{namespace}/{name}/",
		"GET /download/{namespace}-{name}-{version}.tar.gz",
		// API v3 uploads are different; keeping placeholder for completeness
		"POST /api/v3/plugin/ansible/content/published/collections/",
	}
}
