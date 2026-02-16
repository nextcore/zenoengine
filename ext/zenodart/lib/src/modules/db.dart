import 'package:sqlite3/sqlite3.dart';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

class DbModule {
  static Database? _db;

  static void register(Executor executor) {
    executor.registerHandler('db.connect', _handleConnect);
    executor.registerHandler('db.query', _handleQuery);
    executor.registerHandler('db.execute', _handleExecute);
    executor.registerHandler('db.close', _handleClose);
  }

  static Future<void> _handleConnect(
      Node node, Scope scope, Executor executor) async {
    final pathRaw = executor.evaluateExpression(node.value, scope);
    final path = pathRaw?.toString();

    if (path == null) {
      // In-memory
      _db = sqlite3.openInMemory();
      print("Connected to in-memory DB");
    } else {
      _db = sqlite3.open(path);
      print("Connected to DB: $path");
    }
  }

  static Future<void> _handleExecute(
      Node node, Scope scope, Executor executor) async {
    if (_db == null) {
      print("Error: DB not connected. Call db.connect first.");
      return;
    }

    final sql = executor.evaluateExpression(node.value, scope)?.toString();
    if (sql == null) return;

    List<Object?> params = [];
    for (final child in node.children) {
      if (child.name == 'args') {
        final args = executor.evaluateExpression(child.value, scope);
        if (args is List) {
          params = args;
        }
      }
    }

    try {
      _db!.execute(sql, params);
    } catch (e) {
      print("db.execute error: $e");
    }
  }

  static Future<void> _handleQuery(
      Node node, Scope scope, Executor executor) async {
    if (_db == null) {
      print("Error: DB not connected. Call db.connect first.");
      return;
    }

    final sql = executor.evaluateExpression(node.value, scope)?.toString();
    if (sql == null) return;

    List<Object?> params = [];
    String? asVar;

    for (final child in node.children) {
      if (child.name == 'args') {
        final args = executor.evaluateExpression(child.value, scope);
        if (args is List) {
          params = args;
        }
      } else if (child.name == 'as') {
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

    try {
      final ResultSet resultSet = _db!.select(sql, params);
      // Convert to List<Map>
      final List<Map<String, dynamic>> rows = [];
      for (final row in resultSet) {
        rows.add(Map<String, dynamic>.from(row));
      }

      if (asVar != null) {
        if ((asVar.startsWith('"') && asVar.endsWith('"')) ||
            (asVar.startsWith("'") && asVar.endsWith("'"))) {
          asVar = asVar.substring(1, asVar.length - 1);
        }
        scope.set(asVar, rows);
      } else {
        scope.set('rows', rows);
      }
    } catch (e) {
      print("db.query error: $e");
    }
  }

  static Future<void> _handleClose(
      Node node, Scope scope, Executor executor) async {
    _db?.dispose();
    _db = null;
    print("DB Closed");
  }
}
