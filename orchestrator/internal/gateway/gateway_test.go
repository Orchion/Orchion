package gateway

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
)

func TestNewGateway(t *testing.T) {
	gateway := NewGateway("localhost:8080")
	assert.NotNil(t, gateway)
	assert.Equal(t, "localhost:8080", gateway.orchestratorAddr)
	assert.Empty(t, gateway.apiKey)
}

func TestGateway_SetAPIKey(t *testing.T) {
	gateway := NewGateway("localhost:8080")
	assert.Empty(t, gateway.apiKey)

	gateway.SetAPIKey("test-key")
	assert.Equal(t, "test-key", gateway.apiKey)
}

func TestGateway_authenticate(t *testing.T) {
	gateway := NewGateway("localhost:8080")

	// Test no API key required
	req := &http.Request{}
	assert.True(t, gateway.authenticate(req))

	// Set API key
	gateway.SetAPIKey("secret-key")

	// Test missing Authorization header
	assert.False(t, gateway.authenticate(req))

	// Test Bearer token format
	req.Header = make(http.Header)
	req.Header.Set("Authorization", "Bearer secret-key")
	assert.True(t, gateway.authenticate(req))

	// Test wrong Bearer token
	req.Header.Set("Authorization", "Bearer wrong-key")
	assert.False(t, gateway.authenticate(req))

	// Test sk- prefix format
	req.Header.Set("Authorization", "sk-secret-key")
	assert.True(t, gateway.authenticate(req))

	// Test wrong sk- token
	req.Header.Set("Authorization", "sk-wrong-key")
	assert.False(t, gateway.authenticate(req))

	// Test direct API key
	req.Header.Set("Authorization", "secret-key")
	assert.True(t, gateway.authenticate(req))
}

func TestGateway_convertChatCompletionRequest(t *testing.T) {
	gateway := NewGateway("localhost:8080")

	// Test successful conversion
	reqData := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Hello"},
		},
		"temperature": 0.7,
		"stream":     true,
		"max_tokens": 100.0,
	}

	grpcReq, err := gateway.convertChatCompletionRequest(reqData)
	require.NoError(t, err)
	assert.Equal(t, "gpt-3.5-turbo", grpcReq.Model)
	assert.Len(t, grpcReq.Messages, 1)
	assert.Equal(t, "user", grpcReq.Messages[0].Role)
	assert.Equal(t, "Hello", grpcReq.Messages[0].Content)
	assert.Equal(t, float32(0.7), grpcReq.Temperature)
	assert.True(t, grpcReq.Stream)
	assert.Equal(t, int32(100), grpcReq.MaxTokens)

	// Test missing model
	badReq := map[string]interface{}{
		"messages": []interface{}{
			map[string]interface{}{"role": "user", "content": "Hello"},
		},
	}
	_, err = gateway.convertChatCompletionRequest(badReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model is required")

	// Test missing messages
	badReq2 := map[string]interface{}{
		"model": "gpt-3.5-turbo",
	}
	_, err = gateway.convertChatCompletionRequest(badReq2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "messages are required")

	// Test invalid message format
	badReq3 := map[string]interface{}{
		"model": "gpt-3.5-turbo",
		"messages": []interface{}{
			"invalid-message",
		},
	}
	_, err = gateway.convertChatCompletionRequest(badReq3)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid message format")
}

