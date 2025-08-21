package types

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/stretchr/testify/assert"
)

func TestMavenArtifact(t *testing.T) {
	maven := &MavenArtifact{}

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeMaven, maven.GetType())
	})

	t.Run("ParsePath", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			expected *artifact.ArtifactInfo
			hasError bool
		}{
			{
				name: "valid jar path",
				path: "com/example/myapp/1.0.0/myapp-1.0.0.jar",
				expected: &artifact.ArtifactInfo{
					Name:    "myapp",
					Version: "1.0.0",
					Type:    "jar",
					Metadata: map[string]string{
						"groupId":    "com.example",
						"artifactId": "myapp",
					},
				},
				hasError: false,
			},
			{
				name: "valid pom path",
				path: "org/springframework/spring-core/5.3.21/spring-core-5.3.21.pom",
				expected: &artifact.ArtifactInfo{
					Name:    "spring-core",
					Version: "5.3.21",
					Type:    "pom",
					Metadata: map[string]string{
						"groupId":    "org.springframework",
						"artifactId": "spring-core",
					},
				},
				hasError: false,
			},
			{
				name:     "invalid path",
				path:     "invalid/path",
				expected: nil,
				hasError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				info, err := maven.ParsePath(tt.path)
				if tt.hasError {
					assert.Error(t, err)
					assert.Nil(t, info)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected.Name, info.Name)
					assert.Equal(t, tt.expected.Version, info.Version)
					assert.Equal(t, tt.expected.Type, info.Type)
					assert.Equal(t, tt.expected.Metadata["groupId"], info.Metadata["groupId"])
					assert.Equal(t, tt.expected.Metadata["artifactId"], info.Metadata["artifactId"])
				}
			})
		}
	})

	t.Run("GeneratePath", func(t *testing.T) {
		info := &artifact.ArtifactInfo{
			Name:    "myapp",
			Version: "1.0.0",
			Type:    "jar",
			Metadata: map[string]string{
				"groupId":    "com.example",
				"artifactId": "myapp",
			},
		}

		path := maven.GeneratePath(info)
		expected := "com/example/myapp/1.0.0/myapp-1.0.0.jar"
		assert.Equal(t, expected, path)
	})

	t.Run("ValidateArtifact", func(t *testing.T) {
		validJar := bytes.NewReader([]byte("PK\x03\x04")) // JAR file magic bytes
		err := maven.ValidateArtifact(validJar)
		assert.NoError(t, err)

		invalidContent := bytes.NewReader([]byte("invalid content"))
		err = maven.ValidateArtifact(invalidContent)
		assert.Error(t, err)
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		endpoints := maven.GetEndpoints()
		assert.Contains(t, endpoints, "/:groupId/:artifactId/:version/:filename")
		assert.Contains(t, endpoints, "/:groupId/:artifactId/maven-metadata.xml")
	})
}

func TestNPMArtifact(t *testing.T) {
	npm := &NPMArtifact{}

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeNPM, npm.GetType())
	})

	t.Run("ParsePath", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			expected *artifact.ArtifactInfo
			hasError bool
		}{
			{
				name: "scoped package",
				path: "@angular/core/-/core-14.2.0.tgz",
				expected: &artifact.ArtifactInfo{
					Name:    "@angular/core",
					Version: "14.2.0",
					Type:    "tgz",
				},
				hasError: false,
			},
			{
				name: "regular package",
				path: "express/-/express-4.18.1.tgz",
				expected: &artifact.ArtifactInfo{
					Name:    "express",
					Version: "4.18.1",
					Type:    "tgz",
				},
				hasError: false,
			},
			{
				name:     "invalid path",
				path:     "invalid",
				expected: nil,
				hasError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				info, err := npm.ParsePath(tt.path)
				if tt.hasError {
					assert.Error(t, err)
					assert.Nil(t, info)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected.Name, info.Name)
					assert.Equal(t, tt.expected.Version, info.Version)
					assert.Equal(t, tt.expected.Type, info.Type)
				}
			})
		}
	})

	t.Run("ValidateArtifact", func(t *testing.T) {
		validTgz := bytes.NewReader([]byte("\x1f\x8b\x08")) // gzip magic bytes
		err := npm.ValidateArtifact(validTgz)
		assert.NoError(t, err)

		invalidContent := bytes.NewReader([]byte("invalid"))
		err = npm.ValidateArtifact(invalidContent)
		assert.Error(t, err)
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		endpoints := npm.GetEndpoints()
		assert.Contains(t, endpoints, "/:package")
		assert.Contains(t, endpoints, "/:package/-/:filename")
	})
}

