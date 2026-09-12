# CodeX Tutorial, Part 4: Files

*Programs that remember: writing notes, reading them back, analyzing
logs. ~20 minutes. Run these from the folder your files live in —
paths are relative to where you launch `codex.exe`.*

## 1. Write, append, read

Three verbs. `write_file` creates or overwrites, `append_file` adds to
the end, `read_file` slurps the whole file into a string:

```
write_file("notes.txt", "milk\n")
append_file("notes.txt", "bread\n")
append_file("notes.txt", "eggs\n")
print(read_file("notes.txt"))
```

```
milk
bread
eggs
```

`write_file` returns how many bytes it wrote (handy for a quick check),
`read_file` on a missing file stops with an error — so check first:

```
print(exists("notes.txt"))   // true
print(exists("nope.txt"))    // false
```

## 2. Files are just strings

Once loaded, a file is an ordinary string — everything from Part 3 works.
Split into lines and walk them:

```
text := read_file("notes.txt")
lines := split(text, "\n")
print(len(lines))   // 4?!
```

Four, not three! The file ends with `\n`, so splitting leaves one empty
tail piece: `["milk", "bread", "eggs", ""]`. Real scripts skip empties:

```
count := 0
for line in lines {
    if len(line) == 0 {
        continue
    }
    count = count + 1
    print("- ", line)
}
print("items: ", count)   // items: 3
```

## 3. Mini-project: log checker

Server logs are lines with a level word. Count them:

```
write_file("mini.log", "INFO start\nERROR boom\nINFO ok\nERROR boom\n")
log := read_file("mini.log")
errors := 0
for line in split(log, "\n") {
    if contains(line, "ERROR") {
        errors = errors + 1
    }
}
print("errors: ", errors)   // errors: 2
```

The full version lives in the repo as `examples/logstat.cx` (run it from
`examples/`): it groups identical errors, ranks them, and writes
`report.txt`. Same ideas, bigger file.

## 4. Test yourself

**Task 1.** A diary: ask the user for a line with `input()` and append it
to `diary.txt` with a newline. Run it three times, then read the file.

<details>
<summary>Answer</summary>

```
line := input("today: ")
append_file("diary.txt", line + "\n")
print("saved")
```

</details>

**Task 2.** Write `"a\nb\nc\n"` to a file, read it back, print how many
non-empty lines it has (answer: 3).

---

*Part 5 — [functions and methods](tutorial-05-functions.md): recipes with
inputs, outputs, and structs that act.*
