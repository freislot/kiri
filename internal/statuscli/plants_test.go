package statuscli

import (
	"path/filepath"
	"testing"
	"time"

	"kiri/internal/db"
	"kiri/internal/i18n"
	"kiri/internal/model"
)

func TestApplyDegradation_UsesCustomModelConfig(t *testing.T) {
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(24 * time.Hour)
	customCfg := model.NormalizeModelConfig(model.ModelConfig{
		FallbackTempC: 22,
		Season: model.SeasonCoeffs{
			Winter: 0.5,
			Spring: 1.0,
			Summer: 3.0,
			Autumn: 0.75,
		},
		Rain: model.RainCoeffs{LightMM: 1, HeavyMM: 10, BaseShiftHours: 24},
		Temp: model.TempCoeffs{
			CoolThresholdC: 20,
			CoolFactor:     0.8,
			HotThresholdC:  27,
			HotSlope:       0.05,
		},
	})

	custom := model.Plant{
		BaseIntervalDays: 10,
		IsOutdoor:        true,
		WaterLevel:       100,
		State:            model.StateNormal,
		LastUpdatedAt:    start,
	}
	defaultPlant := model.Plant{
		BaseIntervalDays: 10,
		IsOutdoor:        true,
		WaterLevel:       100,
		State:            model.StateNormal,
		LastUpdatedAt:    start,
	}

	applyDegradation(&custom, now, customCfg)
	model.ApplyElapsedDegradationAt(&defaultPlant, start, now, 22)

	if custom.WaterLevel >= defaultPlant.WaterLevel {
		t.Fatalf("custom summer=3.0 should dry more: custom=%.2f default=%.2f",
			custom.WaterLevel, defaultPlant.WaterLevel)
	}
}

func TestPlantNeedsWatering_UsesCustomDueDate(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	cfg := model.DefaultConfig()

	urgent := model.Plant{BaseIntervalDays: 7, WaterLevel: 0, State: model.StateWaterNow}
	if !plantNeedsWatering(urgent, now, cfg) {
		t.Fatal("water_now plant should need watering")
	}

	healthy := model.Plant{BaseIntervalDays: 10, WaterLevel: 100, State: model.StateNormal}
	if plantNeedsWatering(healthy, now, cfg) {
		t.Fatal("full plant should not need watering")
	}
}

func TestLoadPlants_LocalizesFromSettings(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()
	clearPlants(t, store)

	if err := store.SetLanguage(i18n.EN); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	_, _, cat, err := loadPlants(store)
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if got := cat.CLISummaryOK(); got != "🌿 kiri: ok" {
		t.Fatalf("catalog lang = %q", got)
	}
}
