package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"
)

// ========== LEXER ==========

type TokenType int

const (
	TOK_EOF TokenType = iota
	TOK_IDENT
	TOK_NUMBER
	TOK_STRING
	TOK_ASSIGN
	TOK_EQ
	TOK_PLUS
	TOK_MINUS
	TOK_STAR
	TOK_SLASH
	TOK_LPAREN
	TOK_RPAREN
	TOK_LBRACE
	TOK_RBRACE
	TOK_COMMA
	TOK_SEMICOLON
	TOK_IF
	TOK_ELSE
	TOK_FN
	TOK_RETURN
	TOK_PRINT
	TOK_DOT
	TOK_COLON
	TOK_STRUCT
	TOK_DEL
	TOK_NEWLINE
	TOK_EQEQ
	TOK_NEQ
	TOK_LT
	TOK_GT
	TOK_LTE
	TOK_GTE
	TOK_AND
	TOK_OR
	TOK_NOT
	TOK_TRUE
	TOK_FALSE
	TOK_WHILE
	TOK_BREAK
	TOK_CONTINUE
	TOK_FOR
	TOK_LBRACK
	TOK_RBRACK
	TOK_IN
	TOK_IMPORT
	TOK_PERCENT
	TOK_DIVINT
	TOK_NIL
	TOK_TRY
	TOK_CATCH
)

type Token struct {
	Type  TokenType
	Value string
	Line  int
	Col   int
}

type Lexer struct {
	input  []rune
	pos    int
	tokens []Token
	line   int
	col    int
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input), pos: 0, line: 1, col: 1}
}

func (l *Lexer) push(tok TokenType, val string) {
	l.tokens = append(l.tokens, Token{tok, val, l.line, l.col})
}

func (l *Lexer) advancePos(n int) {
	for i := 0; i < n; i++ {
		if l.pos < len(l.input) && l.input[l.pos] == '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
		l.pos++
	}
}

func (l *Lexer) Tokenize() []Token {
	// пропускаем BOM в начале файла
	if l.pos < len(l.input) && l.input[l.pos] == '\uFEFF' {
		l.pos++
		l.col++
	}
	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if unicode.IsSpace(ch) {
			if ch == '\n' {
				l.push(TOK_NEWLINE, "\n")
			}
			l.advancePos(1)
			continue
		}

		if ch == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.advancePos(1)
			}
			continue
		}

		if ch == '"' {
			tokLine, tokCol := l.line, l.col
			l.advancePos(1)
			var sb strings.Builder
			for l.pos < len(l.input) && l.input[l.pos] != '"' {
				if l.input[l.pos] == '\\' && l.pos+1 < len(l.input) {
					nxt := l.input[l.pos+1]
					switch nxt {
					case 'n':
						sb.WriteRune('\n')
					case 't':
						sb.WriteRune('\t')
					case 'r':
						sb.WriteRune('\r')
					case 'e':
						sb.WriteRune('\x1b')
					case '\\':
						sb.WriteRune('\\')
					case '"':
						sb.WriteRune('"')
					default:
						sb.WriteRune('\\')
						sb.WriteRune(nxt)
					}
					l.advancePos(2)
				} else {
					sb.WriteRune(l.input[l.pos])
					l.advancePos(1)
				}
			}
			str := sb.String()
			l.tokens = append(l.tokens, Token{TOK_STRING, str, tokLine, tokCol})
			if l.pos < len(l.input) {
				l.advancePos(1)
			}
			continue
		}

		if unicode.IsLetter(ch) || ch == '_' {
			tokLine, tokCol := l.line, l.col
			start := l.pos
			for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
				l.advancePos(1)
			}
			word := string(l.input[start:l.pos])
			typ := TOK_IDENT
			switch word {
			case "if":
				typ = TOK_IF
			case "else":
				typ = TOK_ELSE
			case "fn":
				typ = TOK_FN
			case "return":
				typ = TOK_RETURN
			case "print":
				typ = TOK_PRINT
			case "struct":
				typ = TOK_STRUCT
			case "del":
				typ = TOK_DEL
			case "true":
				typ = TOK_TRUE
			case "false":
				typ = TOK_FALSE
			case "while":
				typ = TOK_WHILE
			case "break":
				typ = TOK_BREAK
			case "continue":
				typ = TOK_CONTINUE
			case "for":
				typ = TOK_FOR
			case "in":
				typ = TOK_IN
			case "div":
				typ = TOK_DIVINT
			case "nil":
				typ = TOK_NIL
			case "try":
				typ = TOK_TRY
			case "catch":
				typ = TOK_CATCH
			case "import":
				typ = TOK_IMPORT
			}
			l.tokens = append(l.tokens, Token{typ, word, tokLine, tokCol})
			continue
		}

		if unicode.IsDigit(ch) {
			tokLine, tokCol := l.line, l.col
			start := l.pos
			for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
				l.advancePos(1)
			}
			l.tokens = append(l.tokens, Token{TOK_NUMBER, string(l.input[start:l.pos]), tokLine, tokCol})
			continue
		}

		// двухсимвольные операторы — проверяем первыми
		if l.pos+1 < len(l.input) {
			two := string([]rune{ch, l.input[l.pos+1]})
			var typ TokenType = -1
			switch two {
			case ":=":
				typ = TOK_ASSIGN
			case "==":
				typ = TOK_EQEQ
			case "!=":
				typ = TOK_NEQ
			case "<=":
				typ = TOK_LTE
			case ">=":
				typ = TOK_GTE
			case "&&":
				typ = TOK_AND
			case "||":
				typ = TOK_OR
			}
			if typ != -1 {
				l.push(typ, two)
				l.advancePos(2)
				continue
			}
		}

		switch ch {
		case ':':
			l.push(TOK_COLON, ":")
			l.advancePos(1)
		case '=':
			l.push(TOK_EQ, "=")
			l.advancePos(1)
		case '!':
			l.push(TOK_NOT, "!")
			l.advancePos(1)
		case '<':
			l.push(TOK_LT, "<")
			l.advancePos(1)
		case '>':
			l.push(TOK_GT, ">")
			l.advancePos(1)
		case '+':
			l.push(TOK_PLUS, "+")
			l.advancePos(1)
		case '-':
			l.push(TOK_MINUS, "-")
			l.advancePos(1)
		case '*':
			l.push(TOK_STAR, "*")
			l.advancePos(1)
		case '%':
			l.push(TOK_PERCENT, "%")
			l.advancePos(1)
		case '/':
			l.push(TOK_SLASH, "/")
			l.advancePos(1)
		case '(':
			l.push(TOK_LPAREN, "(")
			l.advancePos(1)
		case ')':
			l.push(TOK_RPAREN, ")")
			l.advancePos(1)
		case '{':
			l.push(TOK_LBRACE, "{")
			l.advancePos(1)
		case '}':
			l.push(TOK_RBRACE, "}")
			l.advancePos(1)
		case ',':
			l.push(TOK_COMMA, ",")
			l.advancePos(1)
		case ';':
			l.push(TOK_SEMICOLON, ";")
			l.advancePos(1)
		case '[':
			l.push(TOK_LBRACK, "[")
			l.advancePos(1)
		case ']':
			l.push(TOK_RBRACK, "]")
			l.advancePos(1)
		case '.':
			l.push(TOK_DOT, ".")
			l.advancePos(1)
		default:
			fmt.Fprintf(os.Stderr, "Lexer error at %d:%d: unknown character '%c'\n", l.line, l.col, ch)
			os.Exit(1)
		}
	}
	l.tokens = append(l.tokens, Token{TOK_EOF, "", l.line, l.col})
	return l.tokens
}

// ========== AST ==========

type ASTNode interface{ isASTNode() }

type Program struct{ Statements []ASTNode }

func (p *Program) isASTNode() {}

type VarDecl struct {
	Name  string
	Value ASTNode
}

func (v *VarDecl) isASTNode() {}

type Assign struct {
	Name  string
	Value ASTNode
}

func (a *Assign) isASTNode() {}

type FieldAssign struct {
	Object string
	Field  string
	Value  ASTNode
}

func (f *FieldAssign) isASTNode() {}

type NumberLiteral struct{ Value float64 }

func (n *NumberLiteral) isASTNode() {}

type BoolLiteral struct{ Value bool }

func (b *BoolLiteral) isASTNode() {}

type NilLiteral struct{}

func (n *NilLiteral) isASTNode() {}

type StringLiteral struct{ Value string }

func (s *StringLiteral) isASTNode() {}

type Identifier struct{ Name string }

func (i *Identifier) isASTNode() {}

type BinaryOp struct {
	Left  ASTNode
	Op    string
	Right ASTNode
}

func (b *BinaryOp) isASTNode() {}

type UnaryOp struct {
	Op   string
	Expr ASTNode
}

func (u *UnaryOp) isASTNode() {}

type FuncCall struct {
	Name string
	Args []ASTNode
}

func (f *FuncCall) isASTNode() {}

type IfStatement struct {
	Condition  ASTNode
	Body       []ASTNode
	ElseBranch []ASTNode
}

func (i *IfStatement) isASTNode() {}

type WhileLoop struct {
	Condition ASTNode
	Body      []ASTNode
}

func (w *WhileLoop) isASTNode() {}

type BreakStmt struct{}

func (b *BreakStmt) isASTNode() {}

type ContinueStmt struct{}

func (c *ContinueStmt) isASTNode() {}

type ForLoop struct {
	Init ASTNode
	Cond ASTNode
	Post ASTNode
	Body []ASTNode
}

func (f *ForLoop) isASTNode() {}

type ArrayLiteral struct{ Elements []ASTNode }

func (a *ArrayLiteral) isASTNode() {}

type MapLiteral struct {
	Keys   []string
	Values []ASTNode
}

func (m *MapLiteral) isASTNode() {}

type IndexAccess struct {
	Target ASTNode
	Index  ASTNode
}

func (i *IndexAccess) isASTNode() {}

type IndexAssign struct {
	Target ASTNode
	Index  ASTNode
	Value  ASTNode
}

func (i *IndexAssign) isASTNode() {}

type SliceAccess struct {
	Target ASTNode
	Start  ASTNode
	End    ASTNode
}

func (s *SliceAccess) isASTNode() {}

type MethodCall struct {
	Receiver ASTNode
	Method   string
	Args     []ASTNode
}

func (m *MethodCall) isASTNode() {}

type ForIn struct {
	Var      string
	Iterable ASTNode
	Body     []ASTNode
}

func (f *ForIn) isASTNode() {}

type ImportStmt struct {
	Path string
}

func (i *ImportStmt) isASTNode() {}

type TryCatch struct {
	Body      []ASTNode
	CatchVar  string
	CatchBody []ASTNode
}

func (t *TryCatch) isASTNode() {}

type FuncDef struct {
	Name     string
	Params   []string
	Body     []ASTNode
	RecvName string
	RecvType string
}

func (f *FuncDef) isASTNode() {}

type FuncLit struct {
	Params []string
	Body   []ASTNode
}

func (f *FuncLit) isASTNode() {}

type ReturnStmt struct{ Value ASTNode }

func (r *ReturnStmt) isASTNode() {}

type StructDef struct {
	Name   string
	Fields []string
}

func (s *StructDef) isASTNode() {}

type StructLiteral struct {
	Name   string
	Values []ASTNode
}

func (s *StructLiteral) isASTNode() {}

type FieldAccess struct {
	Object ASTNode
	Field  string
}

func (f *FieldAccess) isASTNode() {}

type DelCall struct{ Target ASTNode }

func (d *DelCall) isASTNode() {}

// ========== PARSER ==========

