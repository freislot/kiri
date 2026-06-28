package tui

import (
	"math"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

const (
	splashRevealDuration = 1500 * time.Millisecond
	splashRevealInterval = 40 * time.Millisecond
)

type splashRevealTickMsg struct{}

func splashRevealTick() tea.Cmd {
	return tea.Tick(splashRevealInterval, func(time.Time) tea.Msg {
		return splashRevealTickMsg{}
	})
}

func splashRevealEase(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	if t < 0.5 {
		return 4 * t * t * t
	}
	return 1 - math.Pow(-2*t+2, 3)/2
}

func (m Model) styleFogLayer(content string, alpha float64) string {
	fg := splashColor(colorBorderGhost, alpha)
	s := lipgloss.NewStyle().Foreground(fg)
	if !skipBackgrounds() {
		s = s.Background(colorBg)
	}
	return s.Render(content)
}

func (m Model) renderSplashReveal() string {
	logoAlpha := splashRevealEase(m.splashRevealProg)
	w, h := m.termWidth(), m.termHeight()
	fog := m.renderSplashFog(w, h, 1)
	if logoAlpha <= 0 {
		return fog
	}
	block := m.renderSplashBlock(logoAlpha)
	if block == "" {
		return fog
	}
	return overlayCompose(fog, block, w, h)
}

func (m *Model) startSplashShow() {
	m.splashPhase = splashPhaseShow
	m.splashRevealProg = 1
	m.splashAlpha = 1
}
