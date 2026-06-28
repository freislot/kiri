package db

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"kiri/internal/i18n"
	"kiri/internal/model"
	"kiri/internal/weather"

	"github.com/pelletier/go-toml/v2"
)

const (
	settingLanguage                = "language"
	settingAutoBackup              = "auto_backup"
	settingTransparentMode         = "transparent_mode"
	settingFastBoot                = "fast_boot"
	settingWeatherCity             = "weather_city"
	settingWeatherRefreshMinutes   = "weather_refresh_minutes"
	settingWeatherPrecipAppliedDay = "weather_precip_applied_day"
)

const (
	DefaultIntervalDaysDefault        = 7
	DefaultFallbackTempC              = 22.0
	DefaultWeatherRefreshMinutesDefault = 10
	minDefaultIntervalDays            = 1
	maxDefaultIntervalDays            = 365
	minFallbackTempC                  = 1.0
	maxFallbackTempC                  = 40.0
	minWeatherRefreshMinutes          = 1
	maxWeatherRefreshMinutes          = 1440 // 24 hours
)

type modelSettingsFile struct {
	FallbackTempC        float64 `toml:"fallback_temp_c"`
	WaterSoonPercent     float64 `toml:"water_soon_percent"`
	PostponeBoostPercent float64 `toml:"postpone_boost_percent"`
	PostponeSuggestAfter int     `toml:"postpone_suggest_after"`
	Season               struct {
		Winter float64 `toml:"winter"`
		Spring float64 `toml:"spring"`
		Summer float64 `toml:"summer"`
		Autumn float64 `toml:"autumn"`
	} `toml:"season"`
	Rain struct {
		LightMM        float64 `toml:"light_mm"`
		HeavyMM        float64 `toml:"heavy_mm"`
		BaseShiftHours float64 `toml:"base_shift_hours"`
	} `toml:"rain"`
	Temp struct {
		CoolThresholdC float64 `toml:"cool_threshold_c"`
		CoolFactor     float64 `toml:"cool_factor"`
		HotThresholdC  float64 `toml:"hot_threshold_c"`
		HotSlope       float64 `toml:"hot_slope"`
	} `toml:"temp"`
}

type settingsFile struct {
	Language                string  `toml:"language"`
	AutoBackup              bool    `toml:"auto_backup"`
	TransparentMode         bool    `toml:"transparent_mode"`
	FastBoot                bool    `toml:"fast_boot"`
	DefaultIntervalDays     int     `toml:"default_interval_days"`
	FallbackTempC           float64 `toml:"fallback_temp_c"` // legacy top-level; migrated into [model]
	WeatherRefreshMinutes   int     `toml:"weather_refresh_minutes"`
	WeatherPrecipAppliedDay string  `toml:"weather_precip_applied_day"`
	Model                   modelSettingsFile `toml:"model"`
	WeatherCity             struct {
		Name      string  `toml:"name"`
		Admin1    string  `toml:"admin1"`
		Country   string  `toml:"country"`
		Latitude  float64 `toml:"latitude"`
		Longitude float64 `toml:"longitude"`
		Timezone  string  `toml:"timezone"`
	} `toml:"weather_city"`
}

type Settings struct {
	Language                i18n.Lang
	AutoBackup              bool
	TransparentMode         bool
	FastBoot                bool
	DefaultIntervalDays     int
	Model                   model.ModelConfig
	WeatherCity             weather.Location
	WeatherRefreshMinutes   int
	WeatherPrecipAppliedDay string
}

func (s Settings) FallbackTempC() float64 {
	return s.Model.FallbackTempC
}

func settingsTOMLHeader() string {
	return strings.TrimSpace(`
# Kiri settings file
# You can edit this file manually while the app is not running.
#
# language: "ru" or "en"
# auto_backup: true/false
# transparent_mode: true/false
# fast_boot: true/false
# default_interval_days: 1–365 (new plant default watering cycle)
# weather_refresh_minutes: 1–1440 (minutes between weather API requests)
# weather_precip_applied_day: YYYY-MM-DD (optional)
#
# [model] — watering model coefficients (optional; defaults match built-in constants)
# fallback_temp_c: 1.0–40.0 (°C when live weather is unavailable)
# water_soon_percent / postpone_boost_percent / postpone_suggest_after
# [model.season] winter/spring/summer/autumn
# [model.rain] light_mm / heavy_mm / base_shift_hours
# [model.temp] cool_threshold_c / cool_factor / hot_threshold_c / hot_slope
#
# [weather_city]
# name/admin1/country/timezone are strings
# latitude/longitude are decimal numbers
`) + "\n\n"
}

