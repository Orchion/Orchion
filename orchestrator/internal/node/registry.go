package node

import (
	"sync"
	"time"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
)

// Registry manages registered nodes and their state
type Registry interface {
	Register(node *pb.Node) error
	UpdateCapabilities(nodeID string, capabilities *pb.Capabilities) error
	UpdateHeartbeat(nodeID string) error
	List() []*pb.Node
	Get(nodeID string) (*pb.Node, bool)
	Remove(nodeID string) error
	CheckHeartbeats(timeout time.Duration) []string // Returns IDs of stale nodes
}

// InMemoryRegistry is an in-memory implementation of Registry
type InMemoryRegistry struct {
	mu    sync.RWMutex
	nodes map[string]*pb.Node
}

// NewInMemoryRegistry creates a new in-memory node registry
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		nodes: make(map[string]*pb.Node),
	}
}

// Register adds or updates a node in the registry
func (r *InMemoryRegistry) Register(node *pb.Node) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node.LastSeenUnix == 0 {
		node.LastSeenUnix = time.Now().Unix()
	}

	r.nodes[node.Id] = node
	return nil
}

// UpdateCapabilities updates the capabilities for a node
func (r *InMemoryRegistry) UpdateCapabilities(nodeID string, capabilities *pb.Capabilities) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node, exists := r.nodes[nodeID]; exists {
		node.Capabilities = capabilities
		node.LastSeenUnix = time.Now().Unix()
		return nil
	}

	return ErrNodeNotFound
}

// UpdateHeartbeat updates the last seen timestamp for a node
func (r *InMemoryRegistry) UpdateHeartbeat(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if node, exists := r.nodes[nodeID]; exists {
		node.LastSeenUnix = time.Now().Unix()
		return nil
	}

	return ErrNodeNotFound
}

// List returns all registered nodes
func (r *InMemoryRegistry) List() []*pb.Node {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*pb.Node, 0, len(r.nodes))
	for _, node := range r.nodes {
		// Return a copy to avoid race conditions
		nodes = append(nodes, &pb.Node{
			Id:           node.Id,
			Hostname:     node.Hostname,
			Capabilities: node.Capabilities,
			LastSeenUnix: node.LastSeenUnix,
			AgentAddress: node.AgentAddress,
		})
	}
	return nodes
}

// Get retrieves a node by ID
func (r *InMemoryRegistry) Get(nodeID string) (*pb.Node, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return nil, false
	}

	// Return a copy
	return &pb.Node{
		Id:           node.Id,
		Hostname:     node.Hostname,
		Capabilities: node.Capabilities,
		LastSeenUnix: node.LastSeenUnix,
		AgentAddress: node.AgentAddress,
	}, true
}

// Remove removes a node from the registry
func (r *InMemoryRegistry) Remove(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeID]; !exists {
		return ErrNodeNotFound
	}

	delete(r.nodes, nodeID)
	return nil
}

// CheckHeartbeats returns IDs of nodes that haven't sent a heartbeat within the timeout
func (r *InMemoryRegistry) CheckHeartbeats(timeout time.Duration) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	now := time.Now().Unix()
	timeoutSeconds := int64(timeout.Seconds())
	stale := []string{}

	for id, node := range r.nodes {
		if now-node.LastSeenUnix > timeoutSeconds {
			stale = append(stale, id)
		}
	}

	return stale
}

var ErrNodeNotFound = &RegistryError{Message: "node not found"}

type RegistryError struct {
	Message string
}

func (e *RegistryError) Error() string {
	return e.Message
}
