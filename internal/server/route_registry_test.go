package server

import (
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/database"
)

func TestNewRouteRegistry(t *testing.T) {
    registry := NewRouteRegistry()
    
    assert.NotNil(t, registry)
    assert.NotNil(t, registry.registrars)
    
    // Check that all default registrars are registered
    supportedTypes := registry.GetSupportedArtifactTypes()
    expectedTypes := []artifact.ArtifactType{
        artifact.ArtifactTypeMaven,
        artifact.ArtifactTypeNPM,
        artifact.ArtifactTypeDocker,
        artifact.ArtifactTypeGolang,
        artifact.ArtifactTypePyPI,
        artifact.ArtifactTypeHelm,
        artifact.ArtifactTypeCargo,
        artifact.ArtifactTypeNuGet,
        artifact.ArtifactTypeRubyGems,
        artifact.ArtifactTypeTerraform,
        artifact.ArtifactTypeAnsible,
        artifact.ArtifactTypeBazel,
        artifact.ArtifactTypeGeneric,
    }
    
    assert.Len(t, supportedTypes, len(expectedTypes))
    for _, expectedType := range expectedTypes {
        assert.Contains(t, supportedTypes, expectedType)
    }
}

func TestTerraformRouteRegistrar(t *testing.T) {
    gin.SetMode(gin.TestMode)
    registrar := NewTerraformRouteRegistrar()

    assert.Equal(t, artifact.ArtifactTypeTerraform, registrar.GetArtifactType())

    router := gin.New()
    group := router.Group("/test-terraform")
    server, _, _, _ := createTestServer()

    registrar.RegisterRoutes(group, server)

    routes := router.Routes()
    expectedPaths := []string{
        "/test-terraform/v1/modules/:namespace/:name/:provider/versions",
        "/test-terraform/v1/modules/:namespace/:name/:provider/:version/download",
    }

    for _, expectedPath := range expectedPaths {
        found := false
        for _, route := range routes {
            if route.Path == expectedPath {
                found = true
                break
            }
        }
        assert.True(t, found, "Expected Terraform route %s should be registered", expectedPath)
    }
}

func TestGetRouteRegistrar(t *testing.T) {
	registry := NewRouteRegistry()
	
	// Test valid artifact type
	registrar, err := registry.GetRouteRegistrar(artifact.ArtifactTypeMaven)
	assert.NoError(t, err)
	assert.NotNil(t, registrar)
	assert.Equal(t, artifact.ArtifactTypeMaven, registrar.GetArtifactType())
	
	// Test invalid artifact type
	registrar, err = registry.GetRouteRegistrar(artifact.ArtifactType("invalid"))
	assert.Error(t, err)
	assert.Nil(t, registrar)
	assert.Contains(t, err.Error(), "no route registrar found")
}

func TestRegisterRouteHandler(t *testing.T) {
	registry := NewRouteRegistry()
	
	// Create a custom route registrar
	customRegistrar := &MavenRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactType("custom")),
	}
	
	// Register custom handler
	registry.RegisterRouteHandler(artifact.ArtifactType("custom"), customRegistrar)
	
	// Verify it was registered
	registrar, err := registry.GetRouteRegistrar(artifact.ArtifactType("custom"))
	assert.NoError(t, err)
	assert.Equal(t, customRegistrar, registrar)
}

func TestRegisterRepositoryRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := NewRouteRegistry()
	router := gin.New()
	
	// Create a mock server
	server := &Server{
		routeRegistry: registry,
	}
	
	// Create a test repository
	repo := &database.Repository{
		Name:         "test-maven-repo",
		ArtifactType: string(artifact.ArtifactTypeMaven),
	}
	
	// Register routes for the repository
	err := registry.RegisterRepositoryRoutes(router, server, repo)
	assert.NoError(t, err)
	
	// Verify routes were registered by checking the router
	routes := router.Routes()
	found := false
	for _, route := range routes {
		if route.Path == "/test-maven-repo/:groupId/:artifactId/:version/:filename" {
			found = true
			break
		}
	}
	assert.True(t, found, "Maven-specific route should be registered")
}

func TestRegisterRepositoryRoutesInvalidType(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registry := NewRouteRegistry()
	router := gin.New()
	
	server := &Server{
		routeRegistry: registry,
	}
	
	// Create a repository with invalid artifact type
	repo := &database.Repository{
		Name:         "test-invalid-repo",
		ArtifactType: "invalid-type",
	}
	
	// Try to register routes
	err := registry.RegisterRepositoryRoutes(router, server, repo)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get route registrar")
}