type Parser struct {
	tokens      []Token
	pos         int
	structNames map[string]bool
}

func NewParser(tokens []Token) *Parser {
	p := &Parser{tokens: tokens, pos: 0, structNames: make(map[string]bool)}
	// прескан: запоминаем все имена структур, чтобы отличать
	// литерал Type{...} от блока после выражения (if cond {).
	for i := 0; i+1 < len(tokens); i++ {
		if tokens[i].Type == TOK_STRUCT && tokens[i+1].Type == TOK_IDENT {
			p.structNames[tokens[i+1].Value] = true
		}
	}
	return p
}

func (p *Parser) skipNewlines() {
	for p.pos < len(p.tokens) && p.tokens[p.pos].Type == TOK_NEWLINE {
		p.pos++
	}
}

func (p *Parser) peek() Token {
	p.skipNewlines()
	if p.pos >= len(p.tokens) {
		return Token{TOK_EOF, "", 0, 0}
	}
	return p.tokens[p.pos]
}

// peekRaw возвращает следующий токен БЕЗ пропуска переводов строк —
// для постфиксных конструкций, обязанных идти на той же строке
// (литерал структуры Name{...}, вызов f(...), индекс a[0]).
func (p *Parser) peekRaw() Token {
	if p.pos < len(p.tokens) {
		return p.tokens[p.pos]
	}
	return Token{TOK_EOF, "", 0, 0}
}

func (p *Parser) next() Token {
	p.skipNewlines()
	tok := p.peek()
	if tok.Type != TOK_EOF {
		p.pos++
	}
	return tok
}

func (p *Parser) expect(typ TokenType) Token {
	tok := p.next()
	if tok.Type != typ {
		fmt.Fprintf(os.Stderr, "Parser error at %d:%d: expected %v, got %v (%q)\n", tok.Line, tok.Col, typ, tok.Type, tok.Value)
		os.Exit(1)
	}
	return tok
}

func (p *Parser) ParseProgram() *Program {
	prog := &Program{}
	for p.peek().Type != TOK_EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
	}
	return prog
}

func (p *Parser) parseStatement() ASTNode {
	p.skipNewlines()
	tok := p.peek()

	switch tok.Type {
	case TOK_STRUCT:
		return p.parseStructDef()
	case TOK_FN:
		return p.parseFuncDef()
	case TOK_IF:
		return p.parseIf()
	case TOK_WHILE:
		return p.parseWhile()
	case TOK_FOR:
		return p.parseFor()
	case TOK_TRY:
		return p.parseTry()
	case TOK_IMPORT:
		p.next()
		pathTok := p.peek()
		if pathTok.Type != TOK_STRING {
			fmt.Fprintf(os.Stderr, "Parser error at %d:%d: import needs a string path\n", pathTok.Line, pathTok.Col)
			os.Exit(1)
		}
		p.next()
		return &ImportStmt{Path: pathTok.Value}
	case TOK_BREAK:
		p.next()
		return &BreakStmt{}
	case TOK_CONTINUE:
		p.next()
		return &ContinueStmt{}
	case TOK_RETURN:
		p.next()
		return &ReturnStmt{Value: p.parseExpr()}
	case TOK_DEL:
		return p.parseDelCall()
	case TOK_PRINT:
		p.next()
		return p.parseFuncCallFinish("print")
	}

	if tok.Type == TOK_IDENT {
		idx := p.pos + 1
		for idx < len(p.tokens) && p.tokens[idx].Type == TOK_NEWLINE {
			idx++
		}

		if p.isIndexAssign() {
			var target ASTNode = &Identifier{Name: p.next().Value}
			for {
				if p.peek().Type == TOK_LBRACK {
					p.next()
					target = p.parseIndexOrSliceTarget(target)
				} else if p.peek().Type == TOK_DOT {
					p.next()
					field := p.expect(TOK_IDENT).Value
					target = &FieldAccess{Object: target, Field: field}
				} else {
					break
				}
			}
			p.expect(TOK_EQ)
			val := p.parseExpr()
			if ia, ok := target.(*IndexAccess); ok {
				return &IndexAssign{Target: ia.Target, Index: ia.Index, Value: val}
			}
			fmt.Fprintf(os.Stderr, "Parser error: invalid assignment target\n")
			os.Exit(1)
		}

		if idx < len(p.tokens) {
			if p.tokens[idx].Type == TOK_ASSIGN {
				name := p.next().Value
				p.expect(TOK_ASSIGN)
				return &VarDecl{Name: name, Value: p.parseExpr()}
			}
			if p.tokens[idx].Type == TOK_EQ {
				name := p.next().Value
				p.expect(TOK_EQ)
				return &Assign{Name: name, Value: p.parseExpr()}
			}
			if p.tokens[idx].Type == TOK_DOT {
				if idx+2 < len(p.tokens) && p.tokens[idx+1].Type == TOK_IDENT && p.tokens[idx+2].Type == TOK_EQ {
					objName := p.next().Value
					p.expect(TOK_DOT)
					fieldName := p.next().Value
					p.expect(TOK_EQ)
					return &FieldAssign{Object: objName, Field: fieldName, Value: p.parseExpr()}
				}
			}
		}
	}

	return p.parseExpr()
}

func (p *Parser) parseStructDef() ASTNode {
	p.expect(TOK_STRUCT)
	name := p.expect(TOK_IDENT).Value
	p.expect(TOK_LBRACE)
	var fields []string
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		fieldName := p.expect(TOK_IDENT).Value
		fields = append(fields, fieldName)

		if p.peek().Type == TOK_COLON {
			p.next()
			p.expect(TOK_IDENT)
		}

		if p.peek().Type == TOK_COMMA {
			p.next()
		}
	}
	p.expect(TOK_RBRACE)
	return &StructDef{Name: name, Fields: fields}
}

func (p *Parser) parseFuncDef() ASTNode {
	p.expect(TOK_FN)
	fn := &FuncDef{}
	// fn (recv Type) name(...) — метод; fn (...) — анонимная функция
	if p.peek().Type == TOK_LPAREN {
		if p.isMethodReceiver() {
			p.next()
			fn.RecvName = p.expect(TOK_IDENT).Value
			fn.RecvType = p.expect(TOK_IDENT).Value
			p.expect(TOK_RPAREN)
		} else {
			params, body := p.parseFnRemainder()
			return &FuncLit{Params: params, Body: body}
		}
	}
	fn.Name = p.expect(TOK_IDENT).Value
	params, body := p.parseFnRemainder()
	fn.Params = params
	fn.Body = body
	return fn
}

// isMethodReceiver checks '(' IDENT IDENT ')' — receiver vs plain params.
func (p *Parser) isMethodReceiver() bool {
	idx := p.pos + 1
	for idx < len(p.tokens) && p.tokens[idx].Type == TOK_NEWLINE {
		idx++
	}
	if idx >= len(p.tokens) || p.tokens[idx].Type != TOK_IDENT {
		return false
	}
	idx++
	for idx < len(p.tokens) && p.tokens[idx].Type == TOK_NEWLINE {
		idx++
	}
	if idx >= len(p.tokens) || p.tokens[idx].Type != TOK_IDENT {
		return false
	}
	idx++
	for idx < len(p.tokens) && p.tokens[idx].Type == TOK_NEWLINE {
		idx++
	}
	return idx < len(p.tokens) && p.tokens[idx].Type == TOK_RPAREN
}

func (p *Parser) parseFnRemainder() ([]string, []ASTNode) {
	p.expect(TOK_LPAREN)
	var params []string
	for p.peek().Type != TOK_RPAREN && p.peek().Type != TOK_EOF {
		params = append(params, p.expect(TOK_IDENT).Value)
		if p.peek().Type == TOK_COMMA {
			p.next()
		}
	}
	p.expect(TOK_RPAREN)
	p.expect(TOK_LBRACE)
	var body []ASTNode
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		body = append(body, p.parseStatement())
	}
	p.expect(TOK_RBRACE)
	return params, body
}

func (p *Parser) parseIf() ASTNode {
	p.expect(TOK_IF)
	cond := p.parseExpr()
	p.expect(TOK_LBRACE)
	var body []ASTNode
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		body = append(body, p.parseStatement())
	}
	p.expect(TOK_RBRACE)

	var elseBranch []ASTNode
	if p.peek().Type == TOK_ELSE {
		p.next()
		// else if — цепочка без вложенного блока
		if p.peek().Type == TOK_IF {
			elseBranch = append(elseBranch, p.parseIf())
		} else {
			p.expect(TOK_LBRACE)
			for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
				elseBranch = append(elseBranch, p.parseStatement())
			}
			p.expect(TOK_RBRACE)
		}
	}
	return &IfStatement{Condition: cond, Body: body, ElseBranch: elseBranch}
}

func (p *Parser) parseWhile() ASTNode {
	start := p.expect(TOK_WHILE)
	_ = start
	cond := p.parseExpr()
	p.expect(TOK_LBRACE)
	var body []ASTNode
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		body = append(body, p.parseStatement())
	}
	p.expect(TOK_RBRACE)
	return &WhileLoop{Condition: cond, Body: body}
}

// isIndexAssign reports whether the statement starting at the current
// identifier is an indexed assignment like a[0] = v or s.f[1] = v.
func (p *Parser) isIndexAssign() bool {
	idx := p.pos + 1
	seenBracket := false
	for {
		for idx < len(p.tokens) && p.tokens[idx].Type == TOK_NEWLINE {
			idx++
		}
		if idx >= len(p.tokens) {
			return false
		}
		t := p.tokens[idx].Type
		if t == TOK_LBRACK {
			seenBracket = true
			depth := 0
			for idx < len(p.tokens) {
				if p.tokens[idx].Type == TOK_LBRACK {
					depth++
				}
				if p.tokens[idx].Type == TOK_RBRACK {
					depth--
					if depth == 0 {
						idx++
						break
					}
				}
				idx++
			}
			continue
		}
		if t == TOK_DOT {
			idx++
			for idx < len(p.tokens) && p.tokens[idx].Type == TOK_NEWLINE {
				idx++
			}
			if idx < len(p.tokens) && p.tokens[idx].Type == TOK_IDENT {
				idx++
				continue
			}
			return false
		}
		if t == TOK_EQ {
			return seenBracket
		}
		return false
	}
}

func (p *Parser) parseFor() ASTNode {
	p.expect(TOK_FOR)
	// for x in iterable { ... }
	if p.peek().Type == TOK_IDENT {
		save := p.pos
		p.next()
		isIn := p.peek().Type == TOK_IN
		p.pos = save
		if isIn {
			varName := p.next().Value
			p.expect(TOK_IN)
			iterable := p.parseExpr()
			p.expect(TOK_LBRACE)
			var body []ASTNode
			for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
				body = append(body, p.parseStatement())
			}
			p.expect(TOK_RBRACE)
			return &ForIn{Var: varName, Iterable: iterable, Body: body}
		}
	}
	var init, cond, post ASTNode
	if p.peek().Type != TOK_SEMICOLON {
		init = p.parseStatement()
	}
	p.expect(TOK_SEMICOLON)
	if p.peek().Type != TOK_SEMICOLON {
		cond = p.parseExpr()
	}
	p.expect(TOK_SEMICOLON)
	if p.peek().Type != TOK_LBRACE {
		post = p.parseStatement()
	}
	p.expect(TOK_LBRACE)
	var body []ASTNode
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		body = append(body, p.parseStatement())
	}
	p.expect(TOK_RBRACE)
	return &ForLoop{Init: init, Cond: cond, Post: post, Body: body}
}

