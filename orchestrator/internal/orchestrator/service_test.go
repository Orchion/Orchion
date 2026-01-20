package orchestrator

import (
	"context"
	"time"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	"github.com/Orchion/Orchion/orchestrator/internal/queue"
)

// MockRegistry is a mock implementation of node.Registry
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
	return args.Get(0).(*pb.Node), args.Bool(1)
}

func (m *MockRegistry) Remove(nodeID string) error {
	args := m.Called(nodeID)
	return args.Error(0)
}

func (m *MockRegistry) CheckHeartbeats(timeout time.Duration) []string {
	args := m.Called(timeout)
	return args.Get(0).([]string)
}

// MockScheduler is a mock implementation of scheduler.Scheduler
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
	mockQueue := queue.NewJobQueue()
	mockScheduler := &MockScheduler{}

	service := NewService(mockRegistry, mockQueue, mockScheduler)

	assert.NotNil(t, service)
	assert.Equal(t, mockRegistry, service.registry)
	assert.Equal(t, mockQueue, service.queue)
	assert.Equal(t, mockScheduler, service.scheduler)
}

func TestService_GetQueue(t *testing.T) {
	mockRegistry := &MockRegistry{}
	mockQueue := queue.NewJobQueue()
	mockScheduler := &MockScheduler{}

	service := NewService(mockRegistry, mockQueue, mockScheduler)

	// GetQueue returns *queue.JobQueue, but we're using MockJobQueue for testing
	assert.NotNil(t, service.GetQueue())
}

func TestService_RegisterNode(t *testing.T) {
	ctx := context.Background()

	t.Run("successful registration", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		node := &pb.Node{
			Id:       "test-node",
			Hostname: "test-host",
		}

		mockRegistry.On("Register", node).Return(nil)

		resp, err := service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockRegistry.AssertExpectations(t)
	})

	t.Run("nil node", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: nil})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "node is required")
	})

	t.Run("empty node ID", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		node := &pb.Node{
			Id:       "",
			Hostname: "test-host",
		}

		resp, err := service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "node.id is required")
	})

	t.Run("registry error", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		node := &pb.Node{
			Id:       "test-node",
			Hostname: "test-host",
		}

		mockRegistry.On("Register", node).Return(assert.AnError)

		resp, err := service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		mockRegistry.AssertExpectations(t)
	})
}

func TestService_Heartbeat(t *testing.T) {
	ctx := context.Background()

	t.Run("successful heartbeat", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		mockRegistry.On("UpdateHeartbeat", "test-node").Return(nil)

		resp, err := service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "test-node"})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockRegistry.AssertExpectations(t)
	})

	t.Run("empty node ID", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: ""})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "node_id is required")
	})

	t.Run("node not found", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		mockRegistry.On("UpdateHeartbeat", "non-existent").Return(node.ErrNodeNotFound)

		resp, err := service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "non-existent"})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), "node not found")
		mockRegistry.AssertExpectations(t)
	})

	t.Run("registry error", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		mockRegistry.On("UpdateHeartbeat", "test-node").Return(assert.AnError)

		resp, err := service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "test-node"})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.Internal, st.Code())
		mockRegistry.AssertExpectations(t)
	})
}

func TestService_UpdateNode(t *testing.T) {
	ctx := context.Background()

	t.Run("successful update", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		capabilities := &pb.Capabilities{
			Cpu:    "8 cores",
			Memory: "16GB",
		}

		mockRegistry.On("UpdateCapabilities", "test-node", capabilities).Return(nil)

		resp, err := service.UpdateNode(ctx, &pb.UpdateNodeRequest{
			NodeId:       "test-node",
			Capabilities: capabilities,
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		mockRegistry.AssertExpectations(t)
	})

	t.Run("empty node ID", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.UpdateNode(ctx, &pb.UpdateNodeRequest{
			NodeId:       "",
			Capabilities: &pb.Capabilities{},
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "node_id is required")
	})

	t.Run("nil capabilities", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.UpdateNode(ctx, &pb.UpdateNodeRequest{
			NodeId:       "test-node",
			Capabilities: nil,
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "capabilities is required")
	})

	t.Run("node not found", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		capabilities := &pb.Capabilities{Cpu: "4 cores"}
		mockRegistry.On("UpdateCapabilities", "non-existent", capabilities).Return(node.ErrNodeNotFound)

		resp, err := service.UpdateNode(ctx, &pb.UpdateNodeRequest{
			NodeId:       "non-existent",
			Capabilities: capabilities,
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), "node not found")
		mockRegistry.AssertExpectations(t)
	})
}

func TestService_ListNodes(t *testing.T) {
	ctx := context.Background()

	t.Run("successful list", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		expectedNodes := []*pb.Node{
			{Id: "node-1", Hostname: "host-1"},
			{Id: "node-2", Hostname: "host-2"},
		}

		mockRegistry.On("List").Return(expectedNodes)

		resp, err := service.ListNodes(ctx, &pb.ListNodesRequest{})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, expectedNodes, resp.Nodes)
		mockRegistry.AssertExpectations(t)
	})

	t.Run("empty list", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		mockRegistry.On("List").Return([]*pb.Node{})

		resp, err := service.ListNodes(ctx, &pb.ListNodesRequest{})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Empty(t, resp.Nodes)
		mockRegistry.AssertExpectations(t)
	})
}

