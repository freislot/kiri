package tui

import (
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	splashFogFPS           = 30
	splashFogTickDur       = time.Second / splashFogFPS
	splashFogMinCount      = 30
	splashFogMaxCount      = 80
	splashFogMinSpeedCPS   = 3.0
	splashFogSpeedRangeCPS = 9.0
	splashFogBackLayerPart = 0.65
	splashFogBackSpeedMul  = 0.45
	splashFogFrontSpeedMul = 1.05
)

var (
	splashFogBackChars  = []rune{'.', ','}
	splashFogFrontChars = []rune{'~', '*'}
)

type splashParticle struct {
	x, y   float64
	speed  float64
	driftY float64
	char   rune
}

type splashFogTickMsg struct{}

func splashFogTick() tea.Cmd {
	return tea.Tick(splashFogTickDur, func(time.Time) tea.Msg {
		return splashFogTickMsg{}
	})
}

func (m *Model) splashFogCount() int {
	w, h := m.termWidth(), m.termHeight()
	if w < 1 || h < 1 {
		return splashFogMinCount
	}
	n := w * h / 50
	if n < splashFogMinCount {
		return splashFogMinCount
	}
	if n > splashFogMaxCount {
		return splashFogMaxCount
	}
	return n
}

func (m *Model) initSplashFog() {
	w, h := m.termWidth(), m.termHeight()
	if w < 1 || h < 1 {
		return
	}
	n := m.splashFogCount()
	m.splashFog = make([]splashParticle, n)
	wf := float64(w)
	hf := float64(h)
	backCount := int(float64(n) * splashFogBackLayerPart)
	for i := 0; i < backCount; i++ {
		m.splashFog[i] = splashParticle{
			x:      rand.Float64() * wf,
			y:      rand.Float64() * hf,
			speed:  (splashFogMinSpeedCPS + rand.Float64()*splashFogSpeedRangeCPS) * splashFogBackSpeedMul,
			driftY: 0.05 + rand.Float64()*0.05,
			char:   splashFogBackChars[rand.Intn(len(splashFogBackChars))],
		}
	}
	for i := backCount; i < n; i++ {
		m.splashFog[i] = splashParticle{
			x:      rand.Float64() * wf,
			y:      rand.Float64() * hf,
			speed:  (splashFogMinSpeedCPS + rand.Float64()*splashFogSpeedRangeCPS) * splashFogFrontSpeedMul,
			driftY: 0.10 + rand.Float64()*0.08,
			char:   splashFogFrontChars[rand.Intn(len(splashFogFrontChars))],
		}
	}
}

func (m *Model) updateSplashFog() {
	w := float64(m.termWidth())
	h := float64(m.termHeight())
	dt := splashFogTickDur.Seconds()
	if w < 1 || h < 1 || len(m.splashFog) == 0 {
		return
	}
	for i := range m.splashFog {
		p := &m.splashFog[i]
		step := p.speed * dt
		p.x -= step
		p.y += step * p.driftY
		if p.x < 0 {
			p.x += w
			p.y = rand.Float64() * h
		}
		if p.y >= h {
			p.y -= h
			p.x = rand.Float64() * w
		}
	}
}

func (m Model) renderSplashFog(w, h int, alpha float64) string {
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}

	rows := make([][]rune, h)
	for y := 0; y < h; y++ {
		row := make([]rune, w)
		for x := 0; x < w; x++ {
			row[x] = ' '
		}
		rows[y] = row
	}

	for _, p := range m.splashFog {
		px := int(p.x)
		py := int(p.y)
		if px < 0 || px >= w || py < 0 || py >= h {
			continue
		}
		rows[py][px] = p.char
	}

	var b strings.Builder
	b.Grow(w*h + h)
	for y := 0; y < h; y++ {
		b.WriteString(string(rows[y]))
		if y < h-1 {
			b.WriteByte('\n')
		}
	}

	return m.styleFogLayer(b.String(), alpha)
}

func (m Model) renderSplashScene(logoAlpha float64) string {
	w, h := m.termWidth(), m.termHeight()
	fog := m.renderSplashFog(w, h, logoAlpha)
	block := m.renderSplashBlock(logoAlpha)
	if block == "" {
		return fog
	}
	return overlayCompose(fog, block, w, h)
}
