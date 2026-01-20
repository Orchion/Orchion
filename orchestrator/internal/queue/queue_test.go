package queue

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestJobStatus_String(t *testing.T) {
	testCases := []struct {
		status   JobStatus
		expected string
	}{
		{JobPending, "pending"},
		{JobAssigned, "assigned"},
		{JobRunning, "running"},
		{JobCompleted, "completed"},
		{JobFailed, "failed"},
		{JobStatus(999), "unknown"}, // Invalid status
	}

	for _, tc := range testCases {
		t.Run(tc.expected, func(t *testing.T) {
			assert.Equal(t, tc.expected, tc.status.String())
		})
	}
}

func TestNewJobQueue(t *testing.T) {
	queue := NewJobQueue()
	assert.NotNil(t, queue)
	assert.NotNil(t, queue.jobs)
	assert.NotNil(t, queue.index)
	assert.Equal(t, 0, len(queue.jobs))
	assert.Equal(t, 0, len(queue.index))
	assert.NotNil(t, queue.cond)
}

func TestJobQueue_Enqueue(t *testing.T) {
	queue := NewJobQueue()

	t.Run("enqueue with default status", func(t *testing.T) {
		job := &Job{
			ID:      "job-1",
			Type:    JobTypeChatCompletion,
			Payload: []byte("test payload"),
		}

		beforeEnqueue := time.Now()
		queue.Enqueue(job)
		afterEnqueue := time.Now()

		// Verify job was added
		assert.Equal(t, 1, len(queue.jobs))
		assert.Contains(t, queue.index, "job-1")

		// Verify timestamps were set
		assert.True(t, job.CreatedAt.After(beforeEnqueue) || job.CreatedAt.Equal(beforeEnqueue))
		assert.True(t, job.CreatedAt.Before(afterEnqueue) || job.CreatedAt.Equal(afterEnqueue))
		assert.True(t, job.UpdatedAt.After(beforeEnqueue) || job.UpdatedAt.Equal(beforeEnqueue))
		assert.True(t, job.UpdatedAt.Before(afterEnqueue) || job.UpdatedAt.Equal(afterEnqueue))

		// Verify default status was set
		assert.Equal(t, JobPending, job.Status)
	})

	t.Run("enqueue with explicit status", func(t *testing.T) {
		job := &Job{
			ID:      "job-2",
			Type:    JobTypeEmbeddings,
			Payload: []byte("embeddings payload"),
			Status:  JobRunning, // Explicit status
		}

		queue.Enqueue(job)

		// Verify explicit status was preserved
		assert.Equal(t, JobRunning, job.Status)
		retrieved, exists := queue.Get("job-2")
		assert.True(t, exists)
		assert.Equal(t, JobRunning, retrieved.Status)
	})

	t.Run("enqueue multiple jobs", func(t *testing.T) {
		queue := NewJobQueue() // Fresh queue

		jobs := []*Job{
			{ID: "multi-1", Type: JobTypeChatCompletion},
			{ID: "multi-2", Type: JobTypeEmbeddings},
			{ID: "multi-3", Type: JobTypeChatCompletion},
		}

		for _, job := range jobs {
			queue.Enqueue(job)
		}

		assert.Equal(t, 3, len(queue.jobs))
		assert.Equal(t, 3, len(queue.index))

		// Verify FIFO order
		assert.Equal(t, "multi-1", queue.jobs[0].ID)
		assert.Equal(t, "multi-2", queue.jobs[1].ID)
		assert.Equal(t, "multi-3", queue.jobs[2].ID)
	})
}

