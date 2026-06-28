package tui

import (
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"
	"kiri/internal/i18n"
	"kiri/internal/model"
)

func (m Model) cat() i18n.Catalog {
	return i18n.New(m.lang)
}

func wateringDate(p model.Plant, cfg model.ModelConfig) time.Time {
	return model.WateringDueDateWithConfig(p, time.Now(), cfg)
}

func (m Model) rowStatusText(p model.Plant) string {
	c := m.cat()
	switch p.State {
	case model.StateWaterNow:
		return c.RowStatusWaterNow()
	case model.StateShiftedByRain:
		return c.RowStatusShiftedRain(wateringDate(p, m.effectiveModelCfg()))
	default:
		return c.RowStatusWatering(wateringDate(p, m.effectiveModelCfg()))
	}
}

func (m Model) detailStatusText(p model.Plant) string {
	c := m.cat()
	switch p.State {
	case model.StateShiftedByRain:
		return c.DetailStatusRain(wateringDate(p, m.effectiveModelCfg()))
	case model.StateWaterNow:
		return c.DetailStatusCritical()
	default:
		return c.DetailStatusScheduled(wateringDate(p, m.effectiveModelCfg()))
	}
}

func (m Model) lastLogText(entries []model.CareLogEntry) string {
	c := m.cat()
	for _, e := range entries {
		if e.EventType == "watered" {
			return c.FormatLogLine(e.CreatedAt, e.Message)
		}
	}
	if len(entries) == 0 {
		return c.NoLogEntries()
	}
	e := entries[0]
	return c.FormatLogLine(e.CreatedAt, e.Message)
}

func (m Model) calendarEventCount() int {
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	n := 0
	for _, p := range m.plants {
		due, _ := model.IsDueOnDateWithConfig(p, today, now, m.effectiveModelCfg())
		if due {
			n++
		}
	}
	return n
}

func (m Model) tabName(t Tab) string {
	c := m.cat()
	switch t {
	case TabPlants:
		return c.TabAllPlants()
	case TabCalendar:
		return c.TabCalendar(m.calendarEventCount())
	case TabCareLog:
		return c.TabCareLog()
	case TabSettings:
		return c.TabSettings()
	case TabAbout:
		return c.TabAbout()
	default:
		return ""
	}
}

func (m Model) tabLabels() []string {
	return []string{
		m.tabName(TabPlants),
		m.tabName(TabCalendar),
		m.tabName(TabCareLog),
		m.tabName(TabSettings),
		m.tabName(TabAbout),
	}
}

func renderToggleOption(active bool, label string) string {
	if active {
		return StyleToggleActive.Render("● " + label)
	}
	return StyleToggleInactive.Render("○ " + label)
}

func settingsRowStyle(active bool) lipgloss.Style {
	if active {
		return lipgloss.NewStyle().Foreground(colorGreen).Bold(true)
	}
	return lipgloss.NewStyle().Foreground(colorMuted)
}

func settingsCursorPrefix(active bool) string {
	if active {
		return "> "
	}
	return "  "
}

func renderSettingsListValue(v string) string {
	return "< " + v + " >"
}

func (m Model) renderSettingsLanguageRow(active bool) string {
	c := m.cat()
	lang := "English"
	if m.lang == i18n.RU {
		lang = "Русский"
	}
	line := settingsCursorPrefix(active) + c.SettingsOptionLanguage() + ": " + renderSettingsListValue(lang)
	return settingsRowStyle(active).Render(line)
}

func (m Model) renderSettingsCityRow(active bool) string {
	c := m.cat()
	value := m.weatherCity.DisplayName()
	if value == "" {
		value = c.SettingsCityNotSelected()
	}
	if m.citySelecting {
		if strings.TrimSpace(m.cityInput) == "" {
			value = c.SettingsCityInputPlaceholder()
		} else {
			value = m.cityInput
		}
	}
	line := settingsCursorPrefix(active) + c.SettingsOptionCity() + ": " + value
	return settingsRowStyle(active).Render(line)
}

