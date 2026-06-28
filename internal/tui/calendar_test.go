package tui

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"kiri/internal/i18n"
	"kiri/internal/model"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
	"github.com/mattn/go-runewidth"
	"github.com/muesli/termenv"
)

func TestCalDayIcon(t *testing.T) {
	if got := calDayIcon(true, true); got != calIconUrgent {
		t.Fatalf("urgent wins: got %q", got)
	}
	if got := calDayIcon(false, true); got != calIconDue {
		t.Fatalf("due: got %q", got)
	}
	if got := calDayIcon(false, false); got != calIconEmpty {
		t.Fatalf("empty: got %q", got)
	}
	if w := runewidth.StringWidth(calIconSlot(calIconUrgent)); w != calIconWidth {
		t.Fatalf("urgent slot width = %d, want %d", w, calIconWidth)
	}
	if w := runewidth.StringWidth(calIconUrgent); w != calIconWidth {
		t.Fatalf("urgent icon width = %d, want %d", w, calIconWidth)
	}
	if w := runewidth.StringWidth(calIconDue); w != calIconWidth {
		t.Fatalf("due icon width = %d, want %d", w, calIconWidth)
	}
}

func TestCalIconSlot(t *testing.T) {
	if got := calIconSlot(calIconEmpty); runewidth.StringWidth(got) != calIconWidth {
		t.Fatalf("empty slot width = %d, want %d (%q)", runewidth.StringWidth(got), calIconWidth, got)
	}
	if got := calIconSlot(calIconDue); got != calIconDue {
		t.Fatalf("due slot = %q", got)
	}
	if got := calIconSlot(calIconUrgent); got != calIconUrgent {
		t.Fatalf("urgent slot = %q", got)
	}
}

func TestFormatCalendarCellContent_TermWidth(t *testing.T) {
	m := Model{}
	for _, cell := range []calDayCell{
		{day: 28, icon: calIconUrgent, urgent: true},
		{day: 5, icon: calIconDue, hasDue: true},
		{day: 12, icon: calIconEmpty},
	} {
		compact := m.formatCalendarCellContent(cell, calCellCompact)
		if w := runewidth.StringWidth(compact); w != calColWidth {
			t.Fatalf("day %d compact term width = %d, want %d (%q)", cell.day, w, calColWidth, compact)
		}
		if w := lipgloss.Width(compact); w != calColWidth {
			t.Fatalf("day %d compact lipgloss width = %d, want %d (%q)", cell.day, w, calColWidth, compact)
		}

		multi := m.formatCalendarCellContent(cell, calCellMulti)
		lines := strings.Split(multi, "\n")
		if len(lines) != 2 {
			t.Fatalf("day %d multi lines = %d", cell.day, len(lines))
		}
		for i, line := range lines {
			if w := runewidth.StringWidth(line); w != calColWidth {
				t.Fatalf("day %d line %d term width = %d, want %d (%q)", cell.day, i, w, calColWidth, line)
			}
			if w := lipgloss.Width(line); w != calColWidth {
				t.Fatalf("day %d line %d lipgloss width = %d, want %d (%q)", cell.day, i, w, calColWidth, line)
			}
		}
	}
}

func TestRenderCalendarCellLine_StyledTermWidth(t *testing.T) {
	m := Model{}
	cell := calDayCell{day: 28, icon: calIconUrgent, urgent: true}
	for line := 0; line < 2; line++ {
		styled := m.renderCalendarCellLine(cell, line, calCellMulti, true)
		if w := styledTermWidth(styled); w != calColWidth {
			t.Fatalf("line %d styled term width = %d, want %d", line, w, calColWidth)
		}
		if w := lipgloss.Width(styled); w != calColWidth {
			t.Fatalf("line %d styled lipgloss width = %d, want %d", line, w, calColWidth)
		}
	}
}

func TestRenderCalendarGrid_LineWidths(t *testing.T) {
	m := Model{calYear: 2026, calMonth: 6, calSelectedDay: 30}
	grid := [][]calDayCell{
		{
			{day: 28, icon: calIconUrgent, urgent: true},
			{day: 29, icon: calIconEmpty},
			{day: 30, icon: calIconDue, hasDue: true},
			{day: 31, icon: calIconEmpty},
			{},
			{},
			{},
		},
	}
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	out := m.renderCalendarGrid(grid, weekdays, calCellMulti, true)
	const wantW = calGridWidth
	for i, line := range strings.Split(out, "\n") {
		if w := runewidth.StringWidth(line); w != wantW {
			t.Fatalf("line %d term width = %d, want %d\n%q", i, w, wantW, line)
		}
		if w := lipgloss.Width(line); w != wantW {
			t.Fatalf("line %d lipgloss width = %d, want %d\n%q", i, w, wantW, line)
		}
	}
}

func TestCalendarPanelFitsGrid(t *testing.T) {
	oldDetect := colorProfileFor
	defer func() { colorProfileFor = oldDetect; initStyles() }()
	colorProfileFor = func(_ io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(nil)

	for _, tw := range []int{100, 120} {
		t.Run(fmt.Sprintf("w%d", tw), func(t *testing.T) {
			m := Model{
				width: tw, height: 40, lang: i18n.EN,
				calYear: 2026, calMonth: 6, calSelectedDay: 28,
				plants: []model.Plant{{ID: 1, Name: "A", State: model.StateWaterNow, BaseIntervalDays: 7}},
			}
			innerW := m.calendarPanelWidth() - sectionHPad
			if innerW < calGridWidth {
				t.Fatalf("innerW=%d < calGridWidth=%d", innerW, calGridWidth)
			}
			assertModalBordersAligned(t, m.renderCalendarPanel())
		})
	}
}

func TestCalendarGridVerticalSeparatorsAlign(t *testing.T) {
	oldDetect := colorProfileFor
	defer func() { colorProfileFor = oldDetect; initStyles() }()
	colorProfileFor = func(_ io.Writer) termenv.Profile { return termenv.TrueColor }
	InitTerminalDisplay(nil)

	m := Model{width: 120, calYear: 2026, calMonth: 6, calSelectedDay: 28}
	grid := [][]calDayCell{
		{
			{day: 28, icon: calIconUrgent, urgent: true, selected: true},
			{day: 29, icon: calIconEmpty},
			{day: 30, icon: calIconDue, hasDue: true, today: true},
			{day: 31, icon: calIconEmpty},
			{}, {}, {},
		},
		{
			{day: 1, icon: calIconDue, hasDue: true},
			{day: 2, icon: calIconEmpty},
			{day: 3, icon: calIconUrgent, urgent: true},
			{}, {}, {}, {},
		},
	}
	weekdays := []string{"Sun", "Mon", "Tue", "Wed", "Thu", "Fri", "Sat"}
	out := m.renderCalendarGrid(grid, weekdays, calCellCompact, true)
	var want []int
	for i, line := range strings.Split(out, "\n") {
		var cols []int
		w := styledTermWidth(line)
		for cell := 0; cell < w; cell++ {
			if ansi.Strip(ansi.Cut(line, cell, cell+1)) == "│" {
				cols = append(cols, cell)
			}
		}
		if len(cols) == 0 {
			continue
		}
		if want == nil {
			want = cols
			continue
		}
		for j := range want {
			if cols[j] != want[j] {
				t.Fatalf("row %d sep %d at %d want %d\n%q", i, j, cols[j], want[j], ansi.Strip(line))
			}
		}
	}
}
