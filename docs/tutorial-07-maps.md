# CodeX Tutorial, Part 7: Maps and data modeling

*Arrays keep order, maps keep meaning. Phonebooks, carts, configs.
~20 minutes.*

## 1. Array or map?

An array answers *"what's 3rd?"*. A map answers *"what's Ann's number?"*.
If you catch yourself writing `if name == "Ann"` chains — you want a map:

```
// array thinking (painful)
names := ["Ann", "Bob"]
numbers := ["111", "222"]   // parallel arrays drift apart...

// map thinking (calm)
phone := {"Ann": "111", "Bob": "222"}
print(phone["Ann"])   // 111
```

## 2. Build, read, update

```
cart := {}
cart["apple"] = 3
cart["bread"] = 1
cart["apple"] = cart["apple"] + 2   // read, add, write back
print(cart)          // {apple: 5, bread: 1}
print(len(cart))     // 2 — number of keys
```

New keys appear on assignment. But reading a missing key is an error,
so guard with `has`:

```
if has(cart, "milk") {
    print(cart["milk"])
} else {
    print("no milk today")
}
```

## 3. Walk all keys

`keys()` returns keys sorted — deterministic order, every run:

```
prices := {"apple": 50, "bread": 30, "milk": 40}
total := 0
for k in keys(prices) {
    print(k, " costs ", prices[k])
    total = total + prices[k]
}
print("total: ", total)   // total: 120
```

## 4. Maps inside maps (and arrays)

Model anything by nesting. A tiny shop:

```
shop := {
    "sword": {"price": 50, "dmg": 15},
    "potion": {"price": 30, "heal": 40}
}
print(shop["sword"]["dmg"])   // 15
shop["potion"]["price"] = 25
```

And an array of maps — the classic table:

```
party := [
    {"name": "Ann", "hp": 100},
    {"name": "Bob", "hp": 80}
]
for hero in party {
    print(hero["name"], " has ", hero["hp"], " hp")
}
```

## 5. Mini-project: word counter

```
text := "cat dog cat bird dog cat"
freq := {}
for w in split(text, " ") {
    if has(freq, w) {
        freq[w] = freq[w] + 1
    } else {
        freq[w] = 1
    }
}
print(freq)   // {bird: 1, cat: 3, dog: 2}
```

This exact pattern counts anything: errors in logs (see
`examples/logstat.cx`), votes, letters.

## 6. Test yourself

**Task 1.** A grade map `{"Ann": 90, "Bob": 70, "Cat": 95}` — print only
the names with 80+.

<details>
<summary>Answer</summary>

```
grades := {"Ann": 90, "Bob": 70, "Cat": 95}
for name in keys(grades) {
    if grades[name] >= 80 {
        print(name)
    }
}
// Ann Cat
```

</details>

**Task 2.** Baskets: `basket := {"apple": 3}`. Add 2 apples and 1 milk
without touching the literal again. Print the map.

---

*Part 8 — [errors and debugging](tutorial-08-errors.md): read the red
text and fix things like a detective.*
