package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/prit-motadata/GoServerProject/internal/models"
)

type BackpressureStrategy int

const (
	StrategyBlock BackpressureStrategy = iota
	StrategyDrop
	StrategyReject
)

func (s BackpressureStrategy) String() string {
	switch s {
	case StrategyDrop:
		return "DROP"
	case StrategyReject:
		return "REJECT"
	default:
		return "BLOCK"
	}
}

type Server struct {
	config        *Config
	httpServer    *http.Server
	logCh         chan models.Log
	metrics       *Metrics
	limiter       *RateLimiter
	limiterCancel context.CancelFunc
	workerPool    *WorkerPool
}

func New(cfg *Config) *Server {
	if cfg == nil {
		cfg = DefaultConfig()
	}

	logCh := make(chan models.Log, cfg.QueueSize)
	metrics := NewMetrics()

	s := &Server{
		config:     cfg,
		logCh:      logCh,
		metrics:    metrics,
		limiter:    NewRateLimiter(cfg.RateLimit, cfg.RateBurst, cfg.IdleTimeout),
		workerPool: NewWorkerPool(logCh, metrics, cfg.WorkerCount),
	}

	mux := http.NewServeMux()
	mux.Handle("/health", s.rateLimitMiddleware(http.HandlerFunc(s.healthHandler)))
	mux.Handle("/logs", s.rateLimitMiddleware(http.HandlerFunc(s.logHandler)))
	mux.Handle("/metrics", s.rateLimitMiddleware(http.HandlerFunc(s.metricsHandler)))

	s.httpServer = &http.Server{
		Addr:    cfg.Addr,
		Handler: mux,
	}

	s.workerPool.Start()

	// Start limiter cleanup
	ctx, cancel := context.WithCancel(context.Background())
	s.limiterCancel = cancel
	go s.limiter.StartCleanup(ctx, 30*time.Second)

	return s
}

func (s *Server) Start() error {
	log.Printf("server starting on %s (Backpressure: %s)", s.config.Addr, s.config.BackpressureStrategy)
	if err := s.httpServer.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	log.Println("Shutting down HTTP server...")
	err := s.httpServer.Shutdown(ctx)

	log.Println("Closing log channel...")
	close(s.logCh)

	log.Println("Waiting for workers to finish...")
	s.workerPool.Stop()
	log.Println("Workers finished.")

	if s.limiterCancel != nil {
		s.limiterCancel()
	}

	return err
}
