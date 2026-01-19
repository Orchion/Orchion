import * as vscode from 'vscode';
import { ClusterTreeProvider } from '../views/clusterTree';

let clusterProvider: ClusterTreeProvider | undefined;

export function setClusterProvider(provider: ClusterTreeProvider): void {
	clusterProvider = provider;
}

export function refreshNodes(): void {
	if (clusterProvider) {
		clusterProvider.refresh();
		vscode.window.showInformationMessage('Orchion: Refreshed nodes');
	} else {
		vscode.window.showWarningMessage('Orchion: Cluster view not initialized');
	}
}
