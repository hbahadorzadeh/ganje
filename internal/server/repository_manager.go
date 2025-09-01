package server

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/metrics"
	"github.com/hbahadorzadeh/ganje/internal/repository"
	"github.com/hbahadorzadeh/ganje/internal/storage"
)

// RepositoryManager manages repositories
type RepositoryManager struct {
	repositories map[string]repository.Repository
	config       *config.Config
	db           *database.DB
	storage      storage.Storage
	factory      artifact.Factory
	metricsService *metrics.MetricsService
	mutex        sync.RWMutex
}

// NewRepositoryManager creates a new repository manager
func NewRepositoryManager(cfg *config.Config, db *database.DB, metricsService *metrics.MetricsService) repository.Manager {
	// Initialize storage
	localStorage := storage.NewLocalStorage(cfg.Storage.LocalPath)
	
	// Initialize artifact factory
	artifactFactory := artifact.NewFactory()

	manager := &RepositoryManager{
		repositories: make(map[string]repository.Repository),
		config:       cfg,
		db:           db,
		storage:      localStorage,
		factory:      artifactFactory,
		metricsService: metricsService,
	}

	// Load repositories from configuration
	for _, repoConfig := range cfg.Repositories {
		repo, err := manager.CreateRepository(&repository.Config{
			Name:         repoConfig.Name,
			Type:         repoConfig.Type,
			ArtifactType: repoConfig.ArtifactType,
			URL:          repoConfig.URL,
			Upstream:     repoConfig.Upstream,
			Options:      repoConfig.Options,
		})
		if err != nil {
			fmt.Printf("Failed to create repository %s: %v\n", repoConfig.Name, err)
			continue
		}
		manager.repositories[repoConfig.Name] = repo
	}

	return manager
}

// GetRepository returns a repository by name
func (rm *RepositoryManager) GetRepository(name string) (repository.Repository, error) {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	repo, exists := rm.repositories[name]
	if !exists {
		return nil, fmt.Errorf("repository not found: %s", name)
	}

	return repo, nil
}

// ListRepositories returns all repositories
func (rm *RepositoryManager) ListRepositories() []repository.Repository {
	rm.mutex.RLock()
	defer rm.mutex.RUnlock()

	var repos []repository.Repository
	for _, repo := range rm.repositories {
		repos = append(repos, repo)
	}

	return repos
}

// CreateRepository creates a new repository
func (rm *RepositoryManager) CreateRepository(config *repository.Config) (repository.Repository, error) {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	// Check if repository already exists
	if _, exists := rm.repositories[config.Name]; exists {
		return nil, fmt.Errorf("repository already exists: %s", config.Name)
	}

	// Parse artifact type
	artifactType := artifact.ArtifactType(config.ArtifactType)

	var repo repository.Repository
	var _ error

	switch repository.Type(config.Type) {
	case repository.Local:
		repo = repository.NewLocalRepository(
			config.Name,
			artifactType,
			rm.storage,
			rm.factory,
			rm.db,
		)

	case repository.Remote:
		if config.URL == "" {
			return nil, fmt.Errorf("URL required for remote repository")
		}
		repo = repository.NewRemoteRepository(
			config.Name,
			artifactType,
			config.URL,
			rm.storage,
			rm.factory,
			rm.db,
		)

	case repository.Virtual:
		// Get upstream repositories
		var upstreams []repository.Repository
		for _, upstreamName := range config.Upstream {
			upstream, exists := rm.repositories[upstreamName]
			if !exists {
				return nil, fmt.Errorf("upstream repository not found: %s", upstreamName)
			}
			upstreams = append(upstreams, upstream)
		}

		repo = repository.NewVirtualRepository(
			config.Name,
			artifactType,
			upstreams,
			rm.storage,
			rm.factory,
			rm.db,
		)

	default:
		return nil, fmt.Errorf("unsupported repository type: %s", config.Type)
	}

	// Save repository to database
	var cfgJSON string
	if config.Options != nil {
		if b, err := json.Marshal(config.Options); err == nil {
			cfgJSON = string(b)
		}
	}

	dbRepo := &database.Repository{
		Name:         config.Name,
		Type:         config.Type,
		ArtifactType: config.ArtifactType,
		URL:          config.URL,
		Config:       cfgJSON,
	}

	if err := rm.db.SaveRepository(context.TODO(), dbRepo); err != nil {
		return nil, fmt.Errorf("failed to save repository to database: %w", err)
	}

	// Wrap repository with metrics if metrics service is available
	if rm.metricsService != nil {
		repo = repository.NewMetricsWrapper(repo, rm.metricsService)
	}

	rm.repositories[config.Name] = repo
	return repo, nil
}

// DeleteRepository deletes a repository
func (rm *RepositoryManager) DeleteRepository(name string) error {
	rm.mutex.Lock()
	defer rm.mutex.Unlock()

	if _, exists := rm.repositories[name]; !exists {
		return fmt.Errorf("repository not found: %s", name)
	}

	// Delete from database
	if err := rm.db.DeleteRepository(context.TODO(), name); err != nil {
		return fmt.Errorf("failed to delete repository from database: %w", err)
	}

	delete(rm.repositories, name)
	return nil
}
