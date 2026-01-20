package integration

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	pb "github.com/Orchion/Orchion/shared/proto/v1"
)

func TestAPI_Contract_Validation(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Test RegisterNode contract
	t.Run("RegisterNode contract", func(t *testing.T) {
		node := &pb.Node{
			Id:       "contract-test-node",
			Hostname: "contract-host",
			Capabilities: &pb.Capabilities{
				Cpu:    "4 cores",
				Memory: "8GB",
				Os:     "linux",
			},
		}

		resp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	// Test Heartbeat contract
	t.Run("Heartbeat contract", func(t *testing.T) {
		resp, err := ts.service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "contract-test-node"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)
	})

	// Test ListNodes contract
	t.Run("ListNodes contract", func(t *testing.T) {
		resp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.NotEmpty(t, resp.Nodes)

		// Verify node structure
		found := false
		for _, node := range resp.Nodes {
			if node.Id == "contract-test-node" {
				found = true
				assert.Equal(t, "contract-host", node.Hostname)
				assert.NotNil(t, node.Capabilities)
				assert.Equal(t, "4 cores", node.Capabilities.Cpu)
				assert.Equal(t, "8GB", node.Capabilities.Memory)
				assert.Equal(t, "linux", node.Capabilities.Os)
				break
			}
		}
		assert.True(t, found, "Registered node should be in list")
	})

	// Test UpdateNode contract
	t.Run("UpdateNode contract", func(t *testing.T) {
		resp, err := ts.service.UpdateNode(ctx, &pb.UpdateNodeRequest{
			NodeId: "contract-test-node",
			Node: &pb.Node{
				Id:       "contract-test-node",
				Hostname: "updated-contract-host",
				Capabilities: &pb.Capabilities{
					Cpu:    "8 cores",
					Memory: "16GB",
					Os:     "linux",
					GpuType: "NVIDIA RTX 3080",
				},
			},
		})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.True(t, resp.Success)

		// Verify update
		listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
		require.NoError(t, err)
		for _, node := range listResp.Nodes {
			if node.Id == "contract-test-node" {
				assert.Equal(t, "updated-contract-host", node.Hostname)
				assert.Equal(t, "8 cores", node.Capabilities.Cpu)
				assert.Equal(t, "16GB", node.Capabilities.Memory)
				assert.Equal(t, "NVIDIA RTX 3080", node.Capabilities.GpuType)
				break
			}
		}
	})

	// Test SubmitJob contract
	t.Run("SubmitJob contract", func(t *testing.T) {
		jobReq := &pb.SubmitJobRequest{
			JobId:   "contract-test-job",
			JobType: pb.JobType_JOB_TYPE_CHAT_COMPLETION,
			Payload: []byte(`{"model":"test","messages":[{"role":"user","content":"test"}]}`),
		}

		resp, err := ts.service.SubmitJob(ctx, jobReq)
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "contract-test-job", resp.JobId)
		assert.True(t, resp.Status == pb.JobStatus_JOB_STATUS_PENDING ||
			resp.Status == pb.JobStatus_JOB_STATUS_ASSIGNED)
	})

	// Test GetJobStatus contract
	t.Run("GetJobStatus contract", func(t *testing.T) {
		resp, err := ts.service.GetJobStatus(ctx, &pb.GetJobStatusRequest{JobId: "contract-test-job"})
		require.NoError(t, err)
		require.NotNil(t, resp)
		assert.Equal(t, "contract-test-job", resp.JobId)
	})
}

func TestAPI_Error_Contracts(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Test error responses for invalid inputs
	t.Run("Heartbeat non-existent node", func(t *testing.T) {
		resp, err := ts.service.Heartbeat(ctx, &pb.HeartbeatRequest{NodeId: "non-existent"})
		require.NoError(t, err) // gRPC doesn't return error for application-level failures
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "not found")
	})

	t.Run("UpdateNode non-existent node", func(t *testing.T) {
		resp, err := ts.service.UpdateNode(ctx, &pb.UpdateNodeRequest{
			NodeId: "non-existent",
			Node: &pb.Node{
				Id:       "non-existent",
				Hostname: "test",
			},
		})
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.Contains(t, resp.Message, "not found")
	})

	t.Run("RegisterNode with empty ID", func(t *testing.T) {
		resp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{
			Node: &pb.Node{
				Id:       "",
				Hostname: "test",
			},
		})
		require.NoError(t, err)
		assert.False(t, resp.Success)
		assert.NotEmpty(t, resp.Message)
	})

	t.Run("GetJobStatus non-existent job", func(t *testing.T) {
		resp, err := ts.service.GetJobStatus(ctx, &pb.GetJobStatusRequest{JobId: "non-existent-job"})
		require.NoError(t, err)
		assert.Equal(t, "non-existent-job", resp.JobId)
		assert.Equal(t, pb.JobStatus_JOB_STATUS_FAILED, resp.Status)
		assert.Contains(t, resp.ErrorMessage, "not found")
	})
}

func TestAPI_Backward_Compatibility(t *testing.T) {
	ts := setupTestServer(t)
	defer ts.cleanup()

	ctx := context.Background()

	// Test that older API versions still work (basic compatibility)
	node := &pb.Node{
		Id:       "compat-test-node",
		Hostname: "compat-host",
		Capabilities: &pb.Capabilities{
			Cpu:    "2 cores",
			Memory: "4GB",
			Os:     "linux",
		},
	}

	// Register with basic capabilities
	regResp, err := ts.service.RegisterNode(ctx, &pb.RegisterNodeRequest{Node: node})
	require.NoError(t, err)
	assert.True(t, regResp.Success)

	// Update with additional capabilities (should not break)
	updateResp, err := ts.service.UpdateNode(ctx, &pb.UpdateNodeRequest{
		NodeId: "compat-test-node",
		Node: &pb.Node{
			Id:       "compat-test-node",
			Hostname: "compat-host",
			Capabilities: &pb.Capabilities{
				Cpu:     "4 cores", // Updated
				Memory:  "8GB",     // Updated
				Os:      "linux",
				GpuType: "NVIDIA GTX 1660", // Added
			},
		},
	})
	require.NoError(t, err)
	assert.True(t, updateResp.Success)

	// Verify updates were applied
	listResp, err := ts.service.ListNodes(ctx, &pb.ListNodesRequest{})
	require.NoError(t, err)
	for _, n := range listResp.Nodes {
		if n.Id == "compat-test-node" {
			assert.Equal(t, "4 cores", n.Capabilities.Cpu)
			assert.Equal(t, "8GB", n.Capabilities.Memory)
			assert.Equal(t, "NVIDIA GTX 1660", n.Capabilities.GpuType)
			break
		}
	}
}