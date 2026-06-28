package tui

import (
	"bytes"
	"io"
	"testing"

	"github.com/muesli/termenv"
)

func TestSkipBackgroundsWithoutTrueColor(t *testing.T) {
	oldDetect := colorProfileFor
	oldTransparent := transparentMode
	oldTrueColor := terminalTrueColor
	defer func() {
		colorProfileFor = oldDetect
		transparentMode = oldTransparent
		terminalTrueColor = oldTrueColor
		initStyles()
	}()

	colorProfileFor = func(io.Writer) termenv.Profile { return termenv.ANSI256 }
	InitTerminalDisplay(bytes.NewBuffer(nil))
	ApplyTransparentMode(false)
	if !skipBackgrounds() {
		t.Fatal("backgrounds should be skipped without TrueColor even when transparent is off")
	}

	colorProfileFor = func(io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(bytes.NewBuffer(nil))
	ApplyTransparentMode(false)
	if skipBackgrounds() {
		t.Fatal("backgrounds should be enabled with TrueColor and transparent off")
	}

	ApplyTransparentMode(true)
	if !skipBackgrounds() {
		t.Fatal("transparent mode should skip backgrounds even with TrueColor")
	}
}

func TestSkipBackgroundsRespectsTransparentMode(t *testing.T) {
	oldDetect := colorProfileFor
	oldTransparent := transparentMode
	defer func() {
		colorProfileFor = oldDetect
		transparentMode = oldTransparent
		initStyles()
	}()

	colorProfileFor = func(io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(bytes.NewBuffer(nil))
	transparentMode = true
	initStyles()
	if !skipBackgrounds() {
		t.Fatal("transparent mode should skip backgrounds")
	}
}