func TestJobQueue_Dequeue(t *testing.T) {
	t.Run("dequeue from non-empty queue", func(t *testing.T) {
		queue := NewJobQueue()

		job1 := &Job{ID: "job-1", Type: JobTypeChatCompletion}
		job2 := &Job{ID: "job-2", Type: JobTypeEmbeddings}

		queue.Enqueue(job1)
		queue.Enqueue(job2)

		// Dequeue first job
		dequeued := queue.Dequeue()
		assert.NotNil(t, dequeued)
		assert.Equal(t, "job-1", dequeued.ID)
		assert.Equal(t, JobTypeChatCompletion, dequeued.Type)

		// Verify queue state - jobs slice should have one less item
		assert.Equal(t, 1, len(queue.jobs))
		assert.Equal(t, "job-2", queue.jobs[0].ID)
		// Note: The index map retains all jobs, this seems to be current behavior
		assert.Contains(t, queue.index, "job-1")
		assert.Contains(t, queue.index, "job-2")
	})

	t.Run("dequeue all jobs", func(t *testing.T) {
		queue := NewJobQueue()

		job := &Job{ID: "only-job", Type: JobTypeChatCompletion}
		queue.Enqueue(job)

		dequeued := queue.Dequeue()
		assert.NotNil(t, dequeued)
		assert.Equal(t, "only-job", dequeued.ID)

		// Jobs slice should be empty, but index retains the job
		assert.Equal(t, 0, len(queue.jobs))
		assert.Equal(t, 1, len(queue.index)) // Current implementation keeps jobs in index
	})
}

func TestJobQueue_DequeueWithTimeout(t *testing.T) {
	t.Run("dequeue with timeout - job available", func(t *testing.T) {
		queue := NewJobQueue()

		job := &Job{ID: "timeout-job", Type: JobTypeChatCompletion}
		queue.Enqueue(job)

		dequeued := queue.DequeueWithTimeout(100 * time.Millisecond)
		assert.NotNil(t, dequeued)
		assert.Equal(t, "timeout-job", dequeued.ID)
	})

	t.Run("dequeue with timeout - no job available", func(t *testing.T) {
		queue := NewJobQueue()

		start := time.Now()
		dequeued := queue.DequeueWithTimeout(50 * time.Millisecond)
		elapsed := time.Since(start)

		assert.Nil(t, dequeued)
		// Should have waited at least the timeout duration
		assert.True(t, elapsed >= 45*time.Millisecond, "Should have waited at least 45ms, waited %v", elapsed)
	})

	t.Run("dequeue with timeout - job arrives in time", func(t *testing.T) {
		queue := NewJobQueue()

		// Start dequeue in background
		var dequeued *Job
		done := make(chan bool)
		go func() {
			dequeued = queue.DequeueWithTimeout(200 * time.Millisecond)
			done <- true
		}()

		// Wait a bit, then enqueue job
		time.Sleep(50 * time.Millisecond)
		job := &Job{ID: "late-job", Type: JobTypeChatCompletion}
		queue.Enqueue(job)

		// Wait for dequeue to complete
		select {
		case <-done:
			assert.NotNil(t, dequeued)
			assert.Equal(t, "late-job", dequeued.ID)
		case <-time.After(300 * time.Millisecond):
			t.Fatal("Dequeue should have completed")
		}
	})
}

func TestJobQueue_DequeueNonBlocking(t *testing.T) {
	t.Run("dequeue non-blocking - job available", func(t *testing.T) {
		queue := NewJobQueue()

		job := &Job{ID: "nonblock-job", Type: JobTypeChatCompletion}
		queue.Enqueue(job)

		dequeued := queue.DequeueNonBlocking()
		assert.NotNil(t, dequeued)
		assert.Equal(t, "nonblock-job", dequeued.ID)

		// Queue should be empty
		assert.Equal(t, 0, len(queue.jobs))
	})

	t.Run("dequeue non-blocking - no job available", func(t *testing.T) {
		queue := NewJobQueue()

		dequeued := queue.DequeueNonBlocking()
		assert.Nil(t, dequeued)
	})
}

