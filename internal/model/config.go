package model

import (
	"fmt"
	"math"
	"time"
)

type ModelConfig struct {
	FallbackTempC        float64
	WaterSoonPercent     float64
	PostponeBoostPercent float64
	PostponeSuggestAfter int
	Season               SeasonCoeffs
	Rain                 RainCoeffs
	Temp                 TempCoeffs
}

type SeasonCoeffs struct {
	Winter float64
	Spring float64
	Summer float64
	Autumn float64
}

type RainCoeffs struct {
	LightMM        float64
	HeavyMM        float64
	BaseShiftHours float64
}

type TempCoeffs struct {
	CoolThresholdC float64
	CoolFactor     float64
	HotThresholdC  float64
	HotSlope       float64
}

const (
	defaultFallbackTempC        = 22.0
	defaultWaterSoonPercent     = 20.0
	defaultPostponeBoostPercent = 15.0
	defaultPostponeSuggestAfter = 2
	defaultRainLightMM          = 1.0
	defaultRainHeavyMM          = 10.0
	defaultRainBaseShiftHours   = 24.0
	defaultTempCoolThresholdC   = 20.0
	defaultTempCoolFactor       = 0.8
	defaultTempHotThresholdC    = 27.0
	defaultTempHotSlope         = 0.05
)

func DefaultConfig() ModelConfig {
	return ModelConfig{
		FallbackTempC:        defaultFallbackTempC,
		WaterSoonPercent:     defaultWaterSoonPercent,
		PostponeBoostPercent: defaultPostponeBoostPercent,
		PostponeSuggestAfter: defaultPostponeSuggestAfter,
		Season: SeasonCoeffs{
			Winter: 0.5,
			Spring: 1.0,
			Summer: 1.5,
			Autumn: 0.75,
		},
		Rain: RainCoeffs{
			LightMM:        defaultRainLightMM,
			HeavyMM:        defaultRainHeavyMM,
			BaseShiftHours: defaultRainBaseShiftHours,
		},
		Temp: TempCoeffs{
			CoolThresholdC: defaultTempCoolThresholdC,
			CoolFactor:     defaultTempCoolFactor,
			HotThresholdC:  defaultTempHotThresholdC,
			HotSlope:       defaultTempHotSlope,
		},
	}
}

func NormalizeModelConfig(c ModelConfig) ModelConfig {
	d := DefaultConfig()

	if c.FallbackTempC < 1.0 || c.FallbackTempC > 40.0 {
		c.FallbackTempC = d.FallbackTempC
	}
	if c.WaterSoonPercent <= 0 || c.WaterSoonPercent > 100 {
		c.WaterSoonPercent = d.WaterSoonPercent
	}
	if c.PostponeBoostPercent <= 0 || c.PostponeBoostPercent > 100 {
		c.PostponeBoostPercent = d.PostponeBoostPercent
	}
	if c.PostponeSuggestAfter < 1 {
		c.PostponeSuggestAfter = d.PostponeSuggestAfter
	}

	c.Season = normalizeSeasonCoeffs(c.Season, d.Season)
	c.Rain = normalizeRainCoeffs(c.Rain, d.Rain)
	c.Temp = normalizeTempCoeffs(c.Temp, d.Temp)
	return c
}

func normalizeSeasonCoeffs(c, d SeasonCoeffs) SeasonCoeffs {
	if c.Winter <= 0 {
		c.Winter = d.Winter
	}
	if c.Spring <= 0 {
		c.Spring = d.Spring
	}
	if c.Summer <= 0 {
		c.Summer = d.Summer
	}
	if c.Autumn <= 0 {
		c.Autumn = d.Autumn
	}
	return c
}

func normalizeRainCoeffs(c, d RainCoeffs) RainCoeffs {
	if c.LightMM <= 0 {
		c.LightMM = d.LightMM
	}
	if c.HeavyMM <= c.LightMM {
		c.HeavyMM = d.HeavyMM
	}
	if c.BaseShiftHours <= 0 {
		c.BaseShiftHours = d.BaseShiftHours
	}
	return c
}

func normalizeTempCoeffs(c, d TempCoeffs) TempCoeffs {
	if c.CoolThresholdC <= 0 {
		c.CoolThresholdC = d.CoolThresholdC
	}
	if c.CoolFactor <= 0 {
		c.CoolFactor = d.CoolFactor
	}
	if c.HotThresholdC <= c.CoolThresholdC {
		c.HotThresholdC = d.HotThresholdC
	}
	if c.HotSlope <= 0 {
		c.HotSlope = d.HotSlope
	}
	return c
}

func (c ModelConfig) seasonReferenceK() float64 {
	return c.Season.Spring
}

func (c ModelConfig) TempFactor(tempC float64) float64 {
	switch {
	case tempC <= c.Temp.CoolThresholdC:
		return c.Temp.CoolFactor
	case tempC <= c.Temp.HotThresholdC:
		return 1.0
	default:
		return 1.0 + (tempC-c.Temp.HotThresholdC)*c.Temp.HotSlope
	}
}

func (c ModelConfig) OutdoorSeasonCoefficient(m time.Month) float64 {
	switch SeasonFromMonth(m) {
	case SeasonSummer:
		return c.Season.Summer
	case SeasonAutumn:
		return c.Season.Autumn
	case SeasonSpring:
		return c.Season.Spring
	default:
		return c.Season.Winter
	}
}

func (c ModelConfig) RainShiftFactor(m time.Month) float64 {
	k := c.OutdoorSeasonCoefficient(m)
	ref := c.seasonReferenceK()
	if k <= 0 {
		k = ref
	}
	factor := ref / k
	if SeasonFromMonth(m) == SeasonWinter && factor > 1.0 {
		factor = 1.0
	}
	return factor
}

