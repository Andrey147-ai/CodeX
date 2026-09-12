# CodeX

A small programming language written from scratch in Go. One file of
interpreter, zero dependencies — grab `codex.exe` from
[Releases](../../releases) and run `.cx` scripts anywhere.

> I'm Andrey, 16, from Kazakhstan. I started CodeX to figure out how
> programming languages actually work inside — lexer, parser, runtime,
> all of it. It's still growing, but it already runs real programs:
> games, log analyzers, HTTP scripts.

## What it looks like

```go
struct Boss {
    id
    hp
    shield
}

b := Boss{777, 1000, 300}
print("Boss spawned! Shields: ", b.shield)

damage := 500
b.hp = b.hp + b.shield - damage
b.shield = 0

if b.hp > 0 {
    print("Boss survives with HP: ", b.hp)
}
```

```
Boss spawned! Shields: 300
Boss survives with HP: 800
```

New here? Start with [Tutorial Part 1: basics](docs/tutorial-01-basics.md) —
it takes about 15 minutes.

## 📖 Learning

Tutorials live in `docs/` and come out one by one, each with runnable
code and exercises:

* **Part 1: basics** — first script, `print`, variables, math,
  conditions, loops ([read it](docs/tutorial-01-basics.md))
* **Part 2: data** — arrays, maps, functions, structs, grade book
  ([read it](docs/tutorial-02-data.md))
* **Part 3: strings** — length, case, split/join, escapes
  ([read it](docs/tutorial-03-strings.md))
* **Part 4: files** — read, write, logs ([read it](docs/tutorial-04-files.md))
* **Part 5: functions** — params, return, scope, methods
  ([read it](docs/tutorial-05-functions.md))
* **Part 6: loops** — nesting, logic, infinite loops
  ([read it](docs/tutorial-06-loops.md))
* **Part 7: maps** — modeling, nesting, counters
  ([read it](docs/tutorial-07-maps.md))
* **Part 8: errors** — reading messages, debugging
  ([read it](docs/tutorial-08-errors.md))
* **Part 9: packages** — GitHub imports, publishing
  ([read it](docs/tutorial-09-packages.md))
* **Part 10: capstone** — dissecting the log analyzer
  ([read it](docs/tutorial-10-capstone.md))

Each snippet in the tutorials is executed against the real interpreter
before publishing, so what you read is what runs.

## How it works

Three stages, all in `main.go`:

1. **Lexer** — turns source text into tokens (numbers, strings, operators,
   keywords). Every token knows its line and column, so errors point at
   the exact spot: `Lexer error at 2:6`.
2. **Parser** — builds a syntax tree (AST) with correct operator
   precedence (`2 + 3 * 4` is `14`). It pre-scans struct names so a block
   like `if hp > 0 {` is never confused with a struct literal.
3. **Interpreter** — walks the tree with chained scopes (global →
   function → block → loop iteration). No garbage collector: when a
   scope ends, its structs are freed and logged (`[del]`). `return`,
   `break` and `continue` travel up through panic signals caught at the
   right level.

## Language tour

```go
// variables: := creates, = updates
x := 5
x = x + 1

// types: numbers, strings, bools, arrays, maps, structs, nil
a := [10, 20, 30]
m := {"name": "CodeX", "ver": 11}

// branches and loops
if x > 10 {
    print("big")
} else if x > 5 {
    print("mid")
} else {
    print("small")
}

for i := 0; i < 3; i = i + 1 {
    print(i)
}
for item in a {
    print(item)
}
while x > 0 {
    x = x - 1
}

// functions and struct methods (methods mutate the receiver)
fn add(a, b) {
    return a + b
}
struct Player { name hp }
fn (p Player) heal(x) {
    p.hp = p.hp + x
    return p.hp
}
p := Player{"hero", 50}
p.heal(30)

// functions are values: literals, higher-order calls, closures
double := fn(x) {
    return x * 2
}
fn makeAdder(n) {
    return fn(x) {
        return x + n
    }
}
add5 := makeAdder(5)
print(add5(100))   // 105

// packages from GitHub, no registry needed
import "./mylib.cx"
import "Andrey147-ai/strutils@v1.2.0"
```

## 📚 Standard Library

