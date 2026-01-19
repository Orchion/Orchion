import * as vscode from 'vscode';
import { ChatPanel } from './webviews/chat/panel';
import { ClusterTreeProvider } from './views/clusterTree';
import { AgentsTreeProvider } from './views/agentsTree';
import { LogsTreeProvider } from './views/logsTree';
import { openChat } from './commands/openChat';
import { refreshNodes, setClusterProvider } from './commands/refreshNodes';
import { promptAgent } from './commands/promptAgent';

export function activate(context: vscode.ExtensionContext) {
	console.log('Orchion Tools extension is now active!');

	// Register tree views
	const clusterProvider = new ClusterTreeProvider();
	const clusterView = vscode.window.createTreeView('orchionCluster', {
		treeDataProvider: clusterProvider,
		showCollapseAll: true,
	});
	context.subscriptions.push(clusterView);
	setClusterProvider(clusterProvider);

	const agentsProvider = new AgentsTreeProvider();
	const agentsView = vscode.window.createTreeView('orchionAgents', {
		treeDataProvider: agentsProvider,
		showCollapseAll: true,
	});
	context.subscriptions.push(agentsView);

	const logsProvider = new LogsTreeProvider();
	const logsView = vscode.window.createTreeView('orchionLogs', {
		treeDataProvider: logsProvider,
		showCollapseAll: true,
	});
	context.subscriptions.push(logsView);

	// Register commands
	const openChatCommand = vscode.commands.registerCommand(
		'orchion.openChat',
		() => openChat(context)
	);
	context.subscriptions.push(openChatCommand);

	const promptAgentCommand = vscode.commands.registerCommand(
		'orchion.promptAgent',
		() => promptAgent(context)
	);
	context.subscriptions.push(promptAgentCommand);

	const refreshNodesCommand = vscode.commands.registerCommand(
		'orchion.refreshNodes',
		() => refreshNodes()
	);
	context.subscriptions.push(refreshNodesCommand);

	const refreshAgentsCommand = vscode.commands.registerCommand(
		'orchion.refreshAgents',
		() => {
			agentsProvider.refresh();
			vscode.window.showInformationMessage('Orchion: Refreshed agents');
		}
	);
	context.subscriptions.push(refreshAgentsCommand);

	const clearLogsCommand = vscode.commands.registerCommand(
		'orchion.clearLogs',
		() => {
			logsProvider.clearLogs();
			vscode.window.showInformationMessage('Orchion: Cleared logs');
		}
	);
	context.subscriptions.push(clearLogsCommand);

	// Add initial log entry
	logsProvider.addLog({
		id: 'init',
		timestamp: Date.now(),
		level: 'info',
		message: 'Orchion Tools extension activated',
		source: 'extension',
	});
}

export function deactivate() {
	// Cleanup if needed
}