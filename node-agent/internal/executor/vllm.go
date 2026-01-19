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

// VLLMExecutor manages vLLM containers and handles inference requests
type VLLMExecutor struct {
	containerManager containers.Manager
	basePort         int            // Starting port for vLLM containers
	runningPorts     map[string]int // model -> port mapping
}

// NewVLLMExecutor creates a new vLLM executor
func NewVLLMExecutor(manager containers.Manager) *VLLMExecutor {
	return &VLLMExecutor{
		containerManager: manager,
		basePort:         8000, // Default vLLM port
		runningPorts:     make(map[string]int),
	}
}

// StartModel starts a vLLM container for the specified model
func (e *VLLMExecutor) StartModel(ctx context.Context, model string) error {
	// Create vLLM config for this model
	config := containers.CreateVLLMContainerConfig(&containers.VLLMConfig{
		Model:              model,
		Port:               e.basePort,
		GPUs:               []string{"all"},
		TensorParallelSize: 1,
		MaxModelLen:        4096,
	})

	// Ensure container is running
	if err := e.containerManager.EnsureRunning(ctx, config); err != nil {
		return fmt.Errorf("failed to start vLLM container: %w", err)
	}

	// Wait for vLLM to be ready
	if err := e.waitForVLLMReady(ctx, config.Port); err != nil {
		return fmt.Errorf("vLLM container failed to become ready: %w", err)
	}

	// Track the port
	e.runningPorts[model] = config.Port

	log.Printf("vLLM model %s ready on port %d", model, config.Port)
	return nil
}

// StopModel stops the vLLM container for the specified model
func (e *VLLMExecutor) StopModel(ctx context.Context, model string) error {
	config := containers.CreateVLLMContainerConfig(&containers.VLLMConfig{
		Model: model,
		Port:  e.basePort,
	})

	if err := e.containerManager.StopContainer(ctx, config.Name); err != nil {
		return fmt.Errorf("failed to stop vLLM container: %w", err)
	}

	delete(e.runningPorts, model)
	log.Printf("Stopped vLLM container for model %s", model)
	return nil
}

// IsModelRunning checks if the vLLM container is running for the specified model
func (e *VLLMExecutor) IsModelRunning(ctx context.Context, model string) (bool, error) {
	config := containers.CreateVLLMContainerConfig(&containers.VLLMConfig{
		Model: model,
		Port:  e.basePort,
	})
	return e.containerManager.IsRunning(ctx, config.Name)
}

