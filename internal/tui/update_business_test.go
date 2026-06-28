package tui

import (
	"path/filepath"
	"strings"
	"testing"
	"time"

	"kiri/internal/db"
	"kiri/internal/model"
	"kiri/internal/weather"
)

func TestWeatherDayKey(t *testing.T) {
	loc := weather.Location{
		Name:      "Moscow",
		Latitude:  55.756,
		Longitude: 37.617,
		Timezone:  "UTC",
	}
	key := weatherDayKey(loc)
	parts := strings.Split(key, "|")
	if len(parts) != 3 {
		t.Fatalf("key = %q, expected 3 parts", key)
	}
	if parts[1] != "moscow" {
		t.Fatalf("city part = %q, want moscow", parts[1])
	}
	if !strings.Contains(parts[2], "55.756") {
		t.Fatalf("coord part = %q", parts[2])
	}
	if _, err := time.Parse("2006-01-02", parts[0]); err != nil {
		t.Fatalf("date part invalid: %q", parts[0])
	}
}

func TestWeatherDayKey_DefaultCity(t *testing.T) {
	key := weatherDayKey(weather.Location{})
	if !strings.Contains(key, "default|") {
		t.Fatalf("empty location key = %q", key)
	}
}

func TestApplyElapsedDegradation(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(24 * time.Hour)
	p := model.Plant{
		BaseIntervalDays: 10,
		WaterLevel:       100,
		State:            model.StateNormal,
		IsOutdoor:        true,
		LastUpdatedAt:    start,
		CreatedAt:        start,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save plant: %v", err)
	}

	m := &Model{
		store:  store,
		plants: []model.Plant{p},
	}
	if err := m.applyElapsedDegradation(now); err != nil {
		t.Fatalf("apply degradation: %v", err)
	}
	if m.plants[0].WaterLevel >= 100 {
		t.Fatalf("water should decrease after 24h, got %.2f", m.plants[0].WaterLevel)
	}
	if !m.plants[0].LastUpdatedAt.Equal(now) {
		t.Fatalf("last updated = %v, want %v", m.plants[0].LastUpdatedAt, now)
	}

	saved, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if saved.WaterLevel != m.plants[0].WaterLevel {
		t.Fatalf("persisted level %.2f != memory %.2f", saved.WaterLevel, m.plants[0].WaterLevel)
	}
}

func TestApplyElapsedDegradation_SkipsNoTimeElapsed(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	now := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	p := model.Plant{
		BaseIntervalDays: 7,
		WaterLevel:       80,
		State:            model.StateNormal,
		LastUpdatedAt:    now,
		CreatedAt:        now,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save plant: %v", err)
	}

	m := &Model{store: store, plants: []model.Plant{p}}
	if err := m.applyElapsedDegradation(now); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if m.plants[0].WaterLevel != 80 {
		t.Fatalf("no elapsed time should not change level, got %.2f", m.plants[0].WaterLevel)
	}
}

func TestParsePlantForm_Validation(t *testing.T) {
	m := &Model{lang: "ru"}

	m.plantForm = plantForm{name: "", location: "desk"}
	if _, _, _, ok := m.parsePlantForm(); ok {
		t.Fatal("empty name should fail")
	}

	m.plantForm = plantForm{name: "Ficus", location: ""}
	if _, _, _, ok := m.parsePlantForm(); ok {
		t.Fatal("empty location should fail")
	}

	m.plantForm = plantForm{name: "Ficus", location: "desk", interval: "abc"}
	if _, _, _, ok := m.parsePlantForm(); ok {
		t.Fatal("invalid interval should fail")
	}

	m.plantForm = plantForm{name: "Ficus", location: "desk", interval: "0"}
	if _, _, _, ok := m.parsePlantForm(); ok {
		t.Fatal("zero interval should fail")
	}

	m.plantForm = plantForm{name: "Ficus", location: "desk", interval: ""}
	name, loc, days, ok := m.parsePlantForm()
	if !ok || name != "Ficus" || loc != "desk" || days != 7 {
		t.Fatalf("default interval: ok=%v name=%q loc=%q days=%d", ok, name, loc, days)
	}

	m.plantForm = plantForm{name: "  Rose  ", location: "  garden  ", interval: "14"}
	name, loc, days, ok = m.parsePlantForm()
	if !ok || name != "Rose" || loc != "garden" || days != 14 {
		t.Fatalf("trimmed valid form: ok=%v name=%q loc=%q days=%d", ok, name, loc, days)
	}
}

