package repository

import (
	"context"
	"io"
	"time"

	"github.com/hbahadorzadeh/ganje/internal/artifact"
	"github.com/hbahadorzadeh/ganje/internal/metrics"
)

// MetricsWrapper wraps a repository with metrics collection
type MetricsWrapper struct {
	repo    Repository
	metrics *metrics.MetricsService
}

// NewMetricsWrapper creates a new metrics wrapper for a repository
func NewMetricsWrapper(repo Repository, metricsService *metrics.MetricsService) Repository {
	return &MetricsWrapper{
		repo:    repo,
		metrics: metricsService,
	}
}

// GetName returns the repository name
func (m *MetricsWrapper) GetName() string {
	return m.repo.GetName()
}

// GetType returns the repository type
func (m *MetricsWrapper) GetType() Type {
	return m.repo.GetType()
}

// GetArtifactType returns the artifact type
func (m *MetricsWrapper) GetArtifactType() artifact.ArtifactType {
	return m.repo.GetArtifactType()
}

// Push pushes an artifact to the repository with metrics
func (m *MetricsWrapper) Push(ctx context.Context, path string, content io.Reader, metadata *artifact.Metadata) error {
	start := time.Now()
	err := m.repo.Push(ctx, path, content, metadata)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"push",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
		
		if err == nil {
			m.metrics.RecordArtifactUpload(m.repo.GetName(), string(m.repo.GetArtifactType()))
		}
	}
	
	return err
}

// Pull pulls an artifact from the repository with metrics
func (m *MetricsWrapper) Pull(ctx context.Context, path string) (io.ReadCloser, *artifact.Metadata, error) {
	start := time.Now()
	content, metadata, err := m.repo.Pull(ctx, path)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"pull",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
		
		if err == nil {
			m.metrics.RecordArtifactDownload(m.repo.GetName(), string(m.repo.GetArtifactType()))
		}
	}
	
	return content, metadata, err
}

// Delete deletes an artifact from the repository with metrics
func (m *MetricsWrapper) Delete(ctx context.Context, path string) error {
	start := time.Now()
	err := m.repo.Delete(ctx, path)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"delete",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
	}
	
	return err
}

// List lists artifacts in the repository with metrics
func (m *MetricsWrapper) List(ctx context.Context, prefix string) ([]string, error) {
	start := time.Now()
	artifacts, err := m.repo.List(ctx, prefix)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"list",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
	}
	
	return artifacts, err
}

// GetIndex returns the repository index with metrics
func (m *MetricsWrapper) GetIndex(ctx context.Context, indexType string) (io.ReadCloser, error) {
	start := time.Now()
	index, err := m.repo.GetIndex(ctx, indexType)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"get_index",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
	}
	
	return index, err
}

// InvalidateCache invalidates cached artifacts with metrics
func (m *MetricsWrapper) InvalidateCache(ctx context.Context, path string) error {
	start := time.Now()
	err := m.repo.InvalidateCache(ctx, path)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"invalidate_cache",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
	}
	
	return err
}

// RebuildIndex rebuilds the repository index with metrics
func (m *MetricsWrapper) RebuildIndex(ctx context.Context) error {
	start := time.Now()
	err := m.repo.RebuildIndex(ctx)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"rebuild_index",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
	}
	
	return err
}

// GetStatistics returns repository statistics with metrics
func (m *MetricsWrapper) GetStatistics(ctx context.Context) (*Statistics, error) {
	start := time.Now()
	stats, err := m.repo.GetStatistics(ctx)
	duration := time.Since(start)
	
	status := "success"
	if err != nil {
		status = "error"
	}
	
	if m.metrics != nil {
		m.metrics.RecordRepositoryOperation(
			m.repo.GetName(),
			"get_statistics",
			string(m.repo.GetArtifactType()),
			status,
			duration,
		)
		
		// Update storage size metrics if available
		if err == nil && stats != nil {
			m.metrics.SetArtifactStorageSize(
				m.repo.GetName(),
				string(m.repo.GetArtifactType()),
				float64(stats.TotalSize),
			)
		}
	}
	
	return stats, err
}
