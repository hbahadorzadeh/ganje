package types

import (
	"encoding/json"
	"fmt"
	"io"
	"regexp"
	"strings"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// DockerArtifact implements Docker artifact handling
type DockerArtifact struct {
	metadata *artifact.Metadata
}

// NewDockerArtifact creates a new Docker artifact
func NewDockerArtifact(metadata *artifact.Metadata) artifact.Artifact {
	return &DockerArtifact{metadata: metadata}
}

// GetType returns the artifact type
func (d *DockerArtifact) GetType() artifact.ArtifactType {
	return artifact.ArtifactTypeDocker
}

// GetArtifactMetadata returns artifact metadata
func (d *DockerArtifact) GetArtifactMetadata() *artifact.Metadata {
	return d.metadata
}

// GetPath returns the storage path for Docker images
func (d *DockerArtifact) GetPath() string {
	return fmt.Sprintf("v2/%s/manifests/%s", d.metadata.Name, d.metadata.Version)
}

// GetIndexPath returns the index path for Docker registry
func (d *DockerArtifact) GetIndexPath() string {
	return fmt.Sprintf("v2/%s/tags/list", d.metadata.Name)
}

// ValidatePath validates Docker image path
func (d *DockerArtifact) ValidatePath(path string) error {
	// OCI Distribution Spec v1.1.1 paths (subset)
	patterns := []string{
		`^v2/$`,
		// name can be nested components separated by '/', reference can be tag or digest
		`^v2/[a-z0-9]+(?:(?:[._]|__|[-]|[/])[a-z0-9]+)*/manifests/[^/]+$`,
		// blobs by digest
		`^v2/[a-z0-9]+(?:(?:[._]|__|[-]|[/])[a-z0-9]+)*/blobs/sha256:[a-f0-9]{64}$`,
		// tags list
		`^v2/[a-z0-9]+(?:(?:[._]|__|[-]|[/])[a-z0-9]+)*/tags/list$`,
		// uploads (init and session)
		`^v2/[a-z0-9]+(?:(?:[._]|__|[-]|[/])[a-z0-9]+)*/blobs/uploads/$`,
		`^v2/[a-z0-9]+(?:(?:[._]|__|[-]|[/])[a-z0-9]+)*/blobs/uploads/[a-f0-9-]+$`,
		// referrers API
		`^v2/[a-z0-9]+(?:(?:[._]|__|[-]|[/])[a-z0-9]+)*/referrers/sha256:[a-f0-9]{64}$`,
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
	return fmt.Errorf("invalid Docker registry path: %s", path)
}

// ParsePath parses Docker image information from path
func (d *DockerArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	if err := d.ValidatePath(path); err != nil {
		return nil, err
	}

	if path == "v2/" {
		return &artifact.ArtifactInfo{
			Name:    "",
			Version: "",
			Type:    artifact.ArtifactTypeDocker,
			Path:    path,
			Metadata: map[string]string{
				"type": "api-root",
			},
		}, nil
	}

	// Common split
	parts := strings.Split(path, "/")
	if len(parts) < 3 {
		return nil, fmt.Errorf("invalid Docker path")
	}

	// name spans from index 1 to the segment before the keyword
	// find keyword index
	var kind string
	var name string
	var tail string
	for i := 2; i < len(parts); i++ {
		if parts[i] == "manifests" || parts[i] == "blobs" || parts[i] == "tags" || parts[i] == "referrers" {
			name = strings.Join(parts[1:i], "/")
			kind = parts[i]
			if i+1 < len(parts) {
				tail = strings.Join(parts[i+1:], "/")
			}
			break
		}
	}

	if kind == "" {
		return nil, fmt.Errorf("unsupported Docker path type")
	}

	info := &artifact.ArtifactInfo{
		Name:    name,
		Version: "",
		Type:    artifact.ArtifactTypeDocker,
		Path:    path,
		Metadata: map[string]string{},
	}

	switch kind {
	case "manifests":
		info.Version = tail // reference (tag or digest)
		info.Metadata["type"] = "manifest"
	case "blobs":
		// either direct blob or uploads
		if strings.HasPrefix(tail, "uploads/") || tail == "uploads" || tail == "uploads/" {
			info.Metadata["type"] = "upload"
			if strings.HasPrefix(tail, "uploads/") {
				sess := strings.TrimPrefix(tail, "uploads/")
				info.Metadata["upload_uuid"] = sess
			}
		} else if strings.HasPrefix(tail, "sha256:") {
			info.Version = tail
			info.Metadata["type"] = "blob"
		} else {
			return nil, fmt.Errorf("invalid blobs path")
		}
	case "tags":
		if tail == "list" {
			info.Metadata["type"] = "tags"
			info.Version = "list"
		} else {
			return nil, fmt.Errorf("invalid tags path")
		}
	case "referrers":
		info.Metadata["type"] = "referrers"
		info.Version = tail // digest
	default:
		return nil, fmt.Errorf("unsupported Docker path type")
	}

	return info, nil
}

// GeneratePath creates a storage path for the artifact
func (d *DockerArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	return fmt.Sprintf("v2/%s/manifests/%s", info.Name, info.Version)
}

// ValidateArtifact validates the artifact content
func (d *DockerArtifact) ValidateArtifact(content io.Reader) error {
	// Minimal validation: manifests are JSON documents in practice
	header := make([]byte, 1)
	if _, err := io.ReadFull(content, header); err != nil {
		if err == io.EOF || err == io.ErrUnexpectedEOF {
			return fmt.Errorf("invalid artifact content: too short")
		}
		return fmt.Errorf("invalid artifact content: %v", err)
	}
	// Allow '{' (likely JSON) or any other for blobs which are arbitrary bytes.
	// Since we don't know the context here, be permissive unless it's obviously invalid.
	return nil
}

// GetMetadata extracts metadata from artifact content
func (d *DockerArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	return map[string]string{
		"type":   "docker-image",
		"format": "oci",
	}, nil
}

// GenerateIndex generates Docker tags list
func (d *DockerArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	type DockerTagsList struct {
		Name string   `json:"name"`
		Tags []string `json:"tags"`
	}

	if len(artifacts) == 0 {
		return []byte{}, nil
	}

	first := artifacts[0]
	tagsList := DockerTagsList{
		Name: first.Name,
		Tags: make([]string, 0, len(artifacts)),
	}

	for _, art := range artifacts {
		tagsList.Tags = append(tagsList.Tags, art.Version)
	}

	return json.MarshalIndent(tagsList, "", "  ")
}

// GetEndpoints returns Docker registry standard endpoints
func (d *DockerArtifact) GetEndpoints() []string {
	return []string{
		"GET /v2/",
		"GET /v2/{name}/tags/list",
		"GET /v2/{name}/manifests/{reference}",
		"HEAD /v2/{name}/manifests/{reference}",
		"PUT /v2/{name}/manifests/{reference}",
		"GET /v2/{name}/blobs/{digest}",
		"HEAD /v2/{name}/blobs/{digest}",
		"POST /v2/{name}/blobs/uploads/",
		"PATCH /v2/{name}/blobs/uploads/{session_id}",
		"PUT /v2/{name}/blobs/uploads/{session_id}",
		"GET /v2/{name}/referrers/{digest}",
	}
}