func (p *Parser) parseTry() ASTNode {
	p.expect(TOK_TRY)
	p.expect(TOK_LBRACE)
	var body []ASTNode
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		body = append(body, p.parseStatement())
	}
	p.expect(TOK_RBRACE)
	p.expect(TOK_CATCH)
	catchVar := ""
	if p.peek().Type == TOK_IDENT {
		catchVar = p.next().Value
	}
	p.expect(TOK_LBRACE)
	var catchBody []ASTNode
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		catchBody = append(catchBody, p.parseStatement())
	}
	p.expect(TOK_RBRACE)
	return &TryCatch{Body: body, CatchVar: catchVar, CatchBody: catchBody}
}

func (p *Parser) parseDelCall() ASTNode {
	p.expect(TOK_DEL)
	p.expect(TOK_LPAREN)
	target := p.parseExpr()
	p.expect(TOK_RPAREN)
	return &DelCall{Target: target}
}

func (p *Parser) parseFuncCallFinish(name string) ASTNode {
	p.expect(TOK_LPAREN)
	var args []ASTNode
	for p.peek().Type != TOK_RPAREN {
		args = append(args, p.parseExpr())
		if p.peek().Type == TOK_COMMA {
			p.next()
		}
	}
	p.expect(TOK_RPAREN)
	return &FuncCall{Name: name, Args: args}
}

func (p *Parser) parseExpr() ASTNode {
	return p.parseOr()
}

func (p *Parser) parseOr() ASTNode {
	left := p.parseAnd()
	for p.peek().Type == TOK_OR {
		opTok := p.next()
		right := p.parseAnd()
		left = &BinaryOp{Left: left, Op: opTok.Value, Right: right}
	}
	return left
}

func (p *Parser) parseAnd() ASTNode {
	left := p.parseEquality()
	for p.peek().Type == TOK_AND {
		opTok := p.next()
		right := p.parseEquality()
		left = &BinaryOp{Left: left, Op: opTok.Value, Right: right}
	}
	return left
}

func (p *Parser) parseEquality() ASTNode {
	left := p.parseComparison()
	for {
		t := p.peek().Type
		if t != TOK_EQEQ && t != TOK_NEQ {
			break
		}
		opTok := p.next()
		right := p.parseComparison()
		left = &BinaryOp{Left: left, Op: opTok.Value, Right: right}
	}
	return left
}

func (p *Parser) parseComparison() ASTNode {
	left := p.parseAdd()
	for {
		t := p.peek().Type
		if t != TOK_LT && t != TOK_GT && t != TOK_LTE && t != TOK_GTE {
			break
		}
		opTok := p.next()
		right := p.parseAdd()
		left = &BinaryOp{Left: left, Op: opTok.Value, Right: right}
	}
	return left
}

func (p *Parser) parseAdd() ASTNode {
	left := p.parseMul()
	for p.peek().Type == TOK_PLUS || p.peek().Type == TOK_MINUS {
		opTok := p.next()
		right := p.parseMul()
		left = &BinaryOp{Left: left, Op: opTok.Value, Right: right}
	}
	return left
}

func (p *Parser) parseMul() ASTNode {
	left := p.parseUnary()
	for p.peek().Type == TOK_STAR || p.peek().Type == TOK_SLASH || p.peek().Type == TOK_PERCENT || p.peek().Type == TOK_DIVINT {
		opTok := p.next()
		right := p.parseUnary()
		left = &BinaryOp{Left: left, Op: opTok.Value, Right: right}
	}
	return left
}

func (p *Parser) parseUnary() ASTNode {
	t := p.peek().Type
	if t == TOK_NOT || t == TOK_MINUS {
		opTok := p.next()
		expr := p.parseUnary()
		return &UnaryOp{Op: opTok.Value, Expr: expr}
	}
	node := p.parsePrimary()
	if ident, ok := node.(*Identifier); ok {
		if p.peekRaw().Type == TOK_LPAREN {
			node = p.parseFuncCallFinish(ident.Name)
		}
	}
	// постфиксные цепочки: поля и индексы в любом порядке (a[0].hp, f()[1])
	for {
		if p.peek().Type == TOK_DOT {
			p.next()
			field := p.expect(TOK_IDENT).Value
			node = &FieldAccess{Object: node, Field: field}
			continue
		}
		if p.peekRaw().Type == TOK_LBRACK {
			p.next()
			node = p.parseIndexOrSlice(node)
			continue
		}
		// вызов метода: obj.method(args)
		if p.peekRaw().Type == TOK_LPAREN {
			if _, ok := node.(*FieldAccess); !ok {
				break
			}
			fa := node.(*FieldAccess)
			p.next()
			var args []ASTNode
			for p.peek().Type != TOK_RPAREN && p.peek().Type != TOK_EOF {
				args = append(args, p.parseExpr())
				if p.peek().Type == TOK_COMMA {
					p.next()
				}
			}
			p.expect(TOK_RPAREN)
			node = &MethodCall{Receiver: fa.Object, Method: fa.Field, Args: args}
			continue
		}
		break
	}
	return node
}

// parseIndexOrSlice parses the rest of [...] after '[' was consumed:
// a[i] (index) or a[start:end] (slice, either side omittable).
func (p *Parser) parseIndexOrSlice(target ASTNode) ASTNode {
	var start, end ASTNode
	if p.peek().Type != TOK_COLON {
		start = p.parseExpr()
	}
	if p.peek().Type == TOK_COLON {
		p.next()
		if p.peek().Type != TOK_RBRACK {
			end = p.parseExpr()
		}
		p.expect(TOK_RBRACK)
		return &SliceAccess{Target: target, Start: start, End: end}
	}
	p.expect(TOK_RBRACK)
	return &IndexAccess{Target: target, Index: start}
}

// parseIndexOrSliceTarget is the assignment-target twin: slices are
// read-only, so a[i:j] = v is a clean error instead of weird behavior.
func (p *Parser) parseIndexOrSliceTarget(target ASTNode) ASTNode {
	var start ASTNode
	if p.peek().Type != TOK_COLON {
		start = p.parseExpr()
	}
	if p.peek().Type == TOK_COLON {
		tok := p.peek()
		fmt.Fprintf(os.Stderr, "Parser error at %d:%d: slice assignment is not supported\n", tok.Line, tok.Col)
		os.Exit(1)
	}
	p.expect(TOK_RBRACK)
	return &IndexAccess{Target: target, Index: start}
}

func (p *Parser) parsePrimary() ASTNode {
	p.skipNewlines()
	tok := p.peek()

	if tok.Type == TOK_TRUE {
		p.next()
		return &BoolLiteral{Value: true}
	}
	if tok.Type == TOK_FN {
		return p.parseFuncDef()
	}
	if tok.Type == TOK_FALSE {
		p.next()
		return &BoolLiteral{Value: false}
	}
	if tok.Type == TOK_NIL {
		p.next()
		return &NilLiteral{}
	}
	if tok.Type == TOK_LBRACK {
		p.next()
		var elems []ASTNode
		for p.peek().Type != TOK_RBRACK && p.peek().Type != TOK_EOF {
			elems = append(elems, p.parseExpr())
			if p.peek().Type == TOK_COMMA {
				p.next()
			}
		}
		p.expect(TOK_RBRACK)
		return &ArrayLiteral{Elements: elems}
	}
	if tok.Type == TOK_LBRACE {
		p.next()
		var mapKeys []string
		var mapVals []ASTNode
		for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
			kt := p.peek()
			var key string
			if kt.Type == TOK_STRING || kt.Type == TOK_IDENT {
				key = p.next().Value
			} else {
				fmt.Fprintf(os.Stderr, "Parser error at %d:%d: map key must be a string, got %v\n", kt.Line, kt.Col, kt.Type)
				os.Exit(1)
			}
			p.expect(TOK_COLON)
			mapKeys = append(mapKeys, key)
			mapVals = append(mapVals, p.parseExpr())
			if p.peek().Type == TOK_COMMA {
				p.next()
			}
		}
		p.expect(TOK_RBRACE)
		return &MapLiteral{Keys: mapKeys, Values: mapVals}
	}
	if tok.Type == TOK_NUMBER {
		p.next()
		val, _ := strconv.ParseFloat(tok.Value, 64)
		return &NumberLiteral{Value: val}
	}
	if tok.Type == TOK_STRING {
		p.next()
		return &StringLiteral{Value: tok.Value}
	}
	if tok.Type == TOK_IDENT {
		name := p.next().Value

		// Литерал структуры только если Name — известная структура
		// и '{' на той же строке: иначе `price` перед блоком if/while
		// съедался бы как тип.
		if p.structNames[name] && p.peekRaw().Type == TOK_LBRACE {
			p.next()
			var vals []ASTNode
			for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
				vals = append(vals, p.parseExpr())
				if p.peek().Type == TOK_COMMA {
					p.next()
				}
			}
			p.expect(TOK_RBRACE)
			return &StructLiteral{Name: name, Values: vals}
		}

		var node ASTNode = &Identifier{Name: name}

		// Фикс: Если за идентификатором идёт точка — это чтение поля структуры прямо внутри математики!
		for p.peek().Type == TOK_DOT {
			p.next() // жрём '.'
			field := p.expect(TOK_IDENT).Value
			node = &FieldAccess{Object: node, Field: field}
		}

		return node
	}
	if tok.Type == TOK_LPAREN {
		p.next()
		expr := p.parseExpr()
		p.expect(TOK_RPAREN)
		return expr
	}

	fmt.Fprintf(os.Stderr, "Parser error at %d:%d: unexpected token %v (%q)\n", tok.Line, tok.Col, tok.Type, tok.Value)
	os.Exit(1)
	return nil
}

// ========== INTERPRETER ==========

// общий stdin-ридер для input(): пересоздание bufio на каждый вызов
// теряло бы уже забуферизованные данные
var stdinReader = bufio.NewReader(os.Stdin)

type Value struct {
	Kind     string
	NumVal   float64
	BoolVal  bool
	StrVal   string
	Items    []Value
	MapVal   map[string]Value
	Fields   map[string]Value
	TypeName string
	Fn       *FuncDef
	Closure  *Environment
}

type Environment struct {
	vars      map[string]Value
	funcs     map[string]*FuncDef
	structs   map[string]*StructDef
	methods   map[string]map[string]*FuncDef
	protected map[string]bool
	parent    *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:      make(map[string]Value),
		funcs:     make(map[string]*FuncDef),
		structs:   make(map[string]*StructDef),
		methods:   make(map[string]map[string]*FuncDef),
		protected: make(map[string]bool),
		parent:    parent,
	}
}

func (env *Environment) getVar(name string) (Value, bool) {
	if val, ok := env.vars[name]; ok {
		return val, true
	}
	if env.parent != nil {
		return env.parent.getVar(name)
	}
	return Value{}, false
}

func (env *Environment) setVar(name string, val Value) {
	env.vars[name] = val
}

func (env *Environment) getFunc(name string) (*FuncDef, bool) {
	if fn, ok := env.funcs[name]; ok {
		return fn, true
	}
	if env.parent != nil {
		return env.parent.getFunc(name)
	}
	return nil, false
}

func (env *Environment) getMethod(typeName, method string) (*FuncDef, bool) {
	if m, ok := env.methods[typeName]; ok {
		if fn, ok := m[method]; ok {
			return fn, true
		}
	}
	if env.parent != nil {
		return env.parent.getMethod(typeName, method)
	}
	return nil, false
}

type Interpreter struct {
	globalEnv   *Environment
	currentFn   string
	mainDir     string
	imported    map[string]bool
	importFiles []string
	importStack []string
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		globalEnv: NewEnvironment(nil),
		imported:  make(map[string]bool),
	}
}

