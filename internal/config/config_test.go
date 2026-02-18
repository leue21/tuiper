package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestResolveValid(t *testing.T) {
	rc, err := Resolve(AppConfig{
		NormalWords:      []string{"a", "b"},
		SpecialCharWords: []string{"!@#", "$%^"},
		Durations:        []string{"15s", "1m"},
		PromptWordCount:  5,
	})
	if err != nil {
		t.Fatalf("Resolve returned error: %v", err)
	}
	if len(rc.DurationOptions) != 2 {
		t.Fatalf("len(DurationOptions) = %d, want 2", len(rc.DurationOptions))
	}
}

func TestResolveInvalidDuration(t *testing.T) {
	_, err := Resolve(AppConfig{
		NormalWords:      []string{"a"},
		SpecialCharWords: []string{"!"},
		Durations:        []string{"bad"},
		PromptWordCount:  1,
	})
	if err == nil {
		t.Fatal("expected error for invalid duration")
	}
}

func TestLoadMissingUsesDefaults(t *testing.T) {
	rc, err := Load(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if len(rc.Words) == 0 || len(rc.DurationOptions) == 0 {
		t.Fatal("expected defaults from missing config")
	}
}

func TestLoadFromFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "tuiper.json")
	data := `{
		"normal_words": ["foo","bar"],
		"special_char_words": ["!@#","$%^"],
		"durations": ["20s","1m"],
		"prompt_word_count": 7
	}`
	if err := os.WriteFile(path, []byte(data), 0o644); err != nil {
		t.Fatalf("write config: %v", err)
	}
	rc, err := Load(path)
	if err != nil {
		t.Fatalf("Load returned error: %v", err)
	}
	if rc.PromptWordCount != 7 {
		t.Fatalf("PromptWordCount = %d, want 7", rc.PromptWordCount)
	}
	if rc.DurationLabels[0] != "20s" {
		t.Fatalf("DurationLabels[0] = %q, want %q", rc.DurationLabels[0], "20s")
	}
}
