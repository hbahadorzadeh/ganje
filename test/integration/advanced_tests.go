package integration

import (
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestArtifactVersioning tests artifact versioning scenarios
func TestArtifactVersioning(t *testing.T) {
	helper := NewTestHelper("http://localhost:8080", t)
	defer helper.Close()

	repoName := "versioning-test-repo"
	helper.CreateTestRepository(repoName, artifact.ArtifactTypeMaven)

	testCases := []struct {
		version string
		content []byte
	}{
		{"1.0.0", createMockJar()},
		{"1.0.1", createMockJar()},
		{"1.1.0", createMockJar()},
		{"2.0.0", createMockJar()},
	}

	// Push multiple versions
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Push_v%s", tc.version), func(t *testing.T) {
			endpoint := fmt.Sprintf("/maven-repo/com/example/test-app/%s/test-app-%s.jar", tc.version, tc.version)
			resp := helper.PushArtifact(endpoint, tc.content)
			defer resp.Body.Close()

			assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
		})
	}

	// Verify all versions exist
	t.Run("List_All_Versions", func(t *testing.T) {
		artifacts, resp := helper.ListArtifacts(repoName)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Len(t, artifacts, len(testCases))
	})

	// Pull each version
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("Pull_v%s", tc.version), func(t *testing.T) {
			endpoint := fmt.Sprintf("/maven-repo/com/example/test-app/%s/test-app-%s.jar", tc.version, tc.version)
			content, resp := helper.PullArtifact(endpoint)
			defer resp.Body.Close()

			assert.Equal(t, http.StatusOK, resp.StatusCode)
			assert.Equal(t, tc.content, content)
		})
	}
}

// TestArtifactOverwrite tests artifact overwrite scenarios
func TestArtifactOverwrite(t *testing.T) {
	helper := NewTestHelper("http://localhost:8080", t)
	defer helper.Close()

	repoName := "overwrite-test-repo"
	helper.CreateTestRepository(repoName, artifact.ArtifactTypeNPM)

	endpoint := "/npm-repo/test-package/-/test-package-1.0.0.tgz"
	originalContent := createMockNpmPackage()
	modifiedContent := append(originalContent, []byte("modified")...)

	// Push original
	t.Run("Push_Original", func(t *testing.T) {
		resp := helper.PushArtifact(endpoint, originalContent)
		defer resp.Body.Close()
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
	})

	// Verify original
	t.Run("Pull_Original", func(t *testing.T) {
		content, resp := helper.PullArtifact(endpoint)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, originalContent, content)
	})

	// Push modified (overwrite)
	t.Run("Push_Modified", func(t *testing.T) {
		resp := helper.PushArtifact(endpoint, modifiedContent)
		defer resp.Body.Close()
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
	})

	// Verify modified
	t.Run("Pull_Modified", func(t *testing.T) {
		content, resp := helper.PullArtifact(endpoint)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, modifiedContent, content)
	})
}

// TestArtifactDeletion tests artifact deletion
func TestArtifactDeletion(t *testing.T) {
	helper := NewTestHelper("http://localhost:8080", t)
	defer helper.Close()

	repoName := "deletion-test-repo"
	helper.CreateTestRepository(repoName, artifact.ArtifactTypePyPI)

	endpoint := "/pypi-repo/simple/test-package/test_package-1.0.0-py3-none-any.whl"
	content := createMockPythonWheel()

	// Push artifact
	t.Run("Push", func(t *testing.T) {
		resp := helper.PushArtifact(endpoint, content)
		defer resp.Body.Close()
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
	})

	// Verify exists
	t.Run("Verify_Exists", func(t *testing.T) {
		_, resp := helper.PullArtifact(endpoint)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusOK, resp.StatusCode)
	})

	// Delete artifact
	t.Run("Delete", func(t *testing.T) {
		resp := helper.DeleteArtifact(endpoint)
		defer resp.Body.Close()
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
	})

	// Verify deleted
	t.Run("Verify_Deleted", func(t *testing.T) {
		_, resp := helper.PullArtifact(endpoint)
		defer resp.Body.Close()
		assert.Equal(t, http.StatusNotFound, resp.StatusCode)
	})
}

