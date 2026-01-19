# Orchion Scripts

Centralized scripts for managing the Orchion monorepo. These scripts simplify common operations like building, running, and testing all components.

---

## 📋 Available Scripts

### `setup-all.ps1` ⭐ **RUN THIS FIRST**

Sets up the entire development environment - installs all dependencies and prerequisites.

**Usage:**
```powershell
.\shared\scripts\setup-all.ps1
```

**What it does:**
- Checks for required tools (Go, Node.js, npm, protoc, Docker)
- Installs Go protobuf plugins if missing
- Installs Go dependencies for orchestrator and node-agent
- Generates protobuf files for both Go components
- Installs npm dependencies for dashboard
- Installs Playwright browsers for dashboard tests
- Reports any missing prerequisites

**When to run:**
- First time setting up the project
- After cloning the repository
- After pulling changes that modify dependencies
- When dependencies are out of sync

**Prerequisites:**
- Go 1.21+ installed
- Node.js 18+ installed
- npm installed (comes with Node.js)
- protoc (optional but recommended)

---

### `build-all.ps1`
Builds all Orchion components (orchestrator, node-agent, and dashboard).

**Usage:**
```powershell
.\shared\scripts\build-all.ps1
```

**What it does:**
- Builds orchestrator to `orchestrator/orchestrator.exe`
- Builds node-agent to `node-agent/node-agent.exe`
- Builds dashboard (production build) to `dashboard/build/`
- Installs dashboard dependencies if needed
- Reports success/failure for each component

**Note:** Dashboard build failures are non-critical (warnings only) since dashboard is typically run in dev mode.

---

### `run-all.ps1`
Starts all Orchion components in separate PowerShell windows.

**Usage:**
```powershell
.\shared\scripts\run-all.ps1
```

**What it does:**
- Starts orchestrator in a new window
- Starts node-agent in a new window (after a 2-second delay)
- Verifies executables exist before starting

**Prerequisites:**
- Components must be built first (`build-all.ps1`)
- Orchestrator must be running before node-agent can connect

**Note:** Each component runs in its own window. Press Ctrl+C in each window to stop them.

---

### `proto-gen.ps1`
Generates protobuf files for all components.

**Usage:**
```powershell
.\shared\scripts\proto-gen.ps1
```

**What it does:**
- Generates protobuf files for orchestrator
- Generates protobuf files for node-agent
- Uses `make proto` if available, falls back to direct `protoc` commands

**When to use:**
- After modifying `shared/proto/v1/orchestrator.proto`
- After cloning the repository
- When protobuf files are missing or outdated

---

### `clean-all.ps1`
Removes all build artifacts.

**Usage:**
```powershell
.\shared\scripts\clean-all.ps1
```

**What it does:**
- Removes `orchestrator/orchestrator.exe`
- Removes `node-agent/node-agent.exe`
- Removes `dashboard/build/` and `dashboard/.svelte-kit/`
- Does NOT remove generated protobuf files (use `git clean` for that)
- Does NOT remove `dashboard/node_modules/` (use `npm run clean` in dashboard/ for that)

---

### `test-all.ps1`
Runs tests for all Orchion components.

**Usage:**
```powershell
.\shared\scripts\test-all.ps1
```

**What it does:**
- Runs Go tests for orchestrator
- Runs Go tests for node-agent
- Runs npm tests for dashboard
- Reports pass/fail for each component

**Prerequisites:**
- Go components must be buildable
- Dashboard dependencies installed (auto-installs if missing)

---

### `test-api.ps1`
Tests the Orchion REST API.

---

### `dev-dashboard.ps1`
Starts the dashboard development server.

**Usage:**
```powershell
.\shared\scripts\dev-dashboard.ps1
```

**What it does:**
- Installs dependencies if needed (`npm install`)
- Starts the SvelteKit development server
- Opens at `http://localhost:5173` (or next available port)

**Prerequisites:**
- Node.js 18+ installed
- Orchestrator should be running for API access

**Note:** This runs the dev server in the current terminal. Press Ctrl+C to stop it.

---

**Usage:**
```powershell
.\shared\scripts\test-api.ps1
```

**What it does:**
- Fetches nodes from `http://localhost:8080/api/nodes`
- Displays registered nodes in JSON format
- Shows helpful error messages if orchestrator isn't running

**Prerequisites:**
- Orchestrator must be running

---

## 🚀 Common Workflows

### Initial Setup
```powershell
# 1. Setup everything (installs dependencies, generates protobuf, etc.)
.\shared\scripts\setup-all.ps1

# 2. Build all components
.\shared\scripts\build-all.ps1
```

**Note:** `setup-all.ps1` handles protobuf generation and dependency installation automatically.

### Daily Development
```powershell
# Build everything (Go + Dashboard)
.\shared\scripts\build-all.ps1

# Run everything (starts in separate windows)
.\shared\scripts\run-all.ps1

# Start dashboard separately (optional, or modify run-all.ps1 to include it)
.\shared\scripts\dev-dashboard.ps1

# Test everything
.\shared\scripts\test-all.ps1

# Test the API
.\shared\scripts\test-api.ps1
```

### After Changing Protobuf Files
```powershell
# Regenerate protobuf
.\shared\scripts\proto-gen.ps1

# Rebuild components
.\shared\scripts\build-all.ps1
```

### Clean Start
```powershell
# Remove build artifacts
.\shared\scripts\clean-all.ps1

# Rebuild
.\shared\scripts\build-all.ps1
```

---

## 🔧 Alternative: Using Make (Cross-Platform)

If you prefer using `make` (works on Windows with `choco install make`):

```bash
# Generate protobuf
make proto

# Build all
make build

# Clean
make clean

# See all commands
make help
```

See the root `Makefile` for available commands.

---

## 📝 Script Requirements

- **PowerShell 5.1+** (Windows 10+ default)
- **Go 1.21+** (for building)
- **protoc** (for `proto-gen.ps1`)
- **make** (optional, for `proto-gen.ps1`)

---

## 🐛 Troubleshooting

### Script Execution Policy

If you get "cannot be loaded because running scripts is disabled":

```powershell
# Run once as Administrator
Set-ExecutionPolicy -ExecutionPolicy RemoteSigned -Scope CurrentUser
```

### "Cannot find path"

Make sure you're running scripts from the project root, not from `shared/scripts/`:
```powershell
# From project root:
.\shared\scripts\build-all.ps1

# NOT from shared/scripts:
cd shared\scripts
.\build-all.ps1  # This will fail
```

### Build Fails

1. Make sure Go is installed and in PATH: `go version`
2. Make sure dependencies are installed: `cd orchestrator && go mod tidy`
3. Make sure protobuf files are generated: `.\shared\scripts\proto-gen.ps1`

---

## 💡 Tips

- Keep scripts in sync with project structure - if you add new components, update these scripts
- All scripts use `$ErrorActionPreference = "Stop"` to fail fast on errors
- Scripts assume they're run from the project root directory
- Consider adding these scripts to your PATH or creating aliases

---

## 🔮 Future Script Ideas

Potential scripts to add:
- `setup-dev.ps1` - Complete development environment setup
- `test-all.ps1` - Run all tests across components
- `deploy-local.ps1` - Deploy to local Docker/Kubernetes
- `watch-build.ps1` - Watch for changes and rebuild automatically
