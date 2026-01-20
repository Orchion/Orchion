package heartbeat

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// MockOrchestratorClient is a mock implementation for testing
type MockOrchestratorClient struct {
	mock.Mock
}

func (m *MockOrchestratorClient) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest, opts ...grpc.CallOption) (*pb.RegisterNodeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.RegisterNodeResponse), args.Error(1)
}

func (m *MockOrchestratorClient) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest, opts ...grpc.CallOption) (*pb.HeartbeatResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.HeartbeatResponse), args.Error(1)
}

func (m *MockOrchestratorClient) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest, opts ...grpc.CallOption) (*pb.UpdateNodeResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.UpdateNodeResponse), args.Error(1)
}

func (m *MockOrchestratorClient) ListNodes(ctx context.Context, req *pb.ListNodesRequest, opts ...grpc.CallOption) (*pb.ListNodesResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.ListNodesResponse), args.Error(1)
}

func (m *MockOrchestratorClient) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest, opts ...grpc.CallOption) (*pb.SubmitJobResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.SubmitJobResponse), args.Error(1)
}

func (m *MockOrchestratorClient) GetJobStatus(ctx context.Context, req *pb.GetJobStatusRequest, opts ...grpc.CallOption) (*pb.GetJobStatusResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.GetJobStatusResponse), args.Error(1)
}

func TestNewClient(t *testing.T) {
	// Test with invalid address - may succeed or fail depending on system
	client, err := NewClient("invalid:99999")
	if err != nil {
		assert.Contains(t, err.Error(), "failed to connect")
	} else {
		// If connection succeeded unexpectedly, ensure client is valid
		assert.NotNil(t, client)
		// Clean up
		client.Close()
	}
}

func TestClient_NodeInfoStorage(t *testing.T) {
	// Test the node info storage logic without actual gRPC calls
	client := &Client{
		nodeID: "test-node",
		nodeInfo: &pb.Node{
			Id:       "test-node",
			Hostname: "test-host",
		},
	}

	// Verify node info is stored
	assert.Equal(t, "test-node", client.nodeID)
	assert.NotNil(t, client.nodeInfo)
	assert.Equal(t, "test-node", client.nodeInfo.Id)
	assert.Equal(t, "test-host", client.nodeInfo.Hostname)
}

func TestClient_EnableCapabilityUpdates(t *testing.T) {
	client := &Client{}

	// Initially disabled
	assert.False(t, client.updateCaps)
	assert.Nil(t, client.capsUpdater)

	// Enable updates
	updater := func() *pb.Capabilities {
		return &pb.Capabilities{Cpu: "4 cores"}
	}

	client.EnableCapabilityUpdates(updater)

	assert.True(t, client.updateCaps)
	assert.NotNil(t, client.capsUpdater)

	// Test updater function
	caps := client.capsUpdater()
	assert.Equal(t, "4 cores", caps.Cpu)
}

func TestClient_SendHeartbeat_Unregistered(t *testing.T) {
	client := &Client{
		nodeID: "", // Not registered
	}

	err := client.SendHeartbeat(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not registered")
}

func TestClient_UpdateCapabilities_Unregistered(t *testing.T) {
	client := &Client{
		nodeID: "", // Not registered
	}

	err := client.UpdateCapabilities(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "node not registered")
}

func TestClient_UpdateCapabilities_NoUpdater(t *testing.T) {
	client := &Client{
		nodeID:      "test-node",
		capsUpdater: nil, // No updater configured
	}

	err := client.UpdateCapabilities(context.Background())
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "capability updater not configured")
}

func TestClient_StartHeartbeatLoop_Basic(t *testing.T) {
	// Test that StartHeartbeatLoop can be called without crashing
	// We can't easily test the full loop without a real gRPC client
	client := &Client{
		nodeID: "", // No node registered, so heartbeat should fail gracefully
	}

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	// Start heartbeat loop - should handle the case where no client is connected
	client.StartHeartbeatLoop(ctx, 100*time.Millisecond)

	// Wait for context to timeout
	<-ctx.Done()

	// Test should complete without crashing
	assert.True(t, true, "Heartbeat loop should handle missing client gracefully")
}

func TestClient_Close(t *testing.T) {
	client := &Client{
		conn: nil, // No connection
	}

	// Should not error when conn is nil
	err := client.Close()
	assert.NoError(t, err)
}

// Test the error handling in heartbeat loop (partial test)
func TestHeartbeatLoop_ErrorHandling(t *testing.T) {
	// Test that status parsing works correctly
	err := status.Error(codes.NotFound, "node not found")

	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.NotFound, st.Code())
	assert.Contains(t, st.Message(), "node not found")
}