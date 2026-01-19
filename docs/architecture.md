# Orchion Architecture

Orchion is a local‑first AI orchestration platform designed for homelab clusters, multi‑node environments, and distributed agent workloads. The system is built as a modular monorepo with clear boundaries between the orchestrator, node agents, dashboard, shared schemas, and developer tooling.

---

## Core Components

### 1. Orchestrator (Go) ✅ **IMPLEMENTED**

The orchestrator is the control plane of Orchion. It is responsible for:
- ✅ Node registration and heartbeat tracking
- ✅ Capability discovery (CPU, memory, OS)
- ✅ Automatic stale node cleanup
- ✅ Proper gRPC error handling with status codes
- ⏳ Job scheduling and routing (interfaces defined, not implemented)
- ✅ Maintaining cluster state (in-memory)
- ✅ Exposing a gRPC/REST API for clients, dashboard, and tools

**Current Implementation Status:**
- ✅ `cmd/orchestrator/main.go` - Main entry point with gRPC & HTTP servers, graceful shutdown
- ✅ `internal/node/registry.go` - In-memory node registry with heartbeat tracking and stale node cleanup
- ✅ `internal/orchestrator/service.go` - gRPC service implementation with proper status codes
- ✅ `api/v1/v1/` - Generated protobuf code

**Key internal modules**
- `node/` — ✅ node registry, health, and telemetry (implemented)
- `orchestrator/` — ✅ gRPC service implementation (implemented)
- `scheduler/` — ⏳ pluggable scheduling strategies (interfaces planned)
- `router/` — ⏳ dispatching jobs to nodes (interfaces planned)
- `api/v1/v1/` — ✅ protobuf definitions and generated code

**Ports:**
- gRPC: `50051` (default)
- HTTP REST API: `8080` (default)

---

### 2. Node Agent (Go) ✅ **IMPLEMENTED**

A lightweight daemon that runs on each machine in the cluster. It:
- ✅ Registers with the orchestrator
- ✅ Sends periodic heartbeats
- ✅ Reports hardware capabilities
- ⏳ Executes jobs sent by the orchestrator (executor.go exists but empty)
- ⏳ Streams logs and telemetry back to the control plane (not implemented)

The agent is intentionally minimal to support:
- Raspberry Pis
- Old desktops
- Servers with GPUs
- Containers or bare metal

**Current Implementation Status:**
- ✅ `cmd/node-agent/main.go` - Complete implementation with registration & heartbeat
- ✅ `internal/capabilities/capabilities.go` - CPU/memory/OS detection using gopsutil for accurate system memory
- ✅ `internal/heartbeat/heartbeat.go` - gRPC client with auto re-registration on orchestrator restart
- ✅ `internal/containers/` - Container management (Docker manager, Ollama/vLLM configs)
- ⏳ `internal/executor/executor.go` - Job execution (empty, placeholder)

---

### 3. Dashboard (SvelteKit) ✅ **PARTIALLY IMPLEMENTED**

A real‑time UI for monitoring and managing the cluster.

**Current Features:**
- ✅ Node list with health and capabilities
- ⏳ Job queue and execution history (not implemented)
- ⏳ Logs and telemetry (not implemented)
- ⏳ Agent status and heartbeat visualization (basic display only)

**Future Features:**
- Model management
- Agent pipelines
- Cluster maps

The dashboard communicates with the orchestrator via HTTP REST API (`/api/nodes` endpoint).

**Current Implementation:**
- ✅ `src/routes/+page.svelte` - Node list display with error handling
- ✅ `src/lib/orchion.ts` - HTTP client with error handling, configurable base URL, and TypeScript types

---

### 4. Shared Schemas

A central location for all cross‑project types.

**Contents:**
- ✅ `proto/v1/orchestrator.proto` — protobuf definitions for orchestrator <-> agent communication
- ⏳ `ts/` — generated TypeScript types for dashboard + VS Code extension (not yet generated)
- ⏳ `zod/` — optional runtime validation schemas (not yet implemented)

This ensures all components speak the same language.

---

### 5. VS Code Extension

Developer tooling for interacting with Orchion directly from the editor.

**Current Status:**
- ✅ Extension structure exists
- ⏳ Basic tree view (placeholder implementation)

