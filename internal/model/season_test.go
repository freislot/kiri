package model

import (
	"math"
	"testing"
	"time"
)

func TestOutdoorSeasonCoefficient(t *testing.T) {
	tests := []struct {
		month time.Month
		want  float64
	}{
		{time.June, 1.5},
		{time.July, 1.5},
		{time.August, 1.5},
		{time.September, 0.75},
		{time.October, 0.75},
		{time.November, 0.75},
		{time.March, 1.0},
		{time.April, 1.0},
		{time.May, 1.0},
		{time.December, 0.5},
		{time.January, 0.5},
		{time.February, 0.5},
	}

	for _, tt := range tests {
		got := OutdoorSeasonCoefficient(tt.month)
		if got != tt.want {
			t.Fatalf("month %v: got %.2f want %.2f", tt.month, got, tt.want)
		}
	}
}

func TestRainShiftFactor_Physics(t *testing.T) {
	summer := RainShiftFactor(time.July)
	autumn := RainShiftFactor(time.October)
	spring := RainShiftFactor(time.April)
	winter := RainShiftFactor(time.January)

	if summer >= spring {
		t.Fatalf("summer shift factor %.2f should be below spring %.2f", summer, spring)
	}
	if autumn <= spring {
		t.Fatalf("autumn shift factor %.2f should be above spring %.2f", autumn, spring)
	}
	if math.Abs(spring-1.0) > 0.01 {
		t.Fatalf("spring shift factor = %.2f, want 1.0", spring)
	}
	if winter > 1.0 {
		t.Fatalf("winter shift factor capped at 1.0, got %.2f", winter)
	}
}

func TestSeasonCoefficient_IndoorAlwaysOne(t *testing.T) {
	p := &Plant{IsOutdoor: false}
	at := time.Date(2026, 7, 15, 12, 0, 0, 0, time.UTC)
	if got := SeasonCoefficient(p, at); got != 1.0 {
		t.Fatalf("indoor K_season = %.1f, want 1.0", got)
	}
}

func TestSeasonFactor(t *testing.T) {
	if got := SeasonSummer.Factor(); got != 1.5 {
		t.Fatalf("summer factor = %.2f, want 1.5", got)
	}
	if got := SeasonWinter.Factor(); got != 0.5 {
		t.Fatalf("winter factor = %.2f, want 0.5", got)
	}
}

func TestRainShiftHoursBySeason(t *testing.T) {
	outdoor := &Plant{IsOutdoor: true, BaseIntervalDays: 7}
	july := time.Date(2026, 7, 10, 8, 0, 0, 0, time.UTC)
	october := time.Date(2026, 10, 10, 8, 0, 0, 0, time.UTC)

	julyShift := RainShiftHours(outdoor, july)
	if math.Abs(julyShift-16.0) > 0.01 {
		t.Fatalf("july shift: got %.1f want 16.0 (24 × 1/1.5)", julyShift)
	}

	octShift := RainShiftHours(outdoor, october)
	if math.Abs(octShift-32.0) > 0.01 {
		t.Fatalf("october shift: got %.1f want 32.0 (24 × 1/0.75)", octShift)
	}
}

func TestSeasonFactor_Autumn(t *testing.T) {
	if got := SeasonAutumn.Factor(); got != 0.75 {
		t.Fatalf("autumn factor = %.2f, want 0.75", got)
	}
}
