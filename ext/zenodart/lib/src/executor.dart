import 'dart:io';
import 'node.dart';
import 'scope.dart';
import 'lexer.dart';
import 'parser.dart';

typedef Handler = Future<void> Function(
    Node node, Scope scope, Executor executor);

class Executor {
  final Map<String, Handler> _handlers = {};

  Executor() {
    registerHandler('log', _handleLog);
    registerHandler('print', _handleLog);
    registerHandler('if', _handleIf);
    registerHandler('foreach', _handleForeach);
    registerHandler('include', _handleInclude);
  }

  void registerHandler(String name, Handler handler) {
    _handlers[name] = handler;
  }

  Future<void> execute(Node node, Scope scope) async {
    // If root, traverse children
    if (node.name == "root") {
      for (final child in node.children) {
        await execute(child, scope);
      }
      return;
    }

    // 1. Check Variable Assignment ($var: value)
    if (node.name.length > 1 && node.name.startsWith(r'$')) {
      final varName = node.name.substring(1); // remove $
      final value = resolveValue(node, scope);
      scope.set(varName, value);
      return;
    }

    final handler = _handlers[node.name];
    if (handler != null) {
      await handler(node, scope, this);
    } else {
      // Default traversal if no handler
      for (final child in node.children) {
        await execute(child, scope);
      }
    }
  }

  // Public wrapper for resolveValue so handlers can use it
  dynamic resolveValue(Node node, Scope scope) {
    // A. If has children, treat as Map
    if (node.children.isNotEmpty) {
      final map = <String, dynamic>{};
      for (final child in node.children) {
        // Recursively resolve children
        map[child.name] = resolveValue(child, scope);
      }
      return map;
    }

    // Use Evaluator for simple expressions/literals
    return evaluateExpression(node.value, scope);
  }

  dynamic evaluateExpression(dynamic expression, Scope scope) {
    if (expression == null) return null;
    final expr = expression.toString().trim();
    if (expr.isEmpty) return null;

    // 1. Literal Checks
    if (expr == "true") return true;
    if (expr == "false") return false;
    if (expr == "null") return null;

    // String literals
    if ((expr.startsWith('"') && expr.endsWith('"')) ||
        (expr.startsWith("'") && expr.endsWith("'"))) {
      return expr.substring(1, expr.length - 1);
    }

    // Numeric literals
    final asInt = int.tryParse(expr);
    if (asInt != null) return asInt;
    final asDouble = double.tryParse(expr);
    if (asDouble != null) return asDouble;

    // List literal (Basic)
    if (expr.startsWith('[') && expr.endsWith(']')) {
      final content = expr.substring(1, expr.length - 1);
      if (content.trim().isEmpty) return [];
      // Naive split by comma (WARNING: breaks on strings with commas)
      return content
          .split(',')
          .map((e) => evaluateExpression(e.trim(), scope))
          .toList();
    }

    // Simple Equality Check (==)
    if (expr.contains("==")) {
      final parts = expr.split("==");
      if (parts.length == 2) {
        final left = evaluateExpression(parts[0].trim(), scope);
        final right = evaluateExpression(parts[1].trim(), scope);
        return left.toString() == right.toString();
      }
    }

    // Simple Addition (+) - String Concatenation or Math
    if (expr.contains("+")) {
      // Naive split (WARNING: breaks on strings with +)
      final parts = expr.split("+");
      if (parts.length >= 2) {
        dynamic result = evaluateExpression(parts[0].trim(), scope);
        for (var i = 1; i < parts.length; i++) {
          final next = evaluateExpression(parts[i].trim(), scope);
          if (result is num && next is num) {
            result += next;
          } else {
            result = result.toString() + next.toString();
          }
        }
        return result;
      }
    }

    // Variable lookup
    if (expr.startsWith(r'$')) {
      // Handle property access $user.name
      final key = expr.substring(1);
      final (val, ok) = scope.lookup(key);
      if (ok) return val;
      return null;
    }

    // Fallback
    return expression;
  }

  Future<void> _handleLog(Node node, Scope scope, Executor executor) async {
    // Evaluate value
    final val = executor.resolveValue(node, scope);
    print(val);
  }

  Future<void> _handleIf(Node node, Scope scope, Executor executor) async {
    // Evaluate condition
    final conditionVal = evaluateExpression(node.value, scope);
    bool isTrue = false;

    if (conditionVal is bool) {
      isTrue = conditionVal;
    } else if (conditionVal != null &&
        conditionVal != "" &&
        conditionVal != 0) {
      isTrue = true; // Truthy
    }

    if (isTrue) {
      // Execute children directly OR check for 'then' block
      bool hasThen = false;
      for (final child in node.children) {
        if (child.name == "then") {
          hasThen = true;
          for (final grandChild in child.children) {
            await executor.execute(grandChild, scope);
          }
        }
      }

      if (!hasThen) {
        // Direct children execution
        for (final child in node.children) {
          if (child.name != "else") {
            await executor.execute(child, scope);
          }
        }
      }
    } else {
      // Execute else block if exists
      for (final child in node.children) {
        if (child.name == "else") {
          for (final grandChild in child.children) {
            await executor.execute(grandChild, scope);
          }
        }
      }
    }
  }

  Future<void> _handleForeach(Node node, Scope scope, Executor executor) async {
    // Note: We use evaluateExpression directly because resolveValue would treat
    // this node as a Map (since it has children), which is NOT what we want for foreach.
    final iterable = executor.evaluateExpression(node.value, scope);

    if (iterable is List) {
      for (var i = 0; i < iterable.length; i++) {
        final item = iterable[i];
        // Create loop scope
        final loopScope = Scope(parent: scope);
        loopScope.set('it', item);
        loopScope.set('index', i);
        loopScope.set('value', item); // Alias

        for (final child in node.children) {
          await executor.execute(child, loopScope);
        }
      }
    } else if (iterable is Map) {
      // Map iteration needs to be async-aware
      // iterable.forEach is synchronous and doesn't support async callback well.
      // Use for-in loop over entries.
      for (final entry in iterable.entries) {
        final key = entry.key;
        final value = entry.value;

        final loopScope = Scope(parent: scope);
        loopScope.set('key', key);
        loopScope.set('value', value);
        loopScope.set('it', value); // Alias

        for (final child in node.children) {
          await executor.execute(child, loopScope);
        }
      }
    } else {
      // Not iterable
    }
  }

  Future<void> _handleInclude(Node node, Scope scope, Executor executor) async {
    final pathRaw = executor.evaluateExpression(node.value, scope);
    final path = pathRaw?.toString();

    if (path == null || path.isEmpty) {
      print("Error: include missing path");
      return;
    }

    // Resolve path relative to current script (ideally)
    // Or relative to CWD. For now CWD.
    final file = File(path);
    if (!file.existsSync()) {
      // Try relative to 'src' if not found?
      // Let's keep it simple: relative to CWD where `dart run` is executed.
      print(
          "Error: include file not found: $path (CWD: ${Directory.current.path})");
      return;
    }

    try {
      final content = await file.readAsString();
      final lexer = Lexer(content);
      final parser = Parser(lexer, filename: path);
      final ast = parser.parse();

      // Execute AST in the SAME scope (mixin style)
      await executor.execute(ast, scope);
    } catch (e) {
      print("Error including file $path: $e");
    }
  }
}