**Planned features:**
- Node list tree view
- Job submission
- Log streaming
- Agent pipeline authoring
- YAML/JSON schema validation

---

### 6. Deployments

A collection of deployment options:
- ⏳ Docker Compose (local dev) - directories exist but empty
- ⏳ Kubernetes manifests (optional) - directories exist but empty
- ⏳ Systemd service files for node agents - not created yet

---

## Project Structure

```
Orchion/
├── orchestrator/
│   ├── cmd/orchestrator/
│   │   └── main.go                    ✅ Main entry point
│   ├── internal/
│   │   ├── node/
│   │   │   └── registry.go            ✅ Node registry implementation
│   │   └── orchestrator/
│   │       └── service.go             ✅ gRPC service
│   ├── api/v1/v1/                     ✅ Generated protobuf files
│   ├── go.mod                         ✅ Go module
│   └── Makefile                       ✅ Protobuf generation
│
├── node-agent/
│   ├── cmd/node-agent/
│   │   └── main.go                    ✅ Main entry point
│   ├── internal/
│   │   ├── capabilities/
│   │   │   └── capabilities.go        ✅ Capability detection
│   │   ├── heartbeat/
│   │   │   └── heartbeat.go           ✅ gRPC client
│   │   ├── executor/
│   │   │   └── executor.go            ⏳ Empty placeholder
│   │   └── proto/v1/                  ✅ Generated protobuf files
│   ├── go.mod                         ✅ Go module
│   └── Makefile                       ✅ Protobuf generation
│
├── dashboard/
│   ├── src/
│   │   ├── routes/
│   │   │   └── +page.svelte           ✅ Basic UI
│   │   └── lib/
│   │       └── orchion.ts             ✅ HTTP client
│   └── package.json                   ✅ SvelteKit config
│
├── shared/
│   └── proto/v1/
│       └── orchestrator.proto         ✅ Protocol definitions
│
├── docs/                              ✅ Documentation
└── README.md                          ✅ Main README
```

---

## Data Flow Overview

### 1. **Node Agent → Orchestrator**
   - ✅ Heartbeats (via gRPC `Heartbeat` RPC)
   - ✅ Capabilities (sent during registration)
   - ⏳ Logs (not implemented)
   - ⏳ Job results (not implemented)

### 2. **Orchestrator → Node Agent**
   - ⏳ Job assignments (not implemented)
   - ⏳ Configuration updates (not implemented)

### 3. **Dashboard → Orchestrator**
   - ✅ Fetch cluster state (via HTTP `GET /api/nodes`)
   - ⏳ Submit jobs (not implemented)
   - ⏳ View logs (not implemented)

### 4. **VS Code Extension → Orchestrator**
   - ⏳ Developer‑focused interactions (not implemented)

---

## Current Implementation Status

### ✅ Working Features
- Node registration via gRPC with proper error handling
- Heartbeat tracking with automatic stale node cleanup
- Accurate capability detection (CPU cores, system memory via gopsutil, OS)
- HTTP REST API for dashboard with CORS support
- Dashboard can display registered nodes with error handling
- Auto re-registration when orchestrator restarts
- Graceful shutdown handling for all components
- Proper gRPC status codes throughout (codes.InvalidArgument, codes.NotFound, etc.)
- Container management infrastructure (Docker, Ollama, vLLM)

### ⏳ Partially Implemented
- Dashboard (basic node list, missing auto-refresh and advanced features)
- VS Code extension (structure exists, needs actual node fetching)

### ❌ Not Yet Implemented
- Job execution (executor.go is empty, but container management ready)
- Job scheduling (scheduler/router not implemented)
- Persistent storage (everything is in-memory)
- Authentication/authorization
- Log streaming
- Node unregistration API
- TypeScript type generation from protobuf
- Docker/Kubernetes deployment configs

---

## Design Principles

- **Local‑first** — everything runs on your hardware
- **Homelab‑friendly** — works on Pis, desktops, servers, and mixed clusters
- **Modular** — each component can evolve independently
- **Extensible** — schedulers, agents, and pipelines are pluggable
- **Open** — MIT‑licensed, transparent, and community‑driven

---

## Next Steps

See `docs/roadmap.md` for planned features and priorities.
