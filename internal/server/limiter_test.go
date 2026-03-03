package server

import (
	"context"
	"testing"
	"time"
)

func TestRateLimiter_Allow(t *testing.T) {
	// Rate: 1 token/sec, Burst: 2
	rl := NewRateLimiter(1, 2, 500*time.Millisecond)
	ip := "127.0.0.1"

	// Initial burst
	if !rl.Allow(ip) {
		t.Error("Expected first call to be allowed")
	}
	if !rl.Allow(ip) {
		t.Error("Expected second call (burst) to be allowed")
	}
	if rl.Allow(ip) {
		t.Error("Expected third call to be blocked (burst exceeded)")
	}

	// Wait for refill (1 token)
	time.Sleep(1100 * time.Millisecond)
	if !rl.Allow(ip) {
		t.Error("Expected call after refill to be allowed")
	}
	if rl.Allow(ip) {
		t.Error("Expected subsequent call to be blocked")
	}
}

func TestRateLimiter_Cleanup(t *testing.T) {
	idleTimeout := 100 * time.Millisecond
	rl := NewRateLimiter(10, 10, idleTimeout)

	rl.Allow("1.1.1.1")
	rl.Allow("2.2.2.2")

	rl.mu.Lock()
	if len(rl.clients) != 2 {
		t.Errorf("Expected 2 clients, got %d", len(rl.clients))
	}
	rl.mu.Unlock()

	// Wait for idle timeout
	time.Sleep(200 * time.Millisecond)

	rl.cleanup()

	rl.mu.Lock()
	if len(rl.clients) != 0 {
		t.Errorf("Expected 0 clients after cleanup, got %d", len(rl.clients))
	}
	rl.mu.Unlock()
}

func TestRateLimiter_StartCleanup(t *testing.T) {
	idleTimeout := 50 * time.Millisecond
	rl := NewRateLimiter(10, 10, idleTimeout)

	rl.Allow("1.1.1.1")

	ctx, cancel := context.WithCancel(context.Background())
	go rl.StartCleanup(ctx, 50*time.Millisecond)

	// Wait for cleanup to run
	time.Sleep(150 * time.Millisecond)
	cancel()

	rl.mu.Lock()
	if len(rl.clients) != 0 {
		t.Errorf("Expected client to be cleaned up by background goroutine, but it remains")
	}
	rl.mu.Unlock()
}
