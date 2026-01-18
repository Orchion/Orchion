package orchestrator

import (
	"context"
	"log"
	"time"

	orchionv1 "github.com/Orchion/Orchion/api/proto/v1"
	"github.com/Orchion/Orchion/internal/node"
)

// Service implements the orchestrator gRPC service
type Service struct {
	orchionv1.UnimplementedOrchestratorServiceServer
	registry          node.Registry
	heartbeatTimeout  time.Duration
}

// NewService creates a new orchestrator service
func NewService(registry node.Registry, heartbeatTimeout time.Duration) *Service {
	return &Service{
		registry:         registry,
		heartbeatTimeout: heartbeatTimeout,
	}
}

// StartHeartbeatMonitor starts a background goroutine that monitors node heartbeats
func (s *Service) StartHeartbeatMonitor(ctx context.Context) {
	ticker := time.NewTicker(s.heartbeatTimeout / 2)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("Stopping heartbeat monitor")
			return
		case <-ticker.C:
			s.registry.MarkInactive(s.heartbeatTimeout)
		}
	}
}

// RegisterNode registers a new node with the orchestrator
func (s *Service) RegisterNode(ctx context.Context, req *orchionv1.RegisterNodeRequest) (*orchionv1.RegisterNodeResponse, error) {
	log.Printf("Registering node: address=%s, metadata=%v", req.Address, req.Metadata)

	nodeID, err := s.registry.Register(req.Address, req.Metadata)
	if err != nil {
		log.Printf("Failed to register node: %v", err)
		return nil, err
	}

	log.Printf("Node registered successfully: nodeID=%s", nodeID)
	return &orchionv1.RegisterNodeResponse{
		NodeId: nodeID,
	}, nil
}

// Heartbeat handles heartbeat requests from nodes
func (s *Service) Heartbeat(ctx context.Context, req *orchionv1.HeartbeatRequest) (*orchionv1.HeartbeatResponse, error) {
	err := s.registry.UpdateHeartbeat(req.NodeId)
	if err != nil {
		log.Printf("Failed to update heartbeat for node %s: %v", req.NodeId, err)
		return &orchionv1.HeartbeatResponse{
			Acknowledged: false,
		}, err
	}

	return &orchionv1.HeartbeatResponse{
		Acknowledged: true,
	}, nil
}

// UnregisterNode removes a node from the orchestrator
func (s *Service) UnregisterNode(ctx context.Context, req *orchionv1.UnregisterNodeRequest) (*orchionv1.UnregisterNodeResponse, error) {
	log.Printf("Unregistering node: nodeID=%s", req.NodeId)

	err := s.registry.Unregister(req.NodeId)
	if err != nil {
		log.Printf("Failed to unregister node: %v", err)
		return &orchionv1.UnregisterNodeResponse{
			Success: false,
		}, err
	}

	log.Printf("Node unregistered successfully: nodeID=%s", req.NodeId)
	return &orchionv1.UnregisterNodeResponse{
		Success: true,
	}, nil
}

// ListNodes returns all registered nodes
func (s *Service) ListNodes(ctx context.Context, req *orchionv1.ListNodesRequest) (*orchionv1.ListNodesResponse, error) {
	nodes, err := s.registry.List(req.StatusFilter)
	if err != nil {
		log.Printf("Failed to list nodes: %v", err)
		return nil, err
	}

	return &orchionv1.ListNodesResponse{
		Nodes: nodes,
	}, nil
}
