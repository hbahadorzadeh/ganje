package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// CargoArtifact implements Rust Cargo artifact handling
type CargoArtifact struct {
	metadata *artifact.Metadata
}

// NewCargoArtifact creates a new Cargo artifact
func NewCargoArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &CargoArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (c *CargoArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeCargo
}

// GetArtifactMetadata returns artifact metadata
func (c *CargoArtifact) GetArtifactMetadata() *artifact.Metadata {
	return c.metadata
}

// GetPath returns the storage path for Cargo crates
func (c *CargoArtifact) GetPath() string {
	return fmt.Sprintf("api/v1/crates/%s/%s/download", c.metadata.Name, c.metadata.Version)
}

// GetIndexPath returns the index path for Cargo registry
func (c *CargoArtifact) GetIndexPath() string {
	name := strings.ToLower(c.metadata.Name)
	switch len(name) {
	case 1:
		return fmt.Sprintf("1/%s", name)
	case 2:
		return fmt.Sprintf("2/%s", name)
	case 3:
		return fmt.Sprintf("3/%s/%s", name[:1], name)
	default:
		return fmt.Sprintf("%s/%s/%s", name[:2], name[2:4], name)
	}
}

// ValidatePath validates Cargo crate path
func (c *CargoArtifact) ValidatePath(path string) error {
	patterns := []string{
		`^api/v1/crates/[a-zA-Z0-9._-]+/[a-zA-Z0-9._-]+/download$`,
		`^[0-9]/[a-zA-Z0-9._-]+$`,
		`^[a-zA-Z0-9._-]{2}/[a-zA-Z0-9._-]+$`,
		`^[a-zA-Z0-9._-]{2}/[a-zA-Z0-9._-]{2}/[a-zA-Z0-9._-]+$`,
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
	return fmt.Errorf("invalid Cargo crate path: %s", path)
}

// ParsePath parses Cargo crate information from path
func (c *CargoArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := c.ValidatePath(path); err != nil {
		return nil, err
	}

	if strings.HasPrefix(path, "api/v1/crates/") {
		parts := strings.Split(path, "/")
		if len(parts) < 6 {
			return nil, fmt.Errorf("invalid Cargo download path")
		}
		
		name := parts[3]
		version := parts[4]
		
		return &artifact.ArtifactInfo{
			Name:    name,
			Version: version,
			Type:    artifact.ArtifactTypeCargo,
			Path:    path,
			Metadata: map[string]string{
				"type": "crate",
			},
		}, nil
	}

	return nil, fmt.Errorf("unsupported Cargo path type")
}

// GeneratePath creates a storage path for the artifact
func (c *CargoArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("api/v1/crates/%s/%s/download", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (c *CargoArtifact) ValidateArtifact(content io.Reader) error {
	buf := make([]byte, 1024)
	_, err := content.Read(buf)
	if err != nil && err != io.EOF {
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	return nil
}

// GetMetadata extracts metadata from artifact content
func (c *CargoArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type": "cargo-crate",
		"format": "crate",
	}, nil
}

// GenerateIndex generates Cargo registry index
func (c *CargoArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
    // Per Cargo registry-index spec, each line is a JSON object describing a
    // single version of a crate. We include minimal required fields and set
    // schema version v=2.
    type CargoVersion struct {
        Name     string              `json:"name"`
        Vers     string              `json:"vers"`
        Deps     []struct{}          `json:"deps"`
        Cksum    string              `json:"cksum"`
        Features map[string][]string `json:"features"`
        Yanked   bool                `json:"yanked"`
        V        int                 `json:"v"`
    }

	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	var lines []string
	for _, art := range artifacts {
        version := CargoVersion{
            Name:     art.Name,
            Vers:     art.Version,
            Deps:     []struct{}{},
            Cksum:    art.Checksum,
            Features: make(map[string][]string),
            Yanked:   art.Yanked,
            V:        2,
        }
		
		jsonLine, err := json.Marshal(version)
		if err != nil {
			return nil, err
		}
		lines = append(lines, string(jsonLine))
	}

	return []byte(strings.Join(lines, "\n")), nil
}

// GetEndpoints returns Cargo registry standard endpoints
func (c *CargoArtifact) GetEndpoints() []string {
    return []string{
        "GET /api/v1/crates/{crate}",
        "GET /api/v1/crates/{crate}/{version}",
        "GET /api/v1/crates/{crate}/{version}/download",
        "PUT /api/v1/crates/new",
        "DELETE /api/v1/crates/{crate}/{version}/yank",
        "PUT /api/v1/crates/{crate}/{version}/unyank",
    }
}
