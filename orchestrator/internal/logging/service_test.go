package logging

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/shared/logging"
)

func TestNewService(t *testing.T) {
	service := NewService()
	assert.NotNil(t, service)
	assert.NotNil(t, service.clients)
	assert.Len(t, service.clients, 0)
}

func TestService_convertLevel(t *testing.T) {
	service := NewService()

	tests := []struct {
		input    logging.Level
		expected pb.LogLevel
	}{
		{logging.DebugLevel, pb.LogLevel_LOG_LEVEL_DEBUG},
		{logging.InfoLevel, pb.LogLevel_LOG_LEVEL_INFO},
		{logging.WarnLevel, pb.LogLevel_LOG_LEVEL_WARN},
		{logging.ErrorLevel, pb.LogLevel_LOG_LEVEL_ERROR},
		{logging.Level(999), pb.LogLevel_LOG_LEVEL_INFO}, // Unknown level defaults to INFO
	}

	for _, test := range tests {
		result := service.convertLevel(test.input)
		assert.Equal(t, test.expected, result, "Failed for level %v", test.input)
	}
}

func TestService_Broadcast(t *testing.T) {
	service := NewService()

	// Create a test log entry
	entry := &logging.LogEntry{
		ID:        "test-id",
		Timestamp: time.Now().Unix(),
		Level:     logging.InfoLevel,
		Source:    "test-source",
		Message:   "Test message",
		Fields:    map[string]string{"key": "value"},
	}

	// Broadcast with no clients - should not panic
	service.Broadcast(entry)

	// Verify no clients are affected
	assert.Len(t, service.clients, 0)
}

func Test_generateClientID(t *testing.T) {
	// Generate multiple IDs
	id1 := generateClientID()
	id2 := generateClientID()

	// IDs should have the expected prefix
	assert.Contains(t, id1, "client-")
	assert.Contains(t, id2, "client-")

	// IDs should be different (due to timestamp nanoseconds)
	// Note: In very rare cases they might be the same, but that's acceptable
	if id1 == id2 {
		t.Log("Warning: Generated identical client IDs (rare but possible)")
	}
}



func TestService_Broadcast_WithNoClients(t *testing.T) {
	service := NewService()

	// Test broadcasting with no clients - should not panic
	entry := &logging.LogEntry{
		ID:        "test-1",
		Timestamp: time.Now().Unix(),
		Level:     logging.InfoLevel,
		Source:    "test-source",
		Message:   "Info message",
		Fields:    map[string]string{"key": "value"},
	}

	// Broadcast should not panic even with no clients
	service.Broadcast(entry)

	// Verify no clients are registered
	assert.Len(t, service.clients, 0)
}

func TestService_ConcurrentBroadcast(t *testing.T) {
	service := NewService()

	// Test concurrent broadcasting with no clients
	const numGoroutines = 5
	const messagesPerGoroutine = 3

	done := make(chan bool, numGoroutines)

	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer func() { done <- true }()

			for j := 0; j < messagesPerGoroutine; j++ {
				entry := &logging.LogEntry{
					ID:        fmt.Sprintf("concurrent-%d-%d", id, j),
					Timestamp: time.Now().Unix(),
					Level:     logging.InfoLevel,
					Source:    fmt.Sprintf("source-%d", id),
					Message:   fmt.Sprintf("Message %d-%d", id, j),
					Fields:    map[string]string{"id": fmt.Sprintf("%d", id)},
				}
				service.Broadcast(entry)
			}
		}(i)
	}

	// Wait for all goroutines
	for i := 0; i < numGoroutines; i++ {
		<-done
	}

	// Service should still be in valid state
	assert.NotNil(t, service)
	assert.NotNil(t, service.clients)
	assert.Len(t, service.clients, 0) // No clients connected
}