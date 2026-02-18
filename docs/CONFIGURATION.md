# Configuration Reference

TUIper reads JSON config from:

- default:
  - Linux/macOS: `~/.config/tuiper/config.json`
  - Windows: `%AppData%\\tuiper\\config.json`
- override with: `-config /path/to/file.json`

If the file does not exist, built-in defaults are used.

## Schema

```json
{
  "normal_words": ["the", "quick"],
  "special_char_words": ["!@#$", "%^&*"],
  "durations": ["15s", "30s", "1m", "2m"],
  "prompt_word_count": 18,
  "quote_endpoint": "https://dummyjson.com/quotes/random",
  "go_example_endpoint": "",
  "go_examples": ["for i := 0; i < 3; i++ { fmt.Println(i) }"]
}
```

## Fields

- `normal_words`: non-empty array of words for normal mode prompt generation.
- `special_char_words`: non-empty array for special-character mode.
- `durations`: non-empty array of Go durations (`15s`, `1m`, etc.) shown in the UI.
- `prompt_word_count`: integer > 0 for generated prompt length.
- `quote_endpoint`: quote API endpoint. Expected JSON keys: `content` or `quote` or `text`.
- `go_example_endpoint`:
  - empty string disables remote code fetch (recommended for local-only operation)
  - otherwise endpoint may return JSON (`content`/`code`/`text`) or plain text.
- `go_examples`: local fallback snippets used by code practice mode.

## Remote Fallback Behavior

- Quote/code remote failures automatically fall back to local prompts.
- Backoff is applied after repeated failures to avoid hammering unstable endpoints.
- Prompts attempt to avoid immediate repetition.

## Recommended Setup

- Keep `go_example_endpoint` empty unless you control the endpoint quality.
- Provide curated `go_examples` for predictable typing content.
- Keep `durations` short and practical for TUI workflows.
