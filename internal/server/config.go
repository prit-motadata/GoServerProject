package server

import (
	"os"
	"strconv"
	"time"
)

// Config holds the server configuration.
type Config struct {
	Addr                 string
	BackpressureStrategy BackpressureStrategy
	RateLimit            float64
	RateBurst            float64
	IdleTimeout          time.Duration
	WorkerCount          int
	MaxBodySize          int64
	QueueSize            int
}

// DefaultConfig returns a configuration with default values.
func DefaultConfig() *Config {
	return &Config{
		Addr:                 ":8080",
		BackpressureStrategy: StrategyBlock,
		RateLimit:            2.0,
		RateBurst:            5.0,
		IdleTimeout:          1 * time.Minute,
		WorkerCount:          3,
		MaxBodySize:          1 << 20, // 1MB
		QueueSize:            5,
	}
}

// ConfigFromEnv loads configuration from environment variables.
func ConfigFromEnv() *Config {
	cfg := DefaultConfig()

	if addr := os.Getenv("SERVER_ADDR"); addr != "" {
		cfg.Addr = addr
	}

	if strategy := os.Getenv("BACKPRESSURE_STRATEGY"); strategy != "" {
		switch strategy {
		case "drop":
			cfg.BackpressureStrategy = StrategyDrop
		case "reject":
			cfg.BackpressureStrategy = StrategyReject
		default:
			cfg.BackpressureStrategy = StrategyBlock
		}
	}

	if rate := os.Getenv("RATE_LIMIT"); rate != "" {
		if r, err := strconv.ParseFloat(rate, 64); err == nil && r > 0 {
			cfg.RateLimit = r
		}
	}

	if burst := os.Getenv("RATE_BURST"); burst != "" {
		if b, err := strconv.ParseFloat(burst, 64); err == nil && b > 0 {
			cfg.RateBurst = b
		}
	}

	if workers := os.Getenv("WORKER_COUNT"); workers != "" {
		if w, err := strconv.Atoi(workers); err == nil && w > 0 {
			cfg.WorkerCount = w
		}
	}

	return cfg
}
