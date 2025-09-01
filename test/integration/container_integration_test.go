package integration

import (
	"context"
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

// ContainerIntegrationTestSuite runs integration tests using real artifact clients in containers
type ContainerIntegrationTestSuite struct {
	suite.Suite
	composeFile string
	projectName string
}

func TestContainerIntegration(t *testing.T) {
	suite.Run(t, new(ContainerIntegrationTestSuite))
}

func (s *ContainerIntegrationTestSuite) SetupSuite() {
	s.composeFile = "docker-compose.test.yml"
	s.projectName = "ganje-integration-test"

	// Ensure Docker Compose is available
	_, err := exec.LookPath("docker")
	if err != nil {
		s.T().Skip("Docker not available, skipping container integration tests")
	}

	// Create test directories
	s.createTestDirectories()

	// Start the test environment
	s.startTestEnvironment()
}

func (s *ContainerIntegrationTestSuite) TearDownSuite() {
	// Stop and clean up the test environment
	s.stopTestEnvironment()
}

func (s *ContainerIntegrationTestSuite) createTestDirectories() {
	dirs := []string{
		"maven-test",
		"npm-test",
		"go-test",
		"cargo-test",
		"docker-test",
		"helm-test",
	}

	for _, dir := range dirs {
		err := os.MkdirAll(dir, 0755)
		require.NoError(s.T(), err, "Failed to create test directory: %s", dir)
	}
}

func (s *ContainerIntegrationTestSuite) startTestEnvironment() {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Build and start services
	cmd := exec.CommandContext(ctx, "docker",
		"compose",
		"-f", s.composeFile,
		"-p", s.projectName,
		"up", "-d", "--build")
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	require.NoError(s.T(), err, "Failed to start test environment: %s", string(output))

	// Wait for services to be healthy
	s.waitForServices()
}

func (s *ContainerIntegrationTestSuite) stopTestEnvironment() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	cmd := exec.CommandContext(ctx, "docker",
		"compose",
		"-f", s.composeFile,
		"-p", s.projectName,
		"down", "-v", "--remove-orphans")
	cmd.Dir = "."

	output, err := cmd.CombinedOutput()
	if err != nil {
		s.T().Logf("Warning: Failed to clean up test environment: %s", string(output))
	}
}

func (s *ContainerIntegrationTestSuite) waitForServices() {
	// Wait for ganje-server to be healthy
	maxRetries := 30
	for i := 0; i < maxRetries; i++ {
		cmd := exec.Command("docker",
			"compose",
			"-f", s.composeFile,
			"-p", s.projectName,
			"ps", "--services", "--filter", "status=running")
		cmd.Dir = "."

		output, err := cmd.CombinedOutput()
		if err == nil && len(output) > 0 {
			s.T().Log("Services are running")
			time.Sleep(5 * time.Second) // Additional wait for full startup
			return
		}

		time.Sleep(2 * time.Second)
	}

	s.T().Fatal("Services failed to start within timeout")
}

func (s *ContainerIntegrationTestSuite) execInContainer(service string, command ...string) ([]byte, error) {
	cmd := exec.Command("docker",
		"compose",
		"-f", s.composeFile,
		"-p", s.projectName,
		"exec", "-T", service)
	cmd.Args = append(cmd.Args, command...)
	cmd.Dir = "."

	return cmd.CombinedOutput()
}

func (s *ContainerIntegrationTestSuite) TestMavenIntegration() {
	// Create a simple Maven project
	s.createMavenTestProject()

	// Deploy artifact using Maven
	output, err := s.execInContainer("maven-client", "mvn", "deploy", "-DskipTests")
	require.NoError(s.T(), err, "Maven deploy failed: %s", string(output))

	// Clean local repository and try to download
	_, err = s.execInContainer("maven-client", "rm", "-rf", "/root/.m2/repository")
	require.NoError(s.T(), err, "Failed to clean Maven local repository")

	// Try to resolve dependencies (this will download from our server)
	output, err = s.execInContainer("maven-client", "mvn", "dependency:resolve")
	require.NoError(s.T(), err, "Maven dependency resolution failed: %s", string(output))
}

