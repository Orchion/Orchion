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

**Status:** All foundational components are in place and minimally working.

---

## Phase 2 — Core Functionality

### Orchestrator
- [x] Node registration API ✅
- [x] Heartbeat tracking ✅
- [x] Capability discovery (CPU/memory/OS) ✅
- [x] HTTP REST API for dashboard ✅
- [x] Automatic stale node cleanup ✅
- [x] Proper gRPC status codes ✅
- [ ] Job scheduling (round‑robin) ⏳
- [ ] Job routing + execution tracking ⏳
- [ ] Logging + telemetry endpoints ⏳
- [ ] Node unregistration API ⏳

### Node Agent
- [x] Heartbeat loop ✅
- [x] Capability reporting ✅
- [x] Auto‑reconnect/re-registration logic ✅
- [x] Accurate system memory detection (gopsutil) ✅
- [x] Container management infrastructure ✅
- [ ] Job executor ⏳
- [ ] Log streaming ⏳

### Dashboard
- [x] Node list UI ✅
- [x] Error handling and user feedback ✅
- [x] Configurable orchestrator URL ✅
- [x] TypeScript type safety ✅
- [ ] Auto-refresh/polling ⏳
- [ ] Node detail view ⏳
- [ ] Job queue view ⏳
- [ ] Log viewer ⏳
- [ ] Cluster health overview ⏳

**Status:** Core node management is working. Job execution and scheduling are the next priorities.

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

## Phase 4 — Advanced Features
- [ ] GPU‑aware scheduling ⏳
- [ ] Model management (vLLM, Exo, Ollama) ⏳
- [ ] Multi‑agent pipelines ⏳
- [ ] Distributed caching ⏳
- [ ] Secrets + config management ⏳
- [ ] Webhooks + triggers ⏳

---

## Phase 5 — Homelab Enhancements
- [ ] Systemd service templates ⏳
- [ ] Docker Compose bundles ⏳
- [ ] Kubernetes manifests ⏳
- [ ] Auto‑discovery of nodes ⏳
- [ ] Local network mesh mode ⏳

---

## Phase 6 — Long‑Term Vision
- [ ] Plugin system for custom agents ⏳
- [ ] Multi‑cluster federation ⏳
- [ ] Enterprise‑grade RBAC + audit logs ⏳

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

### ⏳ In Progress / Next Up
- Job execution framework
- Job scheduling system
- Enhanced dashboard features (auto-refresh, job queue)
- VS Code extension functionality (real node fetching)

### 📅 Timeline
- **Phase 1:** ✅ Complete
- **Phase 2:** ⏳ Core functionality in progress (~60% complete - node management done, job execution pending)
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

## How to Contribute

See `quick-start.md` for testing instructions and `development-setup.md` for development environment setup.
