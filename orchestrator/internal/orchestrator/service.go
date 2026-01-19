package orchestrator

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	"github.com/Orchion/Orchion/orchestrator/internal/queue"
	"github.com/Orchion/Orchion/orchestrator/internal/scheduler"
)

// Service implements the Orchion gRPC service
type Service struct {
	pb.UnimplementedOrchestratorServer
	registry  node.Registry
	queue     *queue.JobQueue
	scheduler scheduler.Scheduler
}

// NewService creates a new orchestrator service
func NewService(registry node.Registry, jobQueue *queue.JobQueue, sched scheduler.Scheduler) *Service {
	return &Service{
		registry:  registry,
		queue:     jobQueue,
		scheduler: sched,
	}
}

// GetQueue returns the job queue (for internal use)
func (s *Service) GetQueue() *queue.JobQueue {
	return s.queue
}

// RegisterNode registers a new node with the orchestrator
func (s *Service) RegisterNode(ctx context.Context, req *pb.RegisterNodeRequest) (*pb.RegisterNodeResponse, error) {
	if req.Node == nil {
		return nil, status.Error(codes.InvalidArgument, "node is required")
	}

	if req.Node.Id == "" {
		return nil, status.Error(codes.InvalidArgument, "node.id is required")
	}

	if err := s.registry.Register(req.Node); err != nil {
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.RegisterNodeResponse{}, nil
}

// Heartbeat updates the heartbeat timestamp for a node
func (s *Service) Heartbeat(ctx context.Context, req *pb.HeartbeatRequest) (*pb.HeartbeatResponse, error) {
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	if err := s.registry.UpdateHeartbeat(req.NodeId); err != nil {
		if err == node.ErrNodeNotFound {
			return nil, status.Error(codes.NotFound, "node not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.HeartbeatResponse{}, nil
}

// UpdateNode updates a node's capabilities
func (s *Service) UpdateNode(ctx context.Context, req *pb.UpdateNodeRequest) (*pb.UpdateNodeResponse, error) {
	if req.NodeId == "" {
		return nil, status.Error(codes.InvalidArgument, "node_id is required")
	}

	if req.Capabilities == nil {
		return nil, status.Error(codes.InvalidArgument, "capabilities is required")
	}

	if err := s.registry.UpdateCapabilities(req.NodeId, req.Capabilities); err != nil {
		if err == node.ErrNodeNotFound {
			return nil, status.Error(codes.NotFound, "node not found")
		}
		return nil, status.Error(codes.Internal, err.Error())
	}

	return &pb.UpdateNodeResponse{}, nil
}

// ListNodes returns all registered nodes
func (s *Service) ListNodes(ctx context.Context, req *pb.ListNodesRequest) (*pb.ListNodesResponse, error) {
	nodes := s.registry.List()
	return &pb.ListNodesResponse{Nodes: nodes}, nil
}

func (s *Service) SubmitJob(ctx context.Context, req *pb.SubmitJobRequest) (*pb.SubmitJobResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}

	// Convert proto job type to internal job type
	var jobType queue.JobType
	switch req.JobType {
	case pb.JobType_JOB_TYPE_CHAT_COMPLETION:
		jobType = queue.JobTypeChatCompletion
	case pb.JobType_JOB_TYPE_EMBEDDINGS:
		jobType = queue.JobTypeEmbeddings
	default:
		return nil, status.Error(codes.InvalidArgument, "job_type is required")
	}

	job := &queue.Job{
		ID:      req.JobId,
		Type:    jobType,
		Payload: req.Payload,
		Status:  queue.JobPending,
	}

	s.queue.Enqueue(job)

	return &pb.SubmitJobResponse{
		JobId:  job.ID,
		Status: pb.JobStatus_JOB_STATUS_PENDING,
	}, nil
}

// GetJobStatus returns the status of a job
func (s *Service) GetJobStatus(ctx context.Context, req *pb.GetJobStatusRequest) (*pb.GetJobStatusResponse, error) {
	if req.JobId == "" {
		return nil, status.Error(codes.InvalidArgument, "job_id is required")
	}

	job, found := s.queue.Get(req.JobId)
	if !found {
		return nil, status.Error(codes.NotFound, "job not found")
	}

	// Convert internal status to proto status
	var protoStatus pb.JobStatus
	switch job.Status {
	case queue.JobPending:
		protoStatus = pb.JobStatus_JOB_STATUS_PENDING
	case queue.JobAssigned:
		protoStatus = pb.JobStatus_JOB_STATUS_ASSIGNED
	case queue.JobRunning:
		protoStatus = pb.JobStatus_JOB_STATUS_RUNNING
	case queue.JobCompleted:
		protoStatus = pb.JobStatus_JOB_STATUS_COMPLETED
	case queue.JobFailed:
		protoStatus = pb.JobStatus_JOB_STATUS_FAILED
	default:
		protoStatus = pb.JobStatus_JOB_STATUS_UNSPECIFIED
	}

	return &pb.GetJobStatusResponse{
		JobId:        job.ID,
		Status:       protoStatus,
		AssignedNode: job.AssignedNode,
		ErrorMessage: job.ErrorMessage,
		Result:       job.Result,
	}, nil
}
