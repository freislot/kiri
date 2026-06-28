package db

import (
	"path/filepath"
	"testing"

	"kiri/internal/model"
)

func TestOpenSeedsNineDemoPlants(t *testing.T) {
	store, err := Open(filepath.Join(t.TempDir(), "fresh.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	plants, err := store.ListPlants()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(plants) != 9 {
		t.Fatalf("seeded plants = %d, want 9", len(plants))
	}

	var outdoor, waterNow, shifted int
	for _, p := range plants {
		if p.IsOutdoor {
			outdoor++
		}
		if p.State == model.StateWaterNow {
			waterNow++
		}
		if p.State == model.StateShiftedByRain {
			shifted++
		}
	}
	if outdoor < 2 {
		t.Fatalf("expected at least 2 outdoor demo plants, got %d", outdoor)
	}
	if waterNow < 1 {
		t.Fatal("expected at least one water_now demo plant")
	}
	if shifted < 1 {
		t.Fatal("expected at least one shifted_by_rain demo plant")
	}

	entries, err := store.AllCareLog(20)
	if err != nil {
		t.Fatalf("care log: %v", err)
	}
	if len(entries) == 0 {
		t.Fatal("expected demo care log entries")
	}
}
