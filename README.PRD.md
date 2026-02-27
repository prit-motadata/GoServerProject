# 🚀 Concurrent Log Processing & Rate-Limited API Service (Go)

## Overview

This project is a production-grade concurrent log ingestion and
analytics service written in pure Go.

It is intentionally designed to force you into real-world edge cases
involving:

-   Goroutines & scheduling
-   Channel edge cases
-   Context cancellation
-   Graceful shutdown
-   Worker pools
-   Backpressure handling
-   Rate limiting
-   Race conditions
-   Slice memory leaks
-   Interface design traps
-   Atomic vs Mutex tradeoffs
-   Panic recovery
-   Testing concurrent systems

By the end of this project, you should be comfortable building backend
systems similar in architecture to observability platforms.

------------------------------------------------------------------------

# 🏗 High-Level Architecture

                       ┌─────────────────────┐
                       │        Client       │
                       └──────────┬──────────┘
                                  │ HTTP
                                  ▼
                       ┌─────────────────────┐
                       │   HTTP Server       │
                       │  (net/http)         │
                       └──────────┬──────────┘
                                  │
                                  ▼
                       ┌─────────────────────┐
                       │   Rate Limiter      │
                       │ (Token Bucket/IP)   │
                       └──────────┬──────────┘
                                  │
                                  ▼
                       ┌─────────────────────┐
                       │ Buffered Channel    │
                       │ (Backpressure)      │
                       └──────────┬──────────┘
                                  │
                                  ▼
                       ┌─────────────────────┐
                       │    Worker Pool      │
                       │  (Fixed/Dynamic)    │
                       └──────────┬──────────┘
                                  │
                                  ▼
                       ┌─────────────────────┐
                       │    Aggregator       │
                       │ (Mutex / Atomic)    │
                       └──────────┬──────────┘
                                  │
                                  ▼
                       ┌─────────────────────┐
                       │  Metrics Endpoint   │
                       └─────────────────────┘

------------------------------------------------------------------------

# 📦 Features To Implement

## 1️⃣ Log Ingestion Endpoint

POST `/logs`

-   Accept JSON logs
-   Validate input
-   Enforce rate limiting
-   Apply backpressure when queue is full

------------------------------------------------------------------------

## 2️⃣ Rate Limiting (Per Client/IP)

Implement:

-   Token bucket algorithm
-   Concurrent-safe storage
-   Expired client cleanup
-   Configurable limits

### Edge Cases You'll Face

-   Concurrent map writes
-   Memory leaks from stale clients
-   Lock contention
-   Atomic vs Mutex tradeoff

------------------------------------------------------------------------

## 3️⃣ Buffered Channel (Backpressure Layer)

Logs should be pushed into a buffered channel.

You must decide:

-   What happens when channel is full?
    -   Block?
    -   Drop?
    -   Return 429?
-   Who closes the channel?
-   How do you avoid deadlock?

### Edge Cases

-   Blocking sends
-   Goroutine leaks
-   Channel close panic
-   Select with default (non-blocking behavior)

------------------------------------------------------------------------

## 4️⃣ Worker Pool

Implement:

-   Fixed worker pool
-   Graceful drain on shutdown
-   Panic recovery inside workers
-   Context-aware processing

### Must Handle

-   Worker panic recovery
-   Shutdown during processing
-   Draining queue safely
-   Ensuring no task loss

------------------------------------------------------------------------

## 5️⃣ Aggregator (Metrics Engine)

Track:

-   Total logs
-   Error logs
-   Logs per service
-   Processing latency

Implement using:

-   sync.Mutex
-   sync.RWMutex
-   atomic operations

### Learnings

-   Data races
-   go run -race
-   Lock granularity
-   False sharing
-   Struct alignment

------------------------------------------------------------------------

## 6️⃣ Graceful Shutdown

Handle:

-   SIGINT / SIGTERM
-   http.Server.Shutdown
-   Context cancellation
-   Worker pool drain
-   Channel closing exactly once

### Classic Pitfalls

-   Closing channel multiple times
-   Forgetting to stop background goroutines
-   Leaked goroutines on shutdown

------------------------------------------------------------------------

## 7️⃣ Memory Safety

You must intentionally handle:

-   Large slice retention
-   Batch processing memory traps
-   Avoiding keeping references to large arrays

Example issue:

Retaining 10 bytes from 1MB slice keeps whole 1MB alive.

------------------------------------------------------------------------

## 8️⃣ Interface Design

Define:

``` go
type Processor interface {
    Process(ctx context.Context, log Log) error
}
```

Implement:

-   JSON processor
-   Filter processor
-   Mock processor for tests

### Traps

-   Nil interface vs nil pointer
-   Pointer vs value receiver
-   Method sets and interface satisfaction

------------------------------------------------------------------------

## 9️⃣ Testing Requirements

You must write tests for:

-   Concurrent ingestion
-   Rate limiting under load
-   Shutdown during heavy traffic
-   Worker panic recovery
-   Backpressure scenario
-   Race detection

Use:

-   httptest
-   sync.WaitGroup
-   Table-driven tests
-   -race flag

------------------------------------------------------------------------

# 🧠 What This Project Forces You To Learn

  Area              What You'll Master
  ----------------- -------------------------------------
  Concurrency       Goroutines, scheduling, leaks
  Channels          Buffered vs unbuffered, close rules
  Synchronization   Mutex vs Atomic
  Memory            Slice backing array traps
  Architecture      Backpressure & worker pools
  Context           Cancellation propagation
  OS Signals        Graceful shutdown
  Performance       Contention & profiling
  Testing           Deterministic concurrent testing
  Debugging         Race detector & panic recovery

------------------------------------------------------------------------

# 🔬 Advanced Extensions (Optional but Recommended)

-   Prometheus metrics endpoint
-   pprof profiling
-   WAL (Write-Ahead Log)
-   Config hot reload
-   Dynamic worker scaling
-   Circuit breaker
-   Retry with exponential backoff

------------------------------------------------------------------------

# 📂 Suggested Project Structure

    cmd/
        main.go

    internal/
        server/
        limiter/
        worker/
        aggregator/
        processor/
        shutdown/

    pkg/
        models/

    configs/
        config.go

    tests/
        integration/

------------------------------------------------------------------------

# 🎯 Final Outcome

When complete, you should confidently explain:

-   Why maps are not thread-safe
-   When to use channels vs mutex
-   How Go scheduler works (G-M-P model)
-   How to prevent goroutine leaks
-   How to implement graceful shutdown
-   How to debug race conditions
-   How backpressure protects systems
-   How to design interfaces properly

This project transforms you from "Go syntax learner" into a backend
engineer who understands system behavior under load.

------------------------------------------------------------------------

# 🏁 Goal

Build it as if you are deploying to production.

No frameworks. No shortcuts.

Pure Go.
