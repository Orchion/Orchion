package node

import (
	"sync"
	"time"

	orchionv1 "github.com/Orchion/Orchion/api/proto/v1"
)

// Registry manages the collection of registered nodes
type Registry interface {
	// Register adds a new node to the registry
	Register(address string, metadata map[string]string) (string, error)
	
	// Unregister removes a node from the registry
	Unregister(nodeID string) error
	
	// Get retrieves information about a specific node
	Get(nodeID string) (*orchionv1.NodeInfo, error)
	
	// List returns all registered nodes
	List(statusFilter orchionv1.NodeStatus) ([]*orchionv1.NodeInfo, error)
	
	// UpdateHeartbeat updates the last heartbeat time for a node
	UpdateHeartbeat(nodeID string) error
	
	// MarkInactive marks nodes as inactive if they haven't sent heartbeat recently
	MarkInactive(timeout time.Duration)
}

// InMemoryRegistry is an in-memory implementation of the Registry interface
type InMemoryRegistry struct {
	mu    sync.RWMutex
	nodes map[string]*orchionv1.NodeInfo
	seq   uint64 // sequence number for generating node IDs
}

// NewInMemoryRegistry creates a new in-memory registry
func NewInMemoryRegistry() *InMemoryRegistry {
	return &InMemoryRegistry{
		nodes: make(map[string]*orchionv1.NodeInfo),
	}
}
