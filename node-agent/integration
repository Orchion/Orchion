package integration

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/node-agent/internal/capabilities"
	"github.com/Orchion/Orchion/node-agent/internal/heartbeat"
)

// testNodeAgent holds the test node agent components
type testNodeAgent struct {
	capabilities *capabilities.Service
	heartbeater  *heartbeat.Service
	cleanup      func()
}

// setupTestNodeAgent creates a test node agent with all components
func setupTestNodeAgent(t *testing.T, orchestratorAddr string) *testNodeAgent {
	// Create capabilities service
	caps := capabilities.NewService()

	// Create heartbeat service
	hb := heartbeat.NewService(orchestratorAddr, "test-node-agent", 30*time.Second)

	cleanup := func() {
		hb.Stop()
	}

	return &testNodeAgent{
		capabilities: caps,
		heartbeater:  hb,
		cleanup:      cleanup,
	}
}

func TestNodeAgent_CapabilityDetection(t *testing.T) {
	// Test capability detection in isolation (no orchestrator needed)
	caps := capabilities.NewService()

	capabilities, err := caps.Detect()
	require.NoError(t, err)
	assert.NotNil(t, capabilities)

	// Verify basic capability fields are populated
	assert.NotEmpty(t, capabilities.OS)
	assert.True(t, capabilities.CPUCores > 0)
	assert.True(t, capabilities.MemoryBytes > 0)

	// CPU and memory should be reasonable values
	assert.Contains(t, []string{"linux", "darwin", "windows"}, capabilities.OS)
	assert.True(t, capabilities.CPUCores >= 1 && capabilities.CPUCores <= 128)
	assert.True(t, capabilities.MemoryBytes >= 1024*1024*1024) // At least 1GB
}

func TestNodeAgent_RegistrationWorkflow(t *testing.T) {
	// Setup test orchestrator server
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Create test node agent (simulating what the real node agent would do)
	caps := capabilities.NewService()
	nodeCapabilities, err := caps.Detect()
	require.NoError(t, err)

	// Create node registration message
	node := &pb.Node{
		Id:       "integration-test-agent",
		Hostname: "test-agent-host",
		Capabilities: &pb.Capabilities{
			Cpu:       fmt.Sprintf("%d cores", nodeCapabilities.CPUCores),
			Memory:    fmt.Sprintf("%d GB", nodeCapabilities.MemoryBytes/(1024*1024*1024)),
			Os:        nodeCapabilities.OS,
			GpuType:   nodeCapabilities.GPUType,
			GpuMemory: fmt.Sprintf("%d GB", nodeCapabilities.GPUMemoryBytes/(1024*1024*1024)),
		},
	}

	// Register with orchestrator
	regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
	require.NoError(t, err)
	assert.True(t, regResp.Success)

	// Verify registration
	listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, 1)

	registeredNode := listResp.Nodes[0]
	assert.Equal(t, "integration-test-agent", registeredNode.Id)
	assert.Equal(t, "test-agent-host", registeredNode.Hostname)
	assert.Equal(t, nodeCapabilities.OS, registeredNode.Capabilities.Os)
	assert.Equal(t, fmt.Sprintf("%d cores", nodeCapabilities.CPUCores), registeredNode.Capabilities.Cpu)
}

func TestNodeAgent_HeartbeatLoop(t *testing.T) {
	// Setup test orchestrator server
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// First register a node
	node := &pb.Node{
		Id:       "heartbeat-test-agent",
		Hostname: "heartbeat-host",
		Capabilities: &pb.Capabilities{
			Cpu:    "4 cores",
			Memory: "8GB",
			Os:     "linux",
		},
	}

	regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
	require.NoError(t, err)
	assert.True(t, regResp.Success)

	// Simulate heartbeat loop (what the real heartbeat service would do)
	initialTime := time.Now()

	// Send multiple heartbeats
	for i := 0; i < 3; i++ {
		hbResp, err := ts.service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "heartbeat-test-agent"})
		require.NoError(t, err)
		assert.True(t, hbResp.Success)

		time.Sleep(10 * time.Millisecond) // Small delay to ensure timestamp difference
	}

	// Verify heartbeat updated the timestamp
	listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, 1)

	updatedNode := listResp.Nodes[0]
	assert.Equal(t, "heartbeat-test-agent", updatedNode.Id)
	assert.True(t, updatedNode.LastSeenUnix >= initialTime.Unix())
}

