package server

import (
	"context"
	"fmt"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/database"
)

// RouteRegistry manages dynamic route registration based on repository types
type RouteRegistry struct {
	registrars map[artifact.ArtifactType]RouteRegistrar
	mu         sync.RWMutex
}

// NewRouteRegistry creates a new route registry with all supported artifact types
func NewRouteRegistry() *RouteRegistry {
	registry := &RouteRegistry{
		registrars: make(map[artifact.ArtifactType]RouteRegistrar),
	}

	// Register all supported artifact type route handlers
	registry.registerDefaults()
	return registry
}

// registerDefaults registers route handlers for all supported artifact types
func (r *RouteRegistry) registerDefaults() {
	r.registrars[artifact.ArtifactTypeMaven] = NewMavenRouteRegistrar()
	r.registrars[artifact.ArtifactTypeNPM] = NewNPMRouteRegistrar()
	r.registrars[artifact.ArtifactTypeDocker] = NewDockerRouteRegistrar()
	r.registrars[artifact.ArtifactTypeGolang] = NewGolangRouteRegistrar()
	r.registrars[artifact.ArtifactTypePyPI] = NewPyPIRouteRegistrar()
	r.registrars[artifact.ArtifactTypeHelm] = NewHelmRouteRegistrar()
	r.registrars[artifact.ArtifactTypeCargo] = NewCargoRouteRegistrar()
	r.registrars[artifact.ArtifactTypeNuGet] = NewNuGetRouteRegistrar()
	r.registrars[artifact.ArtifactTypeRubyGems] = NewRubyGemsRouteRegistrar()
	r.registrars[artifact.ArtifactTypeTerraform] = NewTerraformRouteRegistrar()
	r.registrars[artifact.ArtifactTypeAnsible] = NewAnsibleRouteRegistrar()
	r.registrars[artifact.ArtifactTypeBazel] = NewBazelRouteRegistrar()
	r.registrars[artifact.ArtifactTypeGeneric] = NewGenericRouteRegistrar()
}

// RegisterRouteHandler registers a custom route handler for an artifact type
func (r *RouteRegistry) RegisterRouteHandler(artifactType artifact.ArtifactType, registrar RouteRegistrar) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.registrars[artifactType] = registrar
}

// GetRouteRegistrar returns the route registrar for a given artifact type
func (r *RouteRegistry) GetRouteRegistrar(artifactType artifact.ArtifactType) (RouteRegistrar, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	registrar, exists := r.registrars[artifactType]
	if !exists {
		return nil, fmt.Errorf("no route registrar found for artifact type: %s", artifactType)
	}
	return registrar, nil
}

// RegisterRepositoryRoutes dynamically registers routes for a specific repository
func (r *RouteRegistry) RegisterRepositoryRoutes(router *gin.Engine, server *Server, repo *database.Repository) error {
	artifactType := artifact.ArtifactType(repo.ArtifactType)
	
	registrar, err := r.GetRouteRegistrar(artifactType)
	if err != nil {
		return fmt.Errorf("failed to get route registrar for repository %s: %w", repo.Name, err)
	}

	// Create repository-specific route group
	repoGroup := router.Group("/" + repo.Name)
	
	// Register artifact-specific routes for this repository
	registrar.RegisterRoutes(repoGroup, server)
	
	return nil
}

// RegisterAllRepositoryRoutes registers routes for all repositories in the database
func (r *RouteRegistry) RegisterAllRepositoryRoutes(router *gin.Engine, server *Server, db database.DatabaseInterface) error {
	repositories, err := db.ListRepositories(context.Background())
	if err != nil {
		return fmt.Errorf("failed to list repositories: %w", err)
	}

	for _, repo := range repositories {
		if err := r.RegisterRepositoryRoutes(router, server, repo); err != nil {
			return fmt.Errorf("failed to register routes for repository %s: %w", repo.Name, err)
		}
	}

	return nil
}

// GetSupportedArtifactTypes returns all supported artifact types
func (r *RouteRegistry) GetSupportedArtifactTypes() []artifact.ArtifactType {
	r.mu.RLock()
	defer r.mu.RUnlock()
	
	types := make([]artifact.ArtifactType, 0, len(r.registrars))
	for artifactType := range r.registrars {
		types = append(types, artifactType)
	}
	return types
}
