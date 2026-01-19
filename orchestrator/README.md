# Orchestrator

The control plane for Orchion - manages node registration, heartbeat tracking, and provides APIs for the dashboard and tools.

---

## Overview

The orchestrator is the central component of Orchion. It:
- Registers and tracks nodes in the cluster
- Monitors node health via heartbeats
- Exposes gRPC API for node agents
- Exposes HTTP REST API for dashboard
- Provides node listing and status information

**Current Status:** ✅ Core functionality implemented and working

---

## Architecture

```
orchestrator/
├── cmd/orchestrator/          # Main entry point
│   └── main.go                # gRPC & HTTP server setup
├── internal/
│   ├── node/                  # Node registry implementation
│   │   └── registry.go        # In-memory node storage
│   └── orchestrator/          # gRPC service implementation
│       └── service.go         # RegisterNode, Heartbeat, ListNodes
├── api/v1/v1/                 # Generated protobuf files
├── go.mod                     # Go module definition
└── Makefile                   # Protobuf generation
```

---

## Building

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)
- Go protobuf plugins (`protoc-gen-go`, `protoc-gen-go-grpc`)

### Build

**Using the monorepo script:**
```powershell
# From project root
.\shared\scripts\build-all.ps1
```

**Or manually:**
```powershell
go build -o orchestrator.exe ./cmd/orchestrator
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

```powershell
.\orchestrator.exe
```

### Command-Line Options

```
-port              gRPC server port (default: 50051)
-http-port         HTTP REST API port (default: 8080)
-heartbeat-timeout Node heartbeat timeout duration (default: 30s)
```

### Examples

```powershell
# Use custom ports
.\orchestrator.exe -port 50052 -http-port 8081

# Longer heartbeat timeout
.\orchestrator.exe -heartbeat-timeout 60s

# Production-like settings
.\orchestrator.exe -port 50051 -http-port 8080 -heartbeat-timeout 1m
```

---

## API

### gRPC API (Port 50051)

The orchestrator implements the following gRPC service methods:

- **`RegisterNode`** - Register a new node with the orchestrator
- **`Heartbeat`** - Update heartbeat timestamp for a registered node
- **`ListNodes`** - List all registered nodes

See `shared/proto/v1/orchestrator.proto` for protocol definitions.

### HTTP REST API (Port 8080)

- **`GET /api/nodes`** - List all registered nodes (JSON)

**Example:**
```powershell
Invoke-RestMethod http://localhost:8080/api/nodes
```

---

## Components

### Node Registry

The node registry (`internal/node/registry.go`) maintains an in-memory store of all registered nodes.

**Features:**
- Thread-safe operations (mutex-protected)
- Automatic heartbeat timestamp tracking
- Stale node detection (via `CheckHeartbeats`)

**Note:** Currently in-memory only - restarting the orchestrator loses all node data.

### Orchestrator Service

The gRPC service (`internal/orchestrator/service.go`) handles all RPC requests from node agents.

**Responsibilities:**
- Validating registration requests
- Updating heartbeat timestamps
- Returning node lists

### Heartbeat Monitor

A background goroutine in `main.go` periodically checks for stale nodes (every 10 seconds) and logs nodes that haven't sent a heartbeat within the timeout period.

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

Generated files appear in `api/v1/v1/`.

### Project Structure Notes

- `api/v1/v1/` - Generated protobuf files (nested `v1/` is correct - matches proto package)
- `internal/` - Private packages (not imported by external code)
- `cmd/orchestrator/` - Main entry point (following Go best practices)

---

## Configuration

### Environment Variables

Currently none. All configuration via command-line flags.

### Future: Configuration File

Planned: Support for config file (YAML/JSON) for:
- Port settings
- Heartbeat timeout
- Logging configuration
- Storage backend selection

---

## Dependencies

- `google.golang.org/grpc` - gRPC framework
- `google.golang.org/protobuf` - Protocol Buffers runtime

---

## Limitations & Future Work

### Current Limitations

- ⚠️ In-memory storage - data lost on restart
- ⚠️ No authentication/authorization
- ⚠️ No persistent storage
- ⚠️ Single instance only (no clustering)

### Planned Features

- Database-backed storage
- Authentication/authorization
- Health check endpoints
- Metrics/telemetry endpoints
- Configuration file support
- Job scheduling (Phase 2)
- Multi-instance clustering

---

## Troubleshooting

### "Address already in use"

Another process is using port 50051 or 8080. Use different ports:
```powershell
.\orchestrator.exe -port 50052 -http-port 8081
```

### "Cannot find package" errors

Regenerate protobuf files:
```powershell
make proto
```

### Nodes not appearing

- Check that node agents are actually connecting
- Verify gRPC port matches node agent configuration
- Check orchestrator logs for registration messages

---

## Related Documentation

- **Project README:** `../README.md`
- **Architecture:** `../docs/architecture.md`
- **Quick Start:** `../docs/quick-start.md`
- **API Documentation:** `../docs/api.md` (when implemented)
