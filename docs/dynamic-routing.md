# Dynamic Routing System

The Ganje artifact repository now supports dynamic routing based on repository types. Instead of hardcoded routes for all artifact types, routes are registered dynamically based on the specific artifact type of each repository.

## How It Works

### 1. Route Registration System

Each artifact type has its own route registrar that defines the specific HTTP endpoints for that artifact type:

```go
// Example: Maven routes
type MavenRouteRegistrar struct {
    *BaseRouteRegistrar
}

func (m *MavenRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
    router.GET("/:groupId/:artifactId/:version/:filename", server.authMiddleware(), server.requireRead(), server.pullArtifact)
    router.PUT("/:groupId/:artifactId/:version/:filename", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
    router.GET("/:groupId/:artifactId/maven-metadata.xml", server.authMiddleware(), server.requireRead(), server.getIndex)
}
```

### 2. Repository-Specific Routes

When a repository is created, routes are automatically registered based on its artifact type:

- **Maven repository** → Gets Maven-specific routes (`/:groupId/:artifactId/:version/:filename`)
- **NPM repository** → Gets NPM-specific routes (`/:package`, `/:package/-/:filename`)
- **Docker repository** → Gets Docker-specific routes (`/v2/`, `/v2/:name/manifests/:reference`)
- **Generic repository** → Gets generic routes (`/*path`)

### 3. Automatic Route Registration

Routes are registered automatically in two scenarios:

1. **Server startup**: All existing repositories get their routes registered
2. **Repository creation**: New repositories get routes registered immediately

## Supported Artifact Types

The system supports the following artifact types with their specific route patterns:

| Artifact Type | Example Routes |
|---------------|----------------|
| **Maven** | `GET /:groupId/:artifactId/:version/:filename`<br>`PUT /:groupId/:artifactId/:version/:filename`<br>`GET /:groupId/:artifactId/maven-metadata.xml` |
| **NPM** | `GET /:package`<br>`GET /:package/-/:filename`<br>`PUT /:package` |
| **Docker** | `GET /v2/`<br>`GET /v2/:name/tags/list`<br>`GET /v2/:name/manifests/:reference`<br>`PUT /v2/:name/manifests/:reference` |
| **PyPI** | `GET /simple/`<br>`GET /simple/:package/`<br>`GET /packages/:hash/:filename` |
| **Helm** | `GET /index.yaml`<br>`GET /:chart-:version.tgz` |
| **Go Modules** | `GET /:module/@v/list`<br>`GET /:module/@v/:version.zip`<br>`GET /:module/@v/:version.info`<br>`GET /:module/@v/:version.mod` |
| **Cargo** | `GET /api/v1/crates/:name`<br>`GET /api/v1/crates/:name/:version/download`<br>`PUT /api/v1/crates/new` |
| **NuGet** | `GET /v3/index.json`<br>`GET /v3-flatcontainer/:package/index.json`<br>`GET /v3-flatcontainer/:package/:version/:package.:version.nupkg` |
| **RubyGems** | `GET /specs.4.8.gz`<br>`GET /quick/Marshal.4.8/:name-:version.gemspec.rz`<br>`GET /gems/:filename` |
| **Terraform** | `GET /v1/modules/:namespace/:name/versions`<br>`GET /v1/modules/:namespace/:name/:version/download` |
| **Ansible** | `GET /api/v2/collections/:namespace/:name/`<br>`GET /download/:namespace-:name-:version.tar.gz` |
| **Generic** | `GET /*path`<br>`PUT /*path`<br>`DELETE /*path` |

## Example Usage

### Creating Repositories with Different Types

```bash
# Create a Maven repository
curl -X POST http://localhost:8080/api/v1/repositories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "maven-central",
    "type": "local",
    "artifact_type": "maven"
  }'

# Create an NPM repository
curl -X POST http://localhost:8080/api/v1/repositories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "npm-registry",
    "type": "local", 
    "artifact_type": "npm"
  }'

# Create a Docker repository
curl -X POST http://localhost:8080/api/v1/repositories \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer <token>" \
  -d '{
    "name": "docker-registry",
    "type": "local",
    "artifact_type": "docker"
  }'
```

### Accessing Repository-Specific Endpoints

After creating repositories, each will have its own artifact-type-specific endpoints:

```bash
# Maven repository endpoints
GET  /maven-central/com/example/myapp/1.0.0/myapp-1.0.0.jar
PUT  /maven-central/com/example/myapp/1.0.0/myapp-1.0.0.jar
GET  /maven-central/com/example/myapp/maven-metadata.xml

# NPM repository endpoints  
GET  /npm-registry/express
GET  /npm-registry/express/-/express-4.18.2.tgz
PUT  /npm-registry/express

# Docker repository endpoints
GET  /docker-registry/v2/
GET  /docker-registry/v2/myapp/tags/list
GET  /docker-registry/v2/myapp/manifests/latest
PUT  /docker-registry/v2/myapp/manifests/latest
```

## Benefits

1. **Type Safety**: Each repository only exposes endpoints relevant to its artifact type
2. **Clean URLs**: No generic catch-all routes that might conflict
3. **Extensibility**: Easy to add new artifact types with their specific routing patterns
4. **Maintainability**: Route logic is separated by artifact type, making it easier to maintain
5. **Performance**: Only relevant routes are registered for each repository

## Adding New Artifact Types

To add support for a new artifact type:

1. **Define the artifact type** in `internal/artifact/artifact.go`
2. **Implement the artifact interface** in `internal/artifact/types/`
3. **Create a route registrar** in `internal/server/routes.go`
4. **Register the route handler** in `internal/server/route_registry.go`

Example for a new "MyArtifact" type:

```go
// 1. Add to artifact types
const ArtifactTypeMyArtifact ArtifactType = "myartifact"

// 2. Create route registrar
type MyArtifactRouteRegistrar struct {
    *BaseRouteRegistrar
}

func NewMyArtifactRouteRegistrar() RouteRegistrar {
    return &MyArtifactRouteRegistrar{
        BaseRouteRegistrar: NewBaseRouteRegistrar(artifact.ArtifactTypeMyArtifact),
    }
}

func (m *MyArtifactRouteRegistrar) RegisterRoutes(router *gin.RouterGroup, server *Server) {
    router.GET("/my-endpoint/:id", server.authMiddleware(), server.requireRead(), server.pullArtifact)
    router.PUT("/my-endpoint/:id", server.authMiddleware(), server.requireWrite(), server.pushArtifact)
}

// 3. Register in route registry
r.registrars[artifact.ArtifactTypeMyArtifact] = NewMyArtifactRouteRegistrar()
```

The dynamic routing system will automatically handle the new artifact type when repositories are created with `"artifact_type": "myartifact"`.