func TestApplyWeatherPrecipitation_OutdoorOnly(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	now := time.Now()
	outdoor := model.Plant{
		Name: "Rose", Location: "garden", BaseIntervalDays: 7,
		IsOutdoor: true, WaterLevel: 50, State: model.StateNormal,
		LastUpdatedAt: now, CreatedAt: now,
	}
	indoor := model.Plant{
		Name: "Ficus", Location: "room", BaseIntervalDays: 7,
		IsOutdoor: false, WaterLevel: 50, State: model.StateNormal,
		LastUpdatedAt: now, CreatedAt: now,
	}
	if err := store.SavePlant(&outdoor); err != nil {
		t.Fatalf("save outdoor: %v", err)
	}
	if err := store.SavePlant(&indoor); err != nil {
		t.Fatalf("save indoor: %v", err)
	}

	m := &Model{
		store:        store,
		lang:         "ru",
		weatherReady: true,
		weatherCity:  weather.Location{Name: "Test", Latitude: 55.0, Longitude: 37.0, Timezone: "UTC"},
		weatherNow: weather.Conditions{
			PrevDayPrecipKnown: true,
			PrevDayPrecipMM:    5,
			TemperatureC:       22,
		},
		plants: []model.Plant{outdoor, indoor},
	}

	_ = m.applyWeatherPrecipitation()

	if m.plants[0].State != model.StateShiftedByRain {
		t.Fatalf("outdoor state = %s, want shifted_by_rain", m.plants[0].State)
	}
	if m.plants[0].WaterLevel <= 50 {
		t.Fatalf("outdoor water should increase, got %.2f", m.plants[0].WaterLevel)
	}
	if m.plants[1].WaterLevel != 50 || m.plants[1].State != model.StateNormal {
		t.Fatalf("indoor plant should be unchanged: level=%.0f state=%s", m.plants[1].WaterLevel, m.plants[1].State)
	}

	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if cfg.WeatherPrecipAppliedDay == "" {
		t.Fatal("precip applied day should be saved")
	}
}

