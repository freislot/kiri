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

func TestPlantGridRowBorderAlignment(t *testing.T) {
	oldDetect := colorProfileFor
	defer func() { colorProfileFor = oldDetect; initStyles() }()
	colorProfileFor = func(_ io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(nil)

	m := Model{
		width: 120, height: 40,
		lang: i18n.EN,
		plants: []model.Plant{
			{ID: 1, Name: "Monstera", Location: "living", State: model.StateWaterNow, WaterLevel: 20, BaseIntervalDays: 7},
			{ID: 2, Name: "Basil", Location: "kitchen", State: model.StateNormal, WaterLevel: 90, BaseIntervalDays: 3},
			{ID: 3, Name: "Fern", Location: "bath", State: model.StateShiftedByRain, WaterLevel: 70, BaseIntervalDays: 5},
		},
	}
	out := m.renderGrid()
	for i, line := range strings.Split(out, "\n") {
		w := styledTermWidth(line)
		want := m.plantGridWidth()
		if w != want {
			t.Fatalf("grid row line %d width=%d want=%d\nplain=%q", i, w, want, ansi.Strip(line))
		}
	}
}
