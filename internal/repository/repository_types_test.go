package repository

import (
	"context"
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/database"
)

// MockStorage for testing
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Store(ctx context.Context, path string, content io.Reader) error {
	args := m.Called(ctx, path, content)
	return args.Error(0)
}

func (m *MockStorage) Retrieve(ctx context.Context, path string) (io.ReadCloser, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(io.ReadCloser), args.Error(1)
}

func (m *MockStorage) Delete(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

func (m *MockStorage) List(ctx context.Context, prefix string) ([]string, error) {
	args := m.Called(ctx, prefix)
	return args.Get(0).([]string), args.Error(1)
}

func (m *MockStorage) Exists(ctx context.Context, path string) (bool, error) {
	args := m.Called(ctx, path)
	return args.Bool(0), args.Error(1)
}

func (m *MockStorage) GetSize(ctx context.Context, path string) (int64, error) {
	args := m.Called(ctx, path)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockStorage) GetChecksum(ctx context.Context, path string) (string, error) {
	args := m.Called(ctx, path)
	return args.String(0), args.Error(1)
}

// MockArtifactFactory for testing
type MockArtifactFactory struct {
	mock.Mock
}

func (m *MockArtifactFactory) CreateArtifact(artifactType artifact.ArtifactType) (artifact.Artifact, error) {
	args := m.Called(artifactType)
	return args.Get(0).(artifact.Artifact), args.Error(1)
}

func (m *MockArtifactFactory) GetSupportedTypes() []artifact.ArtifactType {
	args := m.Called()
	return args.Get(0).([]artifact.ArtifactType)
}

// MockDB for testing
type MockDB struct {
	mock.Mock
}

func (m *MockDB) GetRepository(ctx context.Context, name string) (*database.Repository, error) {
	args := m.Called(ctx, name)
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
	return args.Get(0).([]*database.Repository), args.Error(1)
}

func (m *MockDB) GetArtifactsByRepository(ctx context.Context, repositoryName string) ([]*database.ArtifactInfo, error) {
	args := m.Called(ctx, repositoryName)
	return args.Get(0).([]*database.ArtifactInfo), args.Error(1)
}

func (m *MockDB) SaveArtifact(ctx context.Context, artifact *database.ArtifactInfo) error {
	args := m.Called(ctx, artifact)
	return args.Error(0)
}

func (m *MockDB) GetArtifact(ctx context.Context, repositoryName, name, version string) (*database.ArtifactInfo, error) {
	args := m.Called(ctx, repositoryName, name, version)
	return args.Get(0).(*database.ArtifactInfo), args.Error(1)
}

func (m *MockDB) GetArtifactByPath(ctx context.Context, repositoryName, path string) (*database.ArtifactInfo, error) {
	args := m.Called(ctx, repositoryName, path)
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
	return args.Get(0).(*database.Statistics), args.Error(1)
}

func (m *MockDB) LogAccess(ctx context.Context, log *database.AccessLog) error {
	args := m.Called(ctx, log)
	return args.Error(0)
}

// MockArtifact for testing
type MockArtifact struct {
	mock.Mock
	artifactType artifact.ArtifactType
}

func (m *MockArtifact) GetType() artifact.ArtifactType {
	return m.artifactType
}

func (m *MockArtifact) ParsePath(path string) (*artifact.ArtifactInfo, error) {
	args := m.Called(path)
	return args.Get(0).(*artifact.ArtifactInfo), args.Error(1)
}

func (m *MockArtifact) GeneratePath(info *artifact.ArtifactInfo) string {
	args := m.Called(info)
	return args.String(0)
}

func (m *MockArtifact) ValidateArtifact(content io.Reader) error {
	args := m.Called(content)
	return args.Error(0)
}

func (m *MockArtifact) GetMetadata(content io.Reader) (map[string]string, error) {
	args := m.Called(content)
	return args.Get(0).(map[string]string), args.Error(1)
}

func (m *MockArtifact) GenerateIndex(artifacts []*artifact.ArtifactInfo) ([]byte, error) {
	args := m.Called(artifacts)
	return args.Get(0).([]byte), args.Error(1)
}

func (m *MockArtifact) GetEndpoints() []string {
	args := m.Called()
	return args.Get(0).([]string)
}

func TestLocalRepository(t *testing.T) {
	mockStorage := &MockStorage{}
	mockArtifactFactory := &MockArtifactFactory{}
	mockArtifact := &MockArtifact{artifactType: artifact.ArtifactTypeMaven}
	mockDB := &MockDB{}
	
	config := &Config{
		Name:         "local-maven-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	
	// Setup mocks
	mockArtifactFactory.On("CreateArtifact", artifact.ArtifactTypeMaven).Return(mockArtifact, nil)
	
	// Create repository directly instead of using factory to pass mockDB
	repo := NewLocalRepository(config.Name, artifact.ArtifactTypeMaven, mockStorage, mockArtifactFactory, mockDB)
	assert.NotNil(t, repo)
	
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "local-maven-repo", repo.GetName())
	})
	
	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, Local, repo.GetType())
	})
	
	t.Run("GetArtifactType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeMaven, repo.GetArtifactType())
	})
	
	t.Run("Push", func(t *testing.T) {
		ctx := context.Background()
		path := "com/example/test/1.0.0/test-1.0.0.jar"
		content := strings.NewReader("jar content")
		metadata := &artifact.Metadata{
			Name:    "test",
			Version: "1.0.0",
		}
		
		mockArtifact.On("ValidateArtifact", mock.Anything).Return(nil)
		mockStorage.On("Store", ctx, path, mock.Anything).Return(nil)
		mockStorage.On("GetSize", ctx, path).Return(int64(100), nil)
		mockStorage.On("GetChecksum", ctx, path).Return("sha256:abc123", nil)
		mockDB.On("SaveArtifact", ctx, mock.AnythingOfType("*database.ArtifactInfo")).Return(nil)
		
		err := repo.Push(ctx, path, content, metadata)
		assert.NoError(t, err)
		
		mockArtifact.AssertExpectations(t)
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("Pull", func(t *testing.T) {
		ctx := context.Background()
		path := "com/example/test/1.0.0/test-1.0.0.jar"
		
		mockStorage.On("Exists", ctx, path).Return(true, nil)
		expectedContent := io.NopCloser(strings.NewReader("jar content"))
		mockStorage.On("Retrieve", ctx, path).Return(expectedContent, nil)
		mockDB.On("GetArtifactByPath", ctx, "local-maven-repo", path).Return(&database.ArtifactInfo{
			Name: "test", Version: "1.0.0",
		}, nil)
		mockDB.On("IncrementPullCount", ctx, mock.AnythingOfType("uint")).Return(nil)
		
		content, metadata, err := repo.Pull(ctx, path)
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.NotNil(t, metadata)
		
		mockStorage.AssertExpectations(t)
		mockDB.AssertExpectations(t)
	})
	
	t.Run("Delete", func(t *testing.T) {
		ctx := context.Background()
		path := "com/example/test/1.0.0/test-1.0.0.jar"
		
		mockStorage.On("Delete", ctx, path).Return(nil)
		
		err := repo.Delete(ctx, path)
		assert.NoError(t, err)
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("List", func(t *testing.T) {
		ctx := context.Background()
		prefix := "com/example/"
		
		expectedFiles := []string{
			"com/example/test/1.0.0/test-1.0.0.jar",
			"com/example/test/1.0.0/test-1.0.0.pom",
		}
		mockStorage.On("List", ctx, prefix).Return(expectedFiles, nil)
		
		files, err := repo.List(ctx, prefix)
		assert.NoError(t, err)
		assert.Equal(t, expectedFiles, files)
		
		mockStorage.AssertExpectations(t)
	})
}

func TestRemoteRepository(t *testing.T) {
	mockStorage := &MockStorage{}
	mockArtifactFactory := &MockArtifactFactory{}
	mockArtifact := &MockArtifact{artifactType: artifact.ArtifactTypeNPM}
	
	config := &Config{
		Name:         "remote-npm-repo",
		Type:         "remote",
		ArtifactType: "npm",
		URL:          "https://registry.npmjs.org",
	}
	
	// Setup mocks
	mockArtifactFactory.On("CreateArtifact", artifact.ArtifactTypeNPM).Return(mockArtifact, nil)
	
	factory := &DefaultFactory{}
	repo, err := factory.CreateRepository(config, mockStorage, mockArtifactFactory)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "remote-npm-repo", repo.GetName())
	})
	
	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, Remote, repo.GetType())
	})
	
	t.Run("GetArtifactType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeNPM, repo.GetArtifactType())
	})
	
	t.Run("Pull with caching", func(t *testing.T) {
		ctx := context.Background()
		path := "express/-/express-4.18.1.tgz"
		
		// First check cache (miss)
		mockStorage.On("Exists", ctx, path).Return(false, nil).Once()
		
		// Cache the artifact after fetching from remote
		mockStorage.On("Store", ctx, path, mock.Anything).Return(nil).Once()
		
		// Return cached content
		expectedContent := io.NopCloser(strings.NewReader("tgz content"))
		mockStorage.On("Retrieve", ctx, path).Return(expectedContent, nil).Once()
		
		content, metadata, err := repo.Pull(ctx, path)
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.NotNil(t, metadata)
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("Pull from cache", func(t *testing.T) {
		ctx := context.Background()
		path := "express/-/express-4.18.1.tgz"
		
		// Cache hit
		mockStorage.On("Exists", ctx, path).Return(true, nil).Once()
		
		// Return cached content directly
		expectedContent := io.NopCloser(strings.NewReader("cached tgz content"))
		mockStorage.On("Retrieve", ctx, path).Return(expectedContent, nil).Once()
		
		content, metadata, err := repo.Pull(ctx, path)
		assert.NoError(t, err)
		assert.NotNil(t, content)
		assert.NotNil(t, metadata)
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("InvalidateCache", func(t *testing.T) {
		ctx := context.Background()
		path := "express/-/express-4.18.1.tgz"
		
		mockStorage.On("Delete", ctx, path).Return(nil)
		
		err := repo.InvalidateCache(ctx, path)
		assert.NoError(t, err)
		
		mockStorage.AssertExpectations(t)
	})
}

func TestVirtualRepository(t *testing.T) {
	mockStorage := &MockStorage{}
	mockArtifactFactory := &MockArtifactFactory{}
	mockArtifact := &MockArtifact{artifactType: artifact.ArtifactTypeDocker}
	
	config := &Config{
		Name:         "virtual-docker-repo",
		Type:         "virtual",
		ArtifactType: "docker",
		Upstream:     []string{"local-docker", "remote-docker"},
	}
	
	// Setup mocks
	mockArtifactFactory.On("CreateArtifact", artifact.ArtifactTypeDocker).Return(mockArtifact, nil)
	
	factory := &DefaultFactory{}
	repo, err := factory.CreateRepository(config, mockStorage, mockArtifactFactory)
	assert.NoError(t, err)
	assert.NotNil(t, repo)
	
	t.Run("GetName", func(t *testing.T) {
		assert.Equal(t, "virtual-docker-repo", repo.GetName())
	})
	
	t.Run("GetType", func(t *testing.T) {
		assert.Equal(t, Virtual, repo.GetType())
	})
	
	t.Run("GetArtifactType", func(t *testing.T) {
		assert.Equal(t, artifact.ArtifactTypeDocker, repo.GetArtifactType())
	})
	
	t.Run("Pull from upstream repositories", func(t *testing.T) {
		ctx := context.Background()
		path := "v2/library/nginx/manifests/latest"
		
		// Virtual repository should try upstream repositories
		// This would require more complex mocking of upstream repositories
		// For now, we'll test the basic structure
		
		_, _, err := repo.Pull(ctx, path)
		// Virtual repository implementation would handle upstream resolution
		// The exact behavior depends on the implementation details
		assert.Error(t, err) // Expected since we don't have real upstream repos
	})
	
	t.Run("Push not allowed", func(t *testing.T) {
		ctx := context.Background()
		path := "v2/library/nginx/manifests/latest"
		content := strings.NewReader("manifest content")
		metadata := &artifact.Metadata{Name: "nginx", Version: "latest"}
		
		err := repo.Push(ctx, path, content, metadata)
		assert.Error(t, err) // Virtual repositories typically don't allow direct pushes
	})
}

func TestRepositoryStatistics(t *testing.T) {
	mockStorage := &MockStorage{}
	mockArtifactFactory := &MockArtifactFactory{}
	mockArtifact := &MockArtifact{artifactType: artifact.ArtifactTypeMaven}
	
	config := &Config{
		Name:         "stats-test-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	
	// Setup mocks
	mockArtifactFactory.On("CreateArtifact", artifact.ArtifactTypeMaven).Return(mockArtifact, nil)
	
	factory := &DefaultFactory{}
	repo, err := factory.CreateRepository(config, mockStorage, mockArtifactFactory)
	assert.NoError(t, err)
	
	t.Run("GetStatistics", func(t *testing.T) {
		ctx := context.Background()
		
		// Mock storage list to return some artifacts
		artifacts := []string{
			"com/example/app1/1.0.0/app1-1.0.0.jar",
			"com/example/app1/1.0.0/app1-1.0.0.pom",
			"com/example/app2/2.0.0/app2-2.0.0.jar",
		}
		mockStorage.On("List", ctx, "").Return(artifacts, nil)
		
		stats, err := repo.GetStatistics(ctx)
		assert.NoError(t, err)
		assert.NotNil(t, stats)
		assert.Equal(t, int64(3), stats.TotalArtifacts)
		assert.GreaterOrEqual(t, stats.TotalSize, int64(0))
		
		mockStorage.AssertExpectations(t)
	})
}

func TestRepositoryIndexing(t *testing.T) {
	mockStorage := &MockStorage{}
	mockArtifactFactory := &MockArtifactFactory{}
	mockArtifact := &MockArtifact{artifactType: artifact.ArtifactTypeMaven}
	
	config := &Config{
		Name:         "index-test-repo",
		Type:         "local",
		ArtifactType: "maven",
	}
	
	// Setup mocks
	mockArtifactFactory.On("CreateArtifact", artifact.ArtifactTypeMaven).Return(mockArtifact, nil)
	
	factory := &DefaultFactory{}
	repo, err := factory.CreateRepository(config, mockStorage, mockArtifactFactory)
	assert.NoError(t, err)
	
	t.Run("GetIndex", func(t *testing.T) {
		ctx := context.Background()
		indexType := "maven-metadata"
		
		expectedIndex := io.NopCloser(strings.NewReader(`<?xml version="1.0" encoding="UTF-8"?>
<metadata>
  <groupId>com.example</groupId>
  <artifactId>test-app</artifactId>
  <versioning>
    <versions>
      <version>1.0.0</version>
      <version>1.1.0</version>
    </versions>
  </versioning>
</metadata>`))
		
		mockStorage.On("Retrieve", ctx, ".index/maven-metadata.xml").Return(expectedIndex, nil)
		
		index, err := repo.GetIndex(ctx, indexType)
		assert.NoError(t, err)
		assert.NotNil(t, index)
		
		mockStorage.AssertExpectations(t)
	})
	
	t.Run("RebuildIndex", func(t *testing.T) {
		ctx := context.Background()
		
		// Mock artifacts for index generation
		artifacts := []string{
			"com/example/app/1.0.0/app-1.0.0.jar",
			"com/example/app/1.1.0/app-1.1.0.jar",
		}
		mockStorage.On("List", ctx, "").Return(artifacts, nil)
		
		// Mock artifact parsing
		artifactInfo := &artifact.ArtifactInfo{
			Name:    "app",
			Version: "1.0.0",
			Type:    "jar",
		}
		mockArtifact.On("ParsePath", mock.AnythingOfType("string")).Return(artifactInfo, nil)
		
		// Mock index generation
		indexData := []byte("generated index")
		mockArtifact.On("GenerateIndex", mock.AnythingOfType("[]*artifact.ArtifactInfo")).Return(indexData, nil)
		
		// Mock storing the generated index
		mockStorage.On("Store", ctx, ".index/maven-metadata.xml", mock.Anything).Return(nil)
		
		err := repo.RebuildIndex(ctx)
		assert.NoError(t, err)
		
		mockStorage.AssertExpectations(t)
		mockArtifact.AssertExpectations(t)
	})
}

func TestRepositoryValidation(t *testing.T) {
	tests := []struct {
		name      string
		config    *Config
		shouldErr bool
		errMsg    string
	}{
		{
			name: "valid local repository",
			config: &Config{
				Name:         "valid-local",
				Type:         "local",
				ArtifactType: "maven",
			},
			shouldErr: false,
		},
		{
			name: "valid remote repository",
			config: &Config{
				Name:         "valid-remote",
				Type:         "remote",
				ArtifactType: "npm",
				URL:          "https://registry.npmjs.org",
			},
			shouldErr: false,
		},
		{
			name: "valid virtual repository",
			config: &Config{
				Name:         "valid-virtual",
				Type:         "virtual",
				ArtifactType: "docker",
				Upstream:     []string{"local-docker", "remote-docker"},
			},
			shouldErr: false,
		},
		{
			name: "remote repository without URL",
			config: &Config{
				Name:         "invalid-remote",
				Type:         "remote",
				ArtifactType: "maven",
			},
			shouldErr: true,
			errMsg:    "URL is required for remote repositories",
		},
		{
			name: "virtual repository without upstream",
			config: &Config{
				Name:         "invalid-virtual",
				Type:         "virtual",
				ArtifactType: "maven",
			},
			shouldErr: true,
			errMsg:    "upstream repositories are required for virtual repositories",
		},
		{
			name: "unsupported repository type",
			config: &Config{
				Name:         "invalid-type",
				Type:         "unsupported",
				ArtifactType: "maven",
			},
			shouldErr: true,
			errMsg:    "unsupported repository type",
		},
	}
	
	mockStorage := &MockStorage{}
	mockArtifactFactory := &MockArtifactFactory{}
	mockArtifact := &MockArtifact{}
	
	// Setup common mocks
	mockArtifactFactory.On("CreateArtifact", mock.AnythingOfType("artifact.ArtifactType")).Return(mockArtifact, nil).Maybe()
	
	factory := &DefaultFactory{}
	
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			repo, err := factory.CreateRepository(tt.config, mockStorage, mockArtifactFactory)
			
			if tt.shouldErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
				assert.Nil(t, repo)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, repo)
			}
		})
	}
}
