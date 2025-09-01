package integration

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
)

// createMockJar creates a mock JAR file for Maven testing
func createMockJar() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add META-INF/MANIFEST.MF
	manifest, _ := zw.Create("META-INF/MANIFEST.MF")
	manifest.Write([]byte(`Manifest-Version: 1.0
Created-By: Test
Main-Class: com.example.Main
`))

	// Add a simple class file
	classFile, _ := zw.Create("com/example/Main.class")
	classFile.Write([]byte{0xCA, 0xFE, 0xBA, 0xBE}) // Java class file magic number

	zw.Close()
	return buf.Bytes()
}

// createMockNpmPackage creates a mock NPM package tarball
func createMockNpmPackage() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add package.json
	packageJson := map[string]interface{}{
		"name":        "test-package",
		"version":     "1.0.0",
		"description": "Test NPM package",
		"main":        "index.js",
		"scripts": map[string]string{
			"test": "echo \"Error: no test specified\" && exit 1",
		},
		"keywords": []string{"test"},
		"author":   "Test Author",
		"license":  "MIT",
	}

	packageJsonBytes, _ := json.Marshal(packageJson)

	hdr := &tar.Header{
		Name: "package/package.json",
		Mode: 0644,
		Size: int64(len(packageJsonBytes)),
	}
	tw.WriteHeader(hdr)
	tw.Write(packageJsonBytes)

	// Add index.js
	indexJs := []byte(`console.log("Hello from test-package");
module.exports = function() {
    return "test-package";
};`)

	hdr = &tar.Header{
		Name: "package/index.js",
		Mode: 0644,
		Size: int64(len(indexJs)),
	}
	tw.WriteHeader(hdr)
	tw.Write(indexJs)

	tw.Close()
	gw.Close()
	return buf.Bytes()
}

// createMockPythonWheel creates a mock Python wheel file
func createMockPythonWheel() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add METADATA
	metadata, _ := zw.Create("test_package-1.0.0.dist-info/METADATA")
	metadata.Write([]byte(`Metadata-Version: 2.1
Name: test-package
Version: 1.0.0
Summary: Test Python package
Author: Test Author
Author-email: test@example.com
License: MIT
`))

	// Add WHEEL
	wheel, _ := zw.Create("test_package-1.0.0.dist-info/WHEEL")
	wheel.Write([]byte(`Wheel-Version: 1.0
Generator: test
Root-Is-Purelib: true
Tag: py3-none-any
`))

	// Add __init__.py
	initPy, _ := zw.Create("test_package/__init__.py")
	initPy.Write([]byte(`"""Test package"""
__version__ = "1.0.0"

def hello():
    return "Hello from test-package"
`))

	zw.Close()
	return buf.Bytes()
}

// createMockDockerManifest creates a mock Docker manifest
func createMockDockerManifest() []byte {
	manifest := map[string]interface{}{
		"schemaVersion": 2,
		"mediaType":     "application/vnd.docker.distribution.manifest.v2+json",
		"config": map[string]interface{}{
			"mediaType": "application/vnd.docker.container.image.v1+json",
			"size":      1234,
			"digest":    "sha256:abcd1234",
		},
		"layers": []map[string]interface{}{
			{
				"mediaType": "application/vnd.docker.image.rootfs.diff.tar.gzip",
				"size":      5678,
				"digest":    "sha256:efgh5678",
			},
		},
	}

	manifestBytes, _ := json.Marshal(manifest)
	return manifestBytes
}

