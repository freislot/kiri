package tui

import (
	"fmt"
	"math"
	"strings"
	"time"

	"kiri/internal/model"
	"kiri/internal/weather"

	tea "github.com/charmbracelet/bubbletea"
)

const (
	degradationPersistEpsilon  = 0.5
	degradationPersistInterval = 30 * time.Minute
)

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.recalcLayout()
		m.scrollToFocused()
		m.clampCareLogOffset()
		if m.splashPhase != splashPhaseDone {
			m.initSplashFog()
		}
		return m, nil

	case plantsLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.plants = msg.plants
		if err := m.applyElapsedDegradation(time.Now()); err != nil {
			m.err = err
			return m, nil
		}
		m.recalcLayout() // keeps visRows consistent; clamps cursor
		m.scrollToFocused()
		weatherCmd := m.applyWeatherPrecipitation()
		return m, tea.Batch(m.reloadPlantCareLog(), weatherCmd)

	case careLogLoadedMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.careLog = msg.entries
		m.clampCareLogOffset()
		return m, nil

	case plantCareLogMsg:
		if msg.err != nil {
			m.err = msg.err
			return m, nil
		}
		m.detailLog = msg.entries
		return m, nil

	case weatherLoadedMsg:
		if msg.err != nil {
			m.weatherLoading = false
			m.weatherReady = false
			m.status = m.cat().WeatherLoadFailed()
			return m, nil
		}
		m.weatherNow = msg.cond
		m.weatherReady = true
		m.weatherLoading = false
		return m, m.applyWeatherPrecipitation()

	case weatherCityLocalizedMsg:
		if msg.err != nil || !msg.loc.Matches(m.weatherCity) {
			return m, nil
		}
		m.weatherCity = msg.loc
		m.cityInput = msg.loc.Name
		if err := m.store.SetWeatherCity(msg.loc); err != nil {
			m.err = err
		}
		return m, nil

	case weatherRefreshTickMsg:
		if msg.seq != m.weatherRefreshSeq {
			return m, nil
		}
		if m.weatherCity.Name == "" {
			return m, nil
		}
		if m.weatherLoading {
			return m, weatherRefreshTick(m.weatherRefreshSeq, m.weatherRefreshMinutes)
		}
		m.weatherLoading = true
		return m, tea.Batch(
			fetchCurrentWeather(m.weather, m.weatherCity),
			weatherRefreshTick(m.weatherRefreshSeq, m.weatherRefreshMinutes),
		)

	case dayProgressTickMsg:
		if err := m.applyElapsedDegradation(time.Now()); err != nil {
			m.err = err
			return m, dayProgressTick()
		}
		return m, tea.Batch(dayProgressTick(), m.applyWeatherPrecipitation())

	case citySearchLoadedMsg:
		if strings.TrimSpace(msg.query) != strings.TrimSpace(m.cityInput) {
			return m, nil
		}
		if msg.err != nil {
			m.status = m.cat().SettingsCitySearchFailed()
			m.cityCandidates = nil
			m.cityCursor = 0
			return m, nil
		}
		m.cityCandidates = msg.items
		if m.cityCursor >= len(m.cityCandidates) {
			m.cityCursor = 0
		}
		return m, nil

	case aboutFogTickMsg:
		if m.tab != TabAbout || m.splashPhase != splashPhaseDone {
			return m, nil
		}
		m.aboutFogTime += aboutFogTimeStep
		return m, aboutFogTick()

	case splashRevealTickMsg:
		if m.splashPhase != splashPhaseReveal {
			return m, nil
		}
		step := float64(splashRevealInterval) / float64(splashRevealDuration)
		m.splashRevealProg += step
		if m.splashRevealProg >= 1 {
			m.startSplashShow()
			return m, splashHoldTimer()
		}
		return m, splashRevealTick()

	case splashHoldDoneMsg:
		if m.splashPhase == splashPhaseShow {
			m.startSplashFade()
			return m, splashFadeTick()
		}
		return m, nil

	case splashFogTickMsg:
		if m.splashPhase == splashPhaseDone {
			return m, nil
		}
		if len(m.splashFog) == 0 {
			m.initSplashFog()
		}
		m.updateSplashFog()
		return m, splashFogTick()

	case splashFadeTickMsg:
		if m.splashPhase != splashPhaseFade {
			return m, nil
		}
		step := float64(splashFadeInterval) / float64(splashFadeDuration)
		m.splashFadeProg += step
		if m.splashFadeProg >= 1 {
			m.dismissSplash()
			return m, nil
		}
		m.splashAlpha = 1 - splashEaseOut(m.splashFadeProg)
		return m, splashFadeTick()

	case tea.MouseMsg:
		return m.handleMouse(msg)

	case tea.KeyMsg:
		if m.splashPhase != splashPhaseDone {
			switch msg.String() {
			case "ctrl+c", "q":
				return m, tea.Quit
			case "enter":
				m.dismissSplash()
				return m, nil
			}
			return m, nil
		}

		if m.overlay != OverlayNone {
			if msg.String() == "q" {
				var status string
				switch m.overlay {
				case OverlayDeleteConfirm:
					status = m.cat().StatusDeleteCancelled()
				case OverlayBackupConfirm:
					status = m.cat().SettingsBackupCancelled()
				case OverlayEditPlant:
					status = m.cat().StatusEditCancelled()
				default:
					status = m.cat().StatusAddCancelled()
				}
				m.cancelOverlay()
				m.status = status
				return m, nil
			}
			return m.handleOverlayKey(msg)
		}

		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "tab", "shift+tab":
			delta := 1
			if msg.String() == "shift+tab" {
				delta = -1
			}
			m.tab = Tab((int(m.tab) + delta + tabCount) % tabCount)
			m.status = ""
			if m.tab == TabCareLog {
				m.clampCareLogOffset()
			}
			return m, m.aboutTabCmd()
		case "1":
			m.tab = TabPlants
			return m, nil
		case "2":
			m.tab = TabCalendar
			m.calFocus = calFocusGrid
			m.taskCursor = 0
			return m, nil
		case "3":
			m.tab = TabCareLog
			m.clampCareLogOffset()
			return m, nil
		case "4":
			m.tab = TabSettings
			return m, nil
		case "5":
			m.tab = TabAbout
			return m, m.aboutTabCmd()
		}

		if m.tab == TabAbout {
			return m, nil
		}

		if m.tab == TabSettings {
			return m.handleSettingsKey(msg)
		}

		if m.tab == TabCalendar {
			return m.handleCalendarKey(msg)
		}

		if m.tab == TabCareLog {
			return m.handleCareLogKey(msg)
		}

		if m.tab != TabPlants {
			return m, nil
		}

		return m.handlePlantsKey(msg)
	}

	return m, nil
}

