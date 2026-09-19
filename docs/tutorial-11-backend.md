# Tutorial 11: backend in 10 lines (mini_api)

You already know `print`, arrays, maps. Now a real JSON API.

## The whole server

```go
fn hello(req) {
    if req["path"] == "/hi" {
        return to_json({"hi": "there", "time": date()})
    }
    return {"status": 404, "body": "try /hi"}
}
http_listen(8081, hello)
```

Run: `..\codex.exe examples\mini_api.cx`, open `http://127.0.0.1:8081/hi`.

## How it works

* `http_listen(port, fn)` blocks and calls `fn` per request.
* `req` is a map: `req["path"]`, `req["method"]`, `req["query"]`, `req["body"]`.
* Return a string = 200 text. Return `{"status":..., "body":...}` for codes.
* `to_json` / `parse_json` convert maps. `date()` stamps responses.
* POST from CodeX: `http_post(url, "hi")`. From curl: `curl http://127.0.0.1:8081/hi`.

## Exercises

1. Add `/time` returning `{"now": now()}`.
2. Echo POST: if `req["path"] == "/echo"` return `req["body"]`.
3. Count hits with a global `hits := [0]` + `hits[0] = hits[0] + 1`.