func TestNodeAgent_CapabilityUpdates(t *testing.T) {
	// Setup test orchestrator server
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Register node with initial capabilities
	node := &pb.Node{
		Id:       "capability-update-test",
		Hostname: "cap-update-host",
		Capabilities: &pb.Capabilities{
			Cpu:    "2 cores",
			Memory: "4GB",
			Os:     "linux",
		},
	}

	regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
	require.NoError(t, err)
	assert.True(t, regResp.Success)

	// Update capabilities (simulating what happens when hardware changes or is re-detected)
	updatedCaps := &pb.Capabilities{
		Cpu:       "8 cores",
		Memory:    "16GB",
		Os:        "linux",
		GpuType:   "NVIDIA RTX 3080",
		GpuMemory: "10GB",
	}

	updateResp, err := ts.service.UpdateNode(ctx, &pb.UpdateNodeRequest{
		NodeId: "capability-update-test",
		Node: &pb.Node{
			Id:          "capability-update-test",
			Hostname:    "cap-update-host",
			Capabilities: updatedCaps,
		},
	})
	require.NoError(t, err)
	assert.True(t, updateResp.Success)

	// Verify capability update
	listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	assert.Len(t, listResp.Nodes, 1)

	updatedNode := listResp.Nodes[0]
	assert.Equal(t, "8 cores", updatedNode.Capabilities.Cpu)
	assert.Equal(t, "16GB", updatedNode.Capabilities.Memory)
	assert.Equal(t, "NVIDIA RTX 3080", updatedNode.Capabilities.GpuType)
	assert.Equal(t, "10GB", updatedNode.Capabilities.GpuMemory)
}

func TestNodeAgent_ConcurrentCapabilityDetection(t *testing.T) {
	const numGoroutines = 10

	var wg sync.WaitGroup
	results := make(chan *capabilities.Capabilities, numGoroutines)
	errors := make(chan error, numGoroutines)

	// Run multiple capability detections concurrently
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()

			caps := capabilities.NewService()
			capabilities, err := caps.Detect()
			if err != nil {
				errors <- err
				return
			}
			results <- capabilities
		}()
	}

	wg.Wait()
	close(results)
	close(errors)

	// Check for errors
	select {
	case err := <-errors:
		t.Fatalf("Capability detection failed: %v", err)
	default:
		// No errors
	}

	// Verify all results are consistent (should be the same system capabilities)
	var allResults []*capabilities.Capabilities
	for result := range results {
		allResults = append(allResults, result)
	}

	assert.Len(t, allResults, numGoroutines)

	// All results should be the same (detecting the same system)
	first := allResults[0]
	for i := 1; i < len(allResults); i++ {
		assert.Equal(t, first.OS, allResults[i].OS)
		assert.Equal(t, first.CPUCores, allResults[i].CPUCores)
		assert.Equal(t, first.MemoryBytes, allResults[i].MemoryBytes)
	}
}

func TestNodeAgent_ErrorHandling(t *testing.T) {
	// Test capability detection error handling
	caps := capabilities.NewService()

	// This should work on any reasonable system
	capabilities, err := caps.Detect()
	require.NoError(t, err)
	assert.NotNil(t, capabilities)

	// Test with orchestrator that's not running
	// (We can't easily test heartbeat errors without mocking the connection)
	hb := heartbeat.NewService("localhost:99999", "error-test-node", 30*time.Second)
	defer hb.Stop()

	// The heartbeat service should handle connection errors gracefully
	// In a real scenario, it would retry connections
	assert.NotNil(t, hb)
}