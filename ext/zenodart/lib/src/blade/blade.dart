import '../node.dart';

class BladeCompiler {
  // Compiles a template string into a Zeno AST Node (root)
  static Node compile(String template, String filename) {
    final root = Node(name: "root", filename: filename);
    // Helper to keep track of stack for nesting (if/foreach/section)
    final stack = <Node>[root];

    // Naive parsing: Split by tags
    // We want to capture:
    // 1. {{ ... }} -> Interpolation
    // 2. @if(...) ... @endif
    // 3. @foreach(...) ... @endforeach
    // 4. @extends(...)
    // 5. @section(...) ... @endsection
    // 6. @yield(...)
    // 7. Raw Text

    // Regex for tokens.
    // Group 1: @directive(...)
    // Group 2: {{ ... }}
    final reg = RegExp(r'(@[a-zA-Z]+\s*\(.*?\)|@[a-zA-Z]+)|(\{\{.*?\}\})',
        dotAll: true);

    int lastIndex = 0;

    // We iterate through matches
    for (final match in reg.allMatches(template)) {
      // 1. Text before match
      if (match.start > lastIndex) {
        final text = template.substring(lastIndex, match.start);
        if (text.isNotEmpty) {
          _addWriteNode(stack.last, text, filename);
        }
      }

      final token = match.group(0)!;

      if (token.startsWith('{{')) {
        // Interpolation: {{ expr }}
        final expr = token.substring(2, token.length - 2).trim();
        // Create a write node, but value is expression to evaluate
        // We need a special node type or name for dynamic write.
        // Let's use `__native_write_expr`
        final node =
            Node(name: "__native_write_expr", value: expr, filename: filename);
        stack.last.children.add(node);
        node.parent = stack.last;
      } else if (token.startsWith('@')) {
        _handleDirective(token, stack, filename);
      }

      lastIndex = match.end;
    }

    // Trailing text
    if (lastIndex < template.length) {
      final text = template.substring(lastIndex);
      _addWriteNode(stack.last, text, filename);
    }

    return root;
  }

  static void _addWriteNode(Node parent, String text, String filename) {
    final node = Node(name: "__native_write", value: text, filename: filename);
    parent.children.add(node);
    node.parent = parent;
  }

  static void _handleDirective(
      String token, List<Node> stack, String filename) {
    // Parse directive name and args
    // e.g. @if($x) -> name: if, args: $x

    final parenStart = token.indexOf('(');
    String name;
    String? args;

    if (parenStart != -1 && token.endsWith(')')) {
      name = token.substring(1, parenStart).trim();
      args = token.substring(parenStart + 1, token.length - 1).trim();
    } else {
      name = token.substring(1).trim(); // e.g. @endif
    }

    if (name == 'if') {
      // Map to Zeno 'if' node
      // value = condition
      final node = Node(name: "if", value: args, filename: filename);
      // Implicit 'then' block for children?
      // Zeno 'if' executor expects children directly or 'then'.
      // Let's use direct children for now.
      stack.last.children.add(node);
      node.parent = stack.last;
      stack.add(node); // Push to stack
    } else if (name == 'else') {
      // 'else' is a sibling of 'if' children usually, or a child of 'if'?
      // In Zeno AST, 'else' is a child of 'if'.
      // So we need to pop the current 'if' (which was implicit 'then')?
      // Wait, if we pushed 'if' to stack, its children are the 'true' block.
      // When we see '@else', we should close the 'true' block?
      // Zeno 'if' handler: iterates children.
      // If we add 'else' node to 'if', its children are the else block.
      // But any subsequent nodes in 'if' are 'true' block?
      // No, Zeno structure is:
      // if: ... {
      //    log: "true"
      //    else: { log: "false" }
      // }
      // The executor executes non-else/then children if true.
      // So if we encounter @else, we should stop adding to 'if' directly?
      // We should create an 'else' node, add it to 'if', and push 'else' to stack.
      // But first we must stop adding to 'if'?
      // Actually, if we add 'else' node, subsequent text will be added to 'else' node (since it's on top of stack).
      // Yes.

      // Verify parent is 'if'
      if (stack.last.name != 'if') {
        // Error or maybe nested?
      }
      final node = Node(name: "else", filename: filename);
      stack.last.children.add(node);
      node.parent = stack.last;

      // We DON'T pop 'if' yet because 'else' is part of it.
      // But we want new content to go into 'else'.
      // If we push 'else', next writes go to 'else'.
      // But 'else' is inside 'if'.
      // When we hit @endif, we need to pop 'else' AND 'if'?
      // OR: @else replaces the top of stack?
      // 'if' remains the parent of 'else'.

      // Let's treat 'if' as the container.
      // When parsing 'if' block:
      // children... -> implicit true
      // @else -> add 'else' node.
      // subsequent children -> added to 'else' node?

      // Yes, push 'else'.
      // Stack: [Root, If, Else]
      stack.add(node);
    } else if (name == 'endif') {
      // Pop until we find 'if'.
      // If we are in 'else', pop 'else' then 'if'.
      // If in 'if', pop 'if'.
      if (stack.last.name == 'else') {
        stack.removeLast();
      }
      if (stack.last.name == 'if') {
        stack.removeLast();
      }
    } else if (name == 'foreach') {
      // @foreach($list as $item) or @foreach($list)
      // Zeno: foreach: $list
      // If syntax is `$list as $item`, we need to handle it.
      // ZenoDart executor handles `foreach: $list` and sets `$it`.
      // If blade uses `as`, we might need to shim it.

      String iterable = args ?? "";
      if (iterable.toLowerCase().contains(' as ')) {
        // Split logic if needed, but for now assuming simple usage or Zeno logic compatibility.
        // Let's just use the first part as iterable.
        iterable = iterable.split(' as ')[0].trim();
      }

      final node = Node(name: "foreach", value: iterable, filename: filename);
      stack.last.children.add(node);
      node.parent = stack.last;
      stack.add(node);
    } else if (name == 'endforeach') {
      if (stack.last.name == 'foreach') {
        stack.removeLast();
      }
    } else if (name == 'extends') {
      // @extends('layout')
      // Add a special meta node?
      // Or just standard node.
      String layout = args ?? "";
      // Strip quotes
      if ((layout.startsWith("'") && layout.endsWith("'")) ||
          (layout.startsWith('"') && layout.endsWith('"'))) {
        layout = layout.substring(1, layout.length - 1);
      }
      final node = Node(name: "extends", value: layout, filename: filename);
      stack.last.children.add(node);
      node.parent = stack.last;
    } else if (name == 'section') {
      // @section('name')
      String secName = args ?? "";
      if ((secName.startsWith("'") && secName.endsWith("'")) ||
          (secName.startsWith('"') && secName.endsWith('"'))) {
        secName = secName.substring(1, secName.length - 1);
      }
      final node = Node(name: "section", value: secName, filename: filename);
      stack.last.children.add(node);
      node.parent = stack.last;
      stack.add(node);
    } else if (name == 'endsection') {
      if (stack.last.name == 'section') {
        stack.removeLast();
      }
    } else if (name == 'yield') {
      String secName = args ?? "";
      if ((secName.startsWith("'") && secName.endsWith("'")) ||
          (secName.startsWith('"') && secName.endsWith('"'))) {
        secName = secName.substring(1, secName.length - 1);
      }
      final node = Node(name: "yield", value: secName, filename: filename);
      stack.last.children.add(node);
      node.parent = stack.last;
    }
  }
}
