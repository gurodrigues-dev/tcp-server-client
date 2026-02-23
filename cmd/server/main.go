package main

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/gurodrigues-dev/tcp-server-client/config"
	"github.com/gurodrigues-dev/tcp-server-client/internal/messaging"
	mq "github.com/gurodrigues-dev/tcp-server-client/internal/messaging/rabbitmq"
	"github.com/gurodrigues-dev/tcp-server-client/server"
	"github.com/gurodrigues-dev/tcp-server-client/stats"
	"github.com/gurodrigues-dev/tcp-server-client/submission/store"

	_ "github.com/lib/pq"
)

func main() {
	log.Printf("Starting server...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.DatabaseUser,
		cfg.DatabasePassword,
		cfg.DatabaseHost,
		cfg.DatabasePort,
		cfg.DatabaseName,
		cfg.DatabaseSSLMode,
	)

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("failed to open database: %v", err)
	}
	defer db.Close()

	migrations(dsn)

	submissionRepository := store.NewSubmissionRepository(db)

	rabbitMQOpts := &messaging.RabbitMQOpts{
		Host:          cfg.RabbitMQHost,
		Port:          cfg.RabbitMQPort,
		Username:      cfg.RabbitMQUser,
		Password:      cfg.RabbitMQPassword,
		VHost:         cfg.RabbitMQVHost,
		ExchangeName:  cfg.RabbitMQExchange,
		ExchangeType:  cfg.RabbitMQType,
		ExchangeTopic: cfg.RabbitMQTopic,
	}

	rabbitMQProducer, err := mq.NewRabbitMQProducer(rabbitMQOpts)
	if err != nil {
		log.Fatalf("failed to create RabbitMQ producer: %v", err)
	}

	rabbitMQConsumer, err := mq.NewRabbitMQConsumer(rabbitMQOpts)
	if err != nil {
		log.Fatalf("failed to create RabbitMQ consumer: %v", err)
	}

	statsCollector := stats.NewSubmissionCollector(cfg.RabbitMQTopic, cfg.RabbitMQExchange, rabbitMQProducer)
	statsScheduler := stats.NewStatsScheduler(statsCollector, 1*time.Minute)
	statsScheduler.Start()
	defer statsScheduler.Stop()

	authManager := server.NewAuthManager()
	jobManager := server.NewJobManager(authManager)
	submissionManager := server.NewSubmissionManager(authManager)
	requestHandler := server.NewRequestHandler(authManager, jobManager, submissionManager)

	consumerService := stats.NewStatsConsumer(cfg.RabbitMQTopic, rabbitMQConsumer, submissionRepository)
	if err := consumerService.Start(); err != nil {
		log.Printf("Error to start Consumer Service: %v", err)
		log.Printf("Consumer Service will be disabled")
	} else {
		defer consumerService.Stop()
		log.Printf("Consumer Service started successfully")
	}

	tcpServer := server.NewTCPServer(cfg.ServerAddr, requestHandler, authManager)

	jobManager.Start()
	defer func() {
		log.Printf("Stopping JobManager...")
		jobManager.Stop()
	}()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	serverErrChan := make(chan error, 1)
	go func() {
		log.Printf("Server started on address %s", cfg.ServerAddr)
		if err := tcpServer.Start(cfg.ServerAddr); err != nil {
			serverErrChan <- err
		}
	}()

	select {
	case sig := <-sigChan:
		log.Printf("Received signal %v, initiating graceful shutdown...", sig)
	case err := <-serverErrChan:
		log.Printf("Server error: %v", err)
	}

	log.Printf("Stopping server...")
	if err := tcpServer.Stop(); err != nil {
		log.Printf("Error stopping server: %v", err)
	}

	log.Printf("Server stopped")
}

func migrations(databaseURL string) {
	m, err := migrate.New(
		"file://migrations",
		databaseURL,
	)
	if err != nil {
		log.Fatalf("migrate error: %v", err)
	}

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		log.Fatalf("upload migrations error: %v", err)
	}

	log.Println("migrations finished.")
}
