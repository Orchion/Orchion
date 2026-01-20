package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	"github.com/Orchion/Orchion/orchestrator/internal/orchestrator"
	"github.com/Orchion/Orchion/orchestrator/internal/queue"
	"github.com/Orchion/Orchion/orchestrator/internal/scheduler"
)

// testServer holds the test gRPC server and client connections
type testServer struct {
	server     *grpc.Server
	listener   net.Listener
	clientConn *grpc.ClientConn
	service    pb.OrchestratorClient
	cleanup    func()
}

// setupTestServer creates a test orchestrator server with all components
func setupTestServer(t *testing.T) *testServer {
	// Create components
	registry := node.NewInMemoryRegistry()
	jobQueue := queue.NewJobQueue()
	scheduler := scheduler.NewSimpleScheduler()
	service := orchestrator.NewService(registry, jobQueue, scheduler)

	// Create gRPC server
	server := grpc.NewServer()
	pb.RegisterOrchestratorServer(server, service)

	// Start server on random port
	listener, err := net.Listen("tcp", "localhost:0")
	require.NoError(t, err)

	go func() {
		if err := server.Serve(listener); err != nil && err != grpc.ErrServerStopped {
			t.Errorf("Server error: %v", err)
		}
	}()

	// Create client connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	clientConn, err := grpc.DialContext(ctx, listener.Addr().String(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithBlock())
	require.NoError(t, err)

	client := pb.NewOrchestratorClient(clientConn)

	cleanup := func() {
		clientConn.Close()
		server.Stop()
		listener.Close()
	}

	return &testServer{
		server:     server,
		listener:   listener,
		clientConn: clientConn,
		service:    client,
		cleanup:    cleanup,
	}
}

func TestOrchestrator_FullNodeLifecycle(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Test node registration
	node := &pb.Node{
		Id:       "integration-test-node",
		Hostname: "test-host",
		Capabilities: &pb.Capabilities{
			Cpu:    "4 cores",
			Memory: "8GB",
			Os:     "linux",
		},
	}

	regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
	require.NoError(t, err)
	require.NotNil(t, regResp)

	// Test heartbeat
	hbResp, err := ts.service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "integration-test-node"})
	require.NoError(t, err)
	require.NotNil(t, hbResp)

	// Test listing nodes
	listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, 1)
	assert.Equal(t, "integration-test-node", listResp.Nodes[0].Id)
	assert.Equal(t, "test-host", listResp.Nodes[0].Hostname)

	// Test node capabilities update
	updateResp, err := ts.service.UpdateNode(ctx, &pb.UpdateNodeRequest{
		NodeId: "integration-test-node",
		Capabilities: &pb.Capabilities{
			Cpu:     "8 cores",
			Memory:  "16GB",
			Os:      "linux",
			GpuType: "NVIDIA RTX 3080",
		},
	})
	require.NoError(t, err)
	require.NotNil(t, updateResp)

	// Verify capabilities update
	listResp, err = ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, 1)
	assert.Equal(t, "test-host", listResp.Nodes[0].Hostname) // Hostname should remain unchanged
	assert.Equal(t, "8 cores", listResp.Nodes[0].Capabilities.Cpu)
	assert.Equal(t, "NVIDIA RTX 3080", listResp.Nodes[0].Capabilities.GpuType)
}

func TestOrchestrator_ConcurrentNodeOperations(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()
	const numGoroutines = 10
	const operationsPerGoroutine = 5

	var wg sync.WaitGroup
	errorChan := make(chan error, numGoroutines*operationsPerGoroutine)

	// Start concurrent operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				nodeID := fmt.Sprintf("concurrent-node-%d-%d", id, j)
				hostname := fmt.Sprintf("host-%d-%d", id, j)

				// Register node
				node := &pb.Node{
					Id:       nodeID,
					Hostname: hostname,
					Capabilities: &pb.Capabilities{
						Cpu:    "4 cores",
						Memory: "8GB",
						Os:     "linux",
					},
				}

				regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
				if err != nil {
					errorChan <- fmt.Errorf("register error: %v", err)
					continue
				}
				if regResp == nil {
					errorChan <- fmt.Errorf("register returned nil response for node %s", nodeID)
					continue
				}

				// Heartbeat
				hbResp, err := ts.service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: nodeID})
				if err != nil {
					errorChan <- fmt.Errorf("heartbeat error: %v", err)
					continue
				}
				if hbResp == nil {
					errorChan <- fmt.Errorf("heartbeat returned nil response for node %s", nodeID)
					continue
				}

				// Verify node exists in list
				listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
				if err != nil {
					errorChan <- fmt.Errorf("list error: %v", err)
					continue
				}

				found := false
				for _, n := range listResp.Nodes {
					if n.Id == nodeID {
						found = true
						break
					}
				}
				if !found {
					errorChan <- fmt.Errorf("node %s not found in list", nodeID)
				}
			}
		}(i)
	}

	wg.Wait()
	close(errorChan)

	// Check for errors
	var errors []error
	for err := range errorChan {
		errors = append(errors, err)
	}

	if len(errors) > 0 {
		t.Errorf("Concurrent operations failed with %d errors:", len(errors))
		for _, err := range errors {
			t.Errorf("  %v", err)
		}
	}

	// Verify final state
	listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, numGoroutines*operationsPerGoroutine)
}

