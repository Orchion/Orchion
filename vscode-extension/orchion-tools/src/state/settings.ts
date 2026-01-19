import * as vscode from 'vscode';

export interface OrchionSettings {
	orchestratorUrl: string;
	defaultModel: string;
	refreshInterval: number;
}

const DEFAULT_SETTINGS: OrchionSettings = {
	orchestratorUrl: 'http://localhost:8080',
	defaultModel: 'llama3.2',
	refreshInterval: 5000, // 5 seconds
};

export class SettingsManager {
	private static readonly SECTION = 'orchion';

	static getSettings(): OrchionSettings {
		const config = vscode.workspace.getConfiguration(this.SECTION);
		return {
			orchestratorUrl:
				config.get<string>('orchestratorUrl') || DEFAULT_SETTINGS.orchestratorUrl,
			defaultModel: config.get<string>('defaultModel') || DEFAULT_SETTINGS.defaultModel,
			refreshInterval:
				config.get<number>('refreshInterval') || DEFAULT_SETTINGS.refreshInterval,
		};
	}

	static async updateSettings(updates: Partial<OrchionSettings>): Promise<void> {
		const config = vscode.workspace.getConfiguration(this.SECTION);
		for (const [key, value] of Object.entries(updates)) {
			await config.update(key, value, vscode.ConfigurationTarget.Global);
		}
	}

	static onDidChangeSettings(callback: (settings: OrchionSettings) => void): vscode.Disposable {
		return vscode.workspace.onDidChangeConfiguration((e) => {
			if (e.affectsConfiguration(this.SECTION)) {
				callback(this.getSettings());
			}
		});
	}
}