func (interp *Interpreter) eval(node ASTNode, env *Environment) Value {
	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Statements {
			switch s := stmt.(type) {
			case *StructDef:
				env.structs[s.Name] = s
			case *FuncDef:
				if s.RecvType != "" {
					if env.methods[s.RecvType] == nil {
						env.methods[s.RecvType] = make(map[string]*FuncDef)
					}
					env.methods[s.RecvType][s.Name] = s
				} else {
					env.funcs[s.Name] = s
				}
			}
		}
		var lastVal Value
		for _, stmt := range n.Statements {
			lastVal = interp.evalTopLevel(stmt, env)
		}
		return lastVal

	case *VarDecl:
		val := interp.eval(n.Value, env)
		env.setVar(n.Name, val)
		return val

	case *Assign:
		val := interp.eval(n.Value, env)
		current := env
		found := false
		for current != nil {
			if _, ok := current.vars[n.Name]; ok {
				current.vars[n.Name] = val
				found = true
				break
			}
			current = current.parent
		}
		if !found {
			env.setVar(n.Name, val)
		}
		return val

	case *FieldAssign:
		val := interp.eval(n.Value, env)
		current := env
		for current != nil {
			if obj, ok := current.vars[n.Object]; ok {
				if obj.Kind != "struct" {
					fail("Runtime error: variable '%s' is not a struct\n", n.Object)
				}
				if _, exists := obj.Fields[n.Field]; !exists {
					fail("Runtime error: struct '%s' has no field '%s'\n", obj.TypeName, n.Field)
				}
				obj.Fields[n.Field] = val
				current.vars[n.Object] = obj
				return val
			}
			current = current.parent
		}
		fail("Runtime error: undefined variable '%s'\n", n.Object)

	case *NumberLiteral:
		return Value{Kind: "number", NumVal: n.Value}

	case *BoolLiteral:
		return Value{Kind: "bool", BoolVal: n.Value}

	case *NilLiteral:
		return Value{Kind: "nil"}

	case *StringLiteral:
		return Value{Kind: "string", StrVal: n.Value}

	case *Identifier:
		if val, ok := env.getVar(n.Name); ok {
			return val
		}
		// named function referenced as a value (handler := hello)
		if fn, ok := env.getFunc(n.Name); ok {
			return Value{Kind: "func", Fn: fn, Closure: interp.globalEnv}
		}
		fail("Runtime error: undefined variable '%s'\n", n.Name)

	case *BinaryOp:
		left := interp.eval(n.Left, env)
		right := interp.eval(n.Right, env)
		return interp.evalBinaryOp(left, n.Op, right)

	case *UnaryOp:
		val := interp.eval(n.Expr, env)
		switch n.Op {
		case "!":
			return Value{Kind: "bool", BoolVal: !isTruthy(val)}
		case "-":
			if val.Kind == "number" {
				return Value{Kind: "number", NumVal: -val.NumVal}
			}
			fail("Runtime error: unary - on %s\n", val.Kind)
		default:
			fail("Runtime error: unknown unary op %s\n", n.Op)
		}

	case *FuncCall:
		return interp.evalFuncCall(n, env)

	case *MethodCall:
		return interp.evalMethodCall(n, env)

	case *ImportStmt:
		return interp.evalImport(n.Path, env)

	case *TryCatch:
		return interp.evalTry(n, env)

	case *IfStatement:
		cond := interp.eval(n.Condition, env)
		if isTruthy(cond) {
			return interp.evalBlock(n.Body, env)
		} else if len(n.ElseBranch) > 0 {
			return interp.evalBlock(n.ElseBranch, env)
		}
		return Value{Kind: "nil"}

	case *WhileLoop:
		return interp.evalWhile(n, env)

	case *ForLoop:
		return interp.evalFor(n, env)

	case *ForIn:
		return interp.evalForIn(n, env)

	case *ArrayLiteral:
		items := make([]Value, 0, len(n.Elements))
		for _, e := range n.Elements {
			items = append(items, interp.eval(e, env))
		}
		return Value{Kind: "array", Items: items}

	case *MapLiteral:
		m := make(map[string]Value, len(n.Keys))
		for i, k := range n.Keys {
			m[k] = interp.eval(n.Values[i], env)
		}
		return Value{Kind: "map", MapVal: m}

	case *IndexAccess:
		cont := interp.eval(n.Target, env)
		if cont.Kind == "map" {
			key := interp.evalMapKey(n.Index, env)
			val, ok := cont.MapVal[key]
			if !ok {
				fail("Runtime error: missing key %q\n", key)
			}
			return val
		}
		if cont.Kind != "array" {
			fail("Runtime error: indexing non-array (%s)\n", cont.Kind)
		}
		return cont.Items[interp.evalArrayIndex(n.Index, env, len(cont.Items))]

	case *IndexAssign:
		cont := interp.eval(n.Target, env)
		val := interp.eval(n.Value, env)
		if cont.Kind == "map" {
			key := interp.evalMapKey(n.Index, env)
			cont.MapVal[key] = val
			return val
		}
		if cont.Kind != "array" {
			fail("Runtime error: indexing non-array (%s)\n", cont.Kind)
		}
		cont.Items[interp.evalArrayIndex(n.Index, env, len(cont.Items))] = val
		return val

	case *SliceAccess:
		cont := interp.eval(n.Target, env)
		switch cont.Kind {
		case "array":
			s, e := interp.evalSliceBounds(n.Start, n.End, env, len(cont.Items))
			out := make([]Value, 0, e-s)
			out = append(out, cont.Items[s:e]...)
			return Value{Kind: "array", Items: out}
		case "string":
			runes := []rune(cont.StrVal)
			s, e := interp.evalSliceBounds(n.Start, n.End, env, len(runes))
			return Value{Kind: "string", StrVal: string(runes[s:e])}
		default:
			fail("Runtime error: slicing %s\n", cont.Kind)
		}

	case *BreakStmt:
		panic(&breakSignal{})

	case *ContinueStmt:
		panic(&continueSignal{})

	case *ReturnStmt:
		val := interp.eval(n.Value, env)
		panic(&returnValue{val})

	case *StructDef:
		return Value{Kind: "nil"}

	case *FuncDef:
		return Value{Kind: "nil"}

	case *FuncLit:
		return Value{Kind: "func", Fn: &FuncDef{Params: n.Params, Body: n.Body}, Closure: env}

	case *StructLiteral:
		var current = env
		var structDef *StructDef
		var ok bool
		for current != nil {
			if structDef, ok = current.structs[n.Name]; ok {
				break
			}
			current = current.parent
		}
		if structDef == nil {
			fail("Runtime error: unknown struct '%s'\n", n.Name)
		}
		fields := make(map[string]Value)
		for i, fieldName := range structDef.Fields {
			if i < len(n.Values) {
				fields[fieldName] = interp.eval(n.Values[i], env)
			} else {
				fields[fieldName] = Value{Kind: "nil"}
			}
		}
		return Value{Kind: "struct", TypeName: n.Name, Fields: fields}

	case *FieldAccess:
		obj := interp.eval(n.Object, env)
		if obj.Kind != "struct" {
			fail("Runtime error: field access on non-struct\n")
		}
		if val, ok := obj.Fields[n.Field]; ok {
			return val
		}
		fail("Runtime error: struct has no field '%s'\n", n.Field)

	case *DelCall:
		return interp.evalDel(n, env)

	default:
		fail("Runtime error: unknown node type %T\n", node)
	}
	return Value{Kind: "nil"}
}

func (interp *Interpreter) evalTopLevel(stmt ASTNode, env *Environment) (out Value) {
	defer func() {
		if r := recover(); r != nil {
			switch r.(type) {
			case *breakSignal:
				fmt.Fprintf(os.Stderr, "Runtime error: 'break' outside loop\n")
			case *continueSignal:
				fmt.Fprintf(os.Stderr, "Runtime error: 'continue' outside loop\n")
			case *returnValue:
				fmt.Fprintf(os.Stderr, "Runtime error: 'return' outside function\n")
			case *runtimeError:
				fmt.Fprintf(os.Stderr, "%s\n", r.(*runtimeError).msg)
			default:
				panic(r)
			}
			os.Exit(1)
		}
	}()
	out = interp.eval(stmt, env)
	return out
}

func (interp *Interpreter) evalBlock(stmts []ASTNode, env *Environment) Value {
	blockEnv := NewEnvironment(env)
	// defer: блоковые локали чистятся даже при break/continue/return изнутри
	defer interp.cleanupLocals(blockEnv)
	var lastVal Value = Value{Kind: "nil"}
	for _, stmt := range stmts {
		lastVal = interp.eval(stmt, blockEnv)
	}
	return lastVal
}

type returnValue struct {
	val Value
}

type breakSignal struct{}

type continueSignal struct{}

// runtimeError is a catchable script failure. All interpreter error paths
// panic with it instead of calling os.Exit, so try/catch can intercept;
// uncaught it prints exactly like the old fatal errors.
type runtimeError struct {
	msg string
}

func fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimSuffix(msg, "\n")
	panic(&runtimeError{msg: msg})
}

func (interp *Interpreter) evalFor(node *ForLoop, env *Environment) Value {
	loopEnv := NewEnvironment(env)
	defer interp.cleanupLocals(loopEnv)
	if node.Init != nil {
		interp.eval(node.Init, loopEnv)
	}
	lastVal := Value{Kind: "nil"}
	for {
		if node.Cond != nil {
			if !isTruthy(interp.eval(node.Cond, loopEnv)) {
				break
			}
		}
		hitBreak := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(*breakSignal); ok {
						hitBreak = true
						return
					}
					// continue: выходим из итерации, post ниже всё равно выполнится
					if _, ok := r.(*continueSignal); ok {
						return
					}
					panic(r)
				}
			}()
			lastVal = interp.evalBlock(node.Body, loopEnv)
		}()
		if hitBreak {
			break
		}
		if node.Post != nil {
			interp.eval(node.Post, loopEnv)
		}
	}
	return lastVal
}

func (interp *Interpreter) evalForIn(node *ForIn, env *Environment) Value {
	iterable := interp.eval(node.Iterable, env)
	if iterable.Kind != "array" {
		fail("Runtime error: for-in needs an array, got %s\n", iterable.Kind)
	}
	loopEnv := NewEnvironment(env)
	defer interp.cleanupLocals(loopEnv)
	lastVal := Value{Kind: "nil"}
	for _, item := range iterable.Items {
		loopEnv.setVar(node.Var, item)
		hitBreak := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(*breakSignal); ok {
						hitBreak = true
						return
					}
					if _, ok := r.(*continueSignal); ok {
						return
					}
					panic(r)
				}
			}()
			lastVal = interp.evalBlock(node.Body, loopEnv)
		}()
		if hitBreak {
			break
		}
	}
	return lastVal
}

func (interp *Interpreter) evalMapKey(node ASTNode, env *Environment) string {
	v := interp.eval(node, env)
	if v.Kind != "string" {
		fail("Runtime error: map key must be a string, got %s\n", v.Kind)
	}
	return v.StrVal
}

func (interp *Interpreter) evalArrayIndex(node ASTNode, env *Environment, length int) int {
	v := interp.eval(node, env)
	if v.Kind != "number" {
		fail("Runtime error: array index must be a number, got %s\n", v.Kind)
	}
	if v.NumVal != math.Trunc(v.NumVal) {
		fail("Runtime error: array index must be an integer\n")
	}
	i := int(v.NumVal)
	if i < 0 {
		i += length
	}
	if i < 0 || i >= length {
		fail("Runtime error: index %d out of range (len %d)\n", int(v.NumVal), length)
	}
	return i
}

