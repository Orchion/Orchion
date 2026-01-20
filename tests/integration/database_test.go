package integration

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/Orchion/Orchion/shared/proto/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	queue "github.com/Orchion/Orchion/orchestrator/internal/queue"
)

func TestDatabase_Node_Registry_Persistence(t *testing.T) {
	// Test that node registry operations persist correctly
	registry := node.NewInMemoryRegistry()

	// Register multiple nodes
	node1 := &pb.Node{
		Id:           "db-test-node-1",
		Hostname:     "db-host-1",
		LastSeenUnix: time.Now().Unix(),
		Capabilities: &pb.Capabilities{
			Cpu:    "4 cores",
			Memory: "8GB",
			Os:     "linux",
		},
	}

	node2 := &pb.Node{
		Id:           "db-test-node-2",
		Hostname:     "db-host-2",
		LastSeenUnix: time.Now().Unix(),
		Capabilities: &pb.Capabilities{
			Cpu:    "8 cores",
			Memory: "16GB",
			Os:     "linux",
		},
	}

	// Register nodes
	err := registry.Register(node1)
	require.NoError(t, err)

	err = registry.Register(node2)
	require.NoError(t, err)

	// Verify nodes are stored
	retrieved1, exists1 := registry.Get("db-test-node-1")
	assert.True(t, exists1)
	assert.Equal(t, "db-host-1", retrieved1.Hostname)

	retrieved2, exists2 := registry.Get("db-test-node-2")
	assert.True(t, exists2)
	assert.Equal(t, "db-host-2", retrieved2.Hostname)

	// Test listing
	allNodes := registry.List()
	assert.Len(t, allNodes, 2)

	nodeIDs := make(map[string]bool)
	for _, n := range allNodes {
		nodeIDs[n.Id] = true
	}
	assert.True(t, nodeIDs["db-test-node-1"])
	assert.True(t, nodeIDs["db-test-node-2"])

	// Test capability updates
	updatedCaps := &pb.Capabilities{
		Cpu:     "16 cores",
		Memory:  "32GB",
		Os:      "linux",
		GpuType: "NVIDIA A100",
	}

	err = registry.UpdateCapabilities("db-test-node-1", updatedCaps)
	require.NoError(t, err)

	// Verify update
	updatedNode, exists := registry.Get("db-test-node-1")
	require.True(t, exists)
	assert.Equal(t, "16 cores", updatedNode.Capabilities.Cpu)
	assert.Equal(t, "32GB", updatedNode.Capabilities.Memory)
	assert.Equal(t, "NVIDIA A100", updatedNode.Capabilities.GpuType)

	// Test heartbeat updates
	oldTime := updatedNode.LastSeenUnix
	time.Sleep(1 * time.Millisecond) // Ensure time difference

	err = registry.UpdateHeartbeat("db-test-node-1")
	require.NoError(t, err)

	updatedNode2, exists := registry.Get("db-test-node-1")
	require.True(t, exists)
	assert.True(t, updatedNode2.LastSeenUnix > oldTime)

	// Test removal
	err = registry.Remove("db-test-node-2")
	require.NoError(t, err)

	_, exists = registry.Get("db-test-node-2")
	assert.False(t, exists)

	// Verify only one node remains
	finalList := registry.List()
	assert.Len(t, finalList, 1)
	assert.Equal(t, "db-test-node-1", finalList[0].Id)
}

