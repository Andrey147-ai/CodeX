package main

import (
	"fmt"
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
)

type Token struct {
	Type  TokenType
	Value string
}

type Lexer struct {
	input  []rune
	pos    int
	tokens []Token
}

func NewLexer(input string) *Lexer {
	return &Lexer{input: []rune(input), pos: 0}
}

func (l *Lexer) Tokenize() []Token {
	for l.pos < len(l.input) {
		ch := l.input[l.pos]

		if unicode.IsSpace(ch) {
			if ch == '\n' {
				l.tokens = append(l.tokens, Token{TOK_NEWLINE, "\n"})
			}
			l.pos++
			continue
		}

		if ch == '/' && l.pos+1 < len(l.input) && l.input[l.pos+1] == '/' {
			for l.pos < len(l.input) && l.input[l.pos] != '\n' {
				l.pos++
			}
			continue
		}

		if ch == '"' {
			l.pos++
			start := l.pos
			for l.pos < len(l.input) && l.input[l.pos] != '"' {
				l.pos++
			}
			str := string(l.input[start:l.pos])
			l.tokens = append(l.tokens, Token{TOK_STRING, str})
			l.pos++
			continue
		}

		if unicode.IsLetter(ch) || ch == '_' {
			start := l.pos
			for l.pos < len(l.input) && (unicode.IsLetter(l.input[l.pos]) || unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '_') {
				l.pos++
			}
			word := string(l.input[start:l.pos])
			switch word {
			case "if":
				l.tokens = append(l.tokens, Token{TOK_IF, word})
			case "else":
				l.tokens = append(l.tokens, Token{TOK_ELSE, word})
			case "fn":
				l.tokens = append(l.tokens, Token{TOK_FN, word})
			case "return":
				l.tokens = append(l.tokens, Token{TOK_RETURN, word})
			case "print":
				l.tokens = append(l.tokens, Token{TOK_PRINT, word})
			case "struct":
				l.tokens = append(l.tokens, Token{TOK_STRUCT, word})
			case "del":
				l.tokens = append(l.tokens, Token{TOK_DEL, word})
			default:
				l.tokens = append(l.tokens, Token{TOK_IDENT, word})
			}
			continue
		}

		if unicode.IsDigit(ch) {
			start := l.pos
			for l.pos < len(l.input) && (unicode.IsDigit(l.input[l.pos]) || l.input[l.pos] == '.') {
				l.pos++
			}
			l.tokens = append(l.tokens, Token{TOK_NUMBER, string(l.input[start:l.pos])})
			continue
		}

		switch ch {
		case ':':
			if l.pos+1 < len(l.input) && l.input[l.pos+1] == '=' {
				l.tokens = append(l.tokens, Token{TOK_ASSIGN, ":="})
				l.pos += 2
			} else {
				l.tokens = append(l.tokens, Token{TOK_COLON, ":"})
				l.pos++
			}
		case '=':
			l.tokens = append(l.tokens, Token{TOK_EQ, "="})
			l.pos++
		case '+':
			l.tokens = append(l.tokens, Token{TOK_PLUS, "+"})
			l.pos++
		case '-':
			l.tokens = append(l.tokens, Token{TOK_MINUS, "-"})
			l.pos++
		case '*':
			l.tokens = append(l.tokens, Token{TOK_STAR, "*"})
			l.pos++
		case '/':
			l.tokens = append(l.tokens, Token{TOK_SLASH, "/"})
			l.pos++
		case '(':
			l.tokens = append(l.tokens, Token{TOK_LPAREN, "("})
			l.pos++
		case ')':
			l.tokens = append(l.tokens, Token{TOK_RPAREN, ")"})
			l.pos++
		case '{':
			l.tokens = append(l.tokens, Token{TOK_LBRACE, "{"})
			l.pos++
		case '}':
			l.tokens = append(l.tokens, Token{TOK_RBRACE, "}"})
			l.pos++
		case ',':
			l.tokens = append(l.tokens, Token{TOK_COMMA, ","})
			l.pos++
		case ';':
			l.tokens = append(l.tokens, Token{TOK_SEMICOLON, ";"})
			l.pos++
		case '.':
			l.tokens = append(l.tokens, Token{TOK_DOT, "."})
			l.pos++
		default:
			fmt.Fprintf(os.Stderr, "Lexer error: unknown character '%c'\n", ch)
			os.Exit(1)
		}
	}
	l.tokens = append(l.tokens, Token{TOK_EOF, ""})
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

type StringLiteral struct { Value string }
func (s *StringLiteral) isASTNode() {}

type Identifier struct { Name string }
func (i *Identifier) isASTNode() {}

type BinaryOp struct { Left ASTNode; Op string; Right ASTNode }
func (b *BinaryOp) isASTNode() {}

type FuncCall struct { Name string; Args []ASTNode }
func (f *FuncCall) isASTNode() {}

type IfStatement struct { Condition ASTNode; Body []ASTNode; ElseBranch []ASTNode }
func (i *IfStatement) isASTNode() {}

type FuncDef struct { Name string; Params []string; Body []ASTNode }
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
		return Token{TOK_EOF, ""}
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
		fmt.Fprintf(os.Stderr, "Parser error: expected %v, got %v\n", typ, tok)
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
	name := p.expect(TOK_IDENT).Value
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
	return &FuncDef{Name: name, Params: params, Body: body}
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
	left := p.parsePrimary()

	if ident, ok := left.(*Identifier); ok {
		if p.peek().Type == TOK_LPAREN {
			return p.parseFuncCallFinish(ident.Name)
		}
	}

	for p.peek().Type == TOK_PLUS || p.peek().Type == TOK_MINUS || p.peek().Type == TOK_STAR || p.peek().Type == TOK_SLASH {
		opTok := p.next()
		right := p.parsePrimary()
		left = &BinaryOp{Left: left, Op: opTok.Value, Right: right}
	}

	return left
}

