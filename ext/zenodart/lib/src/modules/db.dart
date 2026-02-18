import 'package:sqlite3/sqlite3.dart' as sqlite;
import 'package:mysql1/mysql1.dart' as mysql;
import '../executor.dart';
import '../node.dart';
import '../scope.dart';

abstract class DatabaseDriver {
  Future<void> execute(String sql, List<Object?> params);
  Future<List<Map<String, dynamic>>> query(String sql, List<Object?> params);
  Future<void> close();
}

class SqliteDriver implements DatabaseDriver {
  final sqlite.Database _db;
  SqliteDriver(this._db);

  @override
  Future<void> execute(String sql, List<Object?> params) async {
    _db.execute(sql, params);
  }

  @override
  Future<List<Map<String, dynamic>>> query(
      String sql, List<Object?> params) async {
    final resultSet = _db.select(sql, params);
    final List<Map<String, dynamic>> rows = [];
    for (final row in resultSet) {
      rows.add(Map<String, dynamic>.from(row));
    }
    return rows;
  }

  @override
  Future<void> close() async {
    _db.dispose();
  }
}

class MysqlDriver implements DatabaseDriver {
  final mysql.MySqlConnection _conn;
  MysqlDriver(this._conn);

  @override
  Future<void> execute(String sql, List<Object?> params) async {
    await _conn.query(sql, params);
  }

  @override
  Future<List<Map<String, dynamic>>> query(
      String sql, List<Object?> params) async {
    final results = await _conn.query(sql, params);
    final List<Map<String, dynamic>> rows = [];
    for (final row in results) {
      rows.add(row.fields);
    }
    return rows;
  }

  @override
  Future<void> close() async {
    await _conn.close();
  }
}

class DbModule {
  static DatabaseDriver? _driver;

  static void register(Executor executor) {
    executor.registerHandler('db.connect', _handleConnect);
    executor.registerHandler('db.query', _handleQuery);
    executor.registerHandler('db.execute', _handleExecute);
    executor.registerHandler('db.close', _handleClose);
  }

  static Future<void> _handleConnect(
      Node node, Scope scope, Executor executor) async {
    // Usage:
    // db.connect: "path.db" (Default Sqlite)
    // db.connect: {
    //    driver: "mysql"
    //    host: "localhost"
    //    port: 3306
    //    user: "root"
    //    password: "..."
    //    db: "mydb"
    // }

    // Check if value is string (Sqlite path) or map config
    dynamic config = executor.evaluateExpression(node.value, scope);

    // If value is null, check children for config map
    if (config == null && node.children.isNotEmpty) {
      // Build config map from children
      final map = <String, dynamic>{};
      for (final child in node.children) {
        final key = child.name;
        final val = executor.evaluateExpression(child.value, scope);
        map[key] = val;
      }
      config = map;
    }

    if (config is String) {
      // SQLite
      if (config.isEmpty) {
        _driver = SqliteDriver(sqlite.sqlite3.openInMemory());
        print("Connected to in-memory SQLite DB");
      } else {
        _driver = SqliteDriver(sqlite.sqlite3.open(config));
        print("Connected to SQLite DB: $config");
      }
    } else if (config is Map) {
      final driverType = config['driver']?.toString().toLowerCase() ?? 'sqlite';

      if (driverType == 'mysql' || driverType == 'mariadb') {
        final settings = mysql.ConnectionSettings(
          host: config['host']?.toString() ?? 'localhost',
          port: int.tryParse(config['port']?.toString() ?? '3306') ?? 3306,
          user: config['user']?.toString(),
          password: config['password']?.toString(),
          db: config['db']?.toString() ?? config['database']?.toString(),
        );
        try {
          final conn = await mysql.MySqlConnection.connect(settings);
          _driver = MysqlDriver(conn);
          print("Connected to MySQL/MariaDB: ${settings.db}");
        } catch (e) {
          print("MySQL connection error: $e");
        }
      } else {
        // Fallback to sqlite if path provided
        final path = config['path']?.toString() ?? config['file']?.toString();
        if (path != null) {
          _driver = SqliteDriver(sqlite.sqlite3.open(path));
          print("Connected to SQLite DB: $path");
        } else {
          _driver = SqliteDriver(sqlite.sqlite3.openInMemory());
          print("Connected to in-memory SQLite DB");
        }
      }
    } else {
      print("Error: db.connect invalid config");
    }
  }

  static Future<void> _handleExecute(
      Node node, Scope scope, Executor executor) async {
    if (_driver == null) {
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
      await _driver!.execute(sql, params);
    } catch (e) {
      print("db.execute error: $e");
    }
  }

  static Future<void> _handleQuery(
      Node node, Scope scope, Executor executor) async {
    if (_driver == null) {
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
      final rows = await _driver!.query(sql, params);

      if (asVar != null) {
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
    await _driver?.close();
    _driver = null;
    print("DB Closed");
  }
}
