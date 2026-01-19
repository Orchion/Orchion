# Shared Resources

Central location for cross-project resources used throughout the Orchion monorepo.

---

## 📁 Directory Structure

```
shared/
├── proto/              # Protocol Buffer definitions
│   └── v1/
│       └── orchestrator.proto
├── scripts/            # Project-wide scripts (build, run, test, etc.)
│   ├── build-all.ps1
│   ├── run-all.ps1
│   ├── proto-gen.ps1
│   ├── clean-all.ps1
│   ├── test-api.ps1
│   └── README.md
├── ts/                 # TypeScript type definitions (planned)
└── zod/                # Zod validation schemas (planned)
```

---

## 🚀 Quick Start

### Using Scripts (Recommended)

**Build all components:**
```powershell
.\shared\scripts\build-all.ps1
```

**Run all components:**
```powershell
.\shared\scripts\run-all.ps1
```

**Generate protobuf files:**
```powershell
.\shared\scripts\proto-gen.ps1
```

See `shared/scripts/README.md` for full documentation.

---

## 📋 Components

### Protocol Buffers (`proto/`)

Shared protocol definitions for gRPC communication between:
- Orchestrator ↔ Node Agent
- Orchestrator ↔ Dashboard (via REST API)
- Future: Orchestrator ↔ VS Code Extension

**Files:**
- `proto/v1/orchestrator.proto` - Main protocol definition

**Usage:**
- After modifying `.proto` files, regenerate code: `.\shared\scripts\proto-gen.ps1`
- Generated files appear in `orchestrator/api/v1/v1/` and `node-agent/internal/proto/v1/`

---

### Scripts (`scripts/`)

Centralized scripts for managing the monorepo. See `shared/scripts/README.md` for details.

**Available Scripts:**
- `build-all.ps1` - Build all components
- `run-all.ps1` - Run all components in separate windows
- `proto-gen.ps1` - Generate protobuf files
- `clean-all.ps1` - Remove build artifacts
- `test-api.ps1` - Test the REST API

---

### TypeScript Types (`ts/`) ⏳ Planned

Future: Generated TypeScript types from protobuf definitions for use in:
- Dashboard
- VS Code Extension

---

### Zod Schemas (`zod/`) ⏳ Planned

Future: Runtime validation schemas for TypeScript projects.

---

## 🔧 Make Commands (Alternative)

If you prefer using `make`, there's a root-level `Makefile`:

```bash
make proto      # Generate protobuf files
make build      # Build all components
make clean      # Remove build artifacts
make help       # Show all commands
```

---

## 📝 Notes

- All scripts assume they're run from the project root directory
- Scripts use PowerShell (Windows default)
- For cross-platform support, use the root `Makefile` or WSL
- Generated protobuf files are git-ignored and must be regenerated after cloning
