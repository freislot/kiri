package tui

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type splashPhase int

const (
	splashPhaseReveal splashPhase = iota
	splashPhaseShow
	splashPhaseFade
	splashPhaseDone
)

const (
	splashHoldDuration   = 1 * time.Second
	splashFadeDuration   = 700 * time.Millisecond
	splashFadeInterval   = 40 * time.Millisecond
	splashLogoWidth      = 19
	splashTextGap        = 3
	splashFadeMinOverlay = 0.12
)

var (
	splashLogoLines = []string{
		" _     _       _    ",
		"| |   (_)     (_)   ",
		"| |  _ _  ____ _    ",
		"| |_/ ) |/ ___) |   ",
		"|  _ (| | |   | |   ",
		"|_| \\_)_|_|   |_|   ",
	}
)

type splashHoldDoneMsg struct{}
type splashFadeTickMsg struct{}

func splashHoldTimer() tea.Cmd {
	return tea.Tick(splashHoldDuration, func(time.Time) tea.Msg {
		return splashHoldDoneMsg{}
	})
}

func splashFadeTick() tea.Cmd {
	return tea.Tick(splashFadeInterval, func(time.Time) tea.Msg {
		return splashFadeTickMsg{}
	})
}

func (m *Model) dismissSplash() {
	m.splashPhase = splashPhaseDone
	m.splashAlpha = 0
	m.splashFog = nil
}

func (m *Model) startSplashFade() {
	m.splashPhase = splashPhaseFade
	m.splashAlpha = 1
	m.splashFadeProg = 0
}

func splashColor(c lipgloss.Color, alpha float64) lipgloss.Color {
	if alpha >= 1 {
		return c
	}
	if alpha <= 0 {
		return colorBg
	}
	return blendHexColors(c, colorBg, alpha)
}

func blendHexColors(fg, bg lipgloss.Color, alpha float64) lipgloss.Color {
	fr, fG, fb := parseHexColor(fg)
	br, bG, bb := parseHexColor(bg)
	r := uint8(math.Round(float64(fr)*alpha + float64(br)*(1-alpha)))
	g := uint8(math.Round(float64(fG)*alpha + float64(bG)*(1-alpha)))
	b := uint8(math.Round(float64(fb)*alpha + float64(bb)*(1-alpha)))
	return lipgloss.Color(fmt.Sprintf("#%02x%02x%02x", r, g, b))
}

func parseHexColor(c lipgloss.Color) (r, g, b uint8) {
	s := strings.TrimPrefix(string(c), "#")
	if len(s) != 6 {
		return 0, 0, 0
	}
	rv, _ := strconv.ParseUint(s[0:2], 16, 8)
	gv, _ := strconv.ParseUint(s[2:4], 16, 8)
	bv, _ := strconv.ParseUint(s[4:6], 16, 8)
	return uint8(rv), uint8(gv), uint8(bv)
}

func splashEaseOut(t float64) float64 {
	if t <= 0 {
		return 0
	}
	if t >= 1 {
		return 1
	}
	return 1 - (1-t)*(1-t)
}

func splashTextStyle(fg lipgloss.Color, alpha float64, bold bool) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(splashColor(fg, alpha))
	if bold {
		s = s.Bold(true)
	}
	if !skipBackgrounds() {
		s = s.Background(colorBg)
	}
	return s
}

func (m Model) splashBlockPlainWidth() int {
	sideText := m.splashSideText()
	width := 0
	for i, logo := range splashLogoLines {
		padded := logo
		if len(padded) < splashLogoWidth {
			padded += strings.Repeat(" ", splashLogoWidth-len(padded))
		}
		line := padded
		if side, ok := sideText[i]; ok {
			line += strings.Repeat(" ", splashTextGap) + side
		}
		if w := lipgloss.Width(line); w > width {
			width = w
		}
	}
	return width
}

func splashSpaceStyle() lipgloss.Style {
	s := lipgloss.NewStyle()
	if !skipBackgrounds() {
		s = s.Background(colorBg)
	}
	return s
}

func splashOverlayTextStyle(fg, bg lipgloss.Color, bold bool) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(fg)
	if !skipBackgrounds() {
		s = s.Background(bg)
	}
	if bold {
		s = s.Bold(true)
	}
	return s
}

