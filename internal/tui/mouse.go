package tui

import (
	"strings"

	"kiri/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) frameContentX() int {
	return appHPad/2 + frameHBorder/2 + frameHPad/2
}

func (m Model) frameHeaderY() int {
	return 1
}

func (m Model) contentStartY() int {
	return m.frameHeaderY() + 1
}

func (m Model) contentBodyHeight() int {
	h := m.termHeight() - 5
	if h < 1 {
		return 1
	}
	return h
}

func (m Model) footerY() int {
	return m.contentStartY() + m.contentBodyHeight() + 1
}

func (m Model) plantsGridY() int {
	return m.contentStartY() + 1
}

func footerHintClickKey(h i18n.FooterHint) (string, bool) {
	key := strings.TrimSpace(h.Key)
	switch {
	case strings.HasPrefix(key, "⌨"), strings.Contains(key, "PgUp"), key == "h/l", key == "← →", key == "1–5", key == "Tab":
		return "", false
	case key == "n/Esc":
		return "n", true
	case key == "Enter/y", key == "Enter/Space":
		return "enter", true
	case key == "Space":
		return " ", true
	case key == "Esc":
		return "esc", true
	case len(key) == 1:
		return key, true
	default:
		return "", false
	}
}

type tabRegion struct {
	tab      Tab
	x0, x1 int
}

func (m Model) tabRegions() []tabRegion {
	labels := m.tabLabels()
	sep := fillStyle().Render("   ")
	sepW := lipgloss.Width(sep)
	x := m.frameContentX()

	var regions []tabRegion
	for i, label := range labels {
		var rendered string
		if Tab(i) == m.tab {
			rendered = StyleTabActive.Render(label)
		} else {
			rendered = StyleTabInactive.Render(label)
		}
		w := lipgloss.Width(rendered)
		regions = append(regions, tabRegion{tab: Tab(i), x0: x, x1: x + w})
		x += w
		if i < len(labels)-1 {
			x += sepW
		}
	}
	return regions
}

func (m Model) tabAt(x, y int) (Tab, bool) {
	if y != m.frameHeaderY() {
		return 0, false
	}
	for _, r := range m.tabRegions() {
		if x >= r.x0 && x < r.x1 {
			return r.tab, true
		}
	}
	return 0, false
}

func (m Model) plantGridRowHeights() []int {
	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	vc := m.visCols
	if vc < 1 {
		vc = maxPlantCols
	}
	totalCols := 0
	if len(m.plants) > 0 {
		totalCols = (len(m.plants) + vr - 1) / vr
	}
	startCol := m.viewportCol

	heights := make([]int, vr)
	for row := 0; row < vr; row++ {
		rowH := cardHeight - cardBorderV
		for i := 0; i < vc; i++ {
			col := startCol + i
			cardW := m.plantCardContentWidth(i)
			idx := row + col*vr
			if col >= totalCols || idx >= len(m.plants) {
				continue
			}
			content, _ := m.plantCardBlock(m.plants[idx], false, cardW)
			if h := lipgloss.Height(content); h > rowH {
				rowH = h
			}
			if rowH > cardBodyMax {
				rowH = cardBodyMax
			}
		}
		heights[row] = rowH + cardBorderV
	}
	return heights
}

func (m Model) plantCardAt(x, y int) (row, col int, ok bool) {
	if m.tab != TabPlants || len(m.plants) == 0 {
		return 0, 0, false
	}

	slots := m.plantCardSlotWidths()
	vc := m.visCols
	if vc < 1 {
		vc = maxPlantCols
	}
	if x < m.frameContentX() {
		return 0, 0, false
	}

	relX := x - m.frameContentX()
	col = -1
	x0 := 0
	for i := 0; i < vc && i < len(slots); i++ {
		if relX >= x0 && relX < x0+slots[i] {
			col = i
			break
		}
		x0 += slots[i]
	}
	if col < 0 {
		return 0, 0, false
	}

	relY := y - m.plantsGridY()
	if relY < 0 {
		return 0, 0, false
	}
	yOff := 0
	for r, h := range m.plantGridRowHeights() {
		if relY >= yOff && relY < yOff+h {
			row = r
			return row, col, true
		}
		yOff += h
	}
	return 0, 0, false
}

func (m Model) currentFooterHints() []i18n.FooterHint {
	c := m.cat()
	if m.overlay != OverlayNone {
		switch m.overlay {
		case OverlayBackupConfirm:
			return c.FooterBackupConfirmHints()
		default:
			return c.FooterOverlayHints()
		}
	}
	switch m.tab {
	case TabCalendar:
		return c.FooterCalendarHints()
	case TabCareLog:
		return c.FooterCareLogHints()
	case TabSettings:
		return c.FooterSettingsHints()
	case TabAbout:
		return c.FooterAboutHints()
	default:
		return c.FooterHints()
	}
}

func (m Model) footerRegionAt(x, y int) (string, bool) {
	var rowW int
	var footerRow int
	var originX int

	if m.tab == TabAbout && m.overlay == OverlayNone {
		footerRow = m.termHeight() - 1
		rowW = m.termWidth()
		originX = 0
	} else {
		footerRow = m.footerY()
		rowW = m.headerTextWidth()
		originX = m.frameContentX()
	}

	if y != footerRow {
		return "", false
	}

	_, regions := footerHintLayout(m.currentFooterHints(), rowW)
	for _, r := range regions {
		if x >= originX+r.X0 && x < originX+r.X1 {
			return r.Key, true
		}
	}
	return "", false
}

