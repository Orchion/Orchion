# Deployments

Deployment configurations and templates for Orchion components.

---

## Overview

This directory contains deployment options for running Orchion in various environments:

- **Docker Compose** - Local development and testing
- **Docker** - Container images and Dockerfiles
- **Kubernetes** - Production cluster deployment

**Current Status:** ⏳ Directories created. Configurations planned.

---

## Directory Structure

```
deployments/
├── compose/          # Docker Compose configurations
├── docker/           # Dockerfiles and build scripts
└── k8s/              # Kubernetes manifests
```

---

## Planned Contents

### Docker Compose (`compose/`)

**Planned files:**
- `docker-compose.yml` - Full stack (orchestrator + node-agents)
- `docker-compose.dev.yml` - Development setup with hot reload
- `.env.example` - Environment variable template

**Use cases:**
- Local development
- Testing multi-node setups
- Quick start demos

---

### Docker (`docker/`)

**Planned files:**
- `Dockerfile.orchestrator` - Orchestrator container image
- `Dockerfile.node-agent` - Node agent container image
- `build.sh` / `build.ps1` - Build scripts
- `.dockerignore` - Exclude files from builds

**Use cases:**
- Containerized deployments
- CI/CD pipelines
- Cloud deployments

---

### Kubernetes (`k8s/`)

**Planned files:**
- `namespace.yaml` - Orchion namespace
- `orchestrator-deployment.yaml` - Orchestrator deployment
- `orchestrator-service.yaml` - Orchestrator service
- `node-agent-daemonset.yaml` - Node agent DaemonSet
- `configmap.yaml` - Configuration
- `secrets.yaml.example` - Secrets template

**Use cases:**
- Production cluster deployment
- Multi-node clusters
- Scalable setups

---

## Future: Systemd Services

**Planned location:** `deployments/systemd/`

**Planned files:**
- `orchestrator.service` - Systemd service for orchestrator
- `node-agent.service` - Systemd service for node agent
- `install.sh` - Installation script

**Use cases:**
- Bare metal deployments
- Raspberry Pi clusters
- Traditional Linux servers

---

## Current Status

These directories are placeholders. Deployment configurations will be added as part of **Phase 5 - Homelab Enhancements** (see `../docs/roadmap.md`).

---

## Contributing

When adding deployment configs:

1. Follow existing patterns and best practices for each technology
2. Include documentation comments in configs
3. Provide `.example` or `.template` files for sensitive settings
4. Test configurations in target environments
5. Update this README with new files and usage instructions

---

## Related Documentation

- **Roadmap:** `../docs/roadmap.md` (Phase 5)
- **Architecture:** `../docs/architecture.md`
- **Quick Start:** `../docs/quick-start.md` (for manual deployment)
