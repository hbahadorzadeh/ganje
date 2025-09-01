package types

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/stretchr/testify/assert"
	yaml "gopkg.in/yaml.v3"
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
					Metadata: map[string]string{
						"groupId":   "com.example",
						"extension": ".jar",
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
					Metadata: map[string]string{
						"groupId":   "org.springframework",
						"extension": ".pom",
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
					assert.Equal(t, tt.expected.Metadata["extension"], info.Metadata["extension"])
					assert.Equal(t, tt.expected.Metadata["groupId"], info.Metadata["groupId"])
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
		assert.Contains(t, endpoints, "GET /{groupId}/{artifactId}/{version}/{filename}")
		assert.Contains(t, endpoints, "PUT /{groupId}/{artifactId}/{version}/{filename}")
		assert.Contains(t, endpoints, "GET /{groupId}/{artifactId}/maven-metadata.xml")
	})
}

func TestHelmArtifact(t *testing.T) {
    h := &HelmArtifact{metadata: &artifact.Metadata{Name: "mychart", Version: "1.0.0"}}

    t.Run("GetType", func(t *testing.T) {
        assert.Equal(t, artifact.ArtifactTypeHelm, h.GetType())
    })

    t.Run("ParsePath", func(t *testing.T) {
        info, err := h.ParsePath("mychart-1.2.3.tgz")
        assert.NoError(t, err)
        assert.Equal(t, "mychart", info.Name)
        assert.Equal(t, "1.2.3", info.Version)
        assert.Equal(t, artifact.ArtifactTypeHelm, info.Type)
        assert.Equal(t, "mychart-1.2.3.tgz", info.Metadata["filename"])
    })

    t.Run("GenerateIndex YAML", func(t *testing.T) {
        arts := []*artifact.ArtifactInfo{
            {Name: "mychart", Version: "1.0.0", Path: "mychart-1.0.0.tgz"},
            {Name: "mychart", Version: "1.2.0", Path: "mychart-1.2.0.tgz"},
        }
        b, err := h.GenerateIndex(arts)
        assert.NoError(t, err)
        // Unmarshal YAML and validate structure
        var idx struct {
            APIVersion string `yaml:"apiVersion"`
            Entries    map[string][]struct {
                APIVersion string   `yaml:"apiVersion"`
                Version    string   `yaml:"version"`
                URLs       []string `yaml:"urls"`
            } `yaml:"entries"`
        }
        err = yaml.Unmarshal(b, &idx)
        assert.NoError(t, err)
        assert.Equal(t, "v1", idx.APIVersion)
        list, ok := idx.Entries["mychart"]
        assert.True(t, ok)
        // Expect two versions and that 1.2.0 is present with correct URL
        versions := []string{list[0].Version, list[1].Version}
        assert.Contains(t, versions, "1.2.0")
        // Ensure URL matches
        foundURL := false
        for _, e := range list {
            if e.Version == "1.2.0" {
                if len(e.URLs) > 0 && e.URLs[0] == "mychart-1.2.0.tgz" {
                    foundURL = true
                }
            }
        }
        assert.True(t, foundURL)
    })

    t.Run("Endpoints", func(t *testing.T) {
        eps := h.GetEndpoints()
        assert.Contains(t, eps, "GET /index.yaml")
        assert.Contains(t, eps, "GET /{chart}-{version}.tgz")
    })
}

