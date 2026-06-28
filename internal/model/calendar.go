package model

import "time"

func WateringDueDate(p Plant, now time.Time) time.Time {
	return DefaultConfig().WateringDueDate(p, now)
}

func WateringDueDateWithConfig(p Plant, now time.Time, cfg ModelConfig) time.Time {
	return NormalizeModelConfig(cfg).WateringDueDate(p, now)
}

func IsDueOnDate(p Plant, day, now time.Time) (due bool, urgent bool) {
	return DefaultConfig().IsDueOnDate(p, day, now)
}

func IsDueOnDateWithConfig(p Plant, day, now time.Time, cfg ModelConfig) (due bool, urgent bool) {
	return NormalizeModelConfig(cfg).IsDueOnDate(p, day, now)
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func SeasonFromTime(t time.Time) Season {
	return SeasonFromMonth(t.Month())
}