// ChatCompletion executes a chat completion request using vLLM
func (e *VLLMExecutor) ChatCompletion(ctx context.Context, model string, req *pb.ChatCompletionRequest) (<-chan *pb.ChatCompletionResponse, error) {
	port, exists := e.runningPorts[model]
	if !exists {
		return nil, fmt.Errorf("model %s is not running", model)
	}

	responseChan := make(chan *pb.ChatCompletionResponse, 10)

	go func() {
		defer close(responseChan)

		// Convert messages to OpenAI format
		messages := make([]map[string]interface{}, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = map[string]interface{}{
				"role":    msg.Role,
				"content": msg.Content,
			}
		}

		// Build OpenAI-compatible request
		openaiReq := map[string]interface{}{
			"model":    model,
			"messages": messages,
			"stream":   req.Stream,
		}
		if req.Temperature > 0 {
			openaiReq["temperature"] = req.Temperature
		}
		if req.MaxTokens > 0 {
			openaiReq["max_tokens"] = req.MaxTokens
		}

		reqBody, err := json.Marshal(openaiReq)
		if err != nil {
			responseChan <- e.createErrorResponse(model, "failed to marshal request")
			return
		}

		// Make request to vLLM
		url := fmt.Sprintf("http://localhost:%d/v1/chat/completions", port)
		httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
		if err != nil {
			responseChan <- e.createErrorResponse(model, "failed to create request")
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		client := &http.Client{Timeout: 10 * time.Minute}
		resp, err := client.Do(httpReq)
		if err != nil {
			responseChan <- e.createErrorResponse(model, "failed to call vLLM")
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			responseChan <- e.createErrorResponse(model, fmt.Sprintf("vLLM returned status %d", resp.StatusCode))
			return
		}

		if req.Stream {
			// Handle streaming response
			e.handleVLLMStreamingResponse(resp.Body, model, responseChan)
		} else {
			// Handle non-streaming response
			e.handleVLLMNonStreamingResponse(resp.Body, model, responseChan)
		}
	}()

	return responseChan, nil
}

// Embeddings executes an embeddings request using vLLM
func (e *VLLMExecutor) Embeddings(ctx context.Context, model string, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	port, exists := e.runningPorts[model]
	if !exists {
		return nil, fmt.Errorf("model %s is not running", model)
	}

	// Build OpenAI-compatible request
	openaiReq := map[string]interface{}{
		"model": model,
		"input": req.Input,
	}

	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("http://localhost:%d/v1/embeddings", port)
	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	client := &http.Client{Timeout: 5 * time.Minute}
	resp, err := client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call vLLM: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("vLLM returned status %d", resp.StatusCode)
	}

	// Parse OpenAI-compatible response
	var openaiResp struct {
		Data []struct {
			Embedding []float32 `json:"embedding"`
			Index     int32     `json:"index"`
		} `json:"data"`
		Usage struct {
			PromptTokens int32 `json:"prompt_tokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	// Convert to protobuf format
	embeddings := make([]*pb.Embedding, len(openaiResp.Data))
	for i, data := range openaiResp.Data {
		embeddings[i] = &pb.Embedding{
			Index:     data.Index,
			Embedding: data.Embedding,
		}
	}

	return &pb.EmbeddingResponse{
		Model:             model,
		Object:            "list",
		Data:              embeddings,
		UsagePromptTokens: openaiResp.Usage.PromptTokens,
	}, nil
}

// waitForVLLMReady waits for vLLM to be ready to accept requests
func (e *VLLMExecutor) waitForVLLMReady(ctx context.Context, port int) error {
	url := fmt.Sprintf("http://localhost:%d/v1/models", port)
	client := &http.Client{Timeout: 10 * time.Second}

	// Try for up to 5 minutes (vLLM can take longer to start)
	for i := 0; i < 300; i++ {
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

	return fmt.Errorf("timeout waiting for vLLM to be ready")
}

// handleVLLMStreamingResponse processes streaming vLLM responses
func (e *VLLMExecutor) handleVLLMStreamingResponse(body io.Reader, model string, responseChan chan<- *pb.ChatCompletionResponse) {
	decoder := json.NewDecoder(body)

	for {
		var openaiResp struct {
			ID      string `json:"id"`
			Object  string `json:"object"`
			Created int64  `json:"created"`
			Model   string `json:"model"`
			Choices []struct {
				Index int `json:"index"`
				Delta struct {
					Content string `json:"content"`
				} `json:"delta"`
				FinishReason *string `json:"finish_reason"`
			} `json:"choices"`
		}

		if err := decoder.Decode(&openaiResp); err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("Error decoding streaming response: %v", err)
			continue
		}

		if len(openaiResp.Choices) == 0 {
			continue
		}

		choice := openaiResp.Choices[0]
		finishReason := ""
		if choice.FinishReason != nil {
			finishReason = *choice.FinishReason
		}

		responseChan <- &pb.ChatCompletionResponse{
			Id:     openaiResp.ID,
			Model:  model,
			Object: "chat.completion.chunk",
			Choices: []*pb.ChatChoice{
				{
					Index: int32(choice.Index),
					Message: &pb.ChatMessage{
						Role:    "assistant",
						Content: choice.Delta.Content,
					},
					FinishReason: finishReason,
				},
			},
			Created: openaiResp.Created,
		}

		// Check if this is the final message
		if finishReason != "" {
			break
		}
	}
}

// handleVLLMNonStreamingResponse processes non-streaming vLLM responses
func (e *VLLMExecutor) handleVLLMNonStreamingResponse(body io.Reader, model string, responseChan chan<- *pb.ChatCompletionResponse) {
	var openaiResp struct {
		ID      string `json:"id"`
		Object  string `json:"object"`
		Created int64  `json:"created"`
		Model   string `json:"model"`
		Choices []struct {
			Index   int `json:"index"`
			Message struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		} `json:"choices"`
	}

	if err := json.NewDecoder(body).Decode(&openaiResp); err != nil {
		responseChan <- e.createErrorResponse(model, "failed to decode response")
		return
	}

	if len(openaiResp.Choices) == 0 {
		responseChan <- e.createErrorResponse(model, "no choices in response")
		return
	}

	choice := openaiResp.Choices[0]
	responseChan <- &pb.ChatCompletionResponse{
		Id:     openaiResp.ID,
		Model:  model,
		Object: "chat.completion",
		Choices: []*pb.ChatChoice{
			{
				Index: int32(choice.Index),
				Message: &pb.ChatMessage{
					Role:    choice.Message.Role,
					Content: choice.Message.Content,
				},
				FinishReason: choice.FinishReason,
			},
		},
		Created: openaiResp.Created,
	}
}

// createErrorResponse creates an error response
func (e *VLLMExecutor) createErrorResponse(model, message string) *pb.ChatCompletionResponse {
	return &pb.ChatCompletionResponse{
		Id:      e.generateID(),
		Model:   model,
		Object:  "error",
		Choices: []*pb.ChatChoice{{FinishReason: "error"}},
		Created: time.Now().Unix(),
	}
}

// generateID generates a unique ID for responses
func (e *VLLMExecutor) generateID() string {
	return fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
}
