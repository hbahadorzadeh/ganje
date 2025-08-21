package server

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/repository"
)

func TestMavenArtifactHandlers(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup authentication
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "maven-repo", auth.PermissionWrite).Return(true)
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "maven-repo", auth.PermissionRead).Return(true)
	
	// Setup repository
	mockRepo := &MockRepository{
		name:         "maven-repo",
		repoType:     "local",
		artifactType: "maven",
	}
	mockRepo.On("GetName").Return("maven-repo")
	mockRepo.On("GetType").Return(repository.Local)
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactTypeMaven)
	
	mockRepoManager.On("GetRepository", "maven-repo").Return(mockRepo, nil)
	
	// Register repository routes
	dbRepo := &database.Repository{
		Name:         "maven-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	server.RegisterRepositoryRoutes(dbRepo)
	
	t.Run("Upload Maven JAR", func(t *testing.T) {
		path := "/maven-repo/com/example/myapp/1.0.0/myapp-1.0.0.jar"
		content := "jar file content"
		
		mockRepo.On("Push", mock.Anything, "com/example/myapp/1.0.0/myapp-1.0.0.jar", mock.Anything, mock.Anything).Return(nil)
		mockDB.On("SaveArtifact", mock.Anything, mock.AnythingOfType("*database.ArtifactInfo")).Return(nil)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("PUT", path, strings.NewReader(content))
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/java-archive")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Download Maven JAR", func(t *testing.T) {
		path := "/maven-repo/com/example/myapp/1.0.0/myapp-1.0.0.jar"
		content := "jar file content"
		
		mockRepo.On("Pull", mock.Anything, "com/example/myapp/1.0.0/myapp-1.0.0.jar").Return(
			io.NopCloser(strings.NewReader(content)),
			&artifact.Metadata{Name: "myapp", Version: "1.0.0"},
			nil,
		)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, content, w.Body.String())
		assert.Equal(t, "application/java-archive", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Get Maven Metadata", func(t *testing.T) {
		path := "/maven-repo/com/example/myapp/maven-metadata.xml"
		metadata := `<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>com.example</groupId>
  <artifactId>myapp</artifactId>
  <versioning>
    <versions>
      <version>1.0.0</version>
    </versions>
  </versioning>
</metadata>`
		
		mockRepo.On("GetIndex", mock.Anything, "maven-metadata").Return(
			io.NopCloser(strings.NewReader(metadata)),
			nil,
		)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "com.example")
		assert.Contains(t, w.Body.String(), "myapp")
		assert.Equal(t, "application/xml", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
}

func TestNPMArtifactHandlers(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup authentication
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "npm-repo", auth.PermissionWrite).Return(true)
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "npm-repo", auth.PermissionRead).Return(true)
	
	// Setup repository
	mockRepo := &MockRepository{
		name:         "npm-repo",
		repoType:     "local",
		artifactType: "npm",
	}
	mockRepo.On("GetName").Return("npm-repo")
	mockRepo.On("GetType").Return(repository.Local)
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactTypeNPM)
	
	mockRepoManager.On("GetRepository", "npm-repo").Return(mockRepo, nil)
	
	t.Run("Publish NPM Package", func(t *testing.T) {
		path := "/npm-repo/express/-/express-4.18.1.tgz"
		content := "tgz file content"
		
		mockRepo.On("Push", mock.Anything, "express/-/express-4.18.1.tgz", mock.Anything, mock.Anything).Return(nil)
		mockDB.On("SaveArtifact", mock.Anything, mock.AnythingOfType("*database.ArtifactInfo")).Return(nil)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("PUT", path, strings.NewReader(content))
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/gzip")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Get NPM Package Info", func(t *testing.T) {
		path := "/npm-repo/express"
		packageInfo := `{
  "name": "express",
  "versions": {
    "4.18.1": {
      "name": "express",
      "version": "4.18.1",
      "description": "Fast, unopinionated, minimalist web framework",
      "dist": {
        "tarball": "http://localhost/npm-repo/express/-/express-4.18.1.tgz"
      }
    }
  }
}`
		
		mockRepo.On("GetIndex", mock.Anything, "package").Return(
			io.NopCloser(strings.NewReader(packageInfo)),
			nil,
		)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "express")
		assert.Contains(t, w.Body.String(), "4.18.1")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Download NPM Package", func(t *testing.T) {
		path := "/npm-repo/express/-/express-4.18.1.tgz"
		content := "tgz file content"
		
		mockRepo.On("Pull", mock.Anything, "express/-/express-4.18.1.tgz").Return(
			io.NopCloser(strings.NewReader(content)),
			&artifact.Metadata{Name: "express", Version: "4.18.1"},
			nil,
		)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, content, w.Body.String())
		assert.Equal(t, "application/gzip", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
}

func TestDockerArtifactHandlers(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup authentication
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "docker-repo", auth.PermissionWrite).Return(true)
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "docker-repo", auth.PermissionRead).Return(true)
	
	// Setup repository
	mockRepo := &MockRepository{
		name:         "docker-repo",
		repoType:     "local",
		artifactType: "docker",
	}
	mockRepo.On("GetName").Return("docker-repo")
	mockRepo.On("GetType").Return(repository.Local)
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactTypeDocker)
	
	mockRepoManager.On("GetRepository", "docker-repo").Return(mockRepo, nil)
	
	t.Run("Docker Registry API v2", func(t *testing.T) {
		path := "/docker-repo/v2/"
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
	})
	
	t.Run("List Tags", func(t *testing.T) {
		path := "/docker-repo/v2/library/nginx/tags/list"
		tags := `{
  "name": "library/nginx",
  "tags": ["latest", "1.21", "1.20"]
}`
		
		mockRepo.On("GetIndex", mock.Anything, "tags").Return(
			io.NopCloser(strings.NewReader(tags)),
			nil,
		)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "library/nginx")
		assert.Contains(t, w.Body.String(), "latest")
		assert.Equal(t, "application/json", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Get Manifest", func(t *testing.T) {
		path := "/docker-repo/v2/library/nginx/manifests/latest"
		manifest := `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.docker.container.image.v1+json",
    "size": 1234,
    "digest": "sha256:abc123"
  }
}`
		
		mockRepo.On("Pull", mock.Anything, "v2/library/nginx/manifests/latest").Return(
			io.NopCloser(strings.NewReader(manifest)),
			&artifact.Metadata{Name: "library/nginx", Version: "latest"},
			nil,
		)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "schemaVersion")
		assert.Equal(t, "application/vnd.docker.distribution.manifest.v2+json", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Push Manifest", func(t *testing.T) {
		path := "/docker-repo/v2/library/nginx/manifests/latest"
		manifest := `{
  "schemaVersion": 2,
  "mediaType": "application/vnd.docker.distribution.manifest.v2+json",
  "config": {
    "mediaType": "application/vnd.docker.container.image.v1+json",
    "size": 1234,
    "digest": "sha256:abc123"
  }
}`
		
		mockRepo.On("Push", mock.Anything, "v2/library/nginx/manifests/latest", mock.Anything, mock.Anything).Return(nil)
		mockDB.On("SaveArtifact", mock.Anything, mock.AnythingOfType("*database.ArtifactInfo")).Return(nil)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("PUT", path, strings.NewReader(manifest))
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/vnd.docker.distribution.manifest.v2+json")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestPyPIArtifactHandlers(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup authentication
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "pypi-repo", auth.PermissionWrite).Return(true)
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "pypi-repo", auth.PermissionRead).Return(true)
	
	// Setup repository
	mockRepo := &MockRepository{
		name:         "pypi-repo",
		repoType:     "local",
		artifactType: "pypi",
	}
	mockRepo.On("GetName").Return("pypi-repo")
	mockRepo.On("GetType").Return(repository.Local)
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactTypePyPI)
	
	mockRepoManager.On("GetRepository", "pypi-repo").Return(mockRepo, nil)
	
	t.Run("Upload Python Package", func(t *testing.T) {
		path := "/pypi-repo/packages/django/Django-4.1.0-py3-none-any.whl"
		content := "wheel file content"
		
		mockRepo.On("Push", mock.Anything, "packages/django/Django-4.1.0-py3-none-any.whl", mock.Anything, mock.Anything).Return(nil)
		mockDB.On("SaveArtifact", mock.Anything, mock.AnythingOfType("*database.ArtifactInfo")).Return(nil)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("PUT", path, strings.NewReader(content))
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/zip")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Get Simple Package Index", func(t *testing.T) {
		path := "/pypi-repo/simple/django/"
		index := `<!DOCTYPE html>
<html>
<head><title>Links for django</title></head>
<body>
<h1>Links for django</h1>
<a href="../../packages/django/Django-4.1.0-py3-none-any.whl">Django-4.1.0-py3-none-any.whl</a><br/>
</body>
</html>`
		
		mockRepo.On("GetIndex", mock.Anything, "simple").Return(
			io.NopCloser(strings.NewReader(index)),
			nil,
		)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "django")
		assert.Contains(t, w.Body.String(), "Django-4.1.0")
		assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Download Python Package", func(t *testing.T) {
		path := "/pypi-repo/packages/django/Django-4.1.0-py3-none-any.whl"
		content := "wheel file content"
		
		mockRepo.On("Pull", mock.Anything, "packages/django/Django-4.1.0-py3-none-any.whl").Return(
			io.NopCloser(strings.NewReader(content)),
			&artifact.Metadata{Name: "Django", Version: "4.1.0"},
			nil,
		)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, content, w.Body.String())
		assert.Equal(t, "application/zip", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
}

func TestGenericArtifactHandlers(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup authentication
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "generic-repo", auth.PermissionWrite).Return(true)
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "generic-repo", auth.PermissionRead).Return(true)
	
	// Setup repository
	mockRepo := &MockRepository{
		name:         "generic-repo",
		repoType:     "local",
		artifactType: "generic",
	}
	mockRepo.On("GetName").Return("generic-repo")
	mockRepo.On("GetType").Return(repository.Local)
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactTypeGeneric)
	
	mockRepoManager.On("GetRepository", "generic-repo").Return(mockRepo, nil)
	
	t.Run("Upload Generic File", func(t *testing.T) {
		path := "/generic-repo/files/document.pdf"
		content := "PDF file content"
		
		mockRepo.On("Push", mock.Anything, "files/document.pdf", mock.Anything, mock.Anything).Return(nil)
		mockDB.On("SaveArtifact", mock.Anything, mock.AnythingOfType("*database.ArtifactInfo")).Return(nil)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("PUT", path, strings.NewReader(content))
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/pdf")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusCreated, w.Code)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Download Generic File", func(t *testing.T) {
		path := "/generic-repo/files/document.pdf"
		content := "PDF file content"
		
		mockRepo.On("Pull", mock.Anything, "files/document.pdf").Return(
			io.NopCloser(strings.NewReader(content)),
			&artifact.Metadata{Name: "document.pdf", Version: "latest"},
			nil,
		)
		mockDB.On("LogAccess", mock.Anything, mock.AnythingOfType("*database.AccessLog")).Return(nil)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Equal(t, content, w.Body.String())
		assert.Equal(t, "application/pdf", w.Header().Get("Content-Type"))
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("List Generic Files", func(t *testing.T) {
		path := "/generic-repo/files/"
		files := []string{
			"files/document.pdf",
			"files/image.png",
			"files/archive.zip",
		}
		
		mockRepo.On("List", mock.Anything, "files/").Return(files, nil)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusOK, w.Code)
		assert.Contains(t, w.Body.String(), "document.pdf")
		assert.Contains(t, w.Body.String(), "image.png")
		assert.Contains(t, w.Body.String(), "archive.zip")
		mockRepo.AssertExpectations(t)
	})
}

func TestArtifactValidation(t *testing.T) {
	server, _, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup authentication
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "maven-repo", auth.PermissionWrite).Return(true)
	
	// Setup repository
	mockRepo := &MockRepository{
		name:         "maven-repo",
		repoType:     "local",
		artifactType: "maven",
	}
	mockRepo.On("GetName").Return("maven-repo")
	mockRepo.On("GetType").Return(repository.Local)
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactTypeMaven)
	
	mockRepoManager.On("GetRepository", "maven-repo").Return(mockRepo, nil)
	
	// Register repository routes
	dbRepo := &database.Repository{
		Name:         "maven-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	server.RegisterRepositoryRoutes(dbRepo)
	
	t.Run("Invalid JAR Upload", func(t *testing.T) {
		path := "/maven-repo/com/example/myapp/1.0.0/myapp-1.0.0.jar"
		content := "invalid jar content"
		
		// Mock validation failure
		mockRepo.On("Push", mock.Anything, "com/example/myapp/1.0.0/myapp-1.0.0.jar", mock.Anything, mock.Anything).Return(
			assert.AnError,
		)
		
		req := httptest.NewRequest("PUT", path, strings.NewReader(content))
		req.Header.Set("Authorization", "Bearer valid-token")
		req.Header.Set("Content-Type", "application/java-archive")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusBadRequest, w.Code)
		mockRepo.AssertExpectations(t)
	})
	
	t.Run("Artifact Not Found", func(t *testing.T) {
		path := "/maven-repo/com/example/nonexistent/1.0.0/nonexistent-1.0.0.jar"
		
		mockRepo.On("Pull", mock.Anything, "com/example/nonexistent/1.0.0/nonexistent-1.0.0.jar").Return(
			nil, nil, assert.AnError,
		)
		
		req := httptest.NewRequest("GET", path, nil)
		req.Header.Set("Authorization", "Bearer valid-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusNotFound, w.Code)
		mockRepo.AssertExpectations(t)
	})
}

