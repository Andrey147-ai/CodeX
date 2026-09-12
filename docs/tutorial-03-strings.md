# CodeX Tutorial, Part 3: Strings

*Text is everywhere: names, files, web pages. Learn to bend it. ~20 minutes.*

## 1. What a string is

Anything in double quotes. Strings can sit in variables like numbers:

```
name := "Ada"
greeting := "hello there"
print(name)
```

An empty string `""` has nothing inside — remember it, it counts as
*false* in `if` (Part 1).

## 2. Length and gluing

`len` counts characters, `+` glues two strings (both sides must be
strings — a number needs `str()` first):

```
print(len("CodeX"))          // 5
print("un" + "real")         // unreal
print("level " + str(7))     // level 7
print("level " + 7)          // ERROR: can't mix string + number
```

## 3. Case and searching

```
print(upper("hello"))               // HELLO
print(lower("HeLLo WoRLD"))         // hello world
print(contains("hello world", "world"))   // true
print(contains("hello world", "mars"))    // false
```

`contains` answers yes/no — perfect inside `if`:

```
msg := "ERROR: disk full"
if contains(msg, "ERROR") {
    print("something broke!")
}
```

## 4. Split and join: string ↔ array

`split` cuts a string into an array, `join` glues an array back.
They are opposites:

```
print(split("a,b,c", ","))       // [a, b, c]
print(join(["red", "blue"], " + "))   // red + blue
```

Real use: peel a CSV line apart.

```
line := "Ann,90,95"
parts := split(line, ",")
print(parts[0])   // Ann
print(parts[1])   // 90
```

And back together with a different separator:

```
print(join(parts, " | "))   // Ann | 90 | 95
```

## 5. Special characters (escapes)

Some characters can't be typed directly, so a backslash encodes them:

| Code | Meaning |
|---|---|
| `\n` | new line |
| `\t` | tab |
| `\\` | a plain backslash |
| `\"` | a quote inside quotes |
| `\e` | escape char for terminal control (Part 10) |

```
print("a\nb")              // a
                           // b
print("say \"hi\"")        // say "hi"
print("C:\\codex\\main.cx")  // C:\codex\main.cx
```

## 6. Mini-project: badge printer

```
first := "ada"
last := "lovelace"
print("=== " + upper(first) + " " + upper(last) + " ===")
for w in split("codex is fun", " ") {
    print(upper(w))
}
```

```
=== ADA LOVELACE ===
CODEX
IS
FUN
```

## 7. Test yourself

**Task 1.** Build an email from a name: `"ada"` → `"ada@codex.dev"`.
Then shout it.

<details>
<summary>Answer</summary>

```
name := "ada"
email := name + "@codex.dev"
print(upper(email))   // ADA@CODEX.DEV
```

</details>

**Task 2.** Count how many items are in `"milk,eggs,bread,butter"`
without counting by hand.

<details>
<summary>Answer</summary>

```
print(len(split("milk,eggs,bread,butter", ",")))   // 4
```

</details>

---

*Part 4 — [files](tutorial-04-files.md): read, write, and analyze a real log.*
