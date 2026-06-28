package tui

import (
	"fmt"
	"strings"

	"kiri/internal/i18n"
	"kiri/internal/model"

	"github.com/charmbracelet/lipgloss"
)


func (m Model) View() string {
	c := m.cat()

	if m.err != nil {
		return StyleApp.Render(c.ErrorQuit(m.err))
	}

	switch m.splashPhase {
	case splashPhaseReveal:
		return m.renderSplashReveal()
	case splashPhaseShow:
		return m.renderSplashScreen()
	case splashPhaseFade:
		return m.renderSplashFade()
	}

	return m.renderMainApp()
}

func (m Model) renderMainApp() string {
	if (m.width > 0 && m.width < 40) || (m.height > 0 && m.height < 10) {
		return lipgloss.NewStyle().Foreground(colorWarning).Bold(true).
			Render("Terminal too small!")
	}

	if m.tab == TabAbout && m.overlay == OverlayNone {
		return m.renderAboutScreen()
	}

	base := m.renderBaseApp()

	if m.overlay != OverlayNone {
		popup := m.renderOverlay()
		return overlayCenter(base, popup, m.termWidth(), m.termHeight())
	}

	return base
}

func (m Model) termHeight() int {
	if m.height > 0 {
		return m.height
	}
	return 24
}

func (m Model) renderBaseApp() string {
	statusText := ""
	if m.overlay == OverlayNone {
		statusText = m.status
	}

	body := strings.Join([]string{
		m.renderHeaderBlock(),
		m.renderContent(),
		m.renderStatusLine(statusText),
		m.renderFooter(),
	}, "\n")

	return StyleApp.Render(StyleFrame.Render(body))
}


func (m Model) renderHeaderBlock() string {
	maxW := m.headerTextWidth()
	if maxW < 1 {
		return ""
	}

	left := m.renderNavLine()
	right := StyleWeather.Render(m.weatherHeaderText())
	leftW := lipgloss.Width(left)
	rightW := lipgloss.Width(right)

	if leftW+rightW > maxW {
		availLeft := maxW - rightW - 1
		if availLeft < 0 {
			availLeft = 0
		}
		left = truncateWidth(left, availLeft)
		leftW = lipgloss.Width(left)
	}

	gap := maxW - leftW - rightW
	if gap < 1 {
		gap = 1
	}
	gapFill := fillStyle().Render(strings.Repeat(" ", gap))

	line := left + gapFill + right
	if lipgloss.Width(line) > maxW {
		line = truncateWidth(line, maxW)
	}
	return line
}

func (m Model) renderNavLine() string {
	labels := m.tabLabels()
	var tabs []string
	for i, label := range labels {
		if Tab(i) == m.tab {
			tabs = append(tabs, StyleTabActive.Render(label))
		} else {
			tabs = append(tabs, StyleTabInactive.Render(label))
		}
	}
	sep := fillStyle().Render("   ")
	return strings.Join(tabs, sep)
}


func (m Model) renderContent() string {
	var content string
	switch m.tab {
	case TabPlants:
		content = m.renderPlantsTab()
	case TabCalendar:
		content = m.renderCalendarTab()
	case TabCareLog:
		content = m.renderCareLogTab()
	case TabSettings:
		content = m.renderSettingsTab()
	case TabAbout:
		content = ""
	default:
		content = ""
	}
	return m.surfaceNormalize(content)
}

func (m Model) surfaceNormalize(block string) string {
	width := m.headerTextWidth()
	if width < 1 {
		return block
	}
	bg := fillStyle()
	lines := strings.Split(block, "\n")
	for i, line := range lines {
		w := lipgloss.Width(line)
		if w < width {
			lines[i] = line + bg.Render(strings.Repeat(" ", width-w))
		}
	}

	targetH := m.termHeight() - 5
	if targetH < 1 {
		targetH = 1
	}
	fill := bg.Render(strings.Repeat(" ", width))
	for len(lines) < targetH {
		lines = append(lines, fill)
	}
	if len(lines) > targetH {
		lines = lines[:targetH]
	}
	return strings.Join(lines, "\n")
}


func (m Model) renderPlantsTab() string {
	if len(m.plants) == 0 {
		return m.renderSection(
			"\n" + StyleRowMuted.Render(m.cat().NoPlants()),
		)
	}

	parts := []string{"", m.renderGrid()}
	if ind := m.renderScrollIndicator(); ind != "" {
		parts = append(parts, ind)
	}

	details := lipgloss.NewStyle().
		MarginTop(1).
		Render(m.renderSection(m.renderDetailPanel(m.innerTextWidth())))
	parts = append(parts, details, "")

	return strings.Join(parts, "\n")
}


