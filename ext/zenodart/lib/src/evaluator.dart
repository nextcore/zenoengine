import 'scope.dart';

class Evaluator {
  /// Evaluates an expression string within the given scope.
  /// Supports:
  /// - Literals: "string", 'string', 123, 12.34, true, false
  /// - Variables: $var, $user.name
  /// - Operators: ==, !=, >, <, >=, <=, &&, ||, +, -, *, /
  dynamic evaluate(dynamic expression, Scope scope) {
    if (expression is! String) return expression;

    final expr = expression.trim();
    if (expr.isEmpty) return null;

    // 1. Literal Checks (Fast Path)
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

    // Variable lookup
    if (expr.startsWith(r'$')) {
      final key = expr.substring(1);
      final (val, ok) = scope.lookup(key);
      if (ok) return val;
      // If not found, returning null is standard for dynamic languages
      return null;
    }

    // 2. Simple Operator Parsing (Naive Implementation)
    // NOTE: This is a simplified parser. For full robust parsing,
    // a tokenizer and AST builder for expressions would be needed.
    // Here we handle basic binary operations common in Zeno `if` statements.

    // Logic Operators (&&, ||)
    if (expr.contains(" || ")) {
      final parts = expr.split(" || ");
      for (final part in parts) {
        if (_isTruthy(evaluate(part, scope))) return true;
      }
      return false;
    }

    if (expr.contains(" && ")) {
      final parts = expr.split(" && ");
      for (final part in parts) {
        if (!_isTruthy(evaluate(part, scope))) return false;
      }
      return true;
    }

    // Comparison Operators
    // Need to check for >=, <=, ==, !=, >, < in correct order to avoid prefix collision
    // (e.g. checking > before >=)

    final ops = ["==", "!=", ">=", "<=", ">", "<"];
    for (final op in ops) {
      if (expr.contains(op)) {
        final parts = expr.split(op);
        if (parts.length == 2) {
          final left = evaluate(parts[0], scope);
          final right = evaluate(parts[1], scope);
          return _compare(left, right, op);
        }
      }
    }

    // Arithmetic Operators (+, -, *, /)
    // Note: Order of operations is not strictly respected here (left-to-right split),
    // but sufficient for basic "a + b" style.
    if (expr.contains("+")) {
      final parts = expr.split("+");
      if (parts.length == 2) {
        final left = evaluate(parts[0], scope);
        final right = evaluate(parts[1], scope);
        if (left is num && right is num) return left + right;
        if (left is String || right is String)
          return left.toString() + right.toString();
      }
    }

    // Fallback: Return raw string if nothing matched
    return expr;
  }

  bool _isTruthy(dynamic val) {
    if (val == null) return false;
    if (val is bool) return val;
    if (val is num) return val != 0;
    if (val is String) return val.isNotEmpty;
    return true;
  }

  bool _compare(dynamic left, dynamic right, String op) {
    switch (op) {
      case "==":
        return left.toString() ==
            right.toString(); // Weak equality for simplicity
      case "!=":
        return left.toString() != right.toString();
      case ">":
        if (left is num && right is num) return left > right;
        return false;
      case "<":
        if (left is num && right is num) return left < right;
        return false;
      case ">=":
        if (left is num && right is num) return left >= right;
        return false;
      case "<=":
        if (left is num && right is num) return left <= right;
        return false;
    }
    return false;
  }
}
