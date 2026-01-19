package heartbeat

import (
	"context"
	"fmt"
	"log"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
)

// Client handles communication with the orchestrator
type Client struct {
	conn        *grpc.ClientConn
	client      pb.OrchestratorClient
	address     string
	nodeID      string
	nodeInfo    *pb.Node                // Store node info for re-registration
	updateCaps  bool                    // Whether to update capabilities periodically
	capsUpdater func() *pb.Capabilities // Function to get updated capabilities
}

// NewClient creates a new heartbeat client
func NewClient(orchestratorAddress string) (*Client, error) {
	conn, err := grpc.NewClient(orchestratorAddress, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("failed to connect to orchestrator: %w", err)
	}

	return &Client{
		conn:    conn,
		client:  pb.NewOrchestratorClient(conn),
		address: orchestratorAddress,
	}, nil
}

// RegisterNode registers a node with the orchestrator
func (c *Client) RegisterNode(ctx context.Context, node *pb.Node) error {
	req := &pb.RegisterNodeRequest{Node: node}
	_, err := c.client.RegisterNode(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to register node: %w", err)
	}
	c.nodeID = node.Id
	// Store node info for potential re-registration
	c.nodeInfo = &pb.Node{
		Id:           node.Id,
		Hostname:     node.Hostname,
		Capabilities: node.Capabilities,
		LastSeenUnix: node.LastSeenUnix,
	}
	return nil
}

// EnableCapabilityUpdates enables periodic capability updates
func (c *Client) EnableCapabilityUpdates(updater func() *pb.Capabilities) {
	c.updateCaps = true
	c.capsUpdater = updater
}

// SendHeartbeat sends a heartbeat to the orchestrator
func (c *Client) SendHeartbeat(ctx context.Context) error {
	if c.nodeID == "" {
		return fmt.Errorf("node not registered, cannot send heartbeat")
	}

	req := &pb.HeartbeatRequest{NodeId: c.nodeID}
	_, err := c.client.Heartbeat(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to send heartbeat: %w", err)
	}

	return nil
}

// UpdateCapabilities sends updated capabilities to the orchestrator
func (c *Client) UpdateCapabilities(ctx context.Context) error {
	if c.nodeID == "" {
		return fmt.Errorf("node not registered, cannot update capabilities")
	}

	if c.capsUpdater == nil {
		return fmt.Errorf("capability updater not configured")
	}

	caps := c.capsUpdater()
	req := &pb.UpdateNodeRequest{
		NodeId:       c.nodeID,
		Capabilities: caps,
	}

	_, err := c.client.UpdateNode(ctx, req)
	if err != nil {
		return fmt.Errorf("failed to update capabilities: %w", err)
	}

	return nil
}

// StartHeartbeatLoop starts a goroutine that sends heartbeats periodically
func (c *Client) StartHeartbeatLoop(ctx context.Context, interval time.Duration) {
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if err := c.SendHeartbeat(ctx); err != nil {
					// Check if error is "node not found" - if so, re-register
					if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
						log.Printf("Node not found in registry, attempting re-registration...")
						if c.nodeInfo != nil {
							// Update timestamp before re-registering
							c.nodeInfo.LastSeenUnix = time.Now().Unix()
							if regErr := c.RegisterNode(ctx, c.nodeInfo); regErr != nil {
								log.Printf("Failed to re-register node: %v", regErr)
							} else {
								log.Printf("Successfully re-registered node %s", c.nodeID)
							}
						} else {
							log.Printf("Cannot re-register: node info not available")
						}
					} else {
						log.Printf("Heartbeat error: %v", err)
					}
				}
			}
		}
	}()
}

// Close closes the connection
func (c *Client) Close() error {
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
