package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
)

func styledTermWidth(s string) int {
	return runewidth.StringWidth(ansi.Strip(s))
}

func modalFillStyle() lipgloss.Style {
	return lipgloss.NewStyle().Background(colorCardBg)
}

func modalPad(n int) string {
	if n <= 0 {
		return ""
	}
	return modalFillStyle().Render(strings.Repeat(" ", n))
}

func modalPadLine(line string, tw int) string {
	if line == "" {
		return modalFillStyle().Width(tw).Render("")
	}
	if pad := tw - styledTermWidth(line); pad > 0 {
		line += modalPad(pad)
	}
	return line
}

func modalOpaque(style lipgloss.Style) lipgloss.Style {
	return style.Background(colorCardBg)
}

func joinModalLines(parts []string, tw int) string {
	var rows []string
	for _, part := range parts {
		for _, line := range strings.Split(part, "\n") {
			rows = append(rows, modalPadLine(line, tw))
		}
	}
	return strings.Join(rows, "\n")
}

func truncateTerm(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if runewidth.StringWidth(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	return runewidth.Truncate(s, max, "…")
}

func wrapBlockTerm(s string, maxW int) string {
	if maxW <= 0 || runewidth.StringWidth(s) <= maxW {
		return s
	}
	var lines []string
	var line string
	for _, word := range strings.Fields(s) {
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if runewidth.StringWidth(candidate) <= maxW {
			line = candidate
			continue
		}
		if line != "" {
			lines = append(lines, line)
			line = ""
		}
		if runewidth.StringWidth(word) <= maxW {
			line = word
		} else {
			lines = append(lines, runewidth.Truncate(word, maxW, "…"))
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func modalStyleWrapped(style lipgloss.Style, s string, tw int) string {
	wrapped := wrapBlockTerm(s, tw)
	lines := strings.Split(wrapped, "\n")
	for i, line := range lines {
		if line == "" {
			lines[i] = modalPadLine("", tw)
			continue
		}
		lines[i] = modalPadLine(modalOpaque(style).Render(line), tw)
	}
	return strings.Join(lines, "\n")
}
