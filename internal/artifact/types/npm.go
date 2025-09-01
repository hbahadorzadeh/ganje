package types

import (
	"bytes"
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

var npmMagicBytes = []byte{0x1f, 0x8b, 0x08}

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
		if len(parts) >= 2 {
			scope := strings.TrimPrefix(parts[0], "@")
			name := parts[1]
			return fmt.Sprintf("%s/%s/-/%s-%s.tgz", scope, name, name, n.metadata.Version)
		}
	}
	return fmt.Sprintf("%s/-/%s-%s.tgz", n.metadata.Name, n.metadata.Name, n.metadata.Version)
}

// GetIndexPath returns the index path for NPM registry
func (n *NPMArtifact) GetIndexPath() string {
	if strings.HasPrefix(n.metadata.Name, "@") {
		// Encode '/' as %2F for scoped packages per npm registry
		return strings.ReplaceAll(n.metadata.Name, "/", "%2F")
	}
	return n.metadata.Name
}

// ValidatePath validates NPM package path
func (n *NPMArtifact) ValidatePath(path string) error {
	// Normalize: decode %2F in leading package segment when present
	decoded := path
	// Only decode package segment if present in encoded scoped form
	if strings.HasPrefix(decoded, "@") && strings.Contains(decoded, "%2F") {
		decoded = strings.ReplaceAll(decoded, "%2F", "/")
		decoded = strings.ReplaceAll(decoded, "%2f", "/")
	}

	// NPM path patterns:
	// Scoped: @scope/package/-/package-version.tgz
	// Unscoped: package/-/package-version.tgz
	// Package and scope identifiers allow: a-z0-9._- (lowercase recommended). We'll accept case-insensitive.
	patterns := []string{
		`^@[A-Za-z0-9._-]+/[A-Za-z0-9._-]+/-/[A-Za-z0-9._-]+-[0-9A-Za-z._-]+\.tgz$`,
		`^[A-Za-z0-9._-]+/-/[A-Za-z0-9._-]+-[0-9A-Za-z._-]+\.tgz$`,
	}

	for _, pattern := range patterns {
		matched, err := regexp.MatchString(pattern, decoded)
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
	// Decode scoped package %2F if present for parsing
	normalized := path
	if strings.HasPrefix(normalized, "@") && strings.Contains(normalized, "%2F") {
		normalized = strings.ReplaceAll(normalized, "%2F", "/")
		normalized = strings.ReplaceAll(normalized, "%2f", "/")
	}

	if err := n.ValidatePath(normalized); err != nil {
		return nil, err
	}

	var name, version string
	if strings.HasPrefix(normalized, "@") {
		// Scoped package
		parts := strings.Split(normalized, "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid NPM path: %s", path)
		}
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
		parts := strings.Split(normalized, "/")
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid NPM path: %s", path)
		}
		name = parts[0]
		filename := parts[2]
		tarballName := strings.TrimSuffix(filename, ".tgz")
		versionPart := strings.TrimPrefix(tarballName, name+"-")
		version = versionPart
	}

	// Ensure index path is encoded for scoped name
	indexPath := name
	if strings.HasPrefix(name, "@") {
		indexPath = strings.ReplaceAll(name, "/", "%2F")
	}

	return &artifact.ArtifactInfo{
		Name:    name,
		Version: version,
		Type:    artifact.ArtifactTypeNPM,
		Path:    normalized,
		Metadata: map[string]string{
			"registry":  "npm",
			"extension": ".tgz",
			"indexPath": indexPath,
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
	header := make([]byte, 3)
	if _, err := io.ReadFull(content, header); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return fmt.Errorf("invalid artifact content: too short")
		}
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	if !bytes.Equal(header, npmMagicBytes) {
		return fmt.Errorf("invalid artifact content: not a gzip file")
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (n *NPMArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type":   "npm-package",
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
		// Ensure tarball URL path uses encoded scoped name if applicable
		tarballPath := art.Path
		if strings.HasPrefix(art.Name, "@") {
			// Build path like @scope/name/-/name-version.tgz but with %2F in name segment if needed by clients
			encodedName := strings.ReplaceAll(art.Name, "/", "%2F")
			// Attempt to reconstruct tarball filename
			baseName := strings.Split(art.Name, "/")[1]
			tarballPath = fmt.Sprintf("%s/-/%s-%s.tgz", encodedName, baseName, art.Version)
		}
		// Prepend leading slash if missing to form a path
		if !strings.HasPrefix(tarballPath, "/") {
			tarballPath = "/" + tarballPath
		}
		// Tarball should be a URL; if repository base URL is needed, upstream code can prefix. Store path here.
		version.Dist.Tarball = tarballPath
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
		"GET /{package}/versions/{version}", // Added version endpoint
	}
}