func (m Model) handleCalendarKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		if m.calFocus == calFocusGrid {
			m.calFocus = calFocusTasks
			m.taskCursor = 0
		}
		return m, nil
	case "esc":
		if m.calFocus == calFocusTasks {
			m.calFocus = calFocusGrid
		}
		return m, nil
	}

	if m.calFocus == calFocusTasks {
		tasks := m.tasksForDate(m.calSelectedDate())
		switch msg.String() {
		case "down", "j":
			if m.taskCursor < len(tasks)-1 {
				m.taskCursor++
			}
			return m, nil
		case "up", "k":
			if m.taskCursor > 0 {
				m.taskCursor--
			}
			return m, nil
		case " ":
			return m, m.toggleSelectedTask()
		case "left", "h":
			m.calFocus = calFocusGrid
			m.moveCalendarDay(-1)
			m.status = ""
			return m, nil
		case "right", "l":
			m.calFocus = calFocusGrid
			m.moveCalendarDay(1)
			m.status = ""
			return m, nil
		default:
			return m, nil
		}
	}

	switch msg.String() {
	case "left", "h":
		m.moveCalendarDay(-1)
		m.status = ""
		return m, nil
	case "right", "l":
		m.moveCalendarDay(1)
		m.status = ""
		return m, nil
	case "up", "k":
		m.moveCalendarDay(-7)
		m.status = ""
		return m, nil
	case "down", "j":
		m.moveCalendarDay(7)
		m.status = ""
		return m, nil
	case " ":
		if len(m.tasksForDate(m.calSelectedDate())) > 0 {
			m.calFocus = calFocusTasks
			m.taskCursor = 0
		}
		return m, nil
	}

	return m, nil
}

func (m Model) handlePlantsKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	n := len(m.plants)
	totalCols := 0
	if n > 0 {
		totalCols = (n + vr - 1) / vr
	}

	switch msg.String() {
	case "right", "l":
		if m.cursorCol+1 < totalCols {
			m.cursorCol++
			m.clampCursorRow() // last column may have fewer rows
			m.scrollToFocused()
			m.status = ""
		}
		return m, m.reloadPlantCareLog()

	case "left", "h":
		if m.cursorCol > 0 {
			m.cursorCol--
			m.scrollToFocused()
			m.status = ""
		}
		return m, m.reloadPlantCareLog()

	case "down", "j":
		nextRow := m.cursorRow + 1
		nextIdx := nextRow + m.cursorCol*vr
		if nextRow < vr && nextIdx < n {
			m.cursorRow = nextRow
			m.status = ""
		}
		return m, m.reloadPlantCareLog()

	case "up", "k":
		if m.cursorRow > 0 {
			m.cursorRow--
			m.status = ""
		}
		return m, m.reloadPlantCareLog()

	case "a":
		m.openAddPlant()
		return m, nil
	case "e":
		m.openEditPlant()
		return m, nil
	case "d":
		m.openDeleteConfirm()
		return m, nil
	case "w":
		return m, m.waterSelected()
	case "s":
		return m, m.postponeSelected()
	}

	return m, nil
}

