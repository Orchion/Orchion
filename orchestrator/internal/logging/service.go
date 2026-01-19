package logging

import (
	"fmt"
	"sync"
	"time"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/shared/logging"
)

// Service implements the LogStreamer gRPC service
type Service struct {
	pb.UnimplementedLogStreamerServer
	mu      sync.RWMutex
	clients map[string]pb.LogStreamer_StreamLogsServer
}

// NewService creates a new logging service
func NewService() *Service {
	return &Service{
		clients: make(map[string]pb.LogStreamer_StreamLogsServer),
	}
}

// StreamLogs handles streaming log entries to connected clients
func (s *Service) StreamLogs(req *pb.StreamLogsRequest, stream pb.LogStreamer_StreamLogsServer) error {
	// Generate a client ID
	clientID := generateClientID()

	s.mu.Lock()
	s.clients[clientID] = stream
	s.mu.Unlock()

	// Clean up when client disconnects
	defer func() {
		s.mu.Lock()
		delete(s.clients, clientID)
		s.mu.Unlock()
	}()

	// Keep the stream open
	<-stream.Context().Done()
	return stream.Context().Err()
}

// Broadcast sends a log entry to all connected clients
func (s *Service) Broadcast(entry *logging.LogEntry) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	pbEntry := &pb.LogEntry{
		Id:        entry.ID,
		Timestamp: entry.Timestamp,
		Level:     s.convertLevel(entry.Level),
		Source:    entry.Source,
		Message:   entry.Message,
		Fields:    entry.Fields,
	}

	for clientID, stream := range s.clients {
		go func(clientID string, stream pb.LogStreamer_StreamLogsServer) {
			if err := stream.Send(&pb.StreamLogsResponse{Entry: pbEntry}); err != nil {
				// Client disconnected, will be cleaned up on next send
			}
		}(clientID, stream)
	}
}

// convertLevel converts logging.Level to pb.LogLevel
func (s *Service) convertLevel(level logging.Level) pb.LogLevel {
	switch level {
	case logging.DebugLevel:
		return pb.LogLevel_LOG_LEVEL_DEBUG
	case logging.InfoLevel:
		return pb.LogLevel_LOG_LEVEL_INFO
	case logging.WarnLevel:
		return pb.LogLevel_LOG_LEVEL_WARN
	case logging.ErrorLevel:
		return pb.LogLevel_LOG_LEVEL_ERROR
	default:
		return pb.LogLevel_LOG_LEVEL_INFO
	}
}

// generateClientID generates a unique client ID
func generateClientID() string {
	return fmt.Sprintf("client-%d", time.Now().UnixNano())
}
