# 🚀 CodeX Programming Language

**CodeX** is a lightweight, fast, and completely standalone general-purpose programming language built from scratch in **Go**.

> ⚡ **Fun Fact:** This project is entirely developed and maintained by a 16-year-old independent developer from Kazakhstan (`Andrey147-ai`). The goal of CodeX is to prove that building custom system architecture, lexers, and parsers requires focus and passion rather than degrees or age.

The core architectural highlight of CodeX is its **hybrid memory management that bypasses heavy Garbage Collection (GC)**. The language combines structural simplicity with an automated scope-based memory cleanup (Scope-based Memory Management) to minimize system overhead.

---

## 🔥 Key Features

* 📦 **Zero Dependencies:** The compiler bundles the entire runtime into a single executable (`codex.exe`). Users don't need to install Go, C, or Python—it works out of the box.
* 🧠 **Smart Scope-Based Despawn:** When variables leave their scope (functions, `if` statements), the CodeX runtime automatically wipes the local structures from memory and prints cleanup logs. Manual control is also available via the `del()` construct.
* ⚡ **Struct Field Math:** Full support for custom objects via the `struct` keyword, with the unique ability to mutate and evaluate object fields directly inside complex mathematical expressions.

---

## 🛠️ Architecture Under the Hood

The CodeX interpreter follows classical systems programming principles:
1. **Lexer (Tokenizer):** Scans raw `.cx` source code and converts it into a stream of tokens (identifiers, numbers, operators).
2. **Parser:** Builds an Abstract Syntax Tree (**AST**) based on the language grammar rules.
3. **Interpreter & Environment:** Evaluates AST nodes, isolates variable scopes using custom layered `Environment` states, and tracks the lifecycle of allocations.

---

## 💻 Code Example (`main.cx`)

Here is how a raid boss simulation looks in CodeX, featuring dynamic damage calculations and automatic object memory freeing once the logic scope ends:

```go
struct Boss {
    id
    hp
    shield
}

print("=== SYSTEM LOG: Raid Initialization ===")
print("-> Player entered the boss room")

// Spawning the struct
b := Boss{777, 1000, 300}
print("-> Boss spawned! Current shields: ", b.shield)

damage := 500
print("-> CRITICAL HIT! Dealing ", damage, " damage")

// Complex math using struct fields directly out of the box!
b.hp = b.hp + b.shield - damage
b.shield = 0

if b.hp {
    print("-> Shields destroyed! Remaining boss HP: ", b.hp)
}

print("-> Player won and is leaving the room...")
// Upon exiting the scope, the CodeX runtime automatically triggers destructors for object 'b'
```
## Console Output on Launch:
```
=== SYSTEM LOG: Raid Initialization ===
-> Player entered the boss room
-> Boss spawned! Current shields: 300
-> CRITICAL HIT! Dealing 500 damage
-> Shields destroyed! Remaining boss HP: 800
-> Player won and is leaving the room...
[del] b.id = 777 (freed)
[del] b.hp = 800 (freed)
[del] b.shield = 0 (freed)
=== SYSTEM LOG: Raid completed, server memory is clean ===
```
## 🚀 Getting Started
For Users (Running Scripts)
Go to the Releases tab on this GitHub repository, download the standalone codex.exe, drop it into your working directory, and run your script via terminal:
```
.\codex.exe main.cx
```
## For Developers (Building from Source)
If you want to compile the interpreter engine yourself, make sure you have Go installed:
```
# Initialize the module
go mod init codex

# Build an optimized executable without debug bloat
go build -ldflags="-s -w" -o codex.exe main.go
```
## 🗺️ Roadmap
CodeX is actively evolving with a focus on future full-stack and game development. Upcoming milestones:

[ ] Implementing loops (for, while) for running proper 2D game loops.

* [ ] Dynamic arrays support ([]Value).

* [ ] Struct methods for complete OOP behavior (e.g., fn (b Boss) attack()).

* [ ] Built-in lightweight networking library for backend routing (http_listen).

## 🔧 Repository Maintenance Commands
**How to clean compiled binaries from local tracking:**
To keep the source tree lightweight, avoid pushing compiled .exe files directly into the repository. Use .gitignore or clean tracked files using the following terminal command:
```
git rm --cached codex.exe
git commit -m "Build: Remove compiled binary from source control"
git push
```
## How to release a new standalone binary version:
1. Navigate to the Releases section on the right side of this GitHub page.
2. Click Create a new release (or Draft a new release).
3. Set the version tag (e.g., v1.0.0) and give your release a title.
4. Drag and drop your compiled codex.exe into the binary attachment box.
5. Click Publish release to deliver a single-file executable to the end-users.
