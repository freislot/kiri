package tui

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"time"

	"kiri/internal/db"
	"kiri/internal/i18n"
	"kiri/internal/model"
	"kiri/internal/weather"

	tea "github.com/charmbracelet/bubbletea"
)

type Tab int

const (
	TabPlants Tab = iota
	TabCalendar
	TabCareLog
	TabSettings
	TabAbout
)

const tabCount = 5

const dayProgressCheckInterval = time.Minute

type Model struct {
	store                 *db.Store
	weather               weather.Service
	plants                []model.Plant
	careLog               []model.CareLogEntry
	detailLog             []model.CareLogEntry
	cursorRow             int // row in the current column (0 … visRows-1)
	cursorCol             int // absolute column index
	viewportCol           int // first visible column
	visCols               int // visible columns in viewport (fixed at maxPlantCols)
	visRows               int // visible rows on screen (≥ 2)
	mainWidth             int // width of main dashboard area
	tab                   Tab
	cursor                int // active row in Settings tab
	lang                  i18n.Lang
	autoBackup            bool
	transparent           bool
	fastBoot              bool
	defaultIntervalDays   int
	modelCfg              model.ModelConfig
	weatherCity           weather.Location
	weatherNow            weather.Conditions
	weatherReady          bool
	weatherLoading        bool
	weatherRefreshMinutes int
	weatherRefreshSeq     int
	weatherAppliedDayKey  string
	cityInput             string
	citySelecting         bool
	cityCandidates        []weather.Location
	cityCursor            int
	confirmBtn            int // 0 = No, 1 = Yes (backup/delete confirm overlays)
	overlay               Overlay
	plantForm             plantForm
	calYear               int
	calMonth              int
	calSelectedDay        int
	calFocus              calFocus
	taskCursor            int
	careLogOffset         int // first visible care-log entry line
	width                 int
	height                int
	status                string
	err                   error
	splashPhase           splashPhase
	splashAlpha           float64
	splashRevealProg      float64
	splashFadeProg        float64
	splashFog             []splashParticle
	aboutFogTime          float64
}

func (m Model) flatIndex() int {
	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	return m.cursorRow + m.cursorCol*vr
}

func (m Model) selectedPlant() *model.Plant {
	idx := m.flatIndex()
	if len(m.plants) == 0 || idx < 0 || idx >= len(m.plants) {
		return nil
	}
	return &m.plants[idx]
}

func (m *Model) recalcLayout() {
	savedIdx := m.flatIndex()

	m.mainWidth = m.headerTextWidth()

	m.visCols = maxPlantCols

	avail := m.termHeight() - uiOverhead
	m.visRows = avail / cardMaxHeight
	if m.visRows < 2 {
		m.visRows = 2
	}

	vr := m.visRows
	m.cursorRow = savedIdx % vr
	m.cursorCol = savedIdx / vr
	m.clampCursor()
}

func (m *Model) clampCursor() {
	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	n := len(m.plants)

	if n == 0 {
		m.cursorRow = 0
		m.cursorCol = 0
		return
	}

	totalCols := (n + vr - 1) / vr

	if m.cursorCol < 0 {
		m.cursorCol = 0
	}
	if m.cursorCol >= totalCols {
		m.cursorCol = totalCols - 1
	}

	colStart := m.cursorCol * vr
	maxRow := vr - 1
	if colStart+maxRow >= n {
		maxRow = n - colStart - 1
	}
	if maxRow < 0 {
		maxRow = 0
	}

	if m.cursorRow < 0 {
		m.cursorRow = 0
	}
	if m.cursorRow > maxRow {
		m.cursorRow = maxRow
	}
}

func (m *Model) scrollToFocused() {
	vc := m.visCols
	if vc < 1 {
		vc = 3
	}

	if m.cursorCol < m.viewportCol {
		m.viewportCol = m.cursorCol
	}
	if m.cursorCol >= m.viewportCol+vc {
		m.viewportCol = m.cursorCol - vc + 1
	}
	if m.viewportCol < 0 {
		m.viewportCol = 0
	}

	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	if len(m.plants) > 0 {
		totalCols := (len(m.plants) + vr - 1) / vr
		maxVP := totalCols - vc
		if maxVP < 0 {
			maxVP = 0
		}
		if m.viewportCol > maxVP {
			m.viewportCol = maxVP
		}
	}
}

func (m Model) effectiveModelCfg() model.ModelConfig {
	if m.modelCfg.Rain.LightMM == 0 && m.modelCfg.Season.Spring == 0 && m.modelCfg.FallbackTempC == 0 {
		return model.DefaultConfig()
	}
	return model.NormalizeModelConfig(m.modelCfg)
}

type Options struct {
	Transparent bool
}

