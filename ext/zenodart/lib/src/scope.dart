class Scope {
  final Map<String, dynamic> _vars = {};
  final Scope? parent;

  Scope({this.parent});

  /// Sets a variable in the current scope.
  void set(String key, dynamic val) {
    _vars[key] = val;
  }

  /// Deletes a variable from the current scope.
  void delete(String key) {
    _vars.remove(key);
  }

  /// Retrieves a variable, supporting dot notation for nested maps.
  /// Returns null if not found.
  dynamic get(String key) {
    final (val, ok) = lookup(key);
    if (ok) return val;
    return null;
  }

  /// Helper that returns a record (value, exists) to handle null values correctly.
  (dynamic, bool) lookup(String key) {
    // 1. Check direct key match in current scope
    if (_vars.containsKey(key)) {
      return (_vars[key], true);
    }

    // 2. Check parent scope (recursively checks dot notation there too)
    if (parent != null) {
      final result = parent!.lookup(key);
      if (result.$2) return result;
    }

    // 3. Check Deep Navigation in THIS scope (if not found in parent)
    // Example: key="user.name". We check if "user" exists here.
    if (key.contains('.')) {
      final parts = key.split('.');
      final rootKey = parts[0];

      if (_vars.containsKey(rootKey)) {
        var current = _vars[rootKey];

        for (var i = 1; i < parts.length; i++) {
          // Basic map navigation
          if (current is Map) {
            final part = parts[i];
            if (current.containsKey(part)) {
              current = current[part];
            } else {
              return (null, false);
            }
          } else {
            // Not a map, can't go deeper
            return (null, false);
          }
        }
        return (current, true);
      }
    }

    return (null, false);
  }

  void reset() {
    _vars.clear();
  }

  Scope clone() {
    // Go implementation: NewScope(nil) -> creates disconnected scope?
    // "Clone creates a copy of the scope (Deep Copy Level 1)"
    // "It copies all variables from old scope to new scope".
    // "Used by router.go".
    // If it's used for request handling, it might want to isolate?
    // Go code: `newScope := NewScope(nil)` -> Parent is nil!

    final newScope = Scope(parent: null);
    newScope._vars.addAll(_vars);
    return newScope;
  }

  Map<String, dynamic> toMap() {
    return Map.from(_vars);
  }
}
