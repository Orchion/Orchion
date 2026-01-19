package gateway

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
)

// Gateway handles HTTP requests and converts them to gRPC
type Gateway struct {
	orchestratorAddr string
	apiKey           string // Optional API key for authentication
}

// NewGateway creates a new gateway
func NewGateway(orchestratorAddr string) *Gateway {
	return &Gateway{
		orchestratorAddr: orchestratorAddr,
	}
}

// SetAPIKey sets an optional API key for authentication
func (g *Gateway) SetAPIKey(apiKey string) {
	g.apiKey = apiKey
}

// authenticate checks if the request is authenticated (if API key is set)
func (g *Gateway) authenticate(r *http.Request) bool {
	if g.apiKey == "" {
		return true // No authentication required
	}

	// Check Authorization header: "Bearer <key>" or "sk-<key>"
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		return false
	}

	// Support both "Bearer <key>" and "sk-<key>" formats
	if strings.HasPrefix(authHeader, "Bearer ") {
		key := strings.TrimPrefix(authHeader, "Bearer ")
		return key == g.apiKey
	}
	if strings.HasPrefix(authHeader, "sk-") {
		key := strings.TrimPrefix(authHeader, "sk-")
		return key == g.apiKey
	}

	return authHeader == g.apiKey
}

// ChatCompletionsHandler handles /v1/chat/completions
func (g *Gateway) ChatCompletionsHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication if API key is set
	if !g.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse OpenAI request
	var openaiReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&openaiReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to gRPC request
	grpcReq, err := g.convertChatCompletionRequest(openaiReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Connect to orchestrator
	conn, err := grpc.NewClient(g.orchestratorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to orchestrator: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewOrchionLLMClient(conn)
	stream, err := client.ChatCompletion(r.Context(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to call orchestrator: %v", err), http.StatusInternalServerError)
		return
	}

	// Stream responses
	if grpcReq.Stream {
		g.streamSSE(w, stream)
	} else {
		g.sendNonStreamingResponse(w, stream)
	}
}

// EmbeddingsHandler handles /v1/embeddings
func (g *Gateway) EmbeddingsHandler(w http.ResponseWriter, r *http.Request) {
	// CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusOK)
		return
	}

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check authentication if API key is set
	if !g.authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Parse OpenAI request
	var openaiReq map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&openaiReq); err != nil {
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	// Convert to gRPC request
	grpcReq, err := g.convertEmbeddingRequest(openaiReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Invalid request: %v", err), http.StatusBadRequest)
		return
	}

	// Connect to orchestrator
	conn, err := grpc.NewClient(g.orchestratorAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to connect to orchestrator: %v", err), http.StatusInternalServerError)
		return
	}
	defer conn.Close()

	client := pb.NewOrchionLLMClient(conn)
	resp, err := client.Embeddings(r.Context(), grpcReq)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to call orchestrator: %v", err), http.StatusInternalServerError)
		return
	}

	// Convert to OpenAI format
	openaiResp := g.convertEmbeddingResponse(resp)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}

// convertChatCompletionRequest converts OpenAI request to gRPC
func (g *Gateway) convertChatCompletionRequest(req map[string]interface{}) (*pb.ChatCompletionRequest, error) {
	grpcReq := &pb.ChatCompletionRequest{}

	// Model
	if model, ok := req["model"].(string); ok {
		grpcReq.Model = model
	} else {
		return nil, fmt.Errorf("model is required")
	}

	// Messages
	if messages, ok := req["messages"].([]interface{}); ok {
		grpcReq.Messages = make([]*pb.ChatMessage, len(messages))
		for i, msg := range messages {
			msgMap, ok := msg.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("invalid message format")
			}
			grpcReq.Messages[i] = &pb.ChatMessage{
				Role:    fmt.Sprintf("%v", msgMap["role"]),
				Content: fmt.Sprintf("%v", msgMap["content"]),
			}
		}
	} else {
		return nil, fmt.Errorf("messages are required")
	}

	// Temperature
	if temp, ok := req["temperature"].(float64); ok {
		grpcReq.Temperature = float32(temp)
	}

	// Stream
	if stream, ok := req["stream"].(bool); ok {
		grpcReq.Stream = stream
	}

	// Max tokens
	if maxTokens, ok := req["max_tokens"].(float64); ok {
		grpcReq.MaxTokens = int32(maxTokens)
	}

	return grpcReq, nil
}