func TestCargoArtifact(t *testing.T) {
	cargo := &CargoArtifact{}

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeCargo, cargo.GetType())
	})

	t.Run("GetIndexPath rules", func(t *testing.T) {
		cases := map[string]string{
			"a":            "1/a",
			"ab":           "2/ab",
			"abc":          "3/a/abc",
			"serde_json":   "se/rd/serde_json",
		}
		for name, want := range cases {
			ca := &CargoArtifact{metadata: &artifact.Metadata{Name: name}}
			assert.Equal(t, want, ca.GetIndexPath(), name)
		}
	})

	t.Run("GenerateIndex v2 schema", func(t *testing.T) {
		artifacts := []*artifact.ArtifactInfo{
			{Name: "serde", Version: "1.0.0", Checksum: "deadbeef"},
			{Name: "serde", Version: "1.0.1", Checksum: "cafebabe"},
		}
		idx, err := cargo.GenerateIndex(artifacts)
		assert.NoError(t, err)
		s := string(idx)
		assert.Contains(t, s, `"name":"serde"`)
		assert.Contains(t, s, `"vers":"1.0.0"`)
		assert.Contains(t, s, `"vers":"1.0.1"`)
		assert.Contains(t, s, `"v":2`)
	})

	t.Run("Endpoints", func(t *testing.T) {
		eps := cargo.GetEndpoints()
		assert.Contains(t, eps, "GET /api/v1/crates/{crate}")
		assert.Contains(t, eps, "GET /api/v1/crates/{crate}/{version}")
		assert.Contains(t, eps, "GET /api/v1/crates/{crate}/{version}/download")
		assert.Contains(t, eps, "PUT /api/v1/crates/new")
		assert.Contains(t, eps, "DELETE /api/v1/crates/{crate}/{version}/yank")
		assert.Contains(t, eps, "PUT /api/v1/crates/{crate}/{version}/unyank")
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
					Type:    ".tgz",
				},
				hasError: false,
			},
			{
				name: "regular package",
				path: "express/-/express-4.18.1.tgz",
				expected: &artifact.ArtifactInfo{
					Name:    "express",
					Version: "4.18.1",
					Type:    ".tgz",
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
		assert.Contains(t, endpoints, "GET /{package}")
		assert.Contains(t, endpoints, "GET /{package}/{version}")
		assert.Contains(t, endpoints, "GET /{package}/-/{filename}")
		assert.Contains(t, endpoints, "PUT /{package}")
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
					Type:    artifact.ArtifactTypeDocker,
				},
				hasError: false,
			},
			{
				name: "blob path",
				path: "v2/library/alpine/blobs/sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
				expected: &artifact.ArtifactInfo{
					Name:    "library/alpine",
					Version: "sha256:aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
					Type:    artifact.ArtifactTypeDocker,
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
					if strings.Contains(tt.path, "/manifests/") {
						assert.Equal(t, "manifest", info.Metadata["type"])
					}
					if strings.Contains(tt.path, "/blobs/") {
						assert.Equal(t, "blob", info.Metadata["type"])
					}
				}
			})
		}
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		endpoints := docker.GetEndpoints()
		assert.Contains(t, endpoints, "GET /v2/")
		assert.Contains(t, endpoints, "GET /v2/{name}/tags/list")
		assert.Contains(t, endpoints, "GET /v2/{name}/manifests/{reference}")
		assert.Contains(t, endpoints, "HEAD /v2/{name}/manifests/{reference}")
		assert.Contains(t, endpoints, "GET /v2/{name}/blobs/{digest}")
		assert.Contains(t, endpoints, "HEAD /v2/{name}/blobs/{digest}")
		assert.Contains(t, endpoints, "POST /v2/{name}/blobs/uploads/")
		assert.Contains(t, endpoints, "PATCH /v2/{name}/blobs/uploads/{session_id}")
		assert.Contains(t, endpoints, "PUT /v2/{name}/blobs/uploads/{session_id}")
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
					Type:    artifact.ArtifactTypePyPI,
				},
				hasError: false,
			},
			{
				name: "source distribution",
				path: "packages/requests/requests-2.28.1.tar.gz",
				expected: &artifact.ArtifactInfo{
					Name:    "requests",
					Version: "2.28.1",
					Type:    artifact.ArtifactTypePyPI,
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
					// Validate extension captured in metadata
					if strings.HasSuffix(tt.path, ".whl") {
						assert.Equal(t, ".whl", info.Metadata["extension"])
					}
					if strings.HasSuffix(tt.path, ".tar.gz") {
						assert.Equal(t, ".tar.gz", info.Metadata["extension"])
					}
				}
			})
		}
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		endpoints := pypi.GetEndpoints()
		assert.Contains(t, endpoints, "GET /simple/")
		assert.Contains(t, endpoints, "GET /simple/{package}/")
		assert.Contains(t, endpoints, "GET /packages/{hash}/{filename}")
	})
}

