# Orchion Tools - VS Code Extension

VS Code extension for interacting with Orchion directly from your editor.

---

## Overview

Orchion Tools provides developer-focused integration with the Orchion orchestrator. Currently in early development with a basic skeleton in place.

**Current Status:** ⏳ Extension structure exists. Core functionality planned.

**Planned Features:**
- Orchion Nodes tree view
- Job submission panel
- Log streaming
- Agent pipeline authoring
- Schema validation

---

## Architecture

```
orchion-tools/
├── src/
│   ├── extension.ts          # Extension entry point
│   ├── nodesView.ts          # Tree view provider (placeholder)
│   └── test/
│       └── extension.test.ts # Tests
├── package.json              # Extension manifest
├── tsconfig.json             # TypeScript configuration
└── esbuild.js                # Build configuration
```

---

## Development

### Prerequisites

- Node.js 18+ and npm
- VS Code (for testing)
- `vsce` (VS Code Extension Manager) - for packaging

### Install Dependencies

```powershell
npm install
```

### Build

```powershell
npm run compile
```

### Package Extension

```powershell
npm install -g @vscode/vsce
vsce package
```

Creates a `.vsix` file for distribution.

### Run Extension in Development

1. Open this folder in VS Code
2. Press F5 to launch Extension Development Host
3. Extension loads in new VS Code window

---

## Current Implementation

### Extension Activation

`src/extension.ts` - Registers the extension and initializes tree view.

### Tree View

`src/nodesView.ts` - Placeholder implementation. Currently shows hardcoded nodes:
```typescript
["node-1", "node-2"] // TODO: fetch from orchestrator
```

---

## Planned Features

### Phase 3 - Core Functionality

- **Node Tree View**
  - List all registered nodes from orchestrator
  - Show node status (active/stale)
  - Display node capabilities
  - Refresh button

- **Job Submission**
  - Form/panel for submitting jobs
  - Job template support
  - Schema validation

- **Log Streaming**
  - Stream logs from nodes
  - Filter by node/job
  - Color coding and formatting

- **Pipeline Authoring**
  - Visual pipeline editor
  - YAML/JSON support
  - Validation

### Configuration

Planned settings:
- Orchestrator API endpoint
- Auto-refresh interval
- Theme preferences

---

## Extension Manifest

Key settings in `package.json`:

```json
{
  "name": "orchion-tools",
  "displayName": "Orchion Tools",
  "contributes": {
    "views": {
      "explorer": [{
        "id": "orchionNodes",
        "name": "Orchion Nodes"
      }]
    }
  }
}
```

---

## Testing

### Running Tests

```powershell
npm test
```

### Manual Testing

1. Build extension: `npm run compile`
2. Press F5 to launch Extension Development Host
3. Test tree view and other features

---

## Troubleshooting

### Extension not loading

- Check VS Code Developer Console (Help → Toggle Developer Tools)
- Verify `package.json` is valid
- Check build output for errors

### Tree view not showing

- Verify view is registered in `package.json` under `contributes.views`
- Check activation events are configured correctly

---

## Building Against Orchestrator

**Future:** Extension will communicate with orchestrator via:
- HTTP REST API (for dashboard-like features)
- WebSocket (for real-time updates)
- Generated TypeScript types from protobuf

Currently extension is standalone - no orchestrator integration yet.

---

## Related Documentation

- **Project README:** `../../README.md`
- **Architecture:** `../../docs/architecture.md`
- **VS Code Extension API:** https://code.visualstudio.com/api
- **Extension Guidelines:** https://code.visualstudio.com/api/references/extension-guidelines

---

## Contributing

This extension is part of the Orchion monorepo. See project root for contribution guidelines.
