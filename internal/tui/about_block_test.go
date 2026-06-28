package tui

import (
	"bytes"
	"io"
	"strings"
	"testing"

	"github.com/charmbracelet/lipgloss"
	"github.com/muesli/termenv"
)

func TestRenderAboutSplashBlockSolidBackground(t *testing.T) {
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

	block := Model{}.renderAboutSplashBlock()
	if !strings.Contains(block, "48;") {
		t.Fatal("about splash block should paint a solid background")
	}

	ApplyTransparentMode(true)
	transparent := Model{}.renderAboutSplashBlock()
	if strings.Contains(transparent, "48;") {
		t.Fatal("transparent mode should not paint about block backgrounds")
	}
}
