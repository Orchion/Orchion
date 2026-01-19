package executor

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"

	"github.com/Orchion/Orchion/node-agent/internal/containers"
	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// OllamaExecutor manages Ollama containers and handles inference requests
type OllamaExecutor struct {
	containerManager containers.Manager
	basePort         int            // Starting port for Ollama containers
	runningPorts     map[string]int // model -> port mapping
	dockerAvailable  bool           // Whether Docker is available
}

// NewOllamaExecutor creates a new Ollama executor
func NewOllamaExecutor(manager containers.Manager) *OllamaExecutor {
	executor := &OllamaExecutor{
		containerManager: manager,
		basePort:         11434, // Default Ollama port
		runningPorts:     make(map[string]int),
		dockerAvailable:  true,
	}

	// Test if container runtime is available
	if err := manager.TestConnection(); err != nil {
		log.Printf("Warning: Container runtime not available for Ollama executor: %v", err)
		log.Printf("OllamaExecutor will assume Ollama is running externally on port %d", executor.basePort)
		executor.dockerAvailable = false
	}

	return executor
}

// StartModel starts an Ollama container for the specified model
func (e *OllamaExecutor) StartModel(ctx context.Context, model string) error {
	if e.dockerAvailable {
		// Use container-based approach
		config := containers.CreateOllamaContainerConfig(containers.DefaultOllamaConfig())

		// Ensure container is running
		if err := e.containerManager.EnsureRunning(ctx, config); err != nil {
			return fmt.Errorf("failed to start Ollama container: %w", err)
		}

		// Wait for Ollama to be ready
		if err := e.waitForOllamaReady(ctx, config.Port); err != nil {
			return fmt.Errorf("Ollama container failed to become ready: %w", err)
		}

		// Pull the model
		if err := containers.PullOllamaModel(ctx, e.containerManager, config.Name, model); err != nil {
			log.Printf("Warning: Failed to pull model %s: %v", model, err)
			// Don't fail here - model might already be available
		}

		// Track the port
		e.runningPorts[model] = config.Port

		log.Printf("Ollama model %s ready on port %d (container)", model, config.Port)
	} else {
		// Assume Ollama is running externally
		port := e.basePort
		if err := e.waitForOllamaReady(ctx, port); err != nil {
			return fmt.Errorf("external Ollama not available on port %d: %w", port, err)
		}

		e.runningPorts[model] = port
		log.Printf("Ollama model %s assumed ready on port %d (external)", model, port)
	}

	return nil
}

// StopModel stops the Ollama container for the specified model
func (e *OllamaExecutor) StopModel(ctx context.Context, model string) error {
	if e.dockerAvailable {
		config := containers.CreateOllamaContainerConfig(containers.DefaultOllamaConfig())

		if err := e.containerManager.StopContainer(ctx, config.Name); err != nil {
			return fmt.Errorf("failed to stop Ollama container: %w", err)
		}

		log.Printf("Stopped Ollama container for model %s", model)
	} else {
		log.Printf("Ollama assumed to be running externally, not stopping model %s", model)
	}

	delete(e.runningPorts, model)
	return nil
}

// IsModelRunning checks if the Ollama container is running for the specified model
func (e *OllamaExecutor) IsModelRunning(ctx context.Context, model string) (bool, error) {
	config := containers.CreateOllamaContainerConfig(containers.DefaultOllamaConfig())
	return e.containerManager.IsRunning(ctx, config.Name)
}

