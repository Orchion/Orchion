package logging

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

// Level represents the logging level
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// String returns string representation of the level
func (l Level) String() string {
	switch l {
	case DebugLevel:
		return "debug"
	case InfoLevel:
		return "info"
	case WarnLevel:
		return "warn"
	case ErrorLevel:
		return "error"
	default:
		return "unknown"
	}
}

// LogStreamer defines the interface for streaming log entries
type LogStreamer interface {
	Stream(entry *LogEntry) error
	Close() error
}

// LogEntry represents a structured log entry
type LogEntry struct {
	ID        string
	Timestamp int64
	Level     Level
	Source    string
	Message   string
	Fields    map[string]string
}

// Logger is the main logger interface
type Logger interface {
	Debug(msg string, fields map[string]interface{})
	Info(msg string, fields map[string]interface{})
	Warn(msg string, fields map[string]interface{})
	Error(msg string, fields map[string]interface{})
	WithField(key string, value interface{}) Logger
	WithFields(fields map[string]interface{}) Logger
	SetLevel(level Level)
	SetOutput(w io.Writer)
	SetStreamer(streamer LogStreamer)
	Close()
}

// Config holds logger configuration
type Config struct {
	Level            Level
	Source           string // Component identifier (e.g., "orchestrator", "node-agent:node123")
	OrchestratorAddr string // Address to stream logs to orchestrator (empty to disable streaming)
}

// orchionLogger implements the Logger interface
type orchionLogger struct {
	logger   *logrus.Logger
	source   string
	streamer LogStreamer
	fields   map[string]interface{}
}

// NewLogger creates a new logger with the given configuration
func NewLogger(config Config) Logger {
	logger := &orchionLogger{
		logger: logrus.New(),
		source: config.Source,
		fields: make(map[string]interface{}),
	}

	// Configure logrus
	logger.logger.SetFormatter(&logrus.JSONFormatter{
		TimestampFormat: time.RFC3339,
	})
	logger.logger.SetOutput(os.Stdout)

	// Set level
	switch config.Level {
	case DebugLevel:
		logger.logger.SetLevel(logrus.DebugLevel)
	case InfoLevel:
		logger.logger.SetLevel(logrus.InfoLevel)
	case WarnLevel:
		logger.logger.SetLevel(logrus.WarnLevel)
	case ErrorLevel:
		logger.logger.SetLevel(logrus.ErrorLevel)
	default:
		logger.logger.SetLevel(logrus.InfoLevel)
	}

	return logger
}

// SetStreamer sets the log streamer for this logger
func (l *orchionLogger) SetStreamer(streamer LogStreamer) {
	l.streamer = streamer
}

// log sends a log entry both to local output and streamer
func (l *orchionLogger) log(level Level, msg string, fields map[string]interface{}) {
	// Merge instance fields with provided fields
	allFields := make(map[string]interface{})
	for k, v := range l.fields {
		allFields[k] = v
	}
	for k, v := range fields {
		allFields[k] = v
	}

	// Create logrus entry
	entry := l.logger.WithFields(allFields)

	// Log locally
	switch level {
	case DebugLevel:
		entry.Debug(msg)
	case InfoLevel:
		entry.Info(msg)
	case WarnLevel:
		entry.Warn(msg)
	case ErrorLevel:
		entry.Error(msg)
	}

	// Send to streamer if available
	if l.streamer != nil {
		logEntry := &LogEntry{
			ID:        fmt.Sprintf("%d-%s", time.Now().UnixMilli(), l.source),
			Timestamp: time.Now().UnixMilli(),
			Level:     level,
			Source:    l.source,
			Message:   msg,
			Fields:    l.convertFields(allFields),
		}

		// Send asynchronously (don't block logging on network issues)
		go func() {
			if err := l.streamer.Stream(logEntry); err != nil {
				// Log locally that streaming failed, but don't spam
				l.logger.WithError(err).WithField("source", l.source).Debug("Failed to stream log entry")
			}
		}()
	}
}

// convertFields converts map[string]interface{} to map[string]string
func (l *orchionLogger) convertFields(fields map[string]interface{}) map[string]string {
	result := make(map[string]string)
	for k, v := range fields {
		result[k] = fmt.Sprintf("%v", v)
	}
	return result
}

func (l *orchionLogger) Debug(msg string, fields map[string]interface{}) {
	l.log(DebugLevel, msg, fields)
}

func (l *orchionLogger) Info(msg string, fields map[string]interface{}) {
	l.log(InfoLevel, msg, fields)
}

func (l *orchionLogger) Warn(msg string, fields map[string]interface{}) {
	l.log(WarnLevel, msg, fields)
}

func (l *orchionLogger) Error(msg string, fields map[string]interface{}) {
	l.log(ErrorLevel, msg, fields)
}

func (l *orchionLogger) WithField(key string, value interface{}) Logger {
	newLogger := &orchionLogger{
		logger:   l.logger,
		source:   l.source,
		streamer: l.streamer,
		fields:   make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new field
	newLogger.fields[key] = value

	return newLogger
}

func (l *orchionLogger) WithFields(fields map[string]interface{}) Logger {
	newLogger := &orchionLogger{
		logger:   l.logger,
		source:   l.source,
		streamer: l.streamer,
		fields:   make(map[string]interface{}),
	}

	// Copy existing fields
	for k, v := range l.fields {
		newLogger.fields[k] = v
	}

	// Add new fields
	for k, v := range fields {
		newLogger.fields[k] = v
	}

	return newLogger
}

func (l *orchionLogger) SetLevel(level Level) {
	var logrusLevel logrus.Level
	switch level {
	case DebugLevel:
		logrusLevel = logrus.DebugLevel
	case InfoLevel:
		logrusLevel = logrus.InfoLevel
	case WarnLevel:
		logrusLevel = logrus.WarnLevel
	case ErrorLevel:
		logrusLevel = logrus.ErrorLevel
	default:
		logrusLevel = logrus.InfoLevel
	}
	l.logger.SetLevel(logrusLevel)
}

func (l *orchionLogger) SetOutput(w io.Writer) {
	l.logger.SetOutput(w)
}

func (l *orchionLogger) Close() {
	if l.streamer != nil {
		l.streamer.Close()
	}
}
