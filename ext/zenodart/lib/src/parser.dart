import 'lexer.dart';
import 'token.dart';
import 'node.dart';

class Parser {
  final Lexer _lexer;
  final String _filename;

  Parser(this._lexer, {String filename = "unknown"}) : _filename = filename;

  Node parse() {
    final root = Node(name: "root", filename: _filename);
    final stack = <Node>[root];
    Node? lastNode;

    while (true) {
      var tok = _lexer.nextToken();
      if (tok.type == TokenType.eof) {
        break;
      }

      switch (tok.type) {
        case TokenType.identifier:
          final node = Node(
            name: tok.literal,
            line: tok.line,
            col: tok.column,
            filename: _filename,
          );
          final parent = stack.last;
          parent.children.add(node);
          node.parent = parent;
          lastNode = node;
          break;

        case TokenType.colon:
          final currentLine = tok.line;
          final valueParts = <String>[];

          while (true) {
            final peek = _lexer.peekToken();
            if (peek.type == TokenType.eof ||
                peek.line != currentLine ||
                peek.type == TokenType.lBrace ||
                peek.type == TokenType.colon) {
              break;
            }
            if (peek.type == TokenType.rBrace) {
              break;
            }

            tok = _lexer.nextToken();
            valueParts.add(tok.literal);
          }

          if (valueParts.isNotEmpty) {
            if (lastNode != null) {
              lastNode.value = valueParts.join(" ");
            }
          }

          final peek = _lexer.peekToken();
          if (peek.type == TokenType.lBrace) {
            _lexer.nextToken(); // consume {
            if (lastNode != null) {
              stack.add(lastNode);
            }
          } else if (peek.type == TokenType.rBrace) {
            // name: } (empty slot)
            // Check if it's strictly the next token or if we just stopped reading value because of it.
            // Go logic:
            // } else if peek.Type == TokenRBrace {
            //    l.NextToken()
            //    if len(stack) > 1 { stack = stack[:len(stack)-1] }
            // }
            // Wait, this consumes the rBrace immediately inside the Colon handler?
            // Yes.
            _lexer.nextToken(); // consume }
            if (stack.length > 1) {
              stack.removeLast();
            }
          }
          break;

        case TokenType.lBrace:
          if (lastNode != null) {
            stack.add(lastNode);
          } else {
            // Anonymous node
            final node = Node(
              name: "",
              line: tok.line,
              col: tok.column,
              filename: _filename,
            );
            final parent = stack.last;
            parent.children.add(node);
            node.parent = parent;
            stack.add(node);
          }
          break;

        case TokenType.rBrace:
          if (stack.length > 1) {
            stack.removeLast();
          }
          break;

        case TokenType.error:
          throw FormatException(
              "Lexical error: ${tok.literal} at line ${tok.line}:${tok.column}");

        default:
          // Ignore strings that appear without context?
          // Or maybe treat as value if following nothing?
          // Go parser only handles Identifier, Colon, LBrace, RBrace, Error.
          // String tokens are consumed inside Colon logic loop.
          // If a string appears at top level (e.g. "foo"), it's skipped/ignored?
          // Go switch has no default or case for String.
          // So String tokens at top level are effectively ignored (loop continues).
          break;
      }
    }

    return root;
  }
}
