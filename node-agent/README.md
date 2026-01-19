# Node Agent

A lightweight daemon that runs on each machine in the Orchion cluster. Registers with the orchestrator, reports capabilities, and sends periodic heartbeats.

---

## Overview

The node agent is responsible for:
- Registering with the orchestrator on startup
- Detecting and reporting hardware capabilities (CPU, memory, OS)
- Sending periodic heartbeats to indicate the node is alive
- (Future) Executing jobs assigned by the orchestrator

**Current Status:** ✅ Registration and heartbeat working. Job execution not yet implemented.

**Design Philosophy:** Minimal and lightweight - designed to run on everything from Raspberry Pis to powerful servers.

---

## Architecture

```
node-agent/
├── cmd/node-agent/             # Main entry point
│   └── main.go                 # Agent lifecycle management
├── internal/
│   ├── capabilities/           # Hardware capability detection
│   │   └── capabilities.go     # CPU, memory, OS detection
│   ├── heartbeat/              # Orchestrator communication
│   │   └── heartbeat.go        # gRPC client for orchestrator
│   ├── containers/             # Container management
│   │   ├── manager.go          # Docker lifecycle management
│   │   ├── vllm.go             # vLLM container config
│   │   └── ollama.go           # Ollama container config
│   ├── executor/               # Job execution (planned)
│   │   └── executor.go         # Empty placeholder
│   └── proto/v1/               # Generated protobuf files
├── go.mod                      # Go module definition
└── Makefile                    # Protobuf generation
```

---

## Building

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
- Go protobuf plugins (`protoc-gen-go`, `protoc-gen-go-grpc`)
- Access to orchestrator (for runtime)

### Build

**Using the monorepo script:**
```powershell
# From project root
.\shared\scripts\build-all.ps1
```

**Or manually:**
```powershell
go build -o node-agent.exe ./cmd/node-agent
```

**Generate protobuf files:**
```powershell
make proto
# Or
.\shared\scripts\proto-gen.ps1
```

---

## Running

### Basic Usage

**Prerequisites:** Orchestrator must be running first!

```powershell
.\node-agent.exe
```

### Command-Line Options

```
-orchestrator      Orchestrator gRPC address (default: localhost:50051)
-heartbeat-interval Heartbeat interval (default: 5s)
-node-id          Custom node ID (auto-generated if not provided)
-hostname         Custom hostname (uses system hostname if not provided)
```

### Examples

```powershell
# Connect to remote orchestrator
.\node-agent.exe -orchestrator orchestrator.example.com:50051

# Custom node ID and slower heartbeat
.\node-agent.exe -node-id my-server-01 -heartbeat-interval 10s

# Custom hostname
.\node-agent.exe -hostname production-db-server
```

---

## Components

### Capability Detection

`internal/capabilities/capabilities.go` detects:
- **CPU:** Number of cores (`runtime.NumCPU()`)
- **Memory:** Total system memory (`runtime.MemStats`)
- **OS:** Operating system and architecture (`runtime.GOOS`, `runtime.GOARCH`)

**Note:** Memory detection shows allocated Go memory, not total system memory. This is a limitation and will be improved.

### Heartbeat Client

`internal/heartbeat/heartbeat.go` provides:
- gRPC client connection to orchestrator
- Node registration on startup
- Periodic heartbeat sending
- Connection management

**Features:**
- Auto-reconnect logic (handled by gRPC client)
- Background heartbeat loop
- Graceful error handling

### Job Executor

`internal/executor/executor.go` - **Not yet implemented**

Planned to handle:
- Job execution from orchestrator
- Log streaming back to orchestrator
- Job status reporting
- Resource isolation

---

## Lifecycle

### Startup Sequence

1. **Detect capabilities** - Read CPU, memory, OS info
2. **Connect to orchestrator** - Establish gRPC connection
3. **Register node** - Send `RegisterNode` RPC with node info
4. **Start heartbeat loop** - Begin sending periodic heartbeats
5. **Wait for shutdown** - Respond to SIGINT/SIGTERM

### Runtime Behavior

- Sends heartbeat every 5 seconds (default)
- Logs connection errors but continues attempting
- Gracefully handles orchestrator restarts (gRPC auto-reconnects)

### Shutdown

- Captures SIGINT/SIGTERM signals
- Closes gRPC connection cleanly
- Exits gracefully

---

## Development

### Running Tests

```powershell
go test ./...
```

### Code Generation

After modifying `shared/proto/v1/orchestrator.proto`:

```powershell
make proto
```

Generated files appear in `internal/proto/v1/`.

### Project Structure Notes

