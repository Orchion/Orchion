package logging

import (
	"github.com/Orchion/Orchion/shared/logging"
)

// OrchestratorStreamer implements logging.LogStreamer for broadcasting logs within the orchestrator
type OrchestratorStreamer struct {
	service *Service
}

// NewOrchestratorStreamer creates a new orchestrator streamer
func NewOrchestratorStreamer(service *Service) *OrchestratorStreamer {
	return &OrchestratorStreamer{
		service: service,
	}
}

// Stream sends a log entry to all connected clients
func (s *OrchestratorStreamer) Stream(entry *logging.LogEntry) error {
	s.service.Broadcast(entry)
	return nil
}

// Close is a no-op for the orchestrator streamer
func (s *OrchestratorStreamer) Close() error {
	return nil
}
