package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	aboutFogFPS      = 30
	aboutFogTickDur  = time.Second / aboutFogFPS
	aboutFogTimeStep = 0.03
)

type aboutFogTickMsg struct{}

func aboutFogTick() tea.Cmd {
	return tea.Tick(aboutFogTickDur, func(time.Time) tea.Msg {
		return aboutFogTickMsg{}
	})
}

func (m Model) aboutTabCmd() tea.Cmd {
	if m.tab == TabAbout {
		return aboutFogTick()
	}
	return nil
}

func (m Model) renderAboutScreen() string {
	w, h := m.termWidth(), m.termHeight()
	canvas := renderProceduralFog(w, h, m.aboutFogTime)

	block := m.renderAboutSplashBlock()
	if block != "" {
		blockLines := strings.Split(block, "\n")
		blockH := len(blockLines)
		maxBlockW := 0
		for _, line := range blockLines {
			if lw := lipgloss.Width(line); lw > maxBlockW {
				maxBlockW = lw
			}
		}
		startX := (w - maxBlockW) / 2
		startY := (h - blockH) / 2
		if startX < 0 {
			startX = 0
		}
		if startY < 0 {
			startY = 0
		}
		canvas = overlayAt(canvas, block, startX, startY, w, h)
	}

	hint := m.renderAboutFooterHint()
	if hint != "" {
		canvas = overlayAt(canvas, hint, 0, h-1, w, h)
	}

	return canvas
}

func (m Model) renderAboutFooterHint() string {
	maxW := m.termWidth()
	if maxW < 1 {
		return ""
	}
	return renderFooterHints(m.cat().FooterAboutHints(), maxW)
}
