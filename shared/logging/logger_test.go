package logging

import (
	"bytes"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockLogStreamer is a mock implementation of LogStreamer
type MockLogStreamer struct {
	mock.Mock
}

func (m *MockLogStreamer) Stream(entry *LogEntry) error {
	args := m.Called(entry)
	return args.Error(0)
}

func (m *MockLogStreamer) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestLevel_String(t *testing.T) {
	testCases := []struct {
		level    Level
		expected string
	}{
		{DebugLevel, "debug"},
		{InfoLevel, "info"},
		{WarnLevel, "warn"},
		{ErrorLevel, "error"},
		{Level(999), "unknown"},
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.level.String())
		})
	}
}

func TestNewLogger(t *testing.T) {
	config := Config{
		Level:  InfoLevel,
		Source: "test-component",
	}

	logger := NewLogger(config)

	assert.NotNil(t, logger)

	// Verify it's the correct type
	orchionLogger, ok := logger.(*orchionLogger)
	assert.True(t, ok)
	assert.Equal(t, "test-component", orchionLogger.source)
	assert.NotNil(t, orchionLogger.fields)
	assert.Empty(t, orchionLogger.fields)
}

func TestNewLogger_LevelConfiguration(t *testing.T) {
	testCases := []struct {
		configLevel Level
		expected    logrus.Level
	}{
		{DebugLevel, logrus.DebugLevel},
		{InfoLevel, logrus.InfoLevel},
		{WarnLevel, logrus.WarnLevel},
		{ErrorLevel, logrus.ErrorLevel},
		{Level(999), logrus.InfoLevel}, // Default to Info
	}

	for _, tc := range testCases {
		t.Run(tc.configLevel.String(), func(t *testing.T) {
			config := Config{
				Level:  tc.configLevel,
				Source: "test",
			}

			logger := NewLogger(config)
			orchionLogger := logger.(*orchionLogger)

			assert.Equal(t, tc.expected, orchionLogger.logger.Level)
		})
	}
}

func TestOrchionLogger_SetStreamer(t *testing.T) {
	logger := NewLogger(Config{Source: "test"})
	orchionLogger := logger.(*orchionLogger)

	mockStreamer := &MockLogStreamer{}
	logger.SetStreamer(mockStreamer)

	assert.Equal(t, mockStreamer, orchionLogger.streamer)
}

func TestOrchionLogger_Debug(t *testing.T) {
	var buf bytes.Buffer
	config := Config{
		Level:  DebugLevel,
		Source: "test-component",
	}

	logger := NewLogger(config)
	logger.SetOutput(&buf)

	fields := map[string]interface{}{
		"key1": "value1",
		"key2": 42,
	}

	logger.Debug("test debug message", fields)

	output := buf.String()
	assert.Contains(t, output, "test debug message")
	assert.Contains(t, output, "key1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "key2")
	assert.Contains(t, output, "42")
}

func TestOrchionLogger_Info(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})
	logger.SetOutput(&buf)

	logger.Info("test info message", nil)

	output := buf.String()
	assert.Contains(t, output, "test info message")
	assert.Contains(t, output, "info")
}

func TestOrchionLogger_Warn(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{Level: WarnLevel, Source: "test"})
	logger.SetOutput(&buf)

	logger.Warn("test warn message", nil)

	output := buf.String()
	assert.Contains(t, output, "test warn message")
	assert.Contains(t, output, "warning")
}

func TestOrchionLogger_Error(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{Level: ErrorLevel, Source: "test"})
	logger.SetOutput(&buf)

	logger.Error("test error message", nil)

	output := buf.String()
	assert.Contains(t, output, "test error message")
	assert.Contains(t, output, "error")
}

func TestOrchionLogger_Debug_WithStreamer(t *testing.T) {
	var buf bytes.Buffer
	mockStreamer := &MockLogStreamer{}

	config := Config{
		Level:  DebugLevel,
		Source: "test-streamer",
	}

	logger := NewLogger(config)
	logger.SetOutput(&buf)
	logger.SetStreamer(mockStreamer)

	fields := map[string]interface{}{
		"test": "value",
	}

	// Expect the streamer to be called
	mockStreamer.On("Stream", mock.MatchedBy(func(entry *LogEntry) bool {
		return entry.Level == DebugLevel &&
			entry.Source == "test-streamer" &&
			entry.Message == "streaming debug message" &&
			entry.Fields["test"] == "value"
	})).Return(nil)

	logger.Debug("streaming debug message", fields)

	// Give async operation time to complete
	time.Sleep(10 * time.Millisecond)

	mockStreamer.AssertExpectations(t)
}

func TestOrchionLogger_Debug_StreamerError(t *testing.T) {
	var buf bytes.Buffer
	mockStreamer := &MockLogStreamer{}

	config := Config{
		Level:  DebugLevel,
		Source: "test-streamer-error",
	}

	logger := NewLogger(config)
	logger.SetOutput(&buf)
	logger.SetStreamer(mockStreamer)

	// Mock streamer to return an error
	mockStreamer.On("Stream", mock.Anything).Return(assert.AnError)

	logger.Debug("message that will fail streaming", nil)

	// Give async operation time to complete
	time.Sleep(10 * time.Millisecond)

	// Should still log locally
	output := buf.String()
	assert.Contains(t, output, "message that will fail streaming")

	mockStreamer.AssertExpectations(t)
}

