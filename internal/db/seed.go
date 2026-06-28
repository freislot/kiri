package db

import (
	"time"

	"kiri/internal/i18n"
	"kiri/internal/model"
)

func seed(s *Store) error {
	now := time.Now()
	june21 := time.Date(now.Year(), time.June, 21, 12, 0, 0, 0, now.Location())
	cat := i18n.New(i18n.RU)

	plants := []model.Plant{
		{
			Name:             "Ficus",
			Location:         "living_room",
			BaseIntervalDays: 10,
			WaterLevel:       100,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Spider Plant",
			Location:         "living_room",
			BaseIntervalDays: 7,
			WaterLevel:       85,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Peace Lily",
			Location:         "office",
			BaseIntervalDays: 5,
			WaterLevel:       65,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Aloe",
			Location:         "office",
			BaseIntervalDays: 14,
			WaterLevel:       47,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Begonia",
			Location:         "living_room",
			BaseIntervalDays: 4,
			WaterLevel:       32,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Rosemary",
			Location:         "terrace",
			BaseIntervalDays: 6,
			IsOutdoor:        true,
			WaterLevel:       18,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Cactus",
			Location:         "office",
			BaseIntervalDays: 21,
			WaterLevel:       7,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Monstera",
			Location:         "office",
			BaseIntervalDays: 5,
			WaterLevel:       0,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
		{
			Name:             "Juniper",
			Location:         "terrace",
			BaseIntervalDays: 7,
			IsOutdoor:        true,
			WaterLevel:       55,
			State:            model.StateShiftedByRain,
			LastUpdatedAt:    now,
			CreatedAt:        now,
		},
	}

	for i := range plants {
		model.RefreshPlantState(&plants[i])
		if err := s.SavePlant(&plants[i]); err != nil {
			return err
		}
	}

	logs := []struct {
		plantIdx  int
		eventType string
		message   string
		at        time.Time
	}{
		{0, "watered", cat.LogWateredNormal(), june21},
		{1, "note", "Демо: 85% влажности", now},
		{2, "note", "Демо: 65% влажности", now},
		{3, "note", "Демо: 47% влажности", now},
		{4, "note", "Демо: 32% влажности", now},
		{5, "note", "Демо: 18% влажности", now},
		{6, "note", "Демо: 7% влажности", now},
		{7, "alert", "Критическая жажда — ПОЛИТЬ СЕГОДНЯ!", now},
		{8, "rain", "Полит дождём — сильные осадки (12мм)", now},
		{8, "weather", "Перенесён из-за дождя: +20% воды", now},
	}

	for _, l := range logs {
		if err := s.addCareLogAt(plants[l.plantIdx].ID, l.eventType, l.message, l.at); err != nil {
			return err
		}
	}

	return nil
}
