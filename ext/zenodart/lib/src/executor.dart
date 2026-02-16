import 'node.dart';

typedef Handler = void Function(Node node, Executor executor);

class Executor {
  final Map<String, Handler> _handlers = {};

  Executor() {
    registerHandler('log', _handleLog);
    registerHandler('print', _handleLog);
  }

  void registerHandler(String name, Handler handler) {
    _handlers[name] = handler;
  }

  void execute(Node node) {
    // If root, traverse children
    if (node.name == "root") {
      for (final child in node.children) {
        execute(child);
      }
      return;
    }

    final handler = _handlers[node.name];
    if (handler != null) {
      handler(node, this);
    } else {
      // Default traversal
      for (final child in node.children) {
        execute(child);
      }
    }
  }

  void _handleLog(Node node, Executor executor) {
    // Simple log implementation
    // In real Zeno, this would evaluate expressions
    print(node.value);
  }
}