func TestOrchionLogger_WithField(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})
	logger.SetOutput(&buf)

	fieldLogger := logger.WithField("persistent", "value")

	// Log with the field logger
	fieldLogger.Info("message with field", map[string]interface{}{
		"additional": "field",
	})

	output := buf.String()
	assert.Contains(t, output, "message with field")
	assert.Contains(t, output, "persistent")
	assert.Contains(t, output, "value")
	assert.Contains(t, output, "additional")
	assert.Contains(t, output, "field")

	// Original logger should not have the field
	logger.Info("message without field", nil)
	originalOutput := buf.String()
	assert.Contains(t, originalOutput, "message without field")
	// Should not contain the persistent field from the derived logger
}

func TestOrchionLogger_WithFields(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})
	logger.SetOutput(&buf)

	fieldsLogger := logger.WithFields(map[string]interface{}{
		"field1": "value1",
		"field2": "value2",
	})

	fieldsLogger.Info("message with fields", nil)

	output := buf.String()
	assert.Contains(t, output, "message with fields")
	assert.Contains(t, output, "field1")
	assert.Contains(t, output, "value1")
	assert.Contains(t, output, "field2")
	assert.Contains(t, output, "value2")
}

func TestOrchionLogger_WithField_Chaining(t *testing.T) {
	var buf bytes.Buffer
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})
	logger.SetOutput(&buf)

	// Chain multiple WithField calls
	chainedLogger := logger.WithField("a", "1").WithField("b", "2")

	chainedLogger.Info("chained fields", map[string]interface{}{
		"c": "3",
	})

	output := buf.String()
	assert.Contains(t, output, "chained fields")
	assert.Contains(t, output, "a")
	assert.Contains(t, output, "1")
	assert.Contains(t, output, "b")
	assert.Contains(t, output, "2")
	assert.Contains(t, output, "c")
	assert.Contains(t, output, "3")
}

func TestOrchionLogger_SetLevel(t *testing.T) {
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})

	// Test setting different levels
	testCases := []struct {
		setLevel    Level
		expectedLogrus logrus.Level
	}{
		{DebugLevel, logrus.DebugLevel},
		{InfoLevel, logrus.InfoLevel},
		{WarnLevel, logrus.WarnLevel},
		{ErrorLevel, logrus.ErrorLevel},
		{Level(999), logrus.InfoLevel}, // Default
	}

	for _, tc := range testCases {
		t.Run(tc.setLevel.String(), func(t *testing.T) {
			logger.SetLevel(tc.setLevel)
			orchionLogger := logger.(*orchionLogger)
			assert.Equal(t, tc.expectedLogrus, orchionLogger.logger.Level)
		})
	}
}

func TestOrchionLogger_SetOutput(t *testing.T) {
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})

	var buf bytes.Buffer
	logger.SetOutput(&buf)

	logger.Info("test output", nil)

	output := buf.String()
	assert.Contains(t, output, "test output")
}

func TestOrchionLogger_Close(t *testing.T) {
	mockStreamer := &MockLogStreamer{}
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})
	logger.SetStreamer(mockStreamer)

	mockStreamer.On("Close").Return(nil)

	logger.Close()

	mockStreamer.AssertExpectations(t)
}

func TestOrchionLogger_Close_NoStreamer(t *testing.T) {
	logger := NewLogger(Config{Level: InfoLevel, Source: "test"})

	// Should not panic
	assert.NotPanics(t, func() {
		logger.Close()
	})
}

func TestOrchionLogger_ConvertFields(t *testing.T) {
	logger := NewLogger(Config{Source: "test"})
	orchionLogger := logger.(*orchionLogger)

	input := map[string]interface{}{
		"string": "value",
		"int":    42,
		"bool":   true,
		"nil":    nil,
	}

	result := orchionLogger.convertFields(input)

	expected := map[string]string{
		"string": "value",
		"int":    "42",
		"bool":   "true",
		"nil":    "<nil>",
	}

	assert.Equal(t, expected, result)
}

func TestLogEntry_Structure(t *testing.T) {
	entry := &LogEntry{
		ID:        "test-id",
		Timestamp: 1234567890,
		Level:     InfoLevel,
		Source:    "test-source",
		Message:   "test message",
		Fields: map[string]string{
			"key": "value",
		},
	}

	assert.Equal(t, "test-id", entry.ID)
	assert.Equal(t, int64(1234567890), entry.Timestamp)
	assert.Equal(t, InfoLevel, entry.Level)
	assert.Equal(t, "test-source", entry.Source)
	assert.Equal(t, "test message", entry.Message)
	assert.Equal(t, map[string]string{"key": "value"}, entry.Fields)
}

// Benchmark tests
func BenchmarkOrchionLogger_Debug(b *testing.B) {
	logger := NewLogger(Config{Level: DebugLevel, Source: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("benchmark message", map[string]interface{}{
			"iteration": i,
		})
	}
}

func BenchmarkOrchionLogger_Debug_WithStreamer(b *testing.B) {
	mockStreamer := &MockLogStreamer{}
	logger := NewLogger(Config{Level: DebugLevel, Source: "bench-stream"})
	logger.SetStreamer(mockStreamer)

	// Mock streamer to not return errors
	mockStreamer.On("Stream", mock.Anything).Return(nil)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		logger.Debug("benchmark streaming message", map[string]interface{}{
			"iteration": i,
		})
	}
}

func BenchmarkOrchionLogger_WithField(b *testing.B) {
	logger := NewLogger(Config{Level: InfoLevel, Source: "bench"})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		fieldLogger := logger.WithField("request_id", i)
		fieldLogger.Info("benchmark with field", nil)
	}
}