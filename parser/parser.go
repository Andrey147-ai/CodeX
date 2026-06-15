package parser

import (
	"fmt"
	"strings"

	"CodeX/lexer"
)

// AST node types
type NodeType int

const (
	NodeProgram NodeType = iota
	NodeStructDecl
	NodeFnDecl
	NodeBlock
	NodeCallExpr
	NodeFieldAccess
	NodeVarDecl
	NodeIdent
	NodeNumber
)

// ASTNode represents any node in the abstract syntax tree.
type ASTNode struct {
	Type     NodeType
	Name     string
	Value    string
	Children []*ASTNode
	Fields   []*ASTNode // for struct field declarations
	Params   []*ASTNode // for function parameters
	Body     []*ASTNode // for function/block body statements
}

// Parser holds the state of the parser.
type Parser struct {
	lexer   *lexer.Lexer
	tokens  []lexer.Token
	pos     int
	current lexer.Token
}

// New creates a new Parser for the given lexer.
func New(l *lexer.Lexer) *Parser {
	tokens := l.Tokenize()
	p := &Parser{
		lexer:  l,
		tokens: tokens,
		pos:    0,
	}
	if len(tokens) > 0 {
		p.current = tokens[0]
	}
	return p
}

// advance moves to the next token.
func (p *Parser) advance() {
	p.pos++
	if p.pos < len(p.tokens) {
		p.current = p.tokens[p.pos]
	}
}

// peek returns the next token without consuming it.
func (p *Parser) peek() lexer.Token {
	if p.pos+1 < len(p.tokens) {
		return p.tokens[p.pos+1]
	}
	return lexer.Token{Type: lexer.TokenEOF}
}

// expect checks if the current token matches the given type and advances if so.
func (p *Parser) expect(tokenType lexer.TokenType) (lexer.Token, error) {
	if p.current.Type != tokenType {
		return p.current, fmt.Errorf("expected %s but got %s (%q) at line %d, col %d",
			tokenType, p.current.Type, p.current.Literal, p.current.Line, p.current.Column)
	}
	tok := p.current
	p.advance()
	return tok, nil
}

// ParseProgram parses the entire source and returns the root AST node.
func (p *Parser) ParseProgram() (*ASTNode, error) {
	program := &ASTNode{Type: NodeProgram, Name: "program"}

	for p.current.Type != lexer.TokenEOF {
		node, err := p.parseTopLevel()
		if err != nil {
			return nil, err
		}
		if node != nil {
			program.Children = append(program.Children, node)
		}
	}

	return program, nil
}

// parseTopLevel parses a top-level declaration (struct, fn, or import).
func (p *Parser) parseTopLevel() (*ASTNode, error) {
	switch p.current.Type {
	case lexer.TokenStruct:
		return p.parseStructDecl()
	case lexer.TokenFn:
		return p.parseFnDecl()
	case lexer.TokenImport:
		return p.parseImport()
	default:
		// Skip unknown top-level tokens
		p.advance()
		return nil, nil
	}
}

// parseStructDecl parses a struct declaration: struct Name { fields }
func (p *Parser) parseStructDecl() (*ASTNode, error) {
	p.advance() // consume 'struct'

	name, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, fmt.Errorf("struct declaration: %w", err)
	}

	node := &ASTNode{Type: NodeStructDecl, Name: name.Literal}

	_, err = p.expect(lexer.TokenLBrace)
	if err != nil {
		return nil, fmt.Errorf("struct %s: %w", name.Literal, err)
	}

	// Parse fields until '}'
	for p.current.Type != lexer.TokenRBrace && p.current.Type != lexer.TokenEOF {
		// Check for 'del' keyword (destructor method inside struct)
		if p.current.Type == lexer.TokenDel {
			delNode, err := p.parseDelDecl()
			if err != nil {
				return nil, err
			}
			node.Children = append(node.Children, delNode)
			continue
		}

		// Parse field: name : type
		fieldName, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, fmt.Errorf("struct %s field: %w", name.Literal, err)
		}

		_, err = p.expect(lexer.TokenColon)
		if err != nil {
			return nil, fmt.Errorf("struct %s field %s: %w", name.Literal, fieldName.Literal, err)
		}

		fieldType, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, fmt.Errorf("struct %s field %s type: %w", name.Literal, fieldName.Literal, err)
		}

		field := &ASTNode{
			Type:  NodeVarDecl,
			Name:  fieldName.Literal,
			Value: fieldType.Literal,
		}
		node.Fields = append(node.Fields, field)

		// Optional comma or semicolon
		if p.current.Type == lexer.TokenComma || p.current.Type == lexer.TokenSemicolon {
			p.advance()
		}
	}

	_, err = p.expect(lexer.TokenRBrace)
	if err != nil {
		return nil, fmt.Errorf("struct %s: %w", name.Literal, err)
	}

	return node, nil
}

