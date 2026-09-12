# CodeX Tutorial, Part 10: Capstone — read a game, then extend it

*Final part. We dissect `examples/shop_game.cx` (~90 lines: shop,
boss fight, save file), then three missions. ~30 minutes.*

## 1. Run it first

```
codex.exe shop_game.cx
```

Type a name, buy `1` (sword), `2` (potion), leave with `0`, watch the
fight. Now let's see why it works.

## 2. Act 1 — the player

```
struct Player {
    name
    gold
    hp
}

fn (p Player) takeDamage(amount) {
    p.hp = p.hp - amount
    return p.hp
}
```

A struct plus methods (Part 2/5). Methods mutate the receiver, so damage
sticks. `status()` prints one summary line — called after every purchase
so the player always sees fresh numbers.

## 3. Act 2 — the shop loop

```
shop := ["sword:50", "potion:30", "shield:40"]
```

Items are `"name:price"` strings (Part 3: `split(shop[i], ":")` peels
them). The loop is `while true` with `break` on `"0"` (Part 6):

```
choice := input("buy? ")
if choice == "0" {
    break
}
if !isNumber(choice) {
    print("numbers only!")
}
```

`isNumber` guards `num()` — converting garbage crashes, so the game
validates first (Part 8 thinking). Buying checks gold, pushes the item
into `inv`, counts potions for the battle.

## 4. Act 3 — the fight

```
dmg := 20
if contains(join(inv, ","), "sword") {
    dmg = dmg + 15
}
boss := 120
while p.hp > 0 && boss > 0 {
    boss = boss - dmg
    if boss > 0 {
        p.takeDamage(15)
        if p.hp < 40 && potions > 0 {
            p.heal(40)
            potions = potions - 1
        }
    }
}
```

Three ideas composed: equipment check via string search, a two-sided
`while` condition, and potions as a conditional heal. Victory adds 100
gold; either way the result lands in `save.txt` via `write_file`
(Part 4) and is read back as proof.

## 5. Missions (do at least one)

**Mission 1 — shield that works.** Buying `shield` currently does nothing.
Make it halve boss damage while equipped (`dmg_taken := 15`, `7` with
shield — division works on numbers).

**Mission 2 — price list.** After leaving the shop, print the inventory
with prices using a price map: `{"sword": 50, "potion": 30}`.

**Mission 3 — second boss.** After victory, offer `input("again? ")`:
`"yes"` spawns a 200-hp boss, anything else ends the game. Wrap the
fight in `while` and count wins.

---

*That's the series. You can read real CodeX programs, write your own,
and debug them. The language keeps growing — check the roadmap in the
README for what's next.*
