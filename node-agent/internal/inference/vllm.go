package inference

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// VLLMEngine implements Engine using vLLM OpenAI-compatible API
type VLLMEngine struct {
	baseURL string
	client  *http.Client
}

// NewVLLMEngine creates a new vLLM engine
func NewVLLMEngine(baseURL string) *VLLMEngine {
	if !strings.HasPrefix(baseURL, "http") {
		baseURL = "http://" + baseURL
	}
	return &VLLMEngine{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 5 * time.Minute,
		},
	}
}

// ChatCompletion implements Engine.ChatCompletion
func (e *VLLMEngine) ChatCompletion(ctx context.Context, req *pb.ChatCompletionRequest) (<-chan *pb.ChatCompletionResponse, error) {
	responseChan := make(chan *pb.ChatCompletionResponse, 10)

	go func() {
		defer close(responseChan)

		// Convert messages to OpenAI format
		messages := make([]map[string]string, len(req.Messages))
		for i, msg := range req.Messages {
			messages[i] = map[string]string{
				"role":    msg.Role,
				"content": msg.Content,
			}
		}

		// Build OpenAI-compatible request
		openaiReq := map[string]interface{}{
			"model":    req.Model,
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
			responseChan <- &pb.ChatCompletionResponse{
				Id:      generateID(),
				Model:   req.Model,
				Object:  "error",
				Choices: []*pb.ChatChoice{{FinishReason: "error"}},
			}
			return
		}

		// Make request to vLLM
		httpReq, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/v1/chat/completions", bytes.NewReader(reqBody))
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
			// Stream SSE responses
			scanner := bufio.NewScanner(resp.Body)
			for scanner.Scan() {
				line := scanner.Text()
				if !strings.HasPrefix(line, "data: ") {
					continue
				}

				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					break
				}

				var openaiResp map[string]interface{}
				if err := json.Unmarshal([]byte(data), &openaiResp); err != nil {
					continue
				}

				// Convert OpenAI response to our format
				choices, _ := openaiResp["choices"].([]interface{})
				if len(choices) == 0 {
					continue
				}

				choice, _ := choices[0].(map[string]interface{})
				delta, _ := choice["delta"].(map[string]interface{})
				content, _ := delta["content"].(string)
				finishReason, _ := choice["finish_reason"].(string)

				responseChan <- &pb.ChatCompletionResponse{
					Id:     fmt.Sprintf("%v", openaiResp["id"]),
					Model:  fmt.Sprintf("%v", openaiResp["model"]),
					Object: "chat.completion.chunk",
					Choices: []*pb.ChatChoice{
						{
							Index: 0,
							Message: &pb.ChatMessage{
								Role:    "assistant",
								Content: content,
							},
							FinishReason: finishReason,
						},
					},
					Created: func() int64 {
						if created, ok := openaiResp["created"].(float64); ok {
							return int64(created)
						}
						return time.Now().Unix()
					}(),
				}

				if finishReason != "" {
					break
				}
			}
		} else {
			// Non-streaming response
			var openaiResp map[string]interface{}
			if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
				responseChan <- &pb.ChatCompletionResponse{
					Id:      generateID(),
					Model:   req.Model,
					Object:  "error",
					Choices: []*pb.ChatChoice{{FinishReason: "error"}},
				}
				return
			}

			choices, _ := openaiResp["choices"].([]interface{})
			if len(choices) == 0 {
				responseChan <- &pb.ChatCompletionResponse{
					Id:      generateID(),
					Model:   req.Model,
					Object:  "error",
					Choices: []*pb.ChatChoice{{FinishReason: "error"}},
				}
				return
			}

			choice, _ := choices[0].(map[string]interface{})
			message, _ := choice["message"].(map[string]interface{})
			content, _ := message["content"].(string)
			finishReason, _ := choice["finish_reason"].(string)

			responseChan <- &pb.ChatCompletionResponse{
				Id:     fmt.Sprintf("%v", openaiResp["id"]),
				Model:  fmt.Sprintf("%v", openaiResp["model"]),
				Object: "chat.completion",
				Choices: []*pb.ChatChoice{
					{
						Index: 0,
						Message: &pb.ChatMessage{
							Role:    "assistant",
							Content: content,
						},
						FinishReason: finishReason,
					},
				},
				Created: func() int64 {
					if created, ok := openaiResp["created"].(float64); ok {
						return int64(created)
					}
					return time.Now().Unix()
				}(),
			}
		}
	}()

	return responseChan, nil
}

// Embeddings implements Engine.Embeddings
func (e *VLLMEngine) Embeddings(ctx context.Context, req *pb.EmbeddingRequest) (*pb.EmbeddingResponse, error) {
	openaiReq := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}

	reqBody, err := json.Marshal(openaiReq)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", e.baseURL+"/v1/embeddings", bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := e.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to call vLLM: %w", err)
	}
	defer resp.Body.Close()

	var openaiResp map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&openaiResp); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	data, _ := openaiResp["data"].([]interface{})
	embeddings := make([]*pb.Embedding, len(data))

	for i, item := range data {
		itemMap, _ := item.(map[string]interface{})
		embeddingSlice, _ := itemMap["embedding"].([]interface{})

		embedding := make([]float32, len(embeddingSlice))
		for j, v := range embeddingSlice {
			if f, ok := v.(float64); ok {
				embedding[j] = float32(f)
			}
		}

		embeddings[i] = &pb.Embedding{
			Index:     int32(i),
			Embedding: embedding,
		}
	}

	usage, _ := openaiResp["usage"].(map[string]interface{})
	promptTokens := 0
	if pt, ok := usage["prompt_tokens"].(float64); ok {
		promptTokens = int(pt)
	}

	return &pb.EmbeddingResponse{
		Model:             fmt.Sprintf("%v", openaiResp["model"]),
		Object:            "list",
		Data:              embeddings,
		UsagePromptTokens: int32(promptTokens),
	}, nil
}
