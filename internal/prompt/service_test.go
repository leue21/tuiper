package prompt

import (
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func testService() *Service {
	return New(Config{
		Words:             []string{"alpha", "beta", "gamma"},
		SpecialCharWords:  []string{"!@#", "$%^"},
		PromptWordCount:   4,
		QuoteEndpoint:     "https://example.test/quote",
		GoExampleEndpoint: "",
		GoExamples:        []string{`fmt.Println("a")`, `fmt.Println("b")`},
	})
}

func TestNextNormalNonEmpty(t *testing.T) {
	s := testService()
	got := s.Next(ModeNormal, "")
	if got == "" {
		t.Fatal("expected non-empty prompt")
	}
}

func TestQuoteRetriesTransientError(t *testing.T) {
	s := testService()
	var calls int32
	s.client = &http.Client{
		Timeout: time.Second,
		Transport: roundTripFunc(func(req *http.Request) (*http.Response, error) {
			n := atomic.AddInt32(&calls, 1)
			if n < 2 {
				return nil, io.EOF
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader(`{"content":"recovered quote"}`)),
				Header:     make(http.Header),
			}, nil
		}),
	}
	got := s.Next(ModeQuote, "old")
	if got != "recovered quote" {
		t.Fatalf("got %q, want %q", got, "recovered quote")
	}
}

func TestCodeModeFallbackWhenEndpointDisabled(t *testing.T) {
	s := testService()
	got := s.Next(ModeCode, `fmt.Println("a")`)
	if got == `fmt.Println("a")` {
		t.Fatalf("expected rotated fallback, got %q", got)
	}
}

func TestCleanGoTypingPromptStripsHeaders(t *testing.T) {
	raw := `// Copyright 2026
package main
import "fmt"
func main() {
	fmt.Println("hello")
}`
	got := cleanGoTypingPrompt(raw)
	want := `func main() { fmt.Println("hello") }`
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

type roundTripFunc func(req *http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req)
}
