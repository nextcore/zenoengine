import 'dart:io';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class EnvModule {
  static void register(Executor executor) {
    executor.registerHandler('env.get', _handleGet);
  }

  static Future<void> _handleGet(
      Node node, Scope scope, Executor executor) async {
    // Usage: env.get: "KEY" { as: "var" }

    final keyRaw = executor.evaluateExpression(node.value, scope);
    final key = keyRaw?.toString();

    if (key == null || key.isEmpty) return;

    final value = Platform.environment[key];

    String? asVar;
    String? defaultVal;

    for (final child in node.children) {
      if (child.name == 'as') {
        asVar = executor.evaluateExpression(child.value, scope)?.toString();
      } else if (child.name == 'default') {
        defaultVal =
            executor.evaluateExpression(child.value, scope)?.toString();
      }
    }

    final finalVal = value ?? defaultVal;

    if (asVar != null) {
      scope.set(asVar, finalVal);
    } else {
      // If no 'as', maybe return value? (Only if used in expression, but handlers are void)
      // Set to $env_KEY?
      scope.set('env_${key}', finalVal);
    }
  }
}
