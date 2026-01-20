# Orchion Roadmap

This roadmap outlines the planned evolution of Orchion from early prototype to a fully featured local‑first AI orchestration platform.

---

## Phase 1 — Foundations ✅ **COMPLETE**

- [x] Orchestrator skeleton
- [x] Monorepo structure
- [x] Node Agent skeleton
- [x] Shared protobuf definitions
- [x] Dashboard SvelteKit starter
- [x] VS Code extension starter
- [x] Architecture + roadmap docs
- [x] **Code Quality Tools**: golangci-lint, ESLint, Prettier, automated scripts
- [x] **Development Automation**: format-all.ps1, lint-all.ps1, setup-all.ps1

**Status:** All foundational components are in place and minimally working.

---

## Phase 2 — Core Functionality (Revised)

### Orchestrator
- [x] Node registration API ✅
- [x] Heartbeat tracking ✅
- [x] Capability discovery (CPU/memory/OS) ✅
- [x] HTTP REST API for dashboard ✅
- [x] Automatic stale node cleanup ✅
- [x] Proper gRPC status codes ✅
- [x] All node-management features (registration, heartbeat, cleanup, REST API) ✅ **COMPLETE**
- [x] Job scheduling (round‑robin) ✅ **COMPLETE**
- [x] Job routing + execution tracking ✅ **COMPLETE**
- [x] Job submission and status APIs ✅ **COMPLETE**
- [x] OpenAI-compatible HTTP gateway ✅ **COMPLETE**
- [x] Centralized logging system ✅ **COMPLETE**
- [ ] Node unregistration API ⏳

### Node Agent
- [x] Heartbeat loop ✅
- [x] Capability reporting ✅
- [x] Auto‑reconnect/re-registration logic ✅
- [x] Accurate system memory detection (gopsutil) ✅
- [x] Container management infrastructure ✅ **COMPLETE**
- [x] Structured logging ✅ **COMPLETE**
- [x] Job executor ✅ **COMPLETE** (full node-agent implementation with container management)
- [ ] Log streaming to orchestrator ⏳ **HIGH PRIORITY**

### Dashboard
- [x] Node list UI ✅
- [x] Error handling and user feedback ✅
- [x] Configurable orchestrator URL ✅
- [x] TypeScript type safety ✅ **COMPLETE**
- [x] Log viewer with real-time streaming ✅ **COMPLETE**
- [ ] Auto-refresh/polling ⏳ **NEXT UP**
- [ ] Job queue view ⏳ **HIGH PRIORITY**
- [ ] Node detail view ⏳ **NEXT UP**
- [ ] Cluster health overview ⏳

**Status:** Node management and full job execution pipeline are complete. Log streaming from node-agent to orchestrator is the next priority.

---

## Phase 3 — Developer Tooling

### VS Code Extension
- [x] Extension structure ✅
- [ ] Orchion Nodes tree view ⏳
- [ ] Job submission panel ⏳
- [ ] Log streaming ⏳
- [ ] Pipeline authoring ⏳
- [ ] Schema validation ⏳

### Shared Schemas
- [ ] TS type generation ⏳
- [ ] Zod validation schemas ⏳
- [ ] Versioned API definitions ⏳

**Status:** Basic extension structure exists. Core functionality needs implementation.

---

## Phase 3 — Developer Tooling (Updated)

### VS Code Extension
- [x] Extension structure ✅
- [x] Log streaming from orchestrator ✅ **COMPLETE**
- [ ] Fetch nodes from orchestrator ⏳
- [ ] Submit jobs ⏳
- [ ] Show job history ⏳
- [ ] Pipeline authoring ⏳ (later)
- [ ] Schema validation ⏳

### Shared Schemas
- [ ] TS type generation ⏳
- [ ] Zod validation schemas ⏳ (optional)
- [ ] Shared types for dashboard + VS Code ⏳

---

## Phase 3.5 — Logging & Observability Enhancements

### Centralized Logging (Implemented)
- [x] Structured logging library with logrus ✅
- [x] Protobuf definitions for log streaming ✅
- [x] Orchestrator log service with gRPC streaming ✅
- [x] Node-agent structured logging ✅
- [x] Dashboard log viewer with Server-Sent Events ✅
- [x] VS Code extension log integration ✅
- [x] HTTP REST API for logs (`/api/logs`) ✅

### Next Logging Enhancements (High Priority)
- [ ] **Persistent log storage** ⏳ **HIGH PRIORITY**
  - SQLite-based log persistence
  - Log retention policies
  - Query historical logs
- [ ] **Complete log streaming pipeline** ⏳ **HIGH PRIORITY**
  - Node-agent → orchestrator gRPC streaming
  - Real-time log aggregation
  - Streaming to all connected clients
