package main

// Strength tests: the interpreter must not fall over on user input.
// Unit level (lexer/formatter) + CLI level (fmt/test over every .cx file).

import (
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func lexTypes(src string) []TokenType {
	toks := NewLexer(src).Tokenize()
	out := make([]TokenType, 0, len(toks))
	for _, t := range toks {
		if t.Type == TOK_NEWLINE {
			continue
		}
		out = append(out, t.Type)
	}
	return out
}

func sameTypes(a, b []TokenType) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func TestLexerBasics(t *testing.T) {
	got := lexTypes("x := 5")
	want := []TokenType{TOK_IDENT, TOK_ASSIGN, TOK_NUMBER, TOK_EOF}
	if !sameTypes(got, want) {
		t.Fatalf("x := 5 -> %v, want %v", got, want)
	}
}

func TestLexerCommentAndString(t *testing.T) {
	toks := NewLexer("print(\"a\\n\") // hi").Tokenize()
	var kinds []TokenType
	for _, tok := range toks {
		kinds = append(kinds, tok.Type)
	}
	want := []TokenType{TOK_PRINT, TOK_LPAREN, TOK_STRING, TOK_RPAREN, TOK_COMMENT, TOK_EOF}
	if !sameTypes(kinds, want) {
		t.Fatalf("kinds = %v, want %v", kinds, want)
	}
	if toks[2].Value != "a\n" {
		t.Fatalf("escape \\n = %q, want real newline", toks[2].Value)
	}
	if toks[4].Value != " hi" {
		t.Fatalf("comment = %q", toks[4].Value)
	}
}

func TestLexerBOMSkipped(t *testing.T) {
	a := lexTypes("x := 1")
	b := lexTypes("\uFEFFx := 1")
	if !sameTypes(a, b) {
		t.Fatalf("BOM changed tokens: %v vs %v", a, b)
	}
}

func TestLexerCompoundAssignTokens(t *testing.T) {
	// += is two tokens (PLUS, EQ) — the parser glues them.
	got := lexTypes("x += 1")
	want := []TokenType{TOK_IDENT, TOK_PLUS, TOK_EQ, TOK_NUMBER, TOK_EOF}
	if !sameTypes(got, want) {
		t.Fatalf("x += 1 -> %v, want %v", got, want)
	}
}

func TestFormatIdempotent(t *testing.T) {
	samples := []string{
		"x := 1\n",
		"if x > 0 {\n    print(x)\n}\n",
		"for item in [1, 2] {\n    print(item)\n}\n",
		"x += 1\n",
		"m := {\"a\": 1}\n",
		"s := \"hi\"[1:2]\n",
	}
	for _, s := range samples {
		once := formatSource(s)
		twice := formatSource(once)
		if once != twice {
			t.Fatalf("formatter not idempotent for %q:\n%s\n---\n%s", s, once, twice)
		}
	}
}

// codexBin finds the CLI binary built by CI (./codex.exe or ./codex).
func codexBin(t *testing.T) string {
	t.Helper()
	name := "codex"
	if runtime.GOOS == "windows" {
		name = "codex.exe"
	}
	if _, err := os.Stat(name); err != nil {
		t.Skipf("no ./%s binary (build first), skipping CLI tests", name)
	}
	abs, err := filepath.Abs(name)
	if err != nil {
		t.Fatal(err)
	}
	return abs
}

func collectCX(t *testing.T, dirs ...string) []string {
	t.Helper()
	var files []string
	for _, dir := range dirs {
		entries, err := os.ReadDir(dir)
		if err != nil {
			t.Fatal(err)
		}
		for _, e := range entries {
			if !e.IsDir() && strings.HasSuffix(e.Name(), ".cx") {
				files = append(files, filepath.Join(dir, e.Name()))
			}
		}
	}
	if len(files) == 0 {
		t.Fatal("no .cx files found")
	}
	return files
}

func TestFmtCheckAllFiles(t *testing.T) {
	bin := codexBin(t)
	for _, f := range collectCX(t, "examples", "tests", "packages/strutils", "packages/mathx", "packages/apix", "packages/csvx", "site") {
		cmd := exec.Command(bin, "fmt", "--check", f)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Errorf("fmt --check %s failed: %v\n%s", f, err, out)
		}
	}
}

func TestSelfTests(t *testing.T) {
	bin := codexBin(t)
	cmd := exec.Command(bin, "test", "tests")
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("codex test tests failed: %v\n%s", err, out)
	}
}
