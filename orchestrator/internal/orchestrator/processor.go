package orchestrator

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/proto"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	"github.com/Orchion/Orchion/orchestrator/internal/queue"
	"github.com/Orchion/Orchion/orchestrator/internal/scheduler"
)

// JobProcessor processes jobs from the queue and assigns them to nodes
type JobProcessor struct {
	queue       *queue.JobQueue
	scheduler   scheduler.Scheduler
	registry    node.Registry
	nodeClients map[string]pb.NodeAgentClient
	mu          sync.RWMutex
}

// NewJobProcessor creates a new job processor
func NewJobProcessor(queue *queue.JobQueue, sched scheduler.Scheduler, registry node.Registry) *JobProcessor {
	return &JobProcessor{
		queue:       queue,
		scheduler:   sched,
		registry:    registry,
		nodeClients: make(map[string]pb.NodeAgentClient),
	}
}

// Start begins processing jobs in a goroutine
func (p *JobProcessor) Start(ctx context.Context) {
	go p.processLoop(ctx)
}

// processLoop continuously processes jobs from the queue
func (p *JobProcessor) processLoop(ctx context.Context) {
	log.Printf("Job processor started")
	defer log.Printf("Job processor stopped")

	for {
		select {
		case <-ctx.Done():
			return
		default:
			// Try to dequeue with a short timeout to allow context checking
			job := p.queue.DequeueWithTimeout(100 * time.Millisecond)
			if job != nil {
				// Process job in a separate goroutine to allow concurrent processing
				go p.processJob(ctx, job)
			}
		}
	}
}

// processJob assigns a job to a node and dispatches it
func (p *JobProcessor) processJob(ctx context.Context, job *queue.Job) {
	log.Printf("Processing job %s (type: %d)", job.ID, job.Type)

	// Update status to assigned
	p.queue.UpdateStatus(job.ID, queue.JobAssigned)

	// Select a node using the scheduler
	selectedNode, err := p.scheduler.SelectNode("", p.registry)
	if err != nil {
		log.Printf("Failed to select node for job %s: %v", job.ID, err)
		p.queue.FailJob(job.ID, fmt.Sprintf("failed to select node: %v", err))
		return
	}

	// Update job with assigned node
	p.queue.UpdateStatusAndNode(job.ID, queue.JobRunning, selectedNode.Id)
	log.Printf("Assigned job %s to node %s (%s)", job.ID, selectedNode.Id, selectedNode.AgentAddress)

	// Get or create gRPC client for this node
	client, err := p.getNodeClient(selectedNode.Id, selectedNode)
	if err != nil {
		log.Printf("Failed to connect to node %s for job %s: %v", selectedNode.Id, job.ID, err)
		p.queue.FailJob(job.ID, fmt.Sprintf("failed to connect to node: %v", err))
		return
	}

	// Dispatch job based on type
	switch job.Type {
	case queue.JobTypeChatCompletion:
		p.executeChatCompletion(ctx, job, client)
	case queue.JobTypeEmbeddings:
		p.executeEmbeddings(ctx, job, client)
	default:
		log.Printf("Unknown job type %d for job %s", job.Type, job.ID)
		p.queue.FailJob(job.ID, fmt.Sprintf("unknown job type: %d", job.Type))
	}
}

// executeChatCompletion executes a chat completion job on a node
func (p *JobProcessor) executeChatCompletion(ctx context.Context, job *queue.Job, client pb.NodeAgentClient) {
	// Deserialize the request from payload
	var req pb.ChatCompletionRequest
	if err := proto.Unmarshal(job.Payload, &req); err != nil {
		log.Printf("Failed to unmarshal chat completion request for job %s: %v", job.ID, err)
		p.queue.FailJob(job.ID, fmt.Sprintf("failed to unmarshal request: %v", err))
		return
	}

	// Call the node agent
	stream, err := client.ChatCompletion(ctx, &req)
	if err != nil {
		log.Printf("Failed to execute chat completion for job %s: %v", job.ID, err)
		p.queue.FailJob(job.ID, fmt.Sprintf("failed to execute: %v", err))
		return
	}

	// Collect all responses (for async jobs, we store the final result)
	var lastResponse *pb.ChatCompletionResponse
	for {
		resp, err := stream.Recv()
		if err != nil {
			// Check if it's end of stream
			if err == io.EOF {
				break
			}
			log.Printf("Error receiving chat completion response for job %s: %v", job.ID, err)
			p.queue.FailJob(job.ID, fmt.Sprintf("error receiving response: %v", err))
			return
		}
		lastResponse = resp
	}

	// Serialize the final response
	if lastResponse != nil {
		result, err := proto.Marshal(lastResponse)
		if err != nil {
			log.Printf("Failed to marshal response for job %s: %v", job.ID, err)
			p.queue.FailJob(job.ID, fmt.Sprintf("failed to marshal response: %v", err))
			return
		}
		p.queue.CompleteJob(job.ID, result)
		log.Printf("Completed chat completion job %s", job.ID)
	} else {
		p.queue.CompleteJob(job.ID, nil)
		log.Printf("Completed chat completion job %s (no response)", job.ID)
	}
}

// executeEmbeddings executes an embeddings job on a node
func (p *JobProcessor) executeEmbeddings(ctx context.Context, job *queue.Job, client pb.NodeAgentClient) {
	// Deserialize the request from payload
	var req pb.EmbeddingRequest
	if err := proto.Unmarshal(job.Payload, &req); err != nil {
		log.Printf("Failed to unmarshal embedding request for job %s: %v", job.ID, err)
		p.queue.FailJob(job.ID, fmt.Sprintf("failed to unmarshal request: %v", err))
		return
	}

	// Call the node agent
	resp, err := client.Embeddings(ctx, &req)
	if err != nil {
		log.Printf("Failed to execute embeddings for job %s: %v", job.ID, err)
		p.queue.FailJob(job.ID, fmt.Sprintf("failed to execute: %v", err))
		return
	}

	// Serialize the response
	result, err := proto.Marshal(resp)
	if err != nil {
		log.Printf("Failed to marshal response for job %s: %v", job.ID, err)
		p.queue.FailJob(job.ID, fmt.Sprintf("failed to marshal response: %v", err))
		return
	}

	p.queue.CompleteJob(job.ID, result)
	log.Printf("Completed embeddings job %s", job.ID)
}

// getNodeClient gets or creates a gRPC client for a node
func (p *JobProcessor) getNodeClient(nodeID string, node *pb.Node) (pb.NodeAgentClient, error) {
	p.mu.RLock()
	if client, exists := p.nodeClients[nodeID]; exists {
		p.mu.RUnlock()
		return client, nil
	}
	p.mu.RUnlock()

	p.mu.Lock()
	defer p.mu.Unlock()

	// Double-check after acquiring write lock
	if client, exists := p.nodeClients[nodeID]; exists {
		return client, nil
	}

	// Determine node agent address
	addr := node.AgentAddress
	if addr == "" {
		// Default to hostname:50052 if not specified
		addr = fmt.Sprintf("%s:50052", node.Hostname)
	}

	// Connect to node agent
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to node %s at %s: %w", nodeID, addr, err)
	}

	client := pb.NewNodeAgentClient(conn)
	p.nodeClients[nodeID] = client

	log.Printf("Connected to node agent %s at %s", nodeID, addr)
	return client, nil
}