// TestArtifactSearch tests artifact search functionality
func TestArtifactSearch(t *testing.T) {
	helper := NewTestHelper("http://localhost:8080", t)
	defer helper.Close()

	repoName := "search-test-repo"
	helper.CreateTestRepository(repoName, artifact.ArtifactTypeGeneric)

	// Push multiple artifacts
	artifacts := []struct {
		name    string
		version string
		content []byte
	}{
		{"search-test-1", "1.0.0", []byte("content1")},
		{"search-test-2", "1.0.0", []byte("content2")},
		{"other-artifact", "1.0.0", []byte("content3")},
	}

	for _, art := range artifacts {
		endpoint := fmt.Sprintf("/generic-repo/%s/%s/%s.txt", art.name, art.version, art.name)
		resp := helper.PushArtifact(endpoint, art.content)
		resp.Body.Close()
		require.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
	}

	// Search for "search-test"
	t.Run("Search_Matching", func(t *testing.T) {
		results, resp := helper.SearchArtifacts(repoName, "search-test")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Len(t, results, 2) // Should find search-test-1 and search-test-2
	})

	// Search for non-existent
	t.Run("Search_NonExistent", func(t *testing.T) {
		results, resp := helper.SearchArtifacts(repoName, "nonexistent")
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Len(t, results, 0)
	})
}

// TestConcurrentAccess tests concurrent artifact operations
func TestConcurrentAccess(t *testing.T) {
	helper := NewTestHelper("http://localhost:8080", t)
	defer helper.Close()

	repoName := "concurrent-test-repo"
	helper.CreateTestRepository(repoName, artifact.ArtifactTypeDocker)

	content := createMockDockerManifest()

	// Concurrent pushes
	t.Run("Concurrent_Pushes", func(t *testing.T) {
		done := make(chan bool, 5)

		for i := 0; i < 5; i++ {
			go func(id int) {
				defer func() { done <- true }()

				versionedEndpoint := fmt.Sprintf("/docker-repo/v2/test-image/manifests/v%d", id)
				resp := helper.PushArtifact(versionedEndpoint, content)
				resp.Body.Close()
				assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < 5; i++ {
			select {
			case <-done:
			case <-time.After(30 * time.Second):
				t.Fatal("Timeout waiting for concurrent pushes")
			}
		}
	})

	// Verify all artifacts were created
	t.Run("Verify_All_Created", func(t *testing.T) {
		artifacts, resp := helper.ListArtifacts(repoName)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Len(t, artifacts, 5)
	})
}

// TestLargeArtifact tests handling of large artifacts
func TestLargeArtifact(t *testing.T) {
	helper := NewTestHelper("http://localhost:8080", t)
	defer helper.Close()

	repoName := "large-artifact-repo"
	helper.CreateTestRepository(repoName, artifact.ArtifactTypeGeneric)

	// Create a 10MB artifact
	largeContent := make([]byte, 10*1024*1024)
	for i := range largeContent {
		largeContent[i] = byte(i % 256)
	}

	endpoint := "/generic-repo/large-file/1.0.0/large-file.bin"

	t.Run("Push_Large", func(t *testing.T) {
		resp := helper.PushArtifact(endpoint, largeContent)
		defer resp.Body.Close()
		assert.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
	})

	t.Run("Pull_Large", func(t *testing.T) {
		content, resp := helper.PullArtifact(endpoint)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, largeContent, content)
		assert.Equal(t, len(largeContent), len(content))
	})
}

// TestRepositoryManagement tests repository-level operations
func TestRepositoryManagement(t *testing.T) {
	helper := NewTestHelper("http://localhost:8080", t)
	defer helper.Close()

	repoName := "management-test-repo"
	helper.CreateTestRepository(repoName, artifact.ArtifactTypeHelm)

	// Get repository info
	t.Run("Get_Repository_Info", func(t *testing.T) {
		info, resp := helper.GetRepositoryInfo(repoName)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Equal(t, repoName, info["name"])
		assert.Equal(t, string(artifact.ArtifactTypeHelm), info["type"])
	})

	// Push some artifacts
	for i := 0; i < 3; i++ {
		endpoint := fmt.Sprintf("/helm-repo/charts/test-chart-%d-1.0.0.tgz", i)
		resp := helper.PushArtifact(endpoint, createMockHelmChart())
		resp.Body.Close()
		require.True(t, resp.StatusCode >= 200 && resp.StatusCode < 300)
	}

	// List artifacts
	t.Run("List_Repository_Artifacts", func(t *testing.T) {
		artifacts, resp := helper.ListArtifacts(repoName)
		defer resp.Body.Close()

		assert.Equal(t, http.StatusOK, resp.StatusCode)
		assert.Len(t, artifacts, 3)
	})
}
