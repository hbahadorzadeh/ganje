package server

import (
	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/artifact"
)

// RouteRegistrar defines the interface for registering artifact-specific routes
type RouteRegistrar interface {
	RegisterRoutes(router *gin.RouterGroup, server *Server)
	GetArtifactType() artifact.ArtifactType
}

// BaseRouteRegistrar provides common functionality for all route registrars
type BaseRouteRegistrar struct {
	artifactType artifact.ArtifactType
}

func NewBaseRouteRegistrar(artifactType artifact.ArtifactType) *BaseRouteRegistrar {
	return &BaseRouteRegistrar{
		artifactType: artifactType,
	}
}

func (b *BaseRouteRegistrar) GetArtifactType() artifact.ArtifactType {
	return b.artifactType
}

// MavenRouteRegistrar handles Maven-specific routes
type MavenRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewMavenRouteRegistrar() RouteRegistrar {
	return &MavenRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeMaven),
	}
}

func (m *MavenRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
    // Generic handlers to support multi-segment group paths and metadata via handler
    router.GET("/*path", server.authMiddleware(), server.requireRead(), server.mavenHandler)
    router.PUT("/*path", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// NPMRouteRegistrar handles NPM-specific routes
type NPMRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewNPMRouteRegistrar() RouteRegistrar {
	return &NPMRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeNPM),
	}
}

func (n *NPMRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/:package", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/:package/-/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.PUT("/:package", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// DockerRouteRegistrar handles Docker-specific routes
type DockerRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewDockerRouteRegistrar() RouteRegistrar {
	return &DockerRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeDocker),
	}
}

func (d *DockerRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/v2/", server.authMiddleware(), server.requireRead(), server.dockerAPIVersion)
	router.GET("/v2/:name/tags/list", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/v2/:name/manifests/:reference", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.PUT("/v2/:name/manifests/:reference", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// GolangRouteRegistrar handles Go module-specific routes
type GolangRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewGolangRouteRegistrar() RouteRegistrar {
	return &GolangRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeGolang),
	}
}

func (g *GolangRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/:module/@v/list", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/:module/@v/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
}

// PyPIRouteRegistrar handles PyPI-specific routes
type PyPIRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewPyPIRouteRegistrar() RouteRegistrar {
	return &PyPIRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypePyPI),
	}
}

func (p *PyPIRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/simple/", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/simple/:package/", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/packages/:hash/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
}

// HelmRouteRegistrar handles Helm-specific routes
type HelmRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewHelmRouteRegistrar() RouteRegistrar {
	return &HelmRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeHelm),
	}
}

func (h *HelmRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/index.yaml", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
}

// GenericRouteRegistrar handles Generic artifact routes
type GenericRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewGenericRouteRegistrar() RouteRegistrar {
	return &GenericRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeGeneric),
	}
}

func (g *GenericRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/*path", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.PUT("/*path", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
	router.DELETE("/*path", server.authMiddleware(), server.requireWrite(), server.deleteArtifact)
}

// CargoRouteRegistrar handles Cargo-specific routes
type CargoRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewCargoRouteRegistrar() RouteRegistrar {
	return &CargoRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeCargo),
	}
}

func (c *CargoRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/api/v1/crates/:name", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/api/v1/crates/:name/:version/download", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.PUT("/api/v1/crates/new", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// NuGetRouteRegistrar handles NuGet-specific routes
type NuGetRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewNuGetRouteRegistrar() RouteRegistrar {
	return &NuGetRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeNuGet),
	}
}

func (n *NuGetRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/v3/index.json", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/v3-flatcontainer/:package/index.json", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/v3-flatcontainer/:package/:version/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.PUT("/api/v2/package", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// RubyGemsRouteRegistrar handles RubyGems-specific routes
type RubyGemsRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewRubyGemsRouteRegistrar() RouteRegistrar {
	return &RubyGemsRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeRubyGems),
	}
}

func (r *RubyGemsRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/specs.4.8.gz", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/quick/Marshal.4.8/:filename", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/gems/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.POST("/api/v1/gems", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// TerraformRouteRegistrar handles Terraform-specific routes
type TerraformRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewTerraformRouteRegistrar() RouteRegistrar {
	return &TerraformRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeTerraform),
	}
}

func (t *TerraformRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/v1/modules/:namespace/:name/versions", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/v1/modules/:namespace/:name/:version/download", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.POST("/v1/modules", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// AnsibleRouteRegistrar handles Ansible-specific routes
type AnsibleRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewAnsibleRouteRegistrar() RouteRegistrar {
	return &AnsibleRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeAnsible),
	}
}

func (a *AnsibleRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/api/v2/collections/:namespace/:name/", server.authMiddleware(), server.requireRead(), server.getIndex)
	router.GET("/download/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.POST("/api/v2/collections/", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// ConanRouteRegistrar handles Conan-specific routes
type ConanRouteRegistrar struct {
	*BaseRouteRegistrar
}

func NewConanRouteRegistrar() RouteRegistrar {
	return &ConanRouteRegistrar{
		BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeConan),
	}
}

func (c *ConanRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
	router.GET("/*path", server.authMiddleware(), server.requireRead(), server.pullArtifact)
	router.PUT("/*path", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
	router.DELETE("/*path", server.authMiddleware(), server.requireWrite(), server.deleteArtifact)
}
