package main

import (
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestPickIndexFromKey(t *testing.T) {
	if idx, ok := pickIndexFromKey("2", 4); !ok || idx != 1 {
		t.Fatalf("pickIndexFromKey(2,4) = (%d,%v), want (1,true)", idx, ok)
	}
	if _, ok := pickIndexFromKey("x", 4); ok {
		t.Fatal("expected invalid for non-numeric key")
	}
}

func TestQuickPickHint(t *testing.T) {
	if got := quickPickHint(1); got != "1" {
		t.Fatalf("quickPickHint(1) = %q, want %q", got, "1")
	}
	if got := quickPickHint(4); got != "1-4" {
		t.Fatalf("quickPickHint(4) = %q, want %q", got, "1-4")
	}
}

func TestMistypeThenImmediateCorrectionRepairsPreviousSlot(t *testing.T) {
	m := model{prompt: "ab"}

	updated, _ := m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'x'}})
	m = updated.(model)
	if got := string(m.inputRunes); got != "x" {
		t.Fatalf("after wrong key, inputRunes = %q, want %q", got, "x")
	}
	if m.totalTyped != 1 {
		t.Fatalf("after wrong key, totalTyped = %d, want 1", m.totalTyped)
	}
	if m.totalCorrect != 0 {
		t.Fatalf("after wrong key, totalCorrect = %d, want 0", m.totalCorrect)
	}

	updated, _ = m.Update(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}})
	m = updated.(model)
	if got := string(m.inputRunes); got != "a" {
		t.Fatalf("after correction, inputRunes = %q, want %q", got, "a")
	}
	if m.totalTyped != 2 {
		t.Fatalf("after correction, totalTyped = %d, want 2", m.totalTyped)
	}
	if m.totalCorrect != 1 {
		t.Fatalf("after correction, totalCorrect = %d, want 1", m.totalCorrect)
	}
}
