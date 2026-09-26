package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

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
