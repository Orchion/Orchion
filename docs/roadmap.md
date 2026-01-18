# Orchion Roadmap

This roadmap outlines the planned evolution of Orchion from early prototype to a fully featured local‑first AI orchestration platform.

---

## Phase 1 — Foundations (Now)
- [x] Orchestrator skeleton
- [x] Monorepo structure
- [ ] Node Agent skeleton
- [ ] Shared protobuf definitions
- [ ] Dashboard SvelteKit starter
- [ ] VS Code extension starter
- [ ] Architecture + roadmap docs

---

## Phase 2 — Core Functionality
### Orchestrator
- [ ] Node registration API
- [ ] Heartbeat tracking
- [ ] Capability discovery (CPU/GPU)
- [ ] Basic scheduler (round‑robin)
- [ ] Job routing + execution tracking
- [ ] Logging + telemetry endpoints

### Node Agent
- [ ] Heartbeat loop
- [ ] Capability reporting
- [ ] Job executor
- [ ] Log streaming
- [ ] Auto‑reconnect logic

### Dashboard
- [ ] Node list UI
- [ ] Node detail view
- [ ] Job queue view
- [ ] Log viewer
- [ ] Cluster health overview

---

## Phase 3 — Developer Tooling
### VS Code Extension
- [ ] Orchion Nodes tree view
- [ ] Job submission panel
- [ ] Log streaming
- [ ] Pipeline authoring
- [ ] Schema validation

### Shared Schemas
- [ ] TS type generation
- [ ] Zod validation schemas
- [ ] Versioned API definitions

---

## Phase 4 — Advanced Features
- [ ] GPU‑aware scheduling
- [ ] Model management (vLLM, Exo, Ollama)
- [ ] Multi‑agent pipelines
- [ ] Distributed caching
- [ ] Secrets + config management
- [ ] Webhooks + triggers

---

## Phase 5 — Homelab Enhancements
- [ ] Systemd service templates
- [ ] Docker Compose bundles
- [ ] Kubernetes manifests
- [ ] Auto‑discovery of nodes
- [ ] Local network mesh mode

---

## Phase 6 — Long‑Term Vision
- [ ] Plugin system for custom agents
- [ ] Multi‑cluster federation
- [ ] Enterprise‑grade RBAC + audit logs

---

## Philosophy
Orchion will always be:
- Local‑first  
- Privacy‑respecting  
- Homelab‑friendly  
- Open source  
- Extensible  