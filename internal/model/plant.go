package model

import "time"

const (
	StateNormal        = "normal"
	StateShiftedByRain = "shifted_by_rain"
	StateWaterNow      = "water_now"
	StateNew           = "new"
)

type Plant struct {
	ID                   int64
	Name                 string
	Location             string
	BaseIntervalDays     int
	IsOutdoor            bool
	WaterLevel           float64
	ConsecutivePostpones int
	State                string
	NextAction           string
	LastUpdatedAt        time.Time
	CreatedAt            time.Time
}

type CareLogEntry struct {
	ID        int64
	PlantID   int64
	PlantName string
	EventType string
	Message   string
	CreatedAt time.Time
}

func DeltaV(baseIntervalDays int, kSeason, tempC float64) float64 {
	return DefaultConfig().DeltaV(baseIntervalDays, kSeason, tempC)
}

func DeltaVForPlant(p *Plant, t time.Time, tempC float64) float64 {
	return DefaultConfig().DeltaVForPlant(p, t, tempC)
}

func TempFactor(tempC float64) float64 {
	return DefaultConfig().TempFactor(tempC)
}

func ApplyElapsedDegradationAt(p *Plant, from, to time.Time, tempC float64) {
	DefaultConfig().ApplyElapsedDegradationAt(p, from, to, tempC)
}

type PrecipitationEffect struct {
	ShiftHours float64
	WaterAdded float64
}

func ApplyPrecipitation(p *Plant, precipMM, tempC float64) (bool, PrecipitationEffect) {
	return ApplyPrecipitationAt(p, precipMM, tempC, time.Now())
}

func ApplyPrecipitationAt(p *Plant, precipMM, tempC float64, at time.Time) (bool, PrecipitationEffect) {
	return DefaultConfig().ApplyPrecipitationAt(p, precipMM, tempC, at)
}

func WaterPlant(p *Plant) {
	p.WaterLevel = 100
	p.ConsecutivePostpones = 0
	p.State = StateNormal
	p.LastUpdatedAt = time.Now()
	DefaultConfig().RefreshPlantState(p)
}

func PostponeWatering(p *Plant) (suggestInterval bool) {
	return DefaultConfig().PostponeWatering(p)
}

func RefreshPlantState(p *Plant) {
	DefaultConfig().RefreshPlantState(p)
}

func FormatNextAction(p *Plant) string {
	return DefaultConfig().FormatNextAction(p)
}

func FormatNextActionWithConfig(p *Plant, cfg ModelConfig) string {
	return NormalizeModelConfig(cfg).FormatNextAction(p)
}

func RefreshPlantStateWithConfig(p *Plant, cfg ModelConfig) {
	NormalizeModelConfig(cfg).RefreshPlantState(p)
}

func PostponeWateringWithConfig(p *Plant, cfg ModelConfig) (suggestInterval bool) {
	return NormalizeModelConfig(cfg).PostponeWatering(p)
}

func ApplyPrecipitationWithConfig(p *Plant, precipMM, tempC float64, cfg ModelConfig) (bool, PrecipitationEffect) {
	return NormalizeModelConfig(cfg).ApplyPrecipitationAt(p, precipMM, tempC, time.Now())
}

func ApplyElapsedDegradationWithConfig(p *Plant, from, to time.Time, tempC float64, cfg ModelConfig) {
	NormalizeModelConfig(cfg).ApplyElapsedDegradationAt(p, from, to, tempC)
}
