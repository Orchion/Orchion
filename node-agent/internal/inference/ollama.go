package inference

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// OllamaEngine implements Engine using Ollama API
type OllamaEngine struct {
	baseURL string
	client  *http.Client
}

// NewOllamaEngine creates a new Ollama engine
func NewOllamaEngine(baseURL string) *OllamaEngine {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return &OllamaEngine{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// ChatCompletion implements Engine.ChatCompletion
func (e *OllamaEngine) ChatCompletion(ctx context.Context, req *pb.ChatCompletionRequest) (<-chan *pb.ChatCompletionResponse, error) {
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
			"model":    req.Model,
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
			responseChan <- &pb.ChatCompletionResponse{
				Id:      generateID(),
				Model:   req.Model,
				Object:  "error",
				Choices: []*pb.ChatChoice{{FinishReason: "error"}},
			}
			return
		}

		// Make request to Ollama
		httpReq, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/chat", bytes.NewReader(reqBody))
		if err != nil {
			responseChan <- &pb.ChatCompletionResponse{
				Id:      generateID(),
				Model:   req.Model,
				Object:  "error",
				Choices: []*pb.ChatChoice{{FinishReason: "error"}},
			}
			return
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(httpReq)
		if err != nil {
			responseChan <- &pb.ChatCompletionResponse{
				Id:      generateID(),
				Model:   req.Model,
				Object:  "error",
				Choices: []*pb.ChatChoice{{FinishReason: "error"}},
			}
			return
		}
		defer resp.Body.Close()

		if req.Stream {
			// Stream responses
			decoder := json.NewDecoder(resp.Body)
			for {
				var ollamaResp map[string]interface{}
				if err := decoder.Decode(&ollamaResp); err != nil {
					if err == io.EOF {
						break
					}
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
					Id:     generateID(),
					Model:  req.Model,
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
		} else {
			// Non-streaming response
			var ollamaResp map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
				responseChan <- &pb.ChatCompletionResponse{
					Id:      generateID(),
					Model:   req.Model,
					Object:  "error",
					Choices: []*pb.ChatChoice{{FinishReason: "error"}},
				}
				return
			}

			message, _ := ollamaResp["message"].(map[string]interface{})
			content, _ := message["content"].(string)

			responseChan <- &pb.ChatCompletionResponse{
				Id:     generateID(),
				Model:  req.Model,
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
	}()

	return responseChan, nil
}

// Embeddings implements Engine.Embeddings
func (e *OllamaEngine) Embeddings(ctx context.Context, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	// Ollama embeddings endpoint
	embeddings := make([]*pb.Embedding, 0, len(req.Input))

	for i, input := range req.Input {
		ollamaReq := map[string]interface{}{
			"model":  req.Model,
			"prompt": input,
		}

		reqBody, err := json.Marshal(ollamaReq)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		httpReq, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/api/embeddings", bytes.NewReader(reqBody))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		httpReq.Header.Set("Content-Type", "application/json")

		resp, err := e.client.Do(httpReq)
		if err != nil {
			return nil, fmt.Errorf("failed to call Ollama: %w", err)
		}
		defer resp.Body.Close()

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
		Model:  req.Model,
		Object: "list",
		Data:   embeddings,
	}, nil
}

func generateID() string {
	return fmt.Sprintf("chatcmpl-%d", time.Now().UnixNano())
}
