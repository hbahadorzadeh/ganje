package types

import (
	"encoding/xml"
	"fmt"
	"io"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// MavenArtifact implements Maven artifact handling
type MavenArtifact struct {
	metadata *artifact.Metadata
}

// NewMavenArtifact creates a new Maven artifact
func NewMavenArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &MavenArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (m *MavenArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeMaven
}

// GetArtifactMetadata returns artifact metadata
func (m *MavenArtifact) GetArtifactMetadata() *artifact.Metadata {
	return m.metadata
}

// GetPath returns the storage path for Maven artifacts
func (m *MavenArtifact) GetPath() string {
	group := strings.ReplaceAll(m.metadata.Group, ".", "/")
	return fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, m.metadata.Name, m.metadata.Version, m.metadata.Name, m.metadata.Version)
}

// ValidateArtifact validates Maven artifact content
func (m *MavenArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetIndexPath returns the index path for Maven repositories
func (m *MavenArtifact) GetIndexPath() string {
	return "maven-metadata.xml"
}

// ValidatePath validates Maven artifact path
func (m *MavenArtifact) ValidatePath(path string) error {
	// Maven path pattern: groupId/artifactId/version/artifactId-version.extension
	pattern := `^[a-zA-Z0-9._-]+(/[a-zA-Z0-9._-]+)+/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.(jar|pom|war|ear)$`
	matched, err := regexp.MatchString(pattern, path)
	if err != nil {
		return err
	}
	if !matched {
		return fmt.Errorf("invalid Maven artifact path: %s", path)
	}
	return nil
}

// ParsePath parses Maven artifact information from path
func (m *MavenArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := m.ValidatePath(path); err != nil {
		return nil, err
	}

	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid Maven path structure")
	}

	version := parts[len(parts)-2]
	filename := parts[len(parts)-1]
	artifactId := parts[len(parts)-3]
	groupId := strings.Join(parts[:len(parts)-3], ".")

	// Extract artifact name and version from filename
	ext := filepath.Ext(filename)
	_ = strings.TrimSuffix(filename, ext) // nameVersion not used currently
	
	return &artifact.ArtifactInfo{
		Name:    artifactId,
		Version: version,
		Type:    artifact.ArtifactTypeMaven,
		Path:    path,
		Metadata: map[string]string{
			"groupId":   groupId,
			"filename":  filename,
			"extension": ext,
		},
	}, nil
}

// GeneratePath creates a storage path for the artifact
func (m *MavenArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	groupId := info.Metadata["groupId"]
	if groupId == "" {
		groupId = "default"
	}
	group := strings.ReplaceAll(groupId, ".", "/")
	return fmt.Sprintf("%s/%s/%s/%s-%s.jar", group, info.Name, info.Version, info.Name, info.Version)
}

// GetMetadata extracts metadata from Maven artifact content
func (m *MavenArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "maven-artifact",
		"format": "jar",
	}, nil
}

// GenerateIndex generates Maven metadata XML
func (m *MavenArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type MavenMetadata struct {
		XMLName    xml.Name `xml:"metadata"`
		GroupId    string   `xml:"groupId"`
		ArtifactId string   `xml:"artifactId"`
		Versioning struct {
			Latest   string   `xml:"latest"`
			Release  string   `xml:"release"`
			Versions []string `xml:"versions>version"`
		} `xml:"versioning"`
	}

	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	first := artifacts[0]
	groupId := first.Metadata["groupId"]
	if groupId == "" {
		groupId = "default"
	}
	metadata := MavenMetadata{
		GroupId:    groupId,
		ArtifactId: first.Name,
	}

	versions := make([]string, 0, len(artifacts))
	latest := ""
	for _, art := range artifacts {
		versions = append(versions, art.Version)
		latest = art.Version // Assume last is latest
	}

	metadata.Versioning.Versions = versions
	metadata.Versioning.Latest = latest
	metadata.Versioning.Release = latest

	return xml.MarshalIndent(metadata, "", "  ")
}

// GetEndpoints returns Maven standard endpoints
func (m *MavenArtifact) GetEndpoints() []string {
	return []string{
		"GET /{groupId}/{artifactId}/{version}/{filename}",
		"PUT /{groupId}/{artifactId}/{version}/{filename}",
		"GET /{groupId}/{artifactId}/maven-metadata.xml",
	}
}
