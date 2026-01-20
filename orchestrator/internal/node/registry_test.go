package node

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
)

func TestNewInMemoryRegistry(t *testing.T) {
	registry := NewInMemoryRegistry()
	assert.NotNil(t, registry)
	assert.NotNil(t, registry.nodes)
	assert.Equal(t, 0, len(registry.nodes))
}

func TestInMemoryRegistry_Register(t *testing.T) {
	registry := NewInMemoryRegistry()

	t.Run("successful registration", func(t *testing.T) {
		node := &pb.Node{
			Id:       "test-node-1",
			Hostname: "test-host",
			Capabilities: &pb.Capabilities{
				Cpu:    "4 cores",
				Memory: "8GB",
				Os:     "linux",
			},
		}

		err := registry.Register(node)
		require.NoError(t, err)

		// Verify node was registered
		retrieved, exists := registry.Get("test-node-1")
		assert.True(t, exists)
		assert.Equal(t, "test-node-1", retrieved.Id)
		assert.Equal(t, "test-host", retrieved.Hostname)
		assert.NotZero(t, retrieved.LastSeenUnix) // Should be set automatically
	})

	t.Run("registration with existing LastSeenUnix", func(t *testing.T) {
		customTime := int64(1234567890)
		node := &pb.Node{
			Id:          "test-node-2",
			Hostname:    "test-host-2",
			LastSeenUnix: customTime,
		}

		err := registry.Register(node)
		require.NoError(t, err)

		retrieved, exists := registry.Get("test-node-2")
		assert.True(t, exists)
		assert.Equal(t, customTime, retrieved.LastSeenUnix) // Should preserve existing time
	})

	t.Run("update existing node", func(t *testing.T) {
		// Register initial node
		node := &pb.Node{
			Id:       "update-test",
			Hostname: "old-host",
			Capabilities: &pb.Capabilities{
				Cpu: "2 cores",
			},
		}
		err := registry.Register(node)
		require.NoError(t, err)

		// Update with new data
		updatedNode := &pb.Node{
			Id:       "update-test",
			Hostname: "new-host",
			Capabilities: &pb.Capabilities{
				Cpu:    "4 cores",
				Memory: "16GB",
			},
		}
		err = registry.Register(updatedNode)
		require.NoError(t, err)

		// Verify update
		retrieved, exists := registry.Get("update-test")
		assert.True(t, exists)
		assert.Equal(t, "new-host", retrieved.Hostname)
		assert.Equal(t, "4 cores", retrieved.Capabilities.Cpu)
		assert.Equal(t, "16GB", retrieved.Capabilities.Memory)
	})
}

func TestInMemoryRegistry_UpdateCapabilities(t *testing.T) {
	registry := NewInMemoryRegistry()

	t.Run("successful update", func(t *testing.T) {
		// Register node first
		node := &pb.Node{
			Id:       "cap-test",
			Hostname: "test-host",
			Capabilities: &pb.Capabilities{
				Cpu: "2 cores",
			},
		}
		err := registry.Register(node)
		require.NoError(t, err)

		// Get the original timestamp after registration
		original, exists := registry.Get("cap-test")
		require.True(t, exists)
		originalTime := original.LastSeenUnix

		// Wait a tiny bit to ensure timestamp difference
		time.Sleep(1 * time.Millisecond)

		// Update capabilities
		newCaps := &pb.Capabilities{
			Cpu:    "8 cores",
			Memory: "32GB",
			Os:     "linux",
			GpuType: "NVIDIA RTX 3080",
		}
		err = registry.UpdateCapabilities("cap-test", newCaps)
		require.NoError(t, err)

		// Verify update
		retrieved, exists := registry.Get("cap-test")
		assert.True(t, exists)
		assert.Equal(t, "8 cores", retrieved.Capabilities.Cpu)
		assert.Equal(t, "32GB", retrieved.Capabilities.Memory)
		assert.Equal(t, "NVIDIA RTX 3080", retrieved.Capabilities.GpuType)
		assert.True(t, retrieved.LastSeenUnix >= originalTime) // Timestamp should be updated (or at least not older)
	})

	t.Run("update non-existent node", func(t *testing.T) {
		newCaps := &pb.Capabilities{Cpu: "4 cores"}
		err := registry.UpdateCapabilities("non-existent", newCaps)
		assert.Error(t, err)
		assert.Equal(t, ErrNodeNotFound, err)
	})
}

