# Ganje Artifact# Integration Tests

This directory contains integration tests for the Ganje artifact repository server using two approaches:

## Test Architecture

### 1. Container-Based Integration Tests (`container_integration_test.go`)
**Recommended approach** - Uses real artifact clients in Docker containers to test actual integration scenarios:

- **Maven**: Uses Maven container with configured `settings.xml` to deploy and resolve artifacts
- **NPM**: Uses Node.js container with npm configured to use Ganje as registry
- **Go Modules**: Uses Go container with `GOPROXY` pointing to Ganje server
- **Cargo**: Uses Rust container with Cargo configured to use Ganje registry
- **Docker**: Uses Docker-in-Docker to push/pull images from Ganje registry
- **Helm**: Uses Helm container to package and deploy charts to Ganje repository

### 2. HTTP API Tests (`artifact_integration_test.go`)
Simplified tests for Generic artifact type using direct HTTP API calls.

## Running Tests

### Container Integration Tests (Recommended)
```bash
# Run container-based integration tests
go test -v -run TestContainerIntegration ./test/integration

# Or run specific artifact type
go test -v -run TestContainerIntegration/TestMavenIntegration ./test/integration
```

### HTTP API Tests
```bash
# Run Generic artifact type test
go test -v -run TestGenericArtifactType ./test/integration
```

## Prerequisites

### For Container Tests
- Docker and Docker Compose installed
- Sufficient system resources for multiple containers
- Network access for container communication

### For HTTP Tests
- Ganje server running locally or accessible

## Test Environment Setup

The container tests use `docker-compose.test.yml` which includes:

1. **Ganje Server**: The artifact repository server under test
2. **Client Containers**: Various tool containers (Maven, npm, Go, Cargo, Docker, Helm)
3. **Configuration Files**: 
   - `maven-settings.xml`: Maven configuration pointing to Ganje
   - `cargo-config.toml`: Cargo registry configuration
   - Test project files for each artifact type

## Configuration Files

### Maven Settings (`maven-settings.xml`)
Configures Maven to use Ganje server as primary repository and mirror for Central.

### Cargo Config (`cargo-config.toml`)
Sets up Cargo to use Ganje as the default registry instead of crates.io.

## Test Flow

Each container integration test follows this pattern:

1. **Setup**: Start Docker Compose environment with Ganje server and client containers
2. **Create**: Generate test projects/artifacts in client containers
3. **Publish**: Use native tools to publish artifacts to Ganje server
4. **Consume**: Use native tools to download/install artifacts from Ganje server
5. **Verify**: Ensure artifacts were correctly stored and retrieved
6. **Cleanup**: Stop and remove containers

## Benefits of Container Approach

- **Real Integration**: Tests actual client-server interactions
- **Tool Compatibility**: Verifies compatibility with real artifact management tools
- **Network Testing**: Tests HTTP/HTTPS communication patterns
- **Authentication**: Can test authentication flows with real clients
- **Protocol Compliance**: Ensures Ganje implements artifact protocols correctly

## Artifact Types Tested

### Container Integration Tests
- **Maven**: JAR/POM deployment and dependency resolution
- **NPM**: Package publishing and installation
- **Go Modules**: Module publishing and `go get` operations
- **Cargo**: Crate publishing and dependency resolution
- **Docker**: Image push/pull operations
- **Helm**: Chart packaging and repository operations

### HTTP API Tests
- **Generic**: Direct file upload/download via HTTP API

## Troubleshooting

### Container Issues
```bash
# Check container logs
docker-compose -f docker-compose.test.yml -p ganje-integration-test logs

# Clean up stuck containers
docker-compose -f docker-compose.test.yml -p ganje-integration-test down -v --remove-orphans
```

### Network Issues
Ensure containers can communicate with Ganje server on the test network.

### Resource Issues
Container tests require significant resources. Ensure Docker has adequate memory and CPU allocation.

## Supported Artifact Types

The tests cover all 12 supported artifact types:

- **Maven** - Java JAR files with proper manifest and class files
- **NPM** - Node.js packages with package.json and source files
- **PyPI** - Python wheels with metadata and source code
- **Docker** - Container manifests with proper schema
- **Helm** - Kubernetes charts with Chart.yaml and templates
- **Golang** - Go modules with go.mod and source files
- **Cargo** - Rust crates with Cargo.toml and library code
- **NuGet** - .NET packages with nuspec and assemblies
- **RubyGems** - Ruby gems with gemspec and library files
- **Ansible** - Collections with galaxy.yml and modules
- **Terraform** - Providers with binaries and manifests
- **Generic** - Generic file storage for any content type

## Test Structure

### Core Files

- `artifact_integration_test.go` - Main test suite with lifecycle tests for each artifact type
- `mock_artifacts.go` - Mock artifact generators that create realistic test content
- `test_helpers.go` - Utility functions for common test operations
- `advanced_tests.go` - Advanced scenarios like versioning, concurrency, and large files

