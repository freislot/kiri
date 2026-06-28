package statuscli

import (
	"time"

	"kiri/internal/db"
	"kiri/internal/i18n"
	"kiri/internal/model"
)

func loadPlants(store *db.Store) ([]model.Plant, model.ModelConfig, i18n.Catalog, error) {
	cfg, err := store.LoadSettings()
	if err != nil {
		return nil, model.ModelConfig{}, i18n.Catalog{}, err
	}
	plants, err := store.ListPlants()
	if err != nil {
		return nil, model.ModelConfig{}, i18n.Catalog{}, err
	}
	now := time.Now()
	modelCfg := cfg.Model
	for i := range plants {
		applyDegradation(&plants[i], now, modelCfg)
	}
	return plants, modelCfg, i18n.New(cfg.Language), nil
}

func applyDegradation(p *model.Plant, now time.Time, cfg model.ModelConfig) {
	if p.State == model.StateNew || p.BaseIntervalDays <= 0 || p.LastUpdatedAt.IsZero() || !now.After(p.LastUpdatedAt) {
		return
	}
	model.ApplyElapsedDegradationWithConfig(p, p.LastUpdatedAt, now, cfg.FallbackTempC, cfg)
	model.RefreshPlantStateWithConfig(p, cfg)
}

func plantNeedsWatering(p model.Plant, now time.Time, cfg model.ModelConfig) bool {
	if p.State == model.StateWaterNow || p.WaterLevel <= 0 {
		return true
	}
	return daysUntil(model.WateringDueDateWithConfig(p, now, cfg), now) <= 0
}

func countNeedsWatering(plants []model.Plant, now time.Time, cfg model.ModelConfig) int {
	n := 0
	for _, p := range plants {
		if plantNeedsWatering(p, now, cfg) {
			n++
		}
	}
	return n
}
