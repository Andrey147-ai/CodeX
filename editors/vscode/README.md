# CodeX for VS Code

Lightweight support for `.cx` files.

## Install (dev)

1. Copy folder `editors/vscode` to:
   `%USERPROFILE%\.vscode\extensions\codex-language-0.23.0`
2. Reload VS Code, open any `.cx` file.

## Features

- Syntax highlight: keywords `if/for/fn/struct`, strings, numbers, 50+ builtins
- Snippets: `print`, `fn`, `forin`, `fori`, `if`, `struct`, `server`, `try`
- Config: `//` comments, auto-closing `{} [] () ""`

## Fmt on save (optional)

Add to `settings.json`:

```json
{
  "[codex]": {
    "editor.formatOnSave": true
  }
}
```

And external tool task calling `codex.exe fmt ${file}`.
