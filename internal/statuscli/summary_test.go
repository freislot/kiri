package statuscli

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kiri/internal/db"
	"kiri/internal/i18n"
	"kiri/internal/model"
)

func clearPlants(t *testing.T, store *db.Store) {
	t.Helper()
	plants, err := store.ListPlants()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, p := range plants {
		if err := store.DeletePlant(p.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
	}
}

func TestPrintSummary_OK(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()
	clearPlants(t, store)

	if err := store.SetLanguage(i18n.RU); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	now := time.Now()
	p := model.Plant{
		Name: "Ромашка", Location: "room", BaseIntervalDays: 7,
		IsOutdoor: false, WaterLevel: 100, State: model.StateNormal,
		LastUpdatedAt: now, CreatedAt: now,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	var buf bytes.Buffer
	if err := PrintSummary(&buf, store); err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "🌿 kiri: ok" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintSummary_NeedsWater(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()
	clearPlants(t, store)

	if err := store.SetLanguage(i18n.RU); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	now := time.Now()
	plants := []model.Plant{
		{Name: "Мята", Location: "garden", BaseIntervalDays: 7, IsOutdoor: true, WaterLevel: 0, State: model.StateWaterNow, LastUpdatedAt: now, CreatedAt: now},
		{Name: "Базилик", Location: "garden", BaseIntervalDays: 7, IsOutdoor: true, WaterLevel: 0, State: model.StateWaterNow, LastUpdatedAt: now, CreatedAt: now},
		{Name: "Ромашка", Location: "room", BaseIntervalDays: 7, IsOutdoor: false, WaterLevel: 100, State: model.StateNormal, LastUpdatedAt: now, CreatedAt: now},
	}
	for i := range plants {
		if err := store.SavePlant(&plants[i]); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := PrintSummary(&buf, store); err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "💧 kiri: 2 требуют полива!" {
		t.Fatalf("got %q", got)
	}
}

func TestPrintSummary_English(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()
	clearPlants(t, store)

	if err := store.SetLanguage(i18n.EN); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	now := time.Now()
	p := model.Plant{
		Name: "Mint", Location: "garden", BaseIntervalDays: 7,
		IsOutdoor: true, WaterLevel: 0, State: model.StateWaterNow,
		LastUpdatedAt: now, CreatedAt: now,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	var buf bytes.Buffer
	if err := PrintSummary(&buf, store); err != nil {
		t.Fatalf("summary: %v", err)
	}
	if got := strings.TrimSpace(buf.String()); got != "💧 kiri: 1 needs watering!" {
		t.Fatalf("got %q", got)
	}
}
