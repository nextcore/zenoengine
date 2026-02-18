import 'dart:async';
import 'package:intl/intl.dart';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class TimeModule {
  static void register(Executor executor) {
    executor.registerHandler('time.now', _handleNow);
    executor.registerHandler('time.format', _handleFormat);
    executor.registerHandler('time.sleep', _handleSleep);
  }

  static Future<void> _handleNow(
      Node node, Scope scope, Executor executor) async {
    // Usage: time.now { as: "ts" }

    final now = DateTime.now().toIso8601String();

    String? asVar;
    // Check direct children for 'as'
    for (final child in node.children) {
      if (child.name == 'as') {
        // Use resolveValue to ensure we get a string from child value or its children (if any)
        // If child is `as: "ts"`, resolveValue returns "ts".
        // If child is `as: ts` and ts is not var, resolveValue might fail or return identifier if not handled?
        // Let's rely on evaluateExpression but also fallback to raw value string if simple identifier.

        dynamic val = executor.evaluateExpression(child.value, scope);
        // If evaluation returns null, but there IS a value (e.g. unquoted identifier), treat as string.
        if (val == null && child.value != null) {
          val = child.value.toString();
        }

        // If child.value is null, maybe child has children? (unlikely for 'as', but possible in complex config)
        if (val == null && child.children.isNotEmpty) {
          // For 'as', we don't expect children structure.
        }

        asVar = val?.toString();

        // Strip quotes if present
        if (asVar != null) {
          if ((asVar.startsWith('"') && asVar.endsWith('"')) ||
              (asVar.startsWith("'") && asVar.endsWith("'"))) {
            asVar = asVar.substring(1, asVar.length - 1);
          }
        }
      }
    }

    // Fallback: Check if node itself has children treated as Map?
    // No, evaluateExpression returns null for node with children.
    // But what if `time.now: { as: "ts" }`?
    // Parser creates: name="time.now", value=null, children=[{name: "as", value: "ts"}]
    // This looks correct.

    if (asVar != null) {
      // Remove quotes if present
      if ((asVar.startsWith('"') && asVar.endsWith('"')) ||
          (asVar.startsWith("'") && asVar.endsWith("'"))) {
        asVar = asVar.substring(1, asVar.length - 1);
      }
      scope.set(asVar, now);
    } else {
      scope.set('time_now', now);
    }
  }

  static Future<void> _handleFormat(
      Node node, Scope scope, Executor executor) async {
    // Usage: time.format: $ts { format: "yyyy-MM-dd", as: "formatted" }

    final timeRaw = executor.evaluateExpression(node.value, scope);
    if (timeRaw == null) return;

    DateTime? dt;
    try {
      dt = DateTime.parse(timeRaw.toString());
    } catch (e) {
      print("time.format error parsing date: $e");
      return;
    }

    String pattern = "yyyy-MM-dd HH:mm:ss";
    String? asVar;

    for (final child in node.children) {
      if (child.name == 'format') {
        pattern = executor.evaluateExpression(child.value, scope)?.toString() ??
            pattern;
      } else if (child.name == 'as') {
        asVar = executor.evaluateExpression(child.value, scope)?.toString();
      }
    }

    try {
      final fmt = DateFormat(pattern);
      final result = fmt.format(dt);

      if (asVar != null) {
        scope.set(asVar, result);
      } else {
        print(result);
      }
    } catch (e) {
      print("time.format error: $e");
    }
  }

  static Future<void> _handleSleep(
      Node node, Scope scope, Executor executor) async {
    // Usage: time.sleep: 1000 (ms)

    final msRaw = executor.evaluateExpression(node.value, scope);
    if (msRaw == null) return;

    final ms = int.tryParse(msRaw.toString()) ?? 0;
    if (ms > 0) {
      await Future.delayed(Duration(milliseconds: ms));
    }
  }
}
