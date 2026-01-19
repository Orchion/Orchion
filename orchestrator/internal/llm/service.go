package llm

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	"github.com/Orchion/Orchion/orchestrator/internal/scheduler"
)

// Service implements the OrchionLLM gRPC service
type Service struct {
	pb.UnimplementedOrchionLLMServer
	registry  node.Registry
	scheduler scheduler.Scheduler
	// nodeClients maintains gRPC connections to node agents
	nodeClients map[string]pb.NodeAgentClient
	mu          sync.RWMutex
}

// NewService creates a new LLM service
func NewService(registry node.Registry, sched scheduler.Scheduler) *Service {
	return &Service{
		registry:    registry,
		scheduler:   sched,
		nodeClients: make(map[string]pb.NodeAgentClient),
	}
}

// ChatCompletion handles chat completion requests
func (s *Service) ChatCompletion(req *pb.ChatCompletionRequest, stream pb.OrchionLLM_ChatCompletionServer) error {
	if req.Model == "" {
		return status.Error(codes.InvalidArgument, "model is required")
	}

	if len(req.Messages) == 0 {
		return status.Error(codes.InvalidArgument, "messages are required")
	}

	// Select a node for this model
	selectedNode, err := s.scheduler.SelectNode(req.Model, s.registry)
	if err != nil {
		return status.Error(codes.NotFound, fmt.Sprintf("no node available for model %s: %v", req.Model, err))
	}

	// Get or create gRPC client for this node
	client, err := s.getNodeClient(selectedNode.Id, selectedNode)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to connect to node: %v", err))
	}

	// Forward request to node agent
	nodeStream, err := client.ChatCompletion(context.Background(), req)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to call node agent: %v", err))
	}

	// Stream responses back to gateway
	for {
		resp, err := nodeStream.Recv()
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				return nil
			}
			return status.Error(codes.Internal, fmt.Sprintf("error receiving from node: %v", err))
		}

		if err := stream.Send(resp); err != nil {
			return err
		}
	}
}

// Embeddings handles embedding requests
func (s *Service) Embeddings(ctx context.Context, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	if req.Model == "" {
		return nil, status.Error(codes.InvalidArgument, "model is required")
	}

	if len(req.Input) == 0 {
		return nil, status.Error(codes.InvalidArgument, "input is required")
	}

	// Select a node for this model
	selectedNode, err := s.scheduler.SelectNode(req.Model, s.registry)
	if err != nil {
		return nil, status.Error(codes.NotFound, fmt.Sprintf("no node available for model %s: %v", req.Model, err))
	}

	// Get or create gRPC client for this node
	client, err := s.getNodeClient(selectedNode.Id, selectedNode)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to connect to node: %v", err))
	}

	// Forward request to node agent
	return client.Embeddings(ctx, req)
}

// getNodeClient gets or creates a gRPC client for a node
func (s *Service) getNodeClient(nodeID string, node *pb.Node) (pb.NodeAgentClient, error) {
	s.mu.RLock()
	if client, exists := s.nodeClients[nodeID]; exists {
		s.mu.RUnlock()
		return client, nil
	}
	s.mu.RUnlock()

	s.mu.Lock()
	defer s.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := s.nodeClients[nodeID]; exists {
		return client, nil
	}

	// Determine node agent address
	addr := node.AgentAddress
	if addr == "" {
		// Default to hostname:50052 if not specified
		addr = fmt.Sprintf("%s:50052", node.Hostname)
	}

	// Connect to node agent
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node %s at %s: %w", nodeID, addr, err)
	}

	client := pb.NewNodeAgentClient(conn)
	s.nodeClients[nodeID] = client

	return client, nil
}