func (m *Model) clampCursorRow() {
	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	colStart := m.cursorCol * vr
	maxRow := vr - 1
	if colStart+maxRow >= len(m.plants) {
		maxRow = len(m.plants) - colStart - 1
	}
	if maxRow < 0 {
		maxRow = 0
	}
	if m.cursorRow > maxRow {
		m.cursorRow = maxRow
	}
}

func (m *Model) waterSelected() tea.Cmd {
	p := m.selectedPlant()
	if p == nil {
		return nil
	}
	c := m.cat()
	model.WaterPlant(p)
	m.logCare(p.ID, "watered", c.LogWateredNormal())
	m.status = c.StatusWatered(p.Name)
	if err := m.persistPlant(p); err != nil {
		m.err = err
	}
	return tea.Batch(loadCareLog(m.store), m.reloadPlantCareLog())
}

func (m *Model) postponeSelected() tea.Cmd {
	p := m.selectedPlant()
	if p == nil {
		return nil
	}
	c := m.cat()
	suggest := model.PostponeWateringWithConfig(p, m.effectiveModelCfg())
	m.logCare(p.ID, "postponed", c.LogPostponed())
	m.status = c.StatusPostponed(p.Name)
	if suggest {
		m.logCare(p.ID, "suggestion", c.LogIntervalSuggestion(p.BaseIntervalDays+1))
	}
	if err := m.persistPlant(p); err != nil {
		m.err = err
	}
	return tea.Batch(loadCareLog(m.store), m.reloadPlantCareLog())
}

func (m *Model) applyWeatherPrecipitation() tea.Cmd {
	if !m.weatherReady || len(m.plants) == 0 {
		return nil
	}
	precipMM := m.weatherNow.PrecipMM
	if m.weatherNow.PrevDayPrecipKnown {
		precipMM = m.weatherNow.PrevDayPrecipMM
	}
	if precipMM < m.effectiveModelCfg().Rain.LightMM {
		return nil
	}
	dayKey := weatherDayKey(m.weatherCity)
	if m.weatherAppliedDayKey == dayKey {
		return nil
	}
	c := m.cat()
	changed := 0
	tempC := m.weatherNow.TemperatureC
	for i := range m.plants {
		p := &m.plants[i]
		affected, effect := m.effectiveModelCfg().ApplyPrecipitationAt(p, precipMM, tempC, time.Now())
		if !affected {
			continue
		}
		if err := m.persistPlant(p); err != nil {
			m.err = err
			return nil
		}
		if precipMM > m.effectiveModelCfg().Rain.HeavyMM {
			m.logCare(p.ID, "weather", c.LogRainWatered(precipMM))
		} else {
			m.logCare(p.ID, "weather", c.LogRainShifted(precipMM, effect.ShiftHours, effect.WaterAdded))
		}
		changed++
	}
	if changed == 0 {
		return nil
	}
	if err := m.store.SetWeatherPrecipAppliedDay(dayKey); err != nil {
		m.err = err
		return nil
	}
	m.weatherAppliedDayKey = dayKey
	m.status = c.StatusRainApplied(changed, precipMM)
	return tea.Batch(loadCareLog(m.store), m.reloadPlantCareLog())
}

func weatherDayKey(loc weather.Location) string {
	cityPart := "default"
	if loc.Name != "" {
		cityPart = strings.ToLower(loc.Name)
	}
	coordPart := fmt.Sprintf("%.3f,%.3f", loc.Latitude, loc.Longitude)
	datePart := time.Now().Format("2006-01-02")
	if loc.Timezone != "" {
		if tz, err := time.LoadLocation(loc.Timezone); err == nil {
			datePart = time.Now().In(tz).Format("2006-01-02")
		}
	}
	return datePart + "|" + cityPart + "|" + coordPart
}

func (m *Model) applyElapsedDegradation(now time.Time) error {
	for i := range m.plants {
		p := &m.plants[i]
		if p.State == model.StateNew || p.BaseIntervalDays <= 0 {
			continue
		}
		if p.LastUpdatedAt.IsZero() {
			p.LastUpdatedAt = now
			if err := m.persistPlant(p); err != nil {
				return err
			}
			continue
		}
		if !now.After(p.LastUpdatedAt) {
			continue
		}
		beforeLevel := p.WaterLevel
		beforeState := p.State
		beforeUpdatedAt := p.LastUpdatedAt
		model.ApplyElapsedDegradationWithConfig(p, p.LastUpdatedAt, now, m.effectiveModelCfg().FallbackTempC, m.effectiveModelCfg())
		changedState := p.State != beforeState
		changedLevel := math.Abs(p.WaterLevel-beforeLevel) >= degradationPersistEpsilon
		persistByInterval := now.Sub(beforeUpdatedAt) >= degradationPersistInterval
		if !changedState && !changedLevel && !persistByInterval {
			continue
		}
		if err := m.persistPlant(p); err != nil {
			return err
		}
	}
	return nil
}
