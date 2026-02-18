import 'dart:io';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class FileModule {
  static void register(Executor executor) {
    executor.registerHandler('file.read', _handleRead);
    executor.registerHandler('file.write', _handleWrite);
    executor.registerHandler('file.delete', _handleDelete);
  }

  static Future<void> _handleRead(
      Node node, Scope scope, Executor executor) async {
    final pathRaw = executor.evaluateExpression(node.value, scope);
    final path = pathRaw?.toString();

    if (path == null || path.isEmpty) return;

    final file = File(path);
    if (!await file.exists()) {
      print("Error: file.read file not found: $path");
      return;
    }

    try {
      final content = await file.readAsString();

      String? asVar;
      for (final child in node.children) {
        if (child.name == 'as') {
          asVar = executor.evaluateExpression(child.value, scope)?.toString();
        }
      }

      if (asVar != null) {
        scope.set(asVar, content);
      } else {
        scope.set('file', content);
      }
    } catch (e) {
      print("Error reading file $path: $e");
    }
  }

  static Future<void> _handleWrite(
      Node node, Scope scope, Executor executor) async {
    final pathRaw = executor.evaluateExpression(node.value, scope);
    final path = pathRaw?.toString();

    if (path == null || path.isEmpty) return;

    String? content;
    for (final child in node.children) {
      if (child.name == 'content') {
        content = executor.evaluateExpression(child.value, scope)?.toString();
      }
    }

    if (content == null) {
      print("Error: file.write missing content for $path");
      return;
    }

    try {
      final file = File(path);
      await file.writeAsString(content);
    } catch (e) {
      print("Error writing file $path: $e");
    }
  }

  static Future<void> _handleDelete(
      Node node, Scope scope, Executor executor) async {
    final pathRaw = executor.evaluateExpression(node.value, scope);
    final path = pathRaw?.toString();

    if (path == null || path.isEmpty) return;

    try {
      final file = File(path);
      if (await file.exists()) {
        await file.delete();
      }
    } catch (e) {
      print("Error deleting file $path: $e");
    }
  }
}
