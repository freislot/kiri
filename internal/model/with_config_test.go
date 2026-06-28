package model

import (
	"math"
	"testing"
	"time"
)

func customTestConfig() ModelConfig {
	return NormalizeModelConfig(ModelConfig{
		FallbackTempC:        18,
		WaterSoonPercent:     25,
		PostponeBoostPercent: 12,
		PostponeSuggestAfter: 3,
		Season: SeasonCoeffs{
			Winter: 0.4,
			Spring: 1.1,
			Summer: 2.0,
			Autumn: 0.8,
		},
		Rain: RainCoeffs{
			LightMM:        1.5,
			HeavyMM:        12,
			BaseShiftHours: 20,
		},
		Temp: TempCoeffs{
			CoolThresholdC: 18,
			CoolFactor:     0.75,
			HotThresholdC:  28,
			HotSlope:       0.04,
		},
	})
}

func TestPostponeWateringWithConfig_CustomBoost(t *testing.T) {
	cfg := customTestConfig()
	p := &Plant{BaseIntervalDays: 7, WaterLevel: 40, State: StateNormal}

	if suggest := PostponeWateringWithConfig(p, cfg); suggest {
		t.Fatal("first postpone should not suggest interval change")
	}
	if math.Abs(p.WaterLevel-52) > 0.01 {
		t.Fatalf("water level = %.1f, want 52 (+12%%)", p.WaterLevel)
	}

	PostponeWateringWithConfig(p, cfg)
	PostponeWateringWithConfig(p, cfg)
	if !PostponeWateringWithConfig(p, cfg) {
		t.Fatal("fourth postpone should suggest interval change with postpone_suggest_after=3")
	}
}

func TestApplyElapsedDegradationWithConfig_CustomSummer(t *testing.T) {
	cfg := customTestConfig()
	cfg.FallbackTempC = 22
	cfg = NormalizeModelConfig(cfg)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	end := start.Add(24 * time.Hour)

	defaultOutdoor := &Plant{BaseIntervalDays: 10, WaterLevel: 100, State: StateNormal, IsOutdoor: true}
	customOutdoor := &Plant{BaseIntervalDays: 10, WaterLevel: 100, State: StateNormal, IsOutdoor: true}

	ApplyElapsedDegradationAt(defaultOutdoor, start, end, 22)
	ApplyElapsedDegradationWithConfig(customOutdoor, start, end, 22, cfg)

	if customOutdoor.WaterLevel >= defaultOutdoor.WaterLevel {
		t.Fatalf("custom summer=2.0 should dry faster: custom=%.2f default=%.2f",
			customOutdoor.WaterLevel, defaultOutdoor.WaterLevel)
	}
}

func TestApplyPrecipitationWithConfig_CustomThresholds(t *testing.T) {
	cfg := customTestConfig()

	below := &Plant{BaseIntervalDays: 7, IsOutdoor: true, WaterLevel: 50, State: StateNormal}
	affected, _ := ApplyPrecipitationWithConfig(below, 1.4, 22, cfg)
	if affected {
		t.Fatal("1.4 mm should be below custom light_mm=1.5")
	}

	light := &Plant{BaseIntervalDays: 7, IsOutdoor: true, WaterLevel: 50, State: StateNormal}
	at := time.Date(2026, 7, 15, 9, 0, 0, 0, time.UTC)
	affected, effect := cfg.ApplyPrecipitationAt(light, 5, 22, at)
	if !affected || light.State != StateShiftedByRain {
		t.Fatalf("light rain: affected=%v state=%s", affected, light.State)
	}
	wantShift := 20.0 * (1.1 / 2.0) // base_shift_hours × spring/summer
	if math.Abs(effect.ShiftHours-wantShift) > 0.05 {
		t.Fatalf("shift hours = %.2f, want %.2f", effect.ShiftHours, wantShift)
	}

	heavy := &Plant{
		BaseIntervalDays:     7,
		IsOutdoor:            true,
		WaterLevel:           50,
		State:                StateNormal,
		ConsecutivePostpones: 2,
	}
	affected, _ = ApplyPrecipitationWithConfig(heavy, 12.5, 22, cfg)
	if !affected || heavy.WaterLevel != 100 || heavy.ConsecutivePostpones != 0 {
		t.Fatalf("heavy rain: level=%.0f postpones=%d", heavy.WaterLevel, heavy.ConsecutivePostpones)
	}
}

func TestWateringDueDateWithConfig_UsesFallbackTemp(t *testing.T) {
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	p := Plant{BaseIntervalDays: 10, WaterLevel: 50, State: StateNormal}

	defaultDue := WateringDueDate(p, now)

	hotCfg := customTestConfig()
	hotCfg.FallbackTempC = 35 // hotter → faster drying → sooner due date
	hotDue := WateringDueDateWithConfig(p, now, hotCfg)

	if !hotDue.Before(defaultDue) && !hotDue.Equal(defaultDue) {
		t.Fatalf("hotter fallback should not postpone due date: default=%v hot=%v", defaultDue, hotDue)
	}
}

func TestIsDueOnDateWithConfig_MatchesDefault(t *testing.T) {
	now := time.Date(2026, 6, 10, 9, 0, 0, 0, time.UTC)
	p := Plant{BaseIntervalDays: 7, WaterLevel: 0, State: StateWaterNow}
	day := now

	due1, urgent1 := IsDueOnDate(p, day, now)
	due2, urgent2 := IsDueOnDateWithConfig(p, day, now, DefaultConfig())
	if due1 != due2 || urgent1 != urgent2 {
		t.Fatalf("default wrappers mismatch: (%v,%v) vs (%v,%v)", due1, urgent1, due2, urgent2)
	}
}

func TestRefreshPlantStateWithConfig_CustomWaterSoon(t *testing.T) {
	cfg := customTestConfig()
	p := &Plant{BaseIntervalDays: 10, WaterLevel: 24, State: StateNormal}
	RefreshPlantStateWithConfig(p, cfg)
	if p.NextAction != "Water soon" {
		t.Fatalf("next_action = %q, want Water soon at 24%% with water_soon_percent=25", p.NextAction)
	}
}

func TestFormatNextActionWithConfig(t *testing.T) {
	cfg := customTestConfig()
	p := &Plant{BaseIntervalDays: 10, WaterLevel: 50, State: StateNormal}
	got := FormatNextActionWithConfig(p, cfg)
	if got == "" {
		t.Fatal("expected non-empty next action")
	}
}
