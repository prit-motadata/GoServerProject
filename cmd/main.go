package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"strconv"
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

	addr := getEnv("SERVER_ADDR", ":8080")
	strategyStr := getEnv("BACKPRESSURE_STRATEGY", "block")

	var strategy server.BackpressureStrategy
	switch strategyStr {
	case "drop":
		strategy = server.StrategyDrop
	case "reject":
		strategy = server.StrategyReject
	default:
		strategy = server.StrategyBlock
	}

	rateStr := getEnv("RATE_LIMIT", "2")
	burstStr := getEnv("RATE_BURST", "5")

	rate, _ := strconv.ParseFloat(rateStr, 64)
	if rate == 0 {
		rate = 2
	}
	burst, _ := strconv.ParseFloat(burstStr, 64)
	if burst == 0 {
		burst = 5
	}

	// Set up signal handling
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	s := server.New(addr, strategy, rate, burst)

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

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
