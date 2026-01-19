package scheduler

import (
	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
)

// Scheduler selects nodes for model execution
type Scheduler interface {
	SelectNode(model string, registry node.Registry) (*pb.Node, error)
}

// SimpleScheduler is a basic scheduler that selects the first available node
type SimpleScheduler struct{}

// NewSimpleScheduler creates a new simple scheduler
func NewSimpleScheduler() *SimpleScheduler {
	return &SimpleScheduler{}
}

// SelectNode selects a node for the given model
// For now, it just picks the first available node
// TODO: Enhance to consider node capabilities, load, and model availability
func (s *SimpleScheduler) SelectNode(model string, registry node.Registry) (*pb.Node, error) {
	nodes := registry.List()
	if len(nodes) == 0 {
		return nil, ErrNoNodesAvailable
	}

	// For now, return the first node
	// In the future, this should:
	// 1. Filter nodes by model capability
	// 2. Consider node load/availability
	// 3. Use load balancing strategies
	return nodes[0], nil
}

var ErrNoNodesAvailable = &SchedulerError{Message: "no nodes available"}

type SchedulerError struct {
	Message string
}

func (e *SchedulerError) Error() string {
	return e.Message
}
