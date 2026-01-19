import * as vscode from 'vscode';
import { OrchestratorClient, Node } from '../api/orchestratorClient';
import { SettingsManager } from '../state/settings';

export class ClusterTreeProvider implements vscode.TreeDataProvider<ClusterItem> {
	private _onDidChangeTreeData: vscode.EventEmitter<
		ClusterItem | undefined | null | void
	> = new vscode.EventEmitter<ClusterItem | undefined | null | void>();
	readonly onDidChangeTreeData: vscode.Event<
		ClusterItem | undefined | null | void
	> = this._onDidChangeTreeData.event;

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

	getTreeItem(element: ClusterItem): vscode.TreeItem {
		return element;
	}

	async getChildren(element?: ClusterItem): Promise<ClusterItem[]> {
		if (!element) {
			// Root level - show nodes
			try {
				const nodes = await this.client.listNodes();
				if (nodes.length === 0) {
					return [
						new ClusterItem(
							'No nodes registered',
							vscode.TreeItemCollapsibleState.None
						),
					];
				}
				return nodes.map(
					(node) =>
						new ClusterItem(
							`${node.hostname} (${node.id})`,
							vscode.TreeItemCollapsibleState.Collapsed,
							node
						)
				);
			} catch (error) {
				return [
					new ClusterItem(
						`Error: ${error instanceof Error ? error.message : String(error)}`,
						vscode.TreeItemCollapsibleState.None
					),
				];
			}
		} else if (element.node) {
			// Node details
			const node = element.node;
			const items: ClusterItem[] = [];

			items.push(
				new ClusterItem(`ID: ${node.id}`, vscode.TreeItemCollapsibleState.None)
			);
			items.push(
				new ClusterItem(
					`Hostname: ${node.hostname}`,
					vscode.TreeItemCollapsibleState.None
				)
			);

			if (node.capabilities) {
				const caps = node.capabilities;
				if (caps.cpu) {
					items.push(
						new ClusterItem(
							`CPU: ${caps.cpu}`,
							vscode.TreeItemCollapsibleState.None
						)
					);
				}
				if (caps.memory) {
					items.push(
						new ClusterItem(
							`Memory: ${caps.memory}`,
							vscode.TreeItemCollapsibleState.None
						)
					);
				}
				if (caps.os) {
					items.push(
						new ClusterItem(
							`OS: ${caps.os}`,
							vscode.TreeItemCollapsibleState.None
						)
					);
				}
			}

			if (node.last_seen_unix) {
				const lastSeen = new Date(node.last_seen_unix * 1000);
				items.push(
					new ClusterItem(
						`Last Seen: ${lastSeen.toLocaleString()}`,
						vscode.TreeItemCollapsibleState.None
					)
				);
			}

			return items;
		}

		return [];
	}

	dispose(): void {
		if (this.refreshInterval) {
			clearInterval(this.refreshInterval);
		}
	}
}

class ClusterItem extends vscode.TreeItem {
	constructor(
		public readonly label: string,
		public readonly collapsibleState: vscode.TreeItemCollapsibleState,
		public readonly node?: Node
	) {
		super(label, collapsibleState);
		this.tooltip = node ? `${node.hostname} - ${node.id}` : label;
		this.contextValue = node ? 'node' : undefined;
	}
}
