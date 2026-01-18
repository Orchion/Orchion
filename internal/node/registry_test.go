package node

import (
	"testing"
	"time"

	orchionv1 "github.com/Orchion/Orchion/api/proto/v1"
)

func TestInMemoryRegistry_Register(t *testing.T) {
	registry := NewInMemoryRegistry()

	nodeID, err := registry.Register("localhost:5000", map[string]string{
		"name": "test-node",
	})

	if err != nil {
		t.Fatalf("Failed to register node: %v", err)
	}

	if nodeID == "" {
		t.Fatal("Expected non-empty node ID")
	}

	// Verify the node was registered
	node, err := registry.Get(nodeID)
	if err != nil {
		t.Fatalf("Failed to get registered node: %v", err)
	}

	if node.Address != "localhost:5000" {
		t.Errorf("Expected address 'localhost:5000', got '%s'", node.Address)
	}

	if node.Metadata["name"] != "test-node" {
		t.Errorf("Expected metadata name 'test-node', got '%s'", node.Metadata["name"])
	}

	if node.Status != orchionv1.NodeStatus_NODE_STATUS_ACTIVE {
		t.Errorf("Expected status ACTIVE, got %v", node.Status)
	}
}

func TestInMemoryRegistry_Unregister(t *testing.T) {
	registry := NewInMemoryRegistry()

	nodeID, _ := registry.Register("localhost:5000", nil)

	err := registry.Unregister(nodeID)
	if err != nil {
		t.Fatalf("Failed to unregister node: %v", err)
	}

	// Verify the node was removed
	_, err = registry.Get(nodeID)
	if err == nil {
		t.Fatal("Expected error when getting unregistered node")
	}
}

func TestInMemoryRegistry_List(t *testing.T) {
	registry := NewInMemoryRegistry()

	// Register multiple nodes
	registry.Register("localhost:5000", nil)
	registry.Register("localhost:5001", nil)
	registry.Register("localhost:5002", nil)

	nodes, err := registry.List(orchionv1.NodeStatus_NODE_STATUS_UNSPECIFIED)
	if err != nil {
		t.Fatalf("Failed to list nodes: %v", err)
	}

	if len(nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(nodes))
	}
}

func TestInMemoryRegistry_UpdateHeartbeat(t *testing.T) {
	registry := NewInMemoryRegistry()

	nodeID, _ := registry.Register("localhost:5000", nil)
	
	// Get initial heartbeat time
	node1, _ := registry.Get(nodeID)
	initialTime := node1.LastHeartbeat.AsTime()

	// Wait a bit
	time.Sleep(10 * time.Millisecond)

	// Update heartbeat
	err := registry.UpdateHeartbeat(nodeID)
	if err != nil {
		t.Fatalf("Failed to update heartbeat: %v", err)
	}

	// Get updated node
	node2, _ := registry.Get(nodeID)
	updatedTime := node2.LastHeartbeat.AsTime()

	if !updatedTime.After(initialTime) {
		t.Error("Expected heartbeat time to be updated")
	}
}

func TestInMemoryRegistry_MarkInactive(t *testing.T) {
	registry := NewInMemoryRegistry()

	nodeID, _ := registry.Register("localhost:5000", nil)

	// Get the node and manually set old heartbeat
	node, _ := registry.Get(nodeID)
	oldTime := time.Now().Add(-2 * time.Minute)
	node.LastHeartbeat.Seconds = oldTime.Unix()

	// Mark inactive nodes
	registry.MarkInactive(1 * time.Minute)

	// Verify status changed to inactive
	updatedNode, _ := registry.Get(nodeID)
	if updatedNode.Status != orchionv1.NodeStatus_NODE_STATUS_INACTIVE {
		t.Errorf("Expected status INACTIVE, got %v", updatedNode.Status)
	}
}