func TestArtifactPermissions(t *testing.T) {
	server, _, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup repository
	mockRepo := &MockRepository{
		name:         "private-repo",
		repoType:     "local",
		artifactType: "maven",
	}
	mockRepo.On("GetName").Return("private-repo")
	mockRepo.On("GetType").Return(repository.Local)
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactTypeMaven)
	
	mockRepoManager.On("GetRepository", "private-repo").Return(mockRepo, nil)
	
	// Register repository routes
	dbRepo := &database.Repository{
		Name:         "private-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	server.RegisterRepositoryRoutes(dbRepo)
	
	t.Run("Unauthorized Access", func(t *testing.T) {
		path := "/private-repo/com/example/app/1.0.0/app-1.0.0.jar"
		
		req := httptest.NewRequest("GET", path, nil)
		// No authorization header
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})
	
	t.Run("Insufficient Permissions", func(t *testing.T) {
		path := "/private-repo/com/example/app/1.0.0/app-1.0.0.jar"
		content := "jar content"
		
		// Setup user with read-only access
		mockAuthService.On("ValidateToken", "Bearer readonly-token").Return(&auth.Claims{
			Username: "readonly",
			Email:    "readonly@example.com",
			Realms:   []string{"user"},
		}, nil)
		
		mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "private-repo", auth.PermissionWrite).Return(false)
		
		req := httptest.NewRequest("PUT", path, strings.NewReader(content))
		req.Header.Set("Authorization", "Bearer readonly-token")
		
		w := httptest.NewRecorder()
		server.router.ServeHTTP(w, req)
		
		assert.Equal(t, http.StatusForbidden, w.Code)
		mockAuthService.AssertExpectations(t)
	})
}