func defaultAppSettings() Settings {
	return Settings{
		Language:                i18n.RU,
		AutoBackup:              false,
		TransparentMode:         false,
		FastBoot:                false,
		DefaultIntervalDays:     DefaultIntervalDaysDefault,
		Model:                   model.DefaultConfig(),
		WeatherCity:             defaultWeatherCity(),
		WeatherRefreshMinutes:   DefaultWeatherRefreshMinutesDefault,
		WeatherPrecipAppliedDay: "",
	}
}

func modelConfigFromFile(file settingsFile) model.ModelConfig {
	cfg := model.DefaultConfig()
	m := file.Model

	if m.FallbackTempC != 0 {
		cfg.FallbackTempC = m.FallbackTempC
	} else if file.FallbackTempC != 0 {
		cfg.FallbackTempC = file.FallbackTempC
	}
	if m.WaterSoonPercent != 0 {
		cfg.WaterSoonPercent = m.WaterSoonPercent
	}
	if m.PostponeBoostPercent != 0 {
		cfg.PostponeBoostPercent = m.PostponeBoostPercent
	}
	if m.PostponeSuggestAfter != 0 {
		cfg.PostponeSuggestAfter = m.PostponeSuggestAfter
	}
	if m.Season.Winter != 0 {
		cfg.Season.Winter = m.Season.Winter
	}
	if m.Season.Spring != 0 {
		cfg.Season.Spring = m.Season.Spring
	}
	if m.Season.Summer != 0 {
		cfg.Season.Summer = m.Season.Summer
	}
	if m.Season.Autumn != 0 {
		cfg.Season.Autumn = m.Season.Autumn
	}
	if m.Rain.LightMM != 0 {
		cfg.Rain.LightMM = m.Rain.LightMM
	}
	if m.Rain.HeavyMM != 0 {
		cfg.Rain.HeavyMM = m.Rain.HeavyMM
	}
	if m.Rain.BaseShiftHours != 0 {
		cfg.Rain.BaseShiftHours = m.Rain.BaseShiftHours
	}
	if m.Temp.CoolThresholdC != 0 {
		cfg.Temp.CoolThresholdC = m.Temp.CoolThresholdC
	}
	if m.Temp.CoolFactor != 0 {
		cfg.Temp.CoolFactor = m.Temp.CoolFactor
	}
	if m.Temp.HotThresholdC != 0 {
		cfg.Temp.HotThresholdC = m.Temp.HotThresholdC
	}
	if m.Temp.HotSlope != 0 {
		cfg.Temp.HotSlope = m.Temp.HotSlope
	}
	return model.NormalizeModelConfig(cfg)
}

func modelSettingsFromConfig(cfg model.ModelConfig) modelSettingsFile {
	cfg = model.NormalizeModelConfig(cfg)
	var file modelSettingsFile
	file.FallbackTempC = cfg.FallbackTempC
	file.WaterSoonPercent = cfg.WaterSoonPercent
	file.PostponeBoostPercent = cfg.PostponeBoostPercent
	file.PostponeSuggestAfter = cfg.PostponeSuggestAfter
	file.Season.Winter = cfg.Season.Winter
	file.Season.Spring = cfg.Season.Spring
	file.Season.Summer = cfg.Season.Summer
	file.Season.Autumn = cfg.Season.Autumn
	file.Rain.LightMM = cfg.Rain.LightMM
	file.Rain.HeavyMM = cfg.Rain.HeavyMM
	file.Rain.BaseShiftHours = cfg.Rain.BaseShiftHours
	file.Temp.CoolThresholdC = cfg.Temp.CoolThresholdC
	file.Temp.CoolFactor = cfg.Temp.CoolFactor
	file.Temp.HotThresholdC = cfg.Temp.HotThresholdC
	file.Temp.HotSlope = cfg.Temp.HotSlope
	return file
}

func NormalizeDefaultIntervalDays(v int) int {
	if v < minDefaultIntervalDays || v > maxDefaultIntervalDays {
		return DefaultIntervalDaysDefault
	}
	return v
}

