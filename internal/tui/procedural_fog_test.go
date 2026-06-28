package tui

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestRenderProceduralFogRespectsTransparentMode(t *testing.T) {
	oldDetect := colorProfileFor
	oldTransparent := transparentMode
	oldProfile := lipgloss.ColorProfile()
	defer func() {
		colorProfileFor = oldDetect
		transparentMode = oldTransparent
		lipgloss.SetColorProfile(oldProfile)
		initStyles()
	}()

	lipgloss.SetColorProfile(termenv.TrueColor)
	colorProfileFor = func(io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(bytes.NewBuffer(nil))

	ApplyTransparentMode(false)
	opaque := renderProceduralFog(4, 2, 0)
	if !strings.Contains(opaque, "48;") {
		t.Fatal("opaque mode should paint fog cells with terminal background color")
	}

	ApplyTransparentMode(true)
	transparent := renderProceduralFog(4, 2, 0)
	if strings.Contains(transparent, "48;") {
		t.Fatal("transparent mode should not paint cell backgrounds")
	}
}
