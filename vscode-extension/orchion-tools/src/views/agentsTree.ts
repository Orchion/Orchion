import * as vscode from 'vscode';
import { OrchestratorClient } from '../api/orchestratorClient';
import { SettingsManager } from '../state/settings';

export interface Agent {
	id: string;
	name: string;
	model: string;
	status: 'active' | 'idle' | 'error';
}

export class AgentsTreeProvider implements vscode.TreeDataProvider<AgentItem> {
	private _onDidChangeTreeData: vscode.EventEmitter<AgentItem | undefined | null | void> =
		new vscode.EventEmitter<AgentItem | undefined | null | void>();
	readonly onDidChangeTreeData: vscode.Event<AgentItem | undefined | null | void> =
		this._onDidChangeTreeData.event;

	private client: OrchestratorClient;
	private refreshInterval: NodeJS.Timeout | undefined;

	constructor() {
		const settings = SettingsManager.getSettings();
		this.client = new OrchestratorClient(settings.orchestratorUrl);

		// Set up auto-refresh
		this.startAutoRefresh(settings.refreshInterval);

		// Update client when settings change
		SettingsManager.onDidChangeSettings((newSettings) => {
			this.client = new OrchestratorClient(newSettings.orchestratorUrl);
			if (this.refreshInterval) {
				clearInterval(this.refreshInterval);
			}
			this.startAutoRefresh(newSettings.refreshInterval);
			this.refresh();
		});
	}

	private startAutoRefresh(interval: number): void {
		if (this.refreshInterval) {
			clearInterval(this.refreshInterval);
		}
		this.refreshInterval = setInterval(() => {
			this.refresh();
		}, interval);
	}

	refresh(): void {
		this._onDidChangeTreeData.fire();
	}

	getTreeItem(element: AgentItem): vscode.TreeItem {
		return element;
	}

	async getChildren(element?: AgentItem): Promise<AgentItem[]> {
		if (!element) {
			// Root level - show agents
			try {
				const agentsList = await this.client.listAgents();
				if (agentsList.length === 0) {
					return [
						new AgentItem('No agents available', vscode.TreeItemCollapsibleState.None),
					];
				}

				// Convert API response to Agent format
				const agents: Agent[] = agentsList.map((a) => ({
					id: a.id,
					name: a.name,
					model: a.model,
					status: 'active', // TODO: Get actual status from API
				}));

				return agents.map(
					(agent) =>
						new AgentItem(
							`${agent.name} (${agent.status})`,
							vscode.TreeItemCollapsibleState.Collapsed,
							agent
						)
				);
			} catch (error) {
				return [
					new AgentItem(
						`Error: ${error instanceof Error ? error.message : String(error)}`,
						vscode.TreeItemCollapsibleState.None
					),
				];
			}
		} else if (element.agent) {
			// Agent details
			const agent = element.agent;
			return [
				new AgentItem(`Model: ${agent.model}`, vscode.TreeItemCollapsibleState.None),
				new AgentItem(`Status: ${agent.status}`, vscode.TreeItemCollapsibleState.None),
			];
		}

		return [];
	}

	dispose(): void {
		if (this.refreshInterval) {
			clearInterval(this.refreshInterval);
		}
	}
}

class AgentItem extends vscode.TreeItem {
	constructor(
		public readonly label: string,
		public readonly collapsibleState: vscode.TreeItemCollapsibleState,
		public readonly agent?: Agent
	) {
		super(label, collapsibleState);
		this.tooltip = agent ? `${agent.name} - ${agent.status}` : label;
		this.contextValue = agent ? 'agent' : undefined;

		// Set icon based on status
		if (agent) {
			switch (agent.status) {
				case 'active':
					this.iconPath = new vscode.ThemeIcon(
						'circle-filled',
						new vscode.ThemeColor('charts.green')
					);
					break;
				case 'idle':
					this.iconPath = new vscode.ThemeIcon('circle-outline');
					break;
				case 'error':
					this.iconPath = new vscode.ThemeIcon(
						'error',
						new vscode.ThemeColor('errorForeground')
					);
					break;
			}
		}
	}
}