func TestGateway_convertEmbeddingRequest(t *testing.T) {
	gateway := NewGateway("localhost:8080")

	// Test successful conversion with single input
	reqData := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": "Hello world",
	}

	grpcReq, err := gateway.convertEmbeddingRequest(reqData)
	require.NoError(t, err)
	assert.Equal(t, "text-embedding-ada-002", grpcReq.Model)
	assert.Equal(t, []string{"Hello world"}, grpcReq.Input)

	// Test successful conversion with multiple inputs
	reqData2 := map[string]interface{}{
		"model": "text-embedding-ada-002",
		"input": []interface{}{"Hello", "world"},
	}

	grpcReq2, err := gateway.convertEmbeddingRequest(reqData2)
	require.NoError(t, err)
	assert.Equal(t, []string{"Hello", "world"}, grpcReq2.Input)

	// Test missing model
	badReq := map[string]interface{}{
		"input": "Hello world",
	}
	_, err = gateway.convertEmbeddingRequest(badReq)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model is required")

	// Test missing input
	badReq2 := map[string]interface{}{
		"model": "text-embedding-ada-002",
	}
	_, err = gateway.convertEmbeddingRequest(badReq2)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "input is required")
}

func TestGateway_convertChatCompletionResponse(t *testing.T) {
	gateway := NewGateway("localhost:8080")

	// Mock gRPC response
	grpcResp := &pb.ChatCompletionResponse{
		Id:      "chatcmpl-123",
		Object:  "chat.completion",
		Created: 1677652288,
		Model:   "gpt-3.5-turbo",
		Choices: []*pb.ChatChoice{
			{
				Index: 0,
				Message: &pb.ChatMessage{
					Role:    "assistant",
					Content: "Hello there!",
				},
				FinishReason: "stop",
			},
		},
	}

	openaiResp := gateway.convertChatCompletionResponse(grpcResp)

	// Verify structure
	assert.Equal(t, "chatcmpl-123", openaiResp["id"])
	assert.Equal(t, "chat.completion", openaiResp["object"])
	assert.Equal(t, int64(1677652288), openaiResp["created"])
	assert.Equal(t, "gpt-3.5-turbo", openaiResp["model"])

	choices, ok := openaiResp["choices"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, choices, 1)

	choice := choices[0]
	assert.Equal(t, int32(0), choice["index"])

	message, ok := choice["message"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, "assistant", message["role"])
	assert.Equal(t, "Hello there!", message["content"])
	assert.Equal(t, "stop", choice["finish_reason"])
}

func TestGateway_convertEmbeddingResponse(t *testing.T) {
	gateway := NewGateway("localhost:8080")

	// Mock gRPC response
	grpcResp := &pb.EmbeddingResponse{
		Object: "list",
		Data: []*pb.Embedding{
			{
				Embedding: []float32{0.1, 0.2, 0.3},
				Index:     0,
			},
		},
		Model:            "text-embedding-ada-002",
		UsagePromptTokens: 2,
	}

	openaiResp := gateway.convertEmbeddingResponse(grpcResp)

	// Verify structure
	assert.Equal(t, "list", openaiResp["object"])
	assert.Equal(t, "text-embedding-ada-002", openaiResp["model"])

	data, ok := openaiResp["data"].([]map[string]interface{})
	require.True(t, ok)
	assert.Len(t, data, 1)

	embedding := data[0]
	assert.Equal(t, "embedding", embedding["object"])
	assert.Equal(t, int32(0), embedding["index"])

	// Check embedding vector conversion
	embVec, ok := embedding["embedding"].([]float64)
	require.True(t, ok)
	assert.Len(t, embVec, 3)
	assert.InDelta(t, 0.1, embVec[0], 0.01)
	assert.InDelta(t, 0.2, embVec[1], 0.01)
	assert.InDelta(t, 0.3, embVec[2], 0.01)

	usage, ok := openaiResp["usage"].(map[string]interface{})
	require.True(t, ok)
	assert.Equal(t, int32(2), usage["prompt_tokens"])
	assert.Equal(t, int32(2), usage["total_tokens"])
}

// Note: HTTP handler integration tests would require complex gRPC server mocking
// and are beyond the scope of basic unit tests. These tests focus on the core
// conversion and validation logic.

// Note: These tests would require more complex mocking of gRPC clients
// For now, we'll test the basic structure and conversion functions
// Full HTTP handler tests would require integration with a test gRPC server