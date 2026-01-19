package executor

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/Orchion/Orchion/node-agent/internal/containers"
	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// Service implements the NodeAgent gRPC service using containerized inference engines
type Service struct {
	pb.UnimplementedNodeAgentServer
	containerManager containers.Manager
	executors        map[string]Executor // model name -> executor
	runningModels    map[string]*ModelInstance
	mu               sync.RWMutex
}

// Executor handles inference for a specific model type (Ollama, vLLM, etc.)
type Executor interface {
	StartModel(ctx context.Context, model string) error
	StopModel(ctx context.Context, model string) error
	IsModelRunning(ctx context.Context, model string) (bool, error)
	ChatCompletion(ctx context.Context, model string, req *pb.ChatCompletionRequest) (<-chan *pb.ChatCompletionResponse, error)
	Embeddings(ctx context.Context, model string, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error)
}

// ModelInstance tracks running model instances
type ModelInstance struct {
	Model     string
	Executor  Executor
	StartTime time.Time
}

// NewService creates a new executor service
func NewService() (*Service, error) {
	manager, err := containers.NewContainerManager()
	if err != nil {
		return nil, fmt.Errorf("failed to create container manager: %w", err)
	}

	service := &Service{
		containerManager: manager,
		executors:        make(map[string]Executor),
		runningModels:    make(map[string]*ModelInstance),
	}

	// Register default executors
	service.executors["ollama"] = NewOllamaExecutor(manager)
	service.executors["vllm"] = NewVLLMExecutor(manager)

	return service, nil
}

// ChatCompletion handles chat completion requests by routing to appropriate executor
func (s *Service) ChatCompletion(req *pb.ChatCompletionRequest, stream pb.NodeAgent_ChatCompletionServer) error {
	if req.Model == "" {
		return status.Error(codes.InvalidArgument, "model is required")
	}

	ctx := stream.Context()

	// Ensure model is running
	if err := s.ensureModelRunning(ctx, req.Model); err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to start model %s: %v", req.Model, err))
	}

	// Get executor for this model
	executor, err := s.getExecutorForModel(req.Model)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("no executor for model %s: %v", req.Model, err))
	}

	// Execute request
	responseChan, err := executor.ChatCompletion(ctx, req.Model, req)
	if err != nil {
		return status.Error(codes.Internal, fmt.Sprintf("failed to execute chat completion: %v", err))
	}

	// Stream responses
	for resp := range responseChan {
		if err := stream.Send(resp); err != nil {
			return err
		}
	}

	return nil
}

// Embeddings handles embedding requests by routing to appropriate executor
func (s *Service) Embeddings(ctx context.Context, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	if req.Model == "" {
		return nil, status.Error(codes.InvalidArgument, "model is required")
	}

	// Ensure model is running
	if err := s.ensureModelRunning(ctx, req.Model); err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("failed to start model %s: %v", req.Model, err))
	}

	// Get executor for this model
	executor, err := s.getExecutorForModel(req.Model)
	if err != nil {
		return nil, status.Error(codes.Internal, fmt.Sprintf("no executor for model %s: %v", req.Model, err))
	}

	// Execute request
	return executor.Embeddings(ctx, req.Model, req)
}

// ensureModelRunning ensures the specified model is running
func (s *Service) ensureModelRunning(ctx context.Context, model string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Check if already running
	if instance, exists := s.runningModels[model]; exists {
		running, err := instance.Executor.IsModelRunning(ctx, model)
		if err != nil {
			log.Printf("Failed to check if model %s is running: %v", model, err)
			// Continue with starting the model
		} else if running {
			return nil // Already running
		}
	}

	// Get executor for this model
	executor, err := s.getExecutorForModel(model)
	if err != nil {
		return fmt.Errorf("no executor for model %s: %w", model, err)
	}

	// Start the model
	log.Printf("Starting model: %s", model)
	if err := executor.StartModel(ctx, model); err != nil {
		return fmt.Errorf("failed to start model %s: %w", model, err)
	}

	// Track the running model
	s.runningModels[model] = &ModelInstance{
		Model:     model,
		Executor:  executor,
		StartTime: time.Now(),
	}

	log.Printf("Model %s started successfully", model)
	return nil
}

// getExecutorForModel determines which executor to use for a given model
func (s *Service) getExecutorForModel(model string) (Executor, error) {
	// Simple routing logic - can be enhanced later
	// For now: use Ollama for models without "/" (like "llama2", "mistral")
	// and vLLM for models with "/" (like "mistralai/Mistral-7B")

	if strings.Contains(model, "/") {
		// Likely a HuggingFace model, use vLLM
		if executor, exists := s.executors["vllm"]; exists {
			return executor, nil
		}
	} else {
		// Likely an Ollama model name, use Ollama
		if executor, exists := s.executors["ollama"]; exists {
			return executor, nil
		}
	}

	// Fallback to Ollama
	if executor, exists := s.executors["ollama"]; exists {
		return executor, nil
	}

	return nil, fmt.Errorf("no suitable executor found for model %s", model)
}

// Shutdown gracefully shuts down all running models
func (s *Service) Shutdown(ctx context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	log.Printf("Shutting down executor service...")

	var lastErr error
	for model, instance := range s.runningModels {
		log.Printf("Stopping model: %s", model)
		if err := instance.Executor.StopModel(ctx, model); err != nil {
			log.Printf("Error stopping model %s: %v", model, err)
			lastErr = err
		}
	}

	// Clear running models
	s.runningModels = make(map[string]*ModelInstance)

	if lastErr != nil {
		return fmt.Errorf("errors occurred during shutdown: %w", lastErr)
	}

	log.Printf("Executor service shutdown complete")
	return nil
}