func (m Model) renderGrid() string {
	vr := m.visRows
	if vr < 1 {
		vr = 2
	}
	vc := m.visCols
	if vc < 1 {
		vc = maxPlantCols
	}

	totalCols := (len(m.plants) + vr - 1) / vr
	startCol := m.viewportCol

	var rows []string
	for row := 0; row < vr; row++ {
		type gridCell struct {
			cardW   int
			content string
			style   lipgloss.Style
			empty   bool
		}
		cells := make([]gridCell, 0, vc)
		rowH := cardHeight - 2
		for i := 0; i < vc; i++ {
			cardW := m.plantCardContentWidth(i)
			col := startCol + i
			focused := row == m.cursorRow && col == m.cursorCol
			idx := row + col*vr
			if col >= totalCols || idx >= len(m.plants) {
				cells = append(cells, gridCell{cardW: cardW, empty: true})
				continue
			}
			content, style := m.plantCardBlock(m.plants[idx], focused, cardW)
			if h := lipgloss.Height(content); h > rowH {
				rowH = h
			}
			if rowH > cardBodyMax {
				rowH = cardBodyMax
			}
			cells = append(cells, gridCell{cardW: cardW, content: content, style: style})
		}

		cards := make([]string, 0, len(cells))
		for _, cl := range cells {
			if cl.empty {
				cards = append(cards, StyleCardInactive.Width(cl.cardW).Height(rowH).Render(" "))
				continue
			}
			cards = append(cards, cl.style.Width(cl.cardW).Height(rowH).Render(cl.content))
		}
		rows = append(rows, lipgloss.JoinHorizontal(lipgloss.Top, cards...))
	}

	if len(rows) == 0 {
		return ""
	}
	return lipgloss.JoinVertical(lipgloss.Left, rows...)
}

func (m Model) renderScrollIndicator() string {
	vr := m.visRows
	if vr < 1 {
		return ""
	}
	vc := m.visCols
	if vc < 1 {
		return ""
	}

	totalCols := (len(m.plants) + vr - 1) / vr
	if totalCols <= vc {
		return "" // all columns visible — no indicator
	}

	const trackW = 20
	maxScroll := totalCols - vc
	if maxScroll < 1 {
		maxScroll = 1
	}

	thumbW := trackW * vc / totalCols
	if thumbW < 1 {
		thumbW = 1
	}
	thumbPos := int(float64(m.viewportCol)/float64(maxScroll)*float64(trackW-thumbW) + 0.5)
	if thumbPos+thumbW > trackW {
		thumbPos = trackW - thumbW
	}
	if thumbPos < 0 {
		thumbPos = 0
	}

	track := make([]rune, trackW)
	for i := range track {
		if i >= thumbPos && i < thumbPos+thumbW {
			track[i] = '█'
		} else {
			track[i] = '░'
		}
	}

	endVisible := m.viewportCol + vc
	if endVisible > totalCols {
		endVisible = totalCols
	}

	leftArr := StyleRowMuted.Render("◀")
	rightArr := StyleRowMuted.Render("▶")
	if m.viewportCol > 0 {
		leftArr = StyleScrollBar.Render("◀")
	}
	if m.viewportCol+vc < totalCols {
		rightArr = StyleScrollBar.Render("▶")
	}

	bar := StyleScrollBar.Render(string(track))
	label := StyleRowMuted.Render(
		fmt.Sprintf("col %d–%d / %d", m.viewportCol+1, endVisible, totalCols),
	)

	bg := fillStyle()
	return bg.Render("  ") +
		leftArr +
		bg.Render(" [") +
		bar +
		bg.Render("] ") +
		rightArr +
		bg.Render("   ") +
		label
}


func (m Model) plantCardBlock(p model.Plant, focused bool, contentW int) (string, lipgloss.Style) {
	c := m.cat()
	tw := contentW - cardHPad
	if tw < 1 {
		tw = 1
	}

	icon := plantIcon(p)
	nameLine := styleWrappedMax(cardNameStyle(focused), icon+" "+p.Name, tw, 2)

	locLine := styleWrappedMax(cardLocationStyle(focused), "📍 "+c.Location(p.Location), tw, 1)

	barW := tw - 7
	if barW < 4 {
		barW = 4
	}
	bg := cardBg(focused)
	bar := styledBarOnBg(p.WaterLevel, barW, bg, focused)
	pct := styledPercentOnBg(p.WaterLevel, bg, focused)
	barGap := " "
	if !skipBackgrounds() {
		barGap = lipgloss.NewStyle().Background(bg).Render(" ")
	}
	barLine := bar + barGap + pct

	actionLine := m.cardActionLine(p, tw, focused)

	content := strings.Join([]string{nameLine, locLine, barLine, actionLine}, "\n")

	urgent := p.State == model.StateWaterNow
	return content, cardBorderStyle(focused, urgent)
}

func plantIcon(p model.Plant) string {
	switch p.State {
	case model.StateWaterNow:
		return "💧"
	case model.StateShiftedByRain:
		return "🌧"
	default:
		return "🌿"
	}
}

