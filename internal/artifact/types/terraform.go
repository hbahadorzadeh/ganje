package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// TerraformArtifact implements Terraform module artifact handling
type TerraformArtifact struct {
	metadata *artifact.Metadata
}

// NewTerraformArtifact creates a new Terraform artifact
func NewTerraformArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &TerraformArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (t *TerraformArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeTerraform
}

// GetArtifactMetadata returns artifact metadata
func (t *TerraformArtifact) GetArtifactMetadata() *artifact.Metadata {
	return t.metadata
}

// GetPath returns the storage path for Terraform modules
func (t *TerraformArtifact) GetPath() string {
	provider := t.metadata.Properties["provider"]
	if provider == "" {
		provider = "provider"
	}
	return fmt.Sprintf("v1/modules/%s/%s/%s/%s/download", t.metadata.Group, t.metadata.Name, provider, t.metadata.Version)
}

// GetIndexPath returns the index path for Terraform registry
func (t *TerraformArtifact) GetIndexPath() string {
	provider := t.metadata.Properties["provider"]
	if provider == "" {
		provider = "provider"
	}
	return fmt.Sprintf("v1/modules/%s/%s/%s/versions", t.metadata.Group, t.metadata.Name, provider)
}

// ValidatePath validates Terraform module path
func (t *TerraformArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^v1/modules/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/download$`,
		`^v1/modules/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/versions$`,
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
	return fmt.Errorf("invalid Terraform module path: %s", path)
}

// ParsePath parses Terraform module information from path
func (t *TerraformArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := t.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.Contains(path, "/download") {
		parts := strings.Split(path, "/")
		if len(parts) < 7 {
			return nil, fmt.Errorf("invalid Terraform download path")
		}
		// v1/modules/{namespace}/{name}/{provider}/{version}/download
		namespace := parts[2]
		name := parts[3]
		provider := parts[4]
		version := parts[5]

		return &artifact.ArtifactInfo{
			Name:    name,
			Version: version,
			Type:    artifact.ArtifactTypeTerraform,
			Path:    path,
			Metadata: map[string]string{
				"namespace": namespace,
				"provider":  provider,
				"type": "module",
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported Terraform path type")
}

// GeneratePath creates a storage path for the artifact
func (t *TerraformArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	namespace := info.Metadata["namespace"]
	if namespace == "" {
		namespace = "default"
	}
	provider := info.Metadata["provider"]
	if provider == "" {
		provider = "provider"
	}
	return fmt.Sprintf("v1/modules/%s/%s/%s/%s/download", namespace, info.Name, provider, info.Version)
}

// ValidateArtifact validates the artifact content
func (t *TerraformArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (t *TerraformArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "terraform-module",
		"format": "tar.gz",
	}, nil
}

// GenerateIndex generates Terraform module versions
func (t *TerraformArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type TerraformVersion struct {
		Version string `json:"version"`
	}

	type TerraformModule struct {
		Source   string             `json:"source"`
		Versions []TerraformVersion `json:"versions"`
	}

	type TerraformVersions struct {
		Modules []TerraformModule `json:"modules"`
	}

	// Derive source from first artifact if present: namespace/name/provider
	source := ""
	if len(artifacts) > 0 {
		ns := artifacts[0].Metadata["namespace"]
		name := artifacts[0].Name
		prov := artifacts[0].Metadata["provider"]
		if ns != "" && name != "" && prov != "" {
			source = fmt.Sprintf("%s/%s/%s", ns, name, prov)
		}
	}

	module := TerraformModule{Source: source, Versions: make([]TerraformVersion, 0, len(artifacts))}
	for _, art := range artifacts {
		module.Versions = append(module.Versions, TerraformVersion{Version: art.Version})
	}

	payload := TerraformVersions{Modules: []TerraformModule{module}}
	return json.MarshalIndent(payload, "", "  ")
}

// GetEndpoints returns Terraform registry standard endpoints
func (t *TerraformArtifact) GetEndpoints() []string {
	return []string{
		"GET /v1/modules/{namespace}/{name}/{provider}/versions",
		"GET /v1/modules/{namespace}/{name}/{provider}/{version}/download",
		"POST /v1/modules",
	}
}
