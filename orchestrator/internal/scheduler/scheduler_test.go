package scheduler

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
)

// MockRegistry is a mock implementation of node.Registry for testing
type MockRegistry struct {
	nodes []*pb.Node
}

func (m *MockRegistry) Register(node *pb.Node) error {
	m.nodes = append(m.nodes, node)
	return nil
}

func (m *MockRegistry) UpdateCapabilities(nodeID string, capabilities *pb.Capabilities) error {
	return nil
}

func (m *MockRegistry) UpdateHeartbeat(nodeID string) error {
	return nil
}

func (m *MockRegistry) List() []*pb.Node {
	return m.nodes
}

func (m *MockRegistry) Get(nodeID string) (*pb.Node, bool) {
	for _, node := range m.nodes {
		if node.Id == nodeID {
			return node, true
		}
	}
	return nil, false
}

func (m *MockRegistry) Remove(nodeID string) error {
	for i, node := range m.nodes {
		if node.Id == nodeID {
			m.nodes = append(m.nodes[:i], m.nodes[i+1:]...)
			break
		}
	}
	return nil
}

func (m *MockRegistry) CheckHeartbeats(timeout time.Duration) []string {
	return []string{}
}

func TestNewSimpleScheduler(t *testing.T) {
	scheduler := NewSimpleScheduler()
	assert.NotNil(t, scheduler)
}

func TestSimpleScheduler_SelectNode(t *testing.T) {
	scheduler := NewSimpleScheduler()

	t.Run("successful node selection", func(t *testing.T) {
		mockRegistry := &MockRegistry{
			nodes: []*pb.Node{
				{Id: "node-1", Hostname: "host-1"},
				{Id: "node-2", Hostname: "host-2"},
				{Id: "node-3", Hostname: "host-3"},
			},
		}

		selectedNode, err := scheduler.SelectNode("llama2", mockRegistry)

		require.NoError(t, err)
		assert.NotNil(t, selectedNode)
		// Should select the first node
		assert.Equal(t, "node-1", selectedNode.Id)
		assert.Equal(t, "host-1", selectedNode.Hostname)
	})

	t.Run("single node available", func(t *testing.T) {
		mockRegistry := &MockRegistry{
			nodes: []*pb.Node{
				{Id: "single-node", Hostname: "single-host"},
			},
		}

		selectedNode, err := scheduler.SelectNode("gpt-3", mockRegistry)

		require.NoError(t, err)
		assert.NotNil(t, selectedNode)
		assert.Equal(t, "single-node", selectedNode.Id)
		assert.Equal(t, "single-host", selectedNode.Hostname)
	})

	t.Run("no nodes available", func(t *testing.T) {
		mockRegistry := &MockRegistry{
			nodes: []*pb.Node{}, // Empty registry
		}

		selectedNode, err := scheduler.SelectNode("any-model", mockRegistry)

		assert.Error(t, err)
		assert.Nil(t, selectedNode)
		assert.Equal(t, ErrNoNodesAvailable, err)
	})

	t.Run("empty registry", func(t *testing.T) {
		mockRegistry := &MockRegistry{
			nodes: nil, // Nil slice
		}

		selectedNode, err := scheduler.SelectNode("model", mockRegistry)

		assert.Error(t, err)
		assert.Nil(t, selectedNode)
		assert.Equal(t, ErrNoNodesAvailable, err)
	})
}

func TestSchedulerError_Error(t *testing.T) {
	err := &SchedulerError{Message: "test scheduler error"}
	assert.Equal(t, "test scheduler error", err.Error())
}

func TestErrNoNodesAvailable(t *testing.T) {
	assert.NotNil(t, ErrNoNodesAvailable)
	assert.Equal(t, "no nodes available", ErrNoNodesAvailable.Error())
}

// Benchmark tests for performance
func BenchmarkSimpleScheduler_SelectNode(b *testing.B) {
	scheduler := NewSimpleScheduler()
	mockRegistry := &MockRegistry{
		nodes: []*pb.Node{
			{Id: "node-1", Hostname: "host-1"},
			{Id: "node-2", Hostname: "host-2"},
			{Id: "node-3", Hostname: "host-3"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scheduler.SelectNode("benchmark-model", mockRegistry)
	}
}

func BenchmarkSimpleScheduler_SelectNode_EmptyRegistry(b *testing.B) {
	scheduler := NewSimpleScheduler()
	mockRegistry := &MockRegistry{
		nodes: []*pb.Node{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = scheduler.SelectNode("benchmark-model", mockRegistry)
	}
}