func (m Model) cardActionLine(p model.Plant, tw int, focused bool) string {
	c := m.cat()
	prefix := "📅 "
	switch p.State {
	case model.StateWaterNow:
		return styleWrappedMax(cardActionUrgentStyle(focused), prefix+c.RowStatusWaterNow(), tw, 1)
	case model.StateShiftedByRain:
		return styleWrappedMax(cardActionRainStyle(focused), prefix+c.RowStatusShiftedRain(wateringDate(p, m.effectiveModelCfg())), tw, 1)
	default:
		return styleWrappedMax(cardActionStyle(focused), prefix+m.rowStatusText(p), tw, 1)
	}
}


func (m Model) renderDetailPanel(maxW int) string {
	c := m.cat()
	p := m.selectedPlant()
	if p == nil {
		return StyleSectionTitle.Render(c.DetailsTitle()) + "\n" +
			StyleRowMuted.Render(c.SelectPlant())
	}

	title := StyleSectionTitle.Render(c.DetailsTitlePlant(p.Name, c.PlantTypeLabel(*p)))
	bullets := []string{
		"• " + StyleDetailBullet.Render(truncateWidth(c.DetailBaseCycle(p.BaseIntervalDays), maxW-2)),
		"• " + StyleDetailBullet.Render(truncateWidth(c.DetailStatusPrefix()+m.detailStatusText(*p), maxW-2)),
		"• " + StyleDetailBullet.Render(truncateWidth(c.DetailLastLogPrefix()+m.lastLogText(m.detailLog), maxW-2)),
	}
	if p.ConsecutivePostpones >= 2 {
		bullets = append(bullets,
			"• "+StyleStatusMsg.Render(truncateWidth(c.PostponeTip(p.ConsecutivePostpones), maxW-2)),
		)
	}
	return title + "\n" + strings.Join(bullets, "\n")
}


func (m Model) renderSettingsTab() string {
	c := m.cat()
	rows := []string{StyleSectionTitle.Render(c.SettingsTitle())}
	rows = append(rows, m.renderSettingsLanguageRow(m.cursor == settingsRowLanguage))
	rows = append(rows, m.renderSettingsCityRow(m.cursor == settingsRowCity))
	if m.citySelecting {
		if list := m.renderCitySuggestions(); list != "" {
			rows = append(rows, list)
			rows = append(rows, "")
		}
	}
	rows = append(rows, m.renderSettingsWeatherRefreshRow(m.cursor == settingsRowWeatherRefresh))
	rows = append(rows, m.renderSettingsDefaultIntervalRow(m.cursor == settingsRowDefaultInterval))
	rows = append(rows, m.renderSettingsFallbackTempRow(m.cursor == settingsRowFallbackTemp))
	rows = append(rows, m.renderSettingsBinaryRow(m.cursor == settingsRowAutoBackup, c.SettingsOptionAutoBackup(), m.autoBackup))
	rows = append(rows, m.renderSettingsBinaryRow(m.cursor == settingsRowTransparent, c.SettingsOptionTransparent(), m.transparent))
	rows = append(rows, m.renderSettingsBinaryRow(m.cursor == settingsRowFastBoot, c.SettingsOptionFastBoot(), m.fastBoot))
	section := m.renderSection(strings.Join(rows, "\n"))
	info := strings.Join([]string{
		m.renderSettingsInfoPathRow(c.SettingsOptionDatabase(), m.store.DBPath()),
		m.renderSettingsInfoPathRow(c.SettingsOptionBackup(), m.store.DBPath()+".bak"),
		m.renderSettingsInfoPathRow(c.SettingsOptionConfig(), m.store.SettingsFilePath()),
		m.renderSettingsInfoPathRow(c.SettingsOptionConfigBackup(), m.store.SettingsBackupFilePath()),
	}, "\n")
	return section + "\n" + m.renderSettingsDescriptionRow() + "\n\n" + info
}

func (m Model) renderFooter() string {
	c := m.cat()
	var hints []i18n.FooterHint
	if m.overlay != OverlayNone {
		switch m.overlay {
		case OverlayBackupConfirm:
			hints = c.FooterBackupConfirmHints()
		default:
			hints = c.FooterOverlayHints()
		}
	} else if m.tab == TabCalendar {
		hints = c.FooterCalendarHints()
	} else if m.tab == TabCareLog {
		hints = c.FooterCareLogHints()
	} else if m.tab == TabSettings {
		hints = c.FooterSettingsHints()
	} else if m.tab == TabAbout {
		hints = c.FooterAboutHints()
	} else {
		hints = c.FooterHints()
	}

	maxW := m.headerTextWidth()
	if maxW < 1 {
		return ""
	}

	return renderFooterHints(hints, maxW)
}

func (m Model) renderStatusLine(status string) string {
	maxW := m.headerTextWidth()
	if maxW < 1 {
		return ""
	}

	text := StyleStatusMsg.Render(truncateWidth(status, maxW))
	if pad := maxW - lipgloss.Width(text); pad > 0 {
		text += fillSpaces(pad)
	}
	return text
}
