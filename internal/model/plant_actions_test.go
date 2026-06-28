package model

import (
	"math"
	"testing"
	"time"
)

func TestTempFactor(t *testing.T) {
	tests := []struct {
		tempC float64
		want  float64
	}{
		{15, 0.8},
		{20, 0.8},
		{21, 1.0},
		{27, 1.0},
		{30, 1.15},
		{37, 1.5},
	}
	for _, tt := range tests {
		got := TempFactor(tt.tempC)
		if math.Abs(got-tt.want) > 0.001 {
			t.Fatalf("TempFactor(%.0f) = %.3f, want %.3f", tt.tempC, got, tt.want)
		}
	}
}

func TestDeltaV_EdgeCases(t *testing.T) {
	if got := DeltaV(0, 1.0, 22); got != 0 {
		t.Fatalf("zero interval: got %.2f want 0", got)
	}
	if got := DeltaV(-1, 1.0, 22); got != 0 {
		t.Fatalf("negative interval: got %.2f want 0", got)
	}
	want := (100.0 / 10.0) * 1.0 * TempFactor(22)
	if got := DeltaV(10, 1.0, 22); math.Abs(got-want) > 0.001 {
		t.Fatalf("DeltaV(10,1,22) = %.3f want %.3f", got, want)
	}
}

func TestWaterPlant(t *testing.T) {
	p := &Plant{
		BaseIntervalDays:     7,
		WaterLevel:           12,
		ConsecutivePostpones: 2,
		State:                StateWaterNow,
	}
	WaterPlant(p)
	if p.WaterLevel != 100 {
		t.Fatalf("water level = %.0f, want 100", p.WaterLevel)
	}
	if p.ConsecutivePostpones != 0 {
		t.Fatalf("postpones = %d, want 0", p.ConsecutivePostpones)
	}
	if p.State != StateNormal {
		t.Fatalf("state = %s, want %s", p.State, StateNormal)
	}
	if p.LastUpdatedAt.IsZero() {
		t.Fatal("last updated should be set")
	}
}

func TestPostponeWatering(t *testing.T) {
	p := &Plant{BaseIntervalDays: 7, WaterLevel: 40, State: StateNormal}
	if suggest := PostponeWatering(p); suggest {
		t.Fatal("first postpone should not suggest interval change")
	}
	if p.WaterLevel != 55 {
		t.Fatalf("water level = %.0f, want 55", p.WaterLevel)
	}
	if p.ConsecutivePostpones != 1 {
		t.Fatalf("postpones = %d, want 1", p.ConsecutivePostpones)
	}

	if suggest := PostponeWatering(p); !suggest {
		t.Fatal("second postpone should suggest interval change")
	}
	if p.ConsecutivePostpones != 2 {
		t.Fatalf("postpones = %d, want 2", p.ConsecutivePostpones)
	}
}

func TestPostponeWatering_CapsAt100(t *testing.T) {
	p := &Plant{BaseIntervalDays: 7, WaterLevel: 95, State: StateNormal}
	PostponeWatering(p)
	if p.WaterLevel != 100 {
		t.Fatalf("water level capped at 100, got %.0f", p.WaterLevel)
	}
}

func TestRefreshPlantState(t *testing.T) {
	tests := []struct {
		name       string
		p          Plant
		wantState  string
		wantAction string
	}{
		{
			name:       "new without interval",
			p:          Plant{BaseIntervalDays: 0, WaterLevel: 50},
			wantState:  StateNew,
			wantAction: "Configure interval",
		},
		{
			name:       "explicit new state",
			p:          Plant{BaseIntervalDays: 7, WaterLevel: 50, State: StateNew},
			wantState:  StateNew,
			wantAction: "Configure interval",
		},
		{
			name:       "critical thirst",
			p:          Plant{BaseIntervalDays: 7, WaterLevel: 0, State: StateNormal},
			wantState:  StateWaterNow,
			wantAction: "WATER NOW!",
		},
		{
			name:       "rain shift preserved",
			p:          Plant{BaseIntervalDays: 7, WaterLevel: 55, State: StateShiftedByRain},
			wantState:  StateShiftedByRain,
		},
		{
			name:       "normal healthy",
			p:          Plant{BaseIntervalDays: 10, WaterLevel: 80, State: StateNormal},
			wantState:  StateNormal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := tt.p
			RefreshPlantState(&p)
			if p.State != tt.wantState {
				t.Fatalf("state = %s, want %s", p.State, tt.wantState)
			}
			if tt.wantAction != "" && p.NextAction != tt.wantAction {
				t.Fatalf("next action = %q, want %q", p.NextAction, tt.wantAction)
			}
		})
	}
}

func TestFormatNextAction(t *testing.T) {
	tests := []struct {
		level float64
		want  string
	}{
		{15, "Water soon"},
		{20, "Water soon"},
		{50, ""}, // dynamic — just check non-empty
	}
	for _, tt := range tests {
		p := &Plant{BaseIntervalDays: 10, WaterLevel: tt.level, State: StateNormal}
		got := FormatNextAction(p)
		if tt.want != "" {
			if got != tt.want {
				t.Fatalf("level %.0f: got %q want %q", tt.level, got, tt.want)
			}
		} else if got == "" {
			t.Fatalf("level %.0f: expected non-empty action", tt.level)
		}
	}
}

func TestApplyElapsedDegradationAt_SkipsNewAndInvalid(t *testing.T) {
	now := time.Now()
	start := now.Add(-24 * time.Hour)

	pNew := &Plant{BaseIntervalDays: 7, WaterLevel: 100, State: StateNew}
	ApplyElapsedDegradationAt(pNew, start, now, 22)
	if pNew.WaterLevel != 100 {
		t.Fatalf("new plant should not degrade, got %.1f", pNew.WaterLevel)
	}

	pZero := &Plant{BaseIntervalDays: 0, WaterLevel: 100, State: StateNormal}
	ApplyElapsedDegradationAt(pZero, start, now, 22)
	if pZero.WaterLevel != 100 {
		t.Fatalf("zero interval should not degrade, got %.1f", pZero.WaterLevel)
	}

	p := &Plant{BaseIntervalDays: 7, WaterLevel: 100, State: StateNormal}
	ApplyElapsedDegradationAt(p, now, start, 22)
	if p.WaterLevel != 100 {
		t.Fatalf("reverse time span should not degrade, got %.1f", p.WaterLevel)
	}
}

func TestApplyElapsedDegradationAt_FloorsAtZero(t *testing.T) {
	p := &Plant{
		BaseIntervalDays: 1,
		WaterLevel:       1,
		State:            StateNormal,
		IsOutdoor:        true,
	}
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(30 * 24 * time.Hour)
	ApplyElapsedDegradationAt(p, start, end, 30)
	if p.WaterLevel != 0 {
		t.Fatalf("water level should floor at 0, got %.2f", p.WaterLevel)
	}
	if p.State != StateWaterNow {
		t.Fatalf("state = %s, want %s", p.State, StateWaterNow)
	}
}