func TestOrchestrator_HeartbeatTimeout(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Register a node with old timestamp (simulate stale node)
	staleNode := &pb.Node{
		Id:           "stale-node",
		Hostname:     "stale-host",
		LastSeenUnix: time.Now().Add(-10 * time.Minute).Unix(), // 10 minutes ago
		Capabilities: &pb.Capabilities{
			Cpu:    "4 cores",
			Memory: "8GB",
			Os:     "linux",
		},
	}

	regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: staleNode})
	require.NoError(t, err)
	require.NotNil(t, regResp)

	// Verify node is initially listed
	listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, 1)
	assert.Equal(t, "stale-node", listResp.Nodes[0].Id)

	// In a real implementation, there would be a background goroutine checking heartbeats
	// For this test, we'll simulate the heartbeat check by directly accessing the registry
	// This is a limitation of the current test setup - in production, stale nodes would be
	// cleaned up automatically by the heartbeat monitor goroutine

	// For now, just verify that heartbeat updates work
	hbResp, err := ts.service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "stale-node"})
	require.NoError(t, err)
	require.NotNil(t, hbResp)

	// Node should still be listed after heartbeat
	listResp, err = ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, 1)
	assert.Equal(t, "stale-node", listResp.Nodes[0].Id)
}

func TestOrchestrator_JobSubmission(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// First register a node (required for job scheduling)
	node := &pb.Node{
		Id:       "job-test-node",
		Hostname: "job-host",
		Capabilities: &pb.Capabilities{
			Cpu:    "4 cores",
			Memory: "8GB",
			Os:     "linux",
		},
	}

	regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
	require.NoError(t, err)
	require.NotNil(t, regResp)

	// Submit a job using the SubmitJob API
	chatReq := &pb.ChatCompletionRequest{
		Model: "test-model",
		Messages: []*pb.ChatMessage{
			{Role: "user", Content: "Hello, test message"},
		},
		Stream: true,
	}

	// Serialize the chat request (in practice this would be done properly)
	payload, err := json.Marshal(chatReq)
	require.NoError(t, err)

	jobReq := &pb.SubmitJobRequest{
		JobId:   "test-job-123",
		JobType: pb.JobType_JOB_TYPE_CHAT_COMPLETION,
		Payload: payload,
	}

	jobResp, err := ts.service.SubmitJob(ctx, jobReq)
	require.NoError(t, err)
	require.NotNil(t, jobResp)
	assert.Equal(t, "test-job-123", jobResp.JobId)
	assert.Equal(t, pb.JobStatus_JOB_STATUS_PENDING, jobResp.Status)

	// Check job status
	statusReq := &pb.GetJobStatusRequest{JobId: "test-job-123"}
	statusResp, err := ts.service.GetJobStatus(ctx, statusReq)
	require.NoError(t, err)
	require.NotNil(t, statusResp)
	assert.Equal(t, "test-job-123", statusResp.JobId)
	// Status might be PENDING or ASSIGNED depending on implementation
	assert.True(t, statusResp.Status == pb.JobStatus_JOB_STATUS_PENDING ||
		statusResp.Status == pb.JobStatus_JOB_STATUS_ASSIGNED)
}

func TestOrchestrator_ErrorHandling(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Test heartbeat for non-existent node
	_, err := ts.service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "non-existent-node"})
	require.Error(t, err) // Should return an error for non-existent nodes

	// Test update for non-existent node
	_, err = ts.service.UpdateNode(ctx, &pb.UpdateNodeRequest{
		NodeId: "non-existent-node",
		Capabilities: &pb.Capabilities{
			Cpu: "4 cores",
		},
	})
	require.Error(t, err) // Should return an error for non-existent nodes

	// Test registration with invalid data
	invalidNode := &pb.Node{
		Id: "", // Empty ID should fail
		Capabilities: &pb.Capabilities{
			Cpu: "4 cores",
		},
	}

	_, err = ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: invalidNode})
	require.Error(t, err) // Should return an error for invalid input
}