func TestDockerArtifact(t *testing.T) {
	docker := &DockerArtifact{}

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeDocker, docker.GetType())
	})

	t.Run("ParsePath", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			expected *artifact.ArtifactInfo
			hasError bool
		}{
			{
				name: "manifest path",
				path: "v2/library/nginx/manifests/latest",
				expected: &artifact.ArtifactInfo{
					Name:    "library/nginx",
					Version: "latest",
					Type:    "manifest",
				},
				hasError: false,
			},
			{
				name: "blob path",
				path: "v2/library/alpine/blobs/sha256:abc123",
				expected: &artifact.ArtifactInfo{
					Name:    "library/alpine",
					Version: "sha256:abc123",
					Type:    "blob",
				},
				hasError: false,
			},
			{
				name:     "invalid path",
				path:     "invalid",
				expected: nil,
				hasError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				info, err := docker.ParsePath(tt.path)
				if tt.hasError {
					assert.Error(t, err)
					assert.Nil(t, info)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected.Name, info.Name)
					assert.Equal(t, tt.expected.Version, info.Version)
					assert.Equal(t, tt.expected.Type, info.Type)
				}
			})
		}
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		endpoints := docker.GetEndpoints()
		assert.Contains(t, endpoints, "/v2/")
		assert.Contains(t, endpoints, "/v2/:name/tags/list")
		assert.Contains(t, endpoints, "/v2/:name/manifests/:reference")
	})
}

func TestPyPIArtifact(t *testing.T) {
	pypi := &PyPIArtifact{}

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypePyPI, pypi.GetType())
	})

	t.Run("ParsePath", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			expected *artifact.ArtifactInfo
			hasError bool
		}{
			{
				name: "wheel package",
				path: "packages/django/Django-4.1.0-py3-none-any.whl",
				expected: &artifact.ArtifactInfo{
					Name:    "Django",
					Version: "4.1.0",
					Type:    "whl",
				},
				hasError: false,
			},
			{
				name: "source distribution",
				path: "packages/requests/requests-2.28.1.tar.gz",
				expected: &artifact.ArtifactInfo{
					Name:    "requests",
					Version: "2.28.1",
					Type:    "tar.gz",
				},
				hasError: false,
			},
			{
				name:     "invalid path",
				path:     "invalid",
				expected: nil,
				hasError: true,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				info, err := pypi.ParsePath(tt.path)
				if tt.hasError {
					assert.Error(t, err)
					assert.Nil(t, info)
				} else {
					assert.NoError(t, err)
					assert.Equal(t, tt.expected.Name, info.Name)
					assert.Equal(t, tt.expected.Version, info.Version)
					assert.Equal(t, tt.expected.Type, info.Type)
				}
			})
		}
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		endpoints := pypi.GetEndpoints()
		assert.Contains(t, endpoints, "/simple/")
		assert.Contains(t, endpoints, "/simple/:package/")
		assert.Contains(t, endpoints, "/packages/:path")
	})
}

