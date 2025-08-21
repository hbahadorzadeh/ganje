package main

import (
	"log"

	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/server"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	srv := server.New(cfg)
	if err := srv.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
