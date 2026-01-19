import * as vscode from 'vscode';
import { ChatPanel } from '../webviews/chat/panel';

export function openChat(context: vscode.ExtensionContext): void {
	const panel = ChatPanel.createOrShow(context.extensionUri);
}
