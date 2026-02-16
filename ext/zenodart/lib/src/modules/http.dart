import 'package:http/http.dart' as http;
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class HttpModule {
  static void register(Executor executor) {
    executor.registerHandler('http.get', _handleGet);
  }

  static Future<void> _handleGet(
      Node node, Scope scope, Executor executor) async {
    // Use evaluateExpression to get the raw URL string, bypassing children-as-map logic
    final urlRaw = executor.evaluateExpression(node.value, scope);
    final url = urlRaw?.toString();

    if (url == null || url.isEmpty) {
      print("Error: http.get missing url");
      return;
    }

    try {
      final uri = Uri.parse(url);
      final response = await http.get(uri);

      final responseMap = {
        'statusCode': response.statusCode,
        'body': response.body,
      };

      final responseScope = Scope(parent: scope);
      responseScope.set('response', responseMap);

      bool hasThen = false;
      for (final child in node.children) {
        if (child.name == 'then') {
          hasThen = true;
          for (final grandChild in child.children) {
            await executor.execute(grandChild, responseScope);
          }
        }
      }

      if (!hasThen) {
        for (final child in node.children) {
          await executor.execute(child, responseScope);
        }
      }
    } catch (e) {
      print("http.get error: $e");
    }
  }
}
