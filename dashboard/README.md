# Dashboard

Web-based UI for monitoring and managing Orchion clusters. Built with SvelteKit.

---

## Overview

The Orchion dashboard provides:

- Real-time node list with health and capabilities
- Cluster overview and status
- (Future) Job queue and execution history
- (Future) Log viewing and streaming
- (Future) Agent pipeline management

**Current Status:** ✅ Basic node list implemented. Advanced features planned.

**Tech Stack:**

- **SvelteKit** - Full-stack Svelte framework
- **TypeScript** - Type safety
- **Vite** - Build tool and dev server

---

## Architecture

```
dashboard/
├── src/
│   ├── routes/               # SvelteKit routes (file-based routing)
│   │   ├── +page.svelte      # Main page (node list)
│   │   └── +layout.svelte    # Layout wrapper
│   └── lib/
│       └── orchion.ts        # HTTP client for orchestrator API
├── static/                   # Static assets
├── package.json              # Dependencies and scripts
└── vite.config.ts            # Vite configuration
```

---

## Development

### Prerequisites

- Node.js 18+ and npm
- Orchestrator running (for API access)

### Install Dependencies

```powershell
npm install
```

### Run Development Server

**Option 1: Using the script (recommended)**

```powershell
.\shared\scripts\dev-dashboard.ps1
```

**Option 2: Manual**

```powershell
npm run dev
```

Opens at `http://localhost:5173` (or next available port).

### Full Stack Development

**Terminal 1: Orchestrator**

```powershell
cd orchestrator
.\orchestrator.exe
```

**Terminal 2: Node Agent**

```powershell
cd node-agent
.\node-agent.exe
```

**Terminal 3: Dashboard**

```powershell
cd dashboard
npm run dev
```

Then open `http://localhost:5173` in your browser.

### Build for Production

```powershell
npm run build
```

Output goes to `build/` directory.

### Preview Production Build

```powershell
npm run preview
```

---

## Configuration

### API Endpoint

The dashboard connects to the orchestrator's HTTP REST API. Default endpoint:

```typescript
// src/lib/orchion.ts
const API_URL = 'http://localhost:8080/api/nodes';
```

To change the endpoint, edit `src/lib/orchion.ts`.

**Future:** Configuration via environment variables or config file.

---

## API Integration

### Current Endpoints Used

- **`GET /api/nodes`** - List all registered nodes

### Expected Response Format

```json
[
	{
		"id": "node-uuid",
		"hostname": "my-server",
		"capabilities": {
			"cpu": "8 cores",
			"memory": "16.00 GB",
			"os": "windows/amd64"
		},
		"lastSeenUnix": 1234567890
	}
]
```

---

## Project Structure

### Routes (`src/routes/`)

SvelteKit uses file-based routing:

- `+page.svelte` - Main page component
- `+layout.svelte` - Layout wrapper (shared across pages)

### Library (`src/lib/`)

Shared utilities and API clients:

- `orchion.ts` - HTTP client for orchestrator REST API

### Static Assets (`static/`)

Files served directly (not processed):

- `robots.txt` - Search engine directives
- Favicons and images (when added)

---

## Development Scripts

```json
{
	"dev": "vite dev", // Start dev server
	"build": "vite build", // Build for production
	"preview": "vite preview", // Preview production build
	"check": "svelte-check", // Type checking
	"lint": "eslint ." // Linting
}
```

---

## Implementation Patterns

### Adding Auto-Refresh (Polling)

```svelte
<script>
	import { onMount, onDestroy } from 'svelte';
	import { getNodes } from '$lib/orchion';

	let nodes = [];
	let interval;

	async function refreshNodes() {
		nodes = await getNodes();
	}

	onMount(() => {
		refreshNodes();
		interval = setInterval(refreshNodes, 5000); // Every 5 seconds
	});

	onDestroy(() => {
		clearInterval(interval);
	});
</script>
```

### Error Handling Pattern

```svelte
<script>
	let nodes = [];
	let error = null;
	let loading = true;

	onMount(async () => {
		try {
			nodes = await getNodes();
		} catch (e) {
			error = 'Failed to fetch nodes. Is orchestrator running?';
			console.error(e);
		} finally {
			loading = false;
		}
	});
</script>

{#if loading}
	<p>Loading nodes...</p>
{:else if error}
	<p class="error">{error}</p>
{:else if nodes.length === 0}
	<p>No nodes registered yet.</p>
{:else}
	<!-- Display nodes -->
{/if}
```

### Node Status Calculation

Calculate status based on `lastSeenUnix`:

```svelte
<script>
	function getNodeStatus(node) {
		const lastSeen = node.lastSeenUnix * 1000;
		const now = Date.now();
		const age = now - lastSeen;

		if (age < 10000) return 'active'; // < 10 seconds
		if (age < 30000) return 'stale'; // < 30 seconds
		return 'offline'; // > 30 seconds
	}
</script>

{#each nodes as node}
	<div class="node-card status-{getNodeStatus(node)}">
		<!-- ... -->
	</div>
{/each}
```

---

## Planned Features

### Phase 2 - Core Features

- [ ] Node detail view (drill-down into individual nodes)
- [ ] Real-time updates (WebSocket or polling)
- [ ] Node status indicators (active/stale/offline)
- [ ] Capability filtering and sorting

### Phase 3 - Advanced Features

- [ ] Job queue view
- [ ] Job execution history
- [ ] Log viewer with streaming
- [ ] Cluster health dashboard
- [ ] Metrics and charts

### Phase 4 - Management Features

- [ ] Job submission interface
- [ ] Agent pipeline authoring
- [ ] Configuration management
- [ ] User authentication (when added to orchestrator)

---

## Troubleshooting

### "Failed to fetch" errors

- Verify orchestrator is running: `Invoke-RestMethod http://localhost:8080/api/nodes`
- Check API endpoint in `src/lib/orchion.ts` matches orchestrator's HTTP port
- Verify CORS is enabled if accessing from different origin

### No nodes displayed

- Check browser console for errors
- Verify orchestrator has registered nodes
- Test API directly: `http://localhost:8080/api/nodes`

### Build errors

```powershell
# Clean and reinstall
Remove-Item -Recurse -Force node_modules
Remove-Item package-lock.json
npm install
```

---

## Styling

Tailwind CSS is configured but minimally used.

**Add Tailwind classes to components:**

```svelte
<div class="container mx-auto p-4">
	<h1 class="text-2xl font-bold">Orchion Nodes</h1>
	<!-- ... -->
</div>
```

**Create component styles in `layout.css`:**

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* Custom styles */
.node-card {
	@apply mb-4 rounded border p-4;
}
```

**Future plans:**

- Component library integration
- Responsive design improvements
- Status indicator styling

---

## Related Documentation

- **Project README:** `../README.md`
- **Quick Start:** `../docs/quick-start.md`
- **Architecture:** `../docs/architecture.md`
- **Orchestrator README:** `../orchestrator/README.md` (for API details)
- **SvelteKit Docs:** https://kit.svelte.dev/
