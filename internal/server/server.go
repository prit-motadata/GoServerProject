package server

import (
	"encoding/json"
	"io"
	"log"
	"net/http"

	"github.com/prit-motadata/GoServerProject/internal/models"
)

const maxBodySize = 1 << 20 // 1MB

type Server struct {
	httpServer *http.Server
}

func New(addr string) *Server {
	mux := http.NewServeMux()
	s := &Server{}

	mux.HandleFunc("/health", s.healthHandler)
	mux.HandleFunc("/logs", s.logHandler)

	s.httpServer = &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	return s
}

func (s *Server) Start() error {
	log.Printf("server starting on %s", s.httpServer.Addr)
	return s.httpServer.ListenAndServe()
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

	log.Printf("received log: %+v\n", logEntry)

	w.WriteHeader(http.StatusAccepted)
}
