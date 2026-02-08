use logos::Logos;

#[derive(Logos, Debug, PartialEq)]
#[logos(skip r"[ \t\n\f]+")]
pub enum Token {
    #[token("//", |lex| {
        let remainder = lex.remainder();
        if let Some(pos) = remainder.find('\n') {
            lex.bump(pos);
        } else {
            lex.bump(remainder.len());
        }
        logos::Skip
    })]
    Comment,

    #[token("print")]
    Print,

    #[token("let")]
    Let,

    #[token("fn")]
    Fn,

    #[token("return")]
    Return,

    #[token("if")]
    If,

    #[token("else")]
    Else,

    #[token("true")]
    True,

    #[token("false")]
    False,

    #[token("null")]
    Null,

    #[token("(")]
    LParen,

    #[token(")")]
    RParen,

    #[token("[")]
    LBracket,

    #[token("]")]
    RBracket,

    #[token(":")]
    Colon,

    #[token(",")]
    Comma,

    #[token("{")]
    LBrace,

    #[token("}")]
    RBrace,

    #[token(";")]
    Semicolon,

    #[token("==")]
    DoubleEquals,

    #[token("!=")]
    NotEquals,

    #[token("<")]
    LessThan,

    #[token(">")]
    GreaterThan,

    #[token("=")]
    Equals,

    #[token("+")]
    Plus,

    #[token("-")]
    Minus,

    #[token("*")]
    Star,

    #[token("/")]
    Slash,

    #[regex(r#""([^"\\]|\\["\\bnfrt]|u[a-fA-F0-9]{4})*""#, |lex| unescape_string(lex.slice()))]
    StringLiteral(String),

    #[regex(r"[0-9]+", |lex| lex.slice().parse::<i64>().ok())]
    Integer(Option<i64>),

    #[regex(r"[a-zA-Z_][a-zA-Z0-9_]*", |lex| lex.slice().to_string())]
    Identifier(String),
}

fn unescape_string(s: &str) -> Option<String> {
    if s.len() < 2 { return None; }
    let content = &s[1..s.len()-1];
    let mut res = String::with_capacity(content.len());
    let mut chars = content.chars();
    while let Some(c) = chars.next() {
        if c == '\\' {
            match chars.next() {
                Some('n') => res.push('\n'),
                Some('r') => res.push('\r'),
                Some('t') => res.push('\t'),
                Some('b') => res.push('\u{0008}'),
                Some('f') => res.push('\u{000c}'),
                Some('\\') => res.push('\\'),
                Some('"') => res.push('"'),
                Some('u') => {
                    let u1 = chars.next()?;
                    let u2 = chars.next()?;
                    let u3 = chars.next()?;
                    let u4 = chars.next()?;
                    let hex_str = format!("{}{}{}{}", u1, u2, u3, u4);
                    if let Ok(code) = u32::from_str_radix(&hex_str, 16) {
                         if let Some(ch) = std::char::from_u32(code) {
                             res.push(ch);
                         } else { return None; }
                    } else { return None; }
                }
                Some(other) => { res.push(other); }
                None => return None,
            }
        } else {
            res.push(c);
        }
    }
    Some(res)
}
