package executor

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

func TestNewService(t *testing.T) {
	service, err := NewService()

	// Service creation may succeed or fail depending on container runtime availability
	if err != nil {
		assert.Contains(t, err.Error(), "failed to create container manager")
		return
	}

	assert.NotNil(t, service)
	assert.NotNil(t, service.containerManager)
	assert.NotNil(t, service.executors)
	assert.NotNil(t, service.runningModels)
}

func TestService_BasicInitialization(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Skip("Skipping test due to container manager unavailability")
	}

	assert.NotNil(t, service)
	assert.NotNil(t, service.containerManager)
	assert.NotNil(t, service.executors)
	assert.NotNil(t, service.runningModels)
}

func TestService_Embeddings_Validation(t *testing.T) {
	service, err := NewService()
	if err != nil {
		t.Skip("Skipping test due to container manager unavailability")
	}

	req := &pb.EmbeddingRequest{
		Input: []string{"test"},
	}

	_, err = service.Embeddings(context.Background(), req)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "model is required")
}

func TestModelInstance_Structure(t *testing.T) {
	instance := &ModelInstance{
		Model:     "test-model",
		StartTime: time.Now(),
	}

	assert.Equal(t, "test-model", instance.Model)
	assert.NotZero(t, instance.StartTime)
	assert.True(t, instance.StartTime.Before(time.Now().Add(time.Second)))
}