func NormalizeFallbackTempC(v float64) float64 {
	if v < minFallbackTempC || v > maxFallbackTempC {
		return DefaultFallbackTempC
	}
	return v
}

func NormalizeWeatherRefreshMinutes(v int) int {
	if v < minWeatherRefreshMinutes || v > maxWeatherRefreshMinutes {
		return DefaultWeatherRefreshMinutesDefault
	}
	return v
}

func ClampWeatherRefreshMinutes(v int) int {
	if v < minWeatherRefreshMinutes {
		return minWeatherRefreshMinutes
	}
	if v > maxWeatherRefreshMinutes {
		return maxWeatherRefreshMinutes
	}
	return v
}

func parseBoolSetting(v string) bool {
	val := strings.ToLower(strings.TrimSpace(v))
	return val == "1" || val == "true" || val == "yes" || val == "on"
}

func (s *Store) settingsFilePath() string {
	if s.dbPath != "" {
		return filepath.Join(filepath.Dir(s.dbPath), "settings.toml")
	}
	dbPath, err := DefaultDBPath()
	if err != nil {
		return "settings.toml"
	}
	return filepath.Join(filepath.Dir(dbPath), "settings.toml")
}

func (s *Store) SettingsFilePath() string {
	return s.settingsFilePath()
}

func (s *Store) SettingsBackupFilePath() string {
	return SettingsBackupPathFor(s.DBPath() + ".bak")
}