- [ ] **Advanced log filtering & search** ⏳ **MEDIUM PRIORITY**
  - Filter by source, level, time range
  - Search log content
  - Export logs functionality

---

## Phase 4 — Advanced Features
- [ ] GPU‑aware scheduling ⏳ (moved down in priority)
- [ ] Model management (vLLM, Exo, Ollama) ⏳
- [ ] Multi‑agent pipelines ⏳
- [ ] Distributed caching ⏳
- [ ] Secrets + config management ⏳
- [ ] Webhooks + triggers ⏳

---

## Phase 5 — Homelab Enhancements
- [ ] Systemd service templates ⏳
- [ ] Docker Compose bundles ⏳
- [ ] Kubernetes manifests ⏳ (moved down in priority)
- [ ] Auto‑discovery of nodes ⏳
- [ ] Local network mesh mode ⏳

---

## Phase 6 — Long‑Term Vision
- [ ] Plugin system for custom agents ⏳
- [ ] Multi‑cluster federation ⏳ (moved down in priority)
- [ ] Enterprise‑grade RBAC + audit logs ⏳
- [ ] Authentication/Authorization ⏳ (moved down in priority)
- [ ] Persistent storage (Postgres) ⏳ (moved down in priority)

---

## Current Status Summary

### ✅ What's Working Now
- Orchestrator runs gRPC and HTTP servers with proper error handling
- Node agents can register and send heartbeats
- Automatic stale node cleanup
- Auto re-registration when orchestrator restarts
- Dashboard can display registered nodes with error handling
- Accurate capability detection (system memory, CPU, OS)
- Proper gRPC status codes throughout
- Container management infrastructure ready
- Build/run scripts for easy development

### ⏳ In Progress / Next Up (Priority Order)
1. **Node-agent job executor** (Highest Priority)
   - Implement executor.go to run Ollama/vLLM containers
   - Execute jobs and return results to orchestrator
   - Add basic log streaming
2. **End-to-end testing** (High Priority)
   - Complete unit test coverage (95%+ achieved on core components)
   - Test complete job submission → execution → completion flow
   - Verify OpenAI-compatible API gateway works
3. **Log streaming completion** (High Priority)
   - Complete node-agent → orchestrator streaming
   - Add persistent log storage
4. **Dashboard enhancements** (High Priority)
   - Auto-refresh
   - Job queue view
   - Node detail view
   - Log viewer

### 📅 Recommended Timeline

**Week 1:**
- Implement node-agent executor.go (run containers, execute jobs)
- Test end-to-end job execution flow
- Complete log streaming (node-agent → orchestrator)

**Week 2:**
- Dashboard auto-refresh
- Dashboard job queue view
- Node detail view
- Add persistent log storage

**Week 3:**
- VS Code extension: fetch nodes
- VS Code extension: submit jobs
- VS Code extension: log streaming

**Week 4:**
- Add SQLite persistence
- Add TS type generation
- Add Zod schemas (optional)

### 📅 Overall Progress
- **Phase 1:** ✅ Complete
- **Phase 2:** ⏳ Core functionality in progress (~90% complete - node management and orchestrator job system done, node-agent executor is next)
- **Phase 3+:** 🔜 Future work

---

## Philosophy

Orchion will always be:
- Local‑first  
- Privacy‑respecting  
- Homelab‑friendly  
- Open source  
- Extensible  

---

---

## Job Model Definition

A minimal Job model is needed to support job execution:

```typescript
Job {
  id: string
  nodeId?: string
  status: "pending" | "running" | "completed" | "failed"
  createdAt: number
  startedAt?: number
  finishedAt?: number
  payload: {
    model: string
    prompt: string
    params?: Record<string, any>
  }
  logs?: string[]
}
```

This should be added to:
- `shared/proto/v1/job.proto` (new file)
- Orchestrator job queue
- Dashboard types
- VS Code extension

---

## Persistence Layer

A minimal persistence layer using SQLite (no Postgres needed yet):

**Minimal persistence to add:**
- Job history table
- Node registry table (optional)
- Log storage (optional)

**Priority:** Can be added after job execution works (Week 4)

---

## Why Job Execution is Highest Priority

Your architecture, node agent, and container manager are already in place. You're 80% of the way to a functional inference cluster. Implementing job execution turns Orchion from "node registry" into "actual orchestrator."

**Current State:**
- ✅ Node registration and heartbeat tracking
- ✅ Container management infrastructure
- ✅ Accurate system memory detection
- ✅ Dashboard with TypeScript type safety
- ✅ All node-management features complete
- ✅ Job execution loop (orchestrator-side)
- ✅ Job execution on node-agent (container management)
- ✅ Job scheduling and API endpoints

**What's Missing:**
- ❌ Log streaming from node-agent to orchestrator

---

## How to Contribute

See `quick-start.md` for testing instructions and `development-setup.md` for development environment setup.
