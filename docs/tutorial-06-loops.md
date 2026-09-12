# CodeX Tutorial, Part 6: Loops and logic workout

*You met loops in Part 1. Now train them until they're boring.
~20 minutes.*

## 1. Truth first

Every condition is a yes/no question. The building blocks:

```
print(true && true)    // true
print(true && false)   // false
print(false || true)   // true
print(!true)           // false
```

`&&` needs both sides true. `||` needs at least one. `!` flips.
Parentheses when in doubt:

```
age := 16
tired := false
if age >= 13 && !tired {
    print("party")
}
```

## 2. Which loop when

| Situation | Loop |
|---|---|
| exact count | `for i := 0; i < n; i = i + 1` |
| every item | `for x in arr` |
| until something changes | `while cond` |
| forever (servers, games) | `while true` + `break` |

## 3. Nested loops: tables

A loop inside a loop. The inner one restarts for every outer step:

```
for i := 1; i <= 3; i = i + 1 {
    row := ""
    for j := 1; j <= 3; j = j + 1 {
        row = row + str(i * j) + " "
    }
    print(row)
}
```

```
1 2 3
4 5 6
7 8 9
```

## 4. `break` vs `continue`, precisely

```
for i := 1; i <= 5; i = i + 1 {
    if i == 2 {
        continue   // skip printing 2
    }
    if i == 4 {
        break      // stop entirely, 4 and 5 never print
    }
    print(i)       // 1 3
}
```

In a classic `for`, `continue` still runs the post step (`i = i + 1`) —
the loop can't get stuck. In `while`, **you** advance the counter, so a
misplaced `continue` before `i = i + 1` loops forever.

## 5. Surviving infinite loops

```
while true {
    print("spinning")
    sleep(500)
}
```

Run it, watch three lines, press `Ctrl+C`. Rules of the craft:

1. Every `while` needs a line that *can* flip its condition.
2. In `while true`, the `break` must be reachable — trace it by hand.
3. When stuck, add `print` of the counter: you'll see where it freezes.

## 6. Test yourself

**Task 1.** Sum of even numbers 1..20 (answer: 110).

<details>
<summary>Answer</summary>

```
total := 0
isEven := false
for i := 1; i <= 20; i = i + 1 {
    if isEven {
        total = total + i
    }
    isEven = !isEven   // flip the flag every step
}
print(total)   // 110
```

(No modulo operator in CodeX yet — a toggling flag does the job.)

</details>

**Task 2.** Countdown `5 4 3 2 1 liftoff!` with a `while`, then the same
with a `for`. Both print the same lines.

---

*Part 7 — [maps and data modeling](tutorial-07-maps.md): phonebooks,
carts and configs.*
