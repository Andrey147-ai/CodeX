# CodeX Tutorial, Part 9: Packages

*Stop copying code between projects — import it. From GitHub, no
registry. ~20 minutes.*

## 1. Local imports

Two files side by side. The library only defines, the app uses:

`mymath.cx`:
```
fn square(x) {
    return x * x
}
fn cube(x) {
    return x * x * x
}
```

`app.cx`:
```
import "./mylib.cx"
```

Wait — the path must match the file. If the library is `mymath.cx`:

```
import "./mymath.cx"
print(square(5))   // 25
print(cube(3))     // 27
```

Imported files run once, even if imported twice, and their functions
land in your program as if written there. (Try `examples/useimport.cx`
in the repo.)

## 2. GitHub imports

Any public repo with `main.cx` in the root is a package:

```
import "Andrey147-ai/strutils"                // latest code
import "Andrey147-ai/strutils@v1.2.0"         // exact tag or branch
import "Andrey147-ai/strutils/lib/text.cx"    // a specific file
```

First run downloads a zip into `~/.codex/pkgs` and reuses it after.
Offline? It just works from cache.

Prefetch without running anything:

```
codex get Andrey147-ai/strutils@v1.2.0
```

## 3. Publish your own in 2 minutes

1. New public repo, e.g. `my-utils`.
2. Add `main.cx` with your functions.
3. Push. Done — anyone can `import "you/my-utils"`.

Pinning versions is a habit, not politeness: `@v1.0.0` means your
program never breaks because someone rewrote `main` upstream.

## 4. Data repos with `pkgdir`

Repos aren't only code. `pkgdir("user/repo")` returns the cached path,
so scripts can read data files straight from GitHub:

```
base := pkgdir("trung-kieen/bad-apple-ascii") + "/frames-ascii/"
print(read_file(base + "out0001.jpg.txt"))
```

That's how `examples/badapple.cx` streams 6572 video frames.

## 5. Test yourself

**Task 1.** Move `max()` from Part 5 into `mymath.cx`, import it, find
the biggest of `[4, 9, 2, 7]` with a loop.

**Task 2.** Publish `mymath.cx` as your first package (any repo name),
then `codex get` it by its full `user/repo` path.

---

*Part 10 — [capstone: shop game](tutorial-10-shop.md): we read a real
program together, then you extend it.*
