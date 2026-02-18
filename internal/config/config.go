package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"
)

var defaultWords = []string{
	"the", "quick", "brown", "fox", "jumps", "over", "lazy", "dog", "focus", "speed",
	"accuracy", "rhythm", "smooth", "motion", "keyboard", "muscle", "memory", "clean",
	"timing", "streak", "practice", "session", "result", "screen", "terminal", "signal",
	"input", "output", "model", "render", "frame", "cursor", "character", "letter",
	"word", "line", "flow", "steady", "sharp", "calm", "consistent", "deliberate",
}

var defaultSpecialCharWords = []string{
	"!@#$", "%^&*", "()_+", "[]{}", "{}[]", "<>[]", "/?\\|", "`~", ";;::", "\"'\"'",
	"==!=", "++--", "<<>>", "||&&", "@@##", "$$%%", "^^~~", ".,<>", "///\\", "(()))",
}

type AppConfig struct {
	NormalWords       []string `json:"normal_words"`
	SpecialCharWords  []string `json:"special_char_words"`
	Durations         []string `json:"durations"`
	PromptWordCount   int      `json:"prompt_word_count"`
	QuoteEndpoint     string   `json:"quote_endpoint"`
	GoExampleEndpoint string   `json:"go_example_endpoint"`
	GoExamples        []string `json:"go_examples"`
}

type RuntimeConfig struct {
	Words             []string
	SpecialCharWords  []string
	DurationOptions   []time.Duration
	DurationLabels    []string
	PromptWordCount   int
	QuoteEndpoint     string
	GoExampleEndpoint string
	GoExamples        []string
}

func Default() AppConfig {
	return AppConfig{
		NormalWords:       append([]string(nil), defaultWords...),
		SpecialCharWords:  append([]string(nil), defaultSpecialCharWords...),
		Durations:         []string{"15s", "30s", "1m", "2m"},
		PromptWordCount:   18,
		QuoteEndpoint:     "https://dummyjson.com/quotes/random",
		GoExampleEndpoint: "",
		GoExamples: []string{
			"for i := 0; i < 10; i++ { fmt.Println(i) }",
			"if err != nil { return fmt.Errorf(\"failed: %w\", err) }",
			"items := []string{\"go\", \"tui\"}; for _, it := range items { fmt.Println(it) }",
		},
	}
}

func Resolve(cfg AppConfig) (RuntimeConfig, error) {
	if len(cfg.NormalWords) == 0 {
		return RuntimeConfig{}, fmt.Errorf("normal_words must not be empty")
	}
	if len(cfg.SpecialCharWords) == 0 {
		return RuntimeConfig{}, fmt.Errorf("special_char_words must not be empty")
	}
	if cfg.PromptWordCount <= 0 {
		return RuntimeConfig{}, fmt.Errorf("prompt_word_count must be > 0")
	}
	if len(cfg.Durations) == 0 {
		return RuntimeConfig{}, fmt.Errorf("durations must not be empty")
	}

	durationOptions := make([]time.Duration, 0, len(cfg.Durations))
	durationLabels := make([]string, 0, len(cfg.Durations))
	for _, raw := range cfg.Durations {
		d, err := time.ParseDuration(raw)
		if err != nil {
			return RuntimeConfig{}, fmt.Errorf("invalid duration %q: %w", raw, err)
		}
		if d <= 0 {
			return RuntimeConfig{}, fmt.Errorf("duration %q must be > 0", raw)
		}
		durationOptions = append(durationOptions, d)
		durationLabels = append(durationLabels, raw)
	}

	quoteEndpoint := strings.TrimSpace(cfg.QuoteEndpoint)
	if quoteEndpoint == "" {
		quoteEndpoint = Default().QuoteEndpoint
	}
	goExampleEndpoint := strings.TrimSpace(cfg.GoExampleEndpoint)
	if goExampleEndpoint == "" {
		goExampleEndpoint = Default().GoExampleEndpoint
	}
	goExamples := append([]string(nil), cfg.GoExamples...)
	if len(goExamples) == 0 {
		goExamples = append([]string(nil), Default().GoExamples...)
	}

	return RuntimeConfig{
		Words:             append([]string(nil), cfg.NormalWords...),
		SpecialCharWords:  append([]string(nil), cfg.SpecialCharWords...),
		DurationOptions:   durationOptions,
		DurationLabels:    durationLabels,
		PromptWordCount:   cfg.PromptWordCount,
		QuoteEndpoint:     quoteEndpoint,
		GoExampleEndpoint: goExampleEndpoint,
		GoExamples:        goExamples,
	}, nil
}

func Load(path string) (RuntimeConfig, error) {
	cfg := Default()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return Resolve(cfg)
		}
		return RuntimeConfig{}, fmt.Errorf("read config %s: %w", path, err)
	}
	if err := json.Unmarshal(data, &cfg); err != nil {
		return RuntimeConfig{}, fmt.Errorf("parse config %s: %w", path, err)
	}
	return Resolve(cfg)
}
