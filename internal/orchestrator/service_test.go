package orchestrator

import (
	"context"
	"testing"
	"time"

	orchionv1 "github.com/Orchion/Orchion/api/proto/v1"
	"github.com/Orchion/Orchion/internal/node"
)

func TestService_RegisterNode(t *testing.T) {
	registry := node.NewInMemoryRegistry()
	svc := NewService(registry, 30*time.Second)

	req := &orchionv1.RegisterNodeRequest{
		Address: "localhost:5000",
		Metadata: map[string]string{
			"name": "test-node",
		},
	}

	resp, err := svc.RegisterNode(context.Background(), req)
	if err != nil {
		t.Fatalf("Failed to register node: %v", err)
	}

	if resp.NodeId == "" {
		t.Fatal("Expected non-empty node ID")
	}
}

func TestService_Heartbeat(t *testing.T) {
	registry := node.NewInMemoryRegistry()
	svc := NewService(registry, 30*time.Second)

	// First register a node
	registerReq := &orchionv1.RegisterNodeRequest{
		Address: "localhost:5000",
	}
	registerResp, _ := svc.RegisterNode(context.Background(), registerReq)

	// Send heartbeat
	heartbeatReq := &orchionv1.HeartbeatRequest{
		NodeId: registerResp.NodeId,
	}

	resp, err := svc.Heartbeat(context.Background(), heartbeatReq)
	if err != nil {
		t.Fatalf("Failed to send heartbeat: %v", err)
	}

	if !resp.Acknowledged {
		t.Error("Expected heartbeat to be acknowledged")
	}
}

func TestService_UnregisterNode(t *testing.T) {
	registry := node.NewInMemoryRegistry()
	svc := NewService(registry, 30*time.Second)

	// Register a node
	registerReq := &orchionv1.RegisterNodeRequest{
		Address: "localhost:5000",
	}
	registerResp, _ := svc.RegisterNode(context.Background(), registerReq)

	// Unregister the node
	unregisterReq := &orchionv1.UnregisterNodeRequest{
		NodeId: registerResp.NodeId,
	}

	resp, err := svc.UnregisterNode(context.Background(), unregisterReq)
	if err != nil {
		t.Fatalf("Failed to unregister node: %v", err)
	}

	if !resp.Success {
		t.Error("Expected unregister to succeed")
	}
}

func TestService_ListNodes(t *testing.T) {
	registry := node.NewInMemoryRegistry()
	svc := NewService(registry, 30*time.Second)

	// Register multiple nodes
	for i := 0; i < 3; i++ {
		req := &orchionv1.RegisterNodeRequest{
			Address: "localhost:5000",
		}
		svc.RegisterNode(context.Background(), req)
	}

	// List all nodes
	listReq := &orchionv1.ListNodesRequest{
		StatusFilter: orchionv1.NodeStatus_NODE_STATUS_UNSPECIFIED,
	}

	resp, err := svc.ListNodes(context.Background(), listReq)
	if err != nil {
		t.Fatalf("Failed to list nodes: %v", err)
	}

	if len(resp.Nodes) != 3 {
		t.Errorf("Expected 3 nodes, got %d", len(resp.Nodes))
	}
}