// parseDelDecl parses a del (destructor) declaration inside a struct:
// del() { statements }
func (p *Parser) parseDelDecl() (*ASTNode, error) {
	p.advance() // consume 'del'

	_, err := p.expect(lexer.TokenLParen)
	if err != nil {
		return nil, fmt.Errorf("del declaration: %w", err)
	}

	_, err = p.expect(lexer.TokenRParen)
	if err != nil {
		return nil, fmt.Errorf("del declaration: %w", err)
	}

	node := &ASTNode{Type: NodeFnDecl, Name: "del"}

	// Parse body block
	body, err := p.parseBlock()
	if err != nil {
		return nil, fmt.Errorf("del body: %w", err)
	}
	node.Body = body

	return node, nil
}

// parseFnDecl parses a function declaration: fn name(params) { body }
func (p *Parser) parseFnDecl() (*ASTNode, error) {
	p.advance() // consume 'fn'

	name, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, fmt.Errorf("function declaration: %w", err)
	}

	node := &ASTNode{Type: NodeFnDecl, Name: name.Literal}

	_, err = p.expect(lexer.TokenLParen)
	if err != nil {
		return nil, fmt.Errorf("function %s: %w", name.Literal, err)
	}

	// Parse parameters
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		paramName, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, fmt.Errorf("function %s params: %w", name.Literal, err)
		}

		_, err = p.expect(lexer.TokenColon)
		if err != nil {
			return nil, fmt.Errorf("function %s param %s: %w", name.Literal, paramName.Literal, err)
		}

		paramType, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, fmt.Errorf("function %s param %s type: %w", name.Literal, paramName.Literal, err)
		}

		param := &ASTNode{
			Type:  NodeVarDecl,
			Name:  paramName.Literal,
			Value: paramType.Literal,
		}
		node.Params = append(node.Params, param)

		if p.current.Type == lexer.TokenComma {
			p.advance()
		}
	}

	_, err = p.expect(lexer.TokenRParen)
	if err != nil {
		return nil, fmt.Errorf("function %s: %w", name.Literal, err)
	}

	// Parse body block
	body, err := p.parseBlock()
	if err != nil {
		return nil, fmt.Errorf("function %s body: %w", name.Literal, err)
	}
	node.Body = body

	return node, nil
}

// parseBlock parses a block of statements: { stmt1 stmt2 ... }
func (p *Parser) parseBlock() ([]*ASTNode, error) {
	_, err := p.expect(lexer.TokenLBrace)
	if err != nil {
		return nil, err
	}

	var statements []*ASTNode
	for p.current.Type != lexer.TokenRBrace && p.current.Type != lexer.TokenEOF {
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	_, err = p.expect(lexer.TokenRBrace)
	if err != nil {
		return nil, err
	}

	return statements, nil
}

// parseStatement parses a single statement inside a block.
func (p *Parser) parseStatement() (*ASTNode, error) {
	switch p.current.Type {
	case lexer.TokenIdentifier:
		return p.parseAssignmentOrCall()
	case lexer.TokenPrint:
		return p.parsePrintCall()
	case lexer.TokenIf:
		return p.parseIfStmt()
	case lexer.TokenFn:
		return p.parseFnDecl()
	default:
		// Skip unknown statements
		p.advance()
		return nil, nil
	}
}

// parseAssignmentOrCall parses either a variable declaration (x := expr)
// or a function call (fn(args)).
func (p *Parser) parseAssignmentOrCall() (*ASTNode, error) {
	// Check if this is a call: ident.ident(...)
	if p.peek().Type == lexer.TokenDot {
		return p.parseCall()
	}

	// Check if this is a call: ident(...)
	if p.peek().Type == lexer.TokenLParen {
		return p.parseCall()
	}

	// Variable declaration: ident := expression
	ident, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, err
	}

	// Check for := (short variable declaration)
	if p.current.Type == lexer.TokenColonEquals {
		p.advance() // consume :=

		// Parse the value expression
		value, err := p.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("variable %s: %w", ident.Literal, err)
		}

		return &ASTNode{
			Type:     NodeVarDecl,
			Name:     ident.Literal,
			Children: []*ASTNode{value},
		}, nil
	}

	// If followed by dot, it's a field access: ident.field or ident.field = expr
	if p.current.Type == lexer.TokenDot {
		left, err := p.parseFieldAccess(&ASTNode{Type: NodeIdent, Name: ident.Literal})
		if err != nil {
			return nil, err
		}
		// Check for assignment: ident.field = expr
		if p.current.Type == lexer.TokenEquals {
			p.advance() // consume '='
			value, err := p.parseExpression()
			if err != nil {
				return nil, fmt.Errorf("assignment: %w", err)
			}
			// Create var decl node for the assignment
			return &ASTNode{
				Type:     NodeVarDecl,
				Name:     "assign",
				Children: []*ASTNode{left, value},
			}, nil
		}
		return left, nil
	}

	// Standalone identifier expression
	return &ASTNode{Type: NodeIdent, Name: ident.Literal}, nil
}

