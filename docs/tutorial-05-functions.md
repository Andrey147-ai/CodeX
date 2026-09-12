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

## 5. Mini-project: bank account

```
struct Account {
    owner
    balance
}

fn (a Account) deposit(x) {
    a.balance = a.balance + x
    return a.balance
}

fn (a Account) withdraw(x) {
    if x > a.balance {
        print("denied: not enough funds")
        return a.balance
    }
    a.balance = a.balance - x
    return a.balance
}

acc := Account{"Ann", 100}
print(acc.deposit(50))    // 150
print(acc.withdraw(30))   // 120
print(acc.withdraw(500))  // denied... 120
print(acc.balance)        // 120
```

Two methods sharing one balance, a guard clause that refuses bad
operations, every step returning the new state — the shape of most real
business logic.

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

**Task 2.** Give `Account` an `interest()` method: adds 10% to the
balance and returns it. (`a.balance / 10` is the tenth part.)

---

*Part 6 — [loops and logic workout](tutorial-06-loops.md): nested loops,
truth tables and killing infinite loops.*
