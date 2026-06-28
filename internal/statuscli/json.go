package statuscli

import (
	"encoding/json"
	"io"
	"time"

	"kiri/internal/db"
	"kiri/internal/model"
)

type plantJSON struct {
	ID                   int64   `json:"id"`
	Name                 string  `json:"name"`
	Location             string  `json:"location"`
	BaseIntervalDays     int     `json:"base_interval_days"`
	IsOutdoor            bool    `json:"is_outdoor"`
	WaterLevel           float64 `json:"water_level"`
	ConsecutivePostpones int     `json:"consecutive_postpones"`
	State                string  `json:"state"`
	NextAction           string  `json:"next_action"`
	LastUpdatedAt        string  `json:"last_updated_at"`
	CreatedAt            string  `json:"created_at"`
	NeedsWatering        bool    `json:"needs_watering"`
	WateringDueDate      string  `json:"watering_due_date"`
	DaysUntilWatering    int     `json:"days_until_watering"`
}

func PrintJSON(w io.Writer, store *db.Store) error {
	plants, modelCfg, _, err := loadPlants(store)
	if err != nil {
		return err
	}

	now := time.Now()
	out := make([]plantJSON, len(plants))
	for i, p := range plants {
		out[i] = toPlantJSON(p, now, modelCfg)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(out)
}

func toPlantJSON(p model.Plant, now time.Time, cfg model.ModelConfig) plantJSON {
	due := model.WateringDueDateWithConfig(p, now, cfg)
	days := daysUntil(due, now)
	return plantJSON{
		ID:                   p.ID,
		Name:                 p.Name,
		Location:             p.Location,
		BaseIntervalDays:     p.BaseIntervalDays,
		IsOutdoor:            p.IsOutdoor,
		WaterLevel:           p.WaterLevel,
		ConsecutivePostpones: p.ConsecutivePostpones,
		State:                p.State,
		NextAction:           p.NextAction,
		LastUpdatedAt:        p.LastUpdatedAt.UTC().Format(time.RFC3339),
		CreatedAt:            p.CreatedAt.UTC().Format(time.RFC3339),
		NeedsWatering:        plantNeedsWatering(p, now, cfg),
		WateringDueDate:      due.Format("2006-01-02"),
		DaysUntilWatering:    days,
	}
}
