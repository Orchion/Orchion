package node

import (
	"fmt"
	"time"

	orchionv1 "github.com/Orchion/Orchion/api/proto/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Register adds a new node to the registry
func (r *InMemoryRegistry) Register(address string, metadata map[string]string) (string, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Generate unique node ID
	r.seq++
	nodeID := fmt.Sprintf("node-%d", r.seq)

	now := timestamppb.Now()
	node := &orchionv1.NodeInfo{
		NodeId:        nodeID,
		Address:       address,
		Metadata:      metadata,
		Status:        orchionv1.NodeStatus_NODE_STATUS_ACTIVE,
		RegisteredAt:  now,
		LastHeartbeat: now,
	}

	r.nodes[nodeID] = node
	return nodeID, nil
}

// Unregister removes a node from the registry
func (r *InMemoryRegistry) Unregister(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.nodes[nodeID]; !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	delete(r.nodes, nodeID)
	return nil
}

// Get retrieves information about a specific node
func (r *InMemoryRegistry) Get(nodeID string) (*orchionv1.NodeInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return nil, fmt.Errorf("node %s not found", nodeID)
	}

	return node, nil
}

// List returns all registered nodes
func (r *InMemoryRegistry) List(statusFilter orchionv1.NodeStatus) ([]*orchionv1.NodeInfo, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]*orchionv1.NodeInfo, 0, len(r.nodes))
	for _, node := range r.nodes {
		// If no filter specified (UNSPECIFIED), return all nodes
		if statusFilter == orchionv1.NodeStatus_NODE_STATUS_UNSPECIFIED || node.Status == statusFilter {
			nodes = append(nodes, node)
		}
	}

	return nodes, nil
}

// UpdateHeartbeat updates the last heartbeat time for a node
func (r *InMemoryRegistry) UpdateHeartbeat(nodeID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	node, exists := r.nodes[nodeID]
	if !exists {
		return fmt.Errorf("node %s not found", nodeID)
	}

	node.LastHeartbeat = timestamppb.Now()
	node.Status = orchionv1.NodeStatus_NODE_STATUS_ACTIVE
	return nil
}

// MarkInactive marks nodes as inactive if they haven't sent heartbeat recently
func (r *InMemoryRegistry) MarkInactive(timeout time.Duration) {
	r.mu.Lock()
	defer r.mu.Unlock()

	now := time.Now()
	for _, node := range r.nodes {
		if node.Status == orchionv1.NodeStatus_NODE_STATUS_ACTIVE {
			lastHeartbeat := node.LastHeartbeat.AsTime()
			if now.Sub(lastHeartbeat) > timeout {
				node.Status = orchionv1.NodeStatus_NODE_STATUS_INACTIVE
			}
		}
	}
}
