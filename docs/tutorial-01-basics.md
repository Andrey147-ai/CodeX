# CodeX Tutorial, Part 1: Basics

*Run your first script, variables, math, conditions, loops. ~15 minutes.*

## 0. Setup (2 minutes)

1. Open [Releases](../..//releases) and download `codex.exe`.
2. Put it in an empty folder, e.g. `C:\codex`.
3. Create `hello.cx` next to it:

```
print("hello, CodeX")
```

4. Run it:

```
codex.exe hello.cx
```

You should see `hello, CodeX`. If the file is missing, you'll get
`Ошибка чтения` — check you're in the right folder.

## 1. Printing things

`print` takes as many values as you want and glues them together:

```
print("2 + 3 =", 2 + 3)
print("pi is about ", 3.14)
```

Comments start with `//` and run to the end of the line:

```
// this line does nothing
print("this one does")
```

## 2. Variables

`:=` creates a variable, `=` updates an existing one:

```
x := 5
print(x)      // 5
x = x + 1
print(x)      // 6
```

Types you'll meet: numbers (`42`, `3.5`), strings (`"hi"`),
booleans (`true`, `false`). Convert between them when needed:

```
print(num("19") + 3)   // 22
print(str(42) + "!")   // 42!
```

## 3. Math that respects math

Multiplication goes before addition, parentheses first — like school:

```
print(2 + 3 * 4)     // 14, not 20
print((2 + 3) * 4)   // 20
```

Comparisons give back `true` / `false`, logic glues them:

```
age := 16
print(age >= 13 && age < 20)   // true
print(!(age == 10))            // true
```

## 4. Branching

```
score := 75
if score > 90 {
    print("S")
} else if score > 70 {
    print("A")
} else {
    print("try again")
}
```

Any value can be a condition: `0`, `""` and `false` count as false,
everything else as true.

## 5. Loops

Count with `for`, walk through values with `for-in`, repeat while
something holds with `while`. `break` exits, `continue` skips ahead:

```
for i := 0; i < 3; i = i + 1 {
    print(i)          // 0 1 2
}

for w in ["a", "b", "c"] {
    if w == "b" {
        continue
    }
    print(w)          // a c
}

n := 3
while n > 0 {
    print(n)          // 3 2 1
    n = n - 1
}
```

## 6. Your turn

Save this as `task1.cx`, guess what it prints, then run it:

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

`49` — the sum of 2..10 except the skipped 5 (2+3+4+6+7+8+9+10).

</details>

---

*Part 2 will cover arrays, maps, functions and structs — enough to write
a grade book. Parts ship one by one.*
