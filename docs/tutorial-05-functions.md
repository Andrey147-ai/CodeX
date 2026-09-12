# CodeX Tutorial, Part 5: Functions and methods

*Named recipes with inputs and outputs — plus structs that act.
~20 minutes.*

## 1. Your first function

`fn`, a name, params in parentheses, body in braces:

```
fn greet(name) {
    print("hello, ", name)
}
greet("Ann")   // hello, Ann
greet("Bob")   // hello, Bob
```

Write it once, call it ten times. If the same three lines appear twice
in your program, they want to be a function.

## 2. `return` gives something back

```
fn add(a, b) {
    return a + b
}
print(add(2, 3))       // 5
print(add(10, 20) * 2) // 60 — result is a normal value
```

`return` exits immediately — lines after it never run:

```
fn first(x) {
    return x
    print("unreachable")
}
```

## 3. Scope: what's inside stays inside

Params and `:=` variables live only inside the function.
Same name outside? Untouched:

```
x := 10
fn f() {
    x := 99
    return x
}
print(f())   // 99
print(x)     // still 10
```

Think of a function as a room with the door closed: it sees the outside
(globals), but the outside can't see its mess.

## 4. Methods: functions glued to structs

A receiver in parentheses before the name attaches the function to a type:

```
struct Player {
    name
    hp
}

fn (p Player) heal(x) {
    p.hp = p.hp + x
    return p.hp
}

p := Player{"hero", 50}
print(p.heal(30))   // 80
print(p.hp)         // 80 — changed for real, not a copy
```

Field writes through the receiver stick to the caller's struct.
That's what makes methods different from plain functions: they *act*
on something.

## 5. Mini-project: battle round

```
struct Enemy {
    name
    hp
}

fn (e Enemy) hit(dmg) {
    e.hp = e.hp - dmg
    if e.hp < 0 {
        e.hp = 0
    }
    return e.hp
}

fn alive(e) {
    return e.hp > 0
}

boss := Enemy{"dragon", 100}
round := 1
while alive(boss) {
    print("round ", round, ": ", boss.hit(30))
    round = round + 1
}
print("down in ", round - 1, " rounds")
```

```
round 1: 70
round 2: 40
round 3: 10
round 4: 0
down in 4 rounds
```

(You'll also see `[del]` lines between rounds — that's `alive()`'s
*copy* of the boss being freed after each check. The real boss lives on
in `main`, only copies get cleaned. Method receivers like `hit` don't
spam: they're protected for the call.)

## 6. Test yourself

**Task 1.** `fn max(a, b)` returning the bigger one — then use it to find
the biggest of three numbers.

<details>
<summary>Answer</summary>

```
fn max(a, b) {
    if a > b {
        return a
    }
    return b
}
print(max(max(3, 9), 5))   // 9
```

</details>

**Task 2.** Give `Enemy` a `rage()` method: doubles nothing, just adds 50
hp and returns the new hp. Call it mid-fight.

---

*Part 6 — [loops and logic workout](tutorial-06-loops.md): nested loops,
truth tables and killing infinite loops.*
