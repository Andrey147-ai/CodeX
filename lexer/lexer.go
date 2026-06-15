package lexer

import (
	"fmt"
	"unicode"
)

// TokenType represents the type of a lexical token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenNumber
	TokenString

	// Keywords
	TokenFn
	TokenStruct
	TokenDel
	TokenImport
	TokenPrint

	// Operators and delimiters
	TokenColonEquals // :=
	TokenIf
	TokenLBrace // {
	TokenRBrace // }
	TokenDot    // .
	TokenLParen // (
	TokenRParen // )
	TokenColon  // :
	TokenSemicolon
	TokenComma
	TokenArrow // ->
	TokenEquals

	// Arithmetic
	TokenPlus
	TokenMinus
	TokenStar
	TokenSlash
)

// Token represents a single lexical token with its type, literal value, and position.
type Token struct {
	Type    TokenType
	Literal string
	Line    int
	Column  int
}

func (t Token) String() string {
	return fmt.Sprintf("Token(%s, %q, line=%d, col=%d)", t.Type, t.Literal, t.Line, t.Column)
}

// Lexer holds the state of the lexer.
type Lexer struct {
	source  []rune
	pos     int
	line    int
	column  int
	current rune
}

// New creates a new Lexer for the given source code.
func New(source string) *Lexer {
	l := &Lexer{
		source: []rune(source),
		pos:    0,
		line:   1,
		column: 1,
	}
	if len(l.source) > 0 {
		l.current = l.source[0]
	}
	return l
}

// advance moves to the next character.
func (l *Lexer) advance() {
	l.pos++
	l.column++
	if l.pos < len(l.source) {
		l.current = l.source[l.pos]
	} else {
		l.current = 0 // EOF
	}
}

// peek returns the next character without consuming it.
func (l *Lexer) peek() rune {
	if l.pos+1 < len(l.source) {
		return l.source[l.pos+1]
	}
	return 0
}

// skipWhitespace skips spaces, tabs, newlines, and comments.
func (l *Lexer) skipWhitespace() {
	for l.current != 0 {
		if l.current == ' ' || l.current == '\t' || l.current == '\r' {
			l.advance()
		} else if l.current == '\n' {
			l.line++
			l.column = 1
			l.advance()
		} else if l.current == '/' && l.peek() == '/' {
			// Single-line comment: skip to end of line
			for l.current != 0 && l.current != '\n' {
				l.advance()
			}
		} else {
			break
		}
	}
}

// readIdentifier reads an identifier or keyword.
func (l *Lexer) readIdentifier() string {
	start := l.pos
	for l.current != 0 && (unicode.IsLetter(l.current) || unicode.IsDigit(l.current) || l.current == '_') {
		l.advance()
	}
	return string(l.source[start:l.pos])
}

// readNumber reads a numeric literal (integer or float).
func (l *Lexer) readNumber() string {
	start := l.pos
	for l.current != 0 && unicode.IsDigit(l.current) {
		l.advance()
	}
	if l.current == '.' && unicode.IsDigit(l.peek()) {
		l.advance() // consume '.'
		for l.current != 0 && unicode.IsDigit(l.current) {
			l.advance()
		}
	}
	return string(l.source[start:l.pos])
}