func (s *ContainerIntegrationTestSuite) TestNPMIntegration() {
	// Create NPM project and configure registry
	s.createNPMTestProject()

	// Configure npm to use our registry
	_, err := s.execInContainer("npm-client", "npm", "config", "set", "registry", "http://ganje-server:8080/npm-repo")
	require.NoError(s.T(), err, "Failed to configure npm registry")

	// Publish package
	output, err := s.execInContainer("npm-client", "npm", "publish")
	require.NoError(s.T(), err, "npm publish failed: %s", string(output))

	// Install package in a different directory
	_, err = s.execInContainer("npm-client", "mkdir", "-p", "/tmp/test-install")
	require.NoError(s.T(), err, "Failed to create install directory")

	output, err = s.execInContainer("npm-client", "sh", "-c", "cd /tmp/test-install && npm install test-package@1.0.0")
	require.NoError(s.T(), err, "npm install failed: %s", string(output))
}

func (s *ContainerIntegrationTestSuite) TestGoModulesIntegration() {
	// Create Go module
	s.createGoTestModule()

	// Try to fetch module (this will use our GOPROXY)
	output, err := s.execInContainer("go-client", "go", "mod", "download", "example.com/test-module@v1.0.0")
	require.NoError(s.T(), err, "Go module download failed: %s", string(output))
}

func (s *ContainerIntegrationTestSuite) TestCargoIntegration() {
	// Create Cargo project
	s.createCargoTestProject()

	// Publish crate
	output, err := s.execInContainer("cargo-client", "cargo", "publish", "--allow-dirty")
	require.NoError(s.T(), err, "Cargo publish failed: %s", string(output))

	// Try to use the crate in a new project
	_, err = s.execInContainer("cargo-client", "mkdir", "-p", "/tmp/cargo-test")
	require.NoError(s.T(), err, "Failed to create cargo test directory")

	output, err = s.execInContainer("cargo-client", "sh", "-c", "cd /tmp/cargo-test && cargo init && cargo add test-crate@1.0.0")
	require.NoError(s.T(), err, "Cargo add failed: %s", string(output))
}

func (s *ContainerIntegrationTestSuite) TestHelmIntegration() {
	// Create Helm chart
	s.createHelmTestChart()

	// Add our repository
	_, err := s.execInContainer("helm-client", "helm", "repo", "add", "ganje", "http://ganje-server:8080/helm-repo")
	require.NoError(s.T(), err, "Failed to add Helm repository")

	// Package and push chart
	output, err := s.execInContainer("helm-client", "helm", "package", "test-chart")
	require.NoError(s.T(), err, "Helm package failed: %s", string(output))

	// Push to repository (if supported)
	output, err = s.execInContainer("helm-client", "helm", "push", "test-chart-1.0.0.tgz", "ganje")
	if err != nil {
		s.T().Logf("Helm push not supported, trying alternative upload method: %s", string(output))
		// Alternative: use curl to upload
		_, err = s.execInContainer("helm-client", "curl", "-X", "PUT",
			"http://ganje-server:8080/helm-repo/charts/test-chart-1.0.0.tgz",
			"--data-binary", "@test-chart-1.0.0.tgz")
		require.NoError(s.T(), err, "Chart upload failed")
	}

	// Update repository and install chart
	_, err = s.execInContainer("helm-client", "helm", "repo", "update")
	require.NoError(s.T(), err, "Failed to update Helm repository")

	output, err = s.execInContainer("helm-client", "helm", "search", "repo", "ganje/test-chart")
	require.NoError(s.T(), err, "Helm search failed: %s", string(output))
}

// Helper methods to create test projects

func (s *ContainerIntegrationTestSuite) createMavenTestProject() {
	pomXML := `<?xml version="1.0" encoding="UTF-8"?>
<project xmlns="http://maven.apache.org/POM/4.0.0"
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://maven.apache.org/POM/4.0.0 
         http://maven.apache.org/xsd/maven-4.0.0.xsd">
    <modelVersion>4.0.0</modelVersion>
    
    <groupId>com.example</groupId>
    <artifactId>test-artifact</artifactId>
    <version>1.0.0</version>
    <packaging>jar</packaging>
    
    <distributionManagement>
        <repository>
            <id>ganje-maven-repo</id>
            <url>http://ganje-server:8080/maven-repo</url>
        </repository>
    </distributionManagement>
    
    <properties>
        <maven.compiler.source>17</maven.compiler.source>
        <maven.compiler.target>17</maven.compiler.target>
    </properties>
</project>`

	err := os.WriteFile("maven-test/pom.xml", []byte(pomXML), 0644)
	require.NoError(s.T(), err, "Failed to create Maven pom.xml")

	// Create source directory and a simple Java class
	err = os.MkdirAll("maven-test/src/main/java/com/example", 0755)
	require.NoError(s.T(), err, "Failed to create Maven source directory")

	javaClass := `package com.example;

public class TestClass {
    public String getMessage() {
        return "Hello from test artifact!";
    }
}`

	err = os.WriteFile("maven-test/src/main/java/com/example/TestClass.java", []byte(javaClass), 0644)
	require.NoError(s.T(), err, "Failed to create Java class")
}