func TestDatabase_Job_Queue_Persistence(t *testing.T) {
	// Test that job queue operations persist correctly
	jobQueue := queue.NewJobQueue()

	// Create test jobs
	job1 := &queue.Job{
		ID:     "db-test-job-1",
		Type:   queue.JobTypeChatCompletion,
		Status: queue.JobPending,
		Payload: []byte(`{
			"model": "gpt-3.5-turbo",
			"messages": [{"role": "user", "content": "Hello"}]
		}`),
		CreatedAt: time.Now(),
	}

	job2 := &queue.Job{
		ID:     "db-test-job-2",
		Type:   queue.JobTypeEmbeddings,
		Status: queue.JobPending,
		Payload: []byte(`{
			"model": "text-embedding-ada-002",
			"input": ["test input"]
		}`),
		CreatedAt: time.Now(),
	}

	// Enqueue jobs
	err := jobQueue.Enqueue(job1)
	require.NoError(t, err)

	err = jobQueue.Enqueue(job2)
	require.NoError(t, err)

	// Verify jobs are queued
	assert.Equal(t, 2, jobQueue.Count())
	assert.Equal(t, 2, jobQueue.CountByStatus(queue.JobPending))

	// Test dequeue
	dequeued1, err := jobQueue.Dequeue(context.Background())
	require.NoError(t, err)
	require.NotNil(t, dequeued1)
	assert.Equal(t, "db-test-job-1", dequeued1.ID)
	assert.Equal(t, pb.JobType_JOB_TYPE_CHAT_COMPLETION, dequeued1.Type)

	// Update job status
	err = jobQueue.UpdateStatus(dequeued1.ID, queue.JobRunning)
	require.NoError(t, err)

	// Verify status update
	retrieved, err := jobQueue.Get(dequeued1.ID)
	require.NoError(t, err)
	assert.Equal(t, queue.JobRunning, retrieved.Status)

	// Update job with node assignment
	err = jobQueue.UpdateStatusAndNode(dequeued1.ID, queue.JobCompleted, "test-node")
	require.NoError(t, err)

	// Verify final state
	finalJob, err := jobQueue.Get(dequeued1.ID)
	require.NoError(t, err)
	assert.Equal(t, queue.JobCompleted, finalJob.Status)
	assert.Equal(t, "test-node", finalJob.AssignedNode)

	// Test remaining job
	assert.Equal(t, 1, jobQueue.Count())
	dequeued2, err := jobQueue.Dequeue(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "db-test-job-2", dequeued2.ID)

	// Test job completion
	err = jobQueue.CompleteJob(dequeued2.ID, []byte(`{"result": "success"}`))
	require.NoError(t, err)

	completedJob, err := jobQueue.Get(dequeued2.ID)
	require.NoError(t, err)
	assert.Equal(t, queue.JobCompleted, completedJob.Status)
	assert.Equal(t, []byte(`{"result": "success"}`), completedJob.Result)

	// Test job failure
	job3 := &queue.Job{
		ID:     "db-test-job-3",
		Type:   queue.JobTypeChatCompletion,
		Status: queue.JobRunning,
		Payload: []byte(`{"test": "data"}`),
		CreatedAt: time.Now(),
	}

	err = jobQueue.Enqueue(job3)
	require.NoError(t, err)

	err = jobQueue.FailJob(job3.ID, "Test failure reason")
	require.NoError(t, err)

	failedJob, err := jobQueue.Get(job3.ID)
	require.NoError(t, err)
	assert.Equal(t, queue.JobFailed, failedJob.Status)
	assert.Equal(t, "Test failure reason", failedJob.ErrorMessage)
}

