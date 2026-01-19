# Orchion Development Setup & Troubleshooting Guide

A practical guide to setting up the Orchion development environment across Windows, VS Code, Go, Protocol Buffers, and Make. This document captures the real-world issues encountered during setup and the fixes that worked.

---

## 🚀 Quick Commands Reference

**TL;DR:** Use these commands to manage the entire Orchion monorepo.

### First Time Setup
```powershell
.\shared\scripts\setup-all.ps1
```

### Format Code (before committing)
```powershell
.\shared\scripts\format-all.ps1
```

### Lint Code (check code quality)
```powershell
.\shared\scripts\lint-all.ps1
```

### Build Everything
```powershell
.\shared\scripts\build-all.ps1
```

### Run Everything (starts in separate windows)
```powershell
.\shared\scripts\run-all.ps1
```

### Generate Protobuf Files
```powershell
.\shared\scripts\proto-gen.ps1
```

### Test the REST API
```powershell
.\tests\test-api.ps1
```

### Clean Build Artifacts
```powershell
.\shared\scripts\clean-all.ps1
```

**Tip:** All scripts are in `shared/scripts/` - see `shared/scripts/README.md` for full documentation.

---

## 📋 Prerequisites

Before setting up Orchion, ensure you have:

- **Windows 10/11** (this guide is Windows-focused)
- **PowerShell** (comes with Windows)
- **VS Code** (recommended IDE)
- **Chocolatey** (for easy package management)

---

## 🎯 Code Quality Tools

Orchion uses automated tools to ensure consistent code formatting and quality across all components. These tools are installed automatically by the setup script.

### Code Formatting
- **Go**: `gofmt` + `goimports` (installed with Go)
- **TypeScript/JavaScript**: Prettier (installed via npm)

### Code Linting
- **Go**: `golangci-lint` (comprehensive linter with 30+ rules including security checks)
- **TypeScript/JavaScript**: ESLint + Prettier integration

### Configuration Files
The project includes shared configuration files at the root level:
- `.prettierrc` - Prettier formatting rules
- `.prettierignore` - Files to exclude from formatting
- `.golangci.yml` - Go linting configuration
- `.eslintignore` - Files to exclude from ESLint
- `.editorconfig` - Editor settings for consistent indentation

**Note:** Individual components may have their own overrides (e.g., dashboard has Svelte-specific Prettier config).

### Automated Scripts
Use these scripts to format and lint the entire codebase:
```powershell
# Format all code (modifies files)
.\shared\scripts\format-all.ps1

# Lint all code (shows warnings, doesn't fail)
.\shared\scripts\lint-all.ps1
```

See `shared/scripts/README.md` for detailed script documentation.

---

## 🔧 Installing Required Tools

### 1. Install Go

**Option A: Using Chocolatey** (Recommended)
```powershell
choco install golang
```

**Option B: Manual Installation**
1. Download from https://go.dev/dl/
2. Run installer
3. Verify: `go version`

**Verify Installation:**
```powershell
go version
```

### 2. Install Protocol Buffers Compiler (protoc)

**Option A: Using Chocolatey**
```powershell
choco install protoc
```

**Option B: Manual Installation**
1. Download from https://github.com/protocolbuffers/protobuf/releases/latest
   - Download: `protoc-<version>-win64.zip`
2. Extract to `C:\tools\protoc\`
3. Add to PATH: `C:\tools\protoc\bin`
4. Restart terminal

**Verify Installation:**
```powershell
protoc --version
```

### 3. Install Make (Optional, for Component Makefiles)

**Using Chocolatey:**
```powershell
choco install make
```

**Verify Installation:**
```powershell
make --version
```

**Note:** Component Makefiles are used for protobuf generation. PowerShell scripts will fall back to direct `protoc` commands if `make` is not available.

### 4. Install Go Protobuf Plugins

```powershell
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

**Add Go bin to PATH:**
1. Open System Environment Variables
2. Add: `C:\Users\<your-username>\go\bin`
3. Or in PowerShell: `$env:Path += ";$env:USERPROFILE\go\bin"`

**Verify Installation:**
```powershell
protoc-gen-go --version
protoc-gen-go-grpc --version
```

