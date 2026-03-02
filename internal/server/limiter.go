package server

import (
	"context"
	"log"
	"sync"
	"time"
)

// ClientLimiter tracks rate limit state for a single IP.
type ClientLimiter struct {
	tokens       float64
	lastRefill   time.Time
	lastActivity time.Time
}

// RateLimiter manages a collection of ClientLimiters.
type RateLimiter struct {
	mu          sync.Mutex
	clients     map[string]*ClientLimiter
	rate        float64       // Tokens per second
	burst       float64       // Max tokens (burst size)
	idleTimeout time.Duration // Time after which a client is removed
}

func NewRateLimiter(rate, burst float64, idleTimeout time.Duration) *RateLimiter {
	return &RateLimiter{
		clients:     make(map[string]*ClientLimiter),
		rate:        rate,
		burst:       burst,
		idleTimeout: idleTimeout,
	}
}

func (rl *RateLimiter) Allow(ip string) bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	client, exists := rl.clients[ip]
	if !exists {
		client = &ClientLimiter{
			tokens:       rl.burst,
			lastRefill:   now,
			lastActivity: now,
		}
		rl.clients[ip] = client
	}

	// Refill tokens based on time elapsed
	elapsed := now.Sub(client.lastRefill).Seconds()
	client.tokens += elapsed * rl.rate
	if client.tokens > rl.burst {
		client.tokens = rl.burst
	}
	client.lastRefill = now
	client.lastActivity = now

	if client.tokens >= 1 {
		client.tokens--
		return true
	}

	return false
}

func (rl *RateLimiter) StartCleanup(ctx context.Context, interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Rate limiter cleanup stopped.")
			return
		case <-ticker.C:
			rl.cleanup()
		}
	}
}

func (rl *RateLimiter) cleanup() {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	removed := 0
	for ip, client := range rl.clients {
		if now.Sub(client.lastActivity) > rl.idleTimeout {
			delete(rl.clients, ip)
			removed++
		}
	}

	if removed > 0 {
		log.Printf("Rate limiter cleanup: removed %d inactive clients\n", removed)
	}
}
