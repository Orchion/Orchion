package orchestrator

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
)

// Service implements the Orchion gRPC service
type Service struct {
	pb.UnimplementedOrchestratorServer
	registry node.Registry
}

// NewService creates a new orchestrator service
func NewService(registry node.Registry) *Service {
	return &Service{
		registry: registry,
	}
}

// RegisterNode registers a new node with the orchestrator
func (s *Service) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*pb.RegisterNodeResponse, error) {
	if req.Node == nil {
		return nil, status.Error(codes.InvalidArgument, "node is required")
	}

	if req.Node.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "node.id is required")
	}

	if err := s.registry.Register(req.Node); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterNodeResponse{}, nil
}

// Heartbeat updates the heartbeat timestamp for a node
func (s *Service) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	if err := s.registry.UpdateHeartbeat(req.NodeId); err != nil {
		if err == node.ErrNodeNotFound {
			return nil, status.Error(codes.NotFound, "node not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.HeartbeatResponse{}, nil
}

// ListNodes returns all registered nodes
func (s *Service) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	nodes := s.registry.List()
	return &pb.ListNodesResponse{Nodes: nodes}, nil
}