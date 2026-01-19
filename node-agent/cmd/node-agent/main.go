package main

import (
	"context"
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/google/uuid"

	pb "github.com/Orchion/Orchion/node-agent/internal/proto/v1"
	"github.com/Orchion/Orchion/node-agent/internal/capabilities"
	"github.com/Orchion/Orchion/node-agent/internal/heartbeat"
)

var (
	orchestratorAddr = flag.String("orchestrator", "localhost:50051", "Orchestrator gRPC address")
	heartbeatInterval = flag.Duration("heartbeat-interval", 5*time.Second, "Heartbeat interval")
	nodeID            = flag.String("node-id", "", "Node ID (auto-generated if empty)")
	nodeHostname      = flag.String("hostname", "", "Node hostname (uses system hostname if empty)")
)

func main() {
	flag.Parse()

	log.Println("Orchion Node Agent starting...")

	// Generate or use provided node ID
	if *nodeID == "" {
		*nodeID = uuid.New().String()
		log.Printf("Generated node ID: %s", *nodeID)
	} else {
		log.Printf("Using node ID: %s", *nodeID)
	}

	// Get hostname
	hostname := *nodeHostname
	if hostname == "" {
		var err error
		hostname, err = os.Hostname()
		if err != nil {
			log.Printf("Failed to get hostname: %v, using 'unknown'", err)
			hostname = "unknown"
		}
	}

	log.Printf("Hostname: %s", hostname)

	// Detect capabilities
	caps := capabilities.Detect()
	log.Printf("Capabilities: CPU=%s, Memory=%s, OS=%s", caps.Cpu, caps.Memory, caps.Os)

	// Create heartbeat client
	client, err := heartbeat.NewClient(*orchestratorAddr)
	if err != nil {
		log.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	log.Printf("Connected to orchestrator at %s", *orchestratorAddr)

	// Create node info
	node := &pb.Node{
		Id:       *nodeID,
		Hostname: hostname,
		Capabilities: caps,
		LastSeenUnix: time.Now().Unix(),
	}

	// Register with orchestrator
	ctx := context.Background()
	if err := client.RegisterNode(ctx, node); err != nil {
		log.Fatalf("Failed to register node: %v", err)
	}
	log.Println("Node registered successfully")

	// Setup graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// Start heartbeat loop
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	client.StartHeartbeatLoop(ctx, *heartbeatInterval)
	log.Printf("Heartbeat loop started (interval: %v)", *heartbeatInterval)

	log.Println("Node agent running. Press Ctrl+C to stop.")

	// Wait for shutdown signal
	sig := <-sigChan
	log.Printf("Received signal: %v, shutting down...", sig)
}