func (interp *Interpreter) evalSliceBounds(start, end ASTNode, env *Environment, length int) (int, int) {
	bound := func(node ASTNode, def int) int {
		if node == nil {
			return def
		}
		v := interp.eval(node, env)
		if v.Kind != "number" {
			fail("Runtime error: slice bound must be a number, got %s\n", v.Kind)
		}
		if v.NumVal != math.Trunc(v.NumVal) {
			fail("Runtime error: slice bound must be an integer\n")
		}
		i := int(v.NumVal)
		if i < 0 {
			i += length
		}
		if i < 0 {
			i = 0
		}
		if i > length {
			i = length
		}
		return i
	}
	s := bound(start, 0)
	e := bound(end, length)
	if s > e {
		return s, s
	}
	return s, e
}

func (interp *Interpreter) evalWhile(node *WhileLoop, env *Environment) Value {
	lastVal := Value{Kind: "nil"}
	for isTruthy(interp.eval(node.Condition, env)) {
		hitBreak := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					if _, ok := r.(*breakSignal); ok {
						hitBreak = true
						return
					}
					// continue: просто прерываем текущую итерацию
					if _, ok := r.(*continueSignal); ok {
						return
					}
					// return и прочее — пробрасываем выше
					panic(r)
				}
			}()
			lastVal = interp.evalBlock(node.Body, env)
		}()
		if hitBreak {
			break
		}
	}
	return lastVal
}

