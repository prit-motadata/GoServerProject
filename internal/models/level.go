package models

type LogLevel string

const (
	InfoLevel  LogLevel = "INFO"
	WarnLevel  LogLevel = "WARN"
	ErrorLevel LogLevel = "ERROR"
)

func (l LogLevel) IsValid() bool {
	switch l {
	case InfoLevel, WarnLevel, ErrorLevel:
		return true
	}
	return false
}
