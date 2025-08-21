# Ganje - Universal Artifact Repository

Ganje is a comprehensive web server artifact repository application that supports multiple artifact types with local, remote, and virtual repository configurations.

## Features

### Supported Artifact Types
- **Maven** - Java artifacts with standard Maven repository layout
- **NPM** - Node.js packages with npm registry compatibility
- **Docker** - Container images with Docker Registry API v2
- **PyPI** - Python packages with PyPI simple index
- **Helm** - Kubernetes Helm charts
- **Go Modules** - Go module proxy compatibility
- **Cargo** - Rust crates registry
- **NuGet** - .NET packages
- **RubyGems** - Ruby gems
- **Ansible Galaxy** - Ansible collections
- **Terraform** - Terraform modules
- **Generic** - Generic file storage
- **Canon** - Custom artifact format

### Repository Types
1. **Local** - Stores artifacts and indexes locally (push and pull)
2. **Remote** - Proxies remote repositories with caching (pull only)
3. **Virtual** - Aggregates multiple local and remote repositories (pull and push)

### Key Features
- **Hash-based Sharding** - Efficient artifact storage using SHA256 hash sharding
- **OAuth Authentication** - JWT token-based authentication with realm-based access control
- **Statistics Tracking** - Pull/push statistics for each artifact
- **Cache Management** - Invalidation endpoints for cached artifacts and indexes
- **Index Rebuilding** - Endpoints to rebuild repository indexes from scratch
- **RESTful API** - Standard endpoints for each artifact type
- **Database Storage** - GORM-based metadata storage with PostgreSQL/MySQL support

## Architecture

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Gin Server    │    │  Authentication │    │   Repository    │
│   (HTTP API)    │◄──►│   (OAuth/JWT)   │    │    Manager      │
└─────────────────┘    └─────────────────┘    └─────────────────┘
         │                                              │
         ▼                                              ▼
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Artifact      │    │    Storage      │    │    Database     │
│   Factory       │    │   (Local FS)    │    │ (PostgreSQL)    │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## Installation

### Prerequisites
- Go 1.21+
- PostgreSQL or MySQL database
- OAuth server for authentication

### Build from Source
```bash
git clone https://github.com/hbahadorzadeh/ganje.git
cd ganje
go mod tidy
go build -o ganje ./cmd
```

### Configuration
Create a `config.yaml` file:

```yaml
server:
  host: "0.0.0.0"
  port: 8080

database:
  driver: "postgres"
  host: "localhost"
  port: 5432
  username: "ganje"
  password: "password"
  database: "ganje"
  ssl_mode: "disable"

storage:
  type: "local"
  local_path: "/var/lib/ganje/storage"

auth:
  oauth_server: "https://auth.example.com"
  jwt_secret: "your-jwt-secret-key-here"
  realms:
    - name: "developers"
      permissions: ["read", "write"]
    - name: "admins"
      permissions: ["admin"]

repositories:
  - name: "maven-local"
    type: "local"
    artifact_type: "maven"
  - name: "npm-registry"
    type: "remote"
    artifact_type: "npm"
    url: "https://registry.npmjs.org"
```

### Running
```bash
./ganje
```

## API Endpoints

### Repository Management
- `GET /api/v1/repositories` - List all repositories
- `GET /api/v1/repositories/{name}` - Get repository details
- `POST /api/v1/repositories` - Create repository (admin)
- `DELETE /api/v1/repositories/{name}` - Delete repository (admin)

### Cache Management
- `DELETE /api/v1/repositories/{name}/cache` - Invalidate cache (admin)
- `POST /api/v1/repositories/{name}/reindex` - Rebuild index (admin)

### Artifact Operations
Each repository supports standard endpoints for its artifact type:

#### Maven
- `GET /{repo}/{groupId}/{artifactId}/{version}/{filename}` - Download artifact
- `PUT /{repo}/{groupId}/{artifactId}/{version}/{filename}` - Upload artifact
- `GET /{repo}/{groupId}/{artifactId}/maven-metadata.xml` - Get metadata

#### NPM
- `GET /{repo}/{package}` - Get package metadata
- `GET /{repo}/{package}/-/{filename}` - Download package
- `PUT /{repo}/{package}` - Publish package

#### Docker
- `GET /{repo}/v2/` - API version check
- `GET /{repo}/v2/{name}/tags/list` - List tags
- `GET /{repo}/v2/{name}/manifests/{reference}` - Get manifest
- `PUT /{repo}/v2/{name}/manifests/{reference}` - Push manifest

### Authentication
All endpoints require JWT authentication via `Authorization: Bearer <token>` header.

## Storage Layout

Artifacts are stored using hash-based sharding:
```
storage/
├── ab/
│   └── cd/
│       └── abcd1234...hash/
│           └── artifact-file
└── cache/
    └── remote-repo/
        └── cached-artifacts
```

## Development

### Project Structure
```
cmd/                    # Application entry point
internal/
├── artifact/          # Artifact interfaces and factory
│   └── types/         # Specific artifact implementations
├── auth/              # Authentication and authorization
├── config/            # Configuration management
├── database/          # Database models and operations
├── repository/        # Repository implementations
├── server/            # HTTP server and handlers
└── storage/           # Storage interfaces and implementations
```

### Adding New Artifact Types
1. Implement the `Artifact` interface in `internal/artifact/types/`
2. Add the new type to the factory in `internal/artifact/factory.go`
3. Update the configuration schema if needed

## License

MIT License - see LICENSE file for details.

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests
5. Submit a pull request

## Support

For issues and questions, please use the GitHub issue tracker.