// ChatCompletion executes a chat completion request using Ollama
func (e *OllamaExecutor) ChatCompletion(ctx context.Context, model string, req *pb.ChatCompletionRequest) (<-chan *pb.ChatCompletionResponse, error) {
	port, exists := e.runningPorts[model]
	if !exists {
		return nil, fmt.Errorf("model %s is not running", model)
	}

	responseChan := make(chan *pb.ChatCompletionResponse, 10)

	go func() {
		defer close(responseChan)

		// Convert messages to Ollama format
		messages := make([]map[string]string, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			}
		}

		// Build Ollama API request
		ollamaReq := map[string]interface{}{
			"model":    model,
			"messages": messages,
			"stream":   req.Stream,
		}
		if req.Temperature > 0 {
			ollamaReq["temperature"] = req.Temperature
		}
		if req.MaxTokens > 0 {
			ollamaReq["num_predict"] = req.MaxTokens
		}

		reqBody, err := json.Marshal(ollamaReq)
		if err != nil {
			responseChan <- e.createErrorResponse(model, "failed to marshal request")
			return
		}

		// Make request to Ollama
		url := fmt.Sprintf("http://localhost:%d/api/chat", port)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
		if err != nil {
			responseChan <- e.createErrorResponse(model, "failed to create request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Minute}
		resp, err := client.Do(httpReq)
		if err != nil {
			responseChan <- e.createErrorResponse(model, "failed to call Ollama")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			responseChan <- e.createErrorResponse(model, fmt.Sprintf("Ollama returned status %d", resp.StatusCode))
			return
		}

		if req.Stream {
			// Stream responses
			e.handleStreamingResponse(resp.Body, model, responseChan)
		} else {
			// Non-streaming response
			e.handleNonStreamingResponse(resp.Body, model, responseChan)
		}
	}()

	return responseChan, nil
}

// Embeddings executes an embeddings request using Ollama
func (e *OllamaExecutor) Embeddings(ctx context.Context, model string, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	port, exists := e.runningPorts[model]
	if !exists {
		return nil, fmt.Errorf("model %s is not running", model)
	}

	embeddings := make([]*pb.Embedding, 0, len(req.Input))

	for i, input := range req.Input {
		ollamaReq := map[string]interface{}{
			"model":  model,
			"prompt": input,
		}

		reqBody, err := json.Marshal(ollamaReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		url := fmt.Sprintf("http://localhost:%d/api/embeddings", port)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 5 * time.Minute}
		resp, err := client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to call Ollama: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return nil, fmt.Errorf("Ollama returned status %d", resp.StatusCode)
		}

		var ollamaResp map[string]interface{}
		if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
			return nil, fmt.Errorf("failed to decode response: %w", err)
		}

		embeddingSlice, ok := ollamaResp["embedding"].([]interface{})
		if !ok {
			return nil, fmt.Errorf("invalid embedding format")
		}

		embedding := make([]float32, len(embeddingSlice))
		for j, v := range embeddingSlice {
			if f, ok := v.(float64); ok {
				embedding[j] = float32(f)
			}
		}

		embeddings = append(embeddings, &pb.Embedding{
			Index:     int32(i),
			Embedding: embedding,
		})
	}

	return &pb.EmbeddingResponse{
		Model:  model,
		Object: "list",
		Data:   embeddings,
	}, nil
}

// waitForOllamaReady waits for Ollama to be ready to accept requests
func (e *OllamaExecutor) waitForOllamaReady(ctx context.Context, port int) error {
	url := fmt.Sprintf("http://localhost:%d/api/tags", port)
	client := &http.Client{Timeout: 10 * time.Second}

	// Try for up to 2 minutes
	for i := 0; i < 120; i++ {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
		if err != nil {
			return err
		}

		resp, err := client.Do(req)
		if err == nil {
			resp.Body.Close()
			if resp.StatusCode == http.StatusOK {
				return nil
			}
		}

		time.Sleep(1 * time.Second)
	}

	return fmt.Errorf("timeout waiting for Ollama to be ready")
}

// handleStreamingResponse processes streaming Ollama responses
func (e *OllamaExecutor) handleStreamingResponse(body io.Reader, model string, responseChan chan<- *pb.ChatCompletionResponse) {
	decoder := json.NewDecoder(body)
	for {
		var ollamaResp map[string]interface{}
		if err := decoder.Decode(&ollamaResp); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error decoding streaming response: %v", err)
			continue
		}

		// Extract content from Ollama response
		message, ok := ollamaResp["message"].(map[string]interface{})
		if !ok {
			continue
		}
		content, _ := message["content"].(string)
		done, _ := ollamaResp["done"].(bool)

		responseChan <- &pb.ChatCompletionResponse{
			Id:     e.generateID(),
			Model:  model,
			Object: "chat.completion.chunk",
			Choices: []*pb.ChatChoice{
				{
					Index: 0,
					Message: &pb.ChatMessage{
						Role:    "assistant",
						Content: content,
					},
					FinishReason: func() string {
						if done {
							return "stop"
						}
						return ""
					}(),
				},
			},
			Created: time.Now().Unix(),
		}

		if done {
			break
		}
	}
}

// handleNonStreamingResponse processes non-streaming Ollama responses
func (e *OllamaExecutor) handleNonStreamingResponse(body io.Reader, model string, responseChan chan<- *pb.ChatCompletionResponse) {
	var ollamaResp map[string]interface{}
	if err := json.NewDecoder(body).Decode(&ollamaResp); err != nil {
		responseChan <- e.createErrorResponse(model, "failed to decode response")
		return
	}

	message, _ := ollamaResp["message"].(map[string]interface{})
	content, _ := message["content"].(string)

	responseChan <- &pb.ChatCompletionResponse{
		Id:     e.generateID(),
		Model:  model,
		Object: "chat.completion",
		Choices: []*pb.ChatChoice{
			{
				Index: 0,
				Message: &pb.ChatMessage{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: "stop",
			},
		},
		Created: time.Now().Unix(),
	}
}

// createErrorResponse creates an error response
func (e *OllamaExecutor) createErrorResponse(model, message string) *pb.ChatCompletionResponse {
	return &pb.ChatCompletionResponse{
		Id:      e.generateID(),
		Model:   model,
		Object:  "error",
		Choices: []*pb.ChatChoice{{FinishReason: "error"}},
		Created: time.Now().Unix(),
	}
}

// generateID generates a unique ID for responses
func (e *OllamaExecutor) generateID() string {
	return fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
}
