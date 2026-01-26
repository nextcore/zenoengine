const vscode = require('vscode');

/**
 * @param {vscode.ExtensionContext} context
 */
function activate(context) {
  console.log('ZenoLang extension is active');

  const provider = vscode.languages.registerCompletionItemProvider('zenolang', {
    provideCompletionItems(document, position, token, context) {

      // Keywords & Control Flow
      const keywords = [
        'if', 'else', 'try', 'catch', 'for', 'while', 'return', 'throw',
        'then', 'do', 'as', 'val', 'value', 'default', 'break', 'continue'
      ];

      // Constants
      const constants = [
        'true', 'false', 'nil', 'null'
      ];

      // Core Slots and Modules from ZENOLANG_AI_DOCS.md
      const slots = [
        // Core
        'var', 'log', 'sleep', 'ctx.timeout', 'include',

        // HTTP
        'http.response', 'http.redirect', 'http.query', 'http.get', 'http.post', 'http.put', 'http.delete', 'http.upload',

        // Cookie & View
        'cookie.set', 'view.render',

        // Database
        'db.select', 'db.execute', 'db.table', 'db.where', 'db.insert', 'db.update', 'db.delete', 'db.count', 'db.order_by', 'db.limit', 'db.offset', 'db.first',

        // Auth & Security
        'auth.login', 'auth.middleware', 'validator.validate',

        // IO & System
        'io.file.read', 'io.file.write', 'io.file.delete', 'mail.send', 'job.enqueue',

        // Math & Strings (Commonly used, even if implies some form of stdlib)
        'math.calc', 'strings.concat', 'strings.split',

        // Null Safety
        'coalesce', 'is_null'
      ];

      const completionItems = [];

      // Add keywords
      for (const keyword of keywords) {
        const item = new vscode.CompletionItem(keyword, vscode.CompletionItemKind.Keyword);
        completionItems.push(item);
      }

      // Add constants
      for (const constant of constants) {
        const item = new vscode.CompletionItem(constant, vscode.CompletionItemKind.Constant);
        completionItems.push(item);
      }

      // Add slots
      for (const slot of slots) {
        const item = new vscode.CompletionItem(slot, vscode.CompletionItemKind.Function);
        item.detail = `ZenoLang Slot: ${slot}`;

        // Smart Snippets
        if (slot.startsWith('http.')) {
          // e.g. http.get: /path { do: ... }
          item.insertText = new vscode.SnippetString(`${slot}: \${1:/path} {\n\tdo: {\n\t\t$0\n\t}\n}`);
        } else if (slot.startsWith('db.select') || slot.startsWith('db.execute')) {
          item.insertText = new vscode.SnippetString(`${slot}: "\${1:SQL_QUERY}" {\n\tval: $0\n}`);
        } else if (slot === 'if') {
          item.insertText = new vscode.SnippetString(`if: \${1:condition} {\n\tthen: {\n\t\t$0\n\t}\n\telse: {\n\t\t\n\t}\n}`);
        } else if (slot === 'for') {
          item.insertText = new vscode.SnippetString(`for: \${1:$list} {\n\tas: \${2:$item}\n\tdo: {\n\t\t$0\n\t}\n}`);
        } else if (slot === 'try') {
          item.insertText = new vscode.SnippetString(`try {\n\tdo: {\n\t\t$0\n\t}\n\tcatch: {\n\t\t\n\t}\n}`);
        } else {
          item.insertText = new vscode.SnippetString(`${slot}: $0`);
        }

        completionItems.push(item);
      }

      return completionItems;
    }
  });

  context.subscriptions.push(provider);
}

function deactivate() { }

module.exports = {
  activate,
  deactivate
};
