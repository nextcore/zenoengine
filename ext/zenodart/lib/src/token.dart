enum TokenType {
  identifier,
  colon,
  string,
  lBrace,
  rBrace,
  eof,
  error,
}

class Token {
  final TokenType type;
  final String literal;
  final int line;
  final int column;

  const Token(this.type, this.literal, this.line, this.column);

  @override
  String toString() {
    return 'Token($type, "$literal", line: $line, col: $column)';
  }
}
