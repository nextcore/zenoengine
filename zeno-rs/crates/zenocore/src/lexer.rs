#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum TokenType {
    Identifier,
    Colon,
    String,
    LBrace,
    RBrace,
    EOF,
    Error,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Token {
    pub r#type: TokenType,
    pub literal: String,
    pub line: usize,
    pub col: usize,
}

pub struct Lexer<'a> {
    input: &'a [u8],
    position: usize,
    read_position: usize,
    ch: u8,
    line: usize,
    col: usize,
}

impl<'a> Lexer<'a> {
    pub fn new(input: &'a str) -> Self {
        let mut l = Lexer {
            input: input.as_bytes(),
            position: 0,
            read_position: 0,
            ch: 0,
            line: 1,
            col: 0,
        };
        l.read_char();
        l
    }

    fn read_char(&mut self) {
        if self.read_position >= self.input.len() {
            self.ch = 0;
        } else {
            self.ch = self.input[self.read_position];
        }
        self.position = self.read_position;
        self.read_position += 1;
        self.col += 1;
    }

    fn peek_char(&self) -> u8 {
        if self.read_position >= self.input.len() {
            0
        } else {
            self.input[self.read_position]
        }
    }

    pub fn next_token(&mut self) -> Token {
        self.skip_whitespace_and_comments();

        let mut tok = Token {
            r#type: TokenType::Error,
            literal: String::new(),
            line: self.line,
            col: self.col,
        };

        match self.ch {
            b':' => {
                tok = self.new_token(TokenType::Colon, (self.ch as char).to_string());
            }
            b'"' | b'\'' => {
                tok.line = self.line;
                tok.col = self.col;
                tok.r#type = TokenType::String;
                tok.literal = self.read_string(self.ch);
                return tok;
            }
            0 => {
                tok.r#type = TokenType::EOF;
                tok.literal = String::new();
                tok.line = self.line;
                tok.col = self.col;
            }
            _ => {
                if is_letter(self.ch) || is_digit(self.ch) || 
                   self.ch == b'$' || self.ch == b'.' || self.ch == b'_' || self.ch == b'/' || 
                   self.ch == b'*' || self.ch == b'!' || self.ch == b'=' || self.ch == b'<' || 
                   self.ch == b'>' || self.ch == b'(' || self.ch == b')' || self.ch == b'+' || 
                   self.ch == b'-' || self.ch == b'%' || self.ch == b'{' || self.ch == b'}' {
                    tok.line = self.line;
                    tok.col = self.col;
                    tok.literal = self.read_identifier();
                    if tok.literal == "{" {
                        tok.r#type = TokenType::LBrace;
                    } else if tok.literal == "}" {
                        tok.r#type = TokenType::RBrace;
                    } else {
                        tok.r#type = TokenType::Identifier;
                    }
                    return tok;
                } else {
                    tok = self.new_token(TokenType::Error, (self.ch as char).to_string());
                }
            }
        }

        self.read_char();
        tok
    }

    fn new_token(&self, token_type: TokenType, literal: String) -> Token {
        Token {
            r#type: token_type,
            literal,
            line: self.line,
            col: self.col,
        }
    }

    fn read_identifier(&mut self) -> String {
        let start = self.position;
        while is_letter(self.ch) || is_digit(self.ch) || 
              self.ch == b'$' || self.ch == b'.' || self.ch == b'_' || self.ch == b'-' || 
              self.ch == b'/' || self.ch == b'*' || self.ch == b'!' || self.ch == b'=' || 
              self.ch == b'<' || self.ch == b'>' || self.ch == b'(' || self.ch == b')' || 
              self.ch == b'+' || self.ch == b'%' || self.ch == b'{' || self.ch == b'}' {
            self.read_char();
        }
        let bytes = &self.input[start..self.position];
        String::from_utf8_lossy(bytes).into_owned()
    }

    fn read_string(&mut self, quote: u8) -> String {
        self.read_char(); // skip starting quote
        let mut bytes = Vec::new();

        while self.ch != quote && self.ch != 0 {
            if self.ch == b'\\' {
                self.read_char();
                match self.ch {
                    b'n' => bytes.push(b'\n'),
                    b't' => bytes.push(b'\t'),
                    b'r' => bytes.push(b'\r'),
                    b'"' => bytes.push(b'"'),
                    b'\'' => bytes.push(b'\''),
                    b'\\' => bytes.push(b'\\'),
                    _ => {
                        bytes.push(b'\\');
                        bytes.push(self.ch);
                    }
                }
            } else {
                if self.ch == b'\n' {
                    self.line += 1;
                    self.col = 0;
                }
                bytes.push(self.ch);
            }
            self.read_char();
        }

        // Consume closing quote
        if self.ch == quote {
            self.read_char();
        }
        String::from_utf8_lossy(&bytes).into_owned()
    }

    fn skip_whitespace_and_comments(&mut self) {
        loop {
            if self.ch.is_ascii_whitespace() {
                if self.ch == b'\n' {
                    self.line += 1;
                    self.col = 0;
                }
                self.read_char();
                continue;
            }

            // Comments
            if self.ch == b'/' && self.peek_char() == b'/' {
                self.skip_line_comment();
                continue;
            }
            if self.ch == b'#' {
                self.skip_line_comment();
                continue;
            }

            break;
        }
    }

    fn skip_line_comment(&mut self) {
        while self.ch != b'\n' && self.ch != 0 {
            self.read_char();
        }
        if self.ch == b'\n' {
            self.line += 1;
            self.col = 0;
            self.read_char();
        }
    }

    pub fn peek_token(&mut self) -> Token {
        let pos = self.position;
        let read_pos = self.read_position;
        let ch = self.ch;
        let line = self.line;
        let col = self.col;

        let tok = self.next_token();

        self.position = pos;
        self.read_position = read_pos;
        self.ch = ch;
        self.line = line;
        self.col = col;
        tok
    }
}

fn is_letter(ch: u8) -> bool {
    ch.is_ascii_alphabetic()
}

fn is_digit(ch: u8) -> bool {
    ch.is_ascii_digit()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_tokens() {
        let input = r#"
            log: "🚀 Started"
            http.get: '/' {
                view: 'welcome'
            }
        "#;

        let mut l = Lexer::new(input);
        let expected = vec![
            (TokenType::Identifier, "log"),
            (TokenType::Colon, ":"),
            (TokenType::String, "🚀 Started"),
            (TokenType::Identifier, "http.get"),
            (TokenType::Colon, ":"),
            (TokenType::String, "/"),
            (TokenType::LBrace, "{"),
            (TokenType::Identifier, "view"),
            (TokenType::Colon, ":"),
            (TokenType::String, "welcome"),
            (TokenType::RBrace, "}"),
            (TokenType::EOF, ""),
        ];

        for (exp_type, exp_lit) in expected {
            let tok = l.next_token();
            assert_eq!(tok.r#type, exp_type);
            assert_eq!(tok.literal, exp_lit);
        }
    }
}