func TestInMemoryRegistry_UpdateHeartbeat(t *testing.T) {
	registry := NewInMemoryRegistry()

	t.Run("successful heartbeat update", func(t *testing.T) {
		// Register node
		node := &pb.Node{
			Id:          "heartbeat-test",
			LastSeenUnix: 1000000000, // Old timestamp
		}
		err := registry.Register(node)
		require.NoError(t, err)

		originalTime := node.LastSeenUnix

		// Update heartbeat
		err = registry.UpdateHeartbeat("heartbeat-test")
		require.NoError(t, err)

		// Verify timestamp was updated
		retrieved, exists := registry.Get("heartbeat-test")
		assert.True(t, exists)
		assert.True(t, retrieved.LastSeenUnix > originalTime)
	})

	t.Run("heartbeat for non-existent node", func(t *testing.T) {
		err := registry.UpdateHeartbeat("non-existent")
		assert.Error(t, err)
		assert.Equal(t, ErrNodeNotFound, err)
	})
}

func TestInMemoryRegistry_List(t *testing.T) {
	registry := NewInMemoryRegistry()

	t.Run("empty registry", func(t *testing.T) {
		nodes := registry.List()
		assert.Empty(t, nodes)
	})

	t.Run("multiple nodes", func(t *testing.T) {
		// Register multiple nodes
		node1 := &pb.Node{Id: "node-1", Hostname: "host-1"}
		node2 := &pb.Node{Id: "node-2", Hostname: "host-2"}
		node3 := &pb.Node{Id: "node-3", Hostname: "host-3"}

		registry.Register(node1)
		registry.Register(node2)
		registry.Register(node3)

		// List nodes
		nodes := registry.List()
		assert.Len(t, nodes, 3)

		// Extract IDs for verification
		ids := make([]string, len(nodes))
		for i, node := range nodes {
			ids[i] = node.Id
		}

		assert.Contains(t, ids, "node-1")
		assert.Contains(t, ids, "node-2")
		assert.Contains(t, ids, "node-3")

		// Verify copies are returned (not references)
		nodes[0].Hostname = "modified"
		retrieved, exists := registry.Get("node-1")
		assert.True(t, exists)
		assert.Equal(t, "host-1", retrieved.Hostname) // Original should be unchanged
	})
}

func TestInMemoryRegistry_Get(t *testing.T) {
	registry := NewInMemoryRegistry()

	t.Run("get existing node", func(t *testing.T) {
		node := &pb.Node{
			Id:       "get-test",
			Hostname: "test-host",
			Capabilities: &pb.Capabilities{
				Cpu: "4 cores",
			},
		}
		registry.Register(node)

		retrieved, exists := registry.Get("get-test")
		assert.True(t, exists)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "get-test", retrieved.Id)
		assert.Equal(t, "test-host", retrieved.Hostname)

		// Verify copy is returned
		retrieved.Hostname = "modified"
		original, exists := registry.Get("get-test")
		assert.True(t, exists)
		assert.Equal(t, "test-host", original.Hostname)
	})

	t.Run("get non-existent node", func(t *testing.T) {
		retrieved, exists := registry.Get("non-existent")
		assert.False(t, exists)
		assert.Nil(t, retrieved)
	})
}

func TestInMemoryRegistry_Remove(t *testing.T) {
	registry := NewInMemoryRegistry()

	t.Run("successful removal", func(t *testing.T) {
		// Register node
		node := &pb.Node{Id: "remove-test", Hostname: "test-host"}
		registry.Register(node)

		// Verify it exists
		_, exists := registry.Get("remove-test")
		assert.True(t, exists)

		// Remove it
		err := registry.Remove("remove-test")
		require.NoError(t, err)

		// Verify it's gone
		_, exists = registry.Get("remove-test")
		assert.False(t, exists)
	})

	t.Run("remove non-existent node", func(t *testing.T) {
		err := registry.Remove("non-existent")
		assert.Error(t, err)
		assert.Equal(t, ErrNodeNotFound, err)
	})
}

