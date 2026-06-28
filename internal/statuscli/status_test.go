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

func TestMoistureBar(t *testing.T) {
	if got := moistureBar(100); got != "[██████████]" {
		t.Fatalf("full bar = %q", got)
	}
	if got := moistureBar(0); got != "[░░░░░░░░░░]" {
		t.Fatalf("empty bar = %q", got)
	}
	if got := moistureBar(40); got != "[████░░░░░░]" {
		t.Fatalf("40%% bar = %q, want [████░░░░░░]", got)
	}
}

func TestWateringLabel(t *testing.T) {
	now := time.Date(2026, 6, 10, 12, 0, 0, 0, time.UTC)
	c := i18n.New(i18n.RU)

	urgent := model.Plant{BaseIntervalDays: 7, WaterLevel: 0, State: model.StateWaterNow}
	if got := wateringLabel(c, urgent, now, model.DefaultConfig()); got != "СРОЧНО!" {
		t.Fatalf("urgent = %q", got)
	}

	healthy := model.Plant{BaseIntervalDays: 10, WaterLevel: 100, State: model.StateNormal}
	if got := wateringLabel(c, healthy, now, model.DefaultConfig()); got != "Уже полито" {
		t.Fatalf("healthy = %q", got)
	}

	soon := model.Plant{BaseIntervalDays: 7, WaterLevel: 40, State: model.StateNormal}
	if got := wateringLabel(c, soon, now, model.DefaultConfig()); !strings.HasPrefix(got, "Через ") {
		t.Fatalf("scheduled = %q, want Через ...", got)
	}
}

func TestPrint_RussianTable(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	if err := store.SetLanguage(i18n.RU); err != nil {
		t.Fatalf("set lang: %v", err)
	}

	now := time.Now()
	plants := []model.Plant{
		{Name: "Флоренция", Location: "garden", BaseIntervalDays: 7, IsOutdoor: true, WaterLevel: 40, State: model.StateNormal, LastUpdatedAt: now, CreatedAt: now},
		{Name: "Ромашка", Location: "room", BaseIntervalDays: 7, IsOutdoor: false, WaterLevel: 100, State: model.StateNormal, LastUpdatedAt: now, CreatedAt: now},
		{Name: "Мята", Location: "garden", BaseIntervalDays: 7, IsOutdoor: true, WaterLevel: 0, State: model.StateWaterNow, LastUpdatedAt: now, CreatedAt: now},
	}
	for i := range plants {
		if err := store.SavePlant(&plants[i]); err != nil {
			t.Fatalf("save: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := Print(&buf, store); err != nil {
		t.Fatalf("print: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"РАСТЕНИЕ", "СТАТУС", "ПОЛИВ",
		"Флоренция (улица)", "[████░░░░░░]",
		"Ромашка (комната)", "[██████████]", "Уже полито",
		"Мята (улица)", "[░░░░░░░░░░]", "СРОЧНО!",
	} {
		if !strings.Contains(out, want) {
			t.Fatalf("output missing %q:\n%s", want, out)
		}
	}
}

func TestPrint_Empty(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "empty.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	plants, err := store.ListPlants()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	for _, p := range plants {
		if err := store.DeletePlant(p.ID); err != nil {
			t.Fatalf("delete: %v", err)
		}
	}

	var buf bytes.Buffer
	if err := Print(&buf, store); err != nil {
		t.Fatalf("print: %v", err)
	}
	if !strings.Contains(buf.String(), "Растений пока нет.") {
		t.Fatalf("output = %q", buf.String())
	}
}
