package db

import (
	"path/filepath"
	"testing"
	"time"

	"kiri/internal/model"
)

func openTestStore(t *testing.T) *Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestPlantCRUD(t *testing.T) {
	store := openTestStore(t)

	now := time.Now().UTC().Truncate(time.Second)
	p := &model.Plant{
		Name:             "Monstera",
		Location:         "living_room",
		BaseIntervalDays: 7,
		IsOutdoor:        true,
		WaterLevel:       72.5,
		State:            model.StateNormal,
		NextAction:       "Water in 3 days",
		LastUpdatedAt:    now,
		CreatedAt:        now,
	}
	if err := store.SavePlant(p); err != nil {
		t.Fatalf("insert plant: %v", err)
	}
	if p.ID == 0 {
		t.Fatal("insert should assign ID")
	}

	got, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if got.Name != p.Name || got.IsOutdoor != p.IsOutdoor || got.WaterLevel != p.WaterLevel {
		t.Fatalf("get mismatch: %+v", got)
	}

	p.WaterLevel = 40
	p.Name = "Monstera XL"
	if err := store.SavePlant(p); err != nil {
		t.Fatalf("update plant: %v", err)
	}
	got, err = store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get updated: %v", err)
	}
	if got.WaterLevel != 40 || got.Name != "Monstera XL" {
		t.Fatalf("update not persisted: %+v", got)
	}

	plants, err := store.ListPlants()
	if err != nil {
		t.Fatalf("list plants: %v", err)
	}
	if len(plants) == 0 {
		t.Fatal("expected seeded or inserted plants")
	}

	if err := store.DeletePlant(p.ID); err != nil {
		t.Fatalf("delete plant: %v", err)
	}
	if _, err := store.GetPlant(p.ID); err == nil {
		t.Fatal("expected error after delete")
	}
}

func TestCareLog(t *testing.T) {
	store := openTestStore(t)

	p := &model.Plant{
		Name:             "Test",
		Location:         "desk",
		BaseIntervalDays: 5,
		WaterLevel:       80,
		State:            model.StateNormal,
		LastUpdatedAt:    time.Now(),
		CreatedAt:        time.Now(),
	}
	if err := store.SavePlant(p); err != nil {
		t.Fatalf("save plant: %v", err)
	}

	if err := store.AddCareLog(p.ID, "watered", "test watering"); err != nil {
		t.Fatalf("add care log: %v", err)
	}
	if err := store.AddCareLog(p.ID, "note", "test note"); err != nil {
		t.Fatalf("add note: %v", err)
	}

	entries, err := store.CareLogForPlant(p.ID, 10)
	if err != nil {
		t.Fatalf("care log for plant: %v", err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
	if entries[0].PlantName != "Test" {
		t.Fatalf("plant name in log = %q", entries[0].PlantName)
	}

	all, err := store.AllCareLog(5)
	if err != nil {
		t.Fatalf("all care log: %v", err)
	}
	if len(all) < 2 {
		t.Fatalf("expected at least 2 entries in all log, got %d", len(all))
	}
}

func TestIsTaskDoneOnDate_WateredEvent(t *testing.T) {
	store := openTestStore(t)

	p := &model.Plant{
		Name:             "Test",
		Location:         "Desk",
		BaseIntervalDays: 7,
		WaterLevel:       50,
		State:            model.StateNormal,
		LastUpdatedAt:    time.Now(),
		CreatedAt:        time.Now(),
	}
	if err := store.SavePlant(p); err != nil {
		t.Fatalf("save plant: %v", err)
	}

	day := time.Date(2026, 6, 10, 0, 0, 0, 0, time.UTC)
	if err := store.addCareLogAt(p.ID, "watered", "manual water", day); err != nil {
		t.Fatalf("log watered: %v", err)
	}
	done, err := store.IsTaskDoneOnDate(p.ID, day)
	if err != nil {
		t.Fatalf("is done: %v", err)
	}
	if !done {
		t.Fatal("watered event should mark day as done")
	}
}