func TestApplyWeatherPrecipitation_SkipsWhenAlreadyApplied(t *testing.T) {
	path := filepath.Join(t.TempDir(), "data.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	now := time.Now()
	p := model.Plant{
		Name: "Rose", Location: "garden", BaseIntervalDays: 7,
		IsOutdoor: true, WaterLevel: 50, State: model.StateNormal,
		LastUpdatedAt: now, CreatedAt: now,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	city := weather.Location{Name: "Test", Latitude: 55.0, Longitude: 37.0, Timezone: "UTC"}
	m := &Model{
		store:                 store,
		lang:                  "ru",
		weatherReady:          true,
		weatherCity:           city,
		weatherAppliedDayKey:  weatherDayKey(city),
		weatherNow:            weather.Conditions{PrevDayPrecipKnown: true, PrevDayPrecipMM: 8, TemperatureC: 20},
		plants:                []model.Plant{p},
	}

	if cmd := m.applyWeatherPrecipitation(); cmd != nil {
		t.Fatal("should skip when already applied today")
	}
	if m.plants[0].WaterLevel != 50 {
		t.Fatalf("level should not change, got %.2f", m.plants[0].WaterLevel)
	}
}

func testModel(t *testing.T) *db.Store {
	t.Helper()
	path := filepath.Join(t.TempDir(), "data.db")
	store, err := db.Open(path)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	t.Cleanup(func() { _ = store.Close() })
	return store
}

func TestWaterSelected_PersistsAndLogs(t *testing.T) {
	store := testModel(t)
	now := time.Now().UTC().Truncate(time.Second)
	p := model.Plant{
		Name:             "Monstera",
		Location:         "room",
		BaseIntervalDays: 7,
		WaterLevel:       15,
		State:            model.StateWaterNow,
		LastUpdatedAt:    now,
		CreatedAt:        now,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	m := &Model{
		store:       store,
		lang:        "ru",
		plants:      []model.Plant{p},
		cursorRow:   0,
		cursorCol:   0,
		visRows:     2,
		visCols:     3,
		splashPhase: splashPhaseDone,
	}
	m.waterSelected()

	if m.plants[0].WaterLevel != 100 || m.plants[0].State != model.StateNormal {
		t.Fatalf("memory plant = level %.0f state %s", m.plants[0].WaterLevel, m.plants[0].State)
	}
	if m.status == "" {
		t.Fatal("expected status message after watering")
	}

	saved, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if saved.WaterLevel != 100 {
		t.Fatalf("persisted level = %.0f, want 100", saved.WaterLevel)
	}

	entries, err := store.CareLogForPlant(p.ID, 5)
	if err != nil {
		t.Fatalf("care log: %v", err)
	}
	if len(entries) == 0 || entries[0].EventType != "watered" {
		t.Fatalf("care log = %+v, want watered entry", entries)
	}
}

func TestPostponeSelected_UsesCustomModelConfig(t *testing.T) {
	store := testModel(t)
	now := time.Now()
	p := model.Plant{
		Name:             "Ficus",
		Location:         "room",
		BaseIntervalDays: 7,
		WaterLevel:       40,
		State:            model.StateNormal,
		LastUpdatedAt:    now,
		CreatedAt:        now,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	cfg := model.NormalizeModelConfig(model.ModelConfig{
		FallbackTempC:        22,
		PostponeBoostPercent: 12,
		PostponeSuggestAfter: 3,
		Season:               model.DefaultConfig().Season,
		Rain:                 model.DefaultConfig().Rain,
		Temp:                 model.DefaultConfig().Temp,
	})

	m := &Model{
		store:       store,
		lang:        "ru",
		modelCfg:    cfg,
		plants:      []model.Plant{p},
		cursorRow:   0,
		cursorCol:   0,
		visRows:     2,
		visCols:     3,
		splashPhase: splashPhaseDone,
	}
	m.postponeSelected()

	if m.plants[0].WaterLevel != 52 {
		t.Fatalf("water level = %.0f, want 52 with +12%% boost", m.plants[0].WaterLevel)
	}
	if m.plants[0].ConsecutivePostpones != 1 {
		t.Fatalf("postpones = %d, want 1", m.plants[0].ConsecutivePostpones)
	}

	saved, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if saved.WaterLevel != 52 {
		t.Fatalf("persisted level = %.0f, want 52", saved.WaterLevel)
	}
}

func TestApplyElapsedDegradation_SkipsPersistWhenMinorChange(t *testing.T) {
	store := testModel(t)
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	now := start.Add(5 * time.Minute)
	p := model.Plant{
		Name:             "Slow",
		Location:         "room",
		BaseIntervalDays: 365,
		WaterLevel:       100,
		State:            model.StateNormal,
		LastUpdatedAt:    start,
		CreatedAt:        start,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	m := &Model{store: store, plants: []model.Plant{p}}
	if err := m.applyElapsedDegradation(now); err != nil {
		t.Fatalf("apply: %v", err)
	}

	saved, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if saved.WaterLevel != 100 {
		t.Fatalf("minor change within 30m should not persist: db level = %.3f", saved.WaterLevel)
	}
	if m.plants[0].WaterLevel >= 100 {
		t.Fatal("in-memory level should still decrease")
	}
}

func TestApplyElapsedDegradation_PersistsAfterInterval(t *testing.T) {
	store := testModel(t)
	start := time.Date(2026, 7, 1, 12, 0, 0, 0, time.UTC)
	now := start.Add(31 * time.Minute)
	p := model.Plant{
		Name:             "Slow",
		Location:         "room",
		BaseIntervalDays: 365,
		WaterLevel:       100,
		State:            model.StateNormal,
		LastUpdatedAt:    start,
		CreatedAt:        start,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	m := &Model{store: store, plants: []model.Plant{p}}
	if err := m.applyElapsedDegradation(now); err != nil {
		t.Fatalf("apply: %v", err)
	}

	saved, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if saved.WaterLevel != m.plants[0].WaterLevel {
		t.Fatalf("persist after 30m: db=%.3f memory=%.3f", saved.WaterLevel, m.plants[0].WaterLevel)
	}
	if !saved.LastUpdatedAt.Equal(now) {
		t.Fatalf("last_updated = %v, want %v", saved.LastUpdatedAt, now)
	}
}

func TestApplyElapsedDegradation_PersistsOnStateChange(t *testing.T) {
	store := testModel(t)
	start := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	now := start.Add(2 * time.Hour)
	p := model.Plant{
		Name:             "Thirsty",
		Location:         "garden",
		BaseIntervalDays: 1,
		IsOutdoor:        true,
		WaterLevel:       2,
		State:            model.StateNormal,
		LastUpdatedAt:    start,
		CreatedAt:        start,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	m := &Model{store: store, plants: []model.Plant{p}}
	if err := m.applyElapsedDegradation(now); err != nil {
		t.Fatalf("apply: %v", err)
	}
	if m.plants[0].State != model.StateWaterNow {
		t.Fatalf("state = %s, want water_now", m.plants[0].State)
	}

	saved, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if saved.State != model.StateWaterNow {
		t.Fatalf("persisted state = %s, want water_now", saved.State)
	}
}

func TestToggleSelectedTask_WatersPlant(t *testing.T) {
	store := testModel(t)
	now := time.Now()
	p := model.Plant{
		Name:             "Monstera",
		Location:         "room",
		BaseIntervalDays: 7,
		WaterLevel:       0,
		State:            model.StateWaterNow,
		LastUpdatedAt:    now,
		CreatedAt:        now,
	}
	if err := store.SavePlant(&p); err != nil {
		t.Fatalf("save: %v", err)
	}

	m := &Model{
		store:          store,
		lang:           "ru",
		plants:         []model.Plant{p},
		calYear:        now.Year(),
		calMonth:       int(now.Month()),
		calSelectedDay: now.Day(),
		calFocus:       calFocusTasks,
		taskCursor:     0,
		splashPhase:    splashPhaseDone,
	}

	if len(m.tasksForDate(m.calSelectedDate())) == 0 {
		t.Fatal("expected watering task for today")
	}

	m.toggleSelectedTask()

	if m.plants[0].WaterLevel != 100 {
		t.Fatalf("plant level = %.0f, want 100", m.plants[0].WaterLevel)
	}
	saved, err := store.GetPlant(p.ID)
	if err != nil {
		t.Fatalf("get plant: %v", err)
	}
	if saved.WaterLevel != 100 {
		t.Fatalf("persisted level = %.0f, want 100", saved.WaterLevel)
	}
	done, err := store.IsTaskDoneOnDate(p.ID, dateOnly(now))
	if err != nil {
		t.Fatalf("is done: %v", err)
	}
	if !done {
		t.Fatal("task should be marked done for today")
	}
}

