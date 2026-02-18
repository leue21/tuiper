package main

import "testing"

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