func (interp *Interpreter) evalFuncCall(call *FuncCall, env *Environment) (result Value) {
	if call.Name == "print" {
		for _, arg := range call.Args {
			val := interp.eval(arg, env)
			fmt.Print(valueToString(val))
		}
		fmt.Println()
		return Value{Kind: "nil"}
	}

	if call.Name == "len" {
		if len(call.Args) != 1 {
			fail("Runtime error: len() takes exactly 1 argument\n")
		}
		v := interp.eval(call.Args[0], env)
		switch v.Kind {
		case "array":
			return Value{Kind: "number", NumVal: float64(len(v.Items))}
		case "map":
			return Value{Kind: "number", NumVal: float64(len(v.MapVal))}
		case "string":
			return Value{Kind: "number", NumVal: float64(len([]rune(v.StrVal)))}
		default:
			fail("Runtime error: len() of %s\n", v.Kind)
		}
	}

	if call.Name == "push" {
		if len(call.Args) != 2 {
			fail("Runtime error: push() takes exactly 2 arguments\n")
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fail("Runtime error: push() target must be a variable\n")
		}
		arr, ok := env.getVar(target.Name)
		if !ok || arr.Kind != "array" {
			fail("Runtime error: push() target '%s' is not an array\n", target.Name)
		}
		arr.Items = append(arr.Items, interp.eval(call.Args[1], env))
		current := env
		for current != nil {
			if _, found := current.vars[target.Name]; found {
				current.vars[target.Name] = arr
				break
			}
			current = current.parent
		}
		return arr
	}

	if call.Name == "str" {
		if len(call.Args) != 1 {
			fail("Runtime error: str() takes exactly 1 argument\n")
		}
		return Value{Kind: "string", StrVal: valueToString(interp.eval(call.Args[0], env))}
	}

	if call.Name == "num" {
		if len(call.Args) != 1 {
			fail("Runtime error: num() takes exactly 1 argument\n")
		}
		v := interp.eval(call.Args[0], env)
		switch v.Kind {
		case "number":
			return v
		case "bool":
			if v.BoolVal {
				return Value{Kind: "number", NumVal: 1}
			}
			return Value{Kind: "number", NumVal: 0}
		case "string":
			f, err := strconv.ParseFloat(strings.TrimSpace(v.StrVal), 64)
			if err != nil {
				fail("Runtime error: num() cannot convert %q\n", v.StrVal)
			}
			return Value{Kind: "number", NumVal: f}
		default:
			fail("Runtime error: num() of %s\n", v.Kind)
		}
	}

	if call.Name == "input" {
		if len(call.Args) > 1 {
			fail("Runtime error: input() takes at most 1 argument\n")
		}
		if len(call.Args) == 1 {
			prompt := interp.eval(call.Args[0], env)
			fmt.Print(valueToString(prompt))
		}
		// общий ридер: новый bufio на каждый вызов выбрасывал бы
		// уже забуферизованный stdin
		line, err := stdinReader.ReadString('\n')
		if err != nil && err != io.EOF {
			fail("Runtime error: input() failed: %v\n", err)
		}
		return Value{Kind: "string", StrVal: strings.TrimRight(line, "\r\n")}
	}

	if call.Name == "upper" || call.Name == "lower" {
		if len(call.Args) != 1 {
			fail("Runtime error: %s() takes exactly 1 argument\n", call.Name)
		}
		v := interp.eval(call.Args[0], env)
		if v.Kind != "string" {
			fail("Runtime error: %s() needs a string, got %s\n", call.Name, v.Kind)
		}
		if call.Name == "upper" {
			return Value{Kind: "string", StrVal: strings.ToUpper(v.StrVal)}
		}
		return Value{Kind: "string", StrVal: strings.ToLower(v.StrVal)}
	}

	if call.Name == "contains" {
		if len(call.Args) != 2 {
			fail("Runtime error: contains() takes exactly 2 arguments\n")
		}
		s := interp.eval(call.Args[0], env)
		sub := interp.eval(call.Args[1], env)
		if s.Kind != "string" || sub.Kind != "string" {
			fail("Runtime error: contains() needs strings\n")
		}
		return Value{Kind: "bool", BoolVal: strings.Contains(s.StrVal, sub.StrVal)}
	}

	if call.Name == "split" {
		if len(call.Args) != 2 {
			fail("Runtime error: split() takes exactly 2 arguments\n")
		}
		s := interp.eval(call.Args[0], env)
		sep := interp.eval(call.Args[1], env)
		if s.Kind != "string" || sep.Kind != "string" {
			fail("Runtime error: split() needs strings\n")
		}
		parts := strings.Split(s.StrVal, sep.StrVal)
		items := make([]Value, 0, len(parts))
		for _, part := range parts {
			items = append(items, Value{Kind: "string", StrVal: part})
		}
		return Value{Kind: "array", Items: items}
	}

	if call.Name == "join" {
		if len(call.Args) != 2 {
			fail("Runtime error: join() takes exactly 2 arguments\n")
		}
		arr := interp.eval(call.Args[0], env)
		sep := interp.eval(call.Args[1], env)
		if arr.Kind != "array" || sep.Kind != "string" {
			fail("Runtime error: join() needs (array, string)\n")
		}
		parts := make([]string, 0, len(arr.Items))
		for _, item := range arr.Items {
			parts = append(parts, valueToString(item))
		}
		return Value{Kind: "string", StrVal: strings.Join(parts, sep.StrVal)}
	}

	if call.Name == "keys" {
		if len(call.Args) != 1 {
			fail("Runtime error: keys() takes exactly 1 argument\n")
		}
		v := interp.eval(call.Args[0], env)
		if v.Kind != "map" {
			fail("Runtime error: keys() needs a map, got %s\n", v.Kind)
		}
		ks := make([]string, 0, len(v.MapVal))
		for k := range v.MapVal {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		items := make([]Value, 0, len(ks))
		for _, k := range ks {
			items = append(items, Value{Kind: "string", StrVal: k})
		}
		return Value{Kind: "array", Items: items}
	}

	if call.Name == "has" {
		if len(call.Args) != 2 {
			fail("Runtime error: has() takes exactly 2 arguments\n")
		}
		m := interp.eval(call.Args[0], env)
		k := interp.eval(call.Args[1], env)
		if m.Kind != "map" || k.Kind != "string" {
			fail("Runtime error: has() needs (map, string)\n")
		}
		_, ok := m.MapVal[k.StrVal]
		return Value{Kind: "bool", BoolVal: ok}
	}

	if call.Name == "sort" {
		if len(call.Args) != 1 {
			fail("Runtime error: sort() takes exactly 1 argument\n")
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fail("Runtime error: sort() target must be a variable\n")
		}
		arr, ok := env.getVar(target.Name)
		if !ok || arr.Kind != "array" {
			fail("Runtime error: sort() target '%s' is not an array\n", target.Name)
		}
		if len(arr.Items) > 0 {
			switch arr.Items[0].Kind {
			case "number":
				for _, item := range arr.Items {
					if item.Kind != "number" {
						fail("Runtime error: sort() needs uniformly typed array\n")
					}
				}
				sort.Slice(arr.Items, func(a, b int) bool { return arr.Items[a].NumVal < arr.Items[b].NumVal })
			case "string":
				for _, item := range arr.Items {
					if item.Kind != "string" {
						fail("Runtime error: sort() needs uniformly typed array\n")
					}
				}
				sort.Slice(arr.Items, func(a, b int) bool { return arr.Items[a].StrVal < arr.Items[b].StrVal })
			default:
				fail("Runtime error: sort() supports numbers and strings\n")
			}
		}
		current := env
		for current != nil {
			if _, found := current.vars[target.Name]; found {
				current.vars[target.Name] = arr
				break
			}
			current = current.parent
		}
		return arr
	}

	if call.Name == "args" {
		if len(call.Args) != 0 {
			fail("Runtime error: args() takes no arguments\n")
		}
		items := make([]Value, 0, len(os.Args)-2)
		for _, a := range os.Args[2:] {
			items = append(items, Value{Kind: "string", StrVal: a})
		}
		return Value{Kind: "array", Items: items}
	}

	if call.Name == "http_listen" {
		if len(call.Args) != 2 {
			fail("Runtime error: http_listen() takes exactly 2 arguments (port, handler)\n")
		}
		portVal := interp.eval(call.Args[0], env)
		if portVal.Kind != "number" {
			fail("Runtime error: http_listen() port must be a number, got %s\n", portVal.Kind)
		}
		handler := interp.eval(call.Args[1], env)
		if handler.Kind != "func" {
			fail("Runtime error: http_listen() handler must be a function, got %s\n", handler.Kind)
		}
		fnDef := handler.Fn
		closure := handler.Closure
		if closure == nil {
			closure = interp.globalEnv
		}
		mux := http.NewServeMux()
		mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rec := recover(); rec != nil {
					msg := "handler crash"
					if re, ok := rec.(*runtimeError); ok {
						msg = re.msg
					}
					http.Error(w, msg, http.StatusInternalServerError)
				}
			}()
			body, _ := io.ReadAll(io.LimitReader(r.Body, 1<<20))
			r.Body.Close()
			query := make(map[string]Value)
			for k, vs := range r.URL.Query() {
				if len(vs) > 0 {
					query[k] = Value{Kind: "string", StrVal: vs[0]}
				}
			}
			reqVal := Value{Kind: "map", MapVal: map[string]Value{
				"method": Value{Kind: "string", StrVal: r.Method},
				"path":   Value{Kind: "string", StrVal: r.URL.Path},
				"query":  Value{Kind: "map", MapVal: query},
				"body":   Value{Kind: "string", StrVal: string(body)},
			}}
			res := interp.invokeUserFunc(fnDef, closure, []Value{reqVal}, "http-handler")
			status := 200
			out := ""
			if res.Kind == "string" {
				out = res.StrVal
			} else if res.Kind == "map" {
				if st, ok := res.MapVal["status"]; ok && st.Kind == "number" {
					status = int(st.NumVal)
				}
				if b, ok := res.MapVal["body"]; ok {
					out = valueToString(b)
				}
			} else if res.Kind != "nil" {
				out = valueToString(res)
			}
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
			w.WriteHeader(status)
			w.Write([]byte(out))
		})
		addr := fmt.Sprintf("127.0.0.1:%d", int(portVal.NumVal))
		if err := http.ListenAndServe(addr, mux); err != nil {
			fail("Runtime error: http_listen() failed: %v\n", err)
		}
		return Value{Kind: "nil"}
	}

	if call.Name == "http_get" {
		if len(call.Args) != 1 {
			fail("Runtime error: http_get() takes exactly 1 argument\n")
		}
		url := interp.eval(call.Args[0], env)
		if url.Kind != "string" {
			fail("Runtime error: http_get() needs a string URL, got %s\n", url.Kind)
		}
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Get(url.StrVal)
		if err != nil {
			fail("Runtime error: http_get() failed: %v\n", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fail("Runtime error: http_get() status %s\n", resp.Status)
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		if err != nil {
			fail("Runtime error: http_get() read failed: %v\n", err)
		}
		return Value{Kind: "string", StrVal: string(body)}
	}

	if call.Name == "read_file" {
		if len(call.Args) != 1 {
			fail("Runtime error: read_file() takes exactly 1 argument\n")
		}
		path := interp.eval(call.Args[0], env)
		if path.Kind != "string" {
			fail("Runtime error: read_file() needs a string path, got %s\n", path.Kind)
		}
		data, err := os.ReadFile(path.StrVal)
		if err != nil {
			fail("Runtime error: read_file() failed: %v\n", err)
		}
		return Value{Kind: "string", StrVal: string(data)}
	}

	if call.Name == "write_file" || call.Name == "append_file" {
		if len(call.Args) != 2 {
			fail("Runtime error: %s() takes exactly 2 arguments\n", call.Name)
		}
		path := interp.eval(call.Args[0], env)
		content := interp.eval(call.Args[1], env)
		if path.Kind != "string" || content.Kind != "string" {
			fail("Runtime error: %s() needs (string, string)\n", call.Name)
		}
		var err error
		if call.Name == "write_file" {
			err = os.WriteFile(path.StrVal, []byte(content.StrVal), 0644)
		} else {
			var f *os.File
			f, err = os.OpenFile(path.StrVal, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err == nil {
				_, err = f.WriteString(content.StrVal)
				f.Close()
			}
		}
		if err != nil {
			fail("Runtime error: %s() failed: %v\n", call.Name, err)
		}
		return Value{Kind: "number", NumVal: float64(len(content.StrVal))}
	}

	if call.Name == "exists" {
		if len(call.Args) != 1 {
			fail("Runtime error: exists() takes exactly 1 argument\n")
		}
		path := interp.eval(call.Args[0], env)
		if path.Kind != "string" {
			fail("Runtime error: exists() needs a string path, got %s\n", path.Kind)
		}
		_, err := os.Stat(path.StrVal)
		return Value{Kind: "bool", BoolVal: err == nil}
	}

	if call.Name == "type" {
		if len(call.Args) != 1 {
			fail("Runtime error: type() takes exactly 1 argument\n")
		}
		v := interp.eval(call.Args[0], env)
		if v.Kind == "struct" {
			return Value{Kind: "string", StrVal: "struct:" + v.TypeName}
		}
		return Value{Kind: "string", StrVal: v.Kind}
	}

	if call.Name == "env" {
		if len(call.Args) != 1 {
			fail("Runtime error: env() takes exactly 1 argument\n")
		}
		name := interp.eval(call.Args[0], env)
		if name.Kind != "string" {
			fail("Runtime error: env() needs a string, got %s\n", name.Kind)
		}
		if val, ok := os.LookupEnv(name.StrVal); ok {
			return Value{Kind: "string", StrVal: val}
		}
		return Value{Kind: "nil"}
	}

	if call.Name == "exit" {
		if len(call.Args) > 1 {
			fail("Runtime error: exit() takes at most 1 argument\n")
		}
		code := 0
		if len(call.Args) == 1 {
			v := interp.eval(call.Args[0], env)
			if v.Kind != "number" {
				fail("Runtime error: exit() needs a number, got %s\n", v.Kind)
			}
			code = int(v.NumVal)
		}
		os.Exit(code)
		return Value{Kind: "nil"}
	}

	if call.Name == "now" {
		if len(call.Args) != 0 {
			fail("Runtime error: now() takes no arguments\n")
		}
		return Value{Kind: "number", NumVal: float64(time.Now().Unix())}
	}

	if call.Name == "parse_json" {
		if len(call.Args) != 1 {
			fail("Runtime error: parse_json() takes exactly 1 argument\n")
		}
		s := interp.eval(call.Args[0], env)
		if s.Kind != "string" {
			fail("Runtime error: parse_json() needs a string, got %s\n", s.Kind)
		}
		val, err := parseJSON(s.StrVal)
		if err != nil {
			fail("Runtime error: parse_json() failed: %v\n", err)
		}
		return val
	}

	if call.Name == "to_json" {
		if len(call.Args) != 1 {
			fail("Runtime error: to_json() takes exactly 1 argument\n")
		}
		return Value{Kind: "string", StrVal: valueToJSON(interp.eval(call.Args[0], env))}
	}

	if call.Name == "sleep" {
		if len(call.Args) != 1 {
			fail("Runtime error: sleep() takes exactly 1 argument\n")
		}
		ms := interp.eval(call.Args[0], env)
		if ms.Kind != "number" || ms.NumVal < 0 {
			fail("Runtime error: sleep() needs non-negative milliseconds\n")
		}
		time.Sleep(time.Duration(ms.NumVal * float64(time.Millisecond)))
		return Value{Kind: "nil"}
	}

	if call.Name == "pkgdir" {
		if len(call.Args) != 1 {
			fail("Runtime error: pkgdir() takes exactly 1 argument\n")
		}
		name := interp.eval(call.Args[0], env)
		if name.Kind != "string" {
			fail("Runtime error: pkgdir() needs a string, got %s\n", name.Kind)
		}
		spec, ok := parsePkgSpec(name.StrVal)
		if !ok {
			fail("Runtime error: pkgdir() wants user/repo[@ver], got %q\n", name.StrVal)
		}
		dest, err := ensurePackage(spec)
		if err != nil {
			fail("Runtime error: pkgdir() fetch failed: %v\n", err)
		}
		return Value{Kind: "string", StrVal: dest}
	}

	// variable holding a function value (first-class functions, closures)?
	if v, found := env.getVar(call.Name); found && v.Kind == "func" {
		argVals := make([]Value, 0, len(call.Args))
		for _, a := range call.Args {
			argVals = append(argVals, interp.eval(a, env))
		}
		return interp.invokeUserFunc(v.Fn, v.Closure, argVals, call.Name)
	}

	fn, ok := env.getFunc(call.Name)
	if !ok {
		fail("Runtime error: undefined function '%s'\n", call.Name)
	}

	argVals := make([]Value, 0, len(call.Args))
	for _, a := range call.Args {
		argVals = append(argVals, interp.eval(a, env))
	}
	return interp.invokeUserFunc(fn, interp.globalEnv, argVals, call.Name)
}

// invokeUserFunc runs a user function body in a fresh frame parented at
// frameParent (global scope for plain functions, captured scope for
// closures). Shared by plain calls, function values and methods.
func (interp *Interpreter) invokeUserFunc(fn *FuncDef, frameParent *Environment, argVals []Value, callerName string) (result Value) {
	fnEnv := NewEnvironment(frameParent)
	for i, param := range fn.Params {
		if i < len(argVals) {
			fnEnv.setVar(param, argVals[i])
		} else {
			fnEnv.setVar(param, Value{Kind: "nil"})
		}
	}

	interp.currentFn = callerName

	defer func() {
		if r := recover(); r != nil {
			if rv, ok := r.(*returnValue); ok {
				interp.cleanupLocals(fnEnv)
				result = rv.val
				return
			}
			if _, ok := r.(*runtimeError); ok {
				interp.cleanupLocals(fnEnv)
			}
			panic(r)
		}
	}()

	var lastVal Value = Value{Kind: "nil"}
	for _, stmt := range fn.Body {
		lastVal = interp.eval(stmt, fnEnv)
	}
	interp.cleanupLocals(fnEnv)
	result = lastVal
	return result
}

// ========== PACKAGES (GitHub) ==========
//
// import "user/repo"            -> latest main branch, entry main.cx
// import "user/repo@v1.2.0"     -> tag or branch v1.2.0
// import "user/repo/lib/a.cx"   -> explicit file inside the repo
// import "./local.cx"           -> local file, relative to the importer
//
// Packages are zip-downloaded from GitHub into ~/.codex/pkgs and cached.
// `codex get user/repo@ver` prefetches without running anything.
// An imported file is evaluated once into the shared global scope.

type pkgSpec struct {
	user string
	repo string
	ver  string
	file string
}

func parsePkgSpec(s string) (pkgSpec, bool) {
	t := strings.TrimSpace(s)
	t = strings.TrimPrefix(t, "github.com/")
	if strings.HasPrefix(t, ".") || strings.HasPrefix(t, "/") || !strings.Contains(t, "/") {
		return pkgSpec{}, false
	}
	parts := strings.SplitN(t, "/", 3)
	if len(parts) < 2 || parts[0] == "" || parts[1] == "" {
		return pkgSpec{}, false
	}
	user := parts[0]
	repo := parts[1]
	ver := ""
	if i := strings.Index(repo, "@"); i >= 0 {
		ver = repo[i+1:]
		repo = repo[:i]
	}
	file := ""
	if len(parts) == 3 {
		file = parts[2]
	}
	if repo == "" || strings.Contains(repo, "..") || strings.Contains(user, "..") {
		return pkgSpec{}, false
	}
	return pkgSpec{user: user, repo: repo, ver: ver, file: file}, true
}

func pkgCacheRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex", "pkgs", "github.com"), nil
}

func githubAPIDefaultBranch(user, repo string) (string, error) {
	url := "https://api.github.com/repos/" + user + "/" + repo
	client := &http.Client{Timeout: 15 * time.Second}
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/vnd.github+json")
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return "", fmt.Errorf("github api: %s for %s/%s", resp.Status, user, repo)
	}
	var info struct {
		DefaultBranch string `json:"default_branch"`
	}
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(body, &info); err != nil {
		return "", err
	}
	if info.DefaultBranch == "" {
		return "", fmt.Errorf("no default branch for %s/%s", user, repo)
	}
	return info.DefaultBranch, nil
}

func downloadAndUnzip(zipURL, dest string) error {
	client := &http.Client{Timeout: 60 * time.Second}
	resp, err := client.Get(zipURL)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return fmt.Errorf("download: %s", resp.Status)
	}
	data, err := io.ReadAll(io.LimitReader(resp.Body, 64<<20))
	if err != nil {
		return err
	}
	zr, err := zip.NewReader(bytes.NewReader(data), int64(len(data)))
	if err != nil {
		return err
	}
	for _, f := range zr.File {
		// strip the top-level user-repo-hash folder
		rel := f.Name
		if i := strings.Index(rel, "/"); i >= 0 {
			rel = rel[i+1:]
		} else {
			continue
		}
		if rel == "" {
			continue
		}
		target := filepath.Join(dest, filepath.FromSlash(rel))
		if !strings.HasPrefix(filepath.Clean(target), filepath.Clean(dest)) {
			return fmt.Errorf("zip slip: %s", f.Name)
		}
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
			continue
		}
		if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
			return err
		}
		rc, err := f.Open()
		if err != nil {
			return err
		}
		content, err := io.ReadAll(rc)
		rc.Close()
		if err != nil {
			return err
		}
		if err := os.WriteFile(target, content, 0644); err != nil {
			return err
		}
	}
	return nil
}

