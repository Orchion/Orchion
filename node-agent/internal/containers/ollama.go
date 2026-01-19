package containers

import (
	"context"
	"fmt"
	"os/exec"
)

// OllamaConfig holds configuration for Ollama container
type OllamaConfig struct {
	Model string
	Port  int
	GPUs  []string
}

// DefaultOllamaConfig returns default Ollama configuration
func DefaultOllamaConfig() *OllamaConfig {
	return &OllamaConfig{
		Model: "llama2",
		Port:  11434,
		GPUs:  []string{"all"},
	}
}

// CreateOllamaContainerConfig creates a ContainerConfig for Ollama
func CreateOllamaContainerConfig(cfg *OllamaConfig) *ContainerConfig {
	name := "orchion-ollama"

	return &ContainerConfig{
		Name:  name,
		Image: "ollama/ollama:latest",
		Port:  cfg.Port,
		Model: cfg.Model,
		GPUs:  cfg.GPUs,
		Volumes: []string{
			"ollama-data:/root/.ollama",
		},
		Environment: []string{
			"OLLAMA_HOST=0.0.0.0",
		},
	}
}

// PullOllamaModel pulls a model into the Ollama container
// This would be done separately after container starts
func PullOllamaModel(ctx context.Context, manager Manager, containerName, model string) error {
	// Type assert to get the runtime path
	var runtimePath string
	if cm, ok := manager.(*ContainerManager); ok {
		runtimePath = cm.runtimePath
	} else {
		return fmt.Errorf("unsupported manager type for PullOllamaModel")
	}

	cmd := exec.CommandContext(ctx, runtimePath, "exec", containerName, "ollama", "pull", model)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to pull Ollama model %s: %w\nOutput: %s", model, err, string(output))
	}
	return nil
}
