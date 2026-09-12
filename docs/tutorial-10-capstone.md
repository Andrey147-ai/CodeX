# CodeX Tutorial, Part 10: Capstone — a log analyzer that earns its keep

*Final part. We dissect `examples/logstat.cx` — a real tool that reads a
log file, counts levels, ranks errors and writes a report. Then three
missions. ~30 minutes.*

## 1. Run it first

From the `examples/` folder:

```
codex.exe logstat.cx app.log
```

```
file: app.log
total: 12
levels: {ERROR: 4, INFO: 5, WARN: 3}
--- unique errors ---
3x  db connection timeout
1x  file not found: avatar.png
report saved: true
```

No filename? It defaults to `app.log`. Now — why it works.

## 2. Act 1 — input

```
file := "app.log"
if len(args()) > 0 {
    file = args()[0]
}
log := read_file(file)
lines := split(log, "\n")
```

CLI args (Part 9) with a default, whole file into one string, split into
lines (Parts 3–4). Note the trailing-`\n` gotcha from Part 4: the last
piece after splitting is empty, hence the guard below.

## 3. Act 2 — counting with maps

```
levels := {}
for line in lines {
    if len(line) == 0 {
        continue
    }
    lv := "INFO"
    if contains(line, "ERROR") {
        lv = "ERROR"
    } else if contains(line, "WARN") {
        lv = "WARN"
    }
    if has(levels, lv) {
        levels[lv] = levels[lv] + 1
    } else {
        levels[lv] = 1
    }
}
```

The classic counter pattern (Part 7): classify, then bump. No counters
declared upfront — keys appear on first sight.

## 4. Act 3 — grouping errors

Timestamps make every line unique, so group by message text:

```
key := line
parts := split(line, "ERROR")
if len(parts) > 1 {
    key = parts[1]
}
if has(errs, key) {
    errs[key] = errs[key] + 1
} else {
    errs[key] = 1
}
```

Same counter pattern, smarter key. Result: `db connection timeout → 3`
instead of three lonely lines.

## 5. Act 4 — the report

Strings accumulate, then land in a file (Parts 3–4):

```
report := "log report: " + file + "\n"
for k in keys(levels) {
    report = report + k + ": " + str(levels[k]) + "\n"
}
write_file("report.txt", report)
```

Numbers become strings via `str()` before gluing — the type discipline
from Part 1, paying off.

## 6. Missions (do at least one)

**Mission 1 — group warnings too.** Mirror the `errs` map with a `warns`
map keyed on text after `"WARN"`, print it under its own header.

**Mission 2 — output filename as second arg.** `out := "report.txt"`,
`if len(args()) > 1 { out = args()[1] }`, then `write_file(out, ...)`.

**Mission 3 — headline.** Track the top error while iterating `keys(errs)`
(the `best` pattern from Part 2's grade book) and print
`"top issue: ..." ` first.

---

*That's the series. You can read real CodeX programs, write your own,
and debug them. The language keeps growing — check the roadmap in the
README for what's next. Games were just the demo jacket: the engine is
general-purpose.*