func TestService_SubmitJob(t *testing.T) {
	ctx := context.Background()

	t.Run("successful chat completion job submission", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		payload := []byte("test payload")

		resp, err := service.SubmitJob(ctx, &pb.SubmitJobRequest{
			JobId:   "job-123",
			JobType: pb.JobType_JOB_TYPE_CHAT_COMPLETION,
			Payload: payload,
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "job-123", resp.JobId)
		assert.Equal(t, pb.JobStatus_JOB_STATUS_PENDING, resp.Status)

		// Verify job was enqueued
		job, found := mockQueue.Get("job-123")
		assert.True(t, found)
		assert.Equal(t, "job-123", job.ID)
		assert.Equal(t, queue.JobTypeChatCompletion, job.Type)
		assert.Equal(t, queue.JobPending, job.Status)
		assert.Equal(t, payload, job.Payload)
	})

	t.Run("successful embeddings job submission", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		payload := []byte("embeddings payload")

		resp, err := service.SubmitJob(ctx, &pb.SubmitJobRequest{
			JobId:   "embed-job",
			JobType: pb.JobType_JOB_TYPE_EMBEDDINGS,
			Payload: payload,
		})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "embed-job", resp.JobId)
		assert.Equal(t, pb.JobStatus_JOB_STATUS_PENDING, resp.Status)

		// Verify job was enqueued
		job, found := mockQueue.Get("embed-job")
		assert.True(t, found)
		assert.Equal(t, "embed-job", job.ID)
		assert.Equal(t, queue.JobTypeEmbeddings, job.Type)
		assert.Equal(t, queue.JobPending, job.Status)
		assert.Equal(t, payload, job.Payload)
	})

	t.Run("empty job ID", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.SubmitJob(ctx, &pb.SubmitJobRequest{
			JobId:   "",
			JobType: pb.JobType_JOB_TYPE_CHAT_COMPLETION,
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "job_id is required")
	})

	t.Run("invalid job type", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.SubmitJob(ctx, &pb.SubmitJobRequest{
			JobId:   "job-123",
			JobType: pb.JobType_JOB_TYPE_UNSPECIFIED,
		})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "job_type is required")
	})
}

func TestService_GetJobStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("successful job status retrieval", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		job := &queue.Job{
			ID:           "job-123",
			Type:         queue.JobTypeChatCompletion,
			Status:       queue.JobRunning,
			AssignedNode: "node-456",
			ErrorMessage: "",
			Result:       nil,
		}

		// Manually add job to queue for testing
		mockQueue.Enqueue(job)

		resp, err := service.GetJobStatus(ctx, &pb.GetJobStatusRequest{JobId: "job-123"})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, "job-123", resp.JobId)
		assert.Equal(t, pb.JobStatus_JOB_STATUS_RUNNING, resp.Status)
		assert.Equal(t, "node-456", resp.AssignedNode)
	})

	t.Run("job with error", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		job := &queue.Job{
			ID:           "failed-job",
			Status:       queue.JobFailed,
			ErrorMessage: "Model not available",
		}

		// Manually add job to queue for testing
		mockQueue.Enqueue(job)

		resp, err := service.GetJobStatus(ctx, &pb.GetJobStatusRequest{JobId: "failed-job"})

		require.NoError(t, err)
		assert.NotNil(t, resp)
		assert.Equal(t, pb.JobStatus_JOB_STATUS_FAILED, resp.Status)
		assert.Equal(t, "Model not available", resp.ErrorMessage)
	})

	t.Run("empty job ID", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.GetJobStatus(ctx, &pb.GetJobStatusRequest{JobId: ""})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.InvalidArgument, st.Code())
		assert.Contains(t, st.Message(), "job_id is required")
	})

	t.Run("job not found", func(t *testing.T) {
		mockRegistry := &MockRegistry{}
		mockQueue := queue.NewJobQueue()
		mockScheduler := &MockScheduler{}

		service := NewService(mockRegistry, mockQueue, mockScheduler)

		resp, err := service.GetJobStatus(ctx, &pb.GetJobStatusRequest{JobId: "non-existent"})

		require.Error(t, err)
		assert.Nil(t, resp)
		st, ok := status.FromError(err)
		assert.True(t, ok)
		assert.Equal(t, codes.NotFound, st.Code())
		assert.Contains(t, st.Message(), "job not found")
	})

	t.Run("all job statuses", func(t *testing.T) {
		statuses := []struct {
			internalStatus queue.JobStatus
			protoStatus    pb.JobStatus
		}{
			{queue.JobPending, pb.JobStatus_JOB_STATUS_PENDING},
			{queue.JobAssigned, pb.JobStatus_JOB_STATUS_ASSIGNED},
			{queue.JobRunning, pb.JobStatus_JOB_STATUS_RUNNING},
			{queue.JobCompleted, pb.JobStatus_JOB_STATUS_COMPLETED},
			{queue.JobFailed, pb.JobStatus_JOB_STATUS_FAILED},
		}

		for _, statusCase := range statuses {
			t.Run(statusCase.internalStatus.String(), func(t *testing.T) {
				mockRegistry := &MockRegistry{}
				mockQueue := queue.NewJobQueue()
				mockScheduler := &MockScheduler{}

				service := NewService(mockRegistry, mockQueue, mockScheduler)

				job := &queue.Job{
					ID:     "test-job",
					Status: statusCase.internalStatus,
				}

				// Manually add job to queue for testing
				mockQueue.Enqueue(job)

				resp, err := service.GetJobStatus(ctx, &pb.GetJobStatusRequest{JobId: "test-job"})

				require.NoError(t, err)
				assert.Equal(t, statusCase.protoStatus, resp.Status)
			})
		}
	})
}