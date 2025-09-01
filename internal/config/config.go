package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Config represents the application configuration
type Config struct {
	Server       ServerConfig       `yaml:"server"`
	Database     DatabaseConfig     `yaml:"database"`
	Storage      StorageConfig      `yaml:"storage"`
	Auth         AuthConfig         `yaml:"auth"`
	Logging      LoggingConfig      `yaml:"logging"`
	Metrics      MetricsConfig      `yaml:"metrics"`
	Messaging    MessagingConfig    `yaml:"messaging"`
	Webhook      WebhookConfig      `yaml:"webhook"`
	Repositories []RepositoryConfig `yaml:"repositories"`
}

// ServerConfig contains server-related configuration
type ServerConfig struct {
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
}

// DatabaseConfig contains database connection configuration
type DatabaseConfig struct {
	Driver   string `yaml:"driver"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Username string `yaml:"username"`
	Password string `yaml:"password"`
	Database string `yaml:"database"`
	SSLMode  string `yaml:"ssl_mode"`
}

// StorageConfig contains storage configuration
type StorageConfig struct {
	Type      string            `yaml:"type"`
	LocalPath string            `yaml:"local_path,omitempty"`
	Options   map[string]string `yaml:"options,omitempty"`
}

// AuthConfig contains authentication configuration
type AuthConfig struct {
	OAuthServer  string     `yaml:"oauth_server"`
	JWTSecret    string     `yaml:"jwt_secret"`
	Realms       []Realm    `yaml:"realms"`
	OIDC         OIDCConfig `yaml:"oidc,omitempty"`
}

// OIDCConfig contains OIDC client configuration
type OIDCConfig struct {
	ClientID     string `yaml:"client_id"`
	ClientSecret string `yaml:"client_secret"`
	RedirectURI  string `yaml:"redirect_uri"`
}

// Realm represents access control realm
type Realm struct {
	Name        string   `yaml:"name"`
	Permissions []string `yaml:"permissions"`
}

// LoggingConfig contains logging configuration
type LoggingConfig struct {
	Level  string `yaml:"level"`
	Format string `yaml:"format"`
	Output string `yaml:"output"`
}

// MetricsConfig contains metrics configuration
type MetricsConfig struct {
	Enabled        bool   `yaml:"enabled"`
	Path           string `yaml:"path"`
	SeparateServer bool   `yaml:"separate_server"`
	Port           int    `yaml:"port"`
}

// MessagingConfig contains RabbitMQ settings for event publishing
type MessagingConfig struct {
	RabbitMQ RabbitMQConfig `yaml:"rabbitmq"`
}

type RabbitMQConfig struct {
	Enabled      bool   `yaml:"enabled"`
	URL          string `yaml:"url"`
	Exchange     string `yaml:"exchange"`
	ExchangeType string `yaml:"exchange_type"`
	RoutingKey   string `yaml:"routing_key"`
}

// WebhookConfig controls webhook dispatcher settings
type WebhookConfig struct {
	Enabled          bool `yaml:"enabled"`
	Workers          int  `yaml:"workers"`
	MaxRetries       int  `yaml:"max_retries"`
	InitialBackoffMs int  `yaml:"initial_backoff_ms"`
	MaxBackoffMs     int  `yaml:"max_backoff_ms"`
	HTTPTimeoutMs    int  `yaml:"http_timeout_ms"`
}

// RepositoryConfig contains repository configuration
type RepositoryConfig struct {
	Name         string            `yaml:"name"`
	Type         string            `yaml:"type"` // local, remote, virtual
	ArtifactType string            `yaml:"artifact_type"`
	URL          string            `yaml:"url,omitempty"`
	Upstream     []string          `yaml:"upstream,omitempty"`
	Options      map[string]string `yaml:"options,omitempty"`
}

// Load loads configuration from a YAML file
func Load(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return &config, nil
}

// GetConnectionString returns database connection string
func (d *DatabaseConfig) GetConnectionString() string {
	switch d.Driver {
	case "postgres":
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			d.Host, d.Port, d.Username, d.Password, d.Database, d.SSLMode)
	case "mysql":
		return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
			d.Username, d.Password, d.Host, d.Port, d.Database)
	default:
		return ""
	}
}
