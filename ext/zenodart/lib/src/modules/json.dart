import 'dart:convert';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class JsonModule {
  static void register(Executor executor) {
    executor.registerHandler('json.parse', _handleParse);
    executor.registerHandler('json.stringify', _handleStringify);
  }

  static Future<void> _handleParse(
      Node node, Scope scope, Executor executor) async {
    // Usage: json.parse: $jsonString { as: "parsed" }

    final jsonStrRaw = executor.evaluateExpression(node.value, scope);
    final jsonStr = jsonStrRaw?.toString();

    if (jsonStr == null) return;

    try {
      final parsed = jsonDecode(jsonStr);

      String? asVar;
      for (final child in node.children) {
        if (child.name == 'as') {
          final v = executor.evaluateExpression(child.value, scope);
          asVar = v?.toString();
        }
      }

      if (asVar != null) {
        scope.set(asVar, parsed);
      } else {
        scope.set('json', parsed); // Default
      }
    } catch (e) {
      print("json.parse error: $e");
    }
  }

  static Future<void> _handleStringify(
      Node node, Scope scope, Executor executor) async {
    // Usage: json.stringify: $obj { as: "str" }

    final inputData = executor.evaluateExpression(node.value, scope);

    if (inputData == null) {
      print("json.stringify error: input data is null");
      return;
    }

    try {
      final jsonStr = jsonEncode(inputData);

      String? asVar;
      for (final child in node.children) {
        if (child.name == 'as') {
          final v = executor.evaluateExpression(child.value, scope);
          asVar = v?.toString();
        }
      }

      if (asVar != null) {
        scope.set(asVar, jsonStr);
      } else {
        scope.set('json', jsonStr);
        print(jsonStr); // Also print for debug?
      }
    } catch (e) {
      print("json.stringify error: $e");
    }
  }
}
