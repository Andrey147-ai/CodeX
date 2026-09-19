package main

import (
	"archive/zip"
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
	"unicode"
)

// codexVersion is printed by `codex version`. Bump on release.
const codexVersion = "v0.22.0"

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
	TOK_COMMENT
	TOK_SWITCH
	TOK_CASE
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
			tokLine, tokCol := l.line, l.col
			l.advancePos(2)
			start := l.pos
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.advancePos(1)
			}
			l.tokens = append(l.tokens, Token{TOK_COMMENT, string(l.input[start:l.pos]), tokLine, tokCol})
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
			case "switch":
				typ = TOK_SWITCH
			case "case":
				typ = TOK_CASE
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
	Label     string
}

func (w *WhileLoop) isASTNode() {}

type BreakStmt struct {
	Label string
}

func (b *BreakStmt) isASTNode() {}

type ContinueStmt struct {
	Label string
}

func (c *ContinueStmt) isASTNode() {}

type ForLoop struct {
	Init  ASTNode
	Cond  ASTNode
	Post  ASTNode
	Body  []ASTNode
	Label string
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
	Label    string
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

type SwitchCase struct {
	Values []ASTNode
	Body   []ASTNode
}

type SwitchStmt struct {
	Target   ASTNode
	Cases    []*SwitchCase
	ElseBody []ASTNode
}

func (s *SwitchStmt) isASTNode() {}

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
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Type == TOK_STRUCT {
			j := i + 1
			for j < len(tokens) && (tokens[j].Type == TOK_NEWLINE || tokens[j].Type == TOK_COMMENT) {
				j++
			}
			if j < len(tokens) && tokens[j].Type == TOK_IDENT {
				p.structNames[tokens[j].Value] = true
			}
		}
	}
	return p
}

