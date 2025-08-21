package types

import (
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// PyPIArtifact implements PyPI artifact handling
type PyPIArtifact struct {
	metadata *artifact.Metadata
}

// NewPyPIArtifact creates a new PyPI artifact
func NewPyPIArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &PyPIArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (p *PyPIArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypePyPI
}

// GetArtifactMetadata returns artifact metadata
func (p *PyPIArtifact) GetArtifactMetadata() *artifact.Metadata {
	return p.metadata
}

// GetPath returns the storage path for PyPI packages
func (p *PyPIArtifact) GetPath() string {
	return fmt.Sprintf("packages/%s/%s/%s-%s.tar.gz", 
		strings.ToLower(p.metadata.Name[:1]), 
		strings.ToLower(p.metadata.Name), 
		p.metadata.Name, 
		p.metadata.Version)
}

// GetIndexPath returns the index path for PyPI packages
func (p *PyPIArtifact) GetIndexPath() string {
	return fmt.Sprintf("simple/%s/", strings.ToLower(p.metadata.Name))
}

// ValidatePath validates PyPI package path
func (p *PyPIArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^packages/[a-z]/[a-z0-9._-]+/[a-zA-Z0-9._-]+-[a-zA-Z0-9._-]+\.(tar\.gz|whl)$`,
		`^simple/[a-z0-9._-]+/$`,
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
	return fmt.Errorf("invalid PyPI package path: %s", path)
}

// ParsePath parses PyPI package information from path
func (p *PyPIArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := p.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.HasPrefix(path, "packages/") {
		parts := strings.Split(path, "/")
		if len(parts) < 4 {
			return nil, fmt.Errorf("invalid PyPI package path structure")
		}
		
		packageName := parts[2]
		filename := parts[3]
		
		// Extract version from filename
		var version string
		if strings.HasSuffix(filename, ".tar.gz") {
			nameVersion := strings.TrimSuffix(filename, ".tar.gz")
			version = strings.TrimPrefix(nameVersion, packageName+"-")
		} else if strings.HasSuffix(filename, ".whl") {
			nameVersion := strings.TrimSuffix(filename, ".whl")
			parts := strings.Split(nameVersion, "-")
			if len(parts) >= 2 {
				version = parts[1]
			}
		}
		
		return &artifact.ArtifactInfo{
			Name:    packageName,
			Version: version,
			Type:    artifact.ArtifactTypePyPI,
			Path:    path,
			Metadata: map[string]string{
				"filename": filename,
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported PyPI path type")
}

// GeneratePath creates a storage path for the artifact
func (p *PyPIArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("packages/%s/%s/%s-%s.tar.gz", 
		strings.ToLower(info.Name[:1]), 
		strings.ToLower(info.Name), 
		info.Name, 
		info.Version)
}

// ValidateArtifact validates the artifact content
func (p *PyPIArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (p *PyPIArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "python-package",
		"format": "tar.gz",
	}, nil
}

// GenerateIndex generates PyPI simple index
func (p *PyPIArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	first := artifacts[0]
	html := fmt.Sprintf(`<!DOCTYPE html>
<html>
<head>
    <title>Links for %s</title>
</head>
<body>
    <h1>Links for %s</h1>
`, first.Name, first.Name)

	for _, art := range artifacts {
		filename := art.Metadata["filename"]
		html += fmt.Sprintf(`    <a href="../../%s">%s</a><br/>
`, art.Path, filename)
	}

	html += `</body>
</html>`

	return []byte(html), nil
}

// GetEndpoints returns PyPI standard endpoints
func (p *PyPIArtifact) GetEndpoints() []string {
	return []string{
		"GET /simple/",
		"GET /simple/{package}/",
		"GET /packages/{hash}/{filename}",
		"POST /legacy/",
	}
}
