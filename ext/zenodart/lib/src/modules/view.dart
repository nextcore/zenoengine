import 'dart:io';
import '../executor.dart';
import '../node.dart';
import '../scope.dart';
import '../blade/blade.dart';

class ViewModule {
  static void register(Executor executor) {
    executor.registerHandler('view', _handleView);
    executor.registerHandler('__native_write', _handleWrite);
    executor.registerHandler('__native_write_expr', _handleWriteExpr);
    // extends, section, yield need handlers?
    // extends: usually loads layout.
    // section: captures content.
    // yield: outputs captured content.
    executor.registerHandler('extends', _handleExtends);
    executor.registerHandler('section', _handleSection);
    executor.registerHandler('yield', _handleYield);
  }

  static Future<void> _handleView(
      Node node, Scope scope, Executor executor) async {
    // Usage: view: "path/to/file.html" { data... }

    final pathRaw = executor.evaluateExpression(node.value, scope);
    final path = pathRaw?.toString();

    if (path == null) {
      print("Error: view missing path");
      return;
    }

    final file = File(path);
    if (!file.existsSync()) {
      print("Error: view file not found: $path");
      return;
    }

    // Check for data in children
    // If children are present, they are usually variables to pass to view.
    // view: "..." { name: "Zeno" }
    // Node has children. We can execute them in the view scope?
    // Or we can manually put them in scope.
    // Zeno logic: children of 'view' are executed in the NEW view scope.

    final viewScope = Scope(parent: scope);
    // Also bind children variables
    // But children might be 'extends', etc if inline? No, view loads file.
    // The children of the VIEW NODE are configuration/data.
    for (final child in node.children) {
      await executor.execute(child, viewScope);
    }

    try {
      final content = await file.readAsString();
      final root = BladeCompiler.compile(content, path);

      // Setup capturing buffer in scope if we want to return string?
      // Usually view outputs to HTTP response directly or stdout.
      // For CLI, stdout is fine.
      // For HTTP server, we might need to capture output.
      // Zeno Go: `view` writes to ResponseWriter if available, or buffer.

      // Let's attach a StringBuffer to the scope for "output".
      // If parent has buffer, use it?
      // ZenoDart ServerModule uses `body` variable.

      // Let's capture output into a string and print/set it.
      final buffer = StringBuffer();
      viewScope.set('__buffer', buffer);

      await executor.execute(root, viewScope);

      // Handle Layouts (extends)
      // If 'extends' was called, it sets '__layout' variable.
      // NOTE: '__layout' might be in viewScope, but we need to check carefully.
      final layoutPath = viewScope.get('__layout');
      // print("DEBUG: Layout path: $layoutPath");
      if (layoutPath != null) {
        // Render layout
        // We need to pass 'sections' to layout.
        // Sections are stored in '__sections' map.

        final layoutFile = File(layoutPath.toString());
        if (layoutFile.existsSync()) {
          final layoutContent = await layoutFile.readAsString();
          final layoutRoot =
              BladeCompiler.compile(layoutContent, layoutPath.toString());

          // New scope for layout, sharing sections
          final layoutScope = Scope(
              parent: viewScope); // Parent is viewScope to access sections
          // layoutScope.set('__buffer', buffer); // Continue writing to same buffer?
          // No, layout usually wraps. We should clear buffer and write layout?
          // Or write layout to NEW buffer?
          // The view content (non-section) is usually discarded if extends is used.
          // Sections are captured.

          final layoutBuffer = StringBuffer();
          layoutScope.set('__buffer', layoutBuffer);

          await executor.execute(layoutRoot, layoutScope);

          // If this was an HTTP request, set the body
          // Check if 'response' scope exists or similar.
          // For now, print to stdout or return?
          print(layoutBuffer.toString());

          // If inside server route, set 'body'
          // Scope chain: Layout -> View -> Route (has 'body')
          // We can traverse up or just use 'body' key.
          // But Route scope expects 'body' variable to be set.
          // Let's set 'body' in the closest scope that has it, or just root.
          // Since Scope.set sets in current scope, we might need to walk up?
          // ZenoDart Scope.set is local.
          // But we want to return the result.
          // In ServerModule, we execute steps. If step is 'view', it should set 'body'.
          // So:
          // viewScope.set('body', layoutBuffer.toString());
          // But viewScope is child of Route scope. Route scope won't see it unless we modify Scope.set to traverse?
          // Or we explicitly set parent variable.
          // For now, let's assume specific "return" logic or just print.
          // Update: Let's hack it: try to set on parent if parent has 'body'?
          // Or just assume `view` is the last statement and returns string?
          // Handlers are void.

          // Let's overwrite 'body' in parent if it exists.
          // Quick hack:
          var s = scope;
          bool handled = false;
          while (s.parent != null) {
            // Check if variable exists in this scope (direct check only, not get())
            // Scope class doesn't expose keys check publicly easily except via toMap
            // But we can check via get() and if it returns non-null.
            // 'body' in route scope is "" initially.
            if (s.get('body') != null || s.get('status') != null) {
              // Heuristic for route scope
              s.set('body', layoutBuffer.toString());
              handled = true;
              break;
            }
            s = s.parent!;
          }

          if (!handled) {
            print(layoutBuffer.toString());
          }

          return;
        }
      }

      print(buffer.toString());
      // Same hack for non-layout view
      var s = scope;
      while (s.parent != null) {
        if (s.get('body') != null || s.get('status') != null) {
          s.set('body', buffer.toString());
          break;
        }
        s = s.parent!;
      }
    } catch (e) {
      print("view error: $e");
    }
  }

  static Future<void> _handleWrite(
      Node node, Scope scope, Executor executor) async {
    final buffer = scope.get('__buffer') as StringBuffer?;
    if (buffer != null) {
      buffer.write(node.value);
    } else {
      // stdout.write(node.value);
    }
  }

  static Future<void> _handleWriteExpr(
      Node node, Scope scope, Executor executor) async {
    final val = executor.evaluateExpression(node.value, scope);
    final buffer = scope.get('__buffer') as StringBuffer?;
    if (buffer != null) {
      buffer.write(val);
    } else {
      // stdout.write(val);
    }
  }

  static Future<void> _handleExtends(
      Node node, Scope scope, Executor executor) async {
    final layout = executor.evaluateExpression(node.value, scope);
    scope.set('__layout', layout);
  }

  static Future<void> _handleSection(
      Node node, Scope scope, Executor executor) async {
    // Start capturing section
    // We need a separate buffer for this section.
    final secName = executor.evaluateExpression(node.value, scope)?.toString();
    if (secName == null) return;

    final sectionBuffer = StringBuffer();
    // Temporarily swap buffer
    final oldBuffer = scope.get('__buffer');
    scope.set('__buffer', sectionBuffer);

    // Execute children
    for (final child in node.children) {
      await executor.execute(child, scope);
    }

    // Restore buffer
    scope.set('__buffer', oldBuffer);

    // Save section
    Map<String, String> sections =
        scope.get('__sections') as Map<String, String>? ?? {};
    sections[secName] = sectionBuffer.toString();
    scope.set('__sections', sections);
  }

  static Future<void> _handleYield(
      Node node, Scope scope, Executor executor) async {
    final secName = executor.evaluateExpression(node.value, scope)?.toString();
    if (secName == null) return;

    final sections = scope.get('__sections')
        as Map<String, dynamic>?; // Scope get returns dynamic
    if (sections != null && sections.containsKey(secName)) {
      final content = sections[secName];
      final buffer = scope.get('__buffer') as StringBuffer?;
      buffer?.write(content);
    }
  }
}
