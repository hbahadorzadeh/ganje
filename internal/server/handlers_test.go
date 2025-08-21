package server

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/repository"
	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

func TestBulkDeleteRepositories(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock repositories without artifacts
	mockDB.On("GetArtifactsByRepository", mock.Anything, "repo1").Return([]*database.ArtifactInfo{}, nil)
	mockDB.On("GetArtifactsByRepository", mock.Anything, "repo2").Return([]*database.ArtifactInfo{}, nil)
	
	mockRepoManager.On("DeleteRepository", "repo1").Return(nil)
	mockRepoManager.On("DeleteRepository", "repo2").Return(nil)
	
	// Create request
	requestBody := map[string]interface{}{
		"names": []string{"repo1", "repo2"},
		"force": false,
	}
	
	req := createAuthenticatedRequest("DELETE", "/api/v1/repositories", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	results := response["results"].(map[string]interface{})
	assert.Len(t, results, 2)
	
	summary := response["summary"].(map[string]interface{})
	assert.Equal(t, float64(2), summary["total"])
	assert.Equal(t, float64(2), summary["success"])
	assert.Equal(t, float64(0), summary["errors"])
	
	mockAuthService.AssertExpectations(t)
	mockRepoManager.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestBulkDeleteRepositoriesWithArtifacts(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock repository with artifacts
	artifacts := []*database.ArtifactInfo{
		{ID: 1, Name: "test-artifact"},
	}
	mockDB.On("GetArtifactsByRepository", mock.Anything, "repo-with-artifacts").Return(artifacts, nil)
	
	// Mock empty repository
	mockDB.On("GetArtifactsByRepository", mock.Anything, "empty-repo").Return([]*database.ArtifactInfo{}, nil)
	mockRepoManager.On("DeleteRepository", "empty-repo").Return(nil)
	
	// Create request
	requestBody := map[string]interface{}{
		"names": []string{"repo-with-artifacts", "empty-repo"},
		"force": false,
	}
	
	req := createAuthenticatedRequest("DELETE", "/api/v1/repositories", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	results := response["results"].(map[string]interface{})
	assert.Len(t, results, 2)
	
	// Check that repo with artifacts failed
	repoWithArtifactsResult := results["repo-with-artifacts"].(map[string]interface{})
	assert.False(t, repoWithArtifactsResult["success"].(bool))
	assert.Contains(t, repoWithArtifactsResult["error"], "contains artifacts")
	
	// Check that empty repo succeeded
	emptyRepoResult := results["empty-repo"].(map[string]interface{})
	assert.True(t, emptyRepoResult["success"].(bool))
	
	summary := response["summary"].(map[string]interface{})
	assert.Equal(t, float64(2), summary["total"])
	assert.Equal(t, float64(1), summary["success"])
	assert.Equal(t, float64(1), summary["errors"])
	
	mockAuthService.AssertExpectations(t)
	mockRepoManager.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestValidateRepositoryConfigInvalidType(t *testing.T) {
	server, _, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Create request with invalid artifact type
	requestBody := map[string]interface{}{
		"name":          "test-repo",
		"type":          "local",
		"artifact_type": "invalid-type",
	}
	
	req := createAuthenticatedRequest("POST", "/api/v1/repositories/validate", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Unsupported artifact type")
	assert.NotNil(t, response["supported_types"])
	
	mockAuthService.AssertExpectations(t)
}

func TestValidateRepositoryConfigExistingName(t *testing.T) {
	server, mockDB, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock existing repository
	existingRepo := &database.Repository{
		Name: "existing-repo",
	}
	mockDB.On("GetRepository", mock.Anything, "existing-repo").Return(existingRepo, nil)
	
	// Create request
	requestBody := map[string]interface{}{
		"name":          "existing-repo",
		"type":          "local",
		"artifact_type": "maven",
	}
	
	req := createAuthenticatedRequest("POST", "/api/v1/repositories/validate", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusConflict, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "already exists")
	
	mockAuthService.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestValidateRepositoryConfigRemoteWithoutURL(t *testing.T) {
	server, mockDB, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock repository not found (name available)
	mockDB.On("GetRepository", mock.Anything, "remote-repo").Return(nil, assert.AnError)
	
	// Create request for remote repo without URL
	requestBody := map[string]interface{}{
		"name":          "remote-repo",
		"type":          "remote",
		"artifact_type": "maven",
	}
	
	req := createAuthenticatedRequest("POST", "/api/v1/repositories/validate", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "URL is required for remote repositories")
	
	mockAuthService.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestValidateRepositoryConfigVirtualWithoutUpstream(t *testing.T) {
	server, mockDB, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock repository not found (name available)
	mockDB.On("GetRepository", mock.Anything, "virtual-repo").Return(nil, assert.AnError)
	
	// Create request for virtual repo without upstream
	requestBody := map[string]interface{}{
		"name":          "virtual-repo",
		"type":          "virtual",
		"artifact_type": "maven",
		"upstream":      []string{},
	}
	
	req := createAuthenticatedRequest("POST", "/api/v1/repositories/validate", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Upstream repositories are required")
	
	mockAuthService.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestUpdateRepositoryWithArtifactTypeChange(t *testing.T) {
	server, mockDB, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	existingRepo := &database.Repository{
		Name:         "test-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	
	// Mock repository with artifacts - should prevent artifact type change
	artifacts := []*database.ArtifactInfo{
		{ID: 1, Name: "test-artifact"},
	}
	
	mockDB.On("GetRepository", mock.Anything, "test-repo").Return(existingRepo, nil)
	mockDB.On("GetArtifactsByRepository", mock.Anything, "test-repo").Return(artifacts, nil)
	
	// Create request to change artifact type
	requestBody := map[string]interface{}{
		"artifact_type": "npm",
	}
	
	req := createAuthenticatedRequest("PUT", "/api/v1/repositories/test-repo", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "Cannot change artifact type")
	
	mockAuthService.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestDeleteRepositoryWithArtifacts(t *testing.T) {
	server, mockDB, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock repository with artifacts
	artifacts := []*database.ArtifactInfo{
		{ID: 1, Name: "test-artifact"},
		{ID: 2, Name: "another-artifact"},
	}
	mockDB.On("GetArtifactsByRepository", mock.Anything, "repo-with-artifacts").Return(artifacts, nil)
	
	// Create request without force flag
	req := createAuthenticatedRequest("DELETE", "/api/v1/repositories/repo-with-artifacts", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusBadRequest, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Contains(t, response["error"], "contains artifacts")
	assert.Equal(t, float64(2), response["artifact_count"])
	
	mockAuthService.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestDeleteRepositoryWithArtifactsForced(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock repository with artifacts
	artifacts := []*database.ArtifactInfo{
		{ID: 1, Name: "test-artifact"},
	}
	mockDB.On("GetArtifactsByRepository", mock.Anything, "repo-with-artifacts").Return(artifacts, nil)
	mockRepoManager.On("DeleteRepository", "repo-with-artifacts").Return(nil)
	
	// Create request with force flag
	req := createAuthenticatedRequest("DELETE", "/api/v1/repositories/repo-with-artifacts?force=true", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Repository deleted successfully", response["message"])
	
	mockAuthService.AssertExpectations(t)
	mockRepoManager.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestGetRepositoryDetailed(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	dbRepo := &database.Repository{
		Name:         "detailed-repo",
		Type:         "local",
		ArtifactType: "maven",
		URL:          "",
	}
	
	mockRepo := &MockRepository{
		name:         "detailed-repo",
		repoType:     "local",
		artifactType: "maven",
	}
	
	// Setup mock repository expectations
	mockRepo.On("GetName").Return("detailed-repo")
	mockRepo.On("GetType").Return(repository.Type("local"))
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactType("maven"))
	
	stats := &repository.Statistics{
		TotalArtifacts: 42,
		TotalSize:      1048576,
		PullCount:      150,
		PushCount:      42,
	}
	
	artifacts := []*database.ArtifactInfo{
		{ID: 1, Name: "artifact1"},
		{ID: 2, Name: "artifact2"},
	}
	
	mockDB.On("GetRepository", mock.Anything, "detailed-repo").Return(dbRepo, nil)
	mockRepoManager.On("GetRepository", "detailed-repo").Return(mockRepo, nil)
	mockRepo.On("GetStatistics", mock.Anything).Return(stats, nil)
	mockDB.On("GetArtifactsByRepository", mock.Anything, "detailed-repo").Return(artifacts, nil)
	
	// Create request
	req := createAuthenticatedRequest("GET", "/api/v1/repositories/detailed-repo", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	assert.Equal(t, "detailed-repo", response["name"])
	assert.Equal(t, "local", response["type"])
	assert.Equal(t, "maven", response["artifact_type"])
	assert.Equal(t, float64(2), response["artifact_count"])
	assert.NotNil(t, response["statistics"])
	
	mockAuthService.AssertExpectations(t)
	mockRepoManager.AssertExpectations(t)
	mockDB.AssertExpectations(t)
	mockRepo.AssertExpectations(t)
}
