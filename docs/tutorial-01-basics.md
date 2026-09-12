# CodeX Tutorial, Part 1: Basics

*First script, variables, math, conditions, loops. ~15–20 minutes.*

## 0. Setup and first run

1. Open [Releases](../..//releases) and download `codex.exe`.
2. Drop it into an empty folder, e.g. `C:\codex`.
3. Create `hello.cx` next to it with a single line:

```
print("hello, CodeX")
```

4. Open a terminal in that folder and run:

```
codex.exe hello.cx
```

You should see `hello, CodeX`.

**If you get `Ошибка чтения hello.cx: ... cannot find the path`** — don't
panic, that's not a language bug. It's the interpreter saying "file not
found". Usually one of three things: you're in the wrong folder (check
with `cd`), a typo in the name (`hello.cx`, not `helo.cx`), or the file
was saved with a hidden extension (Notepad loves making `hello.cx.txt` —
turn on file extensions in Explorer).

## 1. Printing: `print`

`print` takes as many values as you want and glues them together
**with no spaces**:

```
print("2 + 3 =", 2 + 3)
```

That outputs `2 + 3 =5` — the five sticks to the text. Add spaces yourself:

```
print("2 + 3 = ", 2 + 3)   // 2 + 3 = 5
```

Comments start with `//` and run to the end of the line:

```
// the interpreter skips me
print("but runs me")   // trailing comment works too
```

## 2. Variables

A variable is created with `:=` **once**, then updated with `=`:

```
x := 5       // created
print(x)     // 5
x = x + 1    // updated
print(x)     // 6
```

**Classic beginner mistake:** writing `=` instead of `==` in a comparison.
Single `=` is always assignment; comparison is double:

```
y := 10
print(y == 10)   // true — comparison
print(y = 10)    // ERROR: can't assign here
```

There are only a few types, and they're honest:

| Type | Examples |
|---|---|
| numbers | `42`, `3.5`, `-7` |
| strings | `"hi"`, `"42"` |
| booleans | `true`, `false` |

A number and a string are different things — convert explicitly:

```
print(num("19") + 3)   // 22 — string became a number
print(str(42) + "!")   // 42! — number became a string
```

## 3. Math like in school

Multiplication before addition, parentheses beat everything:

```
print(2 + 3 * 4)     // 14, not 20
print((2 + 3) * 4)   // 20
print(10 - 4 / 2)    // 8
```

Comparisons (`== != < > <= >=`) return `true` / `false`, and `&&` (and),
`||` (or), `!` (not) combine them:

```
age := 16
print(age >= 13 && age < 20)   // true: both hold
print(age == 10 || age == 16)  // true: second one holds
print(!(age == 10))            // true: "not (age is 10)"
```

## 4. Branching: `if`

The classic. Note: no parentheses around the condition, block in `{ }`:

```
score := 75
if score > 90 {
    print("S")
} else if score > 70 {
    print("A")
} else {
    print("try again")
}
// prints A
```

Almost any value works as a condition. False is `false`, `0` and the
empty string `""`. Everything else is true:

```
if "hello" {
    print("non-empty string — we get here")
}
if 0 {
    print("never reached")
}
```

## 5. Loops

Three flavors. `for` when you know how many times. `for-in` to walk
values. `while` to repeat while something holds:

```
for i := 0; i < 3; i = i + 1 {
    print(i)          // 0 1 2
}

for w in ["a", "b", "c"] {
    if w == "b" {
        continue      // skip "b"
    }
    print(w)          // a c
}

n := 3
while n > 0 {
    print(n)          // 3 2 1
    n = n - 1         // without this line it spins forever!
}
```

`break` exits a loop early. `continue` drops the iteration and jumps to
the next one. Stuck in an infinite loop — press `Ctrl+C`.

## 6. Test yourself

**Task 1.** Save as `task1.cx`, predict the output, then run it:

```
x := 2
total := 0
while x <= 10 {
    if x == 5 {
        x = x + 1
        continue
    }
    total = total + x
    x = x + 1
}
print(total)
```

<details>
<summary>Answer</summary>

`49`. We add 2..10 but skip five via `continue`:
2+3+4+6+7+8+9+10 = 49.

</details>

**Task 2.** Print numbers 1 to 20, but `tick` instead of multiples of three
and `tock` instead of multiples of five (hint: the language has no modulo
yet — keep separate counters and reset them, work with what you have).

---

*Part 2 — [arrays, maps, functions and structs](tutorial-02-data.md):
we'll write a grade book.*
