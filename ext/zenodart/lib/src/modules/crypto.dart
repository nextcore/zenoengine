import 'dart:convert';
import 'package:crypto/crypto.dart';
import 'package:uuid/uuid.dart';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class CryptoModule {
  static void register(Executor executor) {
    executor.registerHandler('hash.sha256', _handleSha256);
    executor.registerHandler('uuid.v4', _handleUuidV4);
    executor.registerHandler('base64.encode', _handleBase64Encode);
    executor.registerHandler('base64.decode', _handleBase64Decode);
  }

  static Future<void> _handleSha256(
      Node node, Scope scope, Executor executor) async {
    final inputRaw = executor.evaluateExpression(node.value, scope);
    if (inputRaw == null) return;

    final input = inputRaw.toString();
    final bytes = utf8.encode(input);
    final digest = sha256.convert(bytes);

    _setResult(node, scope, executor, digest.toString());
  }

  static Future<void> _handleUuidV4(
      Node node, Scope scope, Executor executor) async {
    final uid = Uuid().v4();
    _setResult(node, scope, executor, uid);
  }

  static Future<void> _handleBase64Encode(
      Node node, Scope scope, Executor executor) async {
    final inputRaw = executor.evaluateExpression(node.value, scope);
    if (inputRaw == null) return;

    final input = inputRaw.toString();
    final encoded = base64.encode(utf8.encode(input));
    _setResult(node, scope, executor, encoded);
  }

  static Future<void> _handleBase64Decode(
      Node node, Scope scope, Executor executor) async {
    final inputRaw = executor.evaluateExpression(node.value, scope);
    if (inputRaw == null) return;

    try {
      final decodedBytes = base64.decode(inputRaw.toString());
      final decoded = utf8.decode(decodedBytes);
      _setResult(node, scope, executor, decoded);
    } catch (e) {
      print("base64.decode error: $e");
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
