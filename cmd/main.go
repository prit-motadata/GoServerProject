package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"github.com/prit-motadata/GoServerProject/internal/server"
)

func main() {
	// Load .env file if it exists
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, using system environment variables")
	}

	cfg := server.ConfigFromEnv()

	// Set up signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := server.New(cfg)

	// Start server in a separate goroutine
	errChan := make(chan error, 1)
	go func() {
		if err := s.Start(); err != nil {
			errChan <- err
		}
	}()

	// Wait for termination signal or server error
	select {
	case <-ctx.Done():
		log.Println("Received termination signal, starting graceful shutdown...")
	case err := <-errChan:
		log.Fatalf("Server failed to start: %+v", err)
	}

	// Create context with timeout for shutdown
	shutdownCtx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	if err := s.Shutdown(shutdownCtx); err != nil {
		log.Printf("Graceful shutdown failed: %+v", err)
		os.Exit(1)
	}

	log.Println("Graceful shutdown finished successfully.")
}
