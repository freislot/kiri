package model

import "time"

type Season int

const (
	SeasonWinter Season = iota
	SeasonSpring
	SeasonSummer
	SeasonAutumn
)

const BaseRainShiftHours = defaultRainBaseShiftHours

func CurrentSeason() Season {
	return SeasonFromMonth(time.Now().Month())
}

func SeasonFromMonth(m time.Month) Season {
	switch m {
	case time.December, time.January, time.February:
		return SeasonWinter
	case time.March, time.April, time.May:
		return SeasonSpring
	case time.June, time.July, time.August:
		return SeasonSummer
	default:
		return SeasonAutumn
	}
}

func OutdoorSeasonCoefficient(m time.Month) float64 {
	return DefaultConfig().OutdoorSeasonCoefficient(m)
}

func RainShiftFactor(m time.Month) float64 {
	return DefaultConfig().RainShiftFactor(m)
}

func SeasonCoefficient(p *Plant, t time.Time) float64 {
	return DefaultConfig().SeasonCoefficient(p, t)
}

func RainShiftHours(p *Plant, t time.Time) float64 {
	return DefaultConfig().RainShiftHours(p, t)
}

func (s Season) Factor() float64 {
	switch s {
	case SeasonSummer:
		return OutdoorSeasonCoefficient(time.June)
	case SeasonAutumn:
		return OutdoorSeasonCoefficient(time.September)
	case SeasonSpring:
		return OutdoorSeasonCoefficient(time.March)
	default:
		return OutdoorSeasonCoefficient(time.January)
	}
}