func New(store *db.Store, wx weather.Service, opts Options) Model {
	cfg, _ := store.LoadSettings()
	lang := cfg.Language
	autoBackup := cfg.AutoBackup
	transparent := cfg.TransparentMode
	fastBoot := cfg.FastBoot
	defaultIntervalDays := cfg.DefaultIntervalDays
	modelCfg := cfg.Model
	weatherRefreshMinutes := cfg.WeatherRefreshMinutes
	weatherCity := cfg.WeatherCity
	weatherAppliedDayKey := cfg.WeatherPrecipAppliedDay
	if opts.Transparent {
		transparent = true
	}
	InitTerminalDisplay(os.Stdout)
	ApplyTransparentMode(transparent)
	now := time.Now()
	m := Model{
		store:                 store,
		weather:               wx,
		tab:                   TabPlants,
		cursor:                0,
		lang:                  lang,
		autoBackup:            autoBackup,
		transparent:           transparent,
		fastBoot:              fastBoot,
		defaultIntervalDays:   db.NormalizeDefaultIntervalDays(defaultIntervalDays),
		modelCfg:              modelCfg,
		weatherCity:           weatherCity,
		weatherLoading:        weatherCity.Name != "",
		weatherRefreshMinutes: db.NormalizeWeatherRefreshMinutes(weatherRefreshMinutes),
		weatherRefreshSeq:     1,
		weatherAppliedDayKey:  weatherAppliedDayKey,
		cityInput:             weatherCity.Name,
		calYear:               now.Year(),
		calMonth:              int(now.Month()),
		calSelectedDay:        now.Day(),
		calFocus:              calFocusGrid,
		splashPhase:           splashPhaseDone,
		splashAlpha:           0,
		splashRevealProg:      0,
	}
	if !fastBoot {
		m.splashPhase = splashPhaseReveal
		m.splashAlpha = 1
	}
	m.recalcLayout()
	return m
}

func (m Model) Init() tea.Cmd {
	cmds := []tea.Cmd{
		loadPlants(m.store),
		loadCareLog(m.store),
		dayProgressTick(),
	}
	if m.weatherCity.Name != "" {
		lang := m.cityLanguageCode()
		cmds = append(cmds,
			fetchCurrentWeather(m.weather, m.weatherCity),
			weatherRefreshTick(m.weatherRefreshSeq, m.weatherRefreshMinutes),
			localizeWeatherCity(m.weather, m.weatherCity, lang),
		)
	}
	if !m.fastBoot {
		cmds = append(cmds, splashRevealTick(), splashFogTick())
	}
	return tea.Batch(cmds...)
}

type plantsLoadedMsg struct {
	plants []model.Plant
	err    error
}

type careLogLoadedMsg struct {
	entries []model.CareLogEntry
	err     error
}

func loadPlants(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		plants, err := store.ListPlants()
		return plantsLoadedMsg{plants: plants, err: err}
	}
}

func loadCareLog(store *db.Store) tea.Cmd {
	return func() tea.Msg {
		entries, err := store.AllCareLog(50)
		return careLogLoadedMsg{entries: entries, err: err}
	}
}

func (m *Model) reloadPlantCareLog() tea.Cmd {
	p := m.selectedPlant()
	if p == nil {
		return nil
	}
	id := p.ID
	return func() tea.Msg {
		entries, err := m.store.CareLogForPlant(id, 10)
		return plantCareLogMsg{plantID: id, entries: entries, err: err}
	}
}

type plantCareLogMsg struct {
	plantID int64
	entries []model.CareLogEntry
	err     error
}

type weatherLoadedMsg struct {
	loc  weather.Location
	cond weather.Conditions
	err  error
}

type citySearchLoadedMsg struct {
	query string
	items []weather.Location
	err   error
}

type weatherRefreshTickMsg struct {
	seq int
}

type dayProgressTickMsg struct{}

type weatherCityLocalizedMsg struct {
	loc weather.Location
	err error
}

func dayProgressTick() tea.Cmd {
	return tea.Tick(dayProgressCheckInterval, func(time.Time) tea.Msg {
		return dayProgressTickMsg{}
	})
}

func (m *Model) persistPlant(p *model.Plant) error {
	return m.store.SavePlant(p)
}

func (m *Model) logCare(plantID int64, eventType, message string) {
	if err := m.store.AddCareLog(plantID, eventType, message); err != nil {
		m.err = err
	}
}

func (m *Model) toggleLanguage() error {
	m.lang = m.lang.Toggle()
	return m.store.SetLanguage(m.lang)
}

func (m *Model) toggleAutoBackup() error {
	m.autoBackup = !m.autoBackup
	return m.store.SetAutoBackup(m.autoBackup)
}

func (m *Model) toggleTransparent() error {
	m.transparent = !m.transparent
	ApplyTransparentMode(m.transparent)
	return m.store.SetTransparentMode(m.transparent)
}

func (m *Model) toggleFastBoot() error {
	m.fastBoot = !m.fastBoot
	return m.store.SetFastBoot(m.fastBoot)
}