func (s *ContainerIntegrationTestSuite) createNPMTestProject() {
	packageJSON := `{
  "name": "test-package",
  "version": "1.0.0",
  "description": "Test package for integration testing",
  "main": "index.js",
  "scripts": {
    "test": "echo \"Error: no test specified\" && exit 1"
  },
  "author": "Test Author",
  "license": "MIT"
}`

	err := os.WriteFile("npm-test/package.json", []byte(packageJSON), 0644)
	require.NoError(s.T(), err, "Failed to create package.json")

	indexJS := `module.exports = {
  getMessage: function() {
    return "Hello from test package!";
  }
};`

	err = os.WriteFile("npm-test/index.js", []byte(indexJS), 0644)
	require.NoError(s.T(), err, "Failed to create index.js")
}

func (s *ContainerIntegrationTestSuite) createGoTestModule() {
	goMod := `module example.com/test-module

go 1.21`

	err := os.WriteFile("go-test/go.mod", []byte(goMod), 0644)
	require.NoError(s.T(), err, "Failed to create go.mod")

	mainGo := `package main

import "fmt"

func main() {
    fmt.Println("Hello from test module!")
}

func GetMessage() string {
    return "Hello from test module!"
}`

	err = os.WriteFile("go-test/main.go", []byte(mainGo), 0644)
	require.NoError(s.T(), err, "Failed to create main.go")
}

func (s *ContainerIntegrationTestSuite) createCargoTestProject() {
	cargoToml := `[package]
name = "test-crate"
version = "1.0.0"
edition = "2021"
description = "Test crate for integration testing"
license = "MIT"

[dependencies]`

	err := os.WriteFile("cargo-test/Cargo.toml", []byte(cargoToml), 0644)
	require.NoError(s.T(), err, "Failed to create Cargo.toml")

	err = os.MkdirAll("cargo-test/src", 0755)
	require.NoError(s.T(), err, "Failed to create Cargo src directory")

	libRS := `pub fn get_message() -> &'static str {
    "Hello from test crate!"
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn it_works() {
        assert_eq!(get_message(), "Hello from test crate!");
    }
}`

	err = os.WriteFile("cargo-test/src/lib.rs", []byte(libRS), 0644)
	require.NoError(s.T(), err, "Failed to create lib.rs")
}

func (s *ContainerIntegrationTestSuite) createHelmTestChart() {
	err := os.MkdirAll("helm-test/test-chart/templates", 0755)
	require.NoError(s.T(), err, "Failed to create Helm chart directory")

	chartYAML := `apiVersion: v2
name: test-chart
description: A test Helm chart for integration testing
type: application
version: 1.0.0
appVersion: "1.0.0"`

	err = os.WriteFile("helm-test/test-chart/Chart.yaml", []byte(chartYAML), 0644)
	require.NoError(s.T(), err, "Failed to create Chart.yaml")

	valuesYAML := `replicaCount: 1
image:
  repository: nginx
  tag: latest
  pullPolicy: IfNotPresent
service:
  type: ClusterIP
  port: 80`

	err = os.WriteFile("helm-test/test-chart/values.yaml", []byte(valuesYAML), 0644)
	require.NoError(s.T(), err, "Failed to create values.yaml")

	deploymentYAML := `apiVersion: apps/v1
kind: Deployment
metadata:
  name: {{ include "test-chart.fullname" . }}
spec:
  replicas: {{ .Values.replicaCount }}
  selector:
    matchLabels:
      app: {{ include "test-chart.name" . }}
  template:
    metadata:
      labels:
        app: {{ include "test-chart.name" . }}
    spec:
      containers:
        - name: {{ .Chart.Name }}
          image: "{{ .Values.image.repository }}:{{ .Values.image.tag }}"
          imagePullPolicy: {{ .Values.image.pullPolicy }}
          ports:
            - name: http
              containerPort: 80
              protocol: TCP`

	err = os.WriteFile("helm-test/test-chart/templates/deployment.yaml", []byte(deploymentYAML), 0644)
	require.NoError(s.T(), err, "Failed to create deployment.yaml")
}
