package server

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/prit-motadata/GoServerProject/internal/models"
)

const (
	maxBodySize = 1 << 20 // 1MB
	queueSize   = 5
)

type Server struct {
	httpServer *http.Server
	logCh      chan models.Log
	metrics    *Metrics

	workerCount int
	wg          sync.WaitGroup
}

func (s *Server) worker(id int) {
	log.Printf("worker %d started\n", id)

	defer func() {
		if r := recover(); r != nil {
			log.Printf("worker %d recovered from panic: %v\n", id, r)
		}
		log.Printf("worker %d stopped\n", id)
	}()

	for logEntry := range s.logCh {

		// Simulate occasional panic
		if logEntry.Message == "panic" {
			panic("simulated worker panic")
		}

		time.Sleep(2 * time.Second)

		s.metrics.Record(logEntry.Level, logEntry.Service)

		log.Printf("worker %d processed: %+v\n", id, logEntry)
	}
}

func (s *Server) startWorkers() {
	for i := 0; i < s.workerCount; i++ {
		s.wg.Add(1)

		go func(id int) {
			defer s.wg.Done()
			s.worker(id)
		}(i)
	}
}

func New(addr string) *Server {
	mux := http.NewServeMux()
	s := &Server{
		logCh:       make(chan models.Log, queueSize),
		metrics:     NewMetrics(),
		workerCount: 3,
	}

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/logs", s.logHandler)
	mux.HandleFunc("/metrics", s.metricsHandler)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	s.startWorkers()

	return s
}

func (s *Server) Start() error {
	log.Printf("server starting on %s", s.httpServer.Addr)
	if err := s.httpServer.ListenAndServe(); err != http.ErrServerClosed {
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
	s.wg.Wait()
	log.Println("Workers finished.")

	return err
}

func (s *Server) healthHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("ok"))
}

func (s *Server) logHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Prevent large payload attacks
	r.Body = http.MaxBytesReader(w, r.Body, maxBodySize)

	defer func() {
		// ensure body fully drained to allow connection reuse
		_, _ = io.Copy(io.Discard, r.Body)
		err := r.Body.Close()
		if err != nil {
			return
		}
	}()

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	var logEntry models.Log
	if err := decoder.Decode(&logEntry); err != nil {
		http.Error(w, "invalid JSON payload", http.StatusBadRequest)
		return
	}

	if err := logEntry.Validate(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Push log entry into channel
	s.logCh <- logEntry

	w.WriteHeader(http.StatusAccepted)
}

func (s *Server) metricsHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	metrics := s.metrics.GetSnapshot()

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(metrics); err != nil {
		http.Error(w, "error encoding metrics", http.StatusInternalServerError)
		return
	}
}
