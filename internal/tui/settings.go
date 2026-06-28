package tui

import tea "github.com/charmbracelet/bubbletea"

const (
	settingsRowLanguage = iota
	settingsRowCity
	settingsRowWeatherRefresh
	settingsRowDefaultInterval
	settingsRowFallbackTemp
	settingsRowAutoBackup
	settingsRowTransparent
	settingsRowFastBoot
	settingsRowCount
)

func settingsKeyDirection(k string) (int, bool) {
	switch k {
	case "h", "left":
		return -1, true
	case "l", "right":
		return 1, true
	default:
		return 0, false
	}
}

func (m *Model) moveSettingsCursor(delta int) {
	if delta == 0 {
		return
	}
	m.cursor = (m.cursor + delta + settingsRowCount) % settingsRowCount
}

func (m Model) handleSettingsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	c := m.cat()
	if m.citySelecting {
		if msg.Type == tea.KeyRunes && len(msg.Runes) > 0 {
			m.cityInput += string(msg.Runes)
			m.cityCursor = 0
			if len([]rune(m.cityInput)) < 2 {
				m.cityCandidates = nil
				return m, nil
			}
			return m, searchCities(m.weather, m.cityInput, m.cityLanguageCode())
		}

		switch msg.String() {
		case "esc":
			m.citySelecting = false
			m.cityCandidates = nil
			m.cityCursor = 0
			m.cityInput = m.weatherCity.Name
			return m, nil
		case "enter":
			return m, m.confirmCitySelection()
		case "up":
			if m.cityCursor > 0 {
				m.cityCursor--
			}
			return m, nil
		case "down":
			if m.cityCursor < len(m.cityCandidates)-1 {
				m.cityCursor++
			}
			return m, nil
		case "backspace", "ctrl+h":
			if len(m.cityInput) > 0 {
				rs := []rune(m.cityInput)
				m.cityInput = string(rs[:len(rs)-1])
			}
			m.cityCursor = 0
			if len([]rune(m.cityInput)) < 2 {
				m.cityCandidates = nil
				return m, nil
			}
			return m, searchCities(m.weather, m.cityInput, m.cityLanguageCode())
		}
		return m, nil
	}

	switch msg.String() {
	case "down", "j":
		m.moveSettingsCursor(1)
		return m, nil
	case "up", "k":
		m.moveSettingsCursor(-1)
		return m, nil
	case "b":
		m.openBackupConfirm()
		return m, nil
	}

	if dir, ok := settingsKeyDirection(msg.String()); ok {
		switch m.cursor {
		case settingsRowLanguage:
			if err := m.toggleLanguage(); err != nil {
				m.err = err
				return m, nil
			}
			m.status = c.LanguageChanged()
			if m.weatherCity.Name == "" {
				return m, nil
			}
			return m, localizeWeatherCity(m.weather, m.weatherCity, m.cityLanguageCode())
		case settingsRowWeatherRefresh:
			if err := m.shiftWeatherRefreshMinutes(dir); err != nil {
				m.err = err
				return m, nil
			}
			m.status = c.SettingsWeatherRefreshChanged(m.weatherRefreshMinutes)
			return m, m.restartWeatherRefreshLoop()
		case settingsRowDefaultInterval:
			if err := m.shiftDefaultIntervalDays(dir); err != nil {
				m.err = err
				return m, nil
			}
			m.status = c.SettingsDefaultIntervalChanged(m.defaultIntervalDays)
			return m, nil
		case settingsRowFallbackTemp:
			if err := m.shiftFallbackTempC(dir); err != nil {
				m.err = err
				return m, nil
			}
			m.status = c.SettingsFallbackTempChanged(m.effectiveModelCfg().FallbackTempC)
			return m, nil
		case settingsRowAutoBackup:
			if err := m.toggleAutoBackup(); err != nil {
				m.err = err
				return m, nil
			}
			if m.autoBackup {
				m.status = c.SettingsAutoBackupEnabled()
			} else {
				m.status = c.SettingsAutoBackupDisabled()
			}
			return m, nil
		case settingsRowTransparent:
			if err := m.toggleTransparent(); err != nil {
				m.err = err
				return m, nil
			}
			if m.transparent {
				m.status = c.SettingsTransparentEnabled()
			} else {
				m.status = c.SettingsTransparentDisabled()
			}
			return m, nil
		case settingsRowFastBoot:
			if err := m.toggleFastBoot(); err != nil {
				m.err = err
				return m, nil
			}
			if m.fastBoot {
				m.status = c.SettingsFastBootEnabled()
			} else {
				m.status = c.SettingsFastBootDisabled()
			}
			return m, nil
		}
	}

	switch msg.String() {
	case " ", "enter":
		switch m.cursor {
		case settingsRowCity:
			return m, m.beginCitySearch()
		case settingsRowAutoBackup:
			if err := m.toggleAutoBackup(); err != nil {
				m.err = err
				return m, nil
			}
			if m.autoBackup {
				m.status = c.SettingsAutoBackupEnabled()
			} else {
				m.status = c.SettingsAutoBackupDisabled()
			}
			return m, nil
		case settingsRowTransparent:
			if err := m.toggleTransparent(); err != nil {
				m.err = err
				return m, nil
			}
			if m.transparent {
				m.status = c.SettingsTransparentEnabled()
			} else {
				m.status = c.SettingsTransparentDisabled()
			}
			return m, nil
		case settingsRowFastBoot:
			if err := m.toggleFastBoot(); err != nil {
				m.err = err
				return m, nil
			}
			if m.fastBoot {
				m.status = c.SettingsFastBootEnabled()
			} else {
				m.status = c.SettingsFastBootDisabled()
			}
			return m, nil
		}
	case "c":
		if m.cursor == settingsRowCity {
			return m, m.beginCitySearch()
		}
	}

	return m, nil
}