func TestDatabase_Concurrent_Access(t *testing.T) {
	registry := node.NewInMemoryRegistry()
	jobQueue := queue.NewJobQueue()

	// Test concurrent node registry access
	t.Run("concurrent registry access", func(t *testing.T) {
		const numGoroutines = 10
		const operationsPerGoroutine = 5

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < operationsPerGoroutine; j++ {
					nodeID := fmt.Sprintf("concurrent-node-%d-%d", id, j)

					node := &pb.Node{
						Id:           nodeID,
						Hostname:     fmt.Sprintf("host-%d-%d", id, j),
						LastSeenUnix: time.Now().Unix(),
						Capabilities: &pb.Capabilities{
							Cpu:    "4 cores",
							Memory: "8GB",
							Os:     "linux",
						},
					}

					// Register node
					err := registry.Register(node)
					assert.NoError(t, err)

					// Update heartbeat
					err = registry.UpdateHeartbeat(nodeID)
					assert.NoError(t, err)

					// Get node
					_, exists := registry.Get(nodeID)
					assert.True(t, exists)
				}
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify final state
		allNodes := registry.List()
		assert.Len(t, allNodes, numGoroutines*operationsPerGoroutine)
	})

	// Test concurrent job queue access
	t.Run("concurrent job queue access", func(t *testing.T) {
		const numGoroutines = 5
		const jobsPerGoroutine = 3

		done := make(chan bool, numGoroutines)

		for i := 0; i < numGoroutines; i++ {
			go func(id int) {
				defer func() { done <- true }()

				for j := 0; j < jobsPerGoroutine; j++ {
					jobID := fmt.Sprintf("concurrent-job-%d-%d", id, j)

					job := &queue.Job{
						ID:     jobID,
						Type:   pb.JobType_JOB_TYPE_CHAT_COMPLETION,
						Status: pb.JobStatus_JOB_STATUS_PENDING,
						Payload: []byte(fmt.Sprintf(`{"id": "%s"}`, jobID)),
						CreatedAt: time.Now(),
					}

					// Enqueue job
					err := jobQueue.Enqueue(job)
					assert.NoError(t, err)

					// Update status
					err = jobQueue.UpdateStatus(jobID, pb.JobStatus_JOB_STATUS_RUNNING)
					assert.NoError(t, err)
				}
			}(i)
		}

		// Wait for all goroutines
		for i := 0; i < numGoroutines; i++ {
			<-done
		}

		// Verify final state
		totalJobs := numGoroutines * jobsPerGoroutine
		assert.Equal(t, totalJobs, jobQueue.Count())
		assert.Equal(t, totalJobs, jobQueue.CountByStatus(queue.JobRunning))
	})
}

func TestDatabase_Data_Integrity(t *testing.T) {
	registry := node.NewInMemoryRegistry()
	jobQueue := queue.NewJobQueue()

	// Test node data integrity
	t.Run("node data integrity", func(t *testing.T) {
		node := &pb.Node{
			Id:           "integrity-test-node",
			Hostname:     "integrity-host",
			LastSeenUnix: time.Now().Unix(),
			Capabilities: &pb.Capabilities{
				Cpu:       "4 cores",
				Memory:    "8GB",
				Os:        "linux",
				GpuType:   "NVIDIA RTX 3080",
				GpuVramTotal: "10GB",
			},
		}

		err := registry.Register(node)
		require.NoError(t, err)

		// Verify all fields are preserved
		retrieved, exists := registry.Get("integrity-test-node")
		require.True(t, exists)
		assert.Equal(t, node.Id, retrieved.Id)
		assert.Equal(t, node.Hostname, retrieved.Hostname)
		assert.Equal(t, node.LastSeenUnix, retrieved.LastSeenUnix)
		assert.Equal(t, node.Capabilities.Cpu, retrieved.Capabilities.Cpu)
		assert.Equal(t, node.Capabilities.Memory, retrieved.Capabilities.Memory)
		assert.Equal(t, node.Capabilities.Os, retrieved.Capabilities.Os)
		assert.Equal(t, node.Capabilities.GpuType, retrieved.Capabilities.GpuType)
		assert.Equal(t, node.Capabilities.GpuVramTotal, retrieved.Capabilities.GpuVramTotal)
	})

	// Test job data integrity
	t.Run("job data integrity", func(t *testing.T) {
		job := &queue.Job{
			ID:     "integrity-test-job",
			Type:   queue.JobTypeEmbeddings,
			Status: queue.JobPending,
			Payload: []byte(`{
				"model": "text-embedding-ada-002",
				"input": ["integrity test input"],
				"user": "test-user"
			}`),
			CreatedAt: time.Now(),
		}

		err := jobQueue.Enqueue(job)
		require.NoError(t, err)

		// Process job
		dequeued, err := jobQueue.Dequeue(context.Background())
		require.NoError(t, err)

		// Complete with result
		result := []byte(`{"embeddings": [[0.1, 0.2, 0.3]]}`)
		err = jobQueue.CompleteJob(dequeued.ID, result)
		require.NoError(t, err)

		// Verify final state
		finalJob, err := jobQueue.Get(dequeued.ID)
		require.NoError(t, err)
		assert.Equal(t, queue.JobCompleted, finalJob.Status)
		assert.Equal(t, result, finalJob.Result)
		assert.NotZero(t, finalJob.CompletedAt)
		assert.True(t, finalJob.CompletedAt.After(finalJob.CreatedAt))
	})
}