func TestJobQueue_Get(t *testing.T) {
	queue := NewJobQueue()

	t.Run("get existing job", func(t *testing.T) {
		job := &Job{ID: "get-job", Type: JobTypeChatCompletion}
		queue.Enqueue(job)

		retrieved, exists := queue.Get("get-job")
		assert.True(t, exists)
		assert.NotNil(t, retrieved)
		assert.Equal(t, "get-job", retrieved.ID)
		assert.Equal(t, JobTypeChatCompletion, retrieved.Type)
	})

	t.Run("get non-existent job", func(t *testing.T) {
		retrieved, exists := queue.Get("non-existent")
		assert.False(t, exists)
		assert.Nil(t, retrieved)
	})
}

func TestJobQueue_UpdateStatus(t *testing.T) {
	queue := NewJobQueue()

	job := &Job{ID: "status-job", Type: JobTypeChatCompletion, Status: JobPending}
	queue.Enqueue(job)

	originalTime := job.UpdatedAt

	// Update status
	time.Sleep(1 * time.Millisecond) // Ensure time difference
	queue.UpdateStatus("status-job", JobRunning)

	// Verify status was updated
	retrieved, exists := queue.Get("status-job")
	assert.True(t, exists)
	assert.Equal(t, JobRunning, retrieved.Status)
	assert.True(t, retrieved.UpdatedAt.After(originalTime))
}

func TestJobQueue_UpdateStatusAndNode(t *testing.T) {
	queue := NewJobQueue()

	job := &Job{ID: "node-job", Type: JobTypeChatCompletion, Status: JobPending}
	queue.Enqueue(job)

	originalTime := job.UpdatedAt

	// Update status and node
	time.Sleep(1 * time.Millisecond)
	queue.UpdateStatusAndNode("node-job", JobAssigned, "node-456")

	// Verify updates
	retrieved, exists := queue.Get("node-job")
	assert.True(t, exists)
	assert.Equal(t, JobAssigned, retrieved.Status)
	assert.Equal(t, "node-456", retrieved.AssignedNode)
	assert.True(t, retrieved.UpdatedAt.After(originalTime))
}

func TestJobQueue_CompleteJob(t *testing.T) {
	queue := NewJobQueue()

	job := &Job{ID: "complete-job", Type: JobTypeChatCompletion}
	queue.Enqueue(job)

	result := []byte("completion result")
	originalTime := job.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	queue.CompleteJob("complete-job", result)

	// Verify completion
	retrieved, exists := queue.Get("complete-job")
	assert.True(t, exists)
	assert.Equal(t, JobCompleted, retrieved.Status)
	assert.Equal(t, result, retrieved.Result)
	assert.True(t, retrieved.UpdatedAt.After(originalTime))
}

func TestJobQueue_FailJob(t *testing.T) {
	queue := NewJobQueue()

	job := &Job{ID: "fail-job", Type: JobTypeChatCompletion}
	queue.Enqueue(job)

	errorMsg := "job failed due to error"
	originalTime := job.UpdatedAt

	time.Sleep(1 * time.Millisecond)
	queue.FailJob("fail-job", errorMsg)

	// Verify failure
	retrieved, exists := queue.Get("fail-job")
	assert.True(t, exists)
	assert.Equal(t, JobFailed, retrieved.Status)
	assert.Equal(t, errorMsg, retrieved.ErrorMessage)
	assert.True(t, retrieved.UpdatedAt.After(originalTime))
}

func TestJobQueue_List(t *testing.T) {
	queue := NewJobQueue()

	// Add jobs
	jobs := []*Job{
		{ID: "list-1", Type: JobTypeChatCompletion},
		{ID: "list-2", Type: JobTypeEmbeddings},
		{ID: "list-3", Type: JobTypeChatCompletion},
	}

	for _, job := range jobs {
		queue.Enqueue(job)
	}

	// List all jobs
	listed := queue.List()
	assert.Len(t, listed, 3)

	// Convert to map for easier checking
	jobMap := make(map[string]*Job)
	for _, job := range listed {
		jobMap[job.ID] = job
	}

	assert.Contains(t, jobMap, "list-1")
	assert.Contains(t, jobMap, "list-2")
	assert.Contains(t, jobMap, "list-3")
	assert.Equal(t, JobTypeChatCompletion, jobMap["list-1"].Type)
	assert.Equal(t, JobTypeEmbeddings, jobMap["list-2"].Type)
}

