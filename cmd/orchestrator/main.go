package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	orchionv1 "github.com/Orchion/Orchion/api/proto/v1"
	"github.com/Orchion/Orchion/internal/node"
	"github.com/Orchion/Orchion/internal/orchestrator"
	"google.golang.org/grpc"
)

var (
	port             = flag.Int("port", 50051, "The gRPC server port")
	heartbeatTimeout = flag.Duration("heartbeat-timeout", 30*time.Second, "Node heartbeat timeout")
)

func main() {
	flag.Parse()

	log.Printf("Starting Orchion Orchestrator on port %d", *port)

	// Create the in-memory node registry
	registry := node.NewInMemoryRegistry()

	// Create the orchestrator service
	svc := orchestrator.NewService(registry, *heartbeatTimeout)

	// Start heartbeat monitor in background
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go svc.StartHeartbeatMonitor(ctx)

	// Create gRPC server
	grpcServer := grpc.NewServer()
	orchionv1.RegisterOrchestratorServiceServer(grpcServer, svc)

	// Setup listener
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", *port))
	if err != nil {
		log.Fatalf("Failed to listen on port %d: %v", *port, err)
	}

	// Handle graceful shutdown
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigCh
		log.Printf("Received signal %v, shutting down gracefully...", sig)
		grpcServer.GracefulStop()
		cancel()
	}()

	// Start serving
	log.Printf("Orchestrator listening on %s", lis.Addr().String())
	if err := grpcServer.Serve(lis); err != nil {
		log.Fatalf("Failed to serve: %v", err)
	}

	log.Println("Orchestrator stopped")
}
