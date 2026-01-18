# Orchion Architecture

Orchion is a local‑first AI orchestration platform designed for homelab clusters, multi‑node environments, and distributed agent workloads. The system is built as a modular monorepo with clear boundaries between the orchestrator, node agents, dashboard, shared schemas, and developer tooling.

---

## Core Components

### 1. Orchestrator (Go)
The orchestrator is the control plane of Orchion. It is responsible for:
- Node registration and heartbeat tracking
- Capability discovery (CPU, GPU, memory, accelerators)
- Job scheduling and routing
- Maintaining cluster state
- Exposing a gRPC/REST API for clients, dashboard, and tools

**Key internal modules**
- `scheduler/` — pluggable scheduling strategies
- `router/` — dispatching jobs to nodes
- `node/` — node registry, health, and telemetry
- `api/` — protobuf definitions and generated code

---

### 2. Node Agent (Go or Rust)
A lightweight daemon that runs on each machine in the cluster. It:
- Registers with the orchestrator
- Sends periodic heartbeats
- Reports hardware capabilities
- Executes jobs sent by the orchestrator
- Streams logs and telemetry back to the control plane

The agent is intentionally minimal to support:
- Raspberry Pis
- Old desktops
- Servers with GPUs
- Containers or bare metal

---

### 3. Dashboard (SvelteKit)
A real‑time UI for monitoring and managing the cluster.

Features:
- Node list with health and capabilities
- Job queue and execution history
- Logs and telemetry
- Agent status and heartbeat visualization
- Future: model management, agent pipelines, cluster maps

The dashboard communicates with the orchestrator via a small TypeScript client generated from shared protobuf definitions.

---

### 4. Shared Schemas
A central location for all cross‑project types.

Contents:
- `proto/` — protobuf definitions for orchestrator <-> agent <-> dashboard
- `ts/` — generated TypeScript types for dashboard + VS Code extension
- `zod/` — optional runtime validation schemas

This ensures all components speak the same language.

---

### 5. VS Code Extension
Developer tooling for interacting with Orchion directly from the editor.

Planned features:
- Node list tree view
- Job submission
- Log streaming
- Agent pipeline authoring
- YAML/JSON schema validation

---

### 6. Deployments
A collection of deployment options:
- Docker Compose (local dev)
- Kubernetes manifests (optional)
- Systemd service files for node agents
- Example homelab setups

---

## Data Flow Overview

1. **Node Agent → Orchestrator**
   - Heartbeats
   - Capabilities
   - Logs
   - Job results

2. **Orchestrator → Node Agent**
   - Job assignments
   - Configuration updates

3. **Dashboard → Orchestrator**
   - Fetch cluster state
   - Submit jobs
   - View logs

4. **VS Code Extension → Orchestrator**
   - Developer‑focused interactions

---

## Design Principles

- **Local‑first** — everything runs on your hardware
- **Homelab‑friendly** — works on Pis, desktops, servers, and mixed clusters
- **Modular** — each component can evolve independently
- **Extensible** — schedulers, agents, and pipelines are pluggable
- **Open** — MIT‑licensed, transparent, and community‑driven