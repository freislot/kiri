package db

import (
	"path/filepath"
	"testing"
	"time"

	"kiri/internal/model"
)

func TestTaskDoneLifecycle(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

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
	done, err := store.IsTaskDoneOnDate(p.ID, day)
	if err != nil {
		t.Fatalf("is done before mark: %v", err)
	}
	if done {
		t.Fatalf("expected not done before mark")
	}

	if err := store.MarkTaskDone(p.ID, day, "task done"); err != nil {
		t.Fatalf("mark done: %v", err)
	}
	done, err = store.IsTaskDoneOnDate(p.ID, day)
	if err != nil {
		t.Fatalf("is done after mark: %v", err)
	}
	if !done {
		t.Fatalf("expected done after mark")
	}

	if err := store.UnmarkTaskDone(p.ID, day); err != nil {
		t.Fatalf("unmark done: %v", err)
	}
	done, err = store.IsTaskDoneOnDate(p.ID, day)
	if err != nil {
		t.Fatalf("is done after unmark: %v", err)
	}
	if done {
		t.Fatalf("expected not done after unmark")
	}
}
