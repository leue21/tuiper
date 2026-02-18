# TUIper

TUIper is a terminal typing trainer written in Go with a Monkeytype-inspired flow.

## Features

- Responsive TUI layout for small and large terminals
- Catppuccin-inspired adaptive theme (dark/light aware)
- Mode selection:
  - `Normal`
  - `Special Chars Practice`
  - `Quote Practice` (remote API + fallback)
  - `Code Practice` (remote/plain-text API + fallback)
- In-app duration selection
- JSON configuration overrides
- Built-in help and man page support

## Requirements

- Go 1.22+

## Quick Start

```bash
make build
./bin/tuiper
```

## Run

```bash
make run
```

Use a custom config path:

```bash
make run CONFIG=./tuiper.json
```

## Build

```bash
make build
./bin/tuiper
```

## Controls

- Global:
  - `Ctrl+C` quit
- Splash:
  - `Enter` continue
- Mode/Duration selection:
  - Arrow keys to move selection
  - Number keys (`1..N`) quick select
  - `Enter` confirm
- Typing screen:
  - `Backspace` delete one character
  - `Enter` after a completed test returns to mode menu
- Auto-advance:
  - `Quote Practice` and `Code Practice` auto-load next prompt when finished

## Help and Man Page

```bash
./bin/tuiper -h
./bin/tuiper -man
make man
```

## Configuration

By default, TUIper looks for config in the user config directory:

- Linux/macOS: `~/.config/tuiper/config.json`

If the file does not exist, built-in defaults are used.

Start from the example:

```bash
mkdir -p ~/.config/tuiper
cp tuiper.example.json ~/.config/tuiper/config.json
```

Supported keys and defaults:

- `normal_words`: list of words for `Normal` mode
- `special_char_words`: list of symbol-heavy tokens for `Special Chars Practice`
- `durations`: list used in duration menu, e.g. `["15s","30s","1m","2m"]`
- `prompt_word_count`: words per generated prompt (normal/special modes)
- `quote_endpoint`: default `https://dummyjson.com/quotes/random`
- `go_example_endpoint`: default `""` (disabled; uses local `go_examples`)
- `go_examples`: fallback snippets for `Code Practice`

Endpoint format notes:

- `quote_endpoint` expects JSON with at least one of: `content`, `quote`, `text`.
- `go_example_endpoint` accepts:
  - JSON with one of: `content`, `code`, `text`
  - plain text response (source/snippet)

Recommended code endpoint approach:

- Prefer your own curated JSON endpoint for stable, clean snippets.
- If using raw source endpoints, TUIper sanitizes headers/imports/comments.

## Architecture

Project structure is intentionally layered:

- `cmd`/root `main.go`: CLI + Bubble Tea UI state machine
- `internal/config`: config parsing, validation, defaults
- `internal/prompt`: prompt providers, retry/backoff, sanitization
- `docs/tuiper.1`: man page source

See:

- `docs/ARCHITECTURE.md`
- `docs/CONFIGURATION.md`

## Testing and Quality

```bash
make fmt
make test
make check
```

## Troubleshooting

- `Terminal too small`:
  - Resize terminal to at least `24x10`.
- Remote prompt failures:
  - App falls back automatically to local prompt pools.
- Want only local code snippets:
  - Keep `"go_example_endpoint": ""` in config.
- Need reproducible behavior in CI:
  - Use local-only config sources and run `make check`.
