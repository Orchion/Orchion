// This file runs in the webview context, not the extension host
// VS Code provides acquireVsCodeApi() function
declare function acquireVsCodeApi(): {
	postMessage: (message: any) => void;
	getState: () => any;
	setState: (state: any) => void;
};

const vscode = acquireVsCodeApi();

interface ChatMessage {
	role: 'system' | 'user' | 'assistant';
	content: string;
}

interface WebviewMessage {
	type: string;
	[key: string]: any;
}

class ChatWebview {
	private messagesContainer: HTMLElement;
	private messageInput: HTMLTextAreaElement;
	private sendButton: HTMLButtonElement;
	private modelSelector: HTMLSelectElement;
	private clearButton: HTMLButtonElement;
	private currentModel: string = 'llama3.2';
	private messages: ChatMessage[] = [];

	constructor() {
		this.messagesContainer = document.getElementById('messages') as HTMLElement;
		this.messageInput = document.getElementById('message-input') as HTMLTextAreaElement;
		this.sendButton = document.getElementById('send-button') as HTMLButtonElement;
		this.modelSelector = document.getElementById('model-selector') as HTMLSelectElement;
		this.clearButton = document.getElementById('clear-chat') as HTMLButtonElement;

		this.setupEventListeners();
		this.restoreState();
	}

	private setupEventListeners(): void {
		this.sendButton.addEventListener('click', () => this.sendMessage());
		this.messageInput.addEventListener('keydown', (e) => {
			if (e.key === 'Enter' && !e.shiftKey) {
				e.preventDefault();
				this.sendMessage();
			}
		});
		this.modelSelector.addEventListener('change', (e) => {
			this.currentModel = (e.target as HTMLSelectElement).value;
			this.saveState();
		});
		this.clearButton.addEventListener('click', () => this.clearChat());

		// Listen for messages from extension host
		window.addEventListener('message', (event) => {
			this.handleMessage(event.data);
		});
	}

	private sendMessage(): void {
		const content = this.messageInput.value.trim();
		if (!content) return;

		const userMessage: ChatMessage = {
			role: 'user',
			content,
		};

		this.addMessage(userMessage);
		this.messageInput.value = '';
		this.saveState();

		// Send message to extension host
		vscode.postMessage({
			type: 'sendMessage',
			message: userMessage,
			model: this.currentModel,
		});

		// Show streaming placeholder
		const assistantMessage: ChatMessage = {
			role: 'assistant',
			content: '',
		};
		const messageElement = this.addMessage(assistantMessage, true);
		this.scrollToBottom();
	}

	private handleMessage(message: WebviewMessage): void {
		switch (message.type) {
			case 'sendMessage':
				// Handle programmatic message sending from extension host
				this.handleIncomingMessage(message.message, message.model);
				break;
			case 'streamChunk':
				this.appendToLastAssistantMessage(message.content);
				this.scrollToBottom();
				break;
			case 'streamComplete':
				this.completeStreaming();
				this.saveState();
				break;
			case 'error':
				this.showError(message.error);
				this.completeStreaming();
				break;
			case 'setModel':
				this.currentModel = message.model;
				this.modelSelector.value = message.model;
				break;
			case 'loadMessages':
				this.messages = message.messages || [];
				this.renderMessages();
				break;
		}
	}

	/**
	 * Handle incoming message from extension host (programmatic send via promptAgent command)
	 */
	private handleIncomingMessage(chatMessage: ChatMessage, model?: string): void {
		if (model) {
			this.currentModel = model;
			this.modelSelector.value = model;
		}

		// Add user message to UI
		this.addMessage(chatMessage);
		this.saveState();

		// Send to extension host for processing
		vscode.postMessage({
			type: 'sendMessage',
			message: chatMessage,
			model: this.currentModel,
		});

		// Show streaming placeholder for assistant response
		const assistantMessage: ChatMessage = {
			role: 'assistant',
			content: '',
		};
		this.addMessage(assistantMessage, true);
		this.scrollToBottom();
	}

	private addMessage(message: ChatMessage, isStreaming: boolean = false): HTMLElement {
		this.messages.push(message);
		const messageDiv = document.createElement('div');
		messageDiv.className = `message ${message.role}${isStreaming ? ' streaming' : ''}`;

		const roleDiv = document.createElement('div');
		roleDiv.className = 'message-role';
		roleDiv.textContent = message.role;

		const contentDiv = document.createElement('div');
		contentDiv.className = 'message-content';
		contentDiv.textContent = message.content;

		messageDiv.appendChild(roleDiv);
		messageDiv.appendChild(contentDiv);
		this.messagesContainer.appendChild(messageDiv);

		return messageDiv;
	}

	private appendToLastAssistantMessage(content: string): void {
		const messages = this.messagesContainer.querySelectorAll('.message.assistant');
		if (messages.length === 0) return;

		const lastMessage = messages[messages.length - 1];
		const contentDiv = lastMessage.querySelector('.message-content');
		if (contentDiv) {
			const lastMsg = this.messages[this.messages.length - 1];
			if (lastMsg.role === 'assistant') {
				lastMsg.content += content;
				contentDiv.textContent = lastMsg.content;
			}
		}
	}

	private completeStreaming(): void {
		const streamingMessages = this.messagesContainer.querySelectorAll('.message.streaming');
		streamingMessages.forEach((msg) => {
			msg.classList.remove('streaming');
		});
		this.sendButton.disabled = false;
	}

	private clearChat(): void {
		this.messages = [];
		this.messagesContainer.innerHTML = '';
		this.saveState();
		vscode.postMessage({ type: 'clearChat' });
	}

	private renderMessages(): void {
		this.messagesContainer.innerHTML = '';
		this.messages.forEach((msg) => {
			this.addMessage(msg);
		});
		this.scrollToBottom();
	}

	private showError(error: string): void {
		const errorDiv = document.createElement('div');
		errorDiv.className = 'error-message';
		errorDiv.textContent = `Error: ${error}`;
		this.messagesContainer.appendChild(errorDiv);
		this.scrollToBottom();
	}

	private scrollToBottom(): void {
		this.messagesContainer.scrollTop = this.messagesContainer.scrollHeight;
	}

	private saveState(): void {
		vscode.setState({
			messages: this.messages,
			model: this.currentModel,
		});
	}

	private restoreState(): void {
		const state = vscode.getState();
		if (state) {
			this.messages = state.messages || [];
			this.currentModel = state.model || 'llama3.2';
			this.modelSelector.value = this.currentModel;
			this.renderMessages();
		}
	}
}

// Initialize when DOM is ready
if (document.readyState === 'loading') {
	document.addEventListener('DOMContentLoaded', () => {
		new ChatWebview();
	});
} else {
	new ChatWebview();
}
