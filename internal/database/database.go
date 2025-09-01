package database

import (
	"context"
	"fmt"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

// DB wraps GORM database connection
type DB struct {
	conn *gorm.DB
}

// New creates a new database connection
func New(driver, connectionString string) (*DB, error) {
	var dialector gorm.Dialector

	switch driver {
	case "postgres":
		dialector = postgres.Open(connectionString)
	case "mysql":
		dialector = mysql.Open(connectionString)
	case "sqlite":
		dialector = sqlite.Open(connectionString)
	default:
		return nil, fmt.Errorf("unsupported database driver: %s", driver)
	}

	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	// Run migrations
	if err := AutoMigrate(db); err != nil {
		return nil, fmt.Errorf("failed to run migrations: %w", err)
	}

	return &DB{conn: db}, nil
}

// SaveArtifact saves artifact information to database
func (db *DB) SaveArtifact(ctx context.Context, artifact *ArtifactInfo) error {
	return db.conn.WithContext(ctx).Create(artifact).Error
}

// GetArtifactByPath retrieves artifact by path
func (db *DB) GetArtifactByPath(ctx context.Context, repoName, path string) (*ArtifactInfo, error) {
	var artifactInfo ArtifactInfo
	err := db.conn.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = artifact_infos.repository_id").
		Where("repositories.name = ? AND artifact_infos.path = ?", repoName, path).
		First(&artifactInfo).Error

	if err != nil {
		return nil, err
	}

	return &artifactInfo, nil
}

// GetArtifactsByRepository retrieves all artifacts for a repository
func (db *DB) GetArtifactsByRepository(ctx context.Context, repoName string) ([]*ArtifactInfo, error) {
	var artifacts []*ArtifactInfo
	err := db.conn.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = artifact_infos.repository_id").
		Where("repositories.name = ?", repoName).
		Find(&artifacts).Error

	if err != nil {
		return nil, err
	}

	return artifacts, nil
}

// DeleteArtifactByPath deletes artifact by path
func (db *DB) DeleteArtifactByPath(ctx context.Context, repoName, path string) error {
	return db.conn.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = artifact_infos.repository_id").
		Where("repositories.name = ? AND artifact_infos.path = ?", repoName, path).
		Delete(&ArtifactInfo{}).Error
}

// IncrementPullCount increments pull count for an artifact
func (db *DB) IncrementPullCount(ctx context.Context, artifactID uint) error {
	return db.conn.WithContext(ctx).
		Model(&ArtifactInfo{}).
		Where("id = ?", artifactID).
		UpdateColumn("pull_count", gorm.Expr("pull_count + ?", 1)).Error
}

// IncrementPushCount increments push count for an artifact
func (db *DB) IncrementPushCount(ctx context.Context, artifactID uint) error {
	return db.conn.WithContext(ctx).
		Model(&ArtifactInfo{}).
		Where("id = ?", artifactID).
		UpdateColumn("push_count", gorm.Expr("push_count + ?", 1)).Error
}

// Statistics represents repository statistics
type Statistics struct {
	TotalArtifacts int64 `json:"total_artifacts"`
	TotalSize      int64 `json:"total_size"`
	PullCount      int64 `json:"pull_count"`
	PushCount      int64 `json:"push_count"`
}

// GetRepositoryStatistics returns repository statistics
func (db *DB) GetRepositoryStatistics(ctx context.Context, repoName string) (*Statistics, error) {
	var stats Statistics

	err := db.conn.WithContext(ctx).
		Model(&ArtifactInfo{}).
		Select("COUNT(*) as total_artifacts, SUM(size) as total_size, SUM(pull_count) as pull_count, SUM(push_count) as push_count").
		Joins("JOIN repositories ON repositories.id = artifact_infos.repository_id").
		Where("repositories.name = ?", repoName).
		Scan(&stats).Error

	return &stats, err
}

// SaveRepository saves repository information
func (db *DB) SaveRepository(ctx context.Context, repo *Repository) error {
	return db.conn.WithContext(ctx).Create(repo).Error
}

// GetRepository retrieves repository by name
func (db *DB) GetRepository(ctx context.Context, name string) (*Repository, error) {
	var repo Repository
	err := db.conn.WithContext(ctx).Where("name = ?", name).First(&repo).Error
	if err != nil {
		return nil, err
	}
	return &repo, nil
}

// ListRepositories lists all repositories
func (db *DB) ListRepositories(ctx context.Context) ([]*Repository, error) {
	var repos []*Repository
	err := db.conn.WithContext(ctx).Find(&repos).Error
	return repos, err
}

// UpdateRepository updates repository fields
func (db *DB) UpdateRepository(ctx context.Context, name string, updates map[string]interface{}) error {
	return db.conn.WithContext(ctx).Model(&Repository{}).Where("name = ?", name).Updates(updates).Error
}

// DeleteRepository deletes repository by name
func (db *DB) DeleteRepository(ctx context.Context, name string) error {
	return db.conn.WithContext(ctx).Where("name = ?", name).Delete(&Repository{}).Error
}

// SaveUser saves user information
func (db *DB) SaveUser(ctx context.Context, user *User) error {
	return db.conn.WithContext(ctx).Create(user).Error
}

// GetUserByUsername retrieves user by username
func (db *DB) GetUserByUsername(ctx context.Context, username string) (*User, error) {
	var user User
	err := db.conn.WithContext(ctx).Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// LogAccess logs user access
func (db *DB) LogAccess(ctx context.Context, log *AccessLog) error {
	return db.conn.WithContext(ctx).Create(log).Error
}

// SaveCacheEntry saves cache entry for remote repositories
func (db *DB) SaveCacheEntry(ctx context.Context, entry *CacheEntry) error {
	return db.conn.WithContext(ctx).Create(entry).Error
}

// GetCacheEntry retrieves cache entry
func (db *DB) GetCacheEntry(ctx context.Context, repoID uint, path string) (*CacheEntry, error) {
	var entry CacheEntry
	err := db.conn.WithContext(ctx).
		Where("repository_id = ? AND path = ?", repoID, path).
		First(&entry).Error
	if err != nil {
		return nil, err
	}
	return &entry, nil
}

// DeleteCacheEntry deletes cache entry
func (db *DB) DeleteCacheEntry(ctx context.Context, repoID uint, path string) error {
	return db.conn.WithContext(ctx).
		Where("repository_id = ? AND path = ?", repoID, path).
		Delete(&CacheEntry{}).Error
}

// GetAllRepositories retrieves all repositories (alias for ListRepositories for consistency)
func (db *DB) GetAllRepositories() ([]Repository, error) {
	var repos []Repository
	err := db.conn.Find(&repos).Error
	return repos, err
}

// GetArtifact retrieves a specific artifact
func (db *DB) GetArtifact(ctx context.Context, repositoryName, name, version string) (*ArtifactInfo, error) {
	var artifact ArtifactInfo
	err := db.conn.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = artifact_infos.repository_id").
		Where("repositories.name = ? AND artifact_infos.name = ? AND artifact_infos.version = ?", repositoryName, name, version).
		First(&artifact).Error
	if err != nil {
		return nil, err
	}
	return &artifact, nil
}

// UpdateArtifactYanked updates the yanked flag for an artifact version scoped to repository
func (db *DB) UpdateArtifactYanked(ctx context.Context, repositoryName, name, version string, yanked bool) error {
	return db.conn.WithContext(ctx).
		Model(&ArtifactInfo{}).
		Joins("JOIN repositories ON repositories.id = artifact_infos.repository_id").
		Where("repositories.name = ? AND artifact_infos.name = ? AND artifact_infos.version = ?", repositoryName, name, version).
		Update("yanked", yanked).Error
}

// CreateWebhook creates a webhook for a repository name
func (db *DB) CreateWebhook(ctx context.Context, repoName string, hook *Webhook) error {
	var repo Repository
	if err := db.conn.WithContext(ctx).Where("name = ?", repoName).First(&repo).Error; err != nil {
		return err
	}
	hook.RepositoryID = repo.ID
	return db.conn.WithContext(ctx).Create(hook).Error
}

// UpdateWebhook updates webhook fields by ID
func (db *DB) UpdateWebhook(ctx context.Context, id uint, updates map[string]interface{}) error {
	return db.conn.WithContext(ctx).Model(&Webhook{}).Where("id = ?", id).Updates(updates).Error
}

// DeleteWebhook deletes a webhook by ID
func (db *DB) DeleteWebhook(ctx context.Context, id uint) error {
	return db.conn.WithContext(ctx).Where("id = ?", id).Delete(&Webhook{}).Error
}

// GetWebhook retrieves a webhook by ID
func (db *DB) GetWebhook(ctx context.Context, id uint) (*Webhook, error) {
	var hook Webhook
	if err := db.conn.WithContext(ctx).Where("id = ?", id).First(&hook).Error; err != nil {
		return nil, err
	}
	return &hook, nil
}

// ListWebhooksByRepository lists webhooks by repository name
func (db *DB) ListWebhooksByRepository(ctx context.Context, repoName string) ([]*Webhook, error) {
	var hooks []*Webhook
	err := db.conn.WithContext(ctx).
		Joins("JOIN repositories ON repositories.id = webhooks.repository_id").
		Where("repositories.name = ?", repoName).
		Find(&hooks).Error
	if err != nil {
		return nil, err
	}
	return hooks, nil
}

// RecordWebhookDelivery stores a delivery log entry
func (db *DB) RecordWebhookDelivery(ctx context.Context, delivery *WebhookDelivery) error {
	return db.conn.WithContext(ctx).Create(delivery).Error
}

// Close closes database connection
func (db *DB) Close() error {
	sqlDB, err := db.conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}
