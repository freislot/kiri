package tui

import (
	"strings"
	"testing"
	"time"

	"kiri/internal/i18n"
	"kiri/internal/model"
)

func TestRowStatusText_ShiftedByRainShowsWateringDate(t *testing.T) {
	m := Model{lang: i18n.RU}
	p := model.Plant{
		BaseIntervalDays: 7,
		WaterLevel:       60,
		State:            model.StateShiftedByRain,
		IsOutdoor:        true,
	}

	text := m.rowStatusText(p)
	if !strings.Contains(text, "Полив:") {
		t.Fatalf("row status = %q, want Полив: prefix", text)
	}
	if !strings.Contains(text, "(дождь)") {
		t.Fatalf("row status = %q, want (дождь) marker", text)
	}

	en := Model{lang: i18n.EN}
	enText := en.rowStatusText(p)
	if !strings.Contains(enText, "Water:") || !strings.Contains(enText, "(rain)") {
		t.Fatalf("EN row status = %q", enText)
	}
}

func TestDetailStatusText_ShiftedByRainShowsMovedDate(t *testing.T) {
	m := Model{lang: i18n.RU}
	p := model.Plant{
		BaseIntervalDays: 10,
		WaterLevel:       55,
		State:            model.StateShiftedByRain,
		IsOutdoor:        true,
	}

	text := m.detailStatusText(p)
	if !strings.Contains(text, "Движок сдвинул полив на") {
		t.Fatalf("detail status = %q", text)
	}
}

func TestRowStatusText_ShiftedByRainDateMatchesWateringDueDate(t *testing.T) {
	now := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	p := model.Plant{
		BaseIntervalDays: 7,
		WaterLevel:       50,
		State:            model.StateNormal,
		IsOutdoor:        true,
	}
	affected, _ := model.ApplyPrecipitationAt(&p, 5, 22, now)
	if !affected || p.State != model.StateShiftedByRain {
		t.Fatalf("precipitation should shift plant, state=%s", p.State)
	}

	due := model.WateringDueDate(p, now)
	m := Model{lang: i18n.RU}
	wantDate := m.cat().FormatDate(due)
	text := m.cat().RowStatusShiftedRain(due)
	if !strings.Contains(text, wantDate) {
		t.Fatalf("status %q should contain due date %q", text, wantDate)
	}
}
