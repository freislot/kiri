package model

import (
	"math"
	"testing"
	"time"
)

func TestWateringDueDate(t *testing.T) {
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	p := Plant{
		BaseIntervalDays: 10,
		WaterLevel:       50,
		State:            StateNormal,
	}

	due := WateringDueDate(p, now)
	if due.Before(dateOnly(now)) {
		t.Fatalf("due date should not be in the past: got %v", due)
	}

	delta := DeltaVForPlant(&p, now, 22)
	wantDays := int(math.Ceil(50 / delta))
	want := dateOnly(now.AddDate(0, 0, wantDays))
	if !due.Equal(want) {
		t.Fatalf("due date = %v, want %v", due, want)
	}
}

func TestWateringDueDate_WaterNowReturnsToday(t *testing.T) {
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	p := Plant{BaseIntervalDays: 7, WaterLevel: 0, State: StateWaterNow}
	due := WateringDueDate(p, now)
	if !due.Equal(dateOnly(now)) {
		t.Fatalf("water_now due = %v, want today %v", due, dateOnly(now))
	}
}

func TestWateringDueDate_ZeroDeltaFallback(t *testing.T) {
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	p := Plant{BaseIntervalDays: 0, WaterLevel: 50, State: StateNormal}
	due := WateringDueDate(p, now)
	want := dateOnly(now.AddDate(0, 0, 7))
	if !due.Equal(want) {
		t.Fatalf("zero delta fallback = %v, want %v", due, want)
	}
}

func TestApplyPrecipitation_PostponesWateringDueDate(t *testing.T) {
	now := time.Date(2026, 7, 15, 10, 0, 0, 0, time.UTC)
	baseline := Plant{
		BaseIntervalDays: 7,
		WaterLevel:       20,
		State:            StateNormal,
		IsOutdoor:        true,
	}
	shifted := Plant{
		BaseIntervalDays: 7,
		WaterLevel:       20,
		State:            StateNormal,
		IsOutdoor:        true,
	}

	affected, _ := ApplyPrecipitationAt(&shifted, 5, 22, now)
	if !affected {
		t.Fatal("light rain should affect outdoor plant")
	}
	if shifted.State != StateShiftedByRain {
		t.Fatalf("state = %s, want %s", shifted.State, StateShiftedByRain)
	}

	dueBaseline := WateringDueDate(baseline, now)
	dueShifted := WateringDueDate(shifted, now)
	if !dueShifted.After(dueBaseline) {
		t.Fatalf("rain should postpone watering: baseline=%v shifted=%v", dueBaseline, dueShifted)
	}
}

func TestIsDueOnDate_WaterNowOnlyToday(t *testing.T) {
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	p := Plant{
		BaseIntervalDays: 7,
		WaterLevel:       0,
		State:            StateWaterNow,
	}

	pastDay := now.AddDate(0, 0, -1)
	today := now
	future := now.AddDate(0, 0, 1)

	due, urgent := IsDueOnDate(p, pastDay, now)
	if due || urgent {
		t.Fatalf("past day should not be due for water_now: due=%v urgent=%v", due, urgent)
	}
	due, urgent = IsDueOnDate(p, today, now)
	if !due || !urgent {
		t.Fatalf("today should be due+urgent for water_now")
	}
	due, _ = IsDueOnDate(p, future, now)
	if due {
		t.Fatalf("future day should not be due for water_now")
	}
}

func TestIsDueOnDate_NormalPlantOnDueDate(t *testing.T) {
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	p := Plant{
		BaseIntervalDays: 10,
		WaterLevel:       50,
		State:            StateNormal,
	}
	dueDate := WateringDueDate(p, now)

	due, urgent := IsDueOnDate(p, dueDate, now)
	if !due || urgent {
		t.Fatalf("due date should be due but not urgent: due=%v urgent=%v", due, urgent)
	}

	other := dueDate.AddDate(0, 0, 1)
	due, _ = IsDueOnDate(p, other, now)
	if due {
		t.Fatalf("day after due date should not be due")
	}
}

func TestSeasonFromTime(t *testing.T) {
	july := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	if got := SeasonFromTime(july); got != SeasonSummer {
		t.Fatalf("July season = %v, want summer", got)
	}
}
