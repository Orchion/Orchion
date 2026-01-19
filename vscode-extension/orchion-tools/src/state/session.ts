import * as vscode from 'vscode';
import { ChatMessage } from '../api/orchestratorClient';

export interface ChatSession {
	messages: ChatMessage[];
	model: string;
	createdAt: number;
}

export class SessionManager {
	private static sessions: Map<string, ChatSession> = new Map();
	private static currentSessionId: string | null = null;

	static createSession(model: string): string {
		const sessionId = `session-${Date.now()}`;
		const session: ChatSession = {
			messages: [],
			model,
			createdAt: Date.now(),
		};
		this.sessions.set(sessionId, session);
		this.currentSessionId = sessionId;
		return sessionId;
	}

	static getSession(sessionId: string): ChatSession | undefined {
		return this.sessions.get(sessionId);
	}

	static getCurrentSession(): ChatSession | null {
		if (!this.currentSessionId) {
			return null;
		}
		return this.sessions.get(this.currentSessionId) || null;
	}

	static setCurrentSession(sessionId: string): void {
		if (this.sessions.has(sessionId)) {
			this.currentSessionId = sessionId;
		}
	}

	static addMessage(sessionId: string, message: ChatMessage): void {
		const session = this.sessions.get(sessionId);
		if (session) {
			session.messages.push(message);
		}
	}

	static clearSession(sessionId: string): void {
		const session = this.sessions.get(sessionId);
		if (session) {
			session.messages = [];
		}
	}

	static deleteSession(sessionId: string): void {
		this.sessions.delete(sessionId);
		if (this.currentSessionId === sessionId) {
			this.currentSessionId = null;
		}
	}

	static listSessions(): string[] {
		return Array.from(this.sessions.keys());
	}
}
