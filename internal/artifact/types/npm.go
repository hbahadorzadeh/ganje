package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// NPMArtifact implements NPM artifact handling
type NPMArtifact struct {
	metadata *artifact.Metadata
}

// NewNPMArtifact creates a new NPM artifact
func NewNPMArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &NPMArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (n *NPMArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeNPM
}

// GetArtifactMetadata returns artifact metadata
func (n *NPMArtifact) GetArtifactMetadata() *artifact.Metadata {
	return n.metadata
}

// GetPath returns the storage path for NPM packages
func (n *NPMArtifact) GetPath() string {
	if strings.HasPrefix(n.metadata.Name, "@") {
		// Scoped package
		parts := strings.Split(n.metadata.Name, "/")
		scope := parts[0][1:] // Remove @
		name := parts[1]
		return fmt.Sprintf("%s/%s/-/%s-%s.tgz", scope, name, name, n.metadata.Version)
	}
	return fmt.Sprintf("%s/-/%s-%s.tgz", n.metadata.Name, n.metadata.Name, n.metadata.Version)
}

// GetIndexPath returns the index path for NPM registry
func (n *NPMArtifact) GetIndexPath() string {
	if strings.HasPrefix(n.metadata.Name, "@") {
		return strings.ReplaceAll(n.metadata.Name, "/", "%2F")
	}
	return n.metadata.Name
}

// ValidatePath validates NPM package path
func (n *NPMArtifact) ValidatePath(path string) error {
	// NPM path patterns:
	// Scoped: @scope/package/-/package-version.tgz
	// Unscoped: package/-/package-version.tgz
	patterns := []string{
		`^@[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/-/[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.tgz$`,
		`^[a-zA-Z0-9._-]+/-/[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.tgz$`,
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
	return fmt.Errorf("invalid NPM package path: %s", path)
}

// ParsePath parses NPM package information from path
func (n *NPMArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := n.ValidatePath(path); err != nil {
		return nil, err
	}

	var name, version string
	if strings.HasPrefix(path, "@") {
		// Scoped package
		parts := strings.Split(path, "/")
		scope := parts[0]
		packageName := parts[1]
		name = scope + "/" + packageName
		
		// Extract version from filename
		filename := parts[3]
		tarballName := strings.TrimSuffix(filename, ".tgz")
		versionPart := strings.TrimPrefix(tarballName, packageName+"-")
		version = versionPart
	} else {
		// Unscoped package
		parts := strings.Split(path, "/")
		name = parts[0]
		filename := parts[2]
		tarballName := strings.TrimSuffix(filename, ".tgz")
		versionPart := strings.TrimPrefix(tarballName, name+"-")
		version = versionPart
	}

	return &artifact.ArtifactInfo{
		Name:    name,
		Version: version,
		Type:    artifact.ArtifactTypeNPM,
		Path:    path,
		Metadata: map[string]string{
			"registry": "npm",
		},
	}, nil
}

// GeneratePath creates a storage path for the artifact
func (n *NPMArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	if strings.HasPrefix(info.Name, "@") {
		// Scoped package
		parts := strings.Split(info.Name, "/")
		scope := parts[0][1:] // Remove @
		name := parts[1]
		return fmt.Sprintf("%s/%s/-/%s-%s.tgz", scope, name, name, info.Version)
	}
	return fmt.Sprintf("%s/-/%s-%s.tgz", info.Name, info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (n *NPMArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (n *NPMArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "npm-package",
		"format": "tgz",
	}, nil
}

// GenerateIndex generates NPM package.json metadata
func (n *NPMArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type NPMVersion struct {
		Name    string `json:"name"`
		Version string `json:"version"`
		Dist    struct {
			Tarball string `json:"tarball"`
			Shasum  string `json:"shasum"`
		} `json:"dist"`
	}

	type NPMMetadata struct {
		Name     string                `json:"name"`
		Versions map[string]NPMVersion `json:"versions"`
		DistTags struct {
			Latest string `json:"latest"`
		} `json:"dist-tags"`
	}

	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	first := artifacts[0]
	metadata := NPMMetadata{
		Name:     first.Name,
		Versions: make(map[string]NPMVersion),
	}

	latest := ""
	for _, art := range artifacts {
		version := NPMVersion{
			Name:    art.Name,
			Version: art.Version,
		}
		version.Dist.Tarball = art.Path
		version.Dist.Shasum = art.Checksum
		
		metadata.Versions[art.Version] = version
		latest = art.Version
	}

	metadata.DistTags.Latest = latest

	return json.MarshalIndent(metadata, "", "  ")
}

// GetEndpoints returns NPM standard endpoints
func (n *NPMArtifact) GetEndpoints() []string {
	return []string{
		"GET /{package}",
		"GET /{package}/{version}",
		"GET /{package}/-/{filename}",
		"PUT /{package}",
	}
}
