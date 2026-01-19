import * as vscode from 'vscode';
import { OrchestratorClient, ChatMessage } from '../../api/orchestratorClient';
import { SettingsManager } from '../../state/settings';
import { SessionManager } from '../../state/session';

export class ChatPanel {
	public static currentPanel: ChatPanel | undefined;
	private static readonly viewType = 'orchionChat';
	private readonly _panel: vscode.WebviewPanel;
	private readonly _extensionUri: vscode.Uri;
	private _disposables: vscode.Disposable[] = [];
	private client: OrchestratorClient;
	private sessionId: string | null = null;

	public static createOrShow(extensionUri: vscode.Uri): ChatPanel {
		const column = vscode.window.activeTextEditor
			? vscode.window.activeTextEditor.viewColumn
			: undefined;

		// If we already have a panel, show it
		if (ChatPanel.currentPanel) {
			ChatPanel.currentPanel._panel.reveal(column);
			return ChatPanel.currentPanel;
		}

		// Otherwise, create a new panel
		const panel = vscode.window.createWebviewPanel(
			ChatPanel.viewType,
			'Orchion Chat',
			column || vscode.ViewColumn.One,
			{
				enableScripts: true,
				localResourceRoots: [vscode.Uri.joinPath(extensionUri, 'dist')],
			}
		);

		ChatPanel.currentPanel = new ChatPanel(panel, extensionUri);
		return ChatPanel.currentPanel;
	}

	private constructor(panel: vscode.WebviewPanel, extensionUri: vscode.Uri) {
		this._panel = panel;
		this._extensionUri = extensionUri;
		const settings = SettingsManager.getSettings();
		this.client = new OrchestratorClient(settings.orchestratorUrl);

		// Set the webview's initial html content
		this._update();

		// Listen for when the panel is disposed
		// This happens when the user closes the panel or when the panel is closed programmatically
		this._panel.onDidDispose(() => this.dispose(), null, this._disposables);

		// Handle messages from the webview
		this._panel.webview.onDidReceiveMessage(
			async (message) => {
				switch (message.type) {
					case 'sendMessage':
						await this.handleSendMessage(message.message, message.model);
						break;
					case 'clearChat':
						this.handleClearChat();
						break;
				}
			},
			null,
			this._disposables
		);

		// Update settings when they change
		SettingsManager.onDidChangeSettings((newSettings) => {
			this.client = new OrchestratorClient(newSettings.orchestratorUrl);
		});
	}

	private async handleSendMessage(message: ChatMessage, model: string): Promise<void> {
		if (!this.sessionId) {
			this.sessionId = SessionManager.createSession(model);
		}

		SessionManager.addMessage(this.sessionId, message);

		try {
			// Get conversation history
			const session = SessionManager.getCurrentSession();
			const messages = session?.messages || [message];

			// Send request to orchestrator
			const request = {
				model,
				messages,
				temperature: 0.7,
				stream: true,
			};

			let assistantContent = '';

			// Stream the response
			for await (const chunk of this.client.chatCompletion(request)) {
				if (chunk.choices && chunk.choices.length > 0) {
					const delta = chunk.choices[0].message?.content || '';
					if (delta) {
						assistantContent += delta;
						this._panel.webview.postMessage({
							type: 'streamChunk',
							content: delta,
						});
					}
				}
			}

			// Save assistant message
			if (assistantContent) {
				SessionManager.addMessage(this.sessionId, {
					role: 'assistant',
					content: assistantContent,
				});
			}

			this._panel.webview.postMessage({ type: 'streamComplete' });
		} catch (error) {
			const errorMessage = error instanceof Error ? error.message : String(error);
			this._panel.webview.postMessage({
				type: 'error',
				error: errorMessage,
			});
			vscode.window.showErrorMessage(`Orchion Chat Error: ${errorMessage}`);
		}
	}

	private handleClearChat(): void {
		if (this.sessionId) {
			SessionManager.clearSession(this.sessionId);
		}
		this._panel.webview.postMessage({
			type: 'loadMessages',
			messages: [],
		});
	}

	/**
	 * Send a message to the chat panel programmatically
	 */
	public sendMessage(
		message: { role: 'user' | 'assistant' | 'system'; content: string },
		model?: string
	): void {
		if (model) {
			this._panel.webview.postMessage({
				type: 'setModel',
				model: model,
			});
		}
		this._panel.webview.postMessage({
			type: 'sendMessage',
			message: message,
			model: model || SettingsManager.getSettings().defaultModel,
		});
	}

	public dispose(): void {
		ChatPanel.currentPanel = undefined;

		// Clean up our resources
		this._panel.dispose();

		while (this._disposables.length) {
			const x = this._disposables.pop();
			if (x) {
				x.dispose();
			}
		}
	}

	private _update(): void {
		const webview = this._panel.webview;
		this._panel.webview.html = this._getHtmlForWebview(webview);
	}

	private _getHtmlForWebview(webview: vscode.Webview): string {
		// Get paths to resources
		const scriptUri = webview.asWebviewUri(
			vscode.Uri.joinPath(this._extensionUri, 'dist', 'webviews', 'chat', 'main.js')
		);
		const styleUri = webview.asWebviewUri(
			vscode.Uri.joinPath(this._extensionUri, 'dist', 'webviews', 'chat', 'styles.css')
		);

		// Inline HTML with webview URIs
		// Note: VS Code automatically provides acquireVsCodeApi() in webview context
		return `<!DOCTYPE html>
<html lang="en">
<head>
	<meta charset="UTF-8">
	<meta name="viewport" content="width=device-width, initial-scale=1.0">
	<title>Orchion Chat</title>
	<link rel="stylesheet" href="${styleUri}">
</head>
<body>
	<div class="chat-container">
		<div class="chat-header">
			<select id="model-selector" class="model-selector">
				<option value="llama3.2">Llama 3.2</option>
				<option value="llama3.1">Llama 3.1</option>
				<option value="mistral">Mistral</option>
			</select>
			<button id="clear-chat" class="clear-button" title="Clear chat">Clear</button>
		</div>
		<div id="messages" class="messages-container"></div>
		<div class="input-container">
			<textarea id="message-input" class="message-input" placeholder="Type your message here... (Shift+Enter for new line, Enter to send)"></textarea>
			<button id="send-button" class="send-button">Send</button>
		</div>
	</div>
	<script src="${scriptUri}"></script>
</body>
</html>`;
	}
}
