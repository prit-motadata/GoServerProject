package server

import (
	"context"
	"errors"
	"log"
	"net/http"
	"net/http/pprof"
	"runtime"
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
	pprofServer   *http.Server
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

	if s.config.EnablePprof {
		pprofMux := http.NewServeMux()
		pprofMux.HandleFunc("/debug/pprof/", pprof.Index)
		pprofMux.HandleFunc("/debug/pprof/cmdline", pprof.Cmdline)
		pprofMux.HandleFunc("/debug/pprof/profile", pprof.Profile)
		pprofMux.HandleFunc("/debug/pprof/symbol", pprof.Symbol)
		pprofMux.HandleFunc("/debug/pprof/trace", pprof.Trace)

		// runtime profiles
		pprofMux.Handle("/debug/pprof/goroutine", pprof.Handler("goroutine"))
		pprofMux.Handle("/debug/pprof/heap", pprof.Handler("heap"))
		pprofMux.Handle("/debug/pprof/allocs", pprof.Handler("allocs"))
		pprofMux.Handle("/debug/pprof/block", pprof.Handler("block"))
		pprofMux.Handle("/debug/pprof/mutex", pprof.Handler("mutex"))
		pprofMux.Handle("/debug/pprof/threadcreate", pprof.Handler("threadcreate"))

		// enable extra profiling
		runtime.SetBlockProfileRate(1)
		runtime.SetMutexProfileFraction(1)

		s.pprofServer = &http.Server{
			Addr:    cfg.PprofAddr,
			Handler: pprofMux,
		}
	}

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

	if s.pprofServer != nil {
		log.Printf("pprof server starting on %s", s.config.PprofAddr)
		go func() {
			if err := s.pprofServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Printf("pprof server error: %v", err)
			}
		}()
	}

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

	if s.pprofServer != nil {
		log.Println("Shutting down pprof server...")
		if pprofErr := s.pprofServer.Shutdown(ctx); pprofErr != nil {
			log.Printf("pprof server shutdown error: %v", pprofErr)
		}
	}

	return err
}