func TestMavenRouteRegistrar(t *testing.T) {
	registrar := NewMavenRouteRegistrar()
	
	assert.Equal(t, artifact.ArtifactTypeMaven, registrar.GetArtifactType())
	
	// Test route registration
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/test-maven")
	server, _, _, _ := createTestServer()
	
	// This should not panic
	assert.NotPanics(t, func() {
		registrar.RegisterRoutes(group, server)
	})
	
	// Verify Maven-specific routes were registered
	routes := router.Routes()
	expectedPaths := []string{
		"/test-maven/:groupId/:artifactId/:version/:filename",
		"/test-maven/:groupId/:artifactId/maven-metadata.xml",
	}
	
	for _, expectedPath := range expectedPaths {
		found := false
		for _, route := range routes {
			if route.Path == expectedPath {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected Maven route %s should be registered", expectedPath)
	}
}

func TestNPMRouteRegistrar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registrar := NewNPMRouteRegistrar()
	
	assert.Equal(t, artifact.ArtifactTypeNPM, registrar.GetArtifactType())
	
	router := gin.New()
	group := router.Group("/test-npm")
	server, _, _, _ := createTestServer()
	
	registrar.RegisterRoutes(group, server)
	
	// Verify NPM-specific routes were registered
	routes := router.Routes()
	expectedPaths := []string{
		"/test-npm/:package",
		"/test-npm/:package/-/:filename",
	}
	
	for _, expectedPath := range expectedPaths {
		found := false
		for _, route := range routes {
			if route.Path == expectedPath {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected NPM route %s should be registered", expectedPath)
	}
}

func TestDockerRouteRegistrar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registrar := NewDockerRouteRegistrar()
	
	assert.Equal(t, artifact.ArtifactTypeDocker, registrar.GetArtifactType())
	
	router := gin.New()
	group := router.Group("/test-docker")
	server, _, _, _ := createTestServer()
	
	registrar.RegisterRoutes(group, server)
	
	// Verify Docker-specific routes were registered
	routes := router.Routes()
	expectedPaths := []string{
		"/test-docker/v2/",
		"/test-docker/v2/:name/tags/list",
		"/test-docker/v2/:name/manifests/:reference",
	}
	
	for _, expectedPath := range expectedPaths {
		found := false
		for _, route := range routes {
			if route.Path == expectedPath {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected Docker route %s should be registered", expectedPath)
	}
}

func TestGenericRouteRegistrar(t *testing.T) {
	gin.SetMode(gin.TestMode)
	registrar := NewGenericRouteRegistrar()
	
	assert.Equal(t, artifact.ArtifactTypeGeneric, registrar.GetArtifactType())
	
	router := gin.New()
	group := router.Group("/test-generic")
	server, _, _, _ := createTestServer()
	
	registrar.RegisterRoutes(group, server)
	
	// Verify Generic routes were registered
	routes := router.Routes()
	expectedPaths := []string{
		"/test-generic/*path",
	}
	
	for _, expectedPath := range expectedPaths {
		found := false
		for _, route := range routes {
			if route.Path == expectedPath {
				found = true
				break
			}
		}
		assert.True(t, found, "Expected Generic route %s should be registered", expectedPath)
	}
}

func TestConcurrentAccess(t *testing.T) {
	registry := NewRouteRegistry()
	
	// Test concurrent read access
	done := make(chan bool, 10)
	
	for i := 0; i < 10; i++ {
		go func() {
			defer func() { done <- true }()
			
			// Multiple goroutines accessing supported types
			types := registry.GetSupportedArtifactTypes()
			assert.NotEmpty(t, types)
			
			// Multiple goroutines getting registrars
			registrar, err := registry.GetRouteRegistrar(artifact.ArtifactTypeMaven)
			assert.NoError(t, err)
			assert.NotNil(t, registrar)
		}()
	}
	
	// Wait for all goroutines to complete
	for i := 0; i < 10; i++ {
		<-done
	}
}

func TestAllArtifactTypeRegistrars(t *testing.T) {
	registry := NewRouteRegistry()
	
	// Test that all artifact types have working registrars
	allTypes := []artifact.ArtifactType{
		artifact.ArtifactTypeMaven,
		artifact.ArtifactTypeNPM,
		artifact.ArtifactTypeDocker,
		artifact.ArtifactTypeGolang,
		artifact.ArtifactTypePyPI,
		artifact.ArtifactTypeHelm,
		artifact.ArtifactTypeCargo,
		artifact.ArtifactTypeNuGet,
		artifact.ArtifactTypeRubyGems,
		artifact.ArtifactTypeTerraform,
		artifact.ArtifactTypeAnsible,
		artifact.ArtifactTypeBazel,
		artifact.ArtifactTypeGeneric,
	}
	
	gin.SetMode(gin.TestMode)
	
	for _, artifactType := range allTypes {
		t.Run(string(artifactType), func(t *testing.T) {
			// Create a properly initialized test server for each test
			server, _, _, _ := createTestServer()
			
			registrar, err := registry.GetRouteRegistrar(artifactType)
			assert.NoError(t, err, "Should be able to get registrar for %s", artifactType)
			assert.NotNil(t, registrar, "Registrar should not be nil for %s", artifactType)
			assert.Equal(t, artifactType, registrar.GetArtifactType(), "Registrar should return correct type")
			
			// Test that routes can be registered without panic
			router := gin.New()
			group := router.Group("/test")
			
			assert.NotPanics(t, func() {
				registrar.RegisterRoutes(group, server)
			}, "Should not panic when registering routes for %s", artifactType)
		})
	}
}
