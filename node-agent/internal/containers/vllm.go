package containers

import (
	"fmt"
	"strings"
)

// VLLMConfig holds configuration for vLLM container
type VLLMConfig struct {
	Model       string
	Port        int
	GPUs        []string
	TensorParallelSize int
	MaxModelLen int
}

// DefaultVLLMConfig returns default vLLM configuration
func DefaultVLLMConfig() *VLLMConfig {
	return &VLLMConfig{
		Model:       "mistralai/Mistral-7B-Instruct-v0.1",
		Port:        8000,
		GPUs:        []string{"all"},
		TensorParallelSize: 1,
		MaxModelLen: 4096,
	}
}

// CreateVLLMContainerConfig creates a ContainerConfig for vLLM
func CreateVLLMContainerConfig(cfg *VLLMConfig) *ContainerConfig {
	name := fmt.Sprintf("orchion-vllm-%s", sanitizeModelName(cfg.Model))
	
	// Build vLLM command arguments
	args := []string{
		"--model", cfg.Model,
		"--port", fmt.Sprintf("%d", cfg.Port),
		"--host", "0.0.0.0",
	}

	if cfg.TensorParallelSize > 1 {
		args = append(args, "--tensor-parallel-size", fmt.Sprintf("%d", cfg.TensorParallelSize))
	}

	if cfg.MaxModelLen > 0 {
		args = append(args, "--max-model-len", fmt.Sprintf("%d", cfg.MaxModelLen))
	}

	return &ContainerConfig{
		Name:        name,
		Image:       "vllm/vllm-openai:latest",
		Port:        cfg.Port,
		Model:       cfg.Model,
		GPUs:        cfg.GPUs,
		Args:        args,
		Environment: []string{
			"VLLM_USE_MODELSCOPE=false",
		},
	}
}

// sanitizeModelName converts model name to container-friendly string
func sanitizeModelName(model string) string {
	// Replace slashes and special chars with dashes
	sanitized := strings.ReplaceAll(model, "/", "-")
	sanitized = strings.ReplaceAll(sanitized, ":", "-")
	sanitized = strings.ReplaceAll(sanitized, "_", "-")
	return sanitized
}