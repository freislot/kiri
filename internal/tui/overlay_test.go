package tui

import (
	"io"
	"strings"
	"testing"

	"kiri/internal/i18n"
	"kiri/internal/model"

	"github.com/charmbracelet/x/ansi"
	"github.com/muesli/termenv"
)

func TestAddPlantModalBorderColumnsWithEmoji(t *testing.T) {
	oldDetect := colorProfileFor
	defer func() { colorProfileFor = oldDetect; initStyles() }()
	colorProfileFor = func(_ io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(nil)

	m := Model{
		width: 120, height: 40,
		lang:    i18n.EN,
		overlay: OverlayAddPlant,
		plantForm: plantForm{
			focus:    plantFieldName,
			interval: "7",
		},
	}
	assertModalBordersAligned(t, m.renderOverlay())
}

func TestDeleteModalBorderColumnsWithEmoji(t *testing.T) {
	oldDetect := colorProfileFor
	defer func() { colorProfileFor = oldDetect; initStyles() }()
	colorProfileFor = func(_ io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(nil)

	m := Model{
		width: 120, height: 40,
		lang:    i18n.EN,
		overlay: OverlayDeleteConfirm,
		plants:  []model.Plant{{ID: 1, Name: "Monstera"}},
		cursor:  0,
	}
	assertModalBordersAligned(t, m.renderOverlay())
}

func TestOverlayCenterLineWidths(t *testing.T) {
	oldDetect := colorProfileFor
	defer func() { colorProfileFor = oldDetect; initStyles() }()
	colorProfileFor = func(_ io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(nil)

	m := Model{width: 120, height: 40, lang: i18n.EN, tab: TabPlants, overlay: OverlayAddPlant, plantForm: plantForm{interval: "7"}}
	out := overlayCenter(m.renderBaseApp(), m.renderOverlay(), m.termWidth(), m.termHeight())
	if len(strings.Split(out, "\n")) != 40 {
		t.Fatalf("overlay height = %d, want 40", len(strings.Split(out, "\n")))
	}
	for i, line := range strings.Split(out, "\n") {
		if w := styledTermWidth(line); w != 120 {
			t.Fatalf("line %d width=%d want 120", i, w)
		}
	}
}

func assertModalBordersAligned(t *testing.T, popup string) {
	t.Helper()
	var leftCol, rightCol int
	set := false
	for i, line := range strings.Split(popup, "\n") {
		l, r, ok := modalBorderCols(line)
		if !ok {
			continue
		}
		if !set {
			leftCol, rightCol = l, r
			set = true
			continue
		}
		if l != leftCol || r != rightCol {
			t.Fatalf("row %d borders (%d,%d) want (%d,%d) plain=%q",
				i, l, r, leftCol, rightCol, ansi.Strip(line))
		}
	}
}

func modalBorderCols(line string) (left, right int, ok bool) {
	w := styledTermWidth(line)
	for cell := 0; cell < w; cell++ {
		ch := ansi.Strip(ansi.Cut(line, cell, cell+1))
		if ch == "│" {
			if !ok {
				left = cell
				ok = true
			}
			right = cell
		}
	}
	return left, right, ok
}
