package server

import (
	"testing"

	"github.com/prit-motadata/GoServerProject/internal/models"
)

func TestMetrics_Record(t *testing.T) {
	m := NewMetrics()

	m.Record(models.InfoLevel, "service-a")
	m.Record(models.ErrorLevel, "service-a")
	m.Record(models.WarnLevel, "service-b")

	snapshot := m.GetSnapshot()

	if snapshot.TotalLogs != 3 {
		t.Errorf("Expected 3 total logs, got %d", snapshot.TotalLogs)
	}
	if snapshot.ErrorLogs != 1 {
		t.Errorf("Expected 1 error log, got %d", snapshot.ErrorLogs)
	}
	if snapshot.ServiceStats["service-a"] != 2 {
		t.Errorf("Expected 2 logs for service-a, got %d", snapshot.ServiceStats["service-a"])
	}
	if snapshot.ServiceStats["service-b"] != 1 {
		t.Errorf("Expected 1 log for service-b, got %d", snapshot.ServiceStats["service-b"])
	}
}

func TestMetrics_RecordDrop(t *testing.T) {
	m := NewMetrics()

	m.RecordDrop()
	m.RecordDrop()

	snapshot := m.GetSnapshot()

	if snapshot.DroppedLogs != 2 {
		t.Errorf("Expected 2 dropped logs, got %d", snapshot.DroppedLogs)
	}
}

func TestMetrics_GetSnapshot_DeepCopy(t *testing.T) {
	m := NewMetrics()
	m.Record(models.InfoLevel, "service-a")

	snapshot1 := m.GetSnapshot()

	// Modify original
	m.Record(models.InfoLevel, "service-a")

	snapshot2 := m.GetSnapshot()

	if snapshot1.ServiceStats["service-a"] != 1 {
		t.Errorf("Expected snapshot1 to have 1 log for service-a despite original modification, got %d", snapshot1.ServiceStats["service-a"])
	}
	if snapshot2.ServiceStats["service-a"] != 2 {
		t.Errorf("Expected snapshot2 to have 2 logs for service-a, got %d", snapshot2.ServiceStats["service-a"])
	}
}