### 5. Install Node.js and npm (for Dashboard)

**Using Chocolatey:**
```powershell
choco install nodejs
```

**Or download from:**
https://nodejs.org/

**Verify Installation:**
```powershell
node --version
npm --version
```

### 6. Install Container Runtime (Podman preferred)

**Why Podman?** Podman is a daemonless container engine that doesn't require root privileges and is more secure than Docker.

**Option A: Install Podman (Recommended)**

**Using Chocolatey:**
```powershell
choco install podman
```

**Manual Installation:**
1. Download from https://github.com/containers/podman/releases/latest
2. Install and configure
3. Initialize: `podman machine init` (on Windows)
4. Start: `podman machine start`

**Option B: Install Docker (Alternative)**

**Using Chocolatey:**
```powershell
choco install docker-desktop
```

**Or download from:**
https://www.docker.com/products/docker-desktop/

**Verify Container Runtime:**
```powershell
# Try Podman first (preferred)
podman version

# If Podman not available, try Docker
docker version
```

**Note:** Orchion will automatically detect and use the available container runtime (Podman first, then Docker).

---

## 📁 Project Structure

```text
Orchion/
├── orchestrator/
│   ├── cmd/orchestrator/main.go
│   ├── internal/
│   │   ├── node/registry.go
│   │   └── orchestrator/service.go
│   ├── api/v1/v1/          # Generated protobuf files appear here
│   ├── Makefile
│   └── go.mod
│
├── node-agent/
│   ├── cmd/node-agent/main.go
│   ├── internal/
│   │   ├── capabilities/capabilities.go
│   │   ├── heartbeat/heartbeat.go
│   │   └── proto/v1/       # Generated protobuf files appear here
│   ├── Makefile
│   └── go.mod
│
└── shared/
    └── proto/v1/
        └── orchestrator.proto
```

---

## 🚀 Initial Setup Steps

### 1. Clone the Repository (if not already done)

```powershell
git clone <repository-url>
cd Orchion
```

### 2. Run Setup Script (Recommended)

**Easiest way - runs everything automatically:**
```powershell
.\shared\scripts\setup-all.ps1
```

This script will:
- Check prerequisites (Go, Node.js, npm, protoc, Podman/Docker)
- Install Go protobuf plugins if missing
- Install Go linting/formatting tools (`golangci-lint`, `goimports`) if missing
- Install Go dependencies for orchestrator and node-agent
- Generate protobuf files
- Install dashboard npm dependencies
- Install Playwright browsers for testing

**After setup, you can skip to step 4 (Build).**

See `shared/scripts/README.md` for details on all available scripts.

### 3. Manual Setup (Alternative)

If you prefer manual setup or the script fails:

#### 3a. Generate Protobuf Files

**For Orchestrator:**
```powershell
cd orchestrator
make proto
```

**For Node Agent:**
```powershell
cd node-agent
make proto
```

