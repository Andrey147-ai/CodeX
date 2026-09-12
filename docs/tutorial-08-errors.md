# CodeX Tutorial, Part 8: Errors and debugging

*Red text is a letter from the interpreter. Learn to read it.
~20 minutes.*

## 1. The three error families

```
Lexer error at 2:6: unknown character '@'     // weird symbol
Parser error at 5:3: unexpected token ...     // grammar broken
Runtime error: index 9 out of range (len 3)   // ran, then fell
```

**Lexer** = "I can't even read this character". Check for typos and
stray symbols. The `2:6` is line 2, column 6 — go look there.

**Parser** = "I read the words but the sentence makes no sense".
Unclosed brackets, missing `{`, stray tokens.

**Runtime** = "I understood you and tried — then reality disagreed".
Wrong types, missing keys, crashed conversions. The program *started*,
so the bug is in logic or data.

## 2. Hall of fame: mistakes everyone makes

**Single `=` in a condition:**

```
if x = 5 {   // ERROR — = assigns, == compares
```

**Unclosed string swallows the file:**

```
print("hello)   // ERROR — the closing " is missing
```

**Off-by-one index:**

```
a := [1, 2, 3]
print(a[3])   // ERROR — last index is 2, use len(a) - 1
```

**Missing map key:**

```
m := {"a": 1}
print(m["b"])          // ERROR — check has(m, "b") first
```

**Unknown name (typo):**

```
print(scor)   // ERROR: undefined variable 'scor' — meant score
```

**`num()` on garbage:**

```
n := num("abc")   // ERROR — validate with your own isNumber() first
```

## 3. Debugging strategy

1. **Read the message literally.** `undefined variable 'scor'` tells
   you the name *and* the problem. `at 5:3` tells you where.
2. **Print the suspects.** Before the crashing line, print every value
   involved. Nine times out of ten you'll go "oh".
3. **Halve the program.** Comment out half with `//`. Still crashes?
   The bug is in the live half. Repeat until one line is left.
4. **Check types.** `print` shows values, and most crashes are
   `"5" + 5` style mixups — `str()`/`num()` at the borders fixes them.

## 4. Practice: find the bug

Each snippet has exactly one bug. Answers below.

**A.**
```
for i := 0; i < 10; i = i + 1
    print(i)
```

**B.**
```
a := [1, 2, 3]
print(a[len(a)])
```

**C.**
```
fn add(a, b) {
    return a + b
}
print(add(1))
```

<details>
<summary>Answers</summary>

**A.** Missing `{` after the `for (...)` line — parser error.
**B.** `a[len(a)]` is one past the end — last valid index is `len(a) - 1`.
**C.** `add` needs 2 args, got 1 — the missing one becomes `nil`, and
`1 + nil` is a runtime type error.

</details>

---

*Part 9 — [packages](tutorial-09-packages.md): share code through GitHub.*
