package llm

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
)


// MockRegistry is a mock implementation of the Registry interface
type MockRegistry struct {
	mock.Mock
}

func (m *MockRegistry) Register(node *pb.Node) error {
	args := m.Called(node)
	return args.Error(0)
}

func (m *MockRegistry) UpdateCapabilities(nodeID string, capabilities *pb.Capabilities) error {
	args := m.Called(nodeID, capabilities)
	return args.Error(0)
}

func (m *MockRegistry) UpdateHeartbeat(nodeID string) error {
	args := m.Called(nodeID)
	return args.Error(0)
}

func (m *MockRegistry) List() []*pb.Node {
	args := m.Called()
	return args.Get(0).([]*pb.Node)
}

func (m *MockRegistry) Get(nodeID string) (*pb.Node, bool) {
	args := m.Called(nodeID)
	if args.Get(0) == nil {
		return nil, args.Bool(1)
	}
	node := args.Get(0).(*pb.Node)
	return node, args.Bool(1)
}

func (m *MockRegistry) Remove(nodeID string) error {
	args := m.Called(nodeID)
	return args.Error(0)
}

func (m *MockRegistry) CheckHeartbeats(timeout time.Duration) []string {
	args := m.Called(timeout)
	return args.Get(0).([]string)
}

// MockScheduler is a mock implementation of the Scheduler interface
type MockScheduler struct {
	mock.Mock
}

func (m *MockScheduler) SelectNode(model string, registry node.Registry) (*pb.Node, error) {
	args := m.Called(model, registry)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pb.Node), args.Error(1)
}


func TestNewService(t *testing.T) {
	mockRegistry := &MockRegistry{}
	mockScheduler := &MockScheduler{}

	service := NewService(mockRegistry, mockScheduler)

	assert.NotNil(t, service)
	assert.Equal(t, mockRegistry, service.registry)
	assert.Equal(t, mockScheduler, service.scheduler)
	assert.NotNil(t, service.nodeClients)
	assert.Len(t, service.nodeClients, 0)
}

func TestService_BasicInitialization(t *testing.T) {
	mockRegistry := &MockRegistry{}
	mockScheduler := &MockScheduler{}
	service := NewService(mockRegistry, mockScheduler)

	// Test that the service can be created properly
	assert.NotNil(t, service)
	assert.NotNil(t, service.registry)
	assert.NotNil(t, service.scheduler)
	assert.NotNil(t, service.nodeClients)
	assert.Len(t, service.nodeClients, 0)
}

func TestService_Embeddings_Validation(t *testing.T) {
	mockRegistry := &MockRegistry{}
	mockScheduler := &MockScheduler{}
	service := NewService(mockRegistry, mockScheduler)

	// Test missing model
	req := &pb.EmbeddingRequest{
		Input: []string{"test"},
	}

	_, err := service.Embeddings(context.Background(), req)
	assert.Error(t, err)
	st, ok := status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "model is required")

	// Test missing input
	req2 := &pb.EmbeddingRequest{
		Model: "text-embedding-ada-002",
	}

	_, err = service.Embeddings(context.Background(), req2)
	assert.Error(t, err)
	st, ok = status.FromError(err)
	require.True(t, ok)
	assert.Equal(t, codes.InvalidArgument, st.Code())
	assert.Contains(t, st.Message(), "input is required")
}

func TestService_getNodeClient_Cache(t *testing.T) {
	mockRegistry := &MockRegistry{}
	mockScheduler := &MockScheduler{}
	service := NewService(mockRegistry, mockScheduler)

	testNode := &pb.Node{
		Id:           "test-node",
		Hostname:     "test-host",
		AgentAddress: "localhost:50051",
	}

	// First call should create new client
	client1, err := service.getNodeClient("test-node", testNode)
	require.NoError(t, err)
	assert.NotNil(t, client1)

	// Second call should return cached client
	client2, err := service.getNodeClient("test-node", testNode)
	require.NoError(t, err)
	assert.Equal(t, client1, client2)

	// Verify only one client was created
	assert.Len(t, service.nodeClients, 1)
}

func TestService_getNodeClient_DefaultAddress(t *testing.T) {
	mockRegistry := &MockRegistry{}
	mockScheduler := &MockScheduler{}
	service := NewService(mockRegistry, mockScheduler)

	testNode := &pb.Node{
		Id:       "test-node",
		Hostname: "test-host",
		// No AgentAddress specified
	}

	// Should use default address format
	client, err := service.getNodeClient("test-node", testNode)
	require.NoError(t, err)
	assert.NotNil(t, client)

	// Verify client was cached
	assert.Len(t, service.nodeClients, 1)
	assert.Contains(t, service.nodeClients, "test-node")
}

func TestService_getNodeClient_ErrorHandling(t *testing.T) {
	mockRegistry := &MockRegistry{}
	mockScheduler := &MockScheduler{}
	service := NewService(mockRegistry, mockScheduler)

	// Test with invalid address (may or may not fail depending on system)
	testNode := &pb.Node{
		Id:           "bad-node",
		Hostname:     "invalid-host",
		AgentAddress: "invalid:99999",
	}

	client, err := service.getNodeClient("bad-node", testNode)
	// Connection may succeed or fail depending on system configuration
	// The important thing is that it doesn't panic
	if err != nil {
		assert.Contains(t, err.Error(), "failed to connect")
	} else {
		// If connection succeeded, client should not be nil
		assert.NotNil(t, client)
	}
}

