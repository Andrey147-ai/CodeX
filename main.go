package main

import (
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
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
			start := l.pos
			startCol := tokCol + 1
			_ = startCol
			for l.pos < len(l.input) && l.input[l.pos] != '"' {
				if l.input[l.pos] == '\n' {
					l.advancePos(1)
				} else {
					l.advancePos(1)
				}
			}
			str := string(l.input[start:l.pos])
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

type ASTNode interface { isASTNode() }

type Program struct { Statements []ASTNode }
func (p *Program) isASTNode() {}

type VarDecl struct { Name string; Value ASTNode }
func (v *VarDecl) isASTNode() {}

type Assign struct { Name string; Value ASTNode }
func (a *Assign) isASTNode() {}

type FieldAssign struct { Object string; Field string; Value ASTNode }
func (f *FieldAssign) isASTNode() {}

type NumberLiteral struct { Value float64 }
func (n *NumberLiteral) isASTNode() {}

type BoolLiteral struct { Value bool }
func (b *BoolLiteral) isASTNode() {}

type StringLiteral struct { Value string }
func (s *StringLiteral) isASTNode() {}

type Identifier struct { Name string }
func (i *Identifier) isASTNode() {}

type BinaryOp struct { Left ASTNode; Op string; Right ASTNode }
func (b *BinaryOp) isASTNode() {}

type UnaryOp struct { Op string; Expr ASTNode }
func (u *UnaryOp) isASTNode() {}

type FuncCall struct { Name string; Args []ASTNode }
func (f *FuncCall) isASTNode() {}

type IfStatement struct { Condition ASTNode; Body []ASTNode; ElseBranch []ASTNode }
func (i *IfStatement) isASTNode() {}

type WhileLoop struct { Condition ASTNode; Body []ASTNode }
func (w *WhileLoop) isASTNode() {}

type BreakStmt struct{}
func (b *BreakStmt) isASTNode() {}

type ContinueStmt struct{}
func (c *ContinueStmt) isASTNode() {}

type ForLoop struct { Init ASTNode; Cond ASTNode; Post ASTNode; Body []ASTNode }
func (f *ForLoop) isASTNode() {}

type ArrayLiteral struct { Elements []ASTNode }
func (a *ArrayLiteral) isASTNode() {}

type IndexAccess struct { Target ASTNode; Index ASTNode }
func (i *IndexAccess) isASTNode() {}

type IndexAssign struct { Target ASTNode; Index ASTNode; Value ASTNode }
func (i *IndexAssign) isASTNode() {}

type MethodCall struct { Receiver ASTNode; Method string; Args []ASTNode }
func (m *MethodCall) isASTNode() {}

type FuncDef struct { Name string; Params []string; Body []ASTNode; RecvName string; RecvType string }
func (f *FuncDef) isASTNode() {}

type ReturnStmt struct { Value ASTNode }
func (r *ReturnStmt) isASTNode() {}

type StructDef struct { Name string; Fields []string }
func (s *StructDef) isASTNode() {}

type StructLiteral struct { Name string; Values []ASTNode }
func (s *StructLiteral) isASTNode() {}

type FieldAccess struct { Object ASTNode; Field string }
func (f *FieldAccess) isASTNode() {}

type DelCall struct { Target ASTNode }
func (d *DelCall) isASTNode() {}

// ========== PARSER ==========

type Parser struct {
	tokens []Token
	pos    int
}

func NewParser(tokens []Token) *Parser {
	return &Parser{tokens: tokens, pos: 0}
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
					index := p.parseExpr()
					p.expect(TOK_RBRACK)
					target = &IndexAccess{Target: target, Index: index}
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
	// метод: fn (recv Type) name(params)
	if p.peek().Type == TOK_LPAREN {
		p.next()
		fn.RecvName = p.expect(TOK_IDENT).Value
		fn.RecvType = p.expect(TOK_IDENT).Value
		p.expect(TOK_RPAREN)
	}
	fn.Name = p.expect(TOK_IDENT).Value
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
	fn.Params = params
	fn.Body = body
	return fn
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
		p.expect(TOK_LBRACE)
		for p.peek().Type != TOK_RBRACE && p.peek().Type != TOK_EOF {
			elseBranch = append(elseBranch, p.parseStatement())
		}
		p.expect(TOK_RBRACE)
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
	for p.peek().Type == TOK_STAR || p.peek().Type == TOK_SLASH {
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
		if p.peek().Type == TOK_LPAREN {
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
		if p.peek().Type == TOK_LBRACK {
			p.next()
			index := p.parseExpr()
			p.expect(TOK_RBRACK)
			node = &IndexAccess{Target: node, Index: index}
			continue
		}
		// вызов метода: obj.method(args)
		if p.peek().Type == TOK_LPAREN {
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

func (p *Parser) parsePrimary() ASTNode {
	p.skipNewlines()
	tok := p.peek()

	if tok.Type == TOK_TRUE {
		p.next()
		return &BoolLiteral{Value: true}
	}
	if tok.Type == TOK_FALSE {
		p.next()
		return &BoolLiteral{Value: false}
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
		
		// Фикс: Если за идентификатором сразу идёт открывающая скобка — это литерал структуры
		if p.peek().Type == TOK_LBRACE {
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

type Value struct {
	Kind     string
	NumVal   float64
	BoolVal  bool
	StrVal   string
	Items    []Value
	Fields   map[string]Value
	TypeName string
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
	globalEnv *Environment
	currentFn string
}

func NewInterpreter() *Interpreter {
	return &Interpreter{
		globalEnv: NewEnvironment(nil),
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
					fmt.Fprintf(os.Stderr, "Runtime error: variable '%s' is not a struct\n", n.Object)
					os.Exit(1)
				}
				if _, exists := obj.Fields[n.Field]; !exists {
					fmt.Fprintf(os.Stderr, "Runtime error: struct '%s' has no field '%s'\n", obj.TypeName, n.Field)
					os.Exit(1)
				}
				obj.Fields[n.Field] = val
				current.vars[n.Object] = obj
				return val
			}
			current = current.parent
		}
		fmt.Fprintf(os.Stderr, "Runtime error: undefined variable '%s'\n", n.Object)
		os.Exit(1)

	case *NumberLiteral:
		return Value{Kind: "number", NumVal: n.Value}

	case *BoolLiteral:
		return Value{Kind: "bool", BoolVal: n.Value}

	case *StringLiteral:
		return Value{Kind: "string", StrVal: n.Value}

	case *Identifier:
		if val, ok := env.getVar(n.Name); ok {
			return val
		}
		fmt.Fprintf(os.Stderr, "Runtime error: undefined variable '%s'\n", n.Name)
		os.Exit(1)

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
			fmt.Fprintf(os.Stderr, "Runtime error: unary - on %s\n", val.Kind)
			os.Exit(1)
		default:
			fmt.Fprintf(os.Stderr, "Runtime error: unknown unary op %s\n", n.Op)
			os.Exit(1)
		}

	case *FuncCall:
		return interp.evalFuncCall(n, env)

	case *MethodCall:
		return interp.evalMethodCall(n, env)

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

	case *ArrayLiteral:
		items := make([]Value, 0, len(n.Elements))
		for _, e := range n.Elements {
			items = append(items, interp.eval(e, env))
		}
		return Value{Kind: "array", Items: items}

	case *IndexAccess:
		arr := interp.eval(n.Target, env)
		if arr.Kind != "array" {
			fmt.Fprintf(os.Stderr, "Runtime error: indexing non-array (%s)\n", arr.Kind)
			os.Exit(1)
		}
		return arr.Items[interp.evalArrayIndex(n.Index, env, len(arr.Items))]

	case *IndexAssign:
		arr := interp.eval(n.Target, env)
		if arr.Kind != "array" {
			fmt.Fprintf(os.Stderr, "Runtime error: indexing non-array (%s)\n", arr.Kind)
			os.Exit(1)
		}
		val := interp.eval(n.Value, env)
		arr.Items[interp.evalArrayIndex(n.Index, env, len(arr.Items))] = val
		return val

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
			fmt.Fprintf(os.Stderr, "Runtime error: unknown struct '%s'\n", n.Name)
			os.Exit(1)
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
			fmt.Fprintf(os.Stderr, "Runtime error: field access on non-struct\n")
			os.Exit(1)
		}
		if val, ok := obj.Fields[n.Field]; ok {
			return val
		}
		fmt.Fprintf(os.Stderr, "Runtime error: struct has no field '%s'\n", n.Field)
		os.Exit(1)

	case *DelCall:
		return interp.evalDel(n, env)

	default:
		fmt.Fprintf(os.Stderr, "Runtime error: unknown node type %T\n", node)
		os.Exit(1)
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

func (interp *Interpreter) evalArrayIndex(node ASTNode, env *Environment, length int) int {
	v := interp.eval(node, env)
	if v.Kind != "number" {
		fmt.Fprintf(os.Stderr, "Runtime error: array index must be a number, got %s\n", v.Kind)
		os.Exit(1)
	}
	if v.NumVal != math.Trunc(v.NumVal) {
		fmt.Fprintf(os.Stderr, "Runtime error: array index must be an integer\n")
		os.Exit(1)
	}
	i := int(v.NumVal)
	if i < 0 || i >= length {
		fmt.Fprintf(os.Stderr, "Runtime error: index %d out of range (len %d)\n", i, length)
		os.Exit(1)
	}
	return i
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
			fmt.Fprintf(os.Stderr, "Runtime error: len() takes exactly 1 argument\n")
			os.Exit(1)
		}
		v := interp.eval(call.Args[0], env)
		switch v.Kind {
		case "array":
			return Value{Kind: "number", NumVal: float64(len(v.Items))}
		case "string":
			return Value{Kind: "number", NumVal: float64(len([]rune(v.StrVal)))}
		default:
			fmt.Fprintf(os.Stderr, "Runtime error: len() of %s\n", v.Kind)
			os.Exit(1)
		}
	}

	if call.Name == "push" {
		if len(call.Args) != 2 {
			fmt.Fprintf(os.Stderr, "Runtime error: push() takes exactly 2 arguments\n")
			os.Exit(1)
		}
		target, ok := call.Args[0].(*Identifier)
		if !ok {
			fmt.Fprintf(os.Stderr, "Runtime error: push() target must be a variable\n")
			os.Exit(1)
		}
		arr, ok := env.getVar(target.Name)
		if !ok || arr.Kind != "array" {
			fmt.Fprintf(os.Stderr, "Runtime error: push() target '%s' is not an array\n", target.Name)
			os.Exit(1)
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

	fn, ok := env.getFunc(call.Name)
	if !ok {
		fmt.Fprintf(os.Stderr, "Runtime error: undefined function '%s'\n", call.Name)
		os.Exit(1)
	}

	fnEnv := NewEnvironment(interp.globalEnv)
	for i, param := range fn.Params {
		if i < len(call.Args) {
			fnEnv.setVar(param, interp.eval(call.Args[i], env))
		} else {
			fnEnv.setVar(param, Value{Kind: "nil"})
		}
	}

	interp.currentFn = call.Name

	defer func() {
		if r := recover(); r != nil {
			if rv, ok := r.(*returnValue); ok {
				interp.cleanupLocals(fnEnv)
				result = rv.val
				return
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

func (interp *Interpreter) evalMethodCall(call *MethodCall, env *Environment) (result Value) {
	recv := interp.eval(call.Receiver, env)
	if recv.Kind != "struct" {
		fmt.Fprintf(os.Stderr, "Runtime error: method '%s' called on non-struct (%s)\n", call.Method, recv.Kind)
		os.Exit(1)
	}
	fn, ok := env.getMethod(recv.TypeName, call.Method)
	if !ok {
		fmt.Fprintf(os.Stderr, "Runtime error: struct '%s' has no method '%s'\n", recv.TypeName, call.Method)
		os.Exit(1)
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
				fmt.Fprintf(os.Stderr, "Runtime error: division by zero\n")
				os.Exit(1)
			}
			return Value{Kind: "number", NumVal: left.NumVal / right.NumVal}
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
	fmt.Fprintf(os.Stderr, "Runtime error: invalid binary op %s on %s and %s\n", op, left.Kind, right.Kind)
	os.Exit(1)
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
	return false
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
	case "array":
		parts := make([]string, 0, len(v.Items))
		for _, item := range v.Items {
			parts = append(parts, valueToString(item))
		}
		return "[" + strings.Join(parts, ", ") + "]"
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
		fmt.Fprintf(os.Stderr, "Использование: codex <имя_файла.cx>\n")
		os.Exit(1)
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
	interp.eval(ast, interp.globalEnv)
}