func (p *Parser) skipNewlines() {
	for p.pos < len(p.tokens) && (p.tokens[p.pos].Type == TOK_NEWLINE || p.tokens[p.pos].Type == TOK_COMMENT) {
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

	if tok.Type == TOK_COMMENT {
		p.next()
		return nil
	}

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
	case TOK_SWITCH:
		return p.parseSwitch()
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
		label := ""
		if p.peekRaw().Type == TOK_IDENT {
			label = p.next().Value
		}
		return &BreakStmt{Label: label}
	case TOK_CONTINUE:
		p.next()
		label := ""
		if p.peekRaw().Type == TOK_IDENT {
			label = p.next().Value
		}
		return &ContinueStmt{Label: label}
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
		for idx < len(p.tokens) && (p.tokens[idx].Type == TOK_NEWLINE || p.tokens[idx].Type == TOK_COMMENT) {
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
			// метка цикла: name: while/for ...
			if p.tokens[idx].Type == TOK_COLON {
				label := p.next().Value
				p.expect(TOK_COLON)
				lt := p.peek().Type
				if lt != TOK_WHILE && lt != TOK_FOR {
					tok := p.peek()
					fmt.Fprintf(os.Stderr, "Parser error at %d:%d: label only on loops\n", tok.Line, tok.Col)
					os.Exit(1)
				}
				loop := p.parseStatement()
				switch ln := loop.(type) {
				case *WhileLoop:
					ln.Label = label
				case *ForLoop:
					ln.Label = label
				case *ForIn:
					ln.Label = label
				}
				return loop
			}
			if p.tokens[idx].Type == TOK_ASSIGN {
				name := p.next().Value
				p.expect(TOK_ASSIGN)
				return &VarDecl{Name: name, Value: p.parseExpr()}
			}
			// составное присваивание: x += 1 → x = x + 1
			if p.tokens[idx].Type == TOK_PLUS || p.tokens[idx].Type == TOK_MINUS ||
				p.tokens[idx].Type == TOK_STAR || p.tokens[idx].Type == TOK_SLASH ||
				p.tokens[idx].Type == TOK_PERCENT || p.tokens[idx].Type == TOK_DIVINT {
				j := idx + 1
				for j < len(p.tokens) && (p.tokens[j].Type == TOK_NEWLINE || p.tokens[j].Type == TOK_COMMENT) {
					j++
				}
				if j < len(p.tokens) && p.tokens[j].Type == TOK_EQ {
					name := p.next().Value
					opTok := p.next()
					p.expect(TOK_EQ)
					val := p.parseExpr()
					return &Assign{Name: name, Value: &BinaryOp{
						Left:  &Identifier{Name: name},
						Op:    opTok.Value,
						Right: val,
					}}
				}
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
	skipGap := func(idx int) int {
		for idx < len(p.tokens) && (p.tokens[idx].Type == TOK_NEWLINE || p.tokens[idx].Type == TOK_COMMENT) {
			idx++
		}
		return idx
	}
	idx := skipGap(p.pos + 1)
	if idx >= len(p.tokens) || p.tokens[idx].Type != TOK_IDENT {
		return false
	}
	idx = skipGap(idx + 1)
	if idx >= len(p.tokens) || p.tokens[idx].Type != TOK_IDENT {
		return false
	}
	idx = skipGap(idx + 1)
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
	skipGap := func(idx int) int {
		for idx < len(p.tokens) && (p.tokens[idx].Type == TOK_NEWLINE || p.tokens[idx].Type == TOK_COMMENT) {
			idx++
		}
		return idx
	}
	idx := skipGap(p.pos + 1)
	seenBracket := false
	for {
		idx = skipGap(idx)
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
			idx = skipGap(idx)
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

func (p *Parser) parseSwitch() ASTNode {
	p.expect(TOK_SWITCH)
	target := p.parseExpr()
	p.expect(TOK_LBRACE)
	stmt := &SwitchStmt{Target: target}
	for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
		if p.peek().Type == TOK_CASE {
			p.next()
			sc := &SwitchCase{}
			for {
				sc.Values = append(sc.Values, p.parseExpr())
				if p.peek().Type == TOK_COMMA {
					p.next()
					continue
				}
				break
			}
			p.expect(TOK_COLON)
			p.expect(TOK_LBRACE)
			for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
				sc.Body = append(sc.Body, p.parseStatement())
			}
			p.expect(TOK_RBRACE)
			stmt.Cases = append(stmt.Cases, sc)
		} else if p.peek().Type == TOK_ELSE {
			p.next()
			if p.peek().Type == TOK_COLON {
				p.next()
			}
			p.expect(TOK_LBRACE)
			for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
				stmt.ElseBody = append(stmt.ElseBody, p.parseStatement())
			}
			p.expect(TOK_RBRACE)
		} else {
			tok := p.peek()
			fmt.Fprintf(os.Stderr, "Parser error at %d:%d: expected case or else, got %v\n", tok.Line, tok.Col, tok.Type)
			os.Exit(1)
		}
	}
	p.expect(TOK_RBRACE)
	return stmt
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
		if t != TOK_EQEQ && t != TOK_NEQ && t != TOK_IN {
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
	Ch       chan Value
	TaskCh   chan Value
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
	structNames map[string]bool
	frames      []string
	assertCount int
	noExit      bool
}

// the interpreter currently evaluating (single-threaded runtime);
// fail() reads its call stack for catchable error messages.
var activeInterp *Interpreter

// framesMu guards call-stack access: spawned tasks share the process.
var framesMu sync.Mutex

func NewInterpreter() *Interpreter {
	return &Interpreter{
		globalEnv:   NewEnvironment(nil),
		imported:    make(map[string]bool),
		structNames: make(map[string]bool),
	}
}

func (interp *Interpreter) eval(node ASTNode, env *Environment) Value {
	switch n := node.(type) {
	case *Program:
		for _, stmt := range n.Statements {
			switch s := stmt.(type) {
			case *StructDef:
				env.structs[s.Name] = s
				interp.structNames[s.Name] = true
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

	case *SwitchStmt:
		target := interp.eval(n.Target, env)
		for _, sc := range n.Cases {
			matched := false
			for _, v := range sc.Values {
				if valuesEqual(target, interp.eval(v, env)) {
					matched = true
					break
				}
			}
			if matched {
				return interp.evalBlock(sc.Body, env)
			}
		}
		if len(n.ElseBody) > 0 {
			return interp.evalBlock(n.ElseBody, env)
		}
		return Value{Kind: "nil"}

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
		if cont.Kind == "string" {
			runes := []rune(cont.StrVal)
			return Value{Kind: "string", StrVal: string(runes[interp.evalArrayIndex(n.Index, env, len(runes))])}
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
		if cont.Kind == "string" {
			fail("Runtime error: strings are immutable\n")
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
		panic(&breakSignal{Label: n.Label})

	case *ContinueStmt:
		panic(&continueSignal{Label: n.Label})

	case *ReturnStmt:
		val := interp.eval(n.Value, env)
		panic(&returnValue{val})

	case *StructDef:
		return Value{Kind: "nil"}

	case *FuncDef:
		if n.RecvType != "" {
			if env.methods[n.RecvType] == nil {
				env.methods[n.RecvType] = make(map[string]*FuncDef)
			}
			env.methods[n.RecvType][n.Name] = n
		} else {
			env.funcs[n.Name] = n
		}
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
			msg := ""
			switch sig := r.(type) {
			case *breakSignal:
				if sig.Label != "" {
					msg = fmt.Sprintf("Runtime error: 'break %s' outside loop", sig.Label)
				} else {
					msg = "Runtime error: 'break' outside loop"
				}
			case *continueSignal:
				if sig.Label != "" {
					msg = fmt.Sprintf("Runtime error: 'continue %s' outside loop", sig.Label)
				} else {
					msg = "Runtime error: 'continue' outside loop"
				}
			case *returnValue:
				msg = "Runtime error: 'return' outside function"
			case *runtimeError:
				msg = sig.msg
			default:
				panic(r)
			}
			if interp.noExit {
				panic(r)
			}
			fmt.Fprintln(os.Stderr, msg)
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

type breakSignal struct {
	Label string
}

type continueSignal struct {
	Label string
}

// runtimeError is a catchable script failure. All interpreter error paths
// panic with it instead of calling os.Exit, so try/catch can intercept;
// uncaught it prints exactly like the old fatal errors.
type runtimeError struct {
	msg string
}

func fail(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	msg = strings.TrimSuffix(msg, "\n")
	if activeInterp != nil {
		framesMu.Lock()
		frames := append([]string(nil), activeInterp.frames...)
		framesMu.Unlock()
		for _, f := range frames {
			msg += "\n  at " + f
		}
	}
	panic(&runtimeError{msg: msg})
}

// copyValue deep-copies containers across task/channel boundaries so
// concurrent tasks never share backing memory. Functions are shared
// (their AST is immutable).
func copyValue(v Value, depth int) Value {
	if depth > 100 {
		return Value{Kind: "nil"}
	}
	switch v.Kind {
	case "array":
		items := make([]Value, len(v.Items))
		for i, item := range v.Items {
			items[i] = copyValue(item, depth+1)
		}
		v.Items = items
		return v
	case "map":
		m := make(map[string]Value, len(v.MapVal))
		for k, item := range v.MapVal {
			m[k] = copyValue(item, depth+1)
		}
		v.MapVal = m
		return v
	case "struct":
		f := make(map[string]Value, len(v.Fields))
		for k, item := range v.Fields {
			f[k] = copyValue(item, depth+1)
		}
		v.Fields = f
		return v
	default:
		return v
	}
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
					if isBreak, mine := loopSignal(r, node.Label); mine {
						hitBreak = isBreak
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
					if isBreak, mine := loopSignal(r, node.Label); mine {
						hitBreak = isBreak
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

// loopSignal matches a recovered break/continue against this loop's label.
// Unlabeled signals hit the innermost loop; labeled ones fly past loops
// with other labels. Returns (isBreak, mine).
func loopSignal(r interface{}, label string) (bool, bool) {
	if bs, ok := r.(*breakSignal); ok {
		return true, bs.Label == "" || bs.Label == label
	}
	if cs, ok := r.(*continueSignal); ok {
		return false, cs.Label == "" || cs.Label == label
	}
	return false, false
}

func (interp *Interpreter) evalWhile(node *WhileLoop, env *Environment) Value {
	lastVal := Value{Kind: "nil"}
	for isTruthy(interp.eval(node.Condition, env)) {
		hitBreak := false
		func() {
			defer func() {
				if r := recover(); r != nil {
					if isBreak, mine := loopSignal(r, node.Label); mine {
						hitBreak = isBreak
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
		case "channel":
			return Value{Kind: "number", NumVal: float64(len(v.Ch))}
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

	if call.Name == "replace" {
		if len(call.Args) != 3 {
			fail("Runtime error: replace() takes exactly 3 arguments\n")
		}
		s := interp.eval(call.Args[0], env)
		oldSub := interp.eval(call.Args[1], env)
		newSub := interp.eval(call.Args[2], env)
		if s.Kind != "string" || oldSub.Kind != "string" || newSub.Kind != "string" {
			fail("Runtime error: replace() needs strings\n")
		}
		return Value{Kind: "string", StrVal: strings.ReplaceAll(s.StrVal, oldSub.StrVal, newSub.StrVal)}
	}

	if call.Name == "trim" {
		if len(call.Args) != 1 {
			fail("Runtime error: trim() takes exactly 1 argument\n")
		}
		s := interp.eval(call.Args[0], env)
		if s.Kind != "string" {
			fail("Runtime error: trim() needs a string, got %s\n", s.Kind)
		}
		return Value{Kind: "string", StrVal: strings.TrimSpace(s.StrVal)}
	}

	if call.Name == "repeat" {
		if len(call.Args) != 2 {
			fail("Runtime error: repeat() takes exactly 2 arguments\n")
		}
		s := interp.eval(call.Args[0], env)
		n := interp.eval(call.Args[1], env)
		if s.Kind != "string" || n.Kind != "number" {
			fail("Runtime error: repeat() needs (string, number)\n")
		}
		count := int(n.NumVal)
		if count < 0 || count > 100000 {
			fail("Runtime error: repeat() count out of range\n")
		}
		return Value{Kind: "string", StrVal: strings.Repeat(s.StrVal, count)}
	}

	if call.Name == "starts_with" || call.Name == "ends_with" {
		if len(call.Args) != 2 {
			fail("Runtime error: %s() takes exactly 2 arguments\n", call.Name)
		}
		s := interp.eval(call.Args[0], env)
		sub := interp.eval(call.Args[1], env)
		if s.Kind != "string" || sub.Kind != "string" {
			fail("Runtime error: %s() needs strings\n", call.Name)
		}
		if call.Name == "starts_with" {
			return Value{Kind: "bool", BoolVal: strings.HasPrefix(s.StrVal, sub.StrVal)}
		}
		return Value{Kind: "bool", BoolVal: strings.HasSuffix(s.StrVal, sub.StrVal)}
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

	if call.Name == "index_of" {
		if len(call.Args) != 2 {
			fail("Runtime error: index_of() takes exactly 2 arguments (s, sub)\n")
		}
		s := interp.eval(call.Args[0], env)
		sub := interp.eval(call.Args[1], env)
		if s.Kind != "string" || sub.Kind != "string" {
			fail("Runtime error: index_of() needs (string, string)\n")
		}
		if sub.StrVal == "" {
			return Value{Kind: "number", NumVal: 0}
		}
		// rune-aware: len() и s[i] в CodeX считают руны, не байты
		rs := []rune(s.StrVal)
		rsub := []rune(sub.StrVal)
		if len(rsub) > len(rs) {
			return Value{Kind: "number", NumVal: -1}
		}
		for i := 0; i+len(rsub) <= len(rs); i++ {
			match := true
			for j := 0; j < len(rsub); j++ {
				if rs[i+j] != rsub[j] {
					match = false
					break
				}
			}
			if match {
				return Value{Kind: "number", NumVal: float64(i)}
			}
		}
		return Value{Kind: "number", NumVal: -1}
	}

	if call.Name == "count" {
		if len(call.Args) != 2 {
			fail("Runtime error: count() takes exactly 2 arguments (s, sub)\n")
		}
		s := interp.eval(call.Args[0], env)
		sub := interp.eval(call.Args[1], env)
		if s.Kind != "string" || sub.Kind != "string" {
			fail("Runtime error: count() needs (string, string)\n")
		}
		if sub.StrVal == "" {
			return Value{Kind: "number", NumVal: 0}
		}
		return Value{Kind: "number", NumVal: float64(strings.Count(s.StrVal, sub.StrVal))}
	}

	if call.Name == "lines" {
		if len(call.Args) != 1 {
			fail("Runtime error: lines() takes exactly 1 argument\n")
		}
		s := interp.eval(call.Args[0], env)
		if s.Kind != "string" {
			fail("Runtime error: lines() needs a string, got %s\n", s.Kind)
		}
		normalized := strings.ReplaceAll(s.StrVal, "\r\n", "\n")
		parts := strings.Split(normalized, "\n")
		items := make([]Value, 0, len(parts))
		for _, part := range parts {
			items = append(items, Value{Kind: "string", StrVal: part})
		}
		return Value{Kind: "array", Items: items}
	}

	if call.Name == "sum" {
		if len(call.Args) != 1 {
			fail("Runtime error: sum() takes exactly 1 argument\n")
		}
		arr := interp.eval(call.Args[0], env)
		if arr.Kind != "array" {
			fail("Runtime error: sum() needs an array, got %s\n", arr.Kind)
		}
		total := 0.0
		for _, item := range arr.Items {
			if item.Kind != "number" {
				fail("Runtime error: sum() needs numbers, got %s\n", item.Kind)
			}
			total += item.NumVal
		}
		return Value{Kind: "number", NumVal: total}
	}

	if call.Name == "avg" {
		if len(call.Args) != 1 {
			fail("Runtime error: avg() takes exactly 1 argument\n")
		}
		arr := interp.eval(call.Args[0], env)
		if arr.Kind != "array" {
			fail("Runtime error: avg() needs an array, got %s\n", arr.Kind)
		}
		if len(arr.Items) == 0 {
			fail("Runtime error: avg() of empty array\n")
		}
		total := 0.0
		for _, item := range arr.Items {
			if item.Kind != "number" {
				fail("Runtime error: avg() needs numbers, got %s\n", item.Kind)
			}
			total += item.NumVal
		}
		return Value{Kind: "number", NumVal: total / float64(len(arr.Items))}
	}

	if call.Name == "unique" {
		if len(call.Args) != 1 {
			fail("Runtime error: unique() takes exactly 1 argument\n")
		}
		arr := interp.eval(call.Args[0], env)
		if arr.Kind != "array" {
			fail("Runtime error: unique() needs an array, got %s\n", arr.Kind)
		}
		out := make([]Value, 0, len(arr.Items))
		for _, item := range arr.Items {
			dup := false
			for _, seen := range out {
				if valuesEqual(item, seen) {
					dup = true
					break
				}
			}
			if !dup {
				out = append(out, item)
			}
		}
		return Value{Kind: "array", Items: out}
	}

	if call.Name == "extend" {
		if len(call.Args) != 2 {
			fail("Runtime error: extend() takes exactly 2 arguments (arr, other)\n")
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fail("Runtime error: extend() target must be a variable\n")
		}
		arr, ok := env.getVar(target.Name)
		if !ok || arr.Kind != "array" {
			fail("Runtime error: extend() target '%s' is not an array\n", target.Name)
		}
		other := interp.eval(call.Args[1], env)
		if other.Kind != "array" {
			fail("Runtime error: extend() needs (array, array)\n")
		}
		arr.Items = append(arr.Items, other.Items...)
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

	if call.Name == "choice" {
		if len(call.Args) != 1 {
			fail("Runtime error: choice() takes exactly 1 argument\n")
		}
		arr := interp.eval(call.Args[0], env)
		if arr.Kind != "array" {
			fail("Runtime error: choice() needs an array, got %s\n", arr.Kind)
		}
		if len(arr.Items) == 0 {
			fail("Runtime error: choice() of empty array\n")
		}
		return arr.Items[rand.Intn(len(arr.Items))]
	}

	if call.Name == "shuffle" {
		if len(call.Args) != 1 {
			fail("Runtime error: shuffle() takes exactly 1 argument\n")
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fail("Runtime error: shuffle() target must be a variable\n")
		}
		arr, ok := env.getVar(target.Name)
		if !ok || arr.Kind != "array" {
			fail("Runtime error: shuffle() target '%s' is not an array\n", target.Name)
		}
		rand.Shuffle(len(arr.Items), func(a, b int) { arr.Items[a], arr.Items[b] = arr.Items[b], arr.Items[a] })
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

	if call.Name == "values" {
		if len(call.Args) != 1 {
			fail("Runtime error: values() takes exactly 1 argument\n")
		}
		v := interp.eval(call.Args[0], env)
		if v.Kind != "map" {
			fail("Runtime error: values() needs a map, got %s\n", v.Kind)
		}
		ks := make([]string, 0, len(v.MapVal))
		for k := range v.MapVal {
			ks = append(ks, k)
		}
		sort.Strings(ks)
		items := make([]Value, 0, len(ks))
		for _, k := range ks {
			items = append(items, v.MapVal[k])
		}
		return Value{Kind: "array", Items: items}
	}

	if call.Name == "merge" {
		if len(call.Args) != 2 {
			fail("Runtime error: merge() takes exactly 2 arguments (m1, m2)\n")
		}
		a := interp.eval(call.Args[0], env)
		b := interp.eval(call.Args[1], env)
		if a.Kind != "map" || b.Kind != "map" {
			fail("Runtime error: merge() needs (map, map)\n")
		}
		out := make(map[string]Value, len(a.MapVal)+len(b.MapVal))
		for k, v := range a.MapVal {
			out[k] = v
		}
		for k, v := range b.MapVal {
			out[k] = v
		}
		return Value{Kind: "map", MapVal: out}
	}

	if call.Name == "delete_key" {
		if len(call.Args) != 2 {
			fail("Runtime error: delete_key() takes exactly 2 arguments (m, key)\n")
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fail("Runtime error: delete_key() target must be a variable\n")
		}
		m, ok := env.getVar(target.Name)
		if !ok || m.Kind != "map" {
			fail("Runtime error: delete_key() target '%s' is not a map\n", target.Name)
		}
		key := interp.eval(call.Args[1], env)
		if key.Kind != "string" {
			fail("Runtime error: delete_key() needs (map, string)\n")
		}
		_, existed := m.MapVal[key.StrVal]
		delete(m.MapVal, key.StrVal)
		current := env
		for current != nil {
			if _, found := current.vars[target.Name]; found {
				current.vars[target.Name] = m
				break
			}
			current = current.parent
		}
		return Value{Kind: "bool", BoolVal: existed}
	}

	if call.Name == "is_nil" || call.Name == "is_array" || call.Name == "is_map" || call.Name == "is_string" || call.Name == "is_num" {
		if len(call.Args) != 1 {
			fail("Runtime error: %s() takes exactly 1 argument\n", call.Name)
		}
		v := interp.eval(call.Args[0], env)
		switch call.Name {
		case "is_nil":
			return Value{Kind: "bool", BoolVal: v.Kind == "nil"}
		case "is_array":
			return Value{Kind: "bool", BoolVal: v.Kind == "array"}
		case "is_map":
			return Value{Kind: "bool", BoolVal: v.Kind == "map"}
		case "is_string":
			return Value{Kind: "bool", BoolVal: v.Kind == "string"}
		case "is_num":
			return Value{Kind: "bool", BoolVal: v.Kind == "number"}
		}
	}

	if call.Name == "chr" {
		if len(call.Args) != 1 {
			fail("Runtime error: chr() takes exactly 1 argument\n")
		}
		v := interp.eval(call.Args[0], env)
		if v.Kind != "number" {
			fail("Runtime error: chr() needs a number, got %s\n", v.Kind)
		}
		n := int(v.NumVal)
		if n < 0 || n > 0x10FFFF {
			fail("Runtime error: chr() code out of range\n")
		}
		return Value{Kind: "string", StrVal: string(rune(n))}
	}

	if call.Name == "ord" {
		if len(call.Args) != 1 {
			fail("Runtime error: ord() takes exactly 1 argument\n")
		}
		v := interp.eval(call.Args[0], env)
		if v.Kind != "string" {
			fail("Runtime error: ord() needs a string, got %s\n", v.Kind)
		}
		rs := []rune(v.StrVal)
		if len(rs) == 0 {
			fail("Runtime error: ord() of empty string\n")
		}
		return Value{Kind: "number", NumVal: float64(rs[0])}
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

	if call.Name == "pop" {
		if len(call.Args) != 1 {
			fail("Runtime error: pop() takes exactly 1 argument\n")
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fail("Runtime error: pop() target must be a variable\n")
		}
		arr, ok := env.getVar(target.Name)
		if !ok || arr.Kind != "array" {
			fail("Runtime error: pop() target '%s' is not an array\n", target.Name)
		}
		if len(arr.Items) == 0 {
			fail("Runtime error: pop() of empty array\n")
		}
		last := arr.Items[len(arr.Items)-1]
		arr.Items = arr.Items[:len(arr.Items)-1]
		current := env
		for current != nil {
			if _, found := current.vars[target.Name]; found {
				current.vars[target.Name] = arr
				break
			}
			current = current.parent
		}
		return last
	}

	if call.Name == "reverse" {
		if len(call.Args) != 1 {
			fail("Runtime error: reverse() takes exactly 1 argument\n")
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fail("Runtime error: reverse() target must be a variable\n")
		}
		arr, ok := env.getVar(target.Name)
		if !ok || arr.Kind != "array" {
			fail("Runtime error: reverse() target '%s' is not an array\n", target.Name)
		}
		for i, j := 0, len(arr.Items)-1; i < j; i, j = i+1, j-1 {
			arr.Items[i], arr.Items[j] = arr.Items[j], arr.Items[i]
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

	if call.Name == "file_size" {
		if len(call.Args) != 1 {
			fail("Runtime error: file_size() takes exactly 1 argument\n")
		}
		path := interp.eval(call.Args[0], env)
		if path.Kind != "string" {
			fail("Runtime error: file_size() needs a string path, got %s\n", path.Kind)
		}
		info, err := os.Stat(path.StrVal)
		if err != nil {
			fail("Runtime error: file_size() failed: %v\n", err)
		}
		return Value{Kind: "number", NumVal: float64(info.Size())}
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

	if call.Name == "http_post" {
		if len(call.Args) != 2 {
			fail("Runtime error: http_post() takes exactly 2 arguments (url, body)\n")
		}
		url := interp.eval(call.Args[0], env)
		payload := interp.eval(call.Args[1], env)
		if url.Kind != "string" || payload.Kind != "string" {
			fail("Runtime error: http_post() needs (string, string)\n")
		}
		client := &http.Client{Timeout: 15 * time.Second}
		resp, err := client.Post(url.StrVal, "text/plain; charset=utf-8", strings.NewReader(payload.StrVal))
		if err != nil {
			fail("Runtime error: http_post() failed: %v\n", err)
		}
		defer resp.Body.Close()
		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			fail("Runtime error: http_post() status %s\n", resp.Status)
		}
		body, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
		if err != nil {
			fail("Runtime error: http_post() read failed: %v\n", err)
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

	if call.Name == "assert" {
		if len(call.Args) < 1 || len(call.Args) > 2 {
			fail("Runtime error: assert() takes 1 or 2 arguments\n")
		}
		interp.assertCount++
		cond := interp.eval(call.Args[0], env)
		if !isTruthy(cond) {
			if len(call.Args) == 2 {
				msg := interp.eval(call.Args[1], env)
				fail("Runtime error: assert failed: %s\n", valueToString(msg))
			}
			fail("Runtime error: assert failed\n")
		}
		return Value{Kind: "nil"}
	}

	if call.Name == "date" {
		if len(call.Args) > 1 {
			fail("Runtime error: date() takes at most 1 argument\n")
		}
		t := time.Now()
		if len(call.Args) == 1 {
			v := interp.eval(call.Args[0], env)
			if v.Kind != "number" {
				fail("Runtime error: date() needs a timestamp, got %s\n", v.Kind)
			}
			t = time.Unix(int64(v.NumVal), 0)
		}
		return Value{Kind: "string", StrVal: t.Format("2006-01-02 15:04:05")}
	}

	if call.Name == "ls" {
		if len(call.Args) > 1 {
			fail("Runtime error: ls() takes at most 1 argument\n")
		}
		dir := "."
		if len(call.Args) == 1 {
			d := interp.eval(call.Args[0], env)
			if d.Kind != "string" {
				fail("Runtime error: ls() needs a string path, got %s\n", d.Kind)
			}
			dir = d.StrVal
		}
		entries, err := os.ReadDir(dir)
		if err != nil {
			fail("Runtime error: ls() failed: %v\n", err)
		}
		items := make([]Value, 0, len(entries))
		for _, e := range entries {
			items = append(items, Value{Kind: "string", StrVal: e.Name()})
		}
		sort.Slice(items, func(a, b int) bool { return items[a].StrVal < items[b].StrVal })
		return Value{Kind: "array", Items: items}
	}

	if call.Name == "mkdir" {
		if len(call.Args) != 1 {
			fail("Runtime error: mkdir() takes exactly 1 argument\n")
		}
		p := interp.eval(call.Args[0], env)
		if p.Kind != "string" {
			fail("Runtime error: mkdir() needs a string path, got %s\n", p.Kind)
		}
		if err := os.MkdirAll(p.StrVal, 0755); err != nil {
			fail("Runtime error: mkdir() failed: %v\n", err)
		}
		return Value{Kind: "nil"}
	}

	if call.Name == "remove" {
		if len(call.Args) != 1 {
			fail("Runtime error: remove() takes exactly 1 argument\n")
		}
		p := interp.eval(call.Args[0], env)
		if p.Kind != "string" {
			fail("Runtime error: remove() needs a string path, got %s\n", p.Kind)
		}
		if err := os.Remove(p.StrVal); err != nil {
			fail("Runtime error: remove() failed: %v\n", err)
		}
		return Value{Kind: "nil"}
	}

	if call.Name == "rename" {
		if len(call.Args) != 2 {
			fail("Runtime error: rename() takes exactly 2 arguments\n")
		}
		oldPath := interp.eval(call.Args[0], env)
		newPath := interp.eval(call.Args[1], env)
		if oldPath.Kind != "string" || newPath.Kind != "string" {
			fail("Runtime error: rename() needs strings\n")
		}
		if err := os.Rename(oldPath.StrVal, newPath.StrVal); err != nil {
			fail("Runtime error: rename() failed: %v\n", err)
		}
		return Value{Kind: "nil"}
	}

	if call.Name == "rand" {
		if len(call.Args) != 0 {
			fail("Runtime error: rand() takes no arguments\n")
		}
		return Value{Kind: "number", NumVal: rand.Float64()}
	}

	if call.Name == "randint" {
		if len(call.Args) != 2 {
			fail("Runtime error: randint() takes exactly 2 arguments\n")
		}
		lo := interp.eval(call.Args[0], env)
		hi := interp.eval(call.Args[1], env)
		if lo.Kind != "number" || hi.Kind != "number" {
			fail("Runtime error: randint() needs numbers\n")
		}
		a, b := int(math.Trunc(lo.NumVal)), int(math.Trunc(hi.NumVal))
		if b < a {
			fail("Runtime error: randint() max < min\n")
		}
		return Value{Kind: "number", NumVal: float64(a + rand.Intn(b-a+1))}
	}

	if call.Name == "seed" {
		if len(call.Args) != 1 {
			fail("Runtime error: seed() takes exactly 1 argument\n")
		}
		s := interp.eval(call.Args[0], env)
		if s.Kind != "number" {
			fail("Runtime error: seed() needs a number\n")
		}
		rand.Seed(int64(s.NumVal))
		return Value{Kind: "nil"}
	}

	if call.Name == "abs" || call.Name == "sqrt" || call.Name == "floor" || call.Name == "ceil" || call.Name == "round" {
		if len(call.Args) != 1 {
			fail("Runtime error: %s() takes exactly 1 argument\n", call.Name)
		}
		v := interp.eval(call.Args[0], env)
		if v.Kind != "number" {
			fail("Runtime error: %s() needs a number, got %s\n", call.Name, v.Kind)
		}
		switch call.Name {
		case "abs":
			return Value{Kind: "number", NumVal: math.Abs(v.NumVal)}
		case "sqrt":
			if v.NumVal < 0 {
				fail("Runtime error: sqrt() of negative\n")
			}
			return Value{Kind: "number", NumVal: math.Sqrt(v.NumVal)}
		case "floor":
			return Value{Kind: "number", NumVal: math.Floor(v.NumVal)}
		case "ceil":
			return Value{Kind: "number", NumVal: math.Ceil(v.NumVal)}
		default:
			return Value{Kind: "number", NumVal: math.Round(v.NumVal)}
		}
	}

	if call.Name == "pow" || call.Name == "min" || call.Name == "max" {
		if len(call.Args) != 2 {
			fail("Runtime error: %s() takes exactly 2 arguments\n", call.Name)
		}
		a := interp.eval(call.Args[0], env)
		b := interp.eval(call.Args[1], env)
		if a.Kind != "number" || b.Kind != "number" {
			fail("Runtime error: %s() needs numbers\n", call.Name)
		}
		switch call.Name {
		case "pow":
			return Value{Kind: "number", NumVal: math.Pow(a.NumVal, b.NumVal)}
		case "min":
			return Value{Kind: "number", NumVal: math.Min(a.NumVal, b.NumVal)}
		default:
			return Value{Kind: "number", NumVal: math.Max(a.NumVal, b.NumVal)}
		}
	}

	if call.Name == "range" {
		if len(call.Args) != 1 && len(call.Args) != 2 {
			fail("Runtime error: range() takes 1 or 2 arguments\n")
		}
		bounds := []int{}
		for _, a := range call.Args {
			v := interp.eval(a, env)
			if v.Kind != "number" || v.NumVal != math.Trunc(v.NumVal) {
				fail("Runtime error: range() needs integers\n")
			}
			bounds = append(bounds, int(v.NumVal))
		}
		lo, hi := 0, bounds[0]
		if len(bounds) == 2 {
			lo, hi = bounds[0], bounds[1]
		}
		items := []Value{}
		for i := lo; i < hi; i++ {
			items = append(items, Value{Kind: "number", NumVal: float64(i)})
		}
		return Value{Kind: "array", Items: items}
	}

	if call.Name == "map" || call.Name == "filter" || call.Name == "each" {
		if len(call.Args) != 2 {
			fail("Runtime error: %s() takes exactly 2 arguments (array, function)\n", call.Name)
		}
		arr := interp.eval(call.Args[0], env)
		fnVal := interp.eval(call.Args[1], env)
		if arr.Kind != "array" {
			fail("Runtime error: %s() needs an array, got %s\n", call.Name, arr.Kind)
		}
		if fnVal.Kind != "func" {
			fail("Runtime error: %s() needs a function, got %s\n", call.Name, fnVal.Kind)
		}
		closure := fnVal.Closure
		if closure == nil {
			closure = interp.globalEnv
		}
		out := []Value{}
		for _, item := range arr.Items {
			r := interp.invokeUserFunc(fnVal.Fn, closure, []Value{item}, call.Name)
			if call.Name == "filter" {
				if isTruthy(r) {
					out = append(out, item)
				}
			} else if call.Name == "map" {
				out = append(out, r)
			}
		}
		if call.Name == "each" {
			return Value{Kind: "nil"}
		}
		return Value{Kind: "array", Items: out}
	}

	if call.Name == "reduce" {
		if len(call.Args) != 3 {
			fail("Runtime error: reduce() takes exactly 3 arguments (array, function, init)\n")
		}
		arr := interp.eval(call.Args[0], env)
		fnVal := interp.eval(call.Args[1], env)
		if arr.Kind != "array" {
			fail("Runtime error: reduce() needs an array, got %s\n", arr.Kind)
		}
		if fnVal.Kind != "func" {
			fail("Runtime error: reduce() needs a function, got %s\n", fnVal.Kind)
		}
		closure := fnVal.Closure
		if closure == nil {
			closure = interp.globalEnv
		}
		acc := interp.eval(call.Args[2], env)
		for _, item := range arr.Items {
			acc = interp.invokeUserFunc(fnVal.Fn, closure, []Value{acc, item}, "reduce")
		}
		return acc
	}

	if call.Name == "find" {
		if len(call.Args) != 2 {
			fail("Runtime error: find() takes exactly 2 arguments (array, function)\n")
		}
		arr := interp.eval(call.Args[0], env)
		fnVal := interp.eval(call.Args[1], env)
		if arr.Kind != "array" {
			fail("Runtime error: find() needs an array, got %s\n", arr.Kind)
		}
		if fnVal.Kind != "func" {
			fail("Runtime error: find() needs a function, got %s\n", fnVal.Kind)
		}
		closure := fnVal.Closure
		if closure == nil {
			closure = interp.globalEnv
		}
		for _, item := range arr.Items {
			if isTruthy(interp.invokeUserFunc(fnVal.Fn, closure, []Value{item}, "find")) {
				return item
			}
		}
		return Value{Kind: "nil"}
	}

	if call.Name == "any" || call.Name == "all" {
		if len(call.Args) != 2 {
			fail("Runtime error: %s() takes exactly 2 arguments (array, function)\n", call.Name)
		}
		arr := interp.eval(call.Args[0], env)
		fnVal := interp.eval(call.Args[1], env)
		if arr.Kind != "array" {
			fail("Runtime error: %s() needs an array, got %s\n", call.Name, arr.Kind)
		}
		if fnVal.Kind != "func" {
			fail("Runtime error: %s() needs a function, got %s\n", call.Name, fnVal.Kind)
		}
		closure := fnVal.Closure
		if closure == nil {
			closure = interp.globalEnv
		}
		for _, item := range arr.Items {
			truth := isTruthy(interp.invokeUserFunc(fnVal.Fn, closure, []Value{item}, call.Name))
			if call.Name == "any" && truth {
				return Value{Kind: "bool", BoolVal: true}
			}
			if call.Name == "all" && !truth {
				return Value{Kind: "bool", BoolVal: false}
			}
		}
		if call.Name == "any" {
			return Value{Kind: "bool", BoolVal: false}
		}
		return Value{Kind: "bool", BoolVal: true}
	}

	if call.Name == "cwd" {
		if len(call.Args) != 0 {
			fail("Runtime error: cwd() takes no arguments\n")
		}
		dir, err := os.Getwd()
		if err != nil {
			fail("Runtime error: cwd() failed: %v\n", err)
		}
		return Value{Kind: "string", StrVal: dir}
	}

	if call.Name == "is_dir" {
		if len(call.Args) != 1 {
			fail("Runtime error: is_dir() takes exactly 1 argument\n")
		}
		path := interp.eval(call.Args[0], env)
		if path.Kind != "string" {
			fail("Runtime error: is_dir() needs a string path, got %s\n", path.Kind)
		}
		info, err := os.Stat(path.StrVal)
		if err != nil {
			return Value{Kind: "bool", BoolVal: false}
		}
		return Value{Kind: "bool", BoolVal: info.IsDir()}
	}

	if call.Name == "clock" {
		if len(call.Args) != 0 {
			fail("Runtime error: clock() takes no arguments\n")
		}
		return Value{Kind: "number", NumVal: float64(time.Now().UnixMilli())}
	}

	if call.Name == "run" {
		if len(call.Args) < 1 || len(call.Args) > 2 {
			fail("Runtime error: run() takes a program and optional args array\n")
		}
		prog := interp.eval(call.Args[0], env)
		if prog.Kind != "string" {
			fail("Runtime error: run() program must be a string, got %s\n", prog.Kind)
		}
		var argv []string
		if len(call.Args) == 2 {
			a := interp.eval(call.Args[1], env)
			if a.Kind != "array" {
				fail("Runtime error: run() args must be an array, got %s\n", a.Kind)
			}
			for _, item := range a.Items {
				if item.Kind != "string" {
					fail("Runtime error: run() args must be strings\n")
				}
				argv = append(argv, item.StrVal)
			}
		}
		cmd := exec.Command(prog.StrVal, argv...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout = &stdout
		cmd.Stderr = &stderr
		code := 0
		if err := cmd.Run(); err != nil {
			if ee, ok := err.(*exec.ExitError); ok {
				code = ee.ExitCode()
			} else {
				fail("Runtime error: run() failed to start: %v\n", err)
			}
		}
		return Value{Kind: "map", MapVal: map[string]Value{
			"code": Value{Kind: "number", NumVal: float64(code)},
			"out":  Value{Kind: "string", StrVal: stdout.String()},
			"err":  Value{Kind: "string", StrVal: stderr.String()},
		}}
	}

	if call.Name == "channel" {
		if len(call.Args) > 1 {
			fail("Runtime error: channel() takes at most 1 argument (capacity)\n")
		}
		capacity := 0
		if len(call.Args) == 1 {
			c := interp.eval(call.Args[0], env)
			if c.Kind != "number" || c.NumVal < 0 || c.NumVal != math.Trunc(c.NumVal) {
				fail("Runtime error: channel() capacity must be a non-negative integer\n")
			}
			capacity = int(c.NumVal)
		}
		return Value{Kind: "channel", Ch: make(chan Value, capacity)}
	}

	if call.Name == "send" {
		if len(call.Args) != 2 {
			fail("Runtime error: send() takes exactly 2 arguments\n")
		}
		ch := interp.eval(call.Args[0], env)
		if ch.Kind != "channel" {
			fail("Runtime error: send() needs a channel, got %s\n", ch.Kind)
		}
		ch.Ch <- copyValue(interp.eval(call.Args[1], env), 0)
		return Value{Kind: "nil"}
	}

	if call.Name == "recv" {
		if len(call.Args) != 1 {
			fail("Runtime error: recv() takes exactly 1 argument\n")
		}
		ch := interp.eval(call.Args[0], env)
		if ch.Kind != "channel" {
			fail("Runtime error: recv() needs a channel, got %s\n", ch.Kind)
		}
		return <-ch.Ch
	}

	if call.Name == "spawn" {
		if len(call.Args) < 1 {
			fail("Runtime error: spawn() takes a function and optional args\n")
		}
		fnVal := interp.eval(call.Args[0], env)
		if fnVal.Kind != "func" {
			fail("Runtime error: spawn() needs a function, got %s\n", fnVal.Kind)
		}
		argVals := make([]Value, 0, len(call.Args)-1)
		for _, a := range call.Args[1:] {
			argVals = append(argVals, copyValue(interp.eval(a, env), 0))
		}
		// Isolated interpreter: registries copied (AST is immutable),
		// variables NOT shared — everything crosses via args/channels.
		// Define functions before spawning; nested defs during flight race.
		child := NewInterpreter()
		child.mainDir = interp.mainDir
		for k, v := range interp.globalEnv.funcs {
			child.globalEnv.funcs[k] = v
		}
		for k, v := range interp.globalEnv.structs {
			child.globalEnv.structs[k] = v
		}
		for k, v := range interp.globalEnv.methods {
			child.globalEnv.methods[k] = v
		}
		done := make(chan Value, 1)
		fnDef := fnVal.Fn
		go func() {
			defer func() {
				if r := recover(); r != nil {
					msg := "task crash"
					if re, ok := r.(*runtimeError); ok {
						msg = re.msg
					}
					done <- Value{Kind: "string", StrVal: "task error: " + msg}
				}
			}()
			done <- child.invokeUserFunc(fnDef, child.globalEnv, argVals, "spawn")
		}()
		return Value{Kind: "task", TaskCh: done}
	}

	if call.Name == "wait" {
		if len(call.Args) != 1 {
			fail("Runtime error: wait() takes exactly 1 argument\n")
		}
		t := interp.eval(call.Args[0], env)
		if t.Kind != "task" {
			fail("Runtime error: wait() needs a task, got %s\n", t.Kind)
		}
		return <-t.TaskCh
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
	framesMu.Lock()
	interp.frames = append(interp.frames, callerName)
	framesMu.Unlock()
	defer func() {
		framesMu.Lock()
		interp.frames = interp.frames[:len(interp.frames)-1]
		framesMu.Unlock()
	}()
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

// ========== PACKAGES (GitHub) ==========//
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
	// + со строкой склеивает всё (числа, булевы, массивы через печать)
	if op == "+" && (left.Kind == "string" || right.Kind == "string") {
		return Value{Kind: "string", StrVal: valueToString(left) + valueToString(right)}
	}
	// равенство работает для number/string/bool/nil
	if op == "==" || op == "!=" {
		eq := valuesEqual(left, right)
		if op == "!=" {
			eq = !eq
		}
		return Value{Kind: "bool", BoolVal: eq}
	}
	if op == "in" {
		switch right.Kind {
		case "array":
			for _, item := range right.Items {
				if valuesEqual(left, item) {
					return Value{Kind: "bool", BoolVal: true}
				}
			}
			return Value{Kind: "bool", BoolVal: false}
		case "map":
			if left.Kind != "string" {
				return Value{Kind: "bool", BoolVal: false}
			}
			_, ok := right.MapVal[left.StrVal]
			return Value{Kind: "bool", BoolVal: ok}
		case "string":
			if left.Kind != "string" {
				fail("Runtime error: 'in' needs a string on the left for strings\n")
			}
			return Value{Kind: "bool", BoolVal: strings.Contains(right.StrVal, left.StrVal)}
		default:
			fail("Runtime error: 'in' needs array, map or string on the right\n")
		}
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
	return valuesEqualDepth(a, b, 0)
}

func valuesEqualDepth(a, b Value, depth int) bool {
	if depth > 50 {
		return false
	}
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
	case "array":
		if len(a.Items) != len(b.Items) {
			return false
		}
		for i := range a.Items {
			if !valuesEqualDepth(a.Items[i], b.Items[i], depth+1) {
				return false
			}
		}
		return true
	case "map":
		if len(a.MapVal) != len(b.MapVal) {
			return false
		}
		for k, av := range a.MapVal {
			bv, ok := b.MapVal[k]
			if !ok || !valuesEqualDepth(av, bv, depth+1) {
				return false
			}
		}
		return true
	case "struct":
		if a.TypeName != b.TypeName || len(a.Fields) != len(b.Fields) {
			return false
		}
		for k, av := range a.Fields {
			bv, ok := b.Fields[k]
			if !ok || !valuesEqualDepth(av, bv, depth+1) {
				return false
			}
		}
		return true
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
	if v.Kind == "channel" || v.Kind == "task" {
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
	case "channel":
		return "<channel>"
	case "task":
		return "<task>"
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

// ========== REPL ==========

func replDepth(s string) int {
	depth := 0
	inStr := false
	esc := false
	for _, c := range s {
		if inStr {
			if esc {
				esc = false
				continue
			}
			if c == '\\' {
				esc = true
				continue
			}
			if c == '"' {
				inStr = false
			}
			continue
		}
		switch c {
		case '"':
			inStr = true
		case '{', '(', '[':
			depth++
		case '}', ')', ']':
			depth--
		}
	}
	return depth
}

func runReplSnippet(interp *Interpreter, src string) {
	defer func() {
		if r := recover(); r != nil {
			switch e := r.(type) {
			case *runtimeError:
				fmt.Println(e.msg)
			case *breakSignal:
				fmt.Println("Runtime error: 'break' outside loop")
			case *continueSignal:
				fmt.Println("Runtime error: 'continue' outside loop")
			case *returnValue:
				fmt.Println("Runtime error: 'return' outside function")
			default:
				panic(r)
			}
		}
	}()
	parser := NewParser(NewLexer(src).Tokenize())
	// REPL parses per snippet: carry struct names from earlier snippets
	for name := range interp.structNames {
		parser.structNames[name] = true
	}
	prog := parser.ParseProgram()
	for _, stmt := range prog.Statements {
		// mini pre-pass like files get: defs register into globals
		switch s := stmt.(type) {
		case *StructDef:
			interp.globalEnv.structs[s.Name] = s
			interp.structNames[s.Name] = true
		case *FuncDef:
			if s.RecvType != "" {
				if interp.globalEnv.methods[s.RecvType] == nil {
					interp.globalEnv.methods[s.RecvType] = make(map[string]*FuncDef)
				}
				interp.globalEnv.methods[s.RecvType][s.Name] = s
			} else {
				interp.globalEnv.funcs[s.Name] = s
			}
		}
		v := interp.eval(stmt, interp.globalEnv)
		switch stmt.(type) {
		case *VarDecl, *Assign, *FieldAssign, *IndexAssign, *FuncDef,
			*StructDef, *ImportStmt, *BreakStmt, *ContinueStmt, *ReturnStmt,
			*IfStatement, *WhileLoop, *ForLoop, *ForIn, *TryCatch, *DelCall:
			// silent: declarations, assignments, control flow
		case *FuncCall:
			if fc, ok := stmt.(*FuncCall); ok && fc.Name == "print" {
				break
			}
			fmt.Println(valueToString(v))
		default:
			fmt.Println(valueToString(v))
		}
	}
}

func repl() {
	interp := NewInterpreter()
	activeInterp = interp
	if cwd, err := os.Getwd(); err == nil {
		interp.mainDir = cwd
	}
	fmt.Println("CodeX interactive — .exit quits, bad syntax ends the session")
	var buf strings.Builder
	for {
		if buf.Len() == 0 {
			fmt.Print(">> ")
		} else {
			fmt.Print(".. ")
		}
		line, err := stdinReader.ReadString('\n')
		if err != nil {
			fmt.Println()
			return
		}
		buf.WriteString(line)
		trimmed := strings.TrimSpace(buf.String())
		if trimmed == ".exit" {
			return
		}
		if trimmed == "" || strings.HasPrefix(trimmed, ".") {
			buf.Reset()
			if trimmed != "" {
				fmt.Println("unknown command (try .exit)")
			}
			continue
		}
		if replDepth(buf.String()) > 0 {
			continue
		}
		src := buf.String()
		buf.Reset()
		runReplSnippet(interp, src)
	}
}

// ========== FORMATTER (codex fmt) ==========

func quoteForFmt(s string) string {
	var sb strings.Builder
	sb.WriteRune('"')
	for _, c := range s {
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
		case 0x1b:
			sb.WriteString("\\e")
		default:
			sb.WriteRune(c)
		}
	}
	sb.WriteRune('"')
	return sb.String()
}

func formatSource(src string) string {
	toks := NewLexer(src).Tokenize()
	var sb strings.Builder
	indent := 0
	atStart := true
	var prev TokenType = TOK_EOF
	noSpaceAfter := false
	ind := func() {
		for i := 0; i < indent; i++ {
			sb.WriteString("    ")
		}
	}
	nextSig := func(i int) TokenType {
		for j := i + 1; j < len(toks); j++ {
			t := toks[j].Type
			if t != TOK_NEWLINE && t != TOK_COMMENT && t != TOK_SEMICOLON {
				return t
			}
		}
		return TOK_EOF
	}
	lastChar := func() byte {
		s := sb.String()
		if len(s) == 0 {
			return 0
		}
		return s[len(s)-1]
	}
	needSpace := func(cur TokenType) bool {
		if atStart || noSpaceAfter {
			return false
		}
		// Составное присваивание += -= *= /= %= div= : клей без пробела
		if cur == TOK_EQ {
			switch prev {
			case TOK_PLUS, TOK_MINUS, TOK_STAR, TOK_SLASH, TOK_PERCENT, TOK_DIVINT:
				return false
			}
		}
		switch cur {
		case TOK_RPAREN, TOK_RBRACK, TOK_DOT, TOK_COMMA, TOK_COLON:
			return false
		}
		switch prev {
		case TOK_LPAREN, TOK_LBRACK, TOK_LBRACE, TOK_DOT, TOK_NOT:
			return false
		}
		if cur == TOK_LPAREN {
			return !(prev == TOK_IDENT || prev == TOK_RPAREN || prev == TOK_RBRACK || prev == TOK_PRINT)
		}
		if cur == TOK_LBRACK {
			return !(prev == TOK_IDENT || prev == TOK_NUMBER || prev == TOK_STRING ||
				prev == TOK_RPAREN || prev == TOK_RBRACK || prev == TOK_TRUE ||
				prev == TOK_FALSE || prev == TOK_NIL)
		}
		return true
	}
	for i := 0; i < len(toks); i++ {
		t := toks[i]
		switch t.Type {
		case TOK_EOF:
		case TOK_NEWLINE:
			if ns := nextSig(i); ns == TOK_ELSE || ns == TOK_CATCH {
				s := sb.String()
				j := len(s)
				for j > 0 && s[j-1] == ' ' {
					j--
				}
				if j > 0 && s[j-1] == '}' {
					continue
				}
			}
			if lastChar() != '\n' && sb.Len() > 0 {
				sb.WriteString("\n")
			}
			atStart = true
			noSpaceAfter = false
		case TOK_COMMENT:
			if !atStart {
				sb.WriteString(" ")
			} else {
				ind()
			}
			sb.WriteString("//" + t.Value)
			atStart = false
			noSpaceAfter = false
		case TOK_SEMICOLON:
			sb.WriteString("; ")
			noSpaceAfter = true
			atStart = false
		case TOK_LBRACE:
			if !atStart {
				// map-литерал как аргумент: to_json({...}, без пробела после (,[,
				if prev != TOK_LPAREN && prev != TOK_LBRACK && prev != TOK_COMMA {
					sb.WriteString(" ")
				}
			} else {
				ind()
			}
			sb.WriteString("{")
			indent++
			atStart = false
			noSpaceAfter = false
		case TOK_RBRACE:
			indent--
			if indent < 0 {
				indent = 0
			}
			if lc := lastChar(); lc != '\n' && lc != '{' && sb.Len() > 0 {
				sb.WriteString("}")
			} else if lc == '\n' {
				ind()
				sb.WriteString("}")
			} else {
				sb.WriteString("}")
			}
			atStart = false
			noSpaceAfter = false
			if ns := nextSig(i); ns == TOK_ELSE || ns == TOK_CATCH {
				sb.WriteString(" ")
			}
		case TOK_ELSE, TOK_CATCH:
			if atStart {
				ind()
			}
			sb.WriteString(t.Value)
			atStart = false
			noSpaceAfter = false
		default:
			if atStart {
				ind()
			} else if needSpace(t.Type) {
				sb.WriteString(" ")
			}
			if t.Type == TOK_STRING {
				sb.WriteString(quoteForFmt(t.Value))
			} else {
				sb.WriteString(t.Value)
			}
			atStart = false
			// unary minus and NOT glue to the next token
			if t.Type == TOK_MINUS {
				switch prev {
				case TOK_EOF, TOK_LPAREN, TOK_LBRACK, TOK_COMMA, TOK_COLON,
					TOK_ASSIGN, TOK_PLUS, TOK_MINUS, TOK_STAR, TOK_SLASH,
					TOK_PERCENT, TOK_DIVINT, TOK_EQ, TOK_EQEQ, TOK_NEQ,
					TOK_LT, TOK_GT, TOK_LTE, TOK_GTE, TOK_AND, TOK_OR, TOK_NOT:
					noSpaceAfter = true
				default:
					noSpaceAfter = false
				}
			} else {
				noSpaceAfter = false
			}
			if t.Type == TOK_COMMA || t.Type == TOK_COLON {
				sb.WriteString(" ")
				noSpaceAfter = true
			}
		}
		if t.Type != TOK_NEWLINE && t.Type != TOK_COMMENT && t.Type != TOK_SEMICOLON {
			prev = t.Type
		}
	}
	out := sb.String()
	if len(out) > 0 && out[len(out)-1] != '\n' {
		out += "\n"
	}
	return out
}

// ========== TEST RUNNER (codex test) ==========

// codexTest runs *_test.cx files: one target file, a directory, or "."
// Each file gets a fresh interpreter. Returns the process exit code.
func codexTest(target string) int {
	var files []string
	info, err := os.Stat(target)
	if err != nil {
		fmt.Fprintf(os.Stderr, "test: %v\n", err)
		return 1
	}
	if !info.IsDir() {
		files = []string{target}
	} else {
		entries, err := os.ReadDir(target)
		if err != nil {
			fmt.Fprintf(os.Stderr, "test: %v\n", err)
			return 1
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), "_test.cx") {
				files = append(files, filepath.Join(target, e.Name()))
			}
		}
		sort.Strings(files)
	}
	if len(files) == 0 {
		fmt.Fprintln(os.Stderr, "test: no test files")
		return 1
	}
	failed := 0
	for _, f := range files {
		if runTestFile(f) {
			fmt.Printf("ok %s\n", f)
		} else {
			failed++
		}
	}
	if failed > 0 {
		fmt.Printf("FAIL (%d/%d files)\n", failed, len(files))
		return 1
	}
	fmt.Printf("ok (%d files)\n", len(files))
	return 0
}

func runTestFile(path string) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("FAIL %s: %v\n", path, err)
		return false
	}
	interp := NewInterpreter()
	activeInterp = interp
	interp.noExit = true
	abs, err := filepath.Abs(path)
	if err != nil {
		fmt.Printf("FAIL %s: %v\n", path, err)
		return false
	}
	interp.mainDir = filepath.Dir(abs)
	ok := true
	func() {
		defer func() {
			if r := recover(); r != nil {
				switch e := r.(type) {
				case *runtimeError:
					fmt.Printf("FAIL %s: %s\n", path, e.msg)
				case *breakSignal, *continueSignal, *returnValue:
					fmt.Printf("FAIL %s: control flow outside loop/function\n", path)
				default:
					panic(r)
				}
				ok = false
			}
		}()
		prog := NewParser(NewLexer(string(data)).Tokenize()).ParseProgram()
		interp.eval(prog, interp.globalEnv)
	}()
	if ok && interp.assertCount == 0 {
		fmt.Printf("WARN %s: no asserts\n", path)
	}
	return ok
}

// ========== HELP + NEW (DX для новичков) ==========

func printHelp() {
	fmt.Println("CodeX " + codexVersion + " — лёгкий язык программирования")
	fmt.Println("")
	fmt.Println("Использование:")
	fmt.Println("  codex.exe <file.cx>        запустить скрипт")
	fmt.Println("  codex.exe                  REPL (интерактивный режим)")
	fmt.Println("  codex.exe help             эта справка")
	fmt.Println("  codex.exe version          версия")
	fmt.Println("  codex.exe fmt <file.cx>    форматировать код")
	fmt.Println("  codex.exe fmt --check <file.cx>  проверить формат (для CI)")
	fmt.Println("  codex.exe test [dir]       тесты *_test.cx с assert()")
	fmt.Println("  codex.exe get <user/repo[@ver]>  скачать пакет с GitHub")
	fmt.Println("  codex.exe new [file.cx]    создать шаблон новичка (по умолч. main.cx)")
	fmt.Println("")
	fmt.Println("Примеры:")
	fmt.Println("  codex.exe new hello.cx")
	fmt.Println("  codex.exe hello.cx")
	fmt.Println("  codex.exe test tests")
	fmt.Println("")
	fmt.Println("Уроки: docs/tutorial-01-basics.md ... docs/tutorial-10-capstone.md")
}

func codexNew(target string) error {
	if !strings.HasSuffix(strings.ToLower(target), ".cx") {
		target += ".cx"
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("файл %s уже существует", target)
	}
	template := `// Привет! Это CodeX — запусти: codex.exe ` + filepath.Base(target) + `
name := input("Как тебя зовут? ")
print("Привет, " + name + "!")

// Массив + цикл
scores := [90, 80, 100]
total := 0
for s in scores {
    total += s
}
print("Средний балл: ", total / len(scores))

// Функция
fn double(n) {
    return n * 2
}
print("double(21) = ", double(21))
`
	if err := os.WriteFile(target, []byte(template), 0644); err != nil {
		return err
	}
	fmt.Printf("ok %s — запусти: codex.exe %s\n", target, target)
	return nil
}

func main() {
	if len(os.Args) < 2 {
		repl()
		return
	}

	if os.Args[1] == "version" || os.Args[1] == "--version" {
		fmt.Println("CodeX " + codexVersion)
		return
	}

	if os.Args[1] == "help" || os.Args[1] == "--help" || os.Args[1] == "-h" {
		printHelp()
		return
	}

	if os.Args[1] == "new" || os.Args[1] == "init" {
		target := "main.cx"
		if len(os.Args) > 2 {
			target = os.Args[2]
		}
		if err := codexNew(target); err != nil {
			fmt.Fprintf(os.Stderr, "new: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if os.Args[1] == "fmt" {
		if len(os.Args) < 3 {
			fmt.Fprintf(os.Stderr, "Usage: codex fmt [--check] <file.cx>\n")
			os.Exit(1)
		}
		check := false
		fileArg := os.Args[2]
		if os.Args[2] == "--check" {
			if len(os.Args) < 4 {
				fmt.Fprintf(os.Stderr, "Usage: codex fmt --check <file.cx>\n")
				os.Exit(1)
			}
			check = true
			fileArg = os.Args[3]
		}
		data, err := os.ReadFile(fileArg)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Ошибка чтения %s: %v\n", fileArg, err)
			os.Exit(1)
		}
		formatted := formatSource(string(data))
		if check {
			normalized := strings.ReplaceAll(string(data), "\r\n", "\n")
			if formatted != normalized {
				fmt.Fprintf(os.Stderr, "fmt --check: %s needs formatting\n", fileArg)
				os.Exit(1)
			}
			fmt.Printf("ok %s\n", fileArg)
			return
		}
		fmt.Print(formatted)
		return
	}

	if os.Args[1] == "test" {
		target := "."
		if len(os.Args) > 2 {
			target = os.Args[2]
		}
		os.Exit(codexTest(target))
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
	activeInterp = interp
	abs, err := filepath.Abs(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Ошибка пути %s: %v\n", sourceFile, err)
		os.Exit(1)
	}
	interp.mainDir = filepath.Dir(abs)
	interp.eval(ast, interp.globalEnv)
}
