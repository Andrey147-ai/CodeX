package main

import (
	"strings"
)

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