func ensurePackage(spec pkgSpec) (string, error) {
	root, err := pkgCacheRoot()
	if err != nil {
		return "", err
	}
	ver := spec.ver
	if ver == "" {
		ver = "latest"
	}
	dest := filepath.Join(root, spec.user, spec.repo+"@"+ver)
	if _, err := os.Stat(filepath.Join(dest, ".ok")); err == nil {
		return dest, nil
	}
	var zipURL string
	if spec.ver == "" {
		branch, err := githubAPIDefaultBranch(spec.user, spec.repo)
		if err != nil {
			return "", err
		}
		zipURL = "https://codeload.github.com/" + spec.user + "/" + spec.repo + "/zip/refs/heads/" + branch
	} else {
		zipURL = "https://codeload.github.com/" + spec.user + "/" + spec.repo + "/zip/refs/tags/" + spec.ver
		if err := os.MkdirAll(dest, 0755); err != nil {
			return "", err
		}
		if err := downloadAndUnzip(zipURL, dest); err != nil {
			// fall back to branch with the same name
			os.RemoveAll(dest)
			if err := os.MkdirAll(dest, 0755); err != nil {
				return "", err
			}
			zipURL = "https://codeload.github.com/" + spec.user + "/" + spec.repo + "/zip/refs/heads/" + spec.ver
			if err := downloadAndUnzip(zipURL, dest); err != nil {
				os.RemoveAll(dest)
				return "", fmt.Errorf("no tag or branch %q in %s/%s", spec.ver, spec.user, spec.repo)
			}
		}
		if err := os.WriteFile(filepath.Join(dest, ".ok"), []byte(spec.ver), 0644); err != nil {
			return "", err
		}
		return dest, nil
	}
	if err := os.MkdirAll(dest, 0755); err != nil {
		return "", err
	}
	if err := downloadAndUnzip(zipURL, dest); err != nil {
		os.RemoveAll(dest)
		return "", err
	}
	if err := os.WriteFile(filepath.Join(dest, ".ok"), []byte("latest"), 0644); err != nil {
		return "", err
	}
	return dest, nil
}

func pkgEntryFile(dest string, spec pkgSpec) (string, error) {
	if spec.file != "" {
		p := filepath.Join(dest, filepath.FromSlash(spec.file))
		if !strings.HasPrefix(filepath.Clean(p), filepath.Clean(dest)) {
			return "", fmt.Errorf("path escapes package: %s", spec.file)
		}
		if !strings.HasSuffix(strings.ToLower(p), ".cx") {
			return "", fmt.Errorf("package file must end with .cx: %s", spec.file)
		}
		if _, err := os.Stat(p); err != nil {
			return "", fmt.Errorf("no such file %s in %s/%s", spec.file, spec.user, spec.repo)
		}
		return p, nil
	}
	for _, c := range []string{"main.cx", spec.repo + ".cx", "lib.cx"} {
		p := filepath.Join(dest, c)
		if _, err := os.Stat(p); err == nil {
			return p, nil
		}
	}
	return "", fmt.Errorf("package %s/%s has no entry (add main.cx)", spec.user, spec.repo)
}

func (interp *Interpreter) evalImport(path string, env *Environment) Value {
	var abs string
	if spec, ok := parsePkgSpec(path); ok {
		dest, err := ensurePackage(spec)
		if err != nil {
			fail("Import error %q: %v", path, err)
		}
		f, err := pkgEntryFile(dest, spec)
		if err != nil {
			fail("Import error %q: %v", path, err)
		}
		abs = f
	} else {
		base := interp.mainDir
		if len(interp.importStack) > 0 {
			base = interp.importStack[len(interp.importStack)-1]
		}
		p := path
		if !strings.HasSuffix(strings.ToLower(p), ".cx") {
			p += ".cx"
		}
		if !filepath.IsAbs(p) {
			p = filepath.Join(base, p)
		}
		abs = p
		if _, err := os.Stat(abs); err != nil {
			fail("Import error: cannot stat %s\n", abs)
		}
	}
	abs = filepath.Clean(abs)
	if interp.imported[abs] {
		return Value{Kind: "nil"}
	}
	for _, f := range interp.importFiles {
		if f == abs {
			fail("Import error: cycle detected at %s\n", abs)
		}
	}
	data, err := os.ReadFile(abs)
	if err != nil {
		fail("Import error: cannot read %s: %v\n", abs, err)
	}
	interp.importFiles = append(interp.importFiles, abs)
	interp.importStack = append(interp.importStack, filepath.Dir(abs))
	lx := NewLexer(string(data))
	prog := NewParser(lx.Tokenize()).ParseProgram()
	interp.eval(prog, interp.globalEnv)
	interp.importStack = interp.importStack[:len(interp.importStack)-1]
	interp.imported[abs] = true
	return Value{Kind: "nil"}
}

func (interp *Interpreter) evalTry(node *TryCatch, env *Environment) (result Value) {
	result = Value{Kind: "nil"}
	func() {
		defer func() {
			if r := recover(); r != nil {
				if re, ok := r.(*runtimeError); ok {
					cenv := NewEnvironment(env)
					if node.CatchVar != "" {
						cenv.setVar(node.CatchVar, Value{Kind: "string", StrVal: re.msg})
					}
					result = interp.evalBlock(node.CatchBody, cenv)
					return
				}
				panic(r)
			}
		}()
		result = interp.evalBlock(node.Body, env)
	}()
	return result
}

func (interp *Interpreter) evalMethodCall(call *MethodCall, env *Environment) (result Value) {
	recv := interp.eval(call.Receiver, env)
	if recv.Kind != "struct" {
		fail("Runtime error: method '%s' called on non-struct (%s)\n", call.Method, recv.Kind)
	}
	fn, ok := env.getMethod(recv.TypeName, call.Method)
	if !ok {
		fail("Runtime error: struct '%s' has no method '%s'\n", recv.TypeName, call.Method)
	}

	fnEnv := NewEnvironment(interp.globalEnv)
	// receiver делит карту полей с оригиналом: мутации полей видны вызывающему
	fnEnv.setVar(fn.RecvName, recv)
	fnEnv.protected[fn.RecvName] = true
	for i, param := range fn.Params {
		if i < len(call.Args) {
			fnEnv.setVar(param, interp.eval(call.Args[i], env))
		} else {
			fnEnv.setVar(param, Value{Kind: "nil"})
		}
	}

	interp.currentFn = recv.TypeName + "." + call.Method

	defer func() {
		if r := recover(); r != nil {
			if rv, ok := r.(*returnValue); ok {
				interp.cleanupLocals(fnEnv)
				result = rv.val
				return
			}
			if _, ok := r.(*runtimeError); ok {
				interp.cleanupLocals(fnEnv)
			}
			panic(r)
		}
	}()

	lastVal := Value{Kind: "nil"}
	for _, stmt := range fn.Body {
		lastVal = interp.eval(stmt, fnEnv)
	}
	interp.cleanupLocals(fnEnv)
	result = lastVal
	return result
}

func (interp *Interpreter) cleanupLocals(env *Environment) {
	for name, val := range env.vars {
		if val.Kind == "struct" && !strings.HasPrefix(name, "_") && !env.protected[name] {
			var current = env
			var structDef *StructDef
			var ok bool
			for current != nil {
				if structDef, ok = current.structs[val.TypeName]; ok {
					break
				}
				current = current.parent
			}
			if structDef != nil {
				for _, fieldName := range structDef.Fields {
					fieldVal := val.Fields[fieldName]
					fmt.Printf("[del] %s.%s = %s (freed)\n", name, fieldName, valueToString(fieldVal))
				}
			}
		}
	}
}

func (interp *Interpreter) evalDel(del *DelCall, env *Environment) Value {
	switch target := del.Target.(type) {
	case *Identifier:
		if val, ok := env.getVar(target.Name); ok {
			if val.Kind == "struct" {
				fmt.Printf("[del] destroying '%s'\n", target.Name)
				for k, v := range val.Fields {
					fmt.Printf("  %s = %s (freed)\n", k, valueToString(v))
				}
			}
		}
	}
	return Value{Kind: "nil"}
}

func (interp *Interpreter) evalBinaryOp(left Value, op string, right Value) Value {
	// логика с short-circuit семантикой через truthiness
	if op == "&&" {
		return Value{Kind: "bool", BoolVal: isTruthy(left) && isTruthy(right)}
	}
	if op == "||" {
		return Value{Kind: "bool", BoolVal: isTruthy(left) || isTruthy(right)}
	}
	// равенство работает для number/string/bool/nil
	if op == "==" || op == "!=" {
		eq := valuesEqual(left, right)
		if op == "!=" {
			eq = !eq
		}
		return Value{Kind: "bool", BoolVal: eq}
	}
	if left.Kind == "number" && right.Kind == "number" {
		switch op {
		case "+":
			return Value{Kind: "number", NumVal: left.NumVal + right.NumVal}
		case "-":
			return Value{Kind: "number", NumVal: left.NumVal - right.NumVal}
		case "*":
			return Value{Kind: "number", NumVal: left.NumVal * right.NumVal}
		case "/":
			if right.NumVal == 0 {
				fail("Runtime error: division by zero\n")
			}
			return Value{Kind: "number", NumVal: left.NumVal / right.NumVal}
		case "%":
			if right.NumVal == 0 {
				fail("Runtime error: modulo by zero\n")
			}
			return Value{Kind: "number", NumVal: math.Mod(left.NumVal, right.NumVal)}
		case "div":
			if right.NumVal == 0 {
				fail("Runtime error: integer division by zero\n")
			}
			return Value{Kind: "number", NumVal: math.Trunc(left.NumVal / right.NumVal)}
		case "<":
			return Value{Kind: "bool", BoolVal: left.NumVal < right.NumVal}
		case ">":
			return Value{Kind: "bool", BoolVal: left.NumVal > right.NumVal}
		case "<=":
			return Value{Kind: "bool", BoolVal: left.NumVal <= right.NumVal}
		case ">=":
			return Value{Kind: "bool", BoolVal: left.NumVal >= right.NumVal}
		}
	}
	if left.Kind == "string" && right.Kind == "string" {
		switch op {
		case "+":
			return Value{Kind: "string", StrVal: left.StrVal + right.StrVal}
		case "<":
			return Value{Kind: "bool", BoolVal: left.StrVal < right.StrVal}
		case ">":
			return Value{Kind: "bool", BoolVal: left.StrVal > right.StrVal}
		case "<=":
			return Value{Kind: "bool", BoolVal: left.StrVal <= right.StrVal}
		case ">=":
			return Value{Kind: "bool", BoolVal: left.StrVal >= right.StrVal}
		}
	}
	fail("Runtime error: invalid binary op %s on %s and %s\n", op, left.Kind, right.Kind)
	return Value{Kind: "nil"}
}

func valuesEqual(a, b Value) bool {
	if a.Kind != b.Kind {
		return false
	}
	switch a.Kind {
	case "number":
		return a.NumVal == b.NumVal
	case "bool":
		return a.BoolVal == b.BoolVal
	case "string":
		return a.StrVal == b.StrVal
	case "nil":
		return true
	case "func":
		return a.Fn == b.Fn
	default:
		return false
	}
}

