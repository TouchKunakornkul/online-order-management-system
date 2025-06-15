package logger

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// LogLevel represents the severity level of a log entry
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARN
	ERROR
	FATAL
)

func (l LogLevel) String() string {
	switch l {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO"
	case WARN:
		return "WARN"
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

// Logger represents a structured logger
type Logger struct {
	level      LogLevel
	service    string
	version    string
	withFields map[string]interface{}
}

// LogEntry represents a single log entry
type LogEntry struct {
	Timestamp string                 `json:"timestamp"`
	Level     string                 `json:"level"`
	Service   string                 `json:"service"`
	Version   string                 `json:"version"`
	Message   string                 `json:"message"`
	Fields    map[string]interface{} `json:"fields,omitempty"`
	Caller    string                 `json:"caller,omitempty"`
	Error     string                 `json:"error,omitempty"`
}

// New creates a new logger instance
func New(service, version string) *Logger {
	level := INFO
	if levelStr := os.Getenv("LOG_LEVEL"); levelStr != "" {
		switch strings.ToUpper(levelStr) {
		case "DEBUG":
			level = DEBUG
		case "INFO":
			level = INFO
		case "WARN", "WARNING":
			level = WARN
		case "ERROR":
			level = ERROR
		case "FATAL":
			level = FATAL
		}
	}

	return &Logger{
		level:      level,
		service:    service,
		version:    version,
		withFields: make(map[string]interface{}),
	}
}

// WithFields returns a new logger with additional fields
func (l *Logger) WithFields(fields map[string]interface{}) *Logger {
	newLogger := &Logger{
		level:      l.level,
		service:    l.service,
		version:    l.version,
		withFields: make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.withFields {
		newLogger.withFields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.withFields[k] = v
	}

	return newLogger
}

// WithField returns a new logger with an additional field
func (l *Logger) WithField(key string, value interface{}) *Logger {
	return l.WithFields(map[string]interface{}{key: value})
}

// WithError returns a new logger with error information
func (l *Logger) WithError(err error) *Logger {
	if err == nil {
		return l
	}
	return l.WithField("error", err.Error())
}

// WithContext returns a new logger with context information
func (l *Logger) WithContext(ctx context.Context) *Logger {
	// Extract trace ID or other context values if available
	if traceID := ctx.Value("trace_id"); traceID != nil {
		return l.WithField("trace_id", traceID)
	}
	return l
}

// getCaller returns the file and line number of the caller
func getCaller(skip int) string {
	_, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown"
	}
	// Get just the filename, not the full path
	parts := strings.Split(file, "/")
	if len(parts) > 0 {
		file = parts[len(parts)-1]
	}
	return fmt.Sprintf("%s:%d", file, line)
}

// log outputs a log entry at the specified level
func (l *Logger) log(level LogLevel, msg string, err error) {
	if level < l.level {
		return
	}

	entry := LogEntry{
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Level:     level.String(),
		Service:   l.service,
		Version:   l.version,
		Message:   msg,
		Fields:    l.withFields,
		Caller:    getCaller(3), // Skip log, Debug/Info/Warn/Error, and caller
	}

	if err != nil {
		entry.Error = err.Error()
	}

	// JSON output for structured logging
	jsonBytes, jsonErr := json.Marshal(entry)
	if jsonErr != nil {
		log.Printf("Failed to marshal log entry: %v", jsonErr)
		return
	}

	log.Println(string(jsonBytes))

	// Exit for fatal logs
	if level == FATAL {
		os.Exit(1)
	}
}

// Debug logs a debug message
func (l *Logger) Debug(msg string) {
	l.log(DEBUG, msg, nil)
}

// Debugf logs a formatted debug message
func (l *Logger) Debugf(format string, args ...interface{}) {
	l.log(DEBUG, fmt.Sprintf(format, args...), nil)
}

// Info logs an info message
func (l *Logger) Info(msg string) {
	l.log(INFO, msg, nil)
}

// Infof logs a formatted info message
func (l *Logger) Infof(format string, args ...interface{}) {
	l.log(INFO, fmt.Sprintf(format, args...), nil)
}

// Warn logs a warning message
func (l *Logger) Warn(msg string) {
	l.log(WARN, msg, nil)
}

// Warnf logs a formatted warning message
func (l *Logger) Warnf(format string, args ...interface{}) {
	l.log(WARN, fmt.Sprintf(format, args...), nil)
}

// Error logs an error message
func (l *Logger) Error(msg string) {
	l.log(ERROR, msg, nil)
}

// Errorf logs a formatted error message
func (l *Logger) Errorf(format string, args ...interface{}) {
	l.log(ERROR, fmt.Sprintf(format, args...), nil)
}

// ErrorWithErr logs an error message with error details
func (l *Logger) ErrorWithErr(msg string, err error) {
	l.log(ERROR, msg, err)
}

// Fatal logs a fatal message and exits
func (l *Logger) Fatal(msg string) {
	l.log(FATAL, msg, nil)
}

// Fatalf logs a formatted fatal message and exits
func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.log(FATAL, fmt.Sprintf(format, args...), nil)
}

// FatalWithErr logs a fatal message with error details and exits
func (l *Logger) FatalWithErr(msg string, err error) {
	l.log(FATAL, msg, err)
}

// Default logger instance
var defaultLogger = New("order-management", "1.0.0")

// Package-level functions for convenience
func Debug(msg string) {
	defaultLogger.Debug(msg)
}

func Debugf(format string, args ...interface{}) {
	defaultLogger.Debugf(format, args...)
}

func Info(msg string) {
	defaultLogger.Info(msg)
}

func Infof(format string, args ...interface{}) {
	defaultLogger.Infof(format, args...)
}

func Warn(msg string) {
	defaultLogger.Warn(msg)
}

func Warnf(format string, args ...interface{}) {
	defaultLogger.Warnf(format, args...)
}

func Error(msg string) {
	defaultLogger.Error(msg)
}

func Errorf(format string, args ...interface{}) {
	defaultLogger.Errorf(format, args...)
}

func ErrorWithErr(msg string, err error) {
	defaultLogger.ErrorWithErr(msg, err)
}

func Fatal(msg string) {
	defaultLogger.Fatal(msg)
}

func Fatalf(format string, args ...interface{}) {
	defaultLogger.Fatalf(format, args...)
}

func FatalWithErr(msg string, err error) {
	defaultLogger.FatalWithErr(msg, err)
}