### Test Categories

1. **Basic Lifecycle Tests** (`TestAllArtifactTypes`)
   - Push artifact to repository
   - Pull artifact and verify content
   - Validate metadata storage
   - List repository artifacts

2. **Versioning Tests** (`TestArtifactVersioning`)
   - Multiple versions of same artifact
   - Version-specific retrieval
   - Version listing and management

3. **Overwrite Tests** (`TestArtifactOverwrite`)
   - Artifact replacement scenarios
   - Content integrity after overwrite

4. **Deletion Tests** (`TestArtifactDeletion`)
   - Artifact removal
   - Cleanup verification

5. **Search Tests** (`TestArtifactSearch`)
   - Artifact discovery
   - Query-based filtering

6. **Concurrency Tests** (`TestConcurrentAccess`)
   - Parallel push operations
   - Race condition handling

7. **Large File Tests** (`TestLargeArtifact`)
   - Multi-megabyte artifact handling
   - Memory efficiency validation

8. **Repository Management** (`TestRepositoryManagement`)
   - Repository-level operations
   - Bulk artifact management

## Running the Tests

### Prerequisites

1. Go 1.19 or later
2. SQLite (for test database)
3. Required Go modules (automatically downloaded)

### Execute Tests

```bash
# Run all integration tests
cd test/integration
go test -v

# Run specific test suite
go test -v -run TestAllArtifactTypes

# Run tests for specific artifact type
go test -v -run "TestAllArtifactTypes/Maven"

# Run with race detection
go test -v -race

# Run with coverage
go test -v -cover -coverprofile=coverage.out
```

### Test Configuration

The tests use in-memory SQLite databases and temporary directories for isolation. Each test creates:

- Fresh test server instance
- Isolated database
- Temporary storage directory
- Clean repository state

## Mock Artifact Details

Each mock artifact is designed to be realistic and representative:

### Maven JAR
- Valid ZIP structure with META-INF/MANIFEST.MF
- Java class files with proper magic numbers
- Maven coordinate structure (groupId/artifactId/version)

### NPM Package
- Gzipped tarball with package/ prefix
- Valid package.json with dependencies
- JavaScript source files

### Python Wheel
- ZIP format with .dist-info directory
- METADATA file with package information
- Python source modules

### Docker Manifest
- JSON manifest following Docker Registry API v2
- Schema version and media types
- Layer and config references

### Helm Chart
- Gzipped tarball structure
- Chart.yaml with API version
- Kubernetes templates and values

### Go Module
- ZIP format with version prefix
- go.mod with module declaration
- Go source files

### Rust Crate
- Gzipped tarball with Cargo.toml
- src/lib.rs with documentation
- Proper crate metadata

### .NET NuGet
- ZIP format with .nuspec file
- Assembly DLL files
- Package metadata

### Ruby Gem
- Gzipped tarball with gemspec
- lib/ directory structure
- Ruby source files

### Ansible Collection
- Gzipped tarball with galaxy.yml
- plugins/modules/ structure
- Ansible module code

### Terraform Provider
- ZIP with provider binary
- Manifest JSON with protocol versions
- Platform-specific structure

### Generic Files
- Raw content storage
- Flexible metadata support
- Any file type support

## Test Server Architecture

The `TestServer` provides:

- **HTTP Server** - Handles artifact upload/download requests
- **Database** - In-memory SQLite for metadata storage
- **Storage** - Temporary filesystem for artifact content
- **Repository Manager** - Handles repository operations
- **Authentication** - Mock auth for testing (optional)

## Extending Tests

To add new test scenarios:

1. **New Artifact Type**
   - Add mock generator in `mock_artifacts.go`
   - Add test case in `TestAllArtifactTypes`
   - Define push/pull endpoints

2. **New Test Scenario**
   - Create test function in `advanced_tests.go`
   - Use `TestHelper` for common operations
   - Follow existing patterns for setup/cleanup

3. **Custom Mock Content**
   - Implement realistic file structures
   - Include proper metadata
   - Use appropriate compression/formats

## Troubleshooting

### Common Issues

1. **Port Conflicts** - Tests use random ports, but conflicts can occur
2. **Temporary Directory** - Cleanup happens automatically, but manual cleanup may be needed
3. **Database Locks** - In-memory databases prevent most locking issues
4. **Large Files** - Memory usage can be high for large artifact tests

### Debug Mode

Enable verbose logging:
```bash
go test -v -args -debug
```

### Test Isolation

Each test runs in complete isolation:
- Separate server instance
- Fresh database
- Clean temporary directory
- No shared state between tests

## Performance Considerations

- Tests create realistic but small artifacts to minimize execution time
- Large file tests use configurable sizes
- Concurrent tests are limited to prevent resource exhaustion
- Database operations use transactions for consistency

## Integration with CI/CD

The tests are designed for automated execution:
- No external dependencies
- Deterministic behavior
- Proper cleanup
- Clear pass/fail criteria
- Detailed error reporting
