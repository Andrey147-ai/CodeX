# CodeX Launch Kit — копируй и пости

Логотип: `site/logo.svg`. Скриншоты: запусти `examples/guess.cx` и `examples/mini_api.cx`, сними терминал.

## 1. Telegram / школьные чаты (RU)

> Мне 16, я написал свой язык программирования — CodeX.
> Первая игра за 15 минут, установка одной командой:
> `irm https://raw.githubusercontent.com/Andrey147-ai/CodeX/main/install.ps1 | iex`
> Бэкенд — 10 строк, пакеты — прямо с GitHub, 76 функций из коробки.
> Глянь: https://github.com/Andrey147-ai/CodeX
> Звезда репозиторию = топливо для v0.24 ⭐

## 2. X / Twitter (EN, 280 chars)

> I'm 16 and I built my own programming language: CodeX.
> First game in 15 min, backend in 10 lines, zero deps.
> `codex new hello.cx && codex hello.cx`
> github.com/Andrey147-ai/CodeX ⭐

## 3. Reddit r/learnprogramming (EN)

Title: `I built a tiny language so beginners ship a game on day one — CodeX`

Body:
> I'm Andrey, 16. I wanted to know how languages work inside, so I wrote CodeX in Go: lexer, parser, interpreter, zero deps.
> Why another language? Python has 1000 ways to do one thing. CodeX has one: `x := 5`, `for s in scores`, `http_listen(8081, hello)` — and you get a JSON API.
> - 1-command install, REPL, formatter, `assert()` tests
> - 76 builtins, GitHub-native packages, VS Code highlighting
> - 11 tutorials, first game in ~15 minutes
> Repo: https://github.com/Andrey147-ai/CodeX — roast my interpreter, I read everything.

## 4. Habr / Dev.to (идея статьи)

«Как школьник написал язык за лето: лексер за вечер, парсер за неделю, рантайм за месяц». Структура: зачем → токены с line:col → прескан структур (чтобы `if` не путать с литералом) → скоупы без GC → `try/catch` через panic → пакеты из zip с GitHub → что дальше. Код-примеры из `main.go` + `examples/`.

## 5. Видео 30 сек (Shorts/TikTok)

1. (0-5с) «Python учишь месяц. Смотри:» — терминал.
2. (5-15с) Печатаешь `codex new hello.cx`, `codex hello.cx` — игра работает.
3. (15-25с) Открываешь `mini_api.cx`, `http_listen` — браузер показывает JSON.
4. (25-30с) «CodeX, ссылка в описании» + логотип.

## 6. Elevator pitch (1 предложение)

CodeX — язык, где школьник пишет первую игру за вечер, а бэкенд — за 10 строк без фреймворков.

## Правила честности

Не пиши «самый быстрый» и «убийца Python» — легко проверяется и бьет по репутации. Говори: маленький, честный, для старта.