func (p *Parser) parsePrimary() ASTNode {
	p.skipNewlines()
	tok := p.peek()

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

	fmt.Fprintf(os.Stderr, "Parser error: unexpected token %v\n", tok)
	os.Exit(1)
	return nil
}

// ========== INTERPRETER ==========

type Value struct {
	Kind     string
	NumVal   float64
	StrVal   string
	Fields   map[string]Value
	TypeName string
}

type Environment struct {
	vars    map[string]Value
	funcs   map[string]*FuncDef
	structs map[string]*StructDef
	parent  *Environment
}

func NewEnvironment(parent *Environment) *Environment {
	return &Environment{
		vars:    make(map[string]Value),
		funcs:   make(map[string]*FuncDef),
		structs: make(map[string]*StructDef),
		parent:  parent,
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
				env.funcs[s.Name] = s
			}
		}
		var lastVal Value
		for _, stmt := range n.Statements {
			lastVal = interp.eval(stmt, env)
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
		obj, ok := env.getVar(n.Object)
		if !ok || obj.Kind != "struct" {
			fmt.Fprintf(os.Stderr, "Runtime error: variable '%s' is not a struct\n", n.Object)
			os.Exit(1)
		}
		if _, exists := obj.Fields[n.Field]; !exists {
			fmt.Fprintf(os.Stderr, "Runtime error: struct '%s' has no field '%s'\n", obj.TypeName, n.Field)
			os.Exit(1)
		}
		obj.Fields[n.Field] = val
		env.setVar(n.Object, obj) 
		return val

	case *NumberLiteral:
		return Value{Kind: "number", NumVal: n.Value}

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

	case *FuncCall:
		return interp.evalFuncCall(n, env)

	case *IfStatement:
		cond := interp.eval(n.Condition, env)
		if isTruthy(cond) {
			return interp.evalBlock(n.Body, env)
		} else if len(n.ElseBranch) > 0 {
			return interp.evalBlock(n.ElseBranch, env)
		}
		return Value{Kind: "nil"}

	case *ReturnStmt:
		val := interp.eval(n.Value, env)
		interp.cleanupLocals(env)
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

func (interp *Interpreter) evalBlock(stmts []ASTNode, env *Environment) Value {
	blockEnv := NewEnvironment(env)
	var lastVal Value
	for _, stmt := range stmts {
		lastVal = interp.eval(stmt, blockEnv)
	}
	interp.cleanupLocals(blockEnv)
	return lastVal
}

type returnValue struct {
	val Value
}

func (interp *Interpreter) evalFuncCall(call *FuncCall, env *Environment) Value {
	if call.Name == "print" {
		for _, arg := range call.Args {
			val := interp.eval(arg, env)
			fmt.Print(valueToString(val))
		}
		fmt.Println()
		return Value{Kind: "nil"}
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
			if _, ok := r.(*returnValue); ok {
				interp.cleanupLocals(fnEnv)
				return
			}
			panic(r)
		}
	}()

	var lastVal Value
	for _, stmt := range fn.Body {
		lastVal = interp.eval(stmt, fnEnv)
	}
	interp.cleanupLocals(fnEnv)
	return lastVal
}

func (interp *Interpreter) cleanupLocals(env *Environment) {
	for name, val := range env.vars {
		if val.Kind == "struct" && !strings.HasPrefix(name, "_") {
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
		}
	}
	if left.Kind == "string" && right.Kind == "string" && op == "+" {
		return Value{Kind: "string", StrVal: left.StrVal + right.StrVal}
	}
	fmt.Fprintf(os.Stderr, "Runtime error: invalid binary op %s on %s and %s\n", op, left.Kind, right.Kind)
	os.Exit(1)
	return Value{Kind: "nil"}
}

func isTruthy(v Value) bool {
	if v.Kind == "number" && v.NumVal != 0 {
		return true
	}
	if v.Kind == "string" && v.StrVal != "" {
		return true
	}
	return false
}

func valueToString(v Value) string {
	switch v.Kind {
	case "number":
		return strconv.FormatFloat(v.NumVal, 'f', -1, 64)
	case "string":
		return v.StrVal
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
