package prompt

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

type Mode int

const (
	ModeNormal Mode = iota
	ModeSpecialChars
	ModeQuote
	ModeCode
)

var modeLabels = []string{
	"Normal",
	"Special Chars Practice",
	"Quote Practice",
	"Code Practice",
}

func ModeLabels() []string {
	return append([]string(nil), modeLabels...)
}

type Config struct {
	Words             []string
	SpecialCharWords  []string
	PromptWordCount   int
	QuoteEndpoint     string
	GoExampleEndpoint string
	GoExamples        []string
}

type Service struct {
	cfg               Config
	client            *http.Client
	rng               *rand.Rand
	quoteBackoffUntil time.Time
	codeBackoffUntil  time.Time
	fallbackQuotes    []string
}

func New(cfg Config) *Service {
	return &Service{
		cfg: Config{
			Words:             append([]string(nil), cfg.Words...),
			SpecialCharWords:  append([]string(nil), cfg.SpecialCharWords...),
			PromptWordCount:   cfg.PromptWordCount,
			QuoteEndpoint:     cfg.QuoteEndpoint,
			GoExampleEndpoint: cfg.GoExampleEndpoint,
			GoExamples:        append([]string(nil), cfg.GoExamples...),
		},
		client: &http.Client{Timeout: 1200 * time.Millisecond},
		rng:    rand.New(rand.NewSource(time.Now().UnixNano())),
		fallbackQuotes: []string{
			"Type with calm precision and let rhythm do the heavy lifting.",
			"Progress in typing is consistency repeated over short focused sessions.",
			"Accuracy builds speed; speed without accuracy always stalls.",
		},
	}
}

func (s *Service) Next(mode Mode, previous string) string {
	switch mode {
	case ModeQuote:
		return s.nextQuote(previous)
	case ModeCode:
		return s.nextCode(previous)
	case ModeSpecialChars:
		return s.nextFromWords(previous, s.cfg.SpecialCharWords, "!@#$ %^&* ()_+ []{} <>? /\\| `~ ;;:: ++--.")
	default:
		return s.nextFromWords(previous, s.cfg.Words, "the quick brown fox jumps over the lazy dog.")
	}
}

func (s *Service) nextFromWords(previous string, words []string, fallback string) string {
	for i := 0; i < 8; i++ {
		buf := make([]string, s.cfg.PromptWordCount)
		for j := range buf {
			buf[j] = words[s.rng.Intn(len(words))]
		}
		p := strings.Join(buf, " ") + "."
		if p != previous {
			return p
		}
	}
	return fallback
}

func pickDifferent(rng *rand.Rand, options []string, previous, emptyFallback string) string {
	if len(options) == 0 {
		return emptyFallback
	}
	if len(options) == 1 {
		return options[0]
	}
	for i := 0; i < 8; i++ {
		c := options[rng.Intn(len(options))]
		if c != previous {
			return c
		}
	}
	for _, c := range options {
		if c != previous {
			return c
		}
	}
	return options[0]
}

type quoteResponse struct {
	Content string `json:"content"`
	Quote   string `json:"quote"`
	Text    string `json:"text"`
}

func (s *Service) fetchQuote(endpoint string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("quote endpoint is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1200*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("quote API status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}
	var payload quoteResponse
	if err := json.Unmarshal(body, &payload); err != nil {
		return "", err
	}
	for _, c := range []string{strings.TrimSpace(payload.Content), strings.TrimSpace(payload.Quote), strings.TrimSpace(payload.Text)} {
		if c != "" {
			return c, nil
		}
	}
	return "", fmt.Errorf("quote API returned empty payload")
}

func (s *Service) nextQuote(previous string) string {
	pickFallback := func() string {
		return pickDifferent(s.rng, s.fallbackQuotes, previous, "keep typing with steady rhythm.")
	}
	if time.Now().Before(s.quoteBackoffUntil) {
		return pickFallback()
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		q, err := s.fetchQuote(s.cfg.QuoteEndpoint)
		if err != nil {
			lastErr = err
			continue
		}
		if q != previous {
			return q
		}
	}
	if lastErr != nil {
		s.quoteBackoffUntil = time.Now().Add(15 * time.Second)
	}
	return pickFallback()
}

type codeResponse struct {
	Content string `json:"content"`
	Code    string `json:"code"`
	Text    string `json:"text"`
}

func cleanGoTypingPrompt(raw string) string {
	lines := strings.Split(strings.ReplaceAll(raw, "\r\n", "\n"), "\n")
	codeLines := make([]string, 0, len(lines))
	inBlockComment := false
	inImportBlock := false

	for _, line := range lines {
		l := strings.TrimSpace(line)
		if l == "" {
			continue
		}
		if strings.HasPrefix(l, "/*") {
			inBlockComment = true
		}
		if inBlockComment {
			if strings.Contains(l, "*/") {
				inBlockComment = false
			}
			continue
		}
		if strings.HasPrefix(l, "//") || strings.HasPrefix(l, "package ") {
			continue
		}
		if strings.HasPrefix(l, "import (") {
			inImportBlock = true
			continue
		}
		if inImportBlock {
			if l == ")" {
				inImportBlock = false
			}
			continue
		}
		if strings.HasPrefix(l, "import ") {
			continue
		}
		codeLines = append(codeLines, l)
	}

	if len(codeLines) == 0 {
		return ""
	}
	out := strings.Join(strings.Fields(strings.Join(codeLines, " ")), " ")
	if len(out) > 260 {
		out = out[:260]
		if i := strings.LastIndex(out, " "); i > 80 {
			out = out[:i]
		}
	}
	return strings.TrimSpace(out)
}

func (s *Service) fetchCode(endpoint string) (string, error) {
	if endpoint == "" {
		return "", fmt.Errorf("go example endpoint is empty")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("Accept", "application/json, text/plain;q=0.9")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("go example API status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	if err != nil {
		return "", err
	}

	var payload codeResponse
	if err := json.Unmarshal(body, &payload); err == nil {
		for _, c := range []string{cleanGoTypingPrompt(payload.Content), cleanGoTypingPrompt(payload.Code), cleanGoTypingPrompt(payload.Text)} {
			if c != "" {
				return c, nil
			}
		}
	}

	plain := cleanGoTypingPrompt(string(body))
	if plain == "" {
		return "", fmt.Errorf("go example payload is empty")
	}
	return plain, nil
}

func (s *Service) nextCode(previous string) string {
	pickFallback := func() string {
		return pickDifferent(s.rng, s.cfg.GoExamples, previous, `fmt.Println("hello, tuiper")`)
	}
	if time.Now().Before(s.codeBackoffUntil) {
		return pickFallback()
	}
	if strings.TrimSpace(s.cfg.GoExampleEndpoint) == "" {
		return pickFallback()
	}

	var lastErr error
	for i := 0; i < 3; i++ {
		ex, err := s.fetchCode(s.cfg.GoExampleEndpoint)
		if err != nil {
			lastErr = err
			continue
		}
		if ex != previous {
			return ex
		}
	}
	if lastErr != nil {
		s.codeBackoffUntil = time.Now().Add(15 * time.Second)
	}
	return pickFallback()
}