func TestGoModuleArtifact(t *testing.T) {
	gom := &GoModuleArtifact{}

	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeGolang, gom.GetType())
	})

	t.Run("ParsePath", func(t *testing.T) {
		tests := []struct {
			name     string
			path     string
			wantName string
			wantVer  string
			kind     string
		}{
			{
				name:     "list",
				path:     "github.com/user/repo/@v/list",
				wantName: "github.com/user/repo",
				kind:     "list",
			},
			{
				name:     "latest",
				path:     "github.com/user/repo/@latest",
				wantName: "github.com/user/repo",
				kind:     "latest",
			},
			{
				name:     "zip semver",
				path:     "github.com/user/repo/@v/v1.2.3.zip",
				wantName: "github.com/user/repo",
				wantVer:  "v1.2.3",
				kind:     "zip",
			},
			{
				name:     "mod semver",
				path:     "github.com/user/repo/@v/v1.2.3.mod",
				wantName: "github.com/user/repo",
				wantVer:  "v1.2.3",
				kind:     "mod",
			},
			{
				name:     "info semver",
				path:     "github.com/user/repo/@v/v1.2.3.info",
				wantName: "github.com/user/repo",
				wantVer:  "v1.2.3",
				kind:     "info",
			},
			{
				name:     "zip pseudo-version",
				path:     "example.org/my/mod/@v/v0.0.0-20180504190223-abcdefabcdef.zip",
				wantName: "example.org/my/mod",
				wantVer:  "v0.0.0-20180504190223-abcdefabcdef",
				kind:     "zip",
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				info, err := gom.ParsePath(tt.path)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantName, info.Name)
				if tt.wantVer != "" {
					assert.Equal(t, tt.wantVer, info.Version)
				}
				assert.Equal(t, artifact.ArtifactTypeGolang, info.Type)
				if tt.kind != "" {
					assert.Equal(t, tt.kind, info.Metadata["kind"])
				}
			})
		}
	})

	t.Run("GeneratePath", func(t *testing.T) {
		info := &artifact.ArtifactInfo{Name: "github.com/foo/bar", Version: "v1.0.0"}
		assert.Equal(t, "github.com/foo/bar/@v/v1.0.0.zip", gom.GeneratePath(info))
	})

	t.Run("ValidateArtifact", func(t *testing.T) {
		content := bytes.NewReader([]byte("dummy"))
		err := gom.ValidateArtifact(content)
		assert.NoError(t, err)
	})

	t.Run("GetEndpoints", func(t *testing.T) {
		eps := gom.GetEndpoints()
		assert.Contains(t, eps, "GET /{module}/@v/list")
		assert.Contains(t, eps, "GET /{module}/@v/{version}.info")
		assert.Contains(t, eps, "GET /{module}/@v/{version}.mod")
		assert.Contains(t, eps, "GET /{module}/@v/{version}.zip")
		assert.Contains(t, eps, "GET /{module}/@latest")
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

func TestAnsibleArtifact(t *testing.T) {
	a := &AnsibleArtifact{metadata: &artifact.Metadata{Group: "namespace", Name: "collection_name", Version: "1.0.0"}}

	t.Run("GetIndexPath v3", func(t *testing.T) {
		got := a.GetIndexPath()
		want := "api/v3/plugin/ansible/content/published/collections/index/namespace/collection_name/"
		assert.Equal(t, want, got)
	})

	t.Run("ValidatePath", func(t *testing.T) {
		ok1 := a.ValidatePath("download/namespace-collection_name-1.0.0.tar.gz")
		assert.NoError(t, ok1)
		ok2 := a.ValidatePath("api/v3/plugin/ansible/content/published/collections/index/namespace/collection_name/")
		assert.NoError(t, ok2)
		err := a.ValidatePath("api/v2/collections/namespace/collection_name/")
		assert.Error(t, err)
	})

	t.Run("ParsePath download", func(t *testing.T) {
		info, err := a.ParsePath("download/namespace-collection_name-1.2.3.tar.gz")
		assert.NoError(t, err)
		assert.Equal(t, "collection_name", info.Name)
		assert.Equal(t, "1.2.3", info.Version)
		assert.Equal(t, artifact.ArtifactTypeAnsible, info.Type)
		assert.Equal(t, "namespace", info.Metadata["namespace"])
		assert.Equal(t, "namespace-collection_name-1.2.3.tar.gz", info.Metadata["filename"])
	})

	t.Run("GenerateIndex schema", func(t *testing.T) {
		arts := []*artifact.ArtifactInfo{
			{Name: "collection_name", Version: "1.0.0", Metadata: map[string]string{"namespace": "namespace"}},
			{Name: "collection_name", Version: "1.2.0", Metadata: map[string]string{"namespace": "namespace"}},
		}
		b, err := a.GenerateIndex(arts)
		assert.NoError(t, err)
		s := string(b)
		assert.Contains(t, s, `"href":"/api/v3/plugin/ansible/content/published/collections/index/namespace/collection_name/"`)
		assert.Contains(t, s, `"namespace":"namespace"`)
		assert.Contains(t, s, `"name":"collection_name"`)
		assert.Contains(t, s, `"versions_url":"/api/v3/plugin/ansible/content/published/collections/index/namespace/collection_name/versions/"`)
		assert.Contains(t, s, `"highest_version"`)
		assert.Contains(t, s, `"version":"1.2.0"`)
	})

	t.Run("Endpoints", func(t *testing.T) {
		eps := a.GetEndpoints()
		assert.Contains(t, eps, "GET /api/v3/plugin/ansible/content/published/collections/index/{namespace}/{name}/")
		assert.Contains(t, eps, "GET /download/{namespace}-{name}-{version}.tar.gz")
	})
}
