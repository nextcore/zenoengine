import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class StringModule {
  static void register(Executor executor) {
    executor.registerHandler('str.upper', _handleUpper);
    executor.registerHandler('str.lower', _handleLower);
    executor.registerHandler('str.len', _handleLen);
    executor.registerHandler('str.replace', _handleReplace);
    executor.registerHandler('str.contains', _handleContains);
    executor.registerHandler('regex.match', _handleRegexMatch);
  }

  static Future<void> _handleUpper(
      Node node, Scope scope, Executor executor) async {
    final input = executor.evaluateExpression(node.value, scope)?.toString();
    if (input != null) _setResult(node, scope, executor, input.toUpperCase());
  }

  static Future<void> _handleLower(
      Node node, Scope scope, Executor executor) async {
    final input = executor.evaluateExpression(node.value, scope)?.toString();
    if (input != null) _setResult(node, scope, executor, input.toLowerCase());
  }

  static Future<void> _handleLen(
      Node node, Scope scope, Executor executor) async {
    final input = executor.evaluateExpression(node.value, scope);
    if (input is String)
      _setResult(node, scope, executor, input.length);
    else if (input is List)
      _setResult(node, scope, executor, input.length);
    else if (input is Map)
      _setResult(node, scope, executor, input.length);
    else
      _setResult(node, scope, executor, 0);
  }

  static Future<void> _handleReplace(
      Node node, Scope scope, Executor executor) async {
    final input = executor.evaluateExpression(node.value, scope)?.toString();
    if (input == null) return;

    String? from;
    String? to;

    for (final child in node.children) {
      if (child.name == 'from')
        from = executor.evaluateExpression(child.value, scope)?.toString();
      if (child.name == 'to')
        to = executor.evaluateExpression(child.value, scope)?.toString();
    }

    if (from != null && to != null) {
      _setResult(node, scope, executor, input.replaceAll(from, to));
    }
  }

  static Future<void> _handleContains(
      Node node, Scope scope, Executor executor) async {
    final input = executor.evaluateExpression(node.value, scope)?.toString();
    if (input == null) return;

    String? sub;
    for (final child in node.children) {
      if (child.name == 'substr')
        sub = executor.evaluateExpression(child.value, scope)?.toString();
    }

    if (sub != null) {
      _setResult(node, scope, executor, input.contains(sub));
    }
  }

  static Future<void> _handleRegexMatch(
      Node node, Scope scope, Executor executor) async {
    final input = executor.evaluateExpression(node.value, scope)?.toString();
    if (input == null) return;

    String? pattern;
    for (final child in node.children) {
      if (child.name == 'pattern')
        pattern = executor.evaluateExpression(child.value, scope)?.toString();
    }

    if (pattern != null) {
      try {
        final reg = RegExp(pattern);
        _setResult(node, scope, executor, reg.hasMatch(input));
      } catch (e) {
        print("regex.match error: $e");
      }
    }
  }

  static void _setResult(
      Node node, Scope scope, Executor executor, dynamic result) {
    String? asVar;
    for (final child in node.children) {
      if (child.name == 'as') {
        dynamic val = executor.evaluateExpression(child.value, scope);
        if (val == null && child.value != null) {
          val = child.value.toString();
        }
        asVar = val?.toString();
        if (asVar != null) {
          if ((asVar.startsWith('"') && asVar.endsWith('"')) ||
              (asVar.startsWith("'") && asVar.endsWith("'"))) {
            asVar = asVar.substring(1, asVar.length - 1);
          }
        }
      }
    }

    if (asVar != null) {
      if ((asVar.startsWith('"') && asVar.endsWith('"')) ||
          (asVar.startsWith("'") && asVar.endsWith("'"))) {
        asVar = asVar.substring(1, asVar.length - 1);
      }
      scope.set(asVar, result);
    } else {
      print(result);
    }
  }
}
