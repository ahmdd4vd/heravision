const vscode = require('vscode');
const { execFile } = require('child_process');
function activate(ctx) {
  ctx.subscriptions.push(vscode.commands.registerCommand('heravision.extract', async () => {
    const uri = await vscode.window.showOpenDialog({ filters: { Images: ['png','jpg','webp'] } });
    if (!uri) return;
    execFile('heravision', ['extract', uri[0].fsPath, '--json'], (err, out) => {
      if (err) return vscode.window.showErrorMessage(err.message);
      vscode.workspace.openTextDocument({ language: 'json', content: out }).then(d => vscode.window.showTextDocument(d));
    });
  }));
}
function deactivate() {}
module.exports = { activate, deactivate };
