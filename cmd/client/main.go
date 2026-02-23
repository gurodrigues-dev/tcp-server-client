package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/gurodrigues-dev/tcp-server-client/client"
	"github.com/gurodrigues-dev/tcp-server-client/config"
)

func main() {
	log.Printf("Starting client...")

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	var username string
	flag.StringVar(&username, "username", username, "Username for authentication")
	flag.Parse()

	if username == "" {
		log.Fatal("Error: username is required. Provide via -username flag or CLIENT_USERNAME environment variable")
	}

	log.Printf("Starting client with username: %s", username)

	tcpClient := client.NewTCPClient(cfg.ServerAddr)

	log.Printf("Connecting to server at %s...", cfg.ServerAddr)
	if err := tcpClient.Connect(); err != nil {
		log.Fatalf("Error connecting to server: %v", err)
	}
	log.Printf("Successfully connected to server")

	defer func() {
		log.Printf("Disconnecting from server...")
		tcpClient.Disconnect()
		log.Printf("Disconnected")
	}()

	authResponse, err := tcpClient.Authorize(username)
	if err != nil {
		log.Fatalf("Error during authorization: %v", err)
	}

	if authResponse.Error != "" {
		log.Fatalf("Authentication failed: %s", authResponse.Error)
	}

	log.Printf("Authentication successful")

	log.Printf("Starting job handler...")
	tcpClient.StartJobHandler()
	log.Printf("Job handler started, waiting for tasks...")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	sig := <-sigChan
	log.Printf("Received signal: %v. Shutting down gracefully...", sig)
}
