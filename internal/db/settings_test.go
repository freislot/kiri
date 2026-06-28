package db

import (
	"os"
	"path/filepath"
	"testing"

	"kiri/internal/i18n"
	"kiri/internal/model"
	"kiri/internal/weather"
)

func TestNormalizeWeatherRefreshMinutes(t *testing.T) {
	tests := []struct {
		in, want int
	}{
		{1, 1},
		{10, 10},
		{100, 100},
		{1440, 1440},
		{0, DefaultWeatherRefreshMinutesDefault},
		{15, 15},
		{2000, DefaultWeatherRefreshMinutesDefault},
	}
	for _, tt := range tests {
		if got := NormalizeWeatherRefreshMinutes(tt.in); got != tt.want {
			t.Fatalf("NormalizeWeatherRefreshMinutes(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestClampWeatherRefreshMinutes(t *testing.T) {
	tests := []struct {
		in, want int
	}{
		{100, 100},
		{0, minWeatherRefreshMinutes},
		{2000, maxWeatherRefreshMinutes},
	}
	for _, tt := range tests {
		if got := ClampWeatherRefreshMinutes(tt.in); got != tt.want {
			t.Fatalf("ClampWeatherRefreshMinutes(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestLoadSettingsCustomWeatherRefreshMinutes(t *testing.T) {
	store := openTestStore(t)
	path := store.settingsFilePath()
	content := []byte(`language = "ru"
weather_refresh_minutes = 100
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}
	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.WeatherRefreshMinutes != 100 {
		t.Fatalf("refresh = %d, want 100", cfg.WeatherRefreshMinutes)
	}
}

func TestNormalizeDefaultIntervalDays(t *testing.T) {
	tests := []struct {
		in, want int
	}{
		{7, 7},
		{1, 1},
		{365, 365},
		{0, DefaultIntervalDaysDefault},
		{500, DefaultIntervalDaysDefault},
	}
	for _, tt := range tests {
		if got := NormalizeDefaultIntervalDays(tt.in); got != tt.want {
			t.Fatalf("NormalizeDefaultIntervalDays(%d) = %d, want %d", tt.in, got, tt.want)
		}
	}
}

func TestNormalizeFallbackTempC(t *testing.T) {
	tests := []struct {
		in, want float64
	}{
		{22.0, 22.0},
		{1.0, 1.0},
		{40.0, 40.0},
		{0, DefaultFallbackTempC},
		{50, DefaultFallbackTempC},
	}
	for _, tt := range tests {
		if got := NormalizeFallbackTempC(tt.in); got != tt.want {
			t.Fatalf("NormalizeFallbackTempC(%v) = %v, want %v", tt.in, got, tt.want)
		}
	}
}

func TestParseBoolSetting(t *testing.T) {
	trues := []string{"1", "true", "TRUE", " yes ", "on"}
	for _, s := range trues {
		if !parseBoolSetting(s) {
			t.Fatalf("parseBoolSetting(%q) should be true", s)
		}
	}
	falses := []string{"0", "false", "no", "off", ""}
	for _, s := range falses {
		if parseBoolSetting(s) {
			t.Fatalf("parseBoolSetting(%q) should be false", s)
		}
	}
}

func TestLoadSettingsCreatesDefaults(t *testing.T) {
	store := openTestStore(t)
	path := store.settingsFilePath()
	_ = os.Remove(path)

	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if cfg.Language != i18n.RU {
		t.Fatalf("default language = %s, want ru", cfg.Language)
	}
	if cfg.WeatherRefreshMinutes != 10 {
		t.Fatalf("default refresh = %d, want 10", cfg.WeatherRefreshMinutes)
	}
	if cfg.DefaultIntervalDays != DefaultIntervalDaysDefault {
		t.Fatalf("default interval = %d, want %d", cfg.DefaultIntervalDays, DefaultIntervalDaysDefault)
	}
	if cfg.Model.FallbackTempC != DefaultFallbackTempC {
		t.Fatalf("default temp = %v, want %v", cfg.Model.FallbackTempC, DefaultFallbackTempC)
	}
	if cfg.Model.Season.Summer != 1.5 {
		t.Fatalf("default summer K = %v, want 1.5", cfg.Model.Season.Summer)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("settings file should be created: %v", err)
	}
}

func TestSettingsRoundTrip(t *testing.T) {
	store := openTestStore(t)

	city := weather.Location{
		Name:      "Berlin",
		Admin1:    "Berlin",
		Country:   "Germany",
		Latitude:  52.52,
		Longitude: 13.405,
		Timezone:  "Europe/Berlin",
	}
	if err := store.SetLanguage(i18n.EN); err != nil {
		t.Fatalf("set language: %v", err)
	}
	if err := store.SetAutoBackup(true); err != nil {
		t.Fatalf("set auto backup: %v", err)
	}
	if err := store.SetTransparentMode(true); err != nil {
		t.Fatalf("set transparent: %v", err)
	}
	if err := store.SetFastBoot(true); err != nil {
		t.Fatalf("set fast boot: %v", err)
	}
	if err := store.SetWeatherCity(city); err != nil {
		t.Fatalf("set weather city: %v", err)
	}
	if err := store.SetWeatherRefreshMinutes(30); err != nil {
		t.Fatalf("set refresh: %v", err)
	}
	if err := store.SetDefaultIntervalDays(14); err != nil {
		t.Fatalf("set default interval: %v", err)
	}
	if err := store.SetFallbackTempC(18.5); err != nil {
		t.Fatalf("set fallback temp: %v", err)
	}
	if err := store.SetWeatherPrecipAppliedDay("2026-06-10"); err != nil {
		t.Fatalf("set precip day: %v", err)
	}

	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("reload settings: %v", err)
	}
	if cfg.Language != i18n.EN {
		t.Fatalf("language = %s, want en", cfg.Language)
	}
	if !cfg.AutoBackup || !cfg.TransparentMode || !cfg.FastBoot {
		t.Fatal("boolean settings not persisted")
	}
	if cfg.WeatherCity.Name != "Berlin" {
		t.Fatalf("city = %q, want Berlin", cfg.WeatherCity.Name)
	}
	if cfg.WeatherRefreshMinutes != 30 {
		t.Fatalf("refresh = %d, want 30", cfg.WeatherRefreshMinutes)
	}
	if cfg.DefaultIntervalDays != 14 {
		t.Fatalf("default interval = %d, want 14", cfg.DefaultIntervalDays)
	}
	if cfg.Model.FallbackTempC != 18.5 {
		t.Fatalf("fallback temp = %v, want 18.5", cfg.Model.FallbackTempC)
	}
	if cfg.WeatherPrecipAppliedDay != "2026-06-10" {
		t.Fatalf("precip day = %q", cfg.WeatherPrecipAppliedDay)
	}

	enabled, err := store.GetAutoBackup()
	if err != nil || !enabled {
		t.Fatalf("GetAutoBackup: enabled=%v err=%v", enabled, err)
	}
}

func TestSettingsModelSectionRoundTrip(t *testing.T) {
	store := openTestStore(t)
	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	cfg.Model = model.NormalizeModelConfig(model.ModelConfig{
		FallbackTempC:        19,
		WaterSoonPercent:     25,
		PostponeBoostPercent: 12,
		PostponeSuggestAfter: 3,
		Season: model.SeasonCoeffs{
			Winter: 0.4,
			Spring: 1.1,
			Summer: 1.6,
			Autumn: 0.8,
		},
		Rain: model.RainCoeffs{
			LightMM:        1.5,
			HeavyMM:        12,
			BaseShiftHours: 20,
		},
		Temp: model.TempCoeffs{
			CoolThresholdC: 18,
			CoolFactor:     0.75,
			HotThresholdC:  28,
			HotSlope:       0.04,
		},
	})
	if err := store.saveSettingsLocked(cfg); err != nil {
		t.Fatalf("save: %v", err)
	}

	got, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("reload: %v", err)
	}
	if got.Model.FallbackTempC != 19 {
		t.Fatalf("model fallback = %v, want 19", got.Model.FallbackTempC)
	}
	if got.Model.Season.Summer != 1.6 {
		t.Fatalf("model summer = %v, want 1.6", got.Model.Season.Summer)
	}
	if got.Model.Rain.BaseShiftHours != 20 {
		t.Fatalf("model rain shift = %v, want 20", got.Model.Rain.BaseShiftHours)
	}
}

func TestSettingsLegacyTopLevelFallbackTempMigrated(t *testing.T) {
	store := openTestStore(t)
	path := store.settingsFilePath()
	content := []byte(`language = "ru"

fallback_temp_c = 17.5

[weather_city]
name = "Moscow"
admin1 = "Moscow"
country = "Russia"
latitude = 55.7558
longitude = 37.6173
timezone = "Europe/Moscow"
`)
	if err := os.WriteFile(path, content, 0o644); err != nil {
		t.Fatalf("write settings: %v", err)
	}

	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("load: %v", err)
	}
	if cfg.Model.FallbackTempC != 17.5 {
		t.Fatalf("legacy fallback temp = %v, want 17.5", cfg.Model.FallbackTempC)
	}
}

func TestMigrateLegacySettings(t *testing.T) {
	tmp := t.TempDir()
	path := filepath.Join(tmp, "data.db")
	store, err := Open(path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	defer store.Close()

	settingsPath := store.settingsFilePath()
	_ = os.Remove(settingsPath)

	_, err = store.db.Exec(`CREATE TABLE IF NOT EXISTS settings (key TEXT PRIMARY KEY, value TEXT NOT NULL)`)
	if err != nil {
		t.Fatalf("create legacy settings: %v", err)
	}
	legacy := map[string]string{
		settingLanguage:              "en",
		settingAutoBackup:            "true",
		settingTransparentMode:       "yes",
		settingFastBoot:              "1",
		settingWeatherRefreshMinutes: "5",
		settingWeatherPrecipAppliedDay: "2026-01-15",
		settingWeatherCity: `{"name":"Paris","admin1":"Paris","country":"France",` +
			`"latitude":48.8566,"longitude":2.3522,"timezone":"Europe/Paris"}`,
	}
	for k, v := range legacy {
		if _, err := store.db.Exec(`INSERT OR REPLACE INTO settings (key, value) VALUES (?, ?)`, k, v); err != nil {
			t.Fatalf("insert %s: %v", k, err)
		}
	}

	if err := store.migrateLegacySettings(); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("load after migrate: %v", err)
	}
	if cfg.Language != i18n.EN {
		t.Fatalf("language = %s, want en", cfg.Language)
	}
	if !cfg.AutoBackup || !cfg.TransparentMode || !cfg.FastBoot {
		t.Fatal("legacy booleans not migrated")
	}
	if cfg.WeatherRefreshMinutes != 5 {
		t.Fatalf("refresh = %d, want 5", cfg.WeatherRefreshMinutes)
	}
	if cfg.WeatherCity.Name != "Paris" {
		t.Fatalf("city = %q, want Paris", cfg.WeatherCity.Name)
	}
	if cfg.WeatherPrecipAppliedDay != "2026-01-15" {
		t.Fatalf("precip day = %q", cfg.WeatherPrecipAppliedDay)
	}

	var count int
	if err := store.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='settings'`).Scan(&count); err != nil {
		t.Fatalf("check table: %v", err)
	}
	if count != 0 {
		t.Fatal("legacy settings table should be dropped")
	}
}
