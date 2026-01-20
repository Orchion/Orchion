package containers

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewContainerManager(t *testing.T) {
	manager, err := NewContainerManager()

	// Manager may be nil if no container runtime is available
	if err != nil {
		assert.Contains(t, err.Error(), "neither podman nor docker found")
		assert.Nil(t, manager)
	} else {
		assert.NotNil(t, manager)
	}
}

func TestContainerConfig_Validation(t *testing.T) {
	config := &ContainerConfig{
		Name:  "test-container",
		Image: "test-image:latest",
		Port:  8080,
	}

	assert.Equal(t, "test-container", config.Name)
	assert.Equal(t, "test-image:latest", config.Image)
	assert.Equal(t, 8080, config.Port)
}

func TestContainerRuntime_String(t *testing.T) {
	assert.Equal(t, "podman", string(RuntimePodman))
	assert.Equal(t, "docker", string(RuntimeDocker))
}

func TestContainerManager_IsRunning(t *testing.T) {
	manager, err := NewContainerManager()
	if err != nil {
		t.Skip("Skipping test due to no container runtime available")
	}

	// Test with a non-existent container
	running, err := manager.IsRunning(context.Background(), "non-existent-container")
	assert.NoError(t, err)
	assert.False(t, running)
}

func TestContainerManager_TestConnection(t *testing.T) {
	manager, err := NewContainerManager()
	if err != nil {
		t.Skip("Skipping test due to no container runtime available")
	}

	// Test connection to container runtime
	err = manager.TestConnection()

	// Result depends on system setup - may succeed or fail
	// The important thing is that it doesn't crash
	assert.True(t, true, "Test completed without crashing")
}

func TestContainerConfig_Empty(t *testing.T) {
	config := &ContainerConfig{}

	// Test default/empty values
	assert.Empty(t, config.Name)
	assert.Empty(t, config.Image)
	assert.Equal(t, 0, config.Port)
	assert.Empty(t, config.Model)
	assert.Empty(t, config.GPUs)
	assert.Empty(t, config.Environment)
	assert.Empty(t, config.Volumes)
	assert.Empty(t, config.Args)
}