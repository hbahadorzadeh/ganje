package types

import (
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// NuGetArtifact implements NuGet artifact handling
type NuGetArtifact struct {
	metadata *artifact.Metadata
}

// NewNuGetArtifact creates a new NuGet artifact
func NewNuGetArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &NuGetArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (n *NuGetArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeNuGet
}

// GetArtifactMetadata returns artifact metadata
func (n *NuGetArtifact) GetArtifactMetadata() *artifact.Metadata {
	return n.metadata
}

// GetPath returns the storage path for NuGet packages
func (n *NuGetArtifact) GetPath() string {
	return fmt.Sprintf("%s/%s/%s.%s.nupkg", 
		strings.ToLower(n.metadata.Name), 
		n.metadata.Version, 
		strings.ToLower(n.metadata.Name), 
		n.metadata.Version)
}

// GetIndexPath returns the index path for NuGet feed
func (n *NuGetArtifact) GetIndexPath() string {
	return "Packages"
}

// ValidatePath validates NuGet package path
func (n *NuGetArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+\.[a-zA-Z0-9._-]+\.nupkg$`,
		`^Packages$`,
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
	return fmt.Errorf("invalid NuGet package path: %s", path)
}

// ParsePath parses NuGet package information from path
func (n *NuGetArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := n.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.HasSuffix(path, ".nupkg") {
		parts := strings.Split(path, "/")
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid NuGet package path structure")
		}
		
		packageId := parts[0]
		version := parts[1]
		
		return &artifact.ArtifactInfo{
			Name:    packageId,
			Version: version,
			Type:    artifact.ArtifactTypeNuGet,
			Path:    path,
			Metadata: map[string]string{
				"filename": parts[2],
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported NuGet path type")
}

// GeneratePath creates a storage path for the artifact
func (n *NuGetArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("%s/%s/%s.%s.nupkg", 
		strings.ToLower(info.Name), 
		info.Version, 
		strings.ToLower(info.Name), 
		info.Version)
}

// ValidateArtifact validates the artifact content
func (n *NuGetArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (n *NuGetArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "nuget-package",
		"format": "nupkg",
	}, nil
}

// GenerateIndex generates NuGet feed XML
func (n *NuGetArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type NuGetEntry struct {
		XMLName xml.Name `xml:"entry"`
		ID      string   `xml:"id"`
		Title   string   `xml:"title"`
		Updated string   `xml:"updated"`
		Content struct {
			Type       string `xml:"type,attr"`
			Properties struct {
				ID      string `xml:"Id"`
				Version string `xml:"Version"`
			} `xml:"properties"`
		} `xml:"content"`
	}

	type NuGetFeed struct {
		XMLName xml.Name     `xml:"feed"`
		Xmlns   string       `xml:"xmlns,attr"`
		Title   string       `xml:"title"`
		ID      string       `xml:"id"`
		Updated string       `xml:"updated"`
		Entries []NuGetEntry `xml:"entry"`
	}

	feed := NuGetFeed{
		Xmlns:   "http://www.w3.org/2005/Atom",
		Title:   "NuGet Feed",
		ID:      "http://localhost/nuget",
		Updated: "2023-01-01T00:00:00Z",
		Entries: make([]NuGetEntry, 0, len(artifacts)),
	}

	for _, art := range artifacts {
		entry := NuGetEntry{
			ID:      fmt.Sprintf("http://localhost/nuget/Packages(Id='%s',Version='%s')", art.Name, art.Version),
			Title:   art.Name,
			Updated: art.UploadTime.Format("2006-01-02T15:04:05Z"),
		}
		entry.Content.Type = "application/xml"
		entry.Content.Properties.ID = art.Name
		entry.Content.Properties.Version = art.Version
		
		feed.Entries = append(feed.Entries, entry)
	}

	return xml.MarshalIndent(feed, "", "  ")
}

// GetEndpoints returns NuGet standard endpoints
func (n *NuGetArtifact) GetEndpoints() []string {
	return []string{
		"GET /Packages",
		"GET /Packages(Id='{id}',Version='{version}')",
		"PUT /api/v2/package",
		"DELETE /api/v2/package/{id}/{version}",
	}
}
