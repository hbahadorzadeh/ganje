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
- **Bazel Remote Cache** - HTTP remote cache compatible with Bazel's /ac and /cas endpoints

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

           ┌─────────────────────────────── Messaging (RabbitMQ) ────────────────────────────────┐
           │                                                                                      │
           │   ┌───────────────────────┐         publish events        ┌───────────────────────┐  │
           │   │      HTTP Server      │ ─────────────────────────────► │  Exchange/Queue(s)   │  │
           │   └───────────────────────┘                                └───────────────────────┘  │
           │                                                                          │            │
           │                                              consume artifact events      ▼            │
           │                                                       ┌───────────────────────────┐   │
           │                                                       │  Webhook Dispatcher (CMD) │   │
           │                                                       │  async + retries + HMAC   │   │
           │                                                       └───────────────────────────┘   │
           └───────────────────────────────────────────────────────────────────────────────────────┘
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
  - name: "bazel-cache"
    type: "local"
    artifact_type: "bazel"

messaging:
  rabbitmq:
    enabled: true
    url: "amqp://guest:guest@localhost:5672/"
    exchange: "ganje.events"
    exchange_type: "topic"
    routing_key: "artifact.*"

webhook:
  enabled: true
  workers: 4
  max_retries: 5
  initial_backoff_ms: 500
  max_backoff_ms: 30000
  http_timeout_ms: 10000
```

### Running
```bash
./ganje
```

### Running the Webhook Dispatcher (standalone)

The webhook dispatcher runs as a separate process that consumes artifact events from RabbitMQ and delivers configured webhooks asynchronously with retries and optional HMAC signing.

```bash
# from repository root
go run ./cmd/webhook-dispatcher --config config.yaml

# optionally specify a durable queue name
go run ./cmd/webhook-dispatcher --config config.yaml --queue ganje-webhooks
```

Requirements:

- `messaging.rabbitmq.enabled: true` and a reachable broker.
- `webhook.enabled: true` and database configured.
- Webhooks can be managed via the admin API under `/api/v1/repositories/:name/webhooks`.

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

### Bazel Remote Cache
Once a repository is created with `artifact_type: "bazel"`, it exposes Bazel HTTP cache endpoints:

- `/{repo}/ac/*key` — Action Cache (AC)
- `/{repo}/cas/*digest` — Content Addressable Store (CAS)

Example Bazel flags:

```bash
# Read + write remote cache
bazel build //... \
  --remote_cache=http://localhost:8080/bazel-cache \
  --remote_timeout=60 \
  --experimental_remote_cache_compression \
  --remote_header=Authorization=Bearer\ <YOUR_TOKEN>

# Read-only (disable uploads)
bazel build //... \
  --remote_cache=http://localhost:8080/bazel-cache \
  --remote_upload_local_results=false \
  --remote_header=Authorization=Bearer\ <YOUR_TOKEN>
```

Notes:

- HEAD/GET on `/ac/*` and `/cas/*` read entries; PUT writes; DELETE removes entries.
- Ensure your token has `read` for pulls and `write` for uploads.

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
├── main.go             # HTTP API server entrypoint
└── webhook-dispatcher/ # Standalone webhook dispatcher service
internal/
├── artifact/          # Artifact interfaces and factory
│   └── types/         # Specific artifact implementations
├── auth/              # Authentication and authorization
├── config/            # Configuration management
├── database/          # Database models and operations
├── messaging/         # Publisher/consumer for RabbitMQ
├── repository/        # Repository implementations
├── server/            # HTTP server and handlers
└── webhook/           # Webhook dispatcher library (used by standalone service)
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
