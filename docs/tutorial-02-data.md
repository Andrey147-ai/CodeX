# CodeX Tutorial, Part 2: Data — arrays, maps, functions, structs

*Part 1 covered variables, math, branches and loops. Now we store and
organize data — and finish with a grade book. ~25 minutes.*

## 1. Arrays: ordered lists

Square brackets, zero-based indexing (first element is `[0]`):

```
scores := [90, 85, 95]
print(scores[0])   // 90
print(len(scores)) // 3
```

Rewrite an element, grow the end:

```
scores[1] = 100
push(scores, 70)
print(scores)      // [90, 100, 95, 70]
```

Sort in place (all numbers or all strings — no mixing):

```
nums := [5, 2, 8]
sort(nums)
print(nums)        // [2, 5, 8]
```

Walk with the loops from Part 1:

```
total := 0
for i := 0; i < len(scores); i = i + 1 {
    total = total + scores[i]
}
for s in scores {
    print(s)
}
```

**Out of range is an error, not a silent bug:**

```
print(scores[99])   // Runtime error: index 99 out of range (len 4)
```

That's deliberate — you'd rather hear about it than debug garbage.

## 2. Maps: key → value

Curly braces with `key: value` pairs. Keys are strings:

```
cfg := {"debug": true, "retries": 3}
print(cfg["debug"])   // true
print(len(cfg))       // 2
```

Add or overwrite keys by assignment — missing keys are created:

```
cfg["timeout"] = 30
cfg["retries"] = 5
print(cfg)   // {debug: true, retries: 5, timeout: 30}
```

Reading a missing key is an error, so check first with `has`,
and list keys (sorted) with `keys`:

```
if has(cfg, "debug") {
    print("debug is on")
}
for k in keys(cfg) {
    print(k, " = ", cfg[k])
}
```

Arrays of maps, maps of arrays — nest freely:

```
users := [{"name": "Ann", "hp": 100}, {"name": "Bob", "hp": 80}]
print(users[1]["name"])   // Bob
```

## 3. Functions: named recipes

`fn` with params, `return` hands a value back:

```
fn add(a, b) {
    return a + b
}
print(add(2, 3))   // 5
```

No `return`? The function gives back the last expression's value.
Params and locals stay inside — the outside world never sees them:

```
x := 10
fn f() {
    x := 99
    return x
}
print(f())   // 99
print(x)     // still 10
```

## 4. Structs: your own types

Declare fields, build with positional values, read and write with dots —
even inside math:

```
struct Player {
    name
    hp
}

p := Player{"hero", 50}
print(p.hp)          // 50
p.hp = p.hp - 20
print(p.hp)          // 30
```

Methods attach with a receiver in parentheses. Field writes inside a
method stick to the caller:

```
fn (p Player) heal(x) {
    p.hp = p.hp + x
    return p.hp
}
print(p.heal(30))   // 60
print(p.hp)         // 60 — changed for real
```

When a struct variable leaves its scope, the runtime frees it and logs
`[del]` — that's the scope-based memory from the README, no GC involved.

## 5. Capstone: grade book

Everything above in one script (`examples/showcase.cx` in the repo):

```
struct Student {
    name
    scores
}

fn (s Student) avg() {
    total := 0
    for x in s.scores {
        total = total + x
    }
    return total / len(s.scores)
}

fn grade(a) {
    if a >= 90 {
        return "A"
    } else if a >= 75 {
        return "B"
    } else if a >= 60 {
        return "C"
    } else {
        return "F"
    }
}

group := [Student{"Ann", [90, 85, 95]}, Student{"Bob", [70, 60, 80]}]
report := {}
for s in group {
    report[s.name] = grade(s.avg())
}
print(report)   // {Ann: A, Bob: C}
```

## 6. Test yourself

**Task 1.** Reverse an array into a new one without `sort`:

```
src := [1, 2, 3, 4]
rev := []
// your code here
print(rev)   // [4, 3, 2, 1]
```

<details>
<summary>Answer</summary>

```
i := len(src) - 1
while i >= 0 {
    push(rev, src[i])
    i = i - 1
}
```

</details>

**Task 2.** Count words in a sentence with a map:

```
text := "cat dog cat bird dog cat"
// your code here → {bird: 1, cat: 3, dog: 2}
```

<details>
<summary>Answer</summary>

```
words := split(text, " ")
freq := {}
for w in words {
    if has(freq, w) {
        freq[w] = freq[w] + 1
    } else {
        freq[w] = 1
    }
}
print(freq)
```

</details>

---

*Part 3 — [strings](tutorial-03-strings.md): text tools and escapes.*
