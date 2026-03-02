package models

import (
	"errors"
	"strings"
	"time"
)

type Log struct {
	Service   string    `json:"service"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Timestamp time.Time `json:"timestamp"`
}

func (l *Log) Validate() error {
	if strings.TrimSpace(l.Service) == "" {
		return errors.New("service is required")
	}
	if strings.TrimSpace(l.Message) == "" {
		return errors.New("message is required")
	}

	if !l.Level.IsValid() {
		return errors.New("invalid level")
	}

	if l.Timestamp.IsZero() {
		return errors.New("timestamp is required")
	}

	return nil
}
