package scheduler

// Task represents a unit of work to be scheduled
type Task struct {
	ID       string
	Metadata map[string]string
	Payload  []byte
}

// Scheduler is responsible for scheduling tasks across nodes
type Scheduler interface {
	// ScheduleTask assigns a task to a suitable node
	ScheduleTask(task *Task) (nodeID string, err error)
	
	// GetSchedulingStats returns statistics about the scheduler's decisions
	GetSchedulingStats() *SchedulingStats
}

// SchedulingStats contains statistics about scheduling decisions
type SchedulingStats struct {
	TotalScheduled int64
	FailedSchedules int64
}
