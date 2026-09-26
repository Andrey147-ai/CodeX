package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
)

// ========== AUTO-API (codex serve) ==========
// Every top-level function becomes an HTTP endpoint:
//   /add?a=2&b=3  ->  add(2, 3). Query values convert to numbers
// when numeric, else stay strings. GET / lists function names.

func codexServe(sourceFile string, port int) {
	sourceBytes, err := os.ReadFile(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Cannot read %s: %v\n", sourceFile, err)
		os.Exit(1)
	}
	interp := NewInterpreter()
	activeInterp = interp
	abs, err := filepath.Abs(sourceFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Bad path %s: %v\n", sourceFile, err)
		os.Exit(1)
	}
	interp.mainDir = filepath.Dir(abs)
	prog := NewParser(NewLexer(string(sourceBytes)).Tokenize()).ParseProgram()
	interp.eval(prog, interp.globalEnv)

	names := make([]string, 0, len(interp.globalEnv.funcs))
	for name := range interp.globalEnv.funcs {
		names = append(names, name)
	}
	sort.Strings(names)
	if len(names) == 0 {
		fmt.Fprintf(os.Stderr, "serve: no top-level functions in %s\n", sourceFile)
		os.Exit(1)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		name := strings.TrimPrefix(r.URL.Path, "/")
		if name == "" {
			quoted := make([]string, 0, len(names))
			for _, n := range names {
				quoted = append(quoted, strconv.Quote(n))
			}
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte("{\"functions\":[" + strings.Join(quoted, ",") + "]}"))
			return
		}
		fnDef, ok := interp.globalEnv.funcs[name]
		if !ok {
			http.Error(w, "no function "+name, http.StatusNotFound)
			return
		}
		q := r.URL.Query()
		args := make([]Value, 0, len(fnDef.Params))
		for _, param := range fnDef.Params {
			raw, present := q[param]
			if !present || len(raw) == 0 {
				http.Error(w, "missing param "+param, http.StatusBadRequest)
				return
			}
			if num, err := strconv.ParseFloat(strings.TrimSpace(raw[0]), 64); err == nil {
				args = append(args, Value{Kind: "number", NumVal: num})
			} else {
				args = append(args, Value{Kind: "string", StrVal: raw[0]})
			}
		}
		var res Value
		var crashed *runtimeError
		func() {
			defer func() {
				if rec := recover(); rec != nil {
					if re, ok := rec.(*runtimeError); ok {
						crashed = re
						return
					}
					panic(rec)
				}
			}()
			res = interp.invokeUserFunc(fnDef, interp.globalEnv, args, "serve:"+name)
		}()
		if crashed != nil {
			http.Error(w, crashed.msg, http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte(valueToString(res)))
	})
	fmt.Printf("serving %d function(s) on 127.0.0.1:%d\n", len(names), port)
	if err := http.ListenAndServe(fmt.Sprintf("127.0.0.1:%d", port), mux); err != nil {
		fmt.Fprintf(os.Stderr, "serve failed: %v\n", err)
		os.Exit(1)
	}
}
