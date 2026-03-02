package server

import (
	"sync"

	"github.com/prit-motadata/GoServerProject/internal/models"
)

type Metrics struct {
	mu           sync.RWMutex
	TotalLogs    uint64            `json:"total_logs"`
	ErrorLogs    uint64            `json:"error_logs"`
	DroppedLogs  uint64            `json:"dropped_logs"`
	ServiceStats map[string]uint64 `json:"service_stats"`
}

func NewMetrics() *Metrics {
	return &Metrics{
		ServiceStats: make(map[string]uint64),
	}
}

func (m *Metrics) Record(level models.LogLevel, service string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.TotalLogs++
	if level == models.ErrorLevel {
		m.ErrorLogs++
	}
	m.ServiceStats[service]++
}

func (m *Metrics) RecordDrop() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.DroppedLogs++
}

type MetricsSnapshot struct {
	TotalLogs    uint64            `json:"total_logs"`
	ErrorLogs    uint64            `json:"error_logs"`
	DroppedLogs  uint64            `json:"dropped_logs"`
	ServiceStats map[string]uint64 `json:"service_stats"`
}

func (m *Metrics) GetSnapshot() MetricsSnapshot {
	m.mu.RLock()
	defer m.mu.RUnlock()

	// Deep copy service stats
	stats := make(map[string]uint64, len(m.ServiceStats))
	for k, v := range m.ServiceStats {
		stats[k] = v
	}

	return MetricsSnapshot{
		TotalLogs:    m.TotalLogs,
		ErrorLogs:    m.ErrorLogs,
		DroppedLogs:  m.DroppedLogs,
		ServiceStats: stats,
	}
}
