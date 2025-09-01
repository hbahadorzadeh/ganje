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
	return fmt.Sprintf("packages/%s/%s-%s.tar.gz",
		strings.ToLower(p.metadata.Name),
		p.metadata.Name,
		p.metadata.Version)
}

// GetIndexPath returns the index path for PyPI packages
func (p *PyPIArtifact) GetIndexPath() string {
	return fmt.Sprintf("simple/%s/", normalizeProjectName(p.metadata.Name))
}

// ValidatePath validates PyPI package path
func (p *PyPIArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^simple/$`,
		`^simple/[A-Za-z0-9][A-Za-z0-9._-]*/$`,
		// Accept common Simple API file layouts
		`^packages/[A-Za-z0-9._-]+/[A-Za-zA-Z0-9._-]+-(?:[0-9][^-]*)[^/]*\.(?:tar\.gz|whl)$`,
		// Also accept deeper paths under packages (e.g., hashed dirs)
		`^packages/.+/[A-Za-zA-Z0-9._-]+-(?:[0-9][^-]*)[^/]*\.(?:tar\.gz|whl)$`,
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
		if len(parts) < 3 {
			return nil, fmt.Errorf("invalid PyPI package path structure")
		}

		// Filename is always the last segment
		filename := parts[len(parts)-1]

		// Extract distribution name and version from filename per wheel/sdist conventions
		var dist, version, ext string
		if strings.HasSuffix(strings.ToLower(filename), ".tar.gz") {
			ext = ".tar.gz"
			base := strings.TrimSuffix(filename, ".tar.gz")
			// sdist: {name}-{version}
			re := regexp.MustCompile(`^(.+)-([0-9][^-]*)$`)
			m := re.FindStringSubmatch(base)
			if len(m) == 3 {
				dist, version = m[1], m[2]
			}
		} else if strings.HasSuffix(strings.ToLower(filename), ".whl") {
			ext = ".whl"
			base := strings.TrimSuffix(filename, ".whl")
			// wheel: {name}-{version}-(rest)
			re := regexp.MustCompile(`^(.+)-([0-9][^-]*)-.*$`)
			m := re.FindStringSubmatch(base)
			if len(m) == 3 {
				dist, version = m[1], m[2]
			}
		}

		if dist == "" || version == "" {
			return nil, fmt.Errorf("unable to parse filename: %s", filename)
		}

		return &artifact.ArtifactInfo{
			Name:    dist,
			Version: version,
			Type:    artifact.ArtifactTypePyPI,
			Path:    path,
			Metadata: map[string]string{
				"filename":  filename,
				"extension": ext,
				"project":   normalizeProjectName(dist),
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported PyPI path type")
}

// GeneratePath creates a storage path for the artifact
func (p *PyPIArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("packages/%s/%s-%s.tar.gz",
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

// normalizeProjectName applies PEP 503 normalization rules:
// - convert to lowercase
// - replace runs of '-', '_' and '.' with a single '-' character
func normalizeProjectName(name string) string {
	lower := strings.ToLower(name)
	re := regexp.MustCompile(`[-_.]+`)
	return re.ReplaceAllString(lower, "-")
}
