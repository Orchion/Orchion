/**
 * HTTP client for communicating with the Orchion orchestrator
 */

export interface Node {
	id: string;
	hostname: string;
	capabilities?: {
		cpu?: string;
		memory?: string;
		os?: string;
	};
	last_seen_unix?: number;
}

export interface ChatMessage {
	role: 'system' | 'user' | 'assistant';
	content: string;
}

export interface ChatCompletionRequest {
	model: string;
	messages: ChatMessage[];
	temperature?: number;
	stream?: boolean;
	max_tokens?: number;
}

export interface ChatChoice {
	index: number;
	message: ChatMessage;
	finish_reason?: string;
}

export interface ChatCompletionResponse {
	id: string;
	model: string;
	choices: ChatChoice[];
	created: number;
	object: string;
}

export interface LogEntry {
	id: string;
	timestamp: number;
	level: 'debug' | 'info' | 'warn' | 'error';
	source: string;
	message: string;
	fields?: Record<string, string>;
}

export class OrchestratorClient {
	private baseUrl: string;

	constructor(baseUrl: string = 'http://localhost:8080') {
		this.baseUrl = baseUrl.replace(/\/$/, ''); // Remove trailing slash
	}

	/**
	 * List all registered nodes
	 */
	async listNodes(): Promise<Node[]> {
		const response = await fetch(`${this.baseUrl}/api/nodes`);
		if (!response.ok) {
			throw new Error(`Failed to list nodes: ${response.statusText}`);
		}
		return await response.json();
	}

	/**
	 * List all available agents/models
	 * Note: This requires the orchestrator to expose an agents endpoint.
	 * Currently returns placeholder data until the endpoint is available.
	 */
	async listAgents(): Promise<
		Array<{ id: string; name: string; model: string; description?: string }>
	> {
		// TODO: Replace with actual API call when orchestrator exposes /api/agents
		// For now, return placeholder based on common models
		try {
			const nodes = await this.listNodes();
			if (nodes.length === 0) {
				return [];
			}
			// Placeholder: return default agents
			// In production, this should call GET /api/agents
			return [
				{
					id: 'llama3.2',
					name: 'Llama 3.2',
					model: 'llama3.2',
					description: 'Llama 3.2 model',
				},
				{
					id: 'llama3.1',
					name: 'Llama 3.1',
					model: 'llama3.1',
					description: 'Llama 3.1 model',
				},
			];
		} catch {
			return [];
		}
	}

	/**
	 * Send a prompt to a specific agent/model
	 * Convenience wrapper around chatCompletion
	 */
	async sendPrompt(
		agent: string,
		message: string
	): Promise<AsyncIterable<ChatCompletionResponse>> {
		return this.chatCompletion({
			model: agent,
			messages: [{ role: 'user', content: message }],
			stream: true,
		});
	}

	/**
	 * Send a chat completion request (streaming)
	 * Returns an async iterator of response chunks
	 *
	 * Note: This requires the orchestrator to expose HTTP endpoints for chat completions.
	 * Currently, the orchestrator only exposes gRPC for chat. HTTP endpoints need to be added.
	 */
	async *chatCompletion(request: ChatCompletionRequest): AsyncIterable<ChatCompletionResponse> {
		const response = await fetch(`${this.baseUrl}/api/chat/completions`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({
				...request,
				stream: true,
			}),
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Chat completion failed: ${response.statusText} - ${errorText}`);
		}

		if (!response.body) {
			throw new Error('Response body is null');
		}

		const reader = response.body.getReader();
		const decoder = new TextDecoder();
		let buffer = '';

		try {
			while (true) {
				const { done, value } = await reader.read();
				if (done) break;

				buffer += decoder.decode(value, { stream: true });
				const lines = buffer.split('\n');
				buffer = lines.pop() || ''; // Keep incomplete line in buffer

				for (const line of lines) {
					const trimmed = line.trim();
					if (!trimmed || trimmed === 'data: [DONE]') continue;

					if (trimmed.startsWith('data: ')) {
						try {
							const json = JSON.parse(trimmed.slice(6));
							yield json;
						} catch (e) {
							console.error('Failed to parse SSE data:', e, trimmed);
						}
					}
				}
			}
		} finally {
			reader.releaseLock();
		}
	}

	/**
	 * Send a non-streaming chat completion request
	 */
	async chatCompletionNonStreaming(
		request: ChatCompletionRequest
	): Promise<ChatCompletionResponse> {
		const response = await fetch(`${this.baseUrl}/api/chat/completions`, {
			method: 'POST',
			headers: {
				'Content-Type': 'application/json',
			},
			body: JSON.stringify({
				...request,
				stream: false,
			}),
		});

		if (!response.ok) {
			const errorText = await response.text();
			throw new Error(`Chat completion failed: ${response.statusText} - ${errorText}`);
		}

		return await response.json();
	}

	/**
	 * Get logs from the orchestrator
	 * Note: This requires the orchestrator to expose a logs endpoint.
	 * Currently returns placeholder data until the endpoint is available.
	 */
	async getLogs(): Promise<LogEntry[]> {
		// TODO: Replace with actual API call when orchestrator exposes /api/logs
		// For now, return placeholder logs
		try {
			const nodes = await this.listNodes();
			const placeholderLogs: LogEntry[] = [];

			// Generate some placeholder logs based on nodes
			for (const node of nodes) {
				placeholderLogs.push({
					id: `log-${Date.now()}-${Math.random()}`,
					timestamp: Date.now() - Math.random() * 3600000, // Random time in last hour
					level: 'info',
					source: `node-agent:${node.id}`,
					message: `Node ${node.hostname || node.id} registered successfully`,
					fields: {
						node_id: node.id,
						hostname: node.hostname,
					},
				});
			}

			// Add some orchestrator logs
			placeholderLogs.push({
				id: `log-${Date.now()}-${Math.random()}`,
				timestamp: Date.now() - 300000, // 5 minutes ago
				level: 'info',
				source: 'orchestrator',
				message: 'Orchestrator started successfully',
				fields: {
					version: '0.1.0',
				},
			});

			return placeholderLogs.sort((a, b) => b.timestamp - a.timestamp);
		} catch {
			return [];
		}
	}

	/**
	 * Test connection to orchestrator
	 */
	async ping(): Promise<boolean> {
		try {
			await this.listNodes();
			return true;
		} catch {
			return false;
		}
	}
}