**Manual alternative (if make doesn't work):**
```powershell
cd orchestrator
protoc -I ../shared/proto `
    --go_out=api/v1 --go_opt=paths=source_relative `
    --go-grpc_out=api/v1 --go-grpc_opt=paths=source_relative `
    ../shared/proto/v1/orchestrator.proto
```

#### 3b. Install Go Dependencies

**Orchestrator:**
```powershell
cd orchestrator
go mod tidy
```

**Node Agent:**
```powershell
cd node-agent
go mod tidy
```

#### 3c. Install Dashboard Dependencies

```powershell
cd dashboard
npm install
npx playwright install --with-deps chromium
```

### 4. Build Everything

**Using scripts (recommended):**
```powershell
.\shared\scripts\build-all.ps1
```

**Or manually:**
```powershell
cd orchestrator
go build -o orchestrator.exe ./cmd/orchestrator

cd ../node-agent
go build -o node-agent.exe ./cmd/node-agent
```

---

## 🛠️ VS Code Configuration

### Component Makefile Configuration

Component Makefiles (in `orchestrator/` and `node-agent/`) require **literal TAB characters** and **LF** line endings.

Add to VS Code `settings.json`:
```json
{
  "[makefile]": {
    "editor.insertSpaces": false,
    "editor.detectIndentation": false,
    "files.eol": "\n"
  }
}
```

**To fix existing Makefiles:**
1. Open Makefile
2. Bottom right: Change line endings from `CRLF` to `LF`
3. Convert spaces to tabs: **Convert Indentation to Tabs** (right-click in editor)

### Recommended VS Code Extensions

- **Go** (by Google) - Go language support
- **Protocol Buffers** (Peter Suter) - `.proto` file syntax highlighting
- **Svelte for VS Code** (Svelte) - SvelteKit support
- **Make** (ms-vscode.makefile-tools) - Makefile support
- **Prettier** (Prettier) - Code formatting (uses project config)
- **ESLint** (Microsoft) - JavaScript/TypeScript linting
- **EditorConfig** (EditorConfig) - Consistent editor settings

---

## 🧪 Common Errors & Fixes

### ❌ `missing separator` (Makefile)

**Cause:** Spaces instead of tabs in Makefile  
**Fix:** Replace indentation with literal TAB characters

---

### ❌ `File does not reside within any path specified using --proto_path`

**Cause:** protoc cannot find the `.proto` file  
**Fix:** Use `-I ../shared/proto` (or correct relative path)

---

### ❌ `'protoc-gen-go' is not recognized`

**Cause:** Go plugins not installed or not on PATH  
**Fix:**
1. Install plugins: `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest`
2. Add `%USERPROFILE%\go\bin` to PATH
3. Restart terminal

---

### ❌ No `.pb.go` files appear after `make proto`

**Causes:**
- Output path not what you expected
- Files generated in nested directory (because of `package orchion.v1`)

**Find generated files:**
```powershell
Get-ChildItem -Recurse -Filter *.pb.go
```

**Expected locations:**
- Orchestrator: `orchestrator/api/v1/v1/*.pb.go`
- Node Agent: `node-agent/internal/proto/v1/*.pb.go`

**Note:** The nested `v1/v1/` structure is correct! It's because the proto package is `orchion.v1`, which creates a `v1/` subdirectory.

---

### ❌ `'make' is not recognized`

**Cause:** Make not installed on Windows  
**Fix:** `choco install make`

Or run protoc commands directly (see "Manual alternative" above).

---

### ❌ Build errors: `cannot find package`

**Cause:** Go modules not initialized or dependencies missing  
**Fix:**
```powershell
cd orchestrator  # or node-agent
go mod tidy
go mod download
```

---

### ❌ `'golangci-lint' is not recognized`

**Cause:** Go linting tool not installed or not on PATH
**Fix:**
```powershell
go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
```
Or run the setup script: `.\shared\scripts\setup-all.ps1`

---

### ❌ `'goimports' is not recognized`

**Cause:** Go import formatting tool not installed or not on PATH
**Fix:**
```powershell
go install golang.org/x/tools/cmd/goimports@latest
```
Or run the setup script: `.\shared\scripts\setup-all.ps1`

---

### ❌ Prettier/ESLint errors in VS Code

**Cause:** Extension not using project configuration
**Fix:**
1. Install Prettier and ESLint extensions
2. VS Code should auto-detect `.prettierrc` and `eslint.config.*` files
3. If not working, check VS Code settings:
   - Ensure "Format on Save" is enabled
   - Check "Prettier: Config Path" points to project root

---

### ❌ `format-all.ps1` fails with permission errors

**Cause:** Files are read-only or VS Code has them locked
**Fix:**
1. Close VS Code
2. Ensure no processes are using the files
3. Run the script again

---

### ❌ Dashboard: `npm install` fails

**Causes:**
- Node.js not installed
- Network/proxy issues
- Corrupted `node_modules`

**Fix:**
```powershell
cd dashboard
Remove-Item -Recurse -Force node_modules
Remove-Item package-lock.json
npm install
```

---

## 🧭 Quick Verification Checklist

After setup, verify everything works:

- [ ] `go version` shows Go 1.21 or later
- [ ] `protoc --version` shows protoc installed
- [ ] `make --version` shows make installed
- [ ] `protoc-gen-go --version` works
- [ ] `protoc-gen-go-grpc --version` works
- [ ] `golangci-lint --version` works (Go linting)
- [ ] `goimports -h` works (Go formatting)
- [ ] `make proto` works in both `orchestrator/` and `node-agent/`
- [ ] `go build ./cmd/orchestrator` works in `orchestrator/`
- [ ] `go build ./cmd/node-agent` works in `node-agent/`
- [ ] `npm install` works in `dashboard/`
- [ ] `.\shared\scripts\lint-all.ps1` runs without critical errors
- [ ] `.\shared\scripts\format-all.ps1` completes successfully

---

## 📄 Example Component Makefiles

**Orchestrator (`orchestrator/Makefile`):**
```makefile
proto:
	protoc -I ../shared/proto \
		--go_out=api/v1 --go_opt=paths=source_relative \
		--go-grpc_out=api/v1 --go-grpc_opt=paths=source_relative \
		../shared/proto/v1/orchestrator.proto
```

**Node Agent (`node-agent/Makefile`):**
```makefile
proto:
	protoc -I ../shared/proto \
		--go_out=internal/proto --go_opt=paths=source_relative \
		--go-grpc_out=internal/proto --go-grpc_opt=paths=source_relative \
		../shared/proto/v1/orchestrator.proto
```

**Important:** These must use **TAB characters** for indentation, not spaces!

---

## 🏃 Running the Project

See `quick-start.md` for detailed instructions on running orchestrator, node-agent, and dashboard.

**Quick test using scripts:**
```powershell
# Build everything first
.\shared\scripts\build-all.ps1

# Run everything (starts in separate windows)
.\shared\scripts\run-all.ps1
```

**Or manually:**
```powershell
# Terminal 1: Start orchestrator
cd orchestrator
go run ./cmd/orchestrator

# Terminal 2: Start node agent
cd node-agent
go run ./cmd/node-agent

# Terminal 3: Start dashboard
.\shared\scripts\dev-dashboard.ps1
# Or: cd dashboard && npm run dev
```

**Available scripts:**
- `.\shared\scripts\setup-all.ps1` - Initial setup (install dependencies, tools)
- `.\shared\scripts\format-all.ps1` - Format all code
- `.\shared\scripts\lint-all.ps1` - Lint all code
- `.\shared\scripts\build-all.ps1` - Build all components
- `.\shared\scripts\run-all.ps1` - Run all components
- `.\shared\scripts\dev-dashboard.ps1` - Start dashboard dev server
- `.\shared\scripts\test-all.ps1` - Run all tests
- `.\tests\test-api.ps1` - Test REST API
- `.\shared\scripts\clean-all.ps1` - Clean build artifacts

See `shared/scripts/README.md` for complete script documentation.

---

## 🔄 Typical Development Workflows

### First Time Setup
```powershell
# 1. Setup everything (installs all dependencies, generates protobuf, etc.)
.\shared\scripts\setup-all.ps1

# 2. Build everything
.\shared\scripts\build-all.ps1
```

### Daily Development
```powershell
# Code quality (recommended before committing)
.\shared\scripts\format-all.ps1
.\shared\scripts\lint-all.ps1

# Build
.\shared\scripts\build-all.ps1

# Run
.\shared\scripts\run-all.ps1

# Test
.\tests\test-api.ps1
```

### After Changing Protobuf
```powershell
# Regenerate
.\shared\scripts\proto-gen.ps1

# Rebuild
.\shared\scripts\build-all.ps1
```

---

## 📚 Additional Resources

- **Protocol Buffers:** https://protobuf.dev/
- **gRPC Go:** https://grpc.io/docs/languages/go/
- **Go Modules:** https://go.dev/ref/mod
- **SvelteKit:** https://kit.svelte.dev/

---

## 🐛 Getting Help

If you encounter issues not covered here:

1. Check `docs/quick-start.md` for testing and runtime verification
2. Check `project-status.md` for project status and next steps
3. Check build errors carefully - they usually indicate missing tools or PATH issues
4. Verify all prerequisites are installed and on PATH
5. Restart terminal/VS Code after PATH changes

---

**Happy coding!** 🚀
