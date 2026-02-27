package main

import (
	"log"
	"os"

	"github.com/prit-motadata/GoServerProject/internal/server"
)

func main() {
	addr := getEnv("SERVER_ADDR", ":8080")

	s := server.New(addr)

	if err := s.Start(); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