func isTruthy(v Value) bool {
	if v.Kind == "bool" {
		return v.BoolVal
	}
	if v.Kind == "number" && v.NumVal != 0 {
		return true
	}
	if v.Kind == "string" && v.StrVal != "" {
		return true
	}
	if v.Kind == "array" && len(v.Items) > 0 {
		return true
	}
	if v.Kind == "map" && len(v.MapVal) > 0 {
		return true
	}
	if v.Kind == "func" {
		return true
	}
	return false
}

// ========== JSON ==========

type jsonParser struct {
	s   []rune
	pos int
}

func parseJSON(s string) (Value, error) {
	p := &jsonParser{s: []rune(s)}
	p.skipWS()
	v, err := p.parseValue()
	if err != nil {
		return Value{}, err
	}
	p.skipWS()
	if p.pos != len(p.s) {
		return Value{}, fmt.Errorf("trailing data at %d", p.pos)
	}
	return v, nil
}

func (p *jsonParser) skipWS() {
	for p.pos < len(p.s) && (p.s[p.pos] == ' ' || p.s[p.pos] == '\t' || p.s[p.pos] == '\n' || p.s[p.pos] == '\r') {
		p.pos++
	}
}

func (p *jsonParser) parseValue() (Value, error) {
	if p.pos >= len(p.s) {
		return Value{}, fmt.Errorf("unexpected end of input")
	}
	switch p.s[p.pos] {
	case '{':
		return p.parseObject()
	case '[':
		return p.parseArray()
	case '"':
		s, err := p.parseJSONString()
		if err != nil {
			return Value{}, err
		}
		return Value{Kind: "string", StrVal: s}, nil
	case 't':
		return p.expectWord("true", Value{Kind: "bool", BoolVal: true})
	case 'f':
		return p.expectWord("false", Value{Kind: "bool", BoolVal: false})
	case 'n':
		return p.expectWord("null", Value{Kind: "nil"})
	default:
		return p.parseJSONNumber()
	}
}

func (p *jsonParser) expectWord(word string, val Value) (Value, error) {
	for _, c := range word {
		if p.pos >= len(p.s) || p.s[p.pos] != c {
			return Value{}, fmt.Errorf("bad literal at %d", p.pos)
		}
		p.pos++
	}
	return val, nil
}

func (p *jsonParser) parseObject() (Value, error) {
	p.pos++ // {
	m := make(map[string]Value)
	p.skipWS()
	if p.pos < len(p.s) && p.s[p.pos] == '}' {
		p.pos++
		return Value{Kind: "map", MapVal: m}, nil
	}
	for {
		p.skipWS()
		if p.pos >= len(p.s) || p.s[p.pos] != '"' {
			return Value{}, fmt.Errorf("object key must be a string at %d", p.pos)
		}
		key, err := p.parseJSONString()
		if err != nil {
			return Value{}, err
		}
		p.skipWS()
		if p.pos >= len(p.s) || p.s[p.pos] != ':' {
			return Value{}, fmt.Errorf("expected ':' at %d", p.pos)
		}
		p.pos++
		p.skipWS()
		val, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}
		m[key] = val
		p.skipWS()
		if p.pos >= len(p.s) {
			return Value{}, fmt.Errorf("unterminated object")
		}
		if p.s[p.pos] == ',' {
			p.pos++
			continue
		}
		if p.s[p.pos] == '}' {
			p.pos++
			return Value{Kind: "map", MapVal: m}, nil
		}
		return Value{}, fmt.Errorf("expected ',' or '}' at %d", p.pos)
	}
}

func (p *jsonParser) parseArray() (Value, error) {
	p.pos++ // [
	items := []Value{}
	p.skipWS()
	if p.pos < len(p.s) && p.s[p.pos] == ']' {
		p.pos++
		return Value{Kind: "array", Items: items}, nil
	}
	for {
		p.skipWS()
		val, err := p.parseValue()
		if err != nil {
			return Value{}, err
		}
		items = append(items, val)
		p.skipWS()
		if p.pos >= len(p.s) {
			return Value{}, fmt.Errorf("unterminated array")
		}
		if p.s[p.pos] == ',' {
			p.pos++
			continue
		}
		if p.s[p.pos] == ']' {
			p.pos++
			return Value{Kind: "array", Items: items}, nil
		}
		return Value{}, fmt.Errorf("expected ',' or ']' at %d", p.pos)
	}
}

func (p *jsonParser) parseJSONString() (string, error) {
	p.pos++ // opening quote
	var sb strings.Builder
	for {
		if p.pos >= len(p.s) {
			return "", fmt.Errorf("unterminated string")
		}
		c := p.s[p.pos]
		if c == '"' {
			p.pos++
			return sb.String(), nil
		}
		if c != '\\' {
			sb.WriteRune(c)
			p.pos++
			continue
		}
		p.pos++
		if p.pos >= len(p.s) {
			return "", fmt.Errorf("bad escape at end")
		}
		e := p.s[p.pos]
		p.pos++
		switch e {
		case '"':
			sb.WriteRune('"')
		case '\\':
			sb.WriteRune('\\')
		case '/':
			sb.WriteRune('/')
		case 'b':
			sb.WriteRune('\b')
		case 'f':
			sb.WriteRune('\f')
		case 'n':
			sb.WriteRune('\n')
		case 'r':
			sb.WriteRune('\r')
		case 't':
			sb.WriteRune('\t')
		case 'u':
			if p.pos+4 > len(p.s) {
				return "", fmt.Errorf("bad \\u escape")
			}
			var code int
			for _, h := range p.s[p.pos : p.pos+4] {
				code *= 16
				switch {
				case h >= '0' && h <= '9':
					code += int(h - '0')
				case h >= 'a' && h <= 'f':
					code += int(h-'a') + 10
				case h >= 'A' && h <= 'F':
					code += int(h-'A') + 10
				default:
					return "", fmt.Errorf("bad hex digit %q", h)
				}
			}
			p.pos += 4
			sb.WriteRune(rune(code))
		default:
			return "", fmt.Errorf("bad escape '\\%c'", e)
		}
	}
}

func (p *jsonParser) parseJSONNumber() (Value, error) {
	start := p.pos
	if p.pos < len(p.s) && p.s[p.pos] == '-' {
		p.pos++
	}
	for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
		p.pos++
	}
	if p.pos < len(p.s) && p.s[p.pos] == '.' {
		p.pos++
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			p.pos++
		}
	}
	if p.pos < len(p.s) && (p.s[p.pos] == 'e' || p.s[p.pos] == 'E') {
		p.pos++
		if p.pos < len(p.s) && (p.s[p.pos] == '+' || p.s[p.pos] == '-') {
			p.pos++
		}
		for p.pos < len(p.s) && p.s[p.pos] >= '0' && p.s[p.pos] <= '9' {
			p.pos++
		}
	}
	text := string(p.s[start:p.pos])
	f, err := strconv.ParseFloat(text, 64)
	if err != nil {
		return Value{}, fmt.Errorf("bad number %q", text)
	}
	return Value{Kind: "number", NumVal: f}, nil
}

func valueToJSON(v Value) string {
	switch v.Kind {
	case "number":
		return strconv.FormatFloat(v.NumVal, 'f', -1, 64)
	case "string":
		var sb strings.Builder
		sb.WriteRune('"')
		for _, c := range v.StrVal {
			switch c {
			case '"':
				sb.WriteString("\\\"")
			case '\\':
				sb.WriteString("\\\\")
			case '\n':
				sb.WriteString("\\n")
			case '\r':
				sb.WriteString("\\r")
			case '\t':
				sb.WriteString("\\t")
			default:
				if c < 0x20 {
					fmt.Fprintf(&sb, "\\u%04x", c)
				} else {
					sb.WriteRune(c)
				}
			}
		}
		sb.WriteRune('"')
		return sb.String()
	case "bool":
		if v.BoolVal {
			return "true"
		}
		return "false"
	case "nil":
		return "null"
	case "array":
		parts := make([]string, 0, len(v.Items))
		for _, item := range v.Items {
			parts = append(parts, valueToJSON(item))
		}
		return "[" + strings.Join(parts, ",") + "]"
	case "map":
		mapKeys := make([]string, 0, len(v.MapVal))
		for k := range v.MapVal {
			mapKeys = append(mapKeys, k)
		}
		sort.Strings(mapKeys)
		parts := make([]string, 0, len(mapKeys))
		for _, k := range mapKeys {
			parts = append(parts, valueToJSON(Value{Kind: "string", StrVal: k})+":"+valueToJSON(v.MapVal[k]))
		}
		return "{" + strings.Join(parts, ",") + "}"
	case "struct":
		fields := make([]string, 0, len(v.Fields))
		for k := range v.Fields {
			fields = append(fields, k)
		}
		sort.Strings(fields)
		parts := make([]string, 0, len(fields))
		for _, k := range fields {
			parts = append(parts, valueToJSON(Value{Kind: "string", StrVal: k})+":"+valueToJSON(v.Fields[k]))
		}
		return "{" + strings.Join(parts, ",") + "}"
	default:
		return "null"
	}
}

func valueToString(v Value) string {
	switch v.Kind {
	case "number":
		return strconv.FormatFloat(v.NumVal, 'f', -1, 64)
	case "bool":
		if v.BoolVal {
			return "true"
		}
		return "false"
	case "string":
		return v.StrVal
	case "func":
		if v.Fn != nil && v.Fn.Name != "" {
			return "<fn " + v.Fn.Name + ">"
		}
		return "<fn>"
	case "array":
		parts := make([]string, 0, len(v.Items))
		for _, item := range v.Items {
			parts = append(parts, valueToString(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
	case "map":
		mapKeys := make([]string, 0, len(v.MapVal))
		for k := range v.MapVal {
			mapKeys = append(mapKeys, k)
		}
		sort.Strings(mapKeys)
		parts := make([]string, 0, len(mapKeys))
		for _, k := range mapKeys {
			parts = append(parts, k+": "+valueToString(v.MapVal[k]))
		}
		return "{" + strings.Join(parts, ", ") + "}"
	case "struct":
		parts := make([]string, 0)
		for k, fv := range v.Fields {
			parts = append(parts, fmt.Sprintf("%s=%s", k, valueToString(fv)))
		}
		return fmt.Sprintf("%s{%s}", v.TypeName, strings.Join(parts, ", "))
	case "nil":
		return "nil"
	}
	return "?"
}

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "Usage:\n  codex <file.cx> [args...]   run a script\n  codex get <user/repo[@ver]>   prefetch packages\n")
		os.Exit(1)
	}

	if os.Args[1] == "get" {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: codex get <user/repo[@ver]> [...]\n")
			os.Exit(1)
		}
		for _, spec := range os.Args[2:] {
			ps, ok := parsePkgSpec(spec)
			if !ok {
				fmt.Fprintf(os.Stderr, "Bad package spec %q (want user/repo[@ver][/path.cx])\n", spec)
				os.Exit(1)
			}
			dest, err := ensurePackage(ps)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Get %q failed: %v\n", spec, err)
				os.Exit(1)
			}
			fmt.Printf("ok %s/%s -> %s\n", ps.user, ps.repo, dest)
		}
		return
	}

	sourceFile := os.Args[1]
	sourceBytes, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка чтения %s: %v\n", sourceFile, err)
		os.Exit(1)
	}

	lexer := NewLexer(string(sourceBytes))
	tokens := lexer.Tokenize()

	parser := NewParser(tokens)
	ast := parser.ParseProgram()

	interp := NewInterpreter()
	abs, err := filepath.Abs(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка пути %s: %v\n", sourceFile, err)
		os.Exit(1)
	}
	interp.mainDir = filepath.Dir(abs)
	interp.eval(ast, interp.globalEnv)
}
