<script lang="ts">
	import { onMount, onDestroy } from 'svelte';

	interface LogEntry {
		id: string;
		timestamp: number;
		level: 'info' | 'warning' | 'error';
		source: string;
		message: string;
		fields?: Record<string, string>;
	}

	let logs: LogEntry[] = [];
	let error: string | null = null;
	let eventSource: EventSource | null = null;

	onMount(() => {
		// Connect to Server-Sent Events for real-time logs
		connectToLogStream();
	});

	onDestroy(() => {
		if (eventSource) {
			eventSource.close();
		}
	});

	function connectToLogStream() {
		try {
			eventSource = new EventSource('/api/logs');

			eventSource.onmessage = (event) => {
				try {
					const data = JSON.parse(event.data);
					if (data.type === 'log') {
						logs = [data.entry, ...logs].slice(0, 100); // Keep last 100 logs
					}
					// For now, we only have keepalive messages
				} catch (err) {
					console.error('Failed to parse log event:', err);
				}
			};

			eventSource.onerror = (err) => {
				console.error('EventSource error:', err);
				error = 'Failed to connect to log stream';
			};
		} catch (err) {
			console.error('Failed to connect to log stream:', err);
			error = 'Failed to connect to log stream';
		}
	}
</script>

<h1>Orchion Dashboard</h1>

<nav>
	<a href="/">Nodes</a> |
	<a href="/logs">Logs</a>
</nav>

<h2>Logs</h2>

{#if error}
	<p style="color: red;">Error: {error}</p>
{:else if logs.length === 0}
	<p>No logs received yet. Logs will appear here in real-time.</p>
{:else}
	<div style="max-height: 600px; overflow-y: auto; border: 1px solid #ccc; padding: 10px;">
		{#each logs as log}
			<div
				style="margin-bottom: 8px; padding: 4px; border-left: 4px solid {log.level ===
				'error'
					? 'red'
					: log.level === 'warning'
						? 'orange'
						: 'blue'};"
			>
				<strong
					>[{new Date(log.timestamp).toLocaleTimeString()}] {log.level.toUpperCase()}</strong
				>
				<span style="color: #666;">{log.source}</span>
				<br />
				{log.message}
				{#if log.fields && Object.keys(log.fields).length > 0}
					<br />
					<small style="color: #888;">
						{#each Object.entries(log.fields) as [key, value]}
							{key}={value}
						{/each}
					</small>
				{/if}
			</div>
		{/each}
	</div>
{/if}

<p><a href="/">← Back to Nodes</a></p>
