# Repository Management API

The Ganje artifact repository provides comprehensive REST API endpoints for managing repositories. All endpoints require authentication and appropriate permissions.

## Base URL
```
http://localhost:8080/api/v1
```

## Authentication
All endpoints require a valid JWT token in the Authorization header:
```
Authorization: Bearer <your-jwt-token>
```

## Repository CRUD Operations

### 1. List All Repositories
**GET** `/repositories`

Returns a list of all repositories.

**Response:**
```json
{
  "repositories": [
    {
      "name": "maven-central",
      "type": "local",
      "artifact_type": "maven"
    }
  ]
}
```

### 2. Get Repository Details
**GET** `/repositories/{name}`

Returns detailed information about a specific repository.

**Response:**
```json
{
  "name": "maven-central",
  "type": "local",
  "artifact_type": "maven",
  "url": "",
  "created_at": "2024-01-01T00:00:00Z",
  "updated_at": "2024-01-01T00:00:00Z",
  "artifact_count": 42,
  "statistics": {
    "total_artifacts": 42,
    "total_size": 1048576,
    "pull_count": 150,
    "push_count": 42
  }
}
```

### 3. Create Repository
**POST** `/repositories`

Creates a new repository with dynamic route registration.

**Request Body:**
```json
{
  "name": "my-maven-repo",
  "type": "local",
  "artifact_type": "maven",
  "url": "",
  "upstream": [],
  "options": {}
}
```

**Response:**
```json
{
  "message": "Repository created successfully"
}
```

### 4. Update Repository
**PUT** `/repositories/{name}`

Updates an existing repository configuration.

**Request Body:**
```json
{
  "type": "remote",
  "artifact_type": "maven",
  "url": "https://repo1.maven.org/maven2/",
  "description": "Maven Central Mirror"
}
```

**Response:**
```json
{
  "message": "Repository updated successfully"
}
```

### 5. Delete Repository
**DELETE** `/repositories/{name}[?force=true]`

Deletes a repository. Use `?force=true` to delete repositories containing artifacts.

**Response:**
```json
{
  "message": "Repository deleted successfully"
}
```

## Advanced Repository Operations

### 6. Validate Repository Configuration
**POST** `/repositories/validate`

Validates repository configuration without creating it.

**Request Body:**
```json
{
  "name": "test-repo",
  "type": "local",
  "artifact_type": "maven"
}
```

**Response:**
```json
{
  "valid": true,
  "message": "Repository configuration is valid"
}
```

### 7. Bulk Delete Repositories
**DELETE** `/repositories`

Deletes multiple repositories in a single operation.

**Request Body:**
```json
{
  "names": ["repo1", "repo2", "repo3"],
  "force": false
}
```

**Response:**
```json
{
  "results": {
    "repo1": {
      "success": true,
      "message": "Repository deleted successfully"
    },
    "repo2": {
      "success": false,
      "error": "Repository contains artifacts. Use force=true to delete anyway",
      "artifact_count": 5
    }
  },
  "summary": {
    "total": 3,
    "success": 1,
    "errors": 2
  }
}
```

### 8. Get Supported Types
**GET** `/repository-types`

Returns supported repository and artifact types.

**Response:**
```json
{
  "repository_types": ["local", "remote", "virtual"],
  "artifact_types": [
    "maven", "npm", "docker", "pypi", "helm", "golang",
    "cargo", "nuget", "rubygems", "terraform", "ansible",
    "canon", "generic"
  ]
}
```

## Cache Management

### 9. Invalidate Cache
**DELETE** `/repositories/{name}/cache[?path=specific/path]`

Invalidates repository cache. Optionally specify a path to invalidate specific cached items.

**Response:**
```json
{
  "message": "Cache invalidated successfully"
}
```

### 10. Rebuild Index
**POST** `/repositories/{name}/reindex`

Rebuilds the repository index.

**Response:**
```json
{
  "message": "Index rebuilt successfully"
}
```

### 11. Get Repository Statistics
**GET** `/repositories/{name}/stats`

Returns detailed repository statistics.

**Response:**
```json
{
  "total_artifacts": 42,
  "total_size": 1048576,
  "pull_count": 150,
  "push_count": 42,
  "last_activity": "2024-01-01T12:00:00Z"
}
```

## Repository Types

### Local Repository
Stores artifacts locally on the server filesystem.
```json
{
  "name": "local-maven",
  "type": "local",
  "artifact_type": "maven"
}
```

### Remote Repository
Proxies artifacts from an upstream repository.
```json
{
  "name": "maven-central-proxy",
  "type": "remote",
  "artifact_type": "maven",
  "url": "https://repo1.maven.org/maven2/"
}
```

### Virtual Repository
Aggregates multiple repositories into a single endpoint.
```json
{
  "name": "maven-virtual",
  "type": "virtual",
  "artifact_type": "maven",
  "upstream": ["local-maven", "maven-central-proxy"]
}
```

## Supported Artifact Types

Each artifact type gets its own specific routes when a repository is created:

- **Maven**: `/:groupId/:artifactId/:version/:filename`
- **NPM**: `/:package`, `/:package/-/:filename`
- **Docker**: `/v2/`, `/v2/:name/manifests/:reference`
- **PyPI**: `/simple/`, `/simple/:package/`
- **Helm**: `/index.yaml`, `/:chart-:version.tgz`
- **Go Modules**: `/:module/@v/list`, `/:module/@v/:version.zip`
- **Cargo**: `/api/v1/crates/:name`
- **NuGet**: `/v3/index.json`, `/v3-flatcontainer/:package/index.json`
- **RubyGems**: `/specs.4.8.gz`, `/gems/:filename`
- **Terraform**: `/v1/modules/:namespace/:name/versions`
- **Ansible**: `/api/v2/collections/:namespace/:name/`
- **Generic**: `/*path` (supports any file structure)

## Error Responses

All endpoints return appropriate HTTP status codes and error messages:

```json
{
  "error": "Repository not found"
}
```

Common status codes:
- `200 OK` - Success
- `201 Created` - Repository created
- `400 Bad Request` - Invalid request data
- `401 Unauthorized` - Authentication required
- `403 Forbidden` - Insufficient permissions
- `404 Not Found` - Repository not found
- `409 Conflict` - Repository already exists
- `500 Internal Server Error` - Server error

## Dynamic Routing

When repositories are created or updated, routes are automatically registered based on the artifact type:

1. **Repository Creation**: Routes are registered immediately after successful creation
2. **Artifact Type Change**: Routes are re-registered when artifact type is updated
3. **Server Startup**: All existing repositories get their routes registered

This ensures that each repository only exposes endpoints relevant to its artifact type, providing a clean and type-safe API surface.