func (m *Model) shiftWeatherRefreshMinutes(delta int) error {
	if delta == 0 {
		return nil
	}
	minutes := db.ClampWeatherRefreshMinutes(m.weatherRefreshMinutes) + delta
	m.weatherRefreshMinutes = db.ClampWeatherRefreshMinutes(minutes)
	return m.store.SetWeatherRefreshMinutes(m.weatherRefreshMinutes)
}

func (m *Model) shiftDefaultIntervalDays(delta int) error {
	if delta == 0 {
		return nil
	}
	days := db.NormalizeDefaultIntervalDays(m.defaultIntervalDays) + delta
	if days < 1 {
		days = 1
	}
	if days > 365 {
		days = 365
	}
	m.defaultIntervalDays = days
	return m.store.SetDefaultIntervalDays(days)
}

func (m *Model) shiftFallbackTempC(delta int) error {
	if delta == 0 {
		return nil
	}
	if m.modelCfg.Rain.LightMM == 0 && m.modelCfg.Season.Spring == 0 && m.modelCfg.FallbackTempC == 0 {
		m.modelCfg = model.DefaultConfig()
	}
	step := 1.0
	if delta < 0 {
		step = -1.0
	}
	temp := db.NormalizeFallbackTempC(m.modelCfg.FallbackTempC) + step
	if temp < 1 {
		temp = 1
	}
	if temp > 40 {
		temp = 40
	}
	m.modelCfg.FallbackTempC = temp
	return m.store.SetFallbackTempC(temp)
}

func (m *Model) restartWeatherRefreshLoop() tea.Cmd {
	if m.weatherCity.Name == "" {
		return nil
	}
	m.weatherRefreshSeq++
	return weatherRefreshTick(m.weatherRefreshSeq, m.weatherRefreshMinutes)
}

func fetchCurrentWeather(wx weather.Service, loc weather.Location) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 9*time.Second)
		defer cancel()
		cond, err := wx.Current(ctx, loc)
		return weatherLoadedMsg{loc: loc, cond: cond, err: err}
	}
}

func weatherRefreshTick(seq, minutes int) tea.Cmd {
	d := time.Duration(db.NormalizeWeatherRefreshMinutes(minutes)) * time.Minute
	return tea.Tick(d, func(time.Time) tea.Msg {
		return weatherRefreshTickMsg{seq: seq}
	})
}

func searchCities(wx weather.Service, query, lang string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		items, err := wx.SearchCities(ctx, query, lang, 8)
		return citySearchLoadedMsg{query: query, items: items, err: err}
	}
}

func localizeWeatherCity(wx weather.Service, loc weather.Location, lang string) tea.Cmd {
	return func() tea.Msg {
		if loc.Name == "" {
			return weatherCityLocalizedMsg{loc: loc}
		}
		ctx, cancel := context.WithTimeout(context.Background(), 8*time.Second)
		defer cancel()
		items, err := wx.SearchCities(ctx, loc.Name, lang, 10)
		if err != nil {
			return weatherCityLocalizedMsg{loc: loc, err: err}
		}
		return weatherCityLocalizedMsg{loc: weather.LocalizeName(items, loc)}
	}
}

func (m Model) cityLanguageCode() string {
	if m.lang == i18n.RU {
		return "ru"
	}
	return "en"
}

func (m *Model) beginCitySearch() tea.Cmd {
	m.citySelecting = true
	if m.cityInput == "" {
		m.cityInput = m.weatherCity.Name
	}
	m.cityCursor = 0
	m.cityCandidates = nil
	query := strings.TrimSpace(m.cityInput)
	if len([]rune(query)) < 2 {
		return nil
	}
	return searchCities(m.weather, query, m.cityLanguageCode())
}

func (m *Model) confirmCitySelection() tea.Cmd {
	if len(m.cityCandidates) == 0 {
		return nil
	}
	if m.cityCursor < 0 || m.cityCursor >= len(m.cityCandidates) {
		m.cityCursor = 0
	}
	loc := m.cityCandidates[m.cityCursor]
	m.weatherCity = loc
	m.cityInput = loc.Name
	m.citySelecting = false
	m.cityCandidates = nil
	m.cityCursor = 0
	if err := m.store.SetWeatherCity(loc); err != nil {
		m.err = err
		return nil
	}
	m.weatherAppliedDayKey = ""
	if err := m.store.SetWeatherPrecipAppliedDay(""); err != nil {
		m.err = err
		return nil
	}
	m.weatherLoading = true
	m.status = m.cat().SettingsCitySaved(loc.DisplayName())
	return fetchCurrentWeather(m.weather, loc)
}

func (m *Model) backupDatabase() (string, error) {
	path, err := db.DefaultBackupPath()
	if err != nil {
		return "", err
	}
	if err := m.store.BackupTo(path); err != nil {
		return "", err
	}
	return filepath.Dir(path), nil
}