// parseCall parses a function or method call: expr(args) or expr.method(args)
func (p *Parser) parseCall() (*ASTNode, error) {
	// Parse the base expression (identifier or field access chain)
	expr, err := p.parseExpression()
	if err != nil {
		return nil, err
	}

	// Expect opening paren for call
	if p.current.Type != lexer.TokenLParen {
		return expr, nil // Not a call, return the expression as-is
	}

	p.advance() // consume '('

	callNode := &ASTNode{Type: NodeCallExpr}
	callNode.Children = append(callNode.Children, expr)

	// Parse arguments
	for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
		arg, err := p.parseExpression()
		if err != nil {
			return nil, fmt.Errorf("call argument: %w", err)
		}
		callNode.Children = append(callNode.Children, arg)

		if p.current.Type == lexer.TokenComma {
			p.advance()
		}
	}

	_, err = p.expect(lexer.TokenRParen)
	if err != nil {
		return nil, fmt.Errorf("call closing paren: %w", err)
	}

	return callNode, nil
}

// parseExpression parses a general expression (field access, identifier, number, call).
func (p *Parser) parseExpression() (*ASTNode, error) {
	switch p.current.Type {
	case lexer.TokenIdentifier:
		ident := &ASTNode{Type: NodeIdent, Name: p.current.Literal}
		p.advance()

		// Check for field access: ident.ident
		if p.current.Type == lexer.TokenDot {
			return p.parseFieldAccess(ident)
		}

		// Check for call: ident(...)
		if p.current.Type == lexer.TokenLParen {
			// We need to call this as a function — backtrack slightly
			p.pos-- // go back to before the identifier
			p.current = p.tokens[p.pos]
			return p.parseCall()
		}

		return ident, nil

	case lexer.TokenNumber:
		node := &ASTNode{Type: NodeNumber, Value: p.current.Literal}
		p.advance()
		return node, nil

	default:
		return nil, fmt.Errorf("unexpected token %s in expression at line %d",
			p.current.Type, p.current.Line)
	}
}

// parseFieldAccess parses field/method access: expr.field or expr.field(...)
func (p *Parser) parseFieldAccess(left *ASTNode) (*ASTNode, error) {
	for p.current.Type == lexer.TokenDot {
		p.advance() // consume '.'

		fieldName, err := p.expect(lexer.TokenIdentifier)
		if err != nil {
			return nil, fmt.Errorf("field access: %w", err)
		}

		accessNode := &ASTNode{
			Type: NodeFieldAccess,
			Name: fieldName.Literal,
			Children: []*ASTNode{
				left,
			},
		}
		left = accessNode

		// Check if this is a method call: expr.field(...)
		if p.current.Type == lexer.TokenLParen {
			callNode := &ASTNode{Type: NodeCallExpr}
			callNode.Children = append(callNode.Children, left)

			p.advance() // consume '('

			for p.current.Type != lexer.TokenRParen && p.current.Type != lexer.TokenEOF {
				arg, err := p.parseExpression()
				if err != nil {
					return nil, fmt.Errorf("method call arg: %w", err)
				}
				callNode.Children = append(callNode.Children, arg)

				if p.current.Type == lexer.TokenComma {
					p.advance()
				}
			}

			_, err = p.expect(lexer.TokenRParen)
			if err != nil {
				return nil, fmt.Errorf("method call: %w", err)
			}

			return callNode, nil
		}
	}

	return left, nil
}

// parseIfStmt parses an if statement: if condition { body }
func (p *Parser) parseIfStmt() (*ASTNode, error) {
	p.advance() // consume 'if'

	node := &ASTNode{Type: NodeBlock, Name: "if"}

	// Parse condition (simple identifier for now)
	cond, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("if condition: %w", err)
	}
	node.Children = append(node.Children, cond)

	// Parse body
	body, err := p.parseBlock()
	if err != nil {
		return nil, fmt.Errorf("if body: %w", err)
	}
	node.Body = body

	return node, nil
}