func TestJobQueue_Count(t *testing.T) {
	queue := NewJobQueue()

	assert.Equal(t, 0, queue.Count())

	queue.Enqueue(&Job{ID: "count-1"})
	assert.Equal(t, 1, queue.Count())

	queue.Enqueue(&Job{ID: "count-2"})
	assert.Equal(t, 2, queue.Count())

	queue.Dequeue()
	assert.Equal(t, 1, queue.Count())
}

func TestJobQueue_CountByStatus(t *testing.T) {
	queue := NewJobQueue()

	// Add jobs with different statuses
	jobs := []*Job{
		{ID: "pending-1", Status: JobPending},
		{ID: "pending-2", Status: JobPending},
		{ID: "running-1", Status: JobRunning},
		{ID: "completed-1", Status: JobCompleted},
		{ID: "failed-1", Status: JobFailed},
	}

	for _, job := range jobs {
		queue.Enqueue(job)
	}

	assert.Equal(t, 2, queue.CountByStatus(JobPending))
	assert.Equal(t, 1, queue.CountByStatus(JobRunning))
	assert.Equal(t, 1, queue.CountByStatus(JobCompleted))
	assert.Equal(t, 1, queue.CountByStatus(JobFailed))
	assert.Equal(t, 0, queue.CountByStatus(JobAssigned))
}

func TestJobQueue_Concurrency(t *testing.T) {
	queue := NewJobQueue()
	const numGoroutines = 10
	const operationsPerGoroutine = 50

	var wg sync.WaitGroup
	errors := make(chan error, numGoroutines*operationsPerGoroutine)

	// Start multiple goroutines performing operations
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			for j := 0; j < operationsPerGoroutine; j++ {
				jobID := fmt.Sprintf("concurrency-%d-%d", id, j)

				// Enqueue
				job := &Job{ID: jobID, Type: JobTypeChatCompletion}
				queue.Enqueue(job)

				// Update status
				queue.UpdateStatus(jobID, JobRunning)

				// Get
				_, exists := queue.Get(jobID)
				if !exists {
					errors <- fmt.Errorf("job %s not found", jobID)
					return
				}

				// Complete job
				queue.CompleteJob(jobID, []byte("result"))

				// Dequeue non-blocking (may or may not get the job due to concurrency)
				queue.DequeueNonBlocking()
			}
		}(i)
	}

	wg.Wait()
	close(errors)

	// Check for errors
	for err := range errors {
		t.Error(err)
	}

	// Final state should be consistent
	jobs := queue.List()
	for _, job := range jobs {
		// All remaining jobs should be in a valid state
		assert.True(t, job.Status >= JobPending && job.Status <= JobFailed)
	}
}

// Benchmark tests
func BenchmarkJobQueue_Enqueue(b *testing.B) {
	queue := NewJobQueue()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		job := &Job{
			ID:   fmt.Sprintf("bench-%d", i),
			Type: JobTypeChatCompletion,
		}
		queue.Enqueue(job)
	}
}

func BenchmarkJobQueue_DequeueNonBlocking(b *testing.B) {
	queue := NewJobQueue()

	// Pre-populate queue
	for i := 0; i < b.N; i++ {
		job := &Job{
			ID:   fmt.Sprintf("bench-%d", i),
			Type: JobTypeChatCompletion,
		}
		queue.Enqueue(job)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.DequeueNonBlocking()
	}
}

func BenchmarkJobQueue_Get(b *testing.B) {
	queue := NewJobQueue()

	// Pre-populate queue
	job := &Job{ID: "bench-job", Type: JobTypeChatCompletion}
	queue.Enqueue(job)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		queue.Get("bench-job")
	}
}