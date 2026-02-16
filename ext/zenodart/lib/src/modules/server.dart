import 'dart:io';
import 'package:shelf/shelf.dart';
import 'package:shelf/shelf_io.dart' as shelf_io;
import 'package:shelf_router/shelf_router.dart';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class ServerModule {
  static void register(Executor executor) {
    executor.registerHandler('http.server', _handleServer);
  }

  static Future<void> _handleServer(
      Node node, Scope scope, Executor executor) async {
    // Usage:
    // http.server: {
    //    port: 3000
    //    routes: {
    //       get: "/hello" { return: "Hello" }
    //    }
    // }

    int port = 8080;
    final router = Router();

    // Config parsing
    for (final child in node.children) {
      if (child.name == 'port') {
        final p = executor.evaluateExpression(child.value, scope);
        if (p is int)
          port = p;
        else if (p is String) port = int.tryParse(p) ?? 8080;
      } else if (child.name == 'routes') {
        for (final routeNode in child.children) {
          final method = routeNode.name.toLowerCase();
          final pathRaw = executor.evaluateExpression(routeNode.value, scope);
          final path = pathRaw?.toString();

          if (path == null) continue;

          // Wrap the handler in a closure that shelf_router accepts
          Future<Response> handler(Request request) async {
            final reqScope = Scope(parent: scope);
            reqScope.set('method', request.method);
            reqScope.set('url', request.url.toString());
            reqScope.set('body', "");
            reqScope.set('status', 200);

            for (final step in routeNode.children) {
              if (step.name == 'return') {
                final val = executor.evaluateExpression(step.value, reqScope);
                reqScope.set('body', val);
              } else if (step.name == 'status') {
                final val = executor.evaluateExpression(step.value, reqScope);
                reqScope.set('status', val);
              } else {
                await executor.execute(step, reqScope);
              }
            }

            final body = reqScope.get('body')?.toString() ?? "";
            final status = reqScope.get('status');
            final statusInt = (status is int)
                ? status
                : int.tryParse(status?.toString() ?? "200") ?? 200;

            return Response(statusInt, body: body);
          }

          if (method == 'get')
            router.get(path, handler);
          else if (method == 'post')
            router.post(path, handler);
          else if (method == 'put')
            router.put(path, handler);
          else if (method == 'delete') router.delete(path, handler);
        }
      }
    }

    var handler =
        Pipeline().addMiddleware(logRequests()).addHandler(router.call);

    try {
      final server =
          await shelf_io.serve(handler, InternetAddress.anyIPv4, port);
      print('Server running on port ${server.port}');
    } catch (e) {
      print("Error starting server: $e");
    }
  }
}
