import * as vscode from 'vscode';
import { OrchestratorClient, LogEntry as API_LogEntry } from '../api/orchestratorClient';
import { SettingsManager } from '../state/settings';

export interface LogEntry {
	id: string;
	timestamp: number;
	level: 'info' | 'warning' | 'error';
	message: string;
	source?: string;
}

export class LogsTreeProvider implements vscode.TreeDataProvider<LogItem> {
	private _onDidChangeTreeData: vscode.EventEmitter<LogItem | undefined | null | void> =
		new vscode.EventEmitter<LogItem | undefined | null | void>();
	readonly onDidChangeTreeData: vscode.Event<LogItem | undefined | null | void> =
		this._onDidChangeTreeData.event;

	private client: OrchestratorClient;
	private logs: LogEntry[] = [];
	private maxLogs: number = 100;
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

		// Initial load
		this.loadLogs();
	}

	private startAutoRefresh(interval: number): void {
		if (this.refreshInterval) {
			clearInterval(this.refreshInterval);
		}
		this.refreshInterval = setInterval(() => {
			this.loadLogs();
		}, interval);
	}

	private async loadLogs(): Promise<void> {
		try {
			const apiLogs = await this.client.getLogs();
			this.logs = apiLogs.map((apiLog) => ({
				id: apiLog.id,
				timestamp: apiLog.timestamp,
				level: this.mapLogLevel(apiLog.level),
				message: apiLog.message,
				source: apiLog.source,
			}));
			this.refresh();
		} catch (error) {
			console.error('Failed to load logs:', error);
			// Keep existing logs on error
		}
	}

	private mapLogLevel(apiLevel: string): 'info' | 'warning' | 'error' {
		switch (apiLevel) {
			case 'debug':
			case 'info':
				return 'info';
			case 'warn':
			case 'warning':
				return 'warning';
			case 'error':
				return 'error';
			default:
				return 'info';
		}
	}

	refresh(): void {
		this.loadLogs();
	}

	addLog(entry: LogEntry): void {
		// For backwards compatibility, but logs are now loaded from orchestrator
		this.logs.unshift(entry); // Add to beginning
		if (this.logs.length > this.maxLogs) {
			this.logs = this.logs.slice(0, this.maxLogs);
		}
		this._onDidChangeTreeData.fire();
	}

	clearLogs(): void {
		this.logs = [];
		this.refresh();
	}

	getTreeItem(element: LogItem): vscode.TreeItem {
		return element;
	}

	getChildren(element?: LogItem): LogItem[] {
		if (!element) {
			// Root level - show logs grouped by date/time
			if (this.logs.length === 0) {
				return [new LogItem('No logs', vscode.TreeItemCollapsibleState.None)];
			}

			return this.logs.map(
				(log) =>
					new LogItem(
						`[${new Date(log.timestamp).toLocaleTimeString()}] ${log.message}`,
						vscode.TreeItemCollapsibleState.None,
						log
					)
			);
		}

		return [];
	}
}

class LogItem extends vscode.TreeItem {
	constructor(
		public readonly label: string,
		public readonly collapsibleState: vscode.TreeItemCollapsibleState,
		public readonly log?: LogEntry
	) {
		super(label, collapsibleState);
		this.tooltip = log ? `${new Date(log.timestamp).toLocaleString()} - ${log.message}` : label;
		this.contextValue = log ? 'log' : undefined;

		// Set icon based on log level
		if (log) {
			switch (log.level) {
				case 'info':
					this.iconPath = new vscode.ThemeIcon('info');
					break;
				case 'warning':
					this.iconPath = new vscode.ThemeIcon(
						'warning',
						new vscode.ThemeColor('warningForeground')
					);
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
