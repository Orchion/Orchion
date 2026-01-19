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
	TestConnection() error
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

// ContainerRuntime represents the type of container runtime
type ContainerRuntime string

const (
	RuntimePodman ContainerRuntime = "podman"
	RuntimeDocker ContainerRuntime = "docker"
)

// ContainerManager implements Manager using container CLI (Podman/Docker)
type ContainerManager struct {
	runtime     ContainerRuntime
	runtimePath string
}

// NewContainerManager creates a new container manager, preferring Podman over Docker
func NewContainerManager() (Manager, error) {
	// Try Podman first (preferred)
	if podmanPath, err := exec.LookPath("podman"); err == nil {
		log.Printf("Using Podman as container runtime")
		return &ContainerManager{
			runtime:     RuntimePodman,
			runtimePath: podmanPath,
		}, nil
	}

	// Fall back to Docker
	if dockerPath, err := exec.LookPath("docker"); err == nil {
		log.Printf("Using Docker as container runtime")
		return &ContainerManager{
			runtime:     RuntimeDocker,
			runtimePath: dockerPath,
		}, nil
	}

	return nil, fmt.Errorf("neither podman nor docker found in PATH")
}

// NewPodmanManager creates a container manager specifically for Podman
func NewPodmanManager() (Manager, error) {
	podmanPath, err := exec.LookPath("podman")
	if err != nil {
		return nil, fmt.Errorf("podman not found in PATH: %w", err)
	}

	return &ContainerManager{
		runtime:     RuntimePodman,
		runtimePath: podmanPath,
	}, nil
}

// NewDockerManager creates a container manager specifically for Docker (legacy)
func NewDockerManager() (Manager, error) {
	dockerPath, err := exec.LookPath("docker")
	if err != nil {
		return nil, fmt.Errorf("docker not found in PATH: %w", err)
	}

	return &ContainerManager{
		runtime:     RuntimeDocker,
		runtimePath: dockerPath,
	}, nil
}

// StartContainer starts a container with the given configuration
func (m *ContainerManager) StartContainer(ctx context.Context, config *ContainerConfig) error {
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

	// Build container run command
	args := []string{"run", "-d", "--name", config.Name}

	// Port mapping
	if config.Port > 0 {
		args = append(args, "-p", fmt.Sprintf("%d:%d", config.Port, config.Port))
	}

	// GPU support (different syntax for Podman vs Docker)
	if len(config.GPUs) > 0 {
		if m.runtime == RuntimePodman {
			// Podman GPU support
			for _, gpu := range config.GPUs {
				if gpu == "all" {
					args = append(args, "--device", "nvidia.com/gpu=all")
				} else {
					args = append(args, "--device", fmt.Sprintf("nvidia.com/gpu=%s", gpu))
				}
			}
		} else {
			// Docker GPU support
			gpuStr := strings.Join(config.GPUs, ",")
			args = append(args, "--gpus", fmt.Sprintf("device=%s", gpuStr))
		}
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

	runtimeName := string(m.runtime)
	log.Printf("Starting container %s: %s %s", config.Name, runtimeName, strings.Join(args, " "))

	cmd := exec.CommandContext(ctx, m.runtimePath, args...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("failed to start container %s: %w\nOutput: %s", config.Name, err, string(output))
	}

	log.Printf("Container %s started successfully", config.Name)
	return nil
}

// StopContainer stops and removes a container
func (m *ContainerManager) StopContainer(ctx context.Context, name string) error {
	// Stop container
	stopCmd := exec.CommandContext(ctx, m.runtimePath, "stop", name)
	_ = stopCmd.Run() // Ignore errors if not running

	// Remove container
	rmCmd := exec.CommandContext(ctx, m.runtimePath, "rm", name)
	_ = rmCmd.Run() // Ignore errors if doesn't exist

	return nil
}

// IsRunning checks if a container is running
func (m *ContainerManager) IsRunning(ctx context.Context, name string) (bool, error) {
	cmd := exec.CommandContext(ctx, m.runtimePath, "ps", "--filter", fmt.Sprintf("name=%s", name), "--format", "{{.Names}}")
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) == name, nil
}

// EnsureRunning ensures a container is running, starting it if necessary
func (m *ContainerManager) EnsureRunning(ctx context.Context, config *ContainerConfig) error {
	running, err := m.IsRunning(ctx, config.Name)
	if err != nil {
		return err
	}

	if !running {
		return m.StartContainer(ctx, config)
	}

	return nil
}

// TestConnection tests if the container runtime is available and working
func (m *ContainerManager) TestConnection() error {
	cmd := exec.Command(m.runtimePath, "version")
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%s not available: %w\nOutput: %s", m.runtime, err, string(output))
	}
	return nil
}
