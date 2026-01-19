# Orchion Scripts

Centralized scripts for managing the Orchion monorepo. These scripts simplify common operations like building, running, and testing all components.

The scripts use a shared PowerShell module (`Orchion.Common.psm1`) that provides common functions and reduces code duplication.

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
- Installs Go linting/formatting tools (golangci-lint, goimports) if missing
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

### `lint-all.ps1`
Runs linting for all Orchion components.

**Usage:**
```powershell
.\shared\scripts\lint-all.ps1
```

**What it does:**
- Runs golangci-lint for Go projects (orchestrator, node-agent, shared/logging)
- Runs ESLint for dashboard (Svelte/TypeScript)
- Runs ESLint for VSCode extension (TypeScript)
- Reports pass/fail for each component

**Prerequisites:**
- golangci-lint must be installed (`go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest`)
- Dashboard dependencies installed (auto-installs if missing)
- VSCode extension dependencies installed (auto-installs if missing)

**Note:** Dashboard linting failures are warnings only. VSCode extension and Go linting failures cause script to exit with error.

---

### `format-all.ps1`
Formats code for all Orchion components.

**Usage:**
```powershell
.\shared\scripts\format-all.ps1
```

**What it does:**
- Runs gofmt and goimports for Go projects (orchestrator, node-agent, shared/logging)
- Runs Prettier for dashboard (Svelte/TypeScript)
- Runs Prettier for VSCode extension (TypeScript)
- Modifies files in-place

**Prerequisites:**
- Dashboard dependencies installed (auto-installs if missing)
- VSCode extension dependencies installed (auto-installs if missing)

**Note:** This script modifies files. Make sure to commit/stash changes before running.

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

## 🚀 Common Workflows

### Initial Setup
```powershell
# 1. Setup everything (installs dependencies, generates protobuf, etc.)
.\shared\scripts\setup-all.ps1

# 2. Build all components
.\shared\scripts\build-all.ps1
```

**Note:** `setup-all.ps1` handles protobuf generation and dependency installation automatically.

**Integration Tests:** End-to-end tests are available in the `tests/` directory.

### Daily Development
```powershell
# Format and lint code
.\shared\scripts\format-all.ps1
.\shared\scripts\lint-all.ps1

# Build everything (Go + Dashboard)
.\shared\scripts\build-all.ps1

# Run everything (starts in separate windows)
.\shared\scripts\run-all.ps1

# Start dashboard separately (optional, or modify run-all.ps1 to include it)
.\shared\scripts\dev-dashboard.ps1

# Test everything
.\shared\scripts\test-all.ps1

# Test the API (integration tests in tests/ directory)
.\tests\test-api.ps1
.\tests\test-job.ps1
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

## 📝 Script Requirements

- **PowerShell 5.1+** (Windows 10+ default)
- **Go 1.21+** (for building)
- **protoc** (for `proto-gen.ps1`)
- **make** (optional, used by `proto-gen.ps1` for component-specific protobuf generation)

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

## 🔧 Shared Module (`Orchion.Common.psm1`)

The scripts use a shared PowerShell module that provides common functionality:

### Utility Functions
- `Write-Step`, `Write-Success`, `Write-Error`, `Write-Warning`, `Write-Info` - Consistent logging
- `Invoke-InDirectory` - Run commands in specific directories safely
- `Get-ProjectRoot` - Get the project root path

### Tool Checking
- `Test-GoInstalled`, `Test-NodeInstalled`, `Test-NpmInstalled` - Check required tools
- `Test-ProtocInstalled`, `Test-GolangciLintInstalled` - Check optional tools

### Component Management
- `Install-GoDependencies`, `Install-NodeDependencies` - Install dependencies
- `Build-GoComponent`, `Build-Dashboard` - Build components
- `Test-GoComponent`, `Test-NodeComponent` - Run tests
- `Lint-GoComponent`, `Lint-NodeComponent` - Run linters
- `Format-GoComponent`, `Format-NodeComponent` - Format code
- `Clean-GoComponent`, `Clean-Dashboard` - Clean build artifacts
- `Generate-Protobuf` - Generate protobuf files

### Configuration
- `$script:ProjectRoot` - Project root path
- `$script:Components` - Lists of Go and Node components

---

## 💡 Tips

- Keep scripts in sync with project structure - if you add new components, update these scripts
- All scripts use `$ErrorActionPreference = "Stop"` to fail fast on errors
- Scripts assume they're run from the project root directory
- The shared module makes it easy to add new components or modify behavior across all scripts
- Consider adding these scripts to your PATH or creating aliases

---

## 🔮 Future Script Ideas

Potential scripts to add:
- `setup-dev.ps1` - Complete development environment setup
- `deploy-local.ps1` - Deploy to local Docker/Kubernetes
- `watch-build.ps1` - Watch for changes and rebuild automatically
- `check-all.ps1` - Run lint + format check without modifying files
