import * as vscode from "vscode";

export class NodesProvider implements vscode.TreeDataProvider<string> {
  getTreeItem(element: string) {
    return new vscode.TreeItem(element);
  }

  getChildren() {
    return ["node-1", "node-2"]; // TODO: fetch from orchestrator
  }
}