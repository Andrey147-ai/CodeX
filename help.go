package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// ========== HELP + NEW (beginner DX) ==========

func printHelp() {
	fmt.Println("CodeX " + codexVersion + " — a lightweight programming language")
	fmt.Println("")
	fmt.Println("Usage:")
	fmt.Println("  codex.exe <file.cx>        run a script")
	fmt.Println("  codex.exe                  REPL (interactive mode)")
	fmt.Println("  codex.exe help             this help")
	fmt.Println("  codex.exe version          version")
	fmt.Println("  codex.exe fmt <file.cx>    format code")
	fmt.Println("  codex.exe fmt --check <file.cx>  check formatting (for CI)")
	fmt.Println("  codex.exe test [dir]       *_test.cx tests with assert()")
	fmt.Println("  codex.exe get <user/repo[@ver]>  fetch a package from GitHub")
	fmt.Println("  codex.exe list               package catalog")
	fmt.Println("  codex.exe search <query>     search the catalog")
	fmt.Println("  codex.exe info <package>    package details")
	fmt.Println("  codex.exe serve <file.cx> [port]  auto-API from functions")
	fmt.Println("  codex.exe new [file.cx]    create a beginner template (default main.cx)")
	fmt.Println("")
	fmt.Println("Examples:")
	fmt.Println("  codex.exe new hello.cx")
	fmt.Println("  codex.exe hello.cx")
	fmt.Println("  codex.exe test tests")
	fmt.Println("")
	fmt.Println("Tutorials: docs/tutorial-01-basics.md ... docs/tutorial-11-backend.md")
}

func codexNew(target string) error {
	if !strings.HasSuffix(strings.ToLower(target), ".cx") {
		target += ".cx"
	}
	if _, err := os.Stat(target); err == nil {
		return fmt.Errorf("file %s already exists", target)
	}
	template := `// Hi! This is CodeX — run: codex.exe ` + filepath.Base(target) + `
name := input("Your name? ")
print("Hi, " + name + "!")

// Array + loop
scores := [90, 80, 100]
total := 0
for s in scores {
    total += s
}
print("Average: ", total / len(scores))

// Function
fn double(n) {
    return n * 2
}
print("double(21) = ", double(21))
`
	if err := os.WriteFile(target, []byte(template), 0644); err != nil {
		return err
	}
	fmt.Printf("ok %s — run: codex.exe %s\n", target, target)
	return nil
}
