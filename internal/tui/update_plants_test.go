package tui

import (
	"testing"
	"time"

	"kiri/internal/i18n"
	"kiri/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandlePlantsKey_OKeyIgnored(t *testing.T) {
	now := time.Now()
	p := model.Plant{
		ID:               1,
		Name:             "Rose",
		BaseIntervalDays: 7,
		WaterLevel:       60,
		State:            model.StateShiftedByRain,
		IsOutdoor:        true,
		LastUpdatedAt:    now,
		CreatedAt:        now,
	}
	m := Model{
		tab:         TabPlants,
		lang:        i18n.EN,
		plants:      []model.Plant{p},
		cursorRow:   0,
		cursorCol:   0,
		visRows:     2,
		visCols:     3,
		splashPhase: splashPhaseDone,
	}

	next, cmd := m.handlePlantsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'o'}})
	got := next.(Model)
	if cmd != nil {
		t.Fatalf("o should not trigger command, got %v", cmd)
	}
	if got.plants[0].State != model.StateShiftedByRain {
		t.Fatalf("state after o = %s, want shifted_by_rain", got.plants[0].State)
	}
	if got.status != "" {
		t.Fatalf("status after o = %q, want empty", got.status)
	}
}