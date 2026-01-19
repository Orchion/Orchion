package main

import (
	"context"
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/gateway"
	"github.com/Orchion/Orchion/orchestrator/internal/llm"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	"github.com/Orchion/Orchion/orchestrator/internal/orchestrator"
	"github.com/Orchion/Orchion/orchestrator/internal/scheduler"
)

var (
	port             = flag.String("port", "50051", "gRPC server port")
	httpPort         = flag.String("http-port", "8080", "HTTP REST API port")
	heartbeatTimeout = flag.Duration("heartbeat-timeout", 30*time.Second, "Node heartbeat timeout duration")
	apiKey           = flag.String("api-key", "", "Optional API key for authentication (leave empty to disable)")
)

func main() {
	flag.Parse()

	log.Printf("Starting Orchion Orchestrator")
	log.Printf("gRPC port: %s", *port)
	log.Printf("HTTP port: %s", *httpPort)
	log.Printf("Heartbeat timeout: %v", *heartbeatTimeout)

	// Create node registry
	registry := node.NewInMemoryRegistry()

	// Create orchestrator service
	service := orchestrator.NewService(registry)

	// Create scheduler
	sched := scheduler.NewSimpleScheduler()

	// Create LLM service
	llmService := llm.NewService(registry, sched)

	// Setup gRPC server
	grpcLis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		log.Fatalf("Failed to listen on gRPC port %s: %v", *port, err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrchestratorServer(grpcServer, service)
	pb.RegisterOrchionLLMServer(grpcServer, llmService)

	// Setup HTTP REST API server
	mux := http.NewServeMux()
	
	// Dashboard API
	mux.HandleFunc("/api/nodes", func(w http.ResponseWriter, r *http.Request) {
		// Add CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		ctx := context.Background()
		resp, err := service.ListNodes(ctx, &pb.ListNodesRequest{})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp.Nodes)
	})

	// OpenAI-compatible API Gateway
	gateway := gateway.NewGateway("localhost:" + *port)
	if *apiKey != "" {
		gateway.SetAPIKey(*apiKey)
		log.Printf("API key authentication enabled")
	}
	mux.HandleFunc("/v1/chat/completions", gateway.ChatCompletionsHandler)
	mux.HandleFunc("/v1/embeddings", gateway.EmbeddingsHandler)

	httpServer := &http.Server{
		Addr:    ":" + *httpPort,
		Handler: mux,
	}

	// Start heartbeat monitor goroutine
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go monitorHeartbeats(ctx, registry, *heartbeatTimeout)

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %v, shutting down gracefully...", sig)

		// Shutdown HTTP server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		httpServer.Shutdown(shutdownCtx)

		// Shutdown gRPC server
		grpcServer.GracefulStop()
	}()

	// Start HTTP server
	go func() {
		log.Printf("HTTP REST API listening on :%s", *httpPort)
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to serve HTTP: %v", err)
		}
	}()

	// Start gRPC server (blocking)
	log.Printf("gRPC server listening on :%s", *port)
	if err := grpcServer.Serve(grpcLis); err != nil {
		log.Fatalf("Failed to serve gRPC: %v", err)
	}
}

// monitorHeartbeats periodically checks for stale nodes and removes them
func monitorHeartbeats(ctx context.Context, registry node.Registry, timeout time.Duration) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			staleNodes := registry.CheckHeartbeats(timeout)
			if len(staleNodes) > 0 {
				log.Printf("Found %d stale node(s) (timeout: %v), removing...", len(staleNodes), timeout)
				for _, nodeID := range staleNodes {
					if err := registry.Remove(nodeID); err != nil {
						log.Printf("  - Failed to remove stale node %s: %v", nodeID, err)
					} else {
						log.Printf("  - Removed stale node %s", nodeID)
					}
				}
			}
		}
	}
}