// convertEmbeddingRequest converts OpenAI request to gRPC
func (g *Gateway) convertEmbeddingRequest(req map[string]interface{}) (*pb.EmbeddingRequest, error) {
	grpcReq := &pb.EmbeddingRequest{}

	// Model
	if model, ok := req["model"].(string); ok {
		grpcReq.Model = model
	} else {
		return nil, fmt.Errorf("model is required")
	}

	// Input
	if input, ok := req["input"].(string); ok {
		grpcReq.Input = []string{input}
	} else if inputs, ok := req["input"].([]interface{}); ok {
		grpcReq.Input = make([]string, len(inputs))
		for i, inp := range inputs {
			grpcReq.Input[i] = fmt.Sprintf("%v", inp)
		}
	} else {
		return nil, fmt.Errorf("input is required")
	}

	return grpcReq, nil
}

// streamSSE streams Server-Sent Events
func (g *Gateway) streamSSE(w http.ResponseWriter, stream pb.OrchionLLM_ChatCompletionClient) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")

	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if err == io.EOF || err == context.Canceled {
				fmt.Fprintf(w, "data: [DONE]\n\n")
				flusher.Flush()
				return
			}
			fmt.Fprintf(w, "data: {\"error\":\"%v\"}\n\n", err)
			flusher.Flush()
			return
		}

		// Convert to OpenAI SSE format
		openaiResp := g.convertChatCompletionResponse(resp)
		data, _ := json.Marshal(openaiResp)
		fmt.Fprintf(w, "data: %s\n\n", data)
		flusher.Flush()

		// Check if finished
		if len(resp.Choices) > 0 && resp.Choices[0].FinishReason != "" {
			fmt.Fprintf(w, "data: [DONE]\n\n")
			flusher.Flush()
			return
		}
	}
}

// sendNonStreamingResponse sends a single response
func (g *Gateway) sendNonStreamingResponse(w http.ResponseWriter, stream pb.OrchionLLM_ChatCompletionClient) {
	resp, err := stream.Recv()
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to receive response: %v", err), http.StatusInternalServerError)
		return
	}

	openaiResp := g.convertChatCompletionResponse(resp)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(openaiResp)
}

// convertChatCompletionResponse converts gRPC response to OpenAI format
func (g *Gateway) convertChatCompletionResponse(resp *pb.ChatCompletionResponse) map[string]interface{} {
	choices := make([]map[string]interface{}, len(resp.Choices))
	for i, choice := range resp.Choices {
		choiceMap := map[string]interface{}{
			"index": choice.Index,
		}

		if resp.Object == "chat.completion.chunk" {
			// Streaming format
			choiceMap["delta"] = map[string]interface{}{
				"role":    choice.Message.Role,
				"content": choice.Message.Content,
			}
		} else {
			// Non-streaming format
			choiceMap["message"] = map[string]interface{}{
				"role":    choice.Message.Role,
				"content": choice.Message.Content,
			}
		}

		if choice.FinishReason != "" {
			choiceMap["finish_reason"] = choice.FinishReason
		}

		choices[i] = choiceMap
	}

	return map[string]interface{}{
		"id":      resp.Id,
		"object":  resp.Object,
		"created": resp.Created,
		"model":   resp.Model,
		"choices": choices,
	}
}

// convertEmbeddingResponse converts gRPC response to OpenAI format
func (g *Gateway) convertEmbeddingResponse(resp *pb.EmbeddingResponse) map[string]interface{} {
	data := make([]map[string]interface{}, len(resp.Data))
	for i, emb := range resp.Data {
		// Convert float32 to float64 for JSON
		embedding64 := make([]float64, len(emb.Embedding))
		for j, v := range emb.Embedding {
			embedding64[j] = float64(v)
		}

		data[i] = map[string]interface{}{
			"object":    "embedding",
			"embedding": embedding64,
			"index":     emb.Index,
		}
	}

	return map[string]interface{}{
		"object": resp.Object,
		"data":   data,
		"model":  resp.Model,
		"usage": map[string]interface{}{
			"prompt_tokens": resp.UsagePromptTokens,
			"total_tokens":  resp.UsagePromptTokens,
		},
	}
}