func TestGenericArtifact(t *testing.T) {
	generic := &GenericArtifact{}

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeGeneric, generic.GetType())
	})

	t.Run("ParsePath", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			expected *artifact.ArtifactInfo
			hasError bool
		}{
			{
				name: "simple file",
				path: "files/document.pdf",
				expected: &artifact.ArtifactInfo{
					Name:    "document.pdf",
					Version: "latest",
					Type:    "pdf",
				},
				hasError: false,
			},
			{
				name: "nested path",
				path: "releases/v1.0.0/app.zip",
				expected: &artifact.ArtifactInfo{
					Name:    "app.zip",
					Version: "v1.0.0",
					Type:    "zip",
				},
				hasError: false,
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				info, err := generic.ParsePath(tt.path)
				assert.NoError(t, err) // Generic should never fail parsing
				assert.Equal(t, tt.expected.Name, info.Name)
				assert.Equal(t, tt.expected.Type, info.Type)
			})
		}
	})

	t.Run("ValidateArtifact", func(t *testing.T) {
		content := bytes.NewReader([]byte("any content"))
		err := generic.ValidateArtifact(content)
		assert.NoError(t, err) // Generic accepts any content
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		endpoints := generic.GetEndpoints()
		assert.Contains(t, endpoints, "/*path")
	})
}

func TestArtifactMetadataExtraction(t *testing.T) {
	tests := []struct {
		name         string
		artifactType artifact.Artifact
		content      string
		expectedMeta map[string]string
	}{
		{
			name:         "Maven POM metadata",
			artifactType: &MavenArtifact{},
			content: `<?xml version="1.0" encoding="UTF-8"?>
<project>
    <groupId>com.example</groupId>
    <artifactId>test-app</artifactId>
    <version>1.0.0</version>
    <name>Test Application</name>
</project>`,
			expectedMeta: map[string]string{
				"groupId":    "com.example",
				"artifactId": "test-app",
				"version":    "1.0.0",
				"name":       "Test Application",
			},
		},
		{
			name:         "NPM package.json metadata",
			artifactType: &NPMArtifact{},
			content: `{
    "name": "@angular/core",
    "version": "14.2.0",
    "description": "Angular - the core framework",
    "author": "Angular Team"
}`,
			expectedMeta: map[string]string{
				"name":        "@angular/core",
				"version":     "14.2.0",
				"description": "Angular - the core framework",
				"author":      "Angular Team",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			reader := strings.NewReader(tt.content)
			metadata, err := tt.artifactType.GetMetadata(reader)
			assert.NoError(t, err)

			for key, expectedValue := range tt.expectedMeta {
				assert.Equal(t, expectedValue, metadata[key], "Metadata key %s mismatch", key)
			}
		})
	}
}

func TestArtifactIndexGeneration(t *testing.T) {
	artifacts := []*artifact.ArtifactInfo{
		{
			Name:    "test-app",
			Version: "1.0.0",
			Type:    "jar",
			Metadata: map[string]string{
				"groupId":    "com.example",
				"artifactId": "test-app",
			},
		},
		{
			Name:    "test-app",
			Version: "1.1.0",
			Type:    "jar",
			Metadata: map[string]string{
				"groupId":    "com.example",
				"artifactId": "test-app",
			},
		},
	}

	tests := []struct {
		name         string
		artifactType artifact.Artifact
		artifacts    []*artifact.ArtifactInfo
	}{
		{
			name:         "Maven index generation",
			artifactType: &MavenArtifact{},
			artifacts:    artifacts,
		},
		{
			name:         "NPM index generation",
			artifactType: &NPMArtifact{},
			artifacts:    artifacts,
		},
		{
			name:         "Docker index generation",
			artifactType: &DockerArtifact{},
			artifacts:    artifacts,
		},
		{
			name:         "Generic index generation",
			artifactType: &GenericArtifact{},
			artifacts:    artifacts,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			index, err := tt.artifactType.GenerateIndex(tt.artifacts)
			assert.NoError(t, err)
			assert.NotEmpty(t, index)

			// Verify index contains artifact information
			indexStr := string(index)
			assert.Contains(t, indexStr, "test-app")
			assert.Contains(t, indexStr, "1.0.0")
			assert.Contains(t, indexStr, "1.1.0")
		})
	}
}
