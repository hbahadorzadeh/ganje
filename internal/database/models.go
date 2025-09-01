package database

import (
	"time"

	"gorm.io/gorm"
)

// Repository represents a repository in the database
type Repository struct {
	ID           uint      `gorm:"primaryKey"`
	Name         string    `gorm:"uniqueIndex;not null"`
	Type         string    `gorm:"not null"` // local, remote, virtual
	ArtifactType string    `gorm:"not null"`
	URL          string    `gorm:""`
	Config       string    `gorm:"type:text"` // JSON configuration
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	
	// Statistics
	TotalArtifacts int64 `gorm:"default:0"`
	TotalSize      int64 `gorm:"default:0"`
	PullCount      int64 `gorm:"default:0"`
	PushCount      int64 `gorm:"default:0"`
	
	// Relationships
	Artifacts []ArtifactInfo `gorm:"foreignKey:RepositoryID"`
}

// ArtifactInfo represents an artifact in the database
type ArtifactInfo struct {
	ID           uint      `gorm:"primaryKey"`
	RepositoryID uint      `gorm:"not null;index"`
	Type         string    `gorm:"not null"`
	Name         string    `gorm:"not null;index"`
	Version      string    `gorm:"not null;index"`
	Group        string    `gorm:"index"`
	Path         string    `gorm:"not null;uniqueIndex"`
	Size         int64     `gorm:"not null"`
	Checksum     string    `gorm:"not null"`
	Yanked       bool      `gorm:"not null;default:false"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	
	// Statistics
	PullCount int64 `gorm:"default:0"`
	PushCount int64 `gorm:"default:0"`
	
	// Relationships
	Repository Repository `gorm:"foreignKey:RepositoryID"`
}

// User represents a user in the system
type User struct {
	ID        uint      `gorm:"primaryKey"`
	Username  string    `gorm:"uniqueIndex;not null"`
	Email     string    `gorm:"uniqueIndex;not null"`
	Realms    string    `gorm:"type:text"` // JSON array of realms
	Active    bool      `gorm:"default:true"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
	UpdatedAt time.Time `gorm:"autoUpdateTime"`
	
	// Relationships
	AccessLogs []AccessLog `gorm:"foreignKey:UserID"`
}

// AccessLog represents access logs
type AccessLog struct {
	ID           uint      `gorm:"primaryKey"`
	UserID       uint      `gorm:"not null;index"`
	RepositoryID uint      `gorm:"not null;index"`
	ArtifactID   uint      `gorm:"index"`
	Action       string    `gorm:"not null"` // pull, push, delete
	Path         string    `gorm:"not null"`
	IPAddress    string    `gorm:""`
	UserAgent    string    `gorm:""`
	Success      bool      `gorm:"not null"`
	ErrorMessage string    `gorm:""`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	
	// Relationships
	User       User         `gorm:"foreignKey:UserID"`
	Repository Repository   `gorm:"foreignKey:RepositoryID"`
	Artifact   ArtifactInfo `gorm:"foreignKey:ArtifactID"`
}

// CacheEntry represents cached artifacts for remote repositories
type CacheEntry struct {
	ID           uint      `gorm:"primaryKey"`
	RepositoryID uint      `gorm:"not null;index"`
	Path         string    `gorm:"not null;index"`
	LocalPath    string    `gorm:"not null"`
	Size         int64     `gorm:"not null"`
	Checksum     string    `gorm:"not null"`
	ExpiresAt    time.Time `gorm:"index"`
	CreatedAt    time.Time `gorm:"autoCreateTime"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime"`
	
	// Relationships
	Repository Repository `gorm:"foreignKey:RepositoryID"`
}

// VirtualRepositoryMapping represents mappings for virtual repositories
type VirtualRepositoryMapping struct {
	ID               uint `gorm:"primaryKey"`
	VirtualRepoID    uint `gorm:"not null;index"`
	UpstreamRepoID   uint `gorm:"not null;index"`
	Priority         int  `gorm:"not null;default:0"`
	
	// Relationships
	VirtualRepo  Repository `gorm:"foreignKey:VirtualRepoID"`
	UpstreamRepo Repository `gorm:"foreignKey:UpstreamRepoID"`
}

// Webhook defines a repository-level webhook configuration
type Webhook struct {
	ID              uint      `gorm:"primaryKey"`
	RepositoryID    uint      `gorm:"not null;index"`
	Name            string    `gorm:"not null"`
	URL             string    `gorm:"not null"`
	Events          string    `gorm:"type:text"`      // comma-separated: add,remove,change
	PayloadTemplate string    `gorm:"type:text"`      // Go template text
	HeadersJSON     string    `gorm:"type:text"`      // JSON of headers map
	Enabled         bool      `gorm:"not null;default:true"`
	// Optional credentials
	BasicUsername   string    `gorm:""`
	BasicPassword   string    `gorm:""`               // consider encryption at rest
	BearerToken     string    `gorm:""`
	SigningSecret   string    `gorm:""`
	CreatedAt       time.Time `gorm:"autoCreateTime"`
	UpdatedAt       time.Time `gorm:"autoUpdateTime"`
}

// WebhookDelivery optionally records deliveries (for observability)
type WebhookDelivery struct {
	ID          uint      `gorm:"primaryKey"`
	WebhookID   uint      `gorm:"not null;index"`
	Event       string    `gorm:"not null"`
	StatusCode  int       `gorm:""`
	Success     bool      `gorm:"index"`
	Error       string    `gorm:"type:text"`
	Payload     string    `gorm:"type:text"`
	CreatedAt   time.Time `gorm:"autoCreateTime"`
}

// AutoMigrate runs database migrations
func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&Repository{},
		&ArtifactInfo{},
		&User{},
		&AccessLog{},
		&CacheEntry{},
		&VirtualRepositoryMapping{},
		&Webhook{},
		&WebhookDelivery{},
	)
}
