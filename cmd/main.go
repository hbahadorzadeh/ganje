package main

import (
	"flag"
	"fmt"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/auth"
	"github.com/hbahadorzadeh/ganje/internal/storage"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/metrics"
	"github.com/hbahadorzadeh/ganje/internal/routes"
)

func main() {
	// Parse command line flags
	configFile := flag.String("config", "config.yaml", "Configuration file path")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Setup database
	db, err := database.New(cfg.Database.Driver, cfg.Database.GetConnectionString())
	if err != nil {
		log.Fatalf("Failed to setup database: %v", err)
	}

	// Setup storage service
	storageService := storage.NewLocalStorage(cfg.Storage.LocalPath)

	// Setup authentication services
	realmPerms := make(map[string][]auth.Permission)
	for _, realm := range cfg.Auth.Realms {
		var perms []auth.Permission
		for _, perm := range realm.Permissions {
			perms = append(perms, auth.Permission(perm))
		}
		realmPerms[realm.Name] = perms
	}
	authService := auth.NewAuthService(cfg.Auth.JWTSecret, cfg.Auth.OAuthServer, realmPerms)

	// Setup OIDC service
	var oidcService *auth.OIDCService
	if cfg.Auth.OIDC.ClientID != "" {
		oidcConfig := auth.OIDCConfig{
			ClientID:     cfg.Auth.OIDC.ClientID,
			ClientSecret: cfg.Auth.OIDC.ClientSecret,
			RedirectURI:  cfg.Auth.OIDC.RedirectURI,
			IssuerURL:    cfg.Auth.OAuthServer,
		}
		oidcService = auth.NewOIDCService(oidcConfig)
	}

	// Setup messaging
	var messagingService messaging.Publisher = &messaging.NoopPublisher{}
	if cfg.Messaging.RabbitMQ.Enabled {
		pub, err := messaging.NewRabbitMQPublisher(
			cfg.Messaging.RabbitMQ.URL,
			cfg.Messaging.RabbitMQ.Exchange,
			cfg.Messaging.RabbitMQ.ExchangeType,
			cfg.Messaging.RabbitMQ.RoutingKey,
		)
		if err == nil {
			messagingService = pub
		} else {
			log.Printf("Warning: RabbitMQ disabled due to init error: %v", err)
		}
	}

	// Setup metrics
	var metricsService *metrics.MetricsService
	if cfg.Metrics.Enabled {
		metricsService = metrics.NewMetricsService()
	}

	// Setup routes
	r := gin.Default()
	routes.SetupRoutes(r, db, storageService, authService, oidcService, messagingService, metricsService, cfg)

	// Start server
	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	log.Printf("Starting Ganje server on %s", addr)
	err = r.Run(addr)
	if err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