// parsePrintCall parses a print statement: print(expression)
func (p *Parser) parsePrintCall() (*ASTNode, error) {
	p.advance() // consume 'print'

	_, err := p.expect(lexer.TokenLParen)
	if err != nil {
		return nil, fmt.Errorf("print: %w", err)
	}

	// Parse the argument expression
	arg, err := p.parseExpression()
	if err != nil {
		return nil, fmt.Errorf("print argument: %w", err)
	}

	_, err = p.expect(lexer.TokenRParen)
	if err != nil {
		return nil, fmt.Errorf("print closing paren: %w", err)
	}

	return &ASTNode{
		Type: NodeCallExpr,
		Name: "print",
		Children: []*ASTNode{
			{Type: NodeIdent, Name: "print"},
			arg,
		},
	}, nil
}

// parseImport parses an import statement: import "path"
func (p *Parser) parseImport() (*ASTNode, error) {
	p.advance() // consume 'import'

	path, err := p.expect(lexer.TokenIdentifier)
	if err != nil {
		return nil, fmt.Errorf("import: %w", err)
	}

	return &ASTNode{Type: NodeBlock, Name: "import", Value: path.Literal}, nil
}

// PrintAST prints the AST in a readable tree format.
func PrintAST(node *ASTNode, indent int) {
	if node == nil {
		return
	}

	prefix := strings.Repeat("  ", indent)
	typeName := nodeTypeName(node.Type)

	switch node.Type {
	case NodeProgram:
		fmt.Printf("%s%s\n", prefix, typeName)
		for _, child := range node.Children {
			PrintAST(child, indent+1)
		}

	case NodeStructDecl:
		fmt.Printf("%s%s: %s\n", prefix, typeName, node.Name)
		if len(node.Fields) > 0 {
			fmt.Printf("%s  Fields:\n", prefix)
			for _, field := range node.Fields {
				PrintAST(field, indent+2)
			}
		}
		if len(node.Children) > 0 {
			fmt.Printf("%s  Methods:\n", prefix)
			for _, child := range node.Children {
				PrintAST(child, indent+2)
			}
		}

	case NodeFnDecl:
		fmt.Printf("%s%s: %s\n", prefix, typeName, node.Name)
		if len(node.Params) > 0 {
			fmt.Printf("%s  Params:\n", prefix)
			for _, param := range node.Params {
				PrintAST(param, indent+2)
			}
		}
		if len(node.Body) > 0 {
			fmt.Printf("%s  Body:\n", prefix)
			for _, stmt := range node.Body {
				PrintAST(stmt, indent+2)
			}
		}

	case NodeVarDecl:
		if node.Value != "" {
			fmt.Printf("%s%s: %s : %s\n", prefix, typeName, node.Name, node.Value)
		} else if len(node.Children) > 0 {
			fmt.Printf("%s%s: %s =\n", prefix, typeName, node.Name)
			for _, child := range node.Children {
				PrintAST(child, indent+2)
			}
		} else {
			fmt.Printf("%s%s: %s\n", prefix, typeName, node.Name)
		}

	case NodeCallExpr:
		fmt.Printf("%s%s\n", prefix, typeName)
		for _, child := range node.Children {
			PrintAST(child, indent+1)
		}

	case NodeFieldAccess:
		fmt.Printf("%s%s: .%s\n", prefix, typeName, node.Name)
		for _, child := range node.Children {
			PrintAST(child, indent+1)
		}

	case NodeIdent:
		fmt.Printf("%s%s: %s\n", prefix, typeName, node.Name)

	case NodeNumber:
		fmt.Printf("%s%s: %s\n", prefix, typeName, node.Value)

	case NodeBlock:
		fmt.Printf("%s%s: %s\n", prefix, typeName, node.Name)
		if len(node.Children) > 0 {
			for _, child := range node.Children {
				PrintAST(child, indent+1)
			}
		}
		if len(node.Body) > 0 {
			fmt.Printf("%s  Body:\n", prefix)
			for _, stmt := range node.Body {
				PrintAST(stmt, indent+2)
			}
		}

	default:
		fmt.Printf("%s%s: %s\n", prefix, typeName, node.Name)
	}
}

func nodeTypeName(t NodeType) string {
	switch t {
	case NodeProgram:
		return "Program"
	case NodeStructDecl:
		return "StructDecl"
	case NodeFnDecl:
		return "FnDecl"
	case NodeBlock:
		return "Block"
	case NodeCallExpr:
		return "CallExpr"
	case NodeFieldAccess:
		return "FieldAccess"
	case NodeVarDecl:
		return "VarDecl"
	case NodeIdent:
		return "Ident"
	case NodeNumber:
		return "Number"
	default:
		return "Unknown"
	}
}
