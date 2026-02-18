package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"tuitype/internal/config"
	"tuitype/internal/prompt"
)

const appName = "TUIper"

//go:embed docs/tuiper.1
var manPage string

var splashArt = strings.Join([]string{
	" _______ _   _ ___ ____  _____ ____  ",
	"|_   _| | | | |_ _|  _ \\| ____|  _ \\ ",
	"  | | | | | | || || |_) |  _| | |_) |",
	"  | | | |_| | || ||  __/| |___|  _ < ",
	"  |_|  \\___/ |___|_|   |_____|_| \\_\\",
}, "\n")

type model struct {
	cfg        config.RuntimeConfig
	prompts    *prompt.Service
	modeLabels []string

	width           int
	height          int
	prompt          string
	inputRunes      []rune
	totalTyped      int
	totalCorrect    int
	sessionDuration time.Duration
	startedAt       time.Time
	finishedAt      time.Time
	selectedMode    prompt.Mode
	selectedOption  int
	showSplash      bool
	selectingMode   bool
	selectingTime   bool
	started         bool
	done            bool
}

type tickMsg time.Time

func tickCmd() tea.Cmd {
	return tea.Tick(100*time.Millisecond, func(t time.Time) tea.Msg { return tickMsg(t) })
}

func initialModel(cfg config.RuntimeConfig) model {
	selected := 0
	defaultDuration := 30 * time.Second
	for i, d := range cfg.DurationOptions {
		if d == defaultDuration {
			selected = i
			break
		}
	}
	if selected >= len(cfg.DurationOptions) {
		selected = 0
	}
	return model{
		cfg: cfg,
		prompts: prompt.New(prompt.Config{
			Words:             cfg.Words,
			SpecialCharWords:  cfg.SpecialCharWords,
			PromptWordCount:   cfg.PromptWordCount,
			QuoteEndpoint:     cfg.QuoteEndpoint,
			GoExampleEndpoint: cfg.GoExampleEndpoint,
			GoExamples:        cfg.GoExamples,
		}),
		modeLabels:      prompt.ModeLabels(),
		sessionDuration: cfg.DurationOptions[selected],
		selectedOption:  selected,
		selectedMode:    prompt.ModeNormal,
		showSplash:      true,
	}
}

func (m *model) resetSession() {
	m.prompt = m.prompts.Next(m.selectedMode, "")
	m.inputRunes = nil
	m.totalTyped = 0
	m.totalCorrect = 0
	m.startedAt = time.Time{}
	m.finishedAt = time.Time{}
	m.started = false
	m.done = false
}

func pickIndexFromKey(key string, max int) (int, bool) {
	if len(key) != 1 {
		return 0, false
	}
	n, err := strconv.Atoi(key)
	if err != nil || n < 1 || n > max {
		return 0, false
	}
	return n - 1, true
}

func quickPickHint(max int) string {
	if max <= 1 {
		return "1"
	}
	return fmt.Sprintf("1-%d", max)
}

func defaultConfigPath() string {
	if dir, err := os.UserConfigDir(); err == nil && dir != "" {
		return filepath.Join(dir, "tuiper", "config.json")
	}
	return "tuiper.json"
}

