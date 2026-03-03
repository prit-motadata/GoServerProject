package server

import (
	"encoding/json"
	"io"
	"net"
	"net/http"

	"github.com/prit-motadata/GoServerProject/internal/models"
)

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
	r.Body = http.MaxBytesReader(w, r.Body, s.config.MaxBodySize)

	defer func() {
		// ensure body fully drained to allow connection reuse
		_, _ = io.Copy(io.Discard, r.Body)
		_ = r.Body.Close()
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

	// Push log entry into channel based on strategy
	switch s.config.BackpressureStrategy {
	case StrategyDrop:
		select {
		case s.logCh <- logEntry:
			w.WriteHeader(http.StatusAccepted)
		default:
			s.metrics.RecordDrop()
			w.WriteHeader(http.StatusAccepted) // Silently drop
		}
	case StrategyReject:
		select {
		case s.logCh <- logEntry:
			w.WriteHeader(http.StatusAccepted)
		default:
			s.metrics.RecordDrop()
			http.Error(w, "server busy, try again later", http.StatusTooManyRequests)
		}
	default: // StrategyBlock
		s.logCh <- logEntry
		w.WriteHeader(http.StatusAccepted)
	}
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

func (s *Server) rateLimitMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ip, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			ip = r.RemoteAddr
		}

		if !s.limiter.Allow(ip) {
			http.Error(w, "rate limit exceeded", http.StatusTooManyRequests)
			return
		}

		next.ServeHTTP(w, r)
	})
}
