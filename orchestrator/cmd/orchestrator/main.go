package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"google.golang.org/grpc"

	pb "github.com/Orchion/Orchion/orchestrator/api/v1"
	"github.com/Orchion/Orchion/orchestrator/internal/gateway"
	"github.com/Orchion/Orchion/orchestrator/internal/llm"
	logServicePkg "github.com/Orchion/Orchion/orchestrator/internal/logging"
	"github.com/Orchion/Orchion/orchestrator/internal/node"
	"github.com/Orchion/Orchion/orchestrator/internal/orchestrator"
	"github.com/Orchion/Orchion/orchestrator/internal/queue"
	"github.com/Orchion/Orchion/orchestrator/internal/scheduler"
	"github.com/Orchion/Orchion/shared/logging"
)

var (
	port             = flag.String("port", "50051", "gRPC server port")
	httpPort         = flag.String("http-port", "8080", "HTTP REST API port")
	heartbeatTimeout = flag.Duration("heartbeat-timeout", 30*time.Second, "Node heartbeat timeout duration")
	apiKey           = flag.String("api-key", "", "Optional API key for authentication (leave empty to disable)")
)

func main() {
	flag.Parse()

	// Initialize structured logger
	logger := logging.NewLogger(logging.Config{
		Level:  logging.InfoLevel,
		Source: "orchestrator",
	})

	logger.Info("Starting Orchion Orchestrator", map[string]interface{}{
		"grpc_port":         *port,
		"http_port":         *httpPort,
		"heartbeat_timeout": *heartbeatTimeout,
	})

	// Create node registry
	registry := node.NewInMemoryRegistry()

	// Create job queue
	jobQueue := queue.NewJobQueue()

	// Create scheduler
	sched := scheduler.NewSimpleScheduler()

	// Create orchestrator service
	service := orchestrator.NewService(registry, jobQueue, sched)

	// Create logging service
	logService := logServicePkg.NewService()

	// Create LLM service
	llmService := llm.NewService(registry, sched)

	// Setup logger with streaming
	streamer := logServicePkg.NewOrchestratorStreamer(logService)
	logger.SetStreamer(streamer)
	defer logger.Close()

	// Setup gRPC server
	grpcLis, err := net.Listen("tcp", ":"+*port)
	if err != nil {
		logger.Error("Failed to listen on gRPC port", map[string]interface{}{
			"port":  *port,
			"error": err.Error(),
		})
		os.Exit(1)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterOrchestratorServer(grpcServer, service)
	pb.RegisterOrchionLLMServer(grpcServer, llmService)
	pb.RegisterLogStreamerServer(grpcServer, logService)

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

	// Logs streaming endpoint (Server-Sent Events)
	mux.HandleFunc("/api/logs", func(w http.ResponseWriter, r *http.Request) {
		// Set SSE headers
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Cache-Control")

		// Handle preflight requests
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusOK)
			return
		}

		if r.Method != http.MethodGet {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		// Register with the logging service to receive broadcasts
		// For now, we'll create a simple broadcaster that sends keep-alive messages
		// In a real implementation, we'd want persistent storage of logs

		// Send a keep-alive message initially
		fmt.Fprintf(w, "data: {\"type\": \"connected\"}\n\n")
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}

		// Simple keep-alive loop (in production, would stream actual logs)
		ticker := time.NewTicker(30 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-r.Context().Done():
				return
			case <-ticker.C:
				fmt.Fprintf(w, "data: {\"type\": \"keepalive\", \"timestamp\": %d}\n\n", time.Now().Unix())
				if f, ok := w.(http.Flusher); ok {
					f.Flush()
				}
			}
		}
	})

	// OpenAI-compatible API Gateway
	gateway := gateway.NewGateway("localhost:" + *port)
	if *apiKey != "" {
		gateway.SetAPIKey(*apiKey)
		logger.Info("API key authentication enabled", nil)
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
	go monitorHeartbeats(ctx, registry, *heartbeatTimeout, logger)

	// Start job processor
	processor := orchestrator.NewJobProcessor(jobQueue, sched, registry)
	processor.Start(ctx)

	// Graceful shutdown handling
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Received shutdown signal, shutting down gracefully", map[string]interface{}{
			"signal": sig.String(),
		})

		// Shutdown HTTP server
		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()
		httpServer.Shutdown(shutdownCtx)

		// Shutdown gRPC server
		grpcServer.GracefulStop()
	}()

	// Start HTTP server
	go func() {
		logger.Info("HTTP REST API listening", map[string]interface{}{
			"port": *httpPort,
		})
		if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Error("Failed to serve HTTP", map[string]interface{}{
				"error": err.Error(),
			})
			os.Exit(1)
		}
	}()

	// Start gRPC server (blocking)
	logger.Info("gRPC server listening", map[string]interface{}{
		"port": *port,
	})
	if err := grpcServer.Serve(grpcLis); err != nil {
		logger.Error("Failed to serve gRPC", map[string]interface{}{
			"error": err.Error(),
		})
		os.Exit(1)
	}
}

// monitorHeartbeats periodically checks for stale nodes and removes them
func monitorHeartbeats(ctx context.Context, registry node.Registry, timeout time.Duration, logger logging.Logger) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			staleNodes := registry.CheckHeartbeats(timeout)
			if len(staleNodes) > 0 {
				logger.Warn("Found stale nodes, removing", map[string]interface{}{
					"count":   len(staleNodes),
					"timeout": timeout,
				})
				for _, nodeID := range staleNodes {
					if err := registry.Remove(nodeID); err != nil {
						logger.Error("Failed to remove stale node", map[string]interface{}{
							"node_id": nodeID,
							"error":   err.Error(),
						})
					} else {
						logger.Info("Removed stale node", map[string]interface{}{
							"node_id": nodeID,
						})
					}
				}
			}
		}
	}
}
