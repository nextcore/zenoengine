import 'node.dart';
import 'scope.dart';

typedef Handler = void Function(Node node, Scope scope, Executor executor);

class Executor {
  final Map<String, Handler> _handlers = {};

  Executor() {
    registerHandler('log', _handleLog);
    registerHandler('print', _handleLog);
    registerHandler('if', _handleIf);
  }

  void registerHandler(String name, Handler handler) {
    _handlers[name] = handler;
  }

  void execute(Node node, Scope scope) {
    // If root, traverse children
    if (node.name == "root") {
      for (final child in node.children) {
        execute(child, scope);
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
      handler(node, scope, this);
    } else {
      // Default traversal if no handler
      for (final child in node.children) {
        execute(child, scope);
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

    // B. Raw Value
    final valStr = node.value?.toString() ?? "";
    if (valStr.isEmpty) return null;

    // C. String Literal (quotes)
    if (valStr.length >= 2) {
      if ((valStr.startsWith('"') && valStr.endsWith('"')) ||
          (valStr.startsWith("'") && valStr.endsWith("'"))) {
        return valStr.substring(1, valStr.length - 1);
      }
    }

    // D. Variable Reference ($var)
    if (valStr.startsWith(r'$')) {
      final key = valStr.substring(1);
      final (val, ok) = scope.lookup(key);
      if (ok) return val;
      // Fallback to raw value if var not found
      return node.value;
    }

    // E. Numeric/Bool parsing (Basic)
    if (valStr == "true") return true;
    if (valStr == "false") return false;
    final asInt = int.tryParse(valStr);
    if (asInt != null) return asInt;
    final asDouble = double.tryParse(valStr);
    if (asDouble != null) return asDouble;

    return node.value;
  }

  void _handleLog(Node node, Scope scope, Executor executor) {
    // Evaluate value
    final val = executor.resolveValue(node, scope);
    print(val);
  }

  void _handleIf(Node node, Scope scope, Executor executor) {
    // Basic condition evaluation
    final conditionRaw = node.value?.toString() ?? "";
    bool isTrue = false;

    if (conditionRaw == "true") {
      isTrue = true;
    } else if (conditionRaw.contains("==")) {
      final parts = conditionRaw.split("==");
      if (parts.length == 2) {
        final leftRaw = parts[0].trim();
        final rightRaw = parts[1].trim();

        dynamic leftVal = leftRaw;
        if (leftRaw.startsWith(r'$')) {
          final (val, ok) = scope.lookup(leftRaw.substring(1));
          if (ok) leftVal = val;
          // else keep raw string
        } else if ((leftRaw.startsWith('"') && leftRaw.endsWith('"')) ||
            (leftRaw.startsWith("'") && leftRaw.endsWith("'"))) {
          leftVal = leftRaw.substring(1, leftRaw.length - 1);
        }

        dynamic rightVal = rightRaw;
        if (rightRaw.startsWith(r'$')) {
          final (val, ok) = scope.lookup(rightRaw.substring(1));
          if (ok) rightVal = val;
          // else keep raw string
        } else if ((rightRaw.startsWith('"') && rightRaw.endsWith('"')) ||
            (rightRaw.startsWith("'") && rightRaw.endsWith("'"))) {
          rightVal = rightRaw.substring(1, rightRaw.length - 1);
        }

        isTrue = leftVal.toString() == rightVal.toString();
      }
    } else if (conditionRaw.startsWith(r'$')) {
      final (val, ok) = scope.lookup(conditionRaw.substring(1));
      if (ok && val != null && val != false && val != "" && val != 0) {
        isTrue = true;
      }
    }

    if (isTrue) {
      bool hasThen = false;
      for (final child in node.children) {
        if (child.name == "then") {
          hasThen = true;
          for (final grandChild in child.children) {
            executor.execute(grandChild, scope);
          }
        }
      }

      if (!hasThen) {
        for (final child in node.children) {
          if (child.name != "else" && child.name != "then") {
            executor.execute(child, scope);
          }
        }
      }
    } else {
      for (final child in node.children) {
        if (child.name == "else") {
          for (final grandChild in child.children) {
            executor.execute(grandChild, scope);
          }
        }
      }
    }
  }
}
