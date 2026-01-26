package engine

import (
	"unicode"
)

type TokenType string

const (
	TokenIdentifier TokenType = "IDENTIFIER"
	TokenColon      TokenType = "COLON"
	TokenString     TokenType = "STRING"
	TokenLBrace     TokenType = "LBRACE"
	TokenRBrace     TokenType = "RBRACE"
	TokenEOF        TokenType = "EOF"
	TokenError      TokenType = "ERROR"
)

type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

type Lexer struct {
	input        string
	position     int  // posisi saat ini di input (karakter saat ini)
	readPosition int  // posisi membaca saat ini (setelah karakter saat ini)
	ch           byte // karakter yang sedang diperiksa
	line         int
	col          int
}

func NewLexer(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
	l.col++
}

func (l *Lexer) NextToken() Token {
	var tok Token

	l.skipWhitespaceAndComments()

	switch l.ch {
	case ':':
		tok = l.newToken(TokenColon, string(l.ch))
	case '"', '\'':
		tok.Line = l.line
		tok.Column = l.col
		tok.Type = TokenString
		tok.Literal = l.readString(l.ch)
		return tok
	case 0:
		tok.Type = TokenEOF
		tok.Literal = ""
		tok.Line = l.line
		tok.Column = l.col
	default:
		if isLetter(l.ch) || isDigit(l.ch) || l.ch == '$' || l.ch == '.' || l.ch == '_' || l.ch == '/' || l.ch == '*' || l.ch == '!' || l.ch == '=' || l.ch == '<' || l.ch == '>' || l.ch == '(' || l.ch == ')' || l.ch == '+' || l.ch == '-' || l.ch == '%' || l.ch == '{' || l.ch == '}' {
			tok.Line = l.line
			tok.Column = l.col
			tok.Literal = l.readIdentifier()
			if tok.Literal == "{" {
				tok.Type = TokenLBrace
			} else if tok.Literal == "}" {
				tok.Type = TokenRBrace
			} else {
				tok.Type = TokenIdentifier
			}
			return tok
		} else {
			tok = l.newToken(TokenError, string(l.ch))
		}
	}

	l.readChar()
	return tok
}

func (l *Lexer) newToken(tokenType TokenType, ch string) Token {
	return Token{Type: tokenType, Literal: ch, Line: l.line, Column: l.col}
}

func (l *Lexer) readIdentifier() string {
	position := l.position
	for isLetter(l.ch) || isDigit(l.ch) || l.ch == '$' || l.ch == '.' || l.ch == '_' || l.ch == '-' || l.ch == '/' || l.ch == '*' || l.ch == '!' || l.ch == '=' || l.ch == '<' || l.ch == '>' || l.ch == '(' || l.ch == ')' || l.ch == '+' || l.ch == '%' || l.ch == '{' || l.ch == '}' {
		l.readChar()
	}
	return l.input[position:l.position]
}

func (l *Lexer) readString(quote byte) string {
	l.readChar() // skip starting quote
	str := ""

	for {
		if l.ch == quote || l.ch == 0 {
			break
		}
		if l.ch == '\\' {
			l.readChar()
			switch l.ch {
			case 'n':
				str += "\n"
			case 't':
				str += "\t"
			case 'r':
				str += "\r"
			case '"':
				str += "\""
			case '\'':
				str += "'"
			case '\\':
				str += "\\"
			default:
				str += "\\" + string(l.ch)
			}
		} else {
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			str += string(l.ch)
		}
		l.readChar()
	}

	// l.ch is now the closing quote or EOF
	if l.ch == quote {
		l.readChar() // consume closing quote
	}
	return str
}

func (l *Lexer) skipWhitespaceAndComments() {
	for {
		if unicode.IsSpace(rune(l.ch)) {
			if l.ch == '\n' {
				l.line++
				l.col = 0
			}
			l.readChar()
			continue
		}

		// Comments
		if l.ch == '/' && l.peekChar() == '/' {
			l.skipLineComment()
			continue
		}
		if l.ch == '#' {
			l.skipLineComment()
			continue
		}

		break
	}
}

func (l *Lexer) skipLineComment() {
	for l.ch != '\n' && l.ch != 0 {
		l.readChar()
	}
	if l.ch == '\n' {
		l.line++
		l.col = 0
		l.readChar()
	}
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z'
}

func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

func (l *Lexer) GetLineInfo() (int, int) {
	return l.line, l.col
}

func (l *Lexer) PeekToken() Token {
	pos := l.position
	readPos := l.readPosition
	ch := l.ch
	line := l.line
	col := l.col

	tok := l.NextToken()

	l.position = pos
	l.readPosition = readPos
	l.ch = ch
	l.line = line
	l.col = col
	return tok
}
