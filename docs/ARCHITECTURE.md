# TUIper Architecture

## Overview

TUIper follows a layered architecture:

- `main.go`: CLI entrypoint + Bubble Tea state machine + rendering
- `internal/config`: config schema, defaults, validation, loading
- `internal/prompt`: prompt generation/fetching, retry/backoff, sanitization

This keeps UI orchestration separate from domain logic and external I/O.

## Runtime Flow

1. `main.go` parses flags (`-config`, `-man`).
2. `config.Load(...)` returns validated `RuntimeConfig`.
3. UI model is initialized with:
   - validated runtime config
   - `prompt.Service` dependency
4. UI state transitions:
   - splash -> mode select -> duration select -> typing session
5. Prompt selection delegates to `prompt.Service` by mode.

## Prompt Service Responsibilities

`internal/prompt.Service` owns:

- mode-aware prompt selection
- quote fetch + fallback + retry/backoff
- go code fetch + payload normalization + retry/backoff
- prompt non-repetition where possible

UI code does not directly handle remote fetch or sanitization details.

## Config Responsibilities

`internal/config` owns:

- user config schema (`AppConfig`)
- validated runtime config (`RuntimeConfig`)
- defaults and missing-file behavior
- duration string parsing/validation

This prevents config semantics from leaking into UI code.

## Testing Strategy

- `internal/config/config_test.go`: validation/load/default behavior
- `internal/prompt/service_test.go`: provider/retry/sanitization behavior
- `main_test.go`: local UI helper behavior

Use `make check` to run fmt + tests + build.

## Extension Guidelines

- Add new modes by:
  - extending `prompt.Mode`
  - implementing selection logic in `prompt.Service`
  - adding UI label via `prompt.ModeLabels()`
- Keep network and parsing logic inside `internal/prompt`.
- Keep terminal rendering and key handling inside `main.go`.