func synthKey(key string) tea.KeyMsg {
	switch key {
	case "enter":
		return tea.KeyMsg{Type: tea.KeyEnter}
	case "esc":
		return tea.KeyMsg{Type: tea.KeyEscape}
	case " ":
		return tea.KeyMsg{Type: tea.KeySpace}
	case "tab":
		return tea.KeyMsg{Type: tea.KeyTab}
	case "left":
		return tea.KeyMsg{Type: tea.KeyLeft}
	case "right":
		return tea.KeyMsg{Type: tea.KeyRight}
	case "up":
		return tea.KeyMsg{Type: tea.KeyUp}
	case "down":
		return tea.KeyMsg{Type: tea.KeyDown}
	default:
		if len(key) == 1 {
			return tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{rune(key[0])}}
		}
	}
	return tea.KeyMsg{}
}

func (m Model) activateTab(t Tab) (tea.Model, tea.Cmd) {
	m.tab = t
	m.status = ""
	if m.tab == TabCareLog {
		m.clampCareLogOffset()
	}
	if m.tab == TabCalendar {
		m.calFocus = calFocusGrid
		m.taskCursor = 0
	}
	return m, m.aboutTabCmd()
}

func (m Model) selectPlantCard(row, col int) (tea.Model, tea.Cmd) {
	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	m.cursorRow = row
	m.cursorCol = m.viewportCol + col
	m.clampCursor()
	m.scrollToFocused()
	m.status = ""
	return m, m.reloadPlantCareLog()
}

func (m Model) handleClickKey(key string) (tea.Model, tea.Cmd) {
	msg := synthKey(key)
	if msg.Type == 0 && len(msg.Runes) == 0 {
		return m, nil
	}

	if m.overlay != OverlayNone {
		if key == "q" {
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

	switch key {
	case "ctrl+c", "q":
		return m, tea.Quit
	}

	switch m.tab {
	case TabAbout:
		if key == "q" {
			return m, tea.Quit
		}
		return m, nil
	case TabSettings:
		return m.handleSettingsKey(msg)
	case TabCalendar:
		return m.handleCalendarKey(msg)
	case TabCareLog:
		return m.handleCareLogKey(msg)
	case TabPlants:
		return m.handlePlantsKey(msg)
	}
	return m, nil
}

func (m Model) careLogPanelContains(x, y int) bool {
	if m.tab != TabCareLog || m.overlay != OverlayNone {
		return false
	}
	contentX := m.frameContentX()
	contentY := m.contentStartY()
	contentW := m.headerTextWidth()
	contentH := m.contentBodyHeight()
	return x >= contentX && x < contentX+contentW && y >= contentY && y < contentY+contentH
}

func (m Model) settingsPanelContains(x, y int) bool {
	if m.tab != TabSettings || m.overlay != OverlayNone {
		return false
	}
	contentX := m.frameContentX()
	contentY := m.contentStartY()
	contentW := m.headerTextWidth()
	contentH := m.contentBodyHeight()
	return x >= contentX && x < contentX+contentW && y >= contentY && y < contentY+contentH
}

func (m Model) careLogWheelStep() int {
	step := m.careLogVisibleLines()
	if step < 3 {
		return 3
	}
	return step
}

func (m Model) careLogScrollBy(delta int) (tea.Model, tea.Cmd) {
	if delta == 0 {
		return m, nil
	}
	m.careLogOffset += delta
	m.clampCareLogOffset()
	return m, nil
}

func (m Model) handleCareLogWheel(button tea.MouseButton) (tea.Model, tea.Cmd) {
	step := m.careLogWheelStep()
	switch button {
	case tea.MouseButtonWheelUp:
		return m.careLogScrollBy(-step)
	case tea.MouseButtonWheelDown:
		return m.careLogScrollBy(step)
	default:
		return m, nil
	}
}

func (m Model) handleSettingsWheel(button tea.MouseButton) (tea.Model, tea.Cmd) {
	if m.citySelecting {
		return m, nil
	}
	switch button {
	case tea.MouseButtonWheelUp:
		m.moveSettingsCursor(-1)
	case tea.MouseButtonWheelDown:
		m.moveSettingsCursor(1)
	default:
		return m, nil
	}
	return m, nil
}

func (m Model) handleMouse(msg tea.MouseMsg) (tea.Model, tea.Cmd) {
	if m.splashPhase != splashPhaseDone {
		return m, nil
	}
	if (m.width > 0 && m.width < 40) || (m.height > 0 && m.height < 10) {
		return m, nil
	}

	if tea.MouseEvent(msg).IsWheel() && msg.Action == tea.MouseActionPress {
		if m.careLogPanelContains(msg.X, msg.Y) {
			return m.handleCareLogWheel(msg.Button)
		}
		if m.settingsPanelContains(msg.X, msg.Y) {
			return m.handleSettingsWheel(msg.Button)
		}
		return m, nil
	}

	if msg.Action != tea.MouseActionPress || msg.Button != tea.MouseButtonLeft {
		return m, nil
	}

	x, y := msg.X, msg.Y

	if m.tab != TabAbout || m.overlay != OverlayNone {
		if tab, ok := m.tabAt(x, y); ok && tab != m.tab {
			return m.activateTab(tab)
		}
	}

	if key, ok := m.footerRegionAt(x, y); ok {
		return m.handleClickKey(key)
	}

	if row, col, ok := m.plantCardAt(x, y); ok {
		return m.selectPlantCard(row, col)
	}

	return m, nil
}