func (c ModelConfig) SeasonCoefficient(p *Plant, t time.Time) float64 {
	if p == nil || !p.IsOutdoor {
		return 1.0
	}
	return c.OutdoorSeasonCoefficient(t.Month())
}

func (c ModelConfig) RainShiftHours(p *Plant, t time.Time) float64 {
	if p == nil || !p.IsOutdoor {
		return c.Rain.BaseShiftHours
	}
	return c.Rain.BaseShiftHours * c.RainShiftFactor(t.Month())
}

func (c ModelConfig) DeltaV(baseIntervalDays int, kSeason, tempC float64) float64 {
	if baseIntervalDays <= 0 {
		return 0
	}
	vBase := 100.0 / float64(baseIntervalDays)
	return vBase * kSeason * c.TempFactor(tempC)
}

func (c ModelConfig) DeltaVForPlant(p *Plant, t time.Time, tempC float64) float64 {
	return c.DeltaV(p.BaseIntervalDays, c.SeasonCoefficient(p, t), tempC)
}

func (c ModelConfig) ApplyElapsedDegradationAt(p *Plant, from, to time.Time, tempC float64) {
	if p.State == StateNew || p.BaseIntervalDays <= 0 {
		return
	}
	if !to.After(from) {
		return
	}
	cur := from
	for cur.Before(to) {
		dayStart := time.Date(cur.Year(), cur.Month(), cur.Day(), 0, 0, 0, 0, cur.Location())
		dayEnd := dayStart.AddDate(0, 0, 1)
		segEnd := to
		if dayEnd.Before(segEnd) {
			segEnd = dayEnd
		}
		fracDay := segEnd.Sub(cur).Hours() / 24.0
		if fracDay > 0 {
			delta := c.DeltaVForPlant(p, cur, tempC) * fracDay
			p.WaterLevel = math.Max(0, p.WaterLevel-delta)
		}
		cur = segEnd
	}
	p.LastUpdatedAt = to
	c.RefreshPlantState(p)
}

func (c ModelConfig) ApplyPrecipitationAt(p *Plant, precipMM, tempC float64, at time.Time) (bool, PrecipitationEffect) {
	if !p.IsOutdoor || precipMM < c.Rain.LightMM {
		return false, PrecipitationEffect{}
	}

	switch {
	case precipMM > c.Rain.HeavyMM:
		p.WaterLevel = 100
		p.State = StateNormal
		p.ConsecutivePostpones = 0
		p.LastUpdatedAt = at
		c.RefreshPlantState(p)
		return true, PrecipitationEffect{}
	case precipMM >= c.Rain.LightMM:
		shiftHours := c.RainShiftHours(p, at)
		vBase := 100.0 / float64(p.BaseIntervalDays)
		waterAdded := vBase * c.TempFactor(tempC) * (shiftHours / 24.0)
		p.WaterLevel = math.Min(100, p.WaterLevel+waterAdded)
		p.State = StateShiftedByRain
		p.LastUpdatedAt = at
		c.RefreshPlantState(p)
		return true, PrecipitationEffect{ShiftHours: shiftHours, WaterAdded: waterAdded}
	default:
		return false, PrecipitationEffect{}
	}
}

func (c ModelConfig) PostponeWatering(p *Plant) (suggestInterval bool) {
	p.ConsecutivePostpones++
	p.WaterLevel = math.Min(100, p.WaterLevel+c.PostponeBoostPercent)
	p.LastUpdatedAt = time.Now()
	c.RefreshPlantState(p)
	return p.ConsecutivePostpones >= c.PostponeSuggestAfter
}

func (c ModelConfig) RefreshPlantState(p *Plant) {
	switch {
	case p.BaseIntervalDays <= 0 || p.State == StateNew:
		p.State = StateNew
		p.NextAction = "Configure interval"
	case p.WaterLevel <= 0:
		p.State = StateWaterNow
		p.NextAction = "WATER NOW!"
	default:
		if p.State != StateShiftedByRain {
			p.State = StateNormal
		}
		p.NextAction = c.FormatNextAction(p)
	}
}

func (c ModelConfig) FormatNextAction(p *Plant) string {
	if p.WaterLevel <= c.WaterSoonPercent {
		return "Water soon"
	}
	days := int(math.Ceil(p.WaterLevel / c.DeltaVForPlant(p, time.Now(), c.FallbackTempC)))
	if days <= 1 {
		return "Water tomorrow"
	}
	return fmt.Sprintf("Water in %d days", days)
}

func (c ModelConfig) WateringDueDate(p Plant, now time.Time) time.Time {
	if p.State == StateWaterNow || p.WaterLevel <= 0 {
		return dateOnly(now)
	}
	delta := c.DeltaVForPlant(&p, now, c.FallbackTempC)
	if delta <= 0 {
		return dateOnly(now.AddDate(0, 0, 7))
	}
	days := int(math.Ceil(p.WaterLevel / delta))
	if days < 1 {
		days = 1
	}
	return dateOnly(now.AddDate(0, 0, days))
}

func (c ModelConfig) IsDueOnDate(p Plant, day, now time.Time) (due bool, urgent bool) {
	d := dateOnly(day)
	today := dateOnly(now)

	if p.State == StateWaterNow || p.WaterLevel <= 0 {
		due := d.Equal(today)
		return due, due
	}

	dueDate := c.WateringDueDate(p, now)
	return d.Equal(dueDate), false
}