- `internal/` - Private packages (not imported by external code)
- `cmd/node-agent/` - Main entry point (following Go best practices)
- `capabilities/`, `executor/`, `telemetry/` - Legacy directories (can be cleaned up)

---

## Configuration

### Environment Variables

Currently none. All configuration via command-line flags.

### Future: Configuration File

Planned: Support for config file (YAML/JSON) for:
- Orchestrator connection settings
- Heartbeat interval
- Node ID and hostname
- Resource limits
- Job execution settings

---

## Dependencies

- `google.golang.org/grpc` - gRPC client
- `google.golang.org/protobuf` - Protocol Buffers runtime
- `github.com/google/uuid` - Node ID generation
- `github.com/shirou/gopsutil` - System information

---

## Container Integration (vLLM & Ollama)

The node agent can manage vLLM and Ollama containers for AI inference workloads.

**Status:** ⏳ Container management code is in place, integration with main.go is pending.

### Prerequisites

- Docker installed and running
- GPU support (optional, but recommended for performance)

### Command-Line Options

**vLLM Options:**
```
-enable-vllm              Enable vLLM container
-vllm-model               Model to use (default: mistralai/Mistral-7B-Instruct-v0.1)
-vllm-port                vLLM API port (default: 8000)
-vllm-gpus                GPU devices (default: all)
-vllm-tensor-parallel     Tensor parallel size (default: 1)
-vllm-max-model-len       Max model length (default: 4096)
```

**Ollama Options:**
```
-enable-ollama            Enable Ollama container
-ollama-model             Model to use (default: llama2)
-ollama-port              Ollama API port (default: 11434)
-ollama-gpus              GPU devices (default: all)
```

### Usage Examples

```powershell
# Start with vLLM
.\node-agent.exe --enable-vllm --vllm-model mistralai/Mistral-7B-Instruct-v0.1

# Start with Ollama
.\node-agent.exe --enable-ollama --ollama-model llama2

# Both containers
.\node-agent.exe --enable-vllm --enable-ollama
```

### Container Architecture

```
node-agent/internal/containers/
├── manager.go    # Docker container lifecycle management
├── vllm.go       # vLLM container configuration
└── ollama.go     # Ollama container configuration
```

### GPU Support

**Options:**
- `all` - Use all available GPUs (default)
- `0` - Use only GPU 0
- `0,1` - Use GPUs 0 and 1

**Requirements:**
- NVIDIA GPU with drivers
- `nvidia-container-toolkit` installed (Linux)
- Docker configured for GPU access

### Container Troubleshooting

**"docker not found in PATH"**
```powershell
docker --version  # Verify Docker is installed
```

**"Cannot connect to Docker daemon"**
```powershell
Get-Service *docker*  # Check Docker status (Windows)
```

**GPU not accessible:**
```powershell
docker run --rm --gpus all nvidia/cuda:11.0.3-base-ubuntu20.04 nvidia-smi
```

---

## Limitations & Future Work

### Current Limitations

- ⚠️ Job execution not implemented (`executor.go` is empty)
- ⚠️ Memory detection shows Go memory, not system memory
- ⚠️ No GPU detection yet
- ⚠️ No reconnection retry logic (relies on gRPC defaults)
- ⚠️ No configuration file support

### Planned Features

- Job execution framework
- Log streaming to orchestrator
- GPU/accelerator detection
- Resource limit enforcement
- Health check endpoint (for local monitoring)
- Configuration file support
- Auto-reconnection with exponential backoff

---

## Deployment

### As a Service (Windows)

**Future:** Systemd service file or Windows service template

### Docker

**Future:** Dockerfile for containerized deployment

### Manual

Currently designed for manual startup:
```powershell
.\node-agent.exe -orchestrator localhost:50051
```

---

## Troubleshooting

### "Failed to connect to orchestrator"

- Verify orchestrator is running: `Invoke-RestMethod http://localhost:8080/api/nodes`
- Check orchestrator address matches `-orchestrator` flag
- Verify network connectivity and firewall rules

### Node not appearing in orchestrator

- Check orchestrator logs for registration messages
- Verify node agent logs show "Node registered successfully"
- Ensure heartbeat interval is less than orchestrator timeout

### Memory detection inaccurate

This is a known limitation. The agent reports Go's allocated memory, not total system memory. This will be improved in a future release.

---

## Related Documentation

- **Project README:** `../README.md`
- **Quick Start:** `../docs/quick-start.md`
- **Architecture:** `../docs/architecture.md`
- **Project Status:** `../docs/project-status.md`
- **Orchestrator README:** `../orchestrator/README.md`
