# CodeX in VS Code

## 1. Install extension (dev, 1 min)

Copy `editors/vscode` to `%USERPROFILE%\.vscode\extensions\codex-language-0.22.0`, reload VS Code.

You get: `.cx` highlight, snippets (`print`, `fn`, `forin`, `if`, `struct`, `server`, `try`), `//` comments, auto-close.

## 2. Run + format

`codex.exe` is CLI-first:

* `codex.exe new hello.cx` — template
* `codex.exe hello.cx` — run
* `codex.exe fmt hello.cx` — format
* `codex.exe fmt --check hello.cx` — CI check (exit 1 if dirty)
* `codex.exe test tests` — asserts

Suggested `tasks.json`:

```json
{
  "version": "2.0.0",
  "tasks": [
    {"label": "codex: run", "type": "shell", "command": "${workspaceFolder}/codex.exe ${file}"},
    {"label": "codex: fmt", "type": "shell", "command": "${workspaceFolder}/codex.exe fmt ${file}"}
  ]
}
```

## 3. New builtins (v0.22 wave 2+3)

`pop`, `reverse`, `shuffle`, `choice`, `extend`, `unique`, `sum`, `avg`,
`file_size`, `is_dir`, `cwd`, `clock`, `http_post`,
`index_of`, `count`, `lines`, `values`, `merge`, `delete_key`,
`is_nil/is_array/is_map/is_string/is_num`, `chr/ord`,
`reduce/find/any/all`. Demo: `examples/funcs_demo.cx`.
