# Orchion Development Setup & Troubleshooting Guide

A practical guide to setting up the Orchion development environment across Windows, VS Code, Go, and Protocol Buffers. This document captures the real-world issues encountered during setup and the fixes that worked.

---

## 📁 Project Structure

```text
Orchion/
├── orchestrator/
│   ├── Makefile
│   └── api/                 # generated Go files appear here
├── shared/
│   └── proto/
│       └── v1/
│           └── orchestrator.proto
```

---

## ⚙️ Installing protoc on Windows

### Download
Official releases:  
https://github.com/protocolbuffers/protobuf/releases/latest

Download:
```text
protoc-<version>-win64.zip
```

Extract to:
```text
C:\tools\protoc\
```

Ensure the folder contains:
```text
C:\tools\protoc\bin\protoc.exe
C:\tools\protoc\include\google\protobuf\*.proto
```

### Add to PATH
```text
C:\tools\protoc\bin
```

Restart your terminal.

### Verify
```bash
protoc --version
```

---

## ⚙️ Installing Go protobuf plugins

Install:
```bash
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
```

Add to PATH:
```text
C:\Users\<you>\go\bin
```

Verify:
```bash
protoc-gen-go --version
protoc-gen-go-grpc --version
```

---

## 🛠️ VS Code Makefile Configuration

Makefiles require **literal TAB characters** and **LF** line endings.

Add to `settings.json`:
```json
"[makefile]": {
  "editor.insertSpaces": false,
  "editor.detectIndentation": false
}
```

- Convert indentation: **Convert Indentation to Tabs**
- Ensure line endings are **LF**, not CRLF

---

## 🧪 Common protoc Errors & Fixes

### ❌ `missing separator`
**Cause:** spaces instead of tabs  
**Fix:** replace indentation with a real TAB

---

### ❌ `File does not reside within any path specified using --proto_path`
**Cause:** protoc cannot find the `.proto` file  
**Fix:**
```bash
protoc -I ../shared/proto ...
```

---

### ❌ `'protoc-gen-go' is not recognized`
**Cause:** Go plugins not installed or not on PATH  
**Fix:** install plugins and add `%USERPROFILE%\go\bin` to PATH

---

### ❌ No `.pb.go` files appear
**Causes:**
- Hidden by `.gitignore` or VS Code excludes
- Output path not what you expected

Find generated files:
```powershell
Get-ChildItem -Recurse -Filter *.pb.go
```

Generate into `orchestrator/api/v1`:
```text
--go_out=api --go_opt=paths=source_relative
--go-grpc_out=api --go-grpc_opt=paths=source_relative
```

---

## 📄 Example Working Makefile

```makefile
proto:
	protoc -I ../shared/proto \
		--go_out=api --go_opt=paths=source_relative \
		--go-grpc_out=api --go-grpc_opt=paths=source_relative \
		../shared/proto/v1/orchestrator.proto
```

---

## 📁 Example `orchestrator.proto`

```proto
syntax = "proto3";

package orchion.v1;

option go_package = "github.com/Orchion/Orchion/shared/proto/v1;v1";

message Capabilities {
  string cpu = 1;
  string memory = 2;
  string os = 3;
}

message Node {
  string id = 1;
  string hostname = 2;
  Capabilities capabilities = 3;
  int64 last_seen_unix = 4;
}

message RegisterNodeRequest { Node node = 1; }
message RegisterNodeResponse {}

message HeartbeatRequest { string node_id = 1; }
message HeartbeatResponse {}

message ListNodesRequest {}
message ListNodesResponse { repeated Node nodes = 1; }

service Orchestrator {
  rpc RegisterNode(RegisterNodeRequest) returns (RegisterNodeResponse);
  rpc Heartbeat(HeartbeatRequest) returns (HeartbeatResponse);
  rpc ListNodes(ListNodesRequest) returns (ListNodesResponse);
}
```

---

## 🧭 Quick Troubleshooting Checklist

- [ ] `protoc` installed and on PATH
- [ ] Go plugins installed and on PATH
- [ ] `.proto` file exists where the Makefile expects
- [ ] Makefile uses **tabs**, not spaces
- [ ] Line endings are **LF**
- [ ] Generated files not hidden by `.gitignore` / VS Code
- [ ] Terminal restarted after PATH changes
