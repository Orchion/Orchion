# Orchion Quick Start Guide

**Status:** ✅ Minimal working system is ready for testing!

## What's Been Implemented

### ✅ Orchestrator
- gRPC server on port 50051 (default) with proper status codes
- HTTP REST API on port 8080 (default) with CORS support
- Node registration and heartbeat tracking
- Automatic heartbeat timeout monitoring with stale node cleanup
- Graceful shutdown handling
- Proper gRPC error handling (codes.InvalidArgument, codes.NotFound, etc.)

### ✅ Node Agent
- Automatic registration with orchestrator
- Comprehensive capability detection (CPU cores, system memory, OS, GPU type/VRAM, power usage)
- Periodic heartbeat sending
- Auto re-registration if orchestrator restarts
- Graceful shutdown

### ✅ Dashboard
- Node list UI with error handling
- Connects to REST API at `localhost:8080/api/nodes` (configurable via `VITE_ORCHESTRATOR_URL`)
- Displays node information (hostname, capabilities, last seen)
- TypeScript type safety

### ✅ Code Quality Tools

Orchion includes automated code formatting and linting tools to ensure consistent code quality:

- **Go**: `golangci-lint` (comprehensive linting) + `gofmt`/`goimports` (formatting)
- **TypeScript/JavaScript**: ESLint + Prettier

**Quick commands:**
```powershell
# Format all code
.\shared\scripts\format-all.ps1

# Lint all code
.\shared\scripts\lint-all.ps1
```

These tools use shared configuration files (`.golangci.yml`, `.prettierrc`, etc.) and are installed automatically by the setup script.

---

## How to Test

### Option A: Use Scripts (Easiest)

**Format and lint code (recommended):**
```powershell
.\shared\scripts\format-all.ps1
.\shared\scripts\lint-all.ps1
```

**Build everything:**
```powershell
.\shared\scripts\build-all.ps1
```

**Run everything:**
```powershell
.\shared\scripts\run-all.ps1
```

This starts orchestrator, node-agent, and dashboard. Press Ctrl+C in the dashboard window to stop all components.

### Option B: Manual Steps

### Step 1: Start the Orchestrator

```powershell
cd orchestrator
go run ./cmd/orchestrator
```

Or build and run:
```powershell
cd orchestrator
go build -o orchestrator.exe ./cmd/orchestrator
.\orchestrator.exe
```

**Expected output:**
```
Starting Orchion Orchestrator
gRPC port: 50051
HTTP port: 8080
Heartbeat timeout: 30s
gRPC server listening on :50051
HTTP REST API listening on :8080
```

### Step 2: Start a Node Agent (in another terminal)

```powershell
cd node-agent
go run ./cmd/node-agent
```

Or build and run:
```powershell
cd node-agent
go build -o bin/node-agent.exe ./cmd/node-agent
.\bin\node-agent.exe
```

**Expected output:**
```
Orchion Node Agent starting...
Generated node ID: <uuid>
Hostname: <your-hostname>
Capabilities: CPU=8 cores, Memory=16.00 GB, OS=windows/amd64
Connected to orchestrator at localhost:50051
Node registered successfully
Heartbeat loop started (interval: 5s)
Node agent running. Press Ctrl+C to stop.
```

### Step 3: Verify Registration

**Option A: Use curl/Invoke-WebRequest**
```powershell
Invoke-WebRequest -Uri http://localhost:8080/api/nodes | Select-Object -ExpandProperty Content
```

**Option B: View in Dashboard**
```powershell
.\shared\scripts\dev-dashboard.ps1
```

Or manually:
```powershell
cd dashboard
npm run dev
```

Then open `http://localhost:5173` (or whatever port SvelteKit assigns) and you should see your registered node.

**Option C: Check orchestrator logs**
The orchestrator will log when nodes send heartbeats (every 5 seconds by default).

---

## Command-Line Options

### Orchestrator

```powershell
.\bin\orchestrator.exe -port 50051 -http-port 8080 -heartbeat-timeout 30s
```

- `-port`: gRPC server port (default: 50051)
- `-http-port`: HTTP REST API port (default: 8080)
- `-heartbeat-timeout`: How long before a node is considered stale (default: 30s)

### Node Agent

```powershell
.\bin\node-agent.exe -orchestrator localhost:50051 -heartbeat-interval 5s -node-id my-node-1 -hostname my-machine
```

- `-orchestrator`: Orchestrator gRPC address (default: localhost:50051)
- `-heartbeat-interval`: How often to send heartbeats (default: 5s)
- `-node-id`: Custom node ID (auto-generated if not provided)
- `-hostname`: Custom hostname (uses system hostname if not provided)

---

## Testing Multiple Nodes

Start multiple node agents in different terminals:

**Terminal 1:**
```powershell
cd node-agent
go run ./cmd/node-agent -node-id node-1
```

**Terminal 2:**
```powershell
cd node-agent
go run ./cmd/node-agent -node-id node-2
```

**Terminal 3:**
```powershell
cd node-agent
go run ./cmd/node-agent -node-id node-3
```

All should register and send heartbeats. Check `/api/nodes` to see all registered nodes.

---

## What's Working

✅ Node registration via gRPC  
✅ Heartbeat tracking  
✅ Automatic capability detection  
✅ REST API for dashboard  
✅ Dashboard can display nodes  
✅ Graceful shutdown  
✅ Heartbeat timeout monitoring (logs stale nodes)  

---

## What's NOT Yet Implemented

❌ Job execution (node-agent executor.go is empty, but container management exists)  
❌ Job scheduling (orchestrator scheduler/router are not implemented)  
❌ Persistent storage (everything is in-memory)  
❌ Authentication/authorization  
❌ Health check endpoint  
❌ Node unregistration API  
❌ Advanced node status tracking  
❌ Dashboard auto-refresh/polling  
❌ TypeScript type generation from protobuf  

---

## Troubleshooting

### "Failed to connect to orchestrator"
- Make sure orchestrator is running first
- Check the `-orchestrator` address matches the orchestrator's gRPC port

### Dashboard shows no nodes
- Verify orchestrator is running on port 8080 (check logs)
- Check browser console for fetch errors
- Try accessing `http://localhost:8080/api/nodes` directly

### Nodes not showing in API
- Check orchestrator logs for registration messages
- Verify node agent is connected (check its logs)
- Ensure heartbeat interval is less than heartbeat timeout

### Build errors
```powershell
# Regenerate protobuf files if needed
.\shared\scripts\proto-gen.ps1

# Then try building again
.\shared\scripts\build-all.ps1
```

---

## Next Steps

1. **Test with real hardware** - Run on different machines
2. **Implement job execution** - Fill in `node-agent/internal/executor/executor.go`
3. **Add job scheduling** - Implement scheduler in orchestrator
4. **Add persistence** - Store node state in database
5. **Add authentication** - Secure the API

---

## Related Documentation

- **Development Setup:** `development-setup.md`
- **Project Status:** `project-status.md`
- **Architecture:** `architecture.md`

---

**Happy testing!** 🚀