func TestInMemoryRegistry_CheckHeartbeats(t *testing.T) {
	registry := NewInMemoryRegistry()

	t.Run("no stale nodes", func(t *testing.T) {
		// Register recent nodes
		node1 := &pb.Node{Id: "fresh-1", LastSeenUnix: time.Now().Unix()}
		node2 := &pb.Node{Id: "fresh-2", LastSeenUnix: time.Now().Unix()}
		registry.Register(node1)
		registry.Register(node2)

		stale := registry.CheckHeartbeats(5 * time.Minute)
		assert.Empty(t, stale)
	})

	t.Run("some stale nodes", func(t *testing.T) {
		// Register mix of fresh and stale nodes
		freshNode := &pb.Node{Id: "fresh", LastSeenUnix: time.Now().Unix()}
		staleNode1 := &pb.Node{Id: "stale-1", LastSeenUnix: time.Now().Add(-10 * time.Minute).Unix()}
		staleNode2 := &pb.Node{Id: "stale-2", LastSeenUnix: time.Now().Add(-15 * time.Minute).Unix()}

		registry.Register(freshNode)
		registry.Register(staleNode1)
		registry.Register(staleNode2)

		stale := registry.CheckHeartbeats(5 * time.Minute)
		assert.Len(t, stale, 2)
		assert.Contains(t, stale, "stale-1")
		assert.Contains(t, stale, "stale-2")
		assert.NotContains(t, stale, "fresh")
	})

	t.Run("all nodes stale", func(t *testing.T) {
		// Clear registry and add only stale nodes
		registry = NewInMemoryRegistry()

		staleNode1 := &pb.Node{Id: "stale-1", LastSeenUnix: time.Now().Add(-10 * time.Minute).Unix()}
		staleNode2 := &pb.Node{Id: "stale-2", LastSeenUnix: time.Now().Add(-20 * time.Minute).Unix()}

		registry.Register(staleNode1)
		registry.Register(staleNode2)

		stale := registry.CheckHeartbeats(5 * time.Minute)
		assert.Len(t, stale, 2)
		assert.Contains(t, stale, "stale-1")
		assert.Contains(t, stale, "stale-2")
	})

	t.Run("edge case - slightly before timeout", func(t *testing.T) {
		// Use a fresh registry for this test
		freshRegistry := NewInMemoryRegistry()

		timeout := 5 * time.Minute
		// Set node to be just within the timeout (1 second before expiry)
		slightlyBeforeTimeout := time.Now().Add(-(timeout - time.Second)).Unix()

		node := &pb.Node{Id: "edge-case", LastSeenUnix: slightlyBeforeTimeout}
		freshRegistry.Register(node)

		stale := freshRegistry.CheckHeartbeats(timeout)
		// Should not be considered stale since it's within timeout
		assert.Empty(t, stale)
	})
}

func TestInMemoryRegistry_Concurrency(t *testing.T) {
	registry := NewInMemoryRegistry()
	const numGoroutines = 10
	const operationsPerGoroutine = 100

	var wg sync.WaitGroup

	// Start multiple goroutines performing operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				nodeID := fmt.Sprintf("node-%d-%d", id, j)

				// Register
				node := &pb.Node{Id: nodeID, Hostname: fmt.Sprintf("host-%d-%d", id, j)}
				registry.Register(node)

				// Update heartbeat
				registry.UpdateHeartbeat(nodeID)

				// Get
				_, exists := registry.Get(nodeID)
				assert.True(t, exists)

				// List (just call it)
				registry.List()

				// Remove
				registry.Remove(nodeID)
			}
		}(i)
	}

	wg.Wait()

	// Final check - registry should be empty
	nodes := registry.List()
	assert.Empty(t, nodes)
}

func TestRegistryError_Error(t *testing.T) {
	err := &RegistryError{Message: "test error"}
	assert.Equal(t, "test error", err.Error())
}

func TestErrNodeNotFound(t *testing.T) {
	assert.NotNil(t, ErrNodeNotFound)
	assert.Equal(t, "node not found", ErrNodeNotFound.Error())
}