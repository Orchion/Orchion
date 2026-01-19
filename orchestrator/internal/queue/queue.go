package queue

import (
	"sync"
	"time"
)

// JobStatus represents the status of a job
type JobStatus int

const (
	JobPending JobStatus = iota
	JobAssigned
	JobRunning
	JobCompleted
	JobFailed
)

// String returns the string representation of JobStatus
func (s JobStatus) String() string {
	switch s {
	case JobPending:
		return "pending"
	case JobAssigned:
		return "assigned"
	case JobRunning:
		return "running"
	case JobCompleted:
		return "completed"
	case JobFailed:
		return "failed"
	default:
		return "unknown"
	}
}

// JobType represents the type of job
type JobType int

const (
	JobTypeUnspecified JobType = iota
	JobTypeChatCompletion
	JobTypeEmbeddings
)

// Job represents a job in the queue
type Job struct {
	ID           string
	Type         JobType
	Payload      []byte // Serialized request (ChatCompletionRequest or EmbeddingRequest)
	Status       JobStatus
	CreatedAt    time.Time
	UpdatedAt    time.Time
	AssignedNode string
	Result       []byte // Serialized response when completed
	ErrorMessage string // Error message if failed
}

// JobQueue is a concurrency-safe in-memory job queue
type JobQueue struct {
	mu    sync.Mutex
	cond  *sync.Cond
	jobs  []*Job
	index map[string]*Job
}

// NewJobQueue creates a new job queue
func NewJobQueue() *JobQueue {
	jq := &JobQueue{
		jobs:  make([]*Job, 0),
		index: make(map[string]*Job),
	}
	jq.cond = sync.NewCond(&jq.mu)
	return jq
}

// Enqueue adds a job to the queue
func (q *JobQueue) Enqueue(job *Job) {
	q.mu.Lock()
	defer q.mu.Unlock()

	job.CreatedAt = time.Now()
	job.UpdatedAt = time.Now()
	if job.Status == 0 {
		job.Status = JobPending
	}

	q.jobs = append(q.jobs, job)
	q.index[job.ID] = job
	q.cond.Signal()
}

// Dequeue removes and returns the next job from the queue
// This blocks until a job is available
func (q *JobQueue) Dequeue() *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	for len(q.jobs) == 0 {
		q.cond.Wait()
	}

	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return job
}

// DequeueWithTimeout attempts to dequeue a job with a timeout
// Returns nil if timeout expires before a job is available
func (q *JobQueue) DequeueWithTimeout(timeout time.Duration) *Job {
	// Start a timer that will broadcast to wake us up
	timer := time.AfterFunc(timeout, func() {
		q.cond.Broadcast()
	})
	defer timer.Stop()

	q.mu.Lock()
	defer q.mu.Unlock()

	deadline := time.Now().Add(timeout)

	for len(q.jobs) == 0 {
		if time.Now().After(deadline) {
			return nil
		}
		q.cond.Wait()
	}

	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return job
}

// DequeueNonBlocking attempts to dequeue a job without blocking
// Returns nil if no jobs are available
func (q *JobQueue) DequeueNonBlocking() *Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	if len(q.jobs) == 0 {
		return nil
	}

	job := q.jobs[0]
	q.jobs = q.jobs[1:]
	return job
}

// Get retrieves a job by ID
func (q *JobQueue) Get(id string) (*Job, bool) {
	q.mu.Lock()
	defer q.mu.Unlock()
	job, ok := q.index[id]
	return job, ok
}

// UpdateStatus updates the status of a job by ID
func (q *JobQueue) UpdateStatus(id string, status JobStatus) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if job, ok := q.index[id]; ok {
		job.Status = status
		job.UpdatedAt = time.Now()
	}
}

// UpdateStatusAndNode updates both the status and assigned node of a job
func (q *JobQueue) UpdateStatusAndNode(id string, status JobStatus, nodeID string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if job, ok := q.index[id]; ok {
		job.Status = status
		job.AssignedNode = nodeID
		job.UpdatedAt = time.Now()
	}
}

// CompleteJob marks a job as completed with a result
func (q *JobQueue) CompleteJob(id string, result []byte) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if job, ok := q.index[id]; ok {
		job.Status = JobCompleted
		job.Result = result
		job.UpdatedAt = time.Now()
	}
}

// FailJob marks a job as failed with an error message
func (q *JobQueue) FailJob(id string, errorMsg string) {
	q.mu.Lock()
	defer q.mu.Unlock()
	if job, ok := q.index[id]; ok {
		job.Status = JobFailed
		job.ErrorMessage = errorMsg
		job.UpdatedAt = time.Now()
	}
}

// List returns all jobs in the queue
func (q *JobQueue) List() []*Job {
	q.mu.Lock()
	defer q.mu.Unlock()

	jobs := make([]*Job, 0, len(q.index))
	for _, job := range q.index {
		jobs = append(jobs, job)
	}
	return jobs
}

// Count returns the number of jobs in the queue
func (q *JobQueue) Count() int {
	q.mu.Lock()
	defer q.mu.Unlock()
	return len(q.jobs)
}

// CountByStatus returns the number of jobs with a specific status
func (q *JobQueue) CountByStatus(status JobStatus) int {
	q.mu.Lock()
	defer q.mu.Unlock()

	count := 0
	for _, job := range q.index {
		if job.Status == status {
			count++
		}
	}
	return count
}
