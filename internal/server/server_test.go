package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/repository"
)

// MockDB is a mock implementation of the database
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetRepository(ctx context.Context, name string) (*database.Repository, error) {
	args := m.Called(ctx, name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.Repository), args.Error(1)
}

func (m *MockDB) SaveRepository(ctx context.Context, repo *database.Repository) error {
	args := m.Called(ctx, repo)
	return args.Error(0)
}

func (m *MockDB) UpdateRepository(ctx context.Context, name string, updates map[string]interface{}) error {
	args := m.Called(ctx, name, updates)
	return args.Error(0)
}

func (m *MockDB) DeleteRepository(ctx context.Context, name string) error {
	args := m.Called(ctx, name)
	return args.Error(0)
}

func (m *MockDB) ListRepositories(ctx context.Context) ([]*database.Repository, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*database.Repository), args.Error(1)
}

func (m *MockDB) GetAllRepositories() ([]database.Repository, error) {
	args := m.Called()
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]database.Repository), args.Error(1)
}

func (m *MockDB) GetArtifactsByRepository(ctx context.Context, repositoryName string) ([]*database.ArtifactInfo, error) {
	args := m.Called(ctx, repositoryName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*database.ArtifactInfo), args.Error(1)
}

func (m *MockDB) SaveArtifact(ctx context.Context, artifact *database.ArtifactInfo) error {
	args := m.Called(ctx, artifact)
	return args.Error(0)
}

func (m *MockDB) GetArtifact(ctx context.Context, repositoryName, name, version string) (*database.ArtifactInfo, error) {
	args := m.Called(ctx, repositoryName, name, version)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.ArtifactInfo), args.Error(1)
}

func (m *MockDB) GetArtifactByPath(ctx context.Context, repositoryName, path string) (*database.ArtifactInfo, error) {
	args := m.Called(ctx, repositoryName, path)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.ArtifactInfo), args.Error(1)
}

func (m *MockDB) DeleteArtifactByPath(ctx context.Context, repositoryName, path string) error {
	args := m.Called(ctx, repositoryName, path)
	return args.Error(0)
}

func (m *MockDB) IncrementPullCount(ctx context.Context, artifactID uint) error {
	args := m.Called(ctx, artifactID)
	return args.Error(0)
}

func (m *MockDB) GetRepositoryStatistics(ctx context.Context, repositoryName string) (*database.Statistics, error) {
	args := m.Called(ctx, repositoryName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*database.Statistics), args.Error(1)
}

func (m *MockDB) LogAccess(ctx context.Context, log *database.AccessLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// MockRepositoryManager is a mock implementation of the repository manager
type MockRepositoryManager struct {
	mock.Mock
}

func (m *MockRepositoryManager) GetRepository(name string) (repository.Repository, error) {
	args := m.Called(name)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repository.Repository), args.Error(1)
}

func (m *MockRepositoryManager) ListRepositories() []repository.Repository {
	args := m.Called()
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).([]repository.Repository)
}

func (m *MockRepositoryManager) CreateRepository(config *repository.Config) (repository.Repository, error) {
	args := m.Called(config)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(repository.Repository), args.Error(1)
}

func (m *MockRepositoryManager) DeleteRepository(name string) error {
	args := m.Called(name)
	return args.Error(0)
}

// MockRepository is a mock implementation of repository.Repository
type MockRepository struct {
	mock.Mock
	name         string
	repoType     string
	artifactType string
}

func (m *MockRepository) GetName() string {
	args := m.Called()
	return args.String(0)
}

func (m *MockRepository) GetType() repository.Type {
	args := m.Called()
	return args.Get(0).(repository.Type)
}

func (m *MockRepository) GetArtifactType() artifact.ArtifactType {
	args := m.Called()
	return args.Get(0).(artifact.ArtifactType)
}

func (m *MockRepository) Pull(ctx context.Context, path string) (io.ReadCloser, *artifact.Metadata, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(io.ReadCloser), args.Get(1).(*artifact.Metadata), args.Error(2)
}

func (m *MockRepository) Push(ctx context.Context, path string, content io.Reader, metadata *artifact.Metadata) error {
	args := m.Called(ctx, path, content, metadata)
	return args.Error(0)
}

func (m *MockRepository) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockRepository) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockRepository) GetIndex(ctx context.Context, indexType string) (io.ReadCloser, error) {
	args := m.Called(ctx, indexType)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockRepository) InvalidateCache(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockRepository) RebuildIndex(ctx context.Context) error {
	args := m.Called(ctx)
	return args.Error(0)
}

func (m *MockRepository) GetStatistics(ctx context.Context) (*repository.Statistics, error) {
	args := m.Called(ctx)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*repository.Statistics), args.Error(1)
}

// MockAuthService is a mock implementation of the auth service
type MockAuthService struct {
	mock.Mock
}

func (m *MockAuthService) ValidateToken(token string) (*auth.Claims, error) {
	args := m.Called(token)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*auth.Claims), args.Error(1)
}

func (m *MockAuthService) CheckPermission(claims *auth.Claims, resource string, permission auth.Permission) bool {
	args := m.Called(claims, resource, permission)
	return args.Bool(0)
}

// createTestServer creates a test server with mocked dependencies
func createTestServer() (*Server, *MockDB, *MockRepositoryManager, *MockAuthService) {
	gin.SetMode(gin.TestMode)
	
	mockDB := &MockDB{}
	mockRepoManager := &MockRepositoryManager{}
	mockAuthService := &MockAuthService{}
	
	// Setup default mock for ListRepositories that gets called during setupRoutes
	mockDB.On("ListRepositories", mock.Anything).Return([]*database.Repository{}, nil)
	
	server := &Server{
		config:        &config.Config{},
		db:            mockDB,
		repoManager:   mockRepoManager,
		authService:   mockAuthService,
		routeRegistry: NewRouteRegistry(),
	}
	
	server.setupRoutes()
	
	return server, mockDB, mockRepoManager, mockAuthService
}

// createAuthenticatedRequest creates a request with valid authentication
func createAuthenticatedRequest(method, url string, body interface{}) *http.Request {
	var reqBody *bytes.Buffer
	if body != nil {
		jsonBody, _ := json.Marshal(body)
		reqBody = bytes.NewBuffer(jsonBody)
	} else {
		reqBody = bytes.NewBuffer(nil)
	}
	
	req := httptest.NewRequest(method, url, reqBody)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer valid-token")
	
	return req
}

