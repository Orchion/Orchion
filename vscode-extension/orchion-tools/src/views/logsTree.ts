import * as vscode from 'vscode';

export interface LogEntry {
	id: string;
	timestamp: number;
	level: 'info' | 'warning' | 'error';
	message: string;
	source?: string;
}

export class LogsTreeProvider implements vscode.TreeDataProvider<LogItem> {
	private _onDidChangeTreeData: vscode.EventEmitter<
		LogItem | undefined | null | void
	> = new vscode.EventEmitter<LogItem | undefined | null | void>();
	readonly onDidChangeTreeData: vscode.Event<
		LogItem | undefined | null | void
	> = this._onDidChangeTreeData.event;

	private logs: LogEntry[] = [];
	private maxLogs: number = 100;

	refresh(): void {
		this._onDidChangeTreeData.fire();
	}

	addLog(entry: LogEntry): void {
		this.logs.unshift(entry); // Add to beginning
		if (this.logs.length > this.maxLogs) {
			this.logs = this.logs.slice(0, this.maxLogs);
		}
		this.refresh();
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
				return [
					new LogItem(
						'No logs',
						vscode.TreeItemCollapsibleState.None
					),
				];
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
		this.tooltip = log
			? `${new Date(log.timestamp).toLocaleString()} - ${log.message}`
			: label;
		this.contextValue = log ? 'log' : undefined;

		// Set icon based on log level
		if (log) {
			switch (log.level) {
				case 'info':
					this.iconPath = new vscode.ThemeIcon('info');
					break;
				case 'warning':
					this.iconPath = new vscode.ThemeIcon('warning', new vscode.ThemeColor('warningForeground'));
					break;
				case 'error':
					this.iconPath = new vscode.ThemeIcon('error', new vscode.ThemeColor('errorForeground'));
					break;
			}
		}
	}
}
