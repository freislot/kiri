package statuscli

import (
	"bytes"
	"encoding/json"
	"path/filepath"
	"testing"
	"time"

	"kiri/internal/db"
	"kiri/internal/i18n"
	"kiri/internal/model"
)

func TestPrintJSON(t *testing.T) {
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
		{Name: "Ромашка", Location: "room", BaseIntervalDays: 7, IsOutdoor: false, WaterLevel: 100, State: model.StateNormal, LastUpdatedAt: now, CreatedAt: now},
	}
	for i := range plants {
		if err := store.SavePlant(&plants[i]); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := PrintJSON(&buf, store); err != nil {
		t.Fatalf("json: %v", err)
	}

	var out []plantJSON
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v\n%s", err, buf.String())
	}
	if len(out) != 2 {
		t.Fatalf("len = %d, want 2", len(out))
	}

	byName := make(map[string]plantJSON, len(out))
	for _, p := range out {
		byName[p.Name] = p
	}

	mint := byName["Мята"]
	if !mint.NeedsWatering || mint.State != model.StateWaterNow || mint.WaterLevel != 0 {
		t.Fatalf("mint = %+v", mint)
	}
	if mint.WateringDueDate == "" {
		t.Fatal("mint watering_due_date empty")
	}

	daisy := byName["Ромашка"]
	if daisy.NeedsWatering || daisy.IsOutdoor {
		t.Fatalf("daisy = %+v", daisy)
	}
	if daisy.DaysUntilWatering < 0 {
		t.Fatalf("daisy days_until_watering = %d", daisy.DaysUntilWatering)
	}
}

func TestPrintJSON_Empty(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()
	clearPlants(t, store)

	var buf bytes.Buffer
	if err := PrintJSON(&buf, store); err != nil {
		t.Fatalf("json: %v", err)
	}

	var out []plantJSON
	if err := json.Unmarshal(buf.Bytes(), &out); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if len(out) != 0 {
		t.Fatalf("len = %d, want 0", len(out))
	}
}