func (m model) Init() tea.Cmd { return tickCmd() }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case tickMsg:
		if m.started && !m.done && time.Since(m.startedAt) >= m.sessionDuration {
			m.done = true
			m.finishedAt = time.Now()
		}
		return m, tickCmd()
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}

		if m.showSplash {
			if msg.String() == "enter" {
				m.showSplash = false
				m.selectingMode = true
			}
			return m, nil
		}

		if m.selectingMode {
			switch msg.String() {
			case "left", "up":
				m.selectedMode--
				if m.selectedMode < 0 {
					m.selectedMode = prompt.Mode(len(m.modeLabels) - 1)
				}
			case "right", "down":
				m.selectedMode++
				if int(m.selectedMode) >= len(m.modeLabels) {
					m.selectedMode = 0
				}
			case "enter":
				m.selectingMode = false
				m.selectingTime = true
			default:
				if idx, ok := pickIndexFromKey(msg.String(), len(m.modeLabels)); ok {
					m.selectedMode = prompt.Mode(idx)
				}
			}
			return m, nil
		}

		if m.selectingTime {
			switch msg.String() {
			case "left", "up":
				m.selectedOption--
				if m.selectedOption < 0 {
					m.selectedOption = len(m.cfg.DurationOptions) - 1
				}
			case "right", "down":
				m.selectedOption++
				if m.selectedOption >= len(m.cfg.DurationOptions) {
					m.selectedOption = 0
				}
			case "enter":
				m.sessionDuration = m.cfg.DurationOptions[m.selectedOption]
				m.selectingTime = false
				m.resetSession()
			default:
				if idx, ok := pickIndexFromKey(msg.String(), len(m.cfg.DurationOptions)); ok {
					m.selectedOption = idx
				}
			}
			return m, nil
		}

		if m.done {
			if msg.String() == "enter" {
				m.selectingMode = true
			}
			return m, nil
		}

		switch msg.String() {
		case "backspace":
			if len(m.inputRunes) > 0 {
				idx := len(m.inputRunes) - 1
				r := m.inputRunes[idx]
				promptRunes := []rune(m.prompt)
				if idx < len(promptRunes) && promptRunes[idx] == r && m.totalCorrect > 0 {
					m.totalCorrect--
				}
				if m.totalTyped > 0 {
					m.totalTyped--
				}
				m.inputRunes = m.inputRunes[:len(m.inputRunes)-1]
			}
		default:
			if len(msg.Runes) > 0 {
				if !m.started {
					m.started = true
					m.startedAt = time.Now()
				}
				for _, r := range msg.Runes {
					promptRunes := []rune(m.prompt)
					if len(m.inputRunes) >= len(promptRunes) {
						m.prompt = m.prompts.Next(m.selectedMode, m.prompt)
						m.inputRunes = m.inputRunes[:0]
						promptRunes = []rune(m.prompt)
					}
					idx := len(m.inputRunes)
					m.totalTyped++
					if idx < len(promptRunes) && r == promptRunes[idx] {
						m.inputRunes = append(m.inputRunes, r)
						m.totalCorrect++
						continue
					}
					if idx > 0 {
						prevIdx := idx - 1
						if prevIdx < len(promptRunes) &&
							m.inputRunes[prevIdx] != promptRunes[prevIdx] &&
							r == promptRunes[prevIdx] {
							// If the user immediately corrects the previously mistyped
							// character, repair that slot instead of shifting everything.
							m.inputRunes[prevIdx] = r
							m.totalCorrect++
							continue
						}
					}
					m.inputRunes = append(m.inputRunes, r)
				}
				if (m.selectedMode == prompt.ModeQuote || m.selectedMode == prompt.ModeCode) && len(m.inputRunes) >= len([]rune(m.prompt)) {
					m.prompt = m.prompts.Next(m.selectedMode, m.prompt)
					m.inputRunes = m.inputRunes[:0]
				}
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.width == 0 || m.height == 0 {
		return "loading..."
	}
	if m.width < 24 || m.height < 10 {
		return "Terminal too small. Resize to at least 24x10."
	}

	base := lipgloss.AdaptiveColor{Light: "#4c4f69", Dark: "#cdd6f4"}
	muted := lipgloss.AdaptiveColor{Light: "#6c6f85", Dark: "#a6adc8"}
	accent := lipgloss.AdaptiveColor{Light: "#df8e1d", Dark: "#f9e2af"}
	errorColor := lipgloss.AdaptiveColor{Light: "#d20f39", Dark: "#f38ba8"}
	surface := lipgloss.AdaptiveColor{Light: "#ccd0da", Dark: "#313244"}

	titleStyle := lipgloss.NewStyle().Bold(true).Foreground(accent)
	subtleStyle := lipgloss.NewStyle().Foreground(muted)
	correctStyle := lipgloss.NewStyle().Foreground(base)
	wrongStyle := lipgloss.NewStyle().Foreground(errorColor).Underline(true)
	pendingStyle := lipgloss.NewStyle().Foreground(muted)
	cursorStyle := lipgloss.NewStyle().Foreground(surface).Background(accent).Bold(true)
	selectedStyle := lipgloss.NewStyle().Foreground(surface).Background(accent).Bold(true).Padding(0, 1)
	cardStyle := lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(muted).Padding(1, 3)

	header := titleStyle.Render(appName)
	contentWidth := m.width - 6
	if contentWidth > 100 {
		contentWidth = 100
	}
	if contentWidth < 24 {
		contentWidth = 24
	}
	compact := contentWidth < 56 || m.height < 18
	renderCentered := func(content string) string {
		layout := lipgloss.NewStyle().Width(m.width).Height(m.height)
		if compact {
			layout = layout.Align(lipgloss.Left, lipgloss.Top).Padding(1, 1)
		} else {
			layout = layout.Align(lipgloss.Center, lipgloss.Center)
		}
		return layout.Render(content)
	}

	if m.showSplash {
		logo := titleStyle.Render(splashArt)
		if compact {
			logo = titleStyle.Render(appName)
		}
		body := strings.Join([]string{
			logo,
			"",
			lipgloss.NewStyle().Foreground(base).Bold(true).Render("Terminal UI typing trainer"),
			"",
			selectedStyle.Render("Enter to Continue"),
			"",
			subtleStyle.Render("ctrl+c quit"),
		}, "\n")
		return renderCentered(cardStyle.Width(contentWidth).Render(body))
	}

	if m.selectingMode {
		opts := make([]string, 0, len(m.modeLabels))
		for i, label := range m.modeLabels {
			s := subtleStyle.Render(label)
			if i == int(m.selectedMode) {
				s = selectedStyle.Render(label)
			}
			opts = append(opts, s)
		}
		line := strings.Join(opts, "    ")
		if compact {
			line = strings.Join(opts, "\n")
		}
		content := strings.Join([]string{
			header, titleStyle.Render("Select Mode"), "", line, "",
			selectedStyle.Render("Enter to Continue"), "",
			subtleStyle.Render("arrows or " + quickPickHint(len(m.modeLabels)) + " • ctrl+c quit"),
		}, "\n")
		return renderCentered(cardStyle.Width(contentWidth).Render(content))
	}

	if m.selectingTime {
		opts := make([]string, 0, len(m.cfg.DurationLabels))
		for i, label := range m.cfg.DurationLabels {
			s := subtleStyle.Render(label)
			if i == m.selectedOption {
				s = selectedStyle.Render(label)
			}
			opts = append(opts, s)
		}
		line := strings.Join(opts, "    ")
		if compact {
			line = strings.Join(opts, "\n")
		}
		content := strings.Join([]string{
			header, titleStyle.Render("Select Duration"), "", line, "",
			selectedStyle.Render("Enter to Start"), "",
			subtleStyle.Render("arrows or " + quickPickHint(len(m.cfg.DurationOptions)) + " • ctrl+c quit"),
		}, "\n")
		return renderCentered(cardStyle.Width(contentWidth).Render(content))
	}

	rPrompt := []rune(m.prompt)
	var b strings.Builder
	for i, r := range rPrompt {
		switch {
		case i < len(m.inputRunes):
			if m.inputRunes[i] == r {
				b.WriteString(correctStyle.Render(string(r)))
			} else {
				b.WriteString(wrongStyle.Render(string(r)))
			}
		case i == len(m.inputRunes) && !m.done:
			b.WriteString(cursorStyle.Render(string(r)))
		default:
			b.WriteString(pendingStyle.Render(string(r)))
		}
	}

	elapsed := time.Second
	if m.started {
		if m.done {
			elapsed = m.finishedAt.Sub(m.startedAt)
		} else {
			elapsed = time.Since(m.startedAt)
		}
		if elapsed <= 0 {
			elapsed = time.Second
		}
	}
	wpm := float64(m.totalCorrect) / 5.0 / elapsed.Minutes()
	accuracy := 100.0
	if m.totalTyped > 0 {
		accuracy = float64(m.totalCorrect) / float64(m.totalTyped) * 100.0
	}
	remaining := m.sessionDuration - elapsed
	if m.done || remaining < 0 {
		remaining = 0
	}
	stats := fmt.Sprintf("mode %s   wpm %.0f   acc %.1f%%   chars %d   time %.1fs",
		m.modeLabels[int(m.selectedMode)], wpm, accuracy, m.totalTyped, remaining.Seconds())
	if compact {
		stats = fmt.Sprintf("wpm %.0f  acc %.0f%%  t %.1fs", wpm, accuracy, remaining.Seconds())
	}
	footer := subtleStyle.Render("backspace edit • ctrl+c quit")
	if m.done {
		footer = subtleStyle.Render("enter menu • ctrl+c quit")
	}

	content := strings.Join([]string{
		header,
		cardStyle.Width(contentWidth).Render(titleStyle.Render(stats)),
		"",
		lipgloss.NewStyle().Width(contentWidth).Render(b.String()),
		"",
		lipgloss.NewStyle().Width(contentWidth).Render(footer),
	}, "\n")
	return renderCentered(content)
}

func main() {
	flag.Usage = func() {
		out := flag.CommandLine.Output()
		fmt.Fprintf(out, "%s - terminal typing trainer\n\n", strings.ToLower(appName))
		fmt.Fprintf(out, "Usage:\n  %s [options]\n\n", strings.ToLower(appName))
		fmt.Fprintln(out, "Options:")
		flag.PrintDefaults()
		fmt.Fprintln(out, "")
		fmt.Fprintln(out, "Config file keys (JSON):")
		fmt.Fprintln(out, `  "normal_words": ["word", ...]`)
		fmt.Fprintln(out, `  "special_char_words": ["!@#$", ...]`)
		fmt.Fprintln(out, `  "durations": ["15s", "30s", "1m", "2m"]`)
		fmt.Fprintln(out, `  "prompt_word_count": 18`)
		fmt.Fprintln(out, `  "quote_endpoint": "https://api.quotable.io/random?minLength=80&maxLength=220"`)
		fmt.Fprintln(out, `  "go_example_endpoint": ""  # empty disables remote go examples`)
		fmt.Fprintln(out, `  "go_examples": ["for i := 0; i < 3; i++ { fmt.Println(i) }", ...]`)
		fmt.Fprintln(out, "")
		fmt.Fprintf(out, "Man page:\n  %s -man\n", strings.ToLower(appName))
	}

	configPath := flag.String("config", defaultConfigPath(), "path to JSON config file")
	man := flag.Bool("man", false, "print the man page and exit")
	flag.Parse()

	if *man {
		fmt.Print(manPage)
		return
	}

	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "config error: %v\n", err)
		os.Exit(1)
	}

	p := tea.NewProgram(initialModel(cfg), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		panic(err)
	}
}