func TestListRepositories(t *testing.T) {
	server, _, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockRepo := &MockRepository{
		name:         "test-repo",
		repoType:     "local",
		artifactType: "maven",
	}
	
	// Setup mock expectations for the repository
	mockRepo.On("GetName").Return("test-repo")
	mockRepo.On("GetType").Return(repository.Type("local"))
	mockRepo.On("GetArtifactType").Return(artifact.ArtifactType("maven"))
	
	mockRepoManager.On("ListRepositories").Return([]repository.Repository{mockRepo})
	
	// Create request
	req := createAuthenticatedRequest("GET", "/api/v1/repositories", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	repositories := response["repositories"].([]interface{})
	assert.Len(t, repositories, 1)
	
	repo := repositories[0].(map[string]interface{})
	assert.Equal(t, "test-repo", repo["name"])
	assert.Equal(t, "local", repo["type"])
	assert.Equal(t, "maven", repo["artifact_type"])
	
	mockAuthService.AssertExpectations(t)
	mockRepoManager.AssertExpectations(t)
}

func TestCreateRepository(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	mockRepo := &MockRepository{
		name:         "new-repo",
		repoType:     "local",
		artifactType: "maven",
	}
	
	mockRepoManager.On("CreateRepository", mock.AnythingOfType("*repository.Config")).Return(mockRepo, nil)
	
	dbRepo := &database.Repository{
		Name:         "new-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	mockDB.On("GetRepository", mock.Anything, "new-repo").Return(dbRepo, nil)
	
	// Create request
	requestBody := map[string]interface{}{
		"name":          "new-repo",
		"type":          "local",
		"artifact_type": "maven",
	}
	
	req := createAuthenticatedRequest("POST", "/api/v1/repositories", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusCreated, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Repository created successfully", response["message"])
	
	mockAuthService.AssertExpectations(t)
	mockRepoManager.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestUpdateRepository(t *testing.T) {
	server, mockDB, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	existingRepo := &database.Repository{
		Name:         "existing-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	
	mockDB.On("GetRepository", mock.Anything, "existing-repo").Return(existingRepo, nil)
	mockDB.On("UpdateRepository", mock.Anything, "existing-repo", mock.AnythingOfType("map[string]interface {}")).Return(nil)
	
	// Create request
	requestBody := map[string]interface{}{
		"type": "remote",
		"url":  "https://repo1.maven.org/maven2/",
	}
	
	req := createAuthenticatedRequest("PUT", "/api/v1/repositories/existing-repo", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.Equal(t, "Repository updated successfully", response["message"])
	
	mockAuthService.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestDeleteRepository(t *testing.T) {
	server, mockDB, mockRepoManager, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	mockDB.On("GetArtifactsByRepository", mock.Anything, "test-repo").Return([]*database.ArtifactInfo{}, nil)
	mockRepoManager.On("DeleteRepository", "test-repo").Return(nil)
	
	// Create request
	req := createAuthenticatedRequest("DELETE", "/api/v1/repositories/test-repo", nil)
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

func TestValidateRepositoryConfig(t *testing.T) {
	server, mockDB, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	mockAuthService.On("CheckPermission", mock.AnythingOfType("*auth.Claims"), "", auth.PermissionAdmin).Return(true)
	
	// Mock repository not found (which means name is available)
	mockDB.On("GetRepository", mock.Anything, "new-repo").Return(nil, assert.AnError)
	
	// Create request
	requestBody := map[string]interface{}{
		"name":          "new-repo",
		"type":          "local",
		"artifact_type": "maven",
	}
	
	req := createAuthenticatedRequest("POST", "/api/v1/repositories/validate", requestBody)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.True(t, response["valid"].(bool))
	assert.Equal(t, "Repository configuration is valid", response["message"])
	
	mockAuthService.AssertExpectations(t)
	mockDB.AssertExpectations(t)
}

func TestGetRepositoryTypes(t *testing.T) {
	server, _, _, mockAuthService := createTestServer()
	
	// Setup mocks
	mockAuthService.On("ValidateToken", "Bearer valid-token").Return(&auth.Claims{
		Username: "testuser",
		Email:    "test@example.com",
		Realms:   []string{"admin"},
	}, nil)
	
	// Create request
	req := createAuthenticatedRequest("GET", "/api/v1/repository-types", nil)
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusOK, w.Code)
	
	var response map[string]interface{}
	err := json.Unmarshal(w.Body.Bytes(), &response)
	assert.NoError(t, err)
	
	repoTypes := response["repository_types"].([]interface{})
	assert.Contains(t, repoTypes, "local")
	assert.Contains(t, repoTypes, "remote")
	assert.Contains(t, repoTypes, "virtual")
	
	artifactTypes := response["artifact_types"].([]interface{})
	assert.Contains(t, artifactTypes, "maven")
	assert.Contains(t, artifactTypes, "npm")
	assert.Contains(t, artifactTypes, "docker")
	
	mockAuthService.AssertExpectations(t)
}

func TestUnauthorizedAccess(t *testing.T) {
	server, _, _, mockAuthService := createTestServer()
	
	// Setup mocks for invalid token
	mockAuthService.On("ValidateToken", "Bearer invalid-token").Return(nil, assert.AnError)
	
	// Create request with invalid token
	req := httptest.NewRequest("GET", "/api/v1/repositories", nil)
	req.Header.Set("Authorization", "Bearer invalid-token")
	w := httptest.NewRecorder()
	
	// Execute request
	server.router.ServeHTTP(w, req)
	
	// Assert response
	assert.Equal(t, http.StatusUnauthorized, w.Code)
	
	mockAuthService.AssertExpectations(t)
}
