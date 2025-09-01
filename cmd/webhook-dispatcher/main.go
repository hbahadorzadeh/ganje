package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/hbahadorzadeh/ganje/internal/config"
	"github.com/hbahadorzadeh/ganje/internal/database"
	"github.com/hbahadorzadeh/ganje/internal/messaging"
	"github.com/hbahadorzadeh/ganje/internal/webhook"
)

func main() {
	configFile := flag.String("config", "config.yaml", "Path to config file")
	queue := flag.String("queue", "", "RabbitMQ queue name (optional)")
	flag.Parse()

	cfg, err := config.Load(*configFile)
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	if !cfg.Webhook.Enabled {
		log.Println("webhook dispatcher disabled by config; exiting")
		return
	}

	db, err := database.New(cfg.Database.Driver, cfg.Database.GetConnectionString())
	if err != nil {
		log.Fatalf("database init: %v", err)
	}
	defer func() { _ = db.Close() }()

	wc := cfg.Webhook
	dcfg := webhook.DispatcherConfig{
		MaxRetries:     wc.MaxRetries,
		InitialBackoff: time.Duration(wc.InitialBackoffMs) * time.Millisecond,
		MaxBackoff:     time.Duration(wc.MaxBackoffMs) * time.Millisecond,
		HTTPTimeout:    time.Duration(wc.HTTPTimeoutMs) * time.Millisecond,
	}
	d := webhook.NewDispatcher(db, dcfg)
	workers := wc.Workers
	if workers <= 0 { workers = 2 }
	d.Start(workers)
	defer func() { _ = d.Stop(context.Background()) }()

	mq := cfg.Messaging.RabbitMQ
	if !mq.Enabled {
		log.Fatalf("messaging.rabbitmq is disabled; dispatcher requires event stream")
	}
	ccfg := messaging.ConsumerConfig{
		URL:          mq.URL,
		Exchange:     mq.Exchange,
		ExchangeType: mq.ExchangeType,
		RoutingKey:   mq.RoutingKey,
		Queue:        *queue,
	}
	consumer, err := messaging.NewEventConsumer(ccfg)
	if err != nil {
		log.Fatalf("consumer init: %v", err)
	}
	defer func() { _ = consumer.Close() }()

	// Signal handling
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)

	log.Printf("webhook-dispatcher started; exchange=%s routingKey=%s workers=%d", mq.Exchange, mq.RoutingKey, workers)

	for {
		select {
		case sig := <-sigCh:
			log.Printf("received signal: %v; shutting down", sig)
			return
		case msg, ok := <-consumer.Messages():
			if !ok {
				log.Println("consumer channel closed; exiting")
				return
			}
			evt, err := messaging.DecodeEvent(msg)
			if err != nil {
				log.Printf("decode event error: %v", err)
				continue
			}
			// Basic validation
			if evt.Repository == "" || evt.Type == "" {
				log.Printf("skip invalid event: %+v", evt)
				continue
			}
			d.Enqueue(evt)
		}
	}
}
