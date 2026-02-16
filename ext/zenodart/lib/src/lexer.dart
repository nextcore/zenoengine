import 'token.dart';

class Lexer {
  final String input;
  int _position = 0; // current position in input (current char)
  int _readPosition =
      0; // current reading position in input (after current char)
  int _ch = 0; // current char under examination
  int _line = 1;
  int _col = 0;

  Lexer(this.input) {
    _readChar();
  }

  void _readChar() {
    if (_readPosition >= input.length) {
      _ch = 0;
    } else {
      _ch = input.codeUnitAt(_readPosition);
    }
    _position = _readPosition;
    _readPosition++;
    _col++;
  }

  int _peekChar() {
    if (_readPosition >= input.length) {
      return 0;
    }
    return input.codeUnitAt(_readPosition);
  }

  Token nextToken() {
    Token tok;

    _skipWhitespaceAndComments();

    switch (_ch) {
      case 58: // ':'
        tok = _newToken(TokenType.colon, String.fromCharCode(_ch));
        break;
      case 34: // '"'
      case 39: // '\''
        final startLine = _line;
        final startCol = _col;
        final literal = _readString(_ch);
        tok = Token(TokenType.string, literal, startLine, startCol);
        // readString advances, so we should return immediately without readChar
        return tok;
      case 0:
        tok = Token(TokenType.eof, "", _line, _col);
        break;
      default:
        if (_isLetter(_ch) ||
                _isDigit(_ch) ||
                _ch == 36 || // '$'
                _ch == 46 || // '.'
                _ch == 95 || // '_'
                _ch == 47 || // '/'
                _ch == 42 || // '*'
                _ch == 33 || // '!'
                _ch == 61 || // '='
                _ch == 60 || // '<'
                _ch == 62 || // '>'
                _ch == 40 || // '('
                _ch == 41 || // ')'
                _ch == 43 || // '+'
                _ch == 45 || // '-'
                _ch == 37 || // '%'
                _ch == 123 || // '{'
                _ch == 125 // '}'
            ) {
          final startLine = _line;
          final startCol = _col;
          final literal = _readIdentifier();

          TokenType type = TokenType.identifier;
          if (literal == "{") {
            type = TokenType.lBrace;
          } else if (literal == "}") {
            type = TokenType.rBrace;
          }

          return Token(type, literal, startLine, startCol);
        } else {
          tok = _newToken(TokenType.error, String.fromCharCode(_ch));
        }
    }

    _readChar();
    return tok;
  }

  Token _newToken(TokenType type, String literal) {
    return Token(type, literal, _line, _col);
  }

  String _readIdentifier() {
    final startPos = _position;
    while (_isLetter(_ch) ||
            _isDigit(_ch) ||
            _ch == 36 || // '$'
            _ch == 46 || // '.'
            _ch == 95 || // '_'
            _ch ==
                45 || // '-' (wait, Go lexer includes '-' in identifier loop but maybe not in initial check? Let's check Go code again)
            _ch == 47 || // '/'
            _ch == 42 || // '*'
            _ch == 33 || // '!'
            _ch == 61 || // '='
            _ch == 60 || // '<'
            _ch == 62 || // '>'
            _ch == 40 || // '('
            _ch == 41 || // ')'
            _ch == 43 || // '+'
            _ch == 37 || // '%'
            _ch == 123 || // '{'
            _ch == 125 // '}'
        ) {
      _readChar();
    }
    return input.substring(startPos, _position);
  }

  String _readString(int quote) {
    _readChar(); // skip starting quote
    final buffer = StringBuffer();

    while (true) {
      if (_ch == quote || _ch == 0) {
        break;
      }
      if (_ch == 92) {
        // backslash
        _readChar();
        switch (_ch) {
          case 110: // 'n'
            buffer.write('\n');
            break;
          case 116: // 't'
            buffer.write('\t');
            break;
          case 114: // 'r'
            buffer.write('\r');
            break;
          case 34: // '"'
            buffer.write('"');
            break;
          case 39: // '\''
            buffer.write("'");
            break;
          case 92: // '\\'
            buffer.write('\\');
            break;
          default:
            buffer.write('\\');
            buffer.writeCharCode(_ch);
        }
      } else {
        if (_ch == 10) {
          // newline
          _line++;
          _col = 0;
        }
        buffer.writeCharCode(_ch);
      }
      _readChar();
    }

    if (_ch == quote) {
      _readChar(); // consume closing quote
    }

    return buffer.toString();
  }

  void _skipWhitespaceAndComments() {
    while (true) {
      if (_isSpace(_ch)) {
        if (_ch == 10) {
          // newline
          _line++;
          _col = 0;
        }
        _readChar();
        continue;
      }

      // Comments //
      if (_ch == 47 && _peekChar() == 47) {
        _skipLineComment();
        continue;
      }
      // Comments #
      if (_ch == 35) {
        _skipLineComment();
        continue;
      }

      break;
    }
  }

  void _skipLineComment() {
    while (_ch != 10 && _ch != 0) {
      _readChar();
    }
    if (_ch == 10) {
      _line++;
      _col = 0;
      _readChar();
    }
  }

  bool _isLetter(int ch) {
    return (ch >= 97 && ch <= 122) || (ch >= 65 && ch <= 90);
  }

  bool _isDigit(int ch) {
    return ch >= 48 && ch <= 57;
  }

  bool _isSpace(int ch) {
    // standard ascii whitespace
    return ch == 32 || ch == 9 || ch == 10 || ch == 13;
  }

  Token peekToken() {
    final pos = _position;
    final readPos = _readPosition;
    final ch = _ch;
    final line = _line;
    final col = _col;

    final tok = nextToken();

    _position = pos;
    _readPosition = readPos;
    _ch = ch;
    _line = line;
    _col = col;

    return tok;
  }
}
