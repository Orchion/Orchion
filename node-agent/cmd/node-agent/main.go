package main

import (
	"context"
	"flag"
	"fmt"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	"github.com/google/uuid"

	"github.com/Orchion/Orchion/node-agent/internal/capabilities"
	"github.com/Orchion/Orchion/node-agent/internal/executor"
	"github.com/Orchion/Orchion/node-agent/internal/heartbeat"
	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
	"github.com/Orchion/Orchion/shared/logging"
)

var (
	orchestratorAddr   = flag.String("orchestrator", "localhost:50051", "Orchestrator gRPC address")
	heartbeatInterval  = flag.Duration("heartbeat-interval", 5*time.Second, "Heartbeat interval")
	capabilityInterval = flag.Duration("capability-interval", 10*time.Second, "Capability update interval")
	nodeID             = flag.String("node-id", "", "Node ID (auto-generated if empty)")
	nodeHostname       = flag.String("hostname", "", "Node hostname (uses system hostname if empty)")
	agentPort          = flag.String("agent-port", "50052", "Node agent gRPC server port")
)

// startCapabilityUpdateLoop periodically updates node capabilities
func startCapabilityUpdateLoop(ctx context.Context, client *heartbeat.Client, interval time.Duration, logger logging.Logger) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			if err := client.UpdateCapabilities(ctx); err != nil {
				logger.Error("Capability update error", map[string]interface{}{
					"error": err.Error(),
				})
			}
		}
	}
}

func main() {
	flag.Parse()

	// Generate or use provided node ID
	if *nodeID == "" {
		*nodeID = uuid.New().String()
	} else {
		// Validate provided node ID
	}

	// Initialize structured logger (will setup streaming later after we know orchestrator address)
	logger := logging.NewLogger(logging.Config{
		Level:  logging.InfoLevel,
		Source: fmt.Sprintf("node-agent:%s", *nodeID),
	})

	logger.Info("Orchion Node Agent starting", map[string]interface{}{
		"node_id": *nodeID,
	})

	// Get hostname
	hostname := *nodeHostname
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			logger.Warn("Failed to get hostname, using 'unknown'", map[string]interface{}{
				"error": err.Error(),
			})
			hostname = "unknown"
		}
	}

	logger.Info("Node information", map[string]interface{}{
		"hostname": hostname,
	})

	// Detect capabilities
	caps := capabilities.Detect()
	logger.Info("Detected capabilities", map[string]interface{}{
		"cpu":    caps.Cpu,
		"memory": caps.Memory,
		"os":     caps.Os,
		"gpu":    caps.GpuType,
	})

	// Create heartbeat client
	client, err := heartbeat.NewClient(*orchestratorAddr)
	if err != nil {
		logger.Error("Failed to create heartbeat client", map[string]interface{}{
			"orchestrator_addr": *orchestratorAddr,
			"error":             err.Error(),
		})
		os.Exit(1)
	}
	defer client.Close()

	logger.Info("Connected to orchestrator", map[string]interface{}{
		"orchestrator_addr": *orchestratorAddr,
	})

	// TODO: Setup log streaming to orchestrator
	// For now, logs are only local. Streaming implementation pending.

	// Create node info
	node := &pb.Node{
		Id:           *nodeID,
		Hostname:     hostname,
		Capabilities: caps,
		LastSeenUnix: time.Now().Unix(),
		AgentAddress: fmt.Sprintf("%s:%s", hostname, *agentPort),
	}

	// Register with orchestrator
	ctx := context.Background()
	if err := client.RegisterNode(ctx, node); err != nil {
		logger.Error("Failed to register node", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	logger.Info("Node registered successfully", nil)

	// Enable periodic capability updates
	client.EnableCapabilityUpdates(capabilities.Detect)
	logger.Info("Capability updates enabled", map[string]interface{}{
		"interval": *capabilityInterval,
	})

	// Create executor service
	executorService, err := executor.NewService()
	if err != nil {
		logger.Error("Failed to create executor service", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
	logger.Info("Created executor service", map[string]interface{}{
		"features": "container management",
	})

	// Setup gRPC server for NodeAgent service
	grpcLis, err := net.Listen("tcp", ":"+*agentPort)
	if err != nil {
		logger.Error("Failed to listen on agent port", map[string]interface{}{
			"port":  *agentPort,
			"error": err.Error(),
		})
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNodeAgentServer(grpcServer, executorService)
	logger.Info("Node agent gRPC server listening", map[string]interface{}{
		"port": *agentPort,
	})

	// Start gRPC server
	go func() {
		if err := grpcServer.Serve(grpcLis); err != nil {
			logger.Error("Failed to serve gRPC", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start heartbeat loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client.StartHeartbeatLoop(ctx, *heartbeatInterval)
	logger.Info("Heartbeat loop started", map[string]interface{}{
		"interval": *heartbeatInterval,
	})

	// Start capability update loop
	go startCapabilityUpdateLoop(ctx, client, *capabilityInterval, logger)
	logger.Info("Capability update loop started", map[string]interface{}{
		"interval": *capabilityInterval,
	})

	logger.Info("Node agent running, waiting for shutdown signal", nil)

	// Wait for shutdown signal
	sig := <-sigChan
	logger.Info("Received shutdown signal", map[string]interface{}{
		"signal": sig.String(),
	})

	// Graceful shutdown
	grpcServer.GracefulStop()

	// Shutdown executor service (stops containers)
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer shutdownCancel()

	if err := executorService.Shutdown(shutdownCtx); err != nil {
		logger.Error("Error during executor shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}
}
