package tui

import (
	"testing"

	"kiri/internal/i18n"
	"kiri/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func TestFooterHintClickKey(t *testing.T) {
	tests := []struct {
		key    string
		want   string
		click  bool
	}{
		{"a", "a", true},
		{"q", "q", true},
		{"Space", " ", true},
		{"Enter/y", "enter", true},
		{"n/Esc", "n", true},
		{"⌨ h/j/k/l", "", false},
		{"h/l", "", false},
		{"1–5", "", false},
	}
	for _, tc := range tests {
		got, ok := footerHintClickKey(i18n.FooterHint{Key: tc.key})
		if ok != tc.click || got != tc.want {
			t.Fatalf("footerHintClickKey(%q) = (%q, %v), want (%q, %v)", tc.key, got, ok, tc.want, tc.click)
		}
	}
}

func TestTabAt(t *testing.T) {
	m := Model{
		width:  100,
		height: 30,
		tab:    TabPlants,
		lang:   i18n.EN,
	}
	regions := m.tabRegions()
	if len(regions) != tabCount {
		t.Fatalf("tab regions = %d, want %d", len(regions), tabCount)
	}
	mid := (regions[2].x0 + regions[2].x1) / 2
	if tab, ok := m.tabAt(mid, m.frameHeaderY()); !ok || tab != TabCareLog {
		t.Fatalf("tabAt care log = (%v, %v), want (%v, true)", tab, ok, TabCareLog)
	}
}

func TestPlantCardAt(t *testing.T) {
	m := Model{
		width:   100,
		height:  30,
		tab:     TabPlants,
		lang:    i18n.EN,
		plants:  []model.Plant{{Name: "A"}, {Name: "B"}, {Name: "C"}},
		visRows: 2,
		visCols: 3,
	}
	m.recalcLayout()

	slots := m.plantCardSlotWidths()
	x := m.frameContentX() + slots[0]/2
	y := m.plantsGridY() + 1

	row, col, ok := m.plantCardAt(x, y)
	if !ok || row != 0 || col != 0 {
		t.Fatalf("plantCardAt first card = (%d,%d,%v), want (0,0,true)", row, col, ok)
	}

	x2 := m.frameContentX() + slots[0] + slots[1]/2
	row, col, ok = m.plantCardAt(x2, y)
	if !ok || row != 0 || col != 1 {
		t.Fatalf("plantCardAt second card = (%d,%d,%v), want (0,1,true)", row, col, ok)
	}
}

func TestFooterRegionAt(t *testing.T) {
	m := Model{
		width:  120,
		height: 30,
		tab:    TabPlants,
		lang:   i18n.EN,
	}
	_, regions := footerHintLayout(m.currentFooterHints(), m.headerTextWidth())
	if len(regions) == 0 {
		t.Fatal("expected clickable footer regions")
	}
	r := regions[0]
	x := m.frameContentX() + (r.X0+r.X1)/2
	key, ok := m.footerRegionAt(x, m.footerY())
	if !ok || key == "" {
		t.Fatalf("footerRegionAt = (%q, %v), want non-empty key", key, ok)
	}
}

func TestCareLogWheelScroll(t *testing.T) {
	m := Model{
		width:       100,
		height:      30,
		tab:         TabCareLog,
		lang:        i18n.EN,
		splashPhase: splashPhaseDone,
	}
	for i := 0; i < 40; i++ {
		m.careLog = append(m.careLog, model.CareLogEntry{PlantName: "p"})
	}

	next, _ := m.handleMouse(tea.MouseMsg{
		X:      m.frameContentX() + 2,
		Y:      m.contentStartY() + 2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelDown,
	})
	got := next.(Model)
	want := m.careLogWheelStep()
	if got.careLogOffset != want {
		t.Fatalf("wheel down offset = %d, want %d", got.careLogOffset, want)
	}

	next, _ = m.handleMouse(tea.MouseMsg{
		X:      m.frameContentX() + 2,
		Y:      m.contentStartY() + 2,
		Action: tea.MouseActionPress,
		Button: tea.MouseButtonWheelUp,
	})
	got = next.(Model)
	if got.careLogOffset != 0 {
		t.Fatalf("wheel up offset = %d, want 0", got.careLogOffset)
	}
}

func TestSettingsWheelMovesCursor(t *testing.T) {
	m := Model{
		width:       100,
		height:      30,
		tab:         TabSettings,
		lang:        i18n.EN,
		cursor:      settingsRowLanguage,
		splashPhase: splashPhaseDone,
	}

	wheel := func(button tea.MouseButton) Model {
		next, _ := m.handleMouse(tea.MouseMsg{
			X:      m.frameContentX() + 2,
			Y:      m.contentStartY() + 2,
			Action: tea.MouseActionPress,
			Button: button,
		})
		m = next.(Model)
		return m
	}

	got := wheel(tea.MouseButtonWheelDown)
	if got.cursor != settingsRowCity {
		t.Fatalf("wheel down cursor = %d, want %d", got.cursor, settingsRowCity)
	}

	got = wheel(tea.MouseButtonWheelUp)
	if got.cursor != settingsRowLanguage {
		t.Fatalf("wheel up cursor = %d, want %d", got.cursor, settingsRowLanguage)
	}
}
