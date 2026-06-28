package tui

import (
	"math"
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var (
	proceduralFogDensity = []string{" ", ".", "~", "*", "░", "▒", "▓", "█"}
	proceduralFogColors  = []lipgloss.Color{
		"#2d353b",
		"#343f44",
		"#3d484d",
		"#475258",
		"#4f5b58",
		"#56635f",
		"#6c7a72",
		"#7fbbb3",
	}
	colorFogBg = lipgloss.Color("#2d353b")
)

func renderProceduralFog(w, h int, t float64) string {
	if w < 1 {
		w = 80
	}
	if h < 1 {
		h = 24
	}

	maxIndex := float64(len(proceduralFogDensity) - 1)

	var b strings.Builder
	b.Grow(w*h + h*20)

	for y := 0; y < h; y++ {
		ny := float64(y) * 0.1
		for x := 0; x < w; x++ {
			nx := float64(x) * 0.05

			noise := math.Sin(nx+t) * math.Cos(ny+t*0.5)
			noise += math.Sin(nx*0.5-t*0.7) * math.Cos(ny*0.8+t)
			noise += math.Sin(nx*2.0+ny) * 0.2

			normalized := (noise + 2.2) / 4.4
			if normalized < 0 {
				normalized = 0
			}
			if normalized > 1 {
				normalized = 1
			}

			idx := int(math.Round(normalized * maxIndex))
			if idx >= len(proceduralFogDensity) {
				idx = len(proceduralFogDensity) - 1
			}

			char := proceduralFogDensity[idx]
			color := proceduralFogColors[idx]
			style := lipgloss.NewStyle().Foreground(color)
			if !skipBackgrounds() {
				style = style.Background(colorBg)
			}
			b.WriteString(style.Render(char))
		}
		if y < h-1 {
			b.WriteByte('\n')
		}
	}

	return b.String()
}
