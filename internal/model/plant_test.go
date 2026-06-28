package model

import (
	"math"
	"testing"
	"time"
)

func TestApplyElapsedDegradationAt_PartialDay(t *testing.T) {
	p := &Plant{
		BaseIntervalDays: 10,
		WaterLevel:       100,
		State:            StateNormal,
		IsOutdoor:        true,
	}
	start := time.Date(2026, 6, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(12 * time.Hour)

	ApplyElapsedDegradationAt(p, start, end, 22)

	want := 100 - (DeltaVForPlant(p, start, 22) * 0.5)
	if math.Abs(p.WaterLevel-want) > 0.01 {
		t.Fatalf("water level mismatch: got %.3f want %.3f", p.WaterLevel, want)
	}
	if !p.LastUpdatedAt.Equal(end) {
		t.Fatalf("last updated mismatch: got %v want %v", p.LastUpdatedAt, end)
	}
}

func TestApplyElapsedDegradationAt_OutdoorSummerVsIndoor(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	outdoor := &Plant{BaseIntervalDays: 10, WaterLevel: 100, State: StateNormal, IsOutdoor: true}
	indoor := &Plant{BaseIntervalDays: 10, WaterLevel: 100, State: StateNormal, IsOutdoor: false}

	ApplyElapsedDegradationAt(outdoor, start, end, 22)
	ApplyElapsedDegradationAt(indoor, start, end, 22)

	if outdoor.WaterLevel >= indoor.WaterLevel {
		t.Fatalf("outdoor summer should dry faster: outdoor=%.2f indoor=%.2f", outdoor.WaterLevel, indoor.WaterLevel)
	}
}

func TestApplyElapsedDegradationAt_CrossSeasonBoundary(t *testing.T) {
	p := &Plant{
		BaseIntervalDays: 10,
		WaterLevel:       100,
		State:            StateNormal,
		IsOutdoor:        true,
	}
	start := time.Date(2026, 5, 31, 12, 0, 0, 0, time.UTC)
	end := time.Date(2026, 6, 1, 12, 0, 0, 0, time.UTC)

	ApplyElapsedDegradationAt(p, start, end, 22)

	deltaSpringHalf := DeltaVForPlant(p, start, 22) * 0.5
	deltaSummerHalf := DeltaVForPlant(p, end.Add(-12*time.Hour), 22) * 0.5
	want := 100 - deltaSpringHalf - deltaSummerHalf
	if math.Abs(p.WaterLevel-want) > 0.01 {
		t.Fatalf("cross-season water level mismatch: got %.3f want %.3f", p.WaterLevel, want)
	}
}

func TestApplyPrecipitationThresholds(t *testing.T) {
	july := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
	october := time.Date(2026, 10, 15, 9, 0, 0, 0, time.UTC)

	tests := []struct {
		name       string
		precipMM   float64
		start      float64
		at         time.Time
		wantLevel  float64
		wantState  string
		wantAffect bool
		wantShift  float64
	}{
		{name: "below threshold", precipMM: 0.9, start: 50, at: july, wantLevel: 50, wantState: StateNormal, wantAffect: false},
		{
			name: "light rain summer", precipMM: 5, start: 50, at: july,
			wantState: StateShiftedByRain, wantAffect: true, wantShift: 16.0,
		},
		{
			name: "light rain autumn", precipMM: 5, start: 50, at: october,
			wantState: StateShiftedByRain, wantAffect: true, wantShift: 32.0,
		},
		{name: "heavy rain", precipMM: 12, start: 50, at: july, wantLevel: 100, wantState: StateNormal, wantAffect: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := &Plant{
				BaseIntervalDays: 7,
				IsOutdoor:        true,
				WaterLevel:       tt.start,
				State:            StateNormal,
			}
			affected, effect := ApplyPrecipitationAt(p, tt.precipMM, 22, tt.at)
			if affected != tt.wantAffect {
				t.Fatalf("affected mismatch: got %v want %v", affected, tt.wantAffect)
			}
			if tt.wantShift > 0 && math.Abs(effect.ShiftHours-tt.wantShift) > 0.01 {
				t.Fatalf("shift hours: got %.1f want %.1f", effect.ShiftHours, tt.wantShift)
			}
			if tt.wantLevel > 0 {
				if math.Abs(p.WaterLevel-tt.wantLevel) > 0.01 {
					t.Fatalf("level mismatch: got %.2f want %.2f", p.WaterLevel, tt.wantLevel)
				}
			} else if tt.wantAffect && tt.precipMM >= 1 && tt.precipMM <= 10 {
				vBase := 100.0 / 7.0
				wantAdded := vBase * TempFactor(22) * (tt.wantShift / 24.0)
				wantLevel := tt.start + wantAdded
				if math.Abs(p.WaterLevel-wantLevel) > 0.01 {
					t.Fatalf("level mismatch: got %.2f want %.2f", p.WaterLevel, wantLevel)
				}
			}
			if p.State != tt.wantState {
				t.Fatalf("state mismatch: got %s want %s", p.State, tt.wantState)
			}
		})
	}
}

func TestApplyPrecipitation_IndoorIgnored(t *testing.T) {
	p := &Plant{
		BaseIntervalDays: 7,
		IsOutdoor:        false,
		WaterLevel:       50,
		State:            StateNormal,
	}
	affected, _ := ApplyPrecipitation(p, 5, 22)
	if affected {
		t.Fatal("indoor plant should not be affected by rain")
	}
	if p.WaterLevel != 50 {
		t.Fatalf("water level changed: got %.1f", p.WaterLevel)
	}
}
