package containers

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"strings"
)

// Manager handles container lifecycle for model servers
type Manager interface {
	StartContainer(ctx context.Context, config *ContainerConfig) error
	StopContainer(ctx context.Context, name string) error
	IsRunning(ctx context.Context, name string) (bool, error)
	EnsureRunning(ctx context.Context, config *ContainerConfig) error
}

// ContainerConfig defines configuration for a container
type ContainerConfig struct {
	Name        string
	Image       string
	Port        int
	Model       string   // For vLLM/Ollama
	GPUs        []string // GPU device IDs
	Environment []string // Environment variables
	Volumes     []string // Volume mounts
	Args        []string // Additional arguments
}

// DockerManager implements Manager using Docker CLI
type DockerManager struct {
	dockerPath string
}

// NewDockerManager creates a new Docker container manager
func NewDockerManager() (*DockerManager, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	return &DockerManager{
		dockerPath: dockerPath,
	}, nil
}

// StartContainer starts a container with the given configuration
func (m *DockerManager) StartContainer(ctx context.Context, config *ContainerConfig) error {
	// Check if already running
	running, err := m.IsRunning(ctx, config.Name)
	if err != nil {
		return err
	}
	if running {
		log.Printf("Container %s is already running", config.Name)
		return nil
	}

	// Stop and remove existing container if it exists
	_ = m.StopContainer(ctx, config.Name)

	// Build docker run command
	args := []string{"run", "-d", "--name", config.Name}

	// Port mapping
	if config.Port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", config.Port, config.Port))
	}

	// GPU support
	if len(config.GPUs) > 0 {
		gpuStr := strings.Join(config.GPUs, ",")
		args = append(args, "--gpus", fmt.Sprintf("device=%s", gpuStr))
	}

	// Environment variables
	for _, env := range config.Environment {
		args = append(args, "-e", env)
	}

	// Volume mounts
	for _, vol := range config.Volumes {
		args = append(args, "-v", vol)
	}

	// Additional args
	args = append(args, config.Args...)

	// Image
	args = append(args, config.Image)

	log.Printf("Starting container %s: docker %s", config.Name, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, m.dockerPath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container %s: %w\nOutput: %s", config.Name, err, string(output))
	}

	log.Printf("Container %s started successfully", config.Name)
	return nil
}

// StopContainer stops and removes a container
func (m *DockerManager) StopContainer(ctx context.Context, name string) error {
	// Stop container
	stopCmd := exec.CommandContext(ctx, m.dockerPath, "stop", name)
	_ = stopCmd.Run() // Ignore errors if not running

	// Remove container
	rmCmd := exec.CommandContext(ctx, m.dockerPath, "rm", name)
	_ = rmCmd.Run() // Ignore errors if doesn't exist

	return nil
}

// IsRunning checks if a container is running
func (m *DockerManager) IsRunning(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, m.dockerPath, "ps", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) == name, nil
}

// EnsureRunning ensures a container is running, starting it if necessary
func (m *DockerManager) EnsureRunning(ctx context.Context, config *ContainerConfig) error {
	running, err := m.IsRunning(ctx, config.Name)
	if err != nil {
		return err
	}

	if !running {
		return m.StartContainer(ctx, config)
	}

	return nil
}