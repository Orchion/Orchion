package inference

import (
	"context"
	"fmt"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// Service implements the NodeAgent gRPC service
type Service struct {
	pb.UnimplementedNodeAgentServer
	engines map[string]Engine // Map of model name to engine
}

// Engine represents an inference engine (vLLM, Ollama, etc.)
type Engine interface {
	ChatCompletion(ctx context.Context, req *pb.ChatCompletionRequest) (<-chan *pb.ChatCompletionResponse, error)
	Embeddings(ctx context.Context, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error)
}

// NewService creates a new inference service
func NewService() *Service {
	return &Service{
		engines: make(map[string]Engine),
	}
}

// RegisterEngine registers an inference engine for a model
func (s *Service) RegisterEngine(model string, engine Engine) {
	s.engines[model] = engine
}

// ChatCompletion handles chat completion requests
func (s *Service) ChatCompletion(req *pb.ChatCompletionRequest, stream pb.NodeAgent_ChatCompletionServer) error {
	if req.Model == "" {
		return status.Error(codes.InvalidArgument, "model is required")
	}

	// Find engine for this model
	engine, exists := s.engines[req.Model]
	if !exists {
		// Try to find a default engine or create one
		// For now, try to use Ollama as default
		engine = NewOllamaEngine("localhost:11434")
		s.engines[req.Model] = engine
	}

	// Get response channel
	responseChan, err := engine.ChatCompletion(stream.Context(), req)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to get completion: %v", err))
	}

	// Stream responses
	for resp := range responseChan {
		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

// Embeddings handles embedding requests
func (s *Service) Embeddings(ctx context.Context, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	if req.Model == "" {
		return nil, status.Error(codes.InvalidArgument, "model is required")
	}

	// Find engine for this model
	engine, exists := s.engines[req.Model]
	if !exists {
		// Try to use Ollama as default
		engine = NewOllamaEngine("localhost:11434")
		s.engines[req.Model] = engine
	}

	return engine.Embeddings(ctx, req)
}