// createMockHelmChart creates a mock Helm chart tarball
func createMockHelmChart() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add Chart.yaml
	chartYaml := `apiVersion: v2
name: test-chart
description: A Helm chart for testing
type: application
version: 1.0.0
appVersion: "1.0.0"
`

	hdr := &tar.Header{
		Name: "test-chart/Chart.yaml",
		Mode: 0644,
		Size: int64(len(chartYaml)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(chartYaml))

	// Add values.yaml
	valuesYaml := `replicaCount: 1
image:
  repository: nginx
  tag: latest
  pullPolicy: IfNotPresent
service:
  type: ClusterIP
  port: 80
`

	hdr = &tar.Header{
		Name: "test-chart/values.yaml",
		Mode: 0644,
		Size: int64(len(valuesYaml)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(valuesYaml))

	// Add deployment template
	deploymentYaml := `apiVersion: apps/v1
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
          ports:
            - containerPort: 80
`

	hdr = &tar.Header{
		Name: "test-chart/templates/deployment.yaml",
		Mode: 0644,
		Size: int64(len(deploymentYaml)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(deploymentYaml))

	tw.Close()
	gw.Close()
	return buf.Bytes()
}

// createMockGoModule creates a mock Go module zip
func createMockGoModule() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add go.mod
	goMod := `module example.com/test/module

go 1.19

require (
    github.com/stretchr/testify v1.8.0
)
`

	goModFile, _ := zw.Create("example.com/test/module@v1.0.0/go.mod")
	goModFile.Write([]byte(goMod))

	// Add main.go
	mainGo := `package module

import "fmt"

// Hello returns a greeting message
func Hello(name string) string {
    return fmt.Sprintf("Hello, %s!", name)
}
`

	mainGoFile, _ := zw.Create("example.com/test/module@v1.0.0/main.go")
	mainGoFile.Write([]byte(mainGo))

	zw.Close()
	return buf.Bytes()
}

// createMockCargoCrate creates a mock Rust crate
func createMockCargoCrate() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add Cargo.toml
	cargoToml := `[package]
name = "test-crate"
version = "1.0.0"
edition = "2021"
description = "A test crate"
license = "MIT"

[dependencies]
`

	hdr := &tar.Header{
		Name: "test-crate-1.0.0/Cargo.toml",
		Mode: 0644,
		Size: int64(len(cargoToml)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(cargoToml))

	// Add lib.rs
	libRs := `//! Test crate

/// Returns a greeting message
pub fn hello(name: &str) -> String {
    format!("Hello, {}!", name)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_hello() {
        assert_eq!(hello("World"), "Hello, World!");
    }
}
`

	hdr = &tar.Header{
		Name: "test-crate-1.0.0/src/lib.rs",
		Mode: 0644,
		Size: int64(len(libRs)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(libRs))

	tw.Close()
	gw.Close()
	return buf.Bytes()
}

// createMockNuGetPackage creates a mock NuGet package
func createMockNuGetPackage() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add nuspec file
	nuspec := `<?xml version="1.0" encoding="utf-8"?>
<package xmlns="http://schemas.microsoft.com/packaging/2010/07/nuspec.xsd">
  <metadata>
    <id>TestPackage</id>
    <version>1.0.0</version>
    <title>Test Package</title>
    <authors>Test Author</authors>
    <description>A test NuGet package</description>
    <language>en-US</language>
    <licenseUrl>https://opensource.org/licenses/MIT</licenseUrl>
  </metadata>
</package>`

	nuspecFile, _ := zw.Create("TestPackage.nuspec")
	nuspecFile.Write([]byte(nuspec))

	// Add assembly
	assembly := []byte{0x4D, 0x5A} // PE header magic number

	assemblyFile, _ := zw.Create("lib/net45/TestPackage.dll")
	assemblyFile.Write(assembly)

	zw.Close()
	return buf.Bytes()
}

// createMockRubyGem creates a mock Ruby gem
func createMockRubyGem() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add gemspec
	gemspec := `Gem::Specification.new do |spec|
  spec.name          = "test-gem"
  spec.version       = "1.0.0"
  spec.authors       = ["Test Author"]
  spec.email         = ["test@example.com"]
  spec.summary       = "A test gem"
  spec.description   = "A test Ruby gem for integration testing"
  spec.homepage      = "https://example.com"
  spec.license       = "MIT"

  spec.files         = ["lib/test_gem.rb"]
  spec.require_paths = ["lib"]
end`

	hdr := &tar.Header{
		Name: "test-gem.gemspec",
		Mode: 0644,
		Size: int64(len(gemspec)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(gemspec))

	// Add lib file
	libFile := `module TestGem
  VERSION = "1.0.0"
  
  def self.hello(name = "World")
    "Hello, #{name}!"
  end
end`

	hdr = &tar.Header{
		Name: "lib/test_gem.rb",
		Mode: 0644,
		Size: int64(len(libFile)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(libFile))

	tw.Close()
	gw.Close()
	return buf.Bytes()
}

// createMockAnsibleCollection creates a mock Ansible collection
func createMockAnsibleCollection() []byte {
	var buf bytes.Buffer
	gw := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gw)

	// Add galaxy.yml
	galaxyYml := `namespace: test
name: collection
version: 1.0.0
readme: README.md
authors:
  - Test Author <test@example.com>
description: A test Ansible collection
license:
  - MIT
tags:
  - test
dependencies: {}
repository: https://github.com/test/collection
documentation: https://github.com/test/collection
homepage: https://github.com/test/collection
issues: https://github.com/test/collection/issues
`

	hdr := &tar.Header{
		Name: "galaxy.yml",
		Mode: 0644,
		Size: int64(len(galaxyYml)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(galaxyYml))

	// Add a simple module
	module := `#!/usr/bin/python
# -*- coding: utf-8 -*-

DOCUMENTATION = '''
---
module: test_module
short_description: A test module
description:
    - This is a test module for integration testing
author:
    - Test Author (@testauthor)
'''

EXAMPLES = '''
- name: Test module
  test.collection.test_module:
    name: test
'''

from ansible.module_utils.basic import AnsibleModule

def main():
    module = AnsibleModule(
        argument_spec=dict(
            name=dict(type='str', required=True),
        ),
        supports_check_mode=True
    )
    
    result = dict(
        changed=False,
        message='Hello from test module'
    )
    
    module.exit_json(**result)

if __name__ == '__main__':
    main()
`

	hdr = &tar.Header{
		Name: "plugins/modules/test_module.py",
		Mode: 0644,
		Size: int64(len(module)),
	}
	tw.WriteHeader(hdr)
	tw.Write([]byte(module))

	tw.Close()
	gw.Close()
	return buf.Bytes()
}

// createMockTerraformProvider creates a mock Terraform provider
func createMockTerraformProvider() []byte {
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)

	// Add provider binary (mock)
	binary := []byte{0x7F, 0x45, 0x4C, 0x46} // ELF header magic number

	binaryFile, _ := zw.Create("terraform-provider-test_v1.0.0")
	binaryFile.Write(binary)

	// Add manifest
	manifest := map[string]interface{}{
		"version": 1,
		"metadata": map[string]interface{}{
			"protocol_versions": []string{"5.0"},
		},
	}

	manifestBytes, _ := json.Marshal(manifest)
	manifestFile, _ := zw.Create("terraform-provider-test_v1.0.0_manifest.json")
	manifestFile.Write(manifestBytes)

	zw.Close()
	return buf.Bytes()
}