func (s *Store) migrateLegacySettings() error {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='settings'`).Scan(&count)
	if err != nil {
		return err
	}
	if count == 0 {
		return nil
	}

	path := s.settingsFilePath()
	if _, err := os.Stat(path); os.IsNotExist(err) {
		cfg, err := s.readLegacySettingsTable()
		if err != nil {
			return err
		}
		if err := s.writeSettingsFile(cfg); err != nil {
			return err
		}
	}

	_, err = s.db.Exec(`DROP TABLE IF EXISTS settings`)
	return err
}

func (s *Store) readLegacySettingsTable() (Settings, error) {
	cfg := defaultAppSettings()
	rows, err := s.db.Query(`SELECT key, value FROM settings`)
	if err != nil {
		return cfg, err
	}
	defer rows.Close()

	for rows.Next() {
		var key, value string
		if err := rows.Scan(&key, &value); err != nil {
			return cfg, err
		}
		switch key {
		case settingLanguage:
			cfg.Language = i18n.Parse(value)
		case settingAutoBackup:
			cfg.AutoBackup = parseBoolSetting(value)
		case settingTransparentMode:
			cfg.TransparentMode = parseBoolSetting(value)
		case settingFastBoot:
			cfg.FastBoot = parseBoolSetting(value)
		case settingWeatherRefreshMinutes:
			if v, convErr := strconv.Atoi(value); convErr == nil {
				cfg.WeatherRefreshMinutes = NormalizeWeatherRefreshMinutes(v)
			}
		case settingWeatherPrecipAppliedDay:
			cfg.WeatherPrecipAppliedDay = strings.TrimSpace(value)
		case settingWeatherCity:
			if strings.TrimSpace(value) == "" {
				continue
			}
			var loc weather.Location
			if unmarshalErr := json.Unmarshal([]byte(value), &loc); unmarshalErr == nil && strings.TrimSpace(loc.Name) != "" {
				cfg.WeatherCity = loc
			}
		}
	}
	return cfg, rows.Err()
}

func (s *Store) LoadSettings() (Settings, error) {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	return s.loadSettingsLocked()
}

func (s *Store) loadSettingsLocked() (Settings, error) {
	path := s.settingsFilePath()
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			cfg := defaultAppSettings()
			if saveErr := s.writeSettingsFile(cfg); saveErr != nil {
				return cfg, saveErr
			}
			return cfg, nil
		}
		return defaultAppSettings(), err
	}

	var file settingsFile
	if err := toml.Unmarshal(data, &file); err != nil {
		return defaultAppSettings(), err
	}

	cfg := defaultAppSettings()
	cfg.Language = i18n.Parse(file.Language)
	cfg.AutoBackup = file.AutoBackup
	cfg.TransparentMode = file.TransparentMode
	cfg.FastBoot = file.FastBoot
	cfg.DefaultIntervalDays = NormalizeDefaultIntervalDays(file.DefaultIntervalDays)
	cfg.Model = modelConfigFromFile(file)
	cfg.WeatherRefreshMinutes = NormalizeWeatherRefreshMinutes(file.WeatherRefreshMinutes)
	cfg.WeatherPrecipAppliedDay = strings.TrimSpace(file.WeatherPrecipAppliedDay)

	loc := weather.Location{
		Name:      strings.TrimSpace(file.WeatherCity.Name),
		Admin1:    strings.TrimSpace(file.WeatherCity.Admin1),
		Country:   strings.TrimSpace(file.WeatherCity.Country),
		Latitude:  file.WeatherCity.Latitude,
		Longitude: file.WeatherCity.Longitude,
		Timezone:  strings.TrimSpace(file.WeatherCity.Timezone),
	}
	if loc.Name != "" {
		cfg.WeatherCity = loc
	}
	return cfg, nil
}

func (s *Store) writeSettingsFile(cfg Settings) error {
	path := s.settingsFilePath()
	file := settingsFile{
		Language:                string(cfg.Language),
		AutoBackup:              cfg.AutoBackup,
		TransparentMode:         cfg.TransparentMode,
		FastBoot:                cfg.FastBoot,
		DefaultIntervalDays:     NormalizeDefaultIntervalDays(cfg.DefaultIntervalDays),
		Model:                   modelSettingsFromConfig(cfg.Model),
		WeatherRefreshMinutes:   NormalizeWeatherRefreshMinutes(cfg.WeatherRefreshMinutes),
		WeatherPrecipAppliedDay: strings.TrimSpace(cfg.WeatherPrecipAppliedDay),
	}
	file.WeatherCity.Name = cfg.WeatherCity.Name
	file.WeatherCity.Admin1 = cfg.WeatherCity.Admin1
	file.WeatherCity.Country = cfg.WeatherCity.Country
	file.WeatherCity.Latitude = cfg.WeatherCity.Latitude
	file.WeatherCity.Longitude = cfg.WeatherCity.Longitude
	file.WeatherCity.Timezone = cfg.WeatherCity.Timezone

	buf, err := toml.Marshal(file)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	content := append([]byte(settingsTOMLHeader()), buf...)
	return os.WriteFile(path, content, 0o644)
}

func (s *Store) saveSettingsLocked(cfg Settings) error {
	return s.writeSettingsFile(cfg)
}

func (s *Store) SetLanguage(lang i18n.Lang) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.Language = lang
	return s.saveSettingsLocked(cfg)
}

func (s *Store) GetAutoBackup() (bool, error) {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	return cfg.AutoBackup, err
}

func (s *Store) SetAutoBackup(enabled bool) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.AutoBackup = enabled
	return s.saveSettingsLocked(cfg)
}

func (s *Store) SetTransparentMode(enabled bool) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.TransparentMode = enabled
	return s.saveSettingsLocked(cfg)
}

func (s *Store) SetFastBoot(enabled bool) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.FastBoot = enabled
	return s.saveSettingsLocked(cfg)
}

func defaultWeatherCity() weather.Location {
	return weather.Location{
		Name:      "Moscow",
		Admin1:    "Moscow",
		Country:   "Russia",
		Latitude:  55.7558,
		Longitude: 37.6173,
		Timezone:  "Europe/Moscow",
	}
}

func (s *Store) SetWeatherCity(loc weather.Location) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.WeatherCity = loc
	return s.saveSettingsLocked(cfg)
}

func (s *Store) SetWeatherRefreshMinutes(minutes int) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.WeatherRefreshMinutes = ClampWeatherRefreshMinutes(minutes)
	return s.saveSettingsLocked(cfg)
}

func (s *Store) SetWeatherPrecipAppliedDay(day string) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.WeatherPrecipAppliedDay = strings.TrimSpace(day)
	return s.saveSettingsLocked(cfg)
}

func (s *Store) SetDefaultIntervalDays(days int) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.DefaultIntervalDays = NormalizeDefaultIntervalDays(days)
	return s.saveSettingsLocked(cfg)
}

func (s *Store) SetFallbackTempC(tempC float64) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()
	cfg, err := s.loadSettingsLocked()
	if err != nil && !os.IsNotExist(err) {
		cfg = defaultAppSettings()
	}
	cfg.Model.FallbackTempC = NormalizeFallbackTempC(tempC)
	cfg.Model = model.NormalizeModelConfig(cfg.Model)
	return s.saveSettingsLocked(cfg)
}
