import * as vscode from 'vscode';
import { ChatPanel } from '../webviews/chat/panel';
import { OrchestratorClient } from '../api/orchestratorClient';
import { SettingsManager } from '../state/settings';

/**
 * Prompt an agent with a message
 * Opens chat panel if not already open, then sends the prompt
 */
export async function promptAgent(
	context: vscode.ExtensionContext,
	agent?: string,
	message?: string
): Promise<void> {
	const settings = SettingsManager.getSettings();
	const client = new OrchestratorClient(settings.orchestratorUrl);

	// If no agent specified, show quick pick
	if (!agent) {
		try {
			const agents = await client.listAgents();
			if (agents.length === 0) {
				vscode.window.showWarningMessage('No agents available');
				return;
			}

			const selectedAgent = await vscode.window.showQuickPick(
				agents.map((a) => ({
					label: a.name,
					description: a.description || a.model,
					agent: a.id,
				})),
				{
					placeHolder: 'Select an agent',
				}
			);

			if (!selectedAgent) {
				return;
			}
			agent = selectedAgent.agent;
		} catch (error) {
			vscode.window.showErrorMessage(
				`Failed to list agents: ${error instanceof Error ? error.message : String(error)}`
			);
			return;
		}
	}

	// If no message specified, prompt user
	if (!message) {
		const input = await vscode.window.showInputBox({
			prompt: `Enter message for ${agent}`,
			placeHolder: 'Type your message...',
		});

		if (!input) {
			return;
		}
		message = input;
	}

	// Get or create chat panel
	const panel = ChatPanel.createOrShow(context.extensionUri);

	// Send message to chat panel
	panel.sendMessage(
		{
			role: 'user',
			content: message!,
		},
		agent
	);
}