func (m Model) renderSettingsWeatherRefreshRow(active bool) string {
	c := m.cat()
	line := settingsCursorPrefix(active) + c.SettingsOptionWeatherRefresh() + ": " +
		renderSettingsListValue(c.SettingsWeatherRefreshValue(m.weatherRefreshMinutes))
	return settingsRowStyle(active).Render(line)
}

func (m Model) renderSettingsDefaultIntervalRow(active bool) string {
	c := m.cat()
	line := settingsCursorPrefix(active) + c.SettingsOptionDefaultInterval() + ": " +
		renderSettingsListValue(c.SettingsDefaultIntervalValue(m.defaultIntervalDays))
	return settingsRowStyle(active).Render(line)
}

func (m Model) renderSettingsFallbackTempRow(active bool) string {
	c := m.cat()
	line := settingsCursorPrefix(active) + c.SettingsOptionFallbackTemp() + ": " +
		renderSettingsListValue(c.SettingsFallbackTempValue(m.effectiveModelCfg().FallbackTempC))
	return settingsRowStyle(active).Render(line)
}

func (m Model) renderSettingsBinaryRow(active bool, label string, enabled bool) string {
	c := m.cat()
	yes := renderToggleOption(enabled, c.SettingsYes())
	no := renderToggleOption(!enabled, c.SettingsNo())
	line := settingsCursorPrefix(active) + label + ": " + yes + fillSpaces(2) + no
	return settingsRowStyle(active).Render(line)
}

func (m Model) renderSettingsInfoPathRow(label, path string) string {
	prefix := "  " + label + ": "
	value := path
	if maxW := m.headerTextWidth(); maxW > lipgloss.Width(prefix)+4 {
		value = truncateWidth(path, maxW-lipgloss.Width(prefix))
	}
	return StyleRowMuted.Render(prefix + value)
}

func (m Model) settingsOptionDescription() string {
	c := m.cat()
	switch m.cursor {
	case settingsRowLanguage:
		return c.SettingsDescLanguage()
	case settingsRowCity:
		return c.SettingsDescCity()
	case settingsRowWeatherRefresh:
		return c.SettingsDescWeatherRefresh()
	case settingsRowDefaultInterval:
		return c.SettingsDescDefaultInterval()
	case settingsRowFallbackTemp:
		return c.SettingsDescFallbackTemp()
	case settingsRowAutoBackup:
		return c.SettingsDescAutoBackup()
	case settingsRowTransparent:
		return c.SettingsDescTransparent()
	case settingsRowFastBoot:
		return c.SettingsDescFastBoot()
	default:
		return ""
	}
}

func (m Model) renderSettingsDescriptionRow() string {
	text := m.settingsOptionDescription()
	if text == "" {
		return ""
	}
	maxW := m.headerTextWidth()
	if maxW < 1 {
		return ""
	}
	return styleWrapped(StyleRowMuted, "  "+text, maxW)
}

func (m Model) weatherHeaderCityName() string {
	if m.weatherCity.Name != "" {
		return m.weatherCity.Name
	}
	return m.cat().WeatherHeaderDefaultCity()
}

func (m Model) weatherHeaderText() string {
	c := m.cat()
	city := m.weatherHeaderCityName()
	if m.weatherLoading {
		return c.WeatherLoading(city)
	}
	if !m.weatherReady {
		return c.WeatherUnavailable(city)
	}
	return c.WeatherHeaderLive(city, m.weatherNow.TemperatureC, c.WeatherCodeLabel(m.weatherNow.WeatherCode))
}

func (m Model) renderCitySuggestions() string {
	if !m.citySelecting || len(m.cityCandidates) == 0 {
		return ""
	}
	c := m.cat()
	lines := make([]string, 0, len(m.cityCandidates)+1)
	lines = append(lines, StyleRowMuted.Render(c.SettingsCityResults()))
	for i, item := range m.cityCandidates {
		line := "  " + item.DisplayName()
		if i == m.cityCursor {
			line = "› " + item.DisplayName()
			lines = append(lines, StyleToggleActive.Render(line))
			continue
		}
		lines = append(lines, StyleToggleInactive.Render(line))
	}
	return strings.Join(lines, "\n")
}
