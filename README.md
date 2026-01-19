# Orchion

Run AI agents across every machine you own

## Overview

Orchion is a distributed orchestrator service for managing and coordinating AI agent workloads across multiple machines. It provides a gRPC-based API for node registration, heartbeat monitoring, and task scheduling.

**Current Status:** ✅ Minimal working system - ready for testing! See `docs/quick-start.md` for how to run it.

## Project Structure

```
Orchion/
├── orchestrator/           # Orchestrator service (Go)
│   ├── cmd/orchestrator/   # Main entry point
│   ├── internal/
│   │   ├── node/          # Node registry
│   │   └── orchestrator/  # gRPC service
│   ├── api/v1/v1/        # Generated protobuf files
│   └── Makefile          # Protobuf generation
│
├── node-agent/            # Node agent daemon (Go)
│   ├── cmd/node-agent/    # Main entry point
│   ├── internal/
│   │   ├── capabilities/  # Hardware detection
│   │   ├── heartbeat/     # Orchestrator client
│   │   └── proto/v1/      # Generated protobuf files
│   └── Makefile          # Protobuf generation
│
├── dashboard/             # Web UI (SvelteKit)
│   └── src/              # SvelteKit application
│
├── shared/               # Shared resources
│   └── proto/v1/        # Protocol definitions
│
└── docs/                # Documentation
```

## Features

### ✅ Implemented
- **gRPC API**: Node registration, heartbeat, and listing with proper status codes
- **HTTP REST API**: Dashboard-compatible endpoints (`/api/nodes`) with CORS support
- **Node Registry**: In-memory registry for tracking registered nodes with thread-safe operations
- **Heartbeat Monitoring**: Automatic detection and removal of stale/inactive nodes
- **Capability Detection**: Accurate CPU, memory (system RAM), and OS detection using gopsutil
- **Auto Re-registration**: Node agents automatically re-register if orchestrator restarts
- **Graceful Shutdown**: Proper signal handling for clean shutdowns
- **Dashboard**: Node list UI with error handling and configurable orchestrator URL
- **VS Code Extension**: Basic structure with tree view provider
- **Container Management**: Docker container management infrastructure (Ollama/vLLM support)
- **Build Scripts**: PowerShell scripts for building, running, and testing all components

### ⏳ Planned
- Job scheduling and execution
- Persistent storage
- Authentication/authorization
- Advanced dashboard features (auto-refresh, job queue, logs)
- Docker/Kubernetes deployment configs

## Quick Start

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
- Make (optional, for protobuf generation)
- Node.js 18+ (for dashboard)
- PowerShell (for build/run scripts)
- **Container Runtime**: [Podman](https://podman.io/) (preferred) or Docker

See `docs/development-setup.md` for detailed setup instructions.

### Quick Setup (Recommended)

**1. Setup everything (first time only):**
```powershell
.\shared\scripts\setup-all.ps1
```

**2. Build all components:**
```powershell
.\shared\scripts\build-all.ps1
```

**3. Run everything:**
```powershell
.\shared\scripts\run-all.ps1
```

This will:
- Start orchestrator in a new window (gRPC: 50051, HTTP: 8080)
- Start node-agent in a new window
- Start dashboard in the current window
- Automatically clean up all processes when you press Ctrl+C

### Manual Running

**1. Start the Orchestrator:**
```powershell
cd orchestrator
go run ./cmd/orchestrator
```

**2. Start a Node Agent** (in another terminal):
```powershell
cd node-agent
go run ./cmd/node-agent
```

**3. View Nodes:**

- **HTTP API:** `http://localhost:8080/api/nodes`
- **Dashboard:** `.\shared\scripts\dev-dashboard.ps1` or `cd dashboard && npm run dev`

For detailed testing instructions, see `docs/quick-start.md`.

## Building

### Using Scripts (Recommended)

**Build everything:**
```powershell
.\shared\scripts\build-all.ps1
```

**Generate protobuf files:**
```powershell
.\shared\scripts\proto-gen.ps1
```

**Clean build artifacts:**
```powershell
.\shared\scripts\clean-all.ps1
```

See `shared/scripts/README.md` for all available scripts.

### Manual Building

**Generate Protobuf Code:**
```powershell
# Orchestrator
cd orchestrator
make proto

# Node Agent
cd node-agent
make proto
```

**Build Binaries:**
```powershell
# Orchestrator
cd orchestrator
go build -o orchestrator.exe ./cmd/orchestrator

# Node Agent
cd node-agent
go build -o node-agent.exe ./cmd/node-agent
```

## Configuration

### Orchestrator Command-Line Options

- `--port`: gRPC server port (default: 50051)
- `--http-port`: HTTP REST API port (default: 8080)
- `--heartbeat-timeout`: Node heartbeat timeout duration (default: 30s)

Example:
```bash
./bin/orchestrator --port 50051 --http-port 8080 --heartbeat-timeout 1m
```

### Node Agent Command-Line Options

- `--orchestrator`: Orchestrator gRPC address (default: localhost:50051)
- `--heartbeat-interval`: Heartbeat interval (default: 5s)
- `--node-id`: Custom node ID (auto-generated if not provided)
- `--hostname`: Custom hostname (uses system hostname if not provided)

Example:
```bash
./bin/node-agent --orchestrator localhost:50051 --heartbeat-interval 10s
```

## API

### gRPC Endpoints

The orchestrator provides the following gRPC endpoints:

- `RegisterNode`: Register a new node with the orchestrator
- `Heartbeat`: Send heartbeat to indicate node is alive
- `ListNodes`: List all registered nodes

### HTTP REST API

- `GET /api/nodes`: List all registered nodes (JSON)

See `shared/proto/v1/orchestrator.proto` for protocol definitions.

## Development

### Code Generation

After modifying `.proto` files, regenerate the code:

```bash
cd orchestrator && make proto
cd ../node-agent && make proto
```

### Testing

```bash
# Test orchestrator
cd orchestrator
go test ./...

# Test node agent
cd node-agent
go test ./...
```

### Code Formatting

```bash
go fmt ./...
```

## Documentation

- **`docs/quick-start.md`** - How to test the system
- **`docs/development-setup.md`** - Development environment setup and quick commands
- **`docs/architecture.md`** - System architecture and components
- **`docs/project-status.md`** - Current project status, gaps, and next steps
- **`docs/roadmap.md`** - Planned features and roadmap
- **`shared/scripts/README.md`** - Available PowerShell scripts for building/running/testing
- **`dashboard/README.md`** - Dashboard development guide
- **`node-agent/README.md`** - Node agent details and container integration

## Architecture

### Node Registry

The node registry is an in-memory data structure that tracks all registered nodes. Each node has:
- Unique ID
- Hostname
- Capabilities (CPU, memory, OS)
- Last heartbeat timestamp

### Heartbeat Monitoring

A background goroutine periodically checks node heartbeats and automatically removes stale nodes that haven't sent a heartbeat within the configured timeout period. This ensures the registry only contains active nodes.

### Component Communication

- **Node Agent ↔ Orchestrator**: gRPC (port 50051)
- **Dashboard ↔ Orchestrator**: HTTP REST (port 8080)

## License

See LICENSE file for details.