| Function | Description | Example |
|---|---|---|
| `print(...)` | Print values | `print("hp=", hp)` |
| `len(x)` | Array / map / string length | `len([1,2])` → `2` |
| `push(arr, v)` | Append to array | `push(inv, "sword")` |
| `sort(arr)` | Sort numbers/strings in place | `sort(scores)` |
| `keys(m)` | Sorted map keys | `keys({"a":1})` → `[a]` |
| `has(m, k)` | Map key check | `has(cfg, "debug")` |
| `str(x)` | Any value to string | `str(42)` → `"42"` |
| `num(x)` | String/bool/number to number | `num("19")+3` → `22` |
| `input(p)` | Read a line from stdin | `name := input("name? ")` |
| `args()` | CLI args after the script | `args()[0]` |
| `upper/lower(s)` | String case | `upper("hi")` → `"HI"` |
| `contains(s, sub)` | Substring check | `contains(s, "err")` |
| `split(s, sep)` | Split to array | `split("a,b", ",")` |
| `join(arr, sep)` | Join array to string | `join(["a"], "-")` |
| `read_file(p)` | Read whole file | `read_file("app.log")` |
| `write_file(p, t)` | Overwrite file, returns bytes | `write_file("o.txt", t)` |
| `append_file(p, t)` | Append to file | `append_file("o.txt", t)` |
| `exists(p)` | Path check | `exists("o.txt")` |
| `http_get(url)` | Fetch URL body | `http_get("https://example.com")` |
| `http_listen(port, fn)` | Serve HTTP, handler gets map | `http_listen(8080, hello)` |
| `sleep(ms)` | Millisecond pause | `sleep(50)` |
| `pkgdir(spec)` | Local path of a GitHub repo | `pkgdir("user/data")` |
| `type(x)` | Runtime type name | `type([1])` → `"array"` |
| `env(name)` | Env var or `nil` | `env("PATH")` |
| `exit(code)` | Exit with code | `exit(3)` |
| `now()` | Unix timestamp | `now()` |
| `parse_json(s)` | JSON string to value | `parse_json("{\"a\":1}")` |
| `to_json(x)` | Value to JSON string | `to_json([1,"a"])` |
| `try/catch` | Catch runtime errors | `try { risky() } catch e { print(e) }` || `a div b`, `a % b` | Integer division, modulo | `7 div 2` → `3`, `7 % 3` → `1` |
| `x[i:j]` | Slices (strings/arrays, negatives ok) | `"hello"[1:4]` → `"ell"` |
| `del(x)` | Manual scope cleanup | `del(b)` |

## 📦 Packages (GitHub)

No registry, no editor lock-in — packages come straight from GitHub repos.

```go
import "./mylib.cx"                            // local file
import "Andrey147-ai/strutils"                 // latest main, entry main.cx
import "Andrey147-ai/strutils@v1.2.0"          // pinned tag or branch
import "Andrey147-ai/strutils/lib/text.cx"     // explicit file in repo
```

Rules:
* A package is any public GitHub repo. Entry file: `main.cx`, else `<repo>.cx`, else `lib.cx`.
* Versions are exact tags/branches after `@`, otherwise the default branch.
* Downloaded zips are cached in `~/.codex/pkgs`, an imported file runs once.
* Prefetch without running: `codex get user/repo@ver`.
* Need a repo's files as data (not code)? `pkgdir("user/repo")` returns
  its cache path — see `examples/badapple.cx` (6572 ASCII frames
  streamed from a frames repo with `sleep()` + ANSI `\e` control).
* To publish a package, push a repo with `main.cx` — done.

## Examples

In `examples/`: `raid.cx` and `shop_game.cx` (games), `inventory.cx`,
`logstat.cx` (log analyzer, try `..\codex.exe logstat.cx app.log`),
`methods.cx`, `stdlib.cx`, `showcase.cx`, `useimport.cx`, `badapple.cx`,
`server.cx` (web server, try `..\codex.exe server.cx`).

## Building from source

You need Go installed:

```
go build -ldflags="-s -w" -o codex.exe .
```

Tests and style: `go vet ./...` must pass, and every push is checked by
GitHub Actions on Ubuntu + Windows.

## 🗺️ Roadmap

Done so far — comparisons and logic, correct precedence, `while` /
`for` / `for-in` with `break` / `continue`, arrays, dictionaries,
struct methods, files, HTTP, packages, terminal control, `%` / `div`,
slices, first-class functions and closures, `http_listen` web server. Up next:

* [ ] Built-in lightweight networking library for backends (`http_listen`)
* [ ] Your idea — open an issue