func aboutBlockContainerStyle() lipgloss.Style {
	s := lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorBorderGhost).
		Padding(1, 2).
		Align(lipgloss.Left)
	if !skipBackgrounds() {
		s = s.Background(colorFogBg).BorderBackground(colorFogBg)
	}
	return s
}

func (m Model) splashSideText() map[int]string {
	return map[int]string{
		3: m.cat().SplashPlantStatusBoard(),
	}
}

func aboutTextStyle(fg lipgloss.Color, bold bool) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(fg)
	if bold {
		s = s.Bold(true)
	}
	if !skipBackgrounds() {
		s = s.Background(colorFogBg) // Используем фон контейнера, чтобы текст не просвечивал терминал
	}
	return s
}

func (m Model) renderAboutInfoBlock() string {
	lines := m.cat().AboutInfoLines()
	logStyle := aboutTextStyle(colorFg, false)
	versionStyle := aboutTextStyle(colorGreen, true)

	rendered := make([]string, len(lines))
	for i, line := range lines {
		if line == "" {
			rendered[i] = " "
			continue
		}
		if i == 0 {
			rendered[i] = versionStyle.Render(line)
		} else {
			rendered[i] = logStyle.Render(line)
		}
	}
	return strings.Join(rendered, "\n")
}

func (m Model) renderAboutSplashBlock() string {
	logoStyle := aboutTextStyle(colorGreen, true)
	logoLines := make([]string, len(splashLogoLines))
	for i, line := range splashLogoLines {
		padded := line
		if len(padded) < splashLogoWidth {
			padded += strings.Repeat(" ", splashLogoWidth-len(padded))
		}
		logoLines[i] = logoStyle.Render(padded)
	}
	logoBlock := strings.Join(logoLines, "\n")

	infoBlock := m.renderAboutInfoBlock()

	columnStyle := lipgloss.NewStyle()
	if !skipBackgrounds() {
		columnStyle = columnStyle.Background(colorFogBg)
	}

	maxHeight := max(lipgloss.Height(logoBlock), lipgloss.Height(infoBlock))

	logoColumn := columnStyle.Height(maxHeight).Render(logoBlock)
	infoColumn := columnStyle.Height(maxHeight).Render(infoBlock)

	gapStr := strings.Repeat(" ", splashTextGap)
	gapColumn := columnStyle.Height(maxHeight).Render(gapStr)

	content := lipgloss.JoinHorizontal(lipgloss.Top, logoColumn, gapColumn, infoColumn)

	return aboutBlockContainerStyle().Render(content)
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func (m Model) renderSplashBlock(alpha float64) string {
	if alpha <= 0 {
		return ""
	}

	logoStyle := splashTextStyle(colorGreen, alpha, true)
	mutedStyle := splashTextStyle(colorMuted, alpha, true)
	gapStyle := splashSpaceStyle()
	sideText := m.splashSideText()

	lines := make([]string, len(splashLogoLines))
	for i := range splashLogoLines {
		padded := splashLogoLines[i]
		if len(padded) < splashLogoWidth {
			padded += strings.Repeat(" ", splashLogoWidth-len(padded))
		}
		styled := logoStyle.Render(padded)
		if side, ok := sideText[i]; ok {
			styled += gapStyle.Render(strings.Repeat(" ", splashTextGap))
			styled += mutedStyle.Render(side)
		}
		lines[i] = styled
	}

	block := strings.Join(lines, "\n")
	container := lipgloss.NewStyle().
		Width(m.splashBlockPlainWidth()).
		Align(lipgloss.Left)
	if !skipBackgrounds() {
		container = container.Background(colorBg)
	}
	return container.Render(block)
}

func (m Model) renderSplashScreen() string {
	return m.renderSplashScene(1)
}

func (m Model) renderSplashFade() string {
	progress := splashEaseOut(m.splashFadeProg)
	splashAlpha := 1 - progress
	w, h := m.termWidth(), m.termHeight()

	if splashAlpha <= splashFadeMinOverlay {
		return m.renderMainApp()
	}

	if splashAlpha > 0.45 {
		return m.renderSplashScene(splashAlpha)
	}

	main := m.renderMainApp()
	block := m.renderSplashBlock(splashAlpha)
	if block == "" {
		return main
	}
	fog := m.renderSplashFog(w, h, splashAlpha*0.35)
	base := overlayCompose(main, fog, w, h)
	return overlayCompose(base, block, w, h)
}