// NextToken returns the next token from the source.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.current == 0 {
		return Token{Type: TokenEOF, Literal: "", Line: l.line, Column: l.column}
	}

	line := l.line
	col := l.column

	// Identifiers and keywords
	if unicode.IsLetter(l.current) || l.current == '_' {
		literal := l.readIdentifier()
		tokenType := lookupKeyword(literal)
		return Token{Type: tokenType, Literal: literal, Line: line, Column: col}
	}

	// Numbers
	if unicode.IsDigit(l.current) {
		literal := l.readNumber()
		return Token{Type: TokenNumber, Literal: literal, Line: line, Column: col}
	}

	// Multi-character operators
	if l.current == ':' && l.peek() == '=' {
		l.advance() // consume ':'
		l.advance() // consume '='
		return Token{Type: TokenColonEquals, Literal: ":=", Line: line, Column: col}
	}

	if l.current == '-' && l.peek() == '>' {
		l.advance() // consume '-'
		l.advance() // consume '>'
		return Token{Type: TokenArrow, Literal: "->", Line: line, Column: col}
	}

	// Single-character tokens
	switch l.current {
	case '{':
		l.advance()
		return Token{Type: TokenLBrace, Literal: "{", Line: line, Column: col}
	case '}':
		l.advance()
		return Token{Type: TokenRBrace, Literal: "}", Line: line, Column: col}
	case '(':
		l.advance()
		return Token{Type: TokenLParen, Literal: "(", Line: line, Column: col}
	case ')':
		l.advance()
		return Token{Type: TokenRParen, Literal: ")", Line: line, Column: col}
	case '.':
		l.advance()
		return Token{Type: TokenDot, Literal: ".", Line: line, Column: col}
	case ':':
		l.advance()
		return Token{Type: TokenColon, Literal: ":", Line: line, Column: col}
	case ';':
		l.advance()
		return Token{Type: TokenSemicolon, Literal: ";", Line: line, Column: col}
	case ',':
		l.advance()
		return Token{Type: TokenComma, Literal: ",", Line: line, Column: col}
	case '=':
		l.advance()
		return Token{Type: TokenEquals, Literal: "=", Line: line, Column: col}
	case '+':
		l.advance()
		return Token{Type: TokenPlus, Literal: "+", Line: line, Column: col}
	case '-':
		l.advance()
		return Token{Type: TokenMinus, Literal: "-", Line: line, Column: col}
	case '*':
		l.advance()
		return Token{Type: TokenStar, Literal: "*", Line: line, Column: col}
	case '/':
		l.advance()
		return Token{Type: TokenSlash, Literal: "/", Line: line, Column: col}
	}

	// Unknown character
	ch := l.current
	l.advance()
	return Token{Type: TokenEOF, Literal: string(ch), Line: line, Column: col}
}

// lookupKeyword checks if an identifier is a keyword and returns its token type.
func lookupKeyword(ident string) TokenType {
	switch ident {
	case "fn":
		return TokenFn
	case "struct":
		return TokenStruct
	case "del":
		return TokenDel
	case "import":
		return TokenImport
	case "if":
		return TokenIf
	case "print":
		return TokenPrint
	default:
		return TokenIdentifier
	}
}

// Tokenize returns all tokens from the source.
func (l *Lexer) Tokenize() []Token {
	var tokens []Token
	for {
		tok := l.NextToken()
		tokens = append(tokens, tok)
		if tok.Type == TokenEOF {
			break
		}
	}
	return tokens
}

// String returns a human-readable name for a token type.
func (t TokenType) String() string {
	switch t {
	case TokenEOF:
		return "EOF"
	case TokenIdentifier:
		return "IDENT"
	case TokenNumber:
		return "NUMBER"
	case TokenString:
		return "STRING"
	case TokenFn:
		return "FN"
	case TokenStruct:
		return "STRUCT"
	case TokenDel:
		return "DEL"
	case TokenImport:
		return "IMPORT"
	case TokenPrint:
		return "PRINT"
	case TokenColonEquals:
		return ":="
	case TokenIf:
		return "IF"
	case TokenLBrace:
		return "{"
	case TokenRBrace:
		return "}"
	case TokenDot:
		return "."
	case TokenLParen:
		return "("
	case TokenRParen:
		return ")"
	case TokenColon:
		return ":"
	case TokenSemicolon:
		return ";"
	case TokenComma:
		return ","
	case TokenArrow:
		return "->"
	case TokenEquals:
		return "="
	case TokenPlus:
		return "+"
	case TokenMinus:
		return "-"
	case TokenStar:
		return "*"
	case TokenSlash:
		return "/"
	default:
		return "UNKNOWN"
	}
}
