# Orchion

Run AI agents across every machine you own

## Overview

Orchion is a distributed orchestrator service for managing and coordinating AI agent workloads across multiple machines. It provides a gRPC-based API for node registration, heartbeat monitoring, and task scheduling.

## Project Structure

```
.
├── api/
│   └── proto/v1/          # gRPC protocol definitions
│       └── orchestrator.proto
├── cmd/
│   └── orchestrator/      # Main orchestrator application
│       └── main.go
├── internal/
│   ├── node/              # Node registry implementation
│   │   ├── registry.go
│   │   ├── registry_impl.go
│   │   └── registry_test.go
│   ├── orchestrator/      # Orchestrator service
│   │   ├── service.go
│   │   └── service_test.go
│   ├── scheduler/         # Scheduler interface (extensible)
│   │   └── scheduler.go
│   └── router/            # Router interface (extensible)
│       └── router.go
├── Makefile              # Build and development tasks
├── go.mod                # Go module definition
└── README.md
```

## Features

- **gRPC API**: Clean, type-safe API for node management
- **Node Registry**: In-memory registry for tracking registered nodes
- **Heartbeat Monitoring**: Automatic detection of inactive nodes
- **Clean Architecture**: Interfaces for schedulers and routers allow easy extension
- **Graceful Shutdown**: Proper signal handling for clean shutdowns

## Building

### Prerequisites

- Go 1.21 or later
- Protocol Buffers compiler (`protoc`)

### Install Tools

```bash
make install-tools
```

### Generate gRPC Code

```bash
make proto
```

### Build the Orchestrator

```bash
make build
```

The binary will be created at `bin/orchestrator`.

## Running

Start the orchestrator service:

```bash
./bin/orchestrator
```

### Command-line Options

- `--port`: gRPC server port (default: 50051)
- `--heartbeat-timeout`: Node heartbeat timeout duration (default: 30s)

Example:

```bash
./bin/orchestrator --port 50051 --heartbeat-timeout 1m
```

## Testing

Run all tests:

```bash
make test
```

Run specific package tests:

```bash
go test ./internal/node/... -v
go test ./internal/orchestrator/... -v
```

## API

The orchestrator provides the following gRPC endpoints:

- `RegisterNode`: Register a new node with the orchestrator
- `Heartbeat`: Send heartbeat to indicate node is alive
- `UnregisterNode`: Remove a node from the orchestrator
- `ListNodes`: List all registered nodes with optional status filtering

## Development

### Code Generation

After modifying `.proto` files, regenerate the code:

```bash
make proto
```

### Code Formatting

```bash
make fmt
```

### Cleaning Build Artifacts

```bash
make clean
```

## Architecture

### Node Registry

The node registry is an in-memory data structure that tracks all registered nodes. Each node has:
- Unique ID
- Address
- Metadata (key-value pairs)
- Status (active, inactive, disconnected)
- Registration timestamp
- Last heartbeat timestamp

### Heartbeat Monitoring

A background goroutine periodically checks node heartbeats and marks nodes as inactive if they haven't sent a heartbeat within the configured timeout period.

### Scheduler & Router Interfaces

The project defines clean interfaces for future scheduler and router implementations:

- **Scheduler**: Responsible for assigning tasks to nodes
- **Router**: Responsible for routing messages between nodes

These can be extended with different strategies (round-robin, least-loaded, geographic, etc.)

## License

See LICENSE file for details.
