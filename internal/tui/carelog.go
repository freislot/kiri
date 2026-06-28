package tui

import (
	"strings"

	"kiri/internal/model"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) careLogBodyLines() int {
	overhead := 11
	lines := m.termHeight() - overhead
	if lines < 3 {
		return 3
	}
	return lines
}

func (m Model) careLogVisibleLines() int {
	lines := m.careLogBodyLines() - 1 // column header
	if len(m.careLog) > lines {
		lines-- // scroll indicator
	}
	if lines < 1 {
		return 1
	}
	return lines
}

func (m *Model) clampCareLogOffset() {
	total := len(m.careLog)
	maxLines := m.careLogVisibleLines()
	maxOffset := 0
	if total > maxLines {
		maxOffset = total - maxLines
	}
	if m.careLogOffset > maxOffset {
		m.careLogOffset = maxOffset
	}
	if m.careLogOffset < 0 {
		m.careLogOffset = 0
	}
}

func (m Model) careLogColWidths(maxW int) (dateW, plantW, eventW int) {
	dateW = 15
	plantW = 16
	eventW = maxW - dateW - plantW - 4
	if eventW < 10 {
		eventW = 10
		plantW = maxW - dateW - eventW - 4
		if plantW < 8 {
			plantW = 8
		}
	}
	return dateW, plantW, eventW
}

func (m Model) renderCareLogHeader(maxW int) string {
	c := m.cat()
	dateW, plantW, eventW := m.careLogColWidths(maxW)
	sep := fillSpaces(2)
	date := StyleCalWeekday.Render(truncateWidth(c.CareLogColDate(), dateW))
	plant := StyleCalWeekday.Render(truncateWidth(c.CareLogColPlant(), plantW))
	event := StyleCalWeekday.Render(truncateWidth(c.CareLogColEvent(), eventW))
	return date + sep + plant + sep + event
}

func (m Model) renderCareLogRow(e model.CareLogEntry, maxW int) string {
	c := m.cat()
	dateW, plantW, eventW := m.careLogColWidths(maxW)
	sep := fillSpaces(2)
	date := StyleRowMuted.Render(truncateWidth(c.FormatDateTime(e.CreatedAt), dateW))
	plant := StyleRowText.Render(truncateWidth(e.PlantName, plantW))
	event := StyleRowText.Render(truncateWidth(c.TranslateLogMessage(e.Message), eventW))
	return date + sep + plant + sep + event
}

func (m Model) careLogEntryLines() []string {
	maxW := m.innerTextWidth()
	lines := make([]string, 0, len(m.careLog))
	for _, e := range m.careLog {
		lines = append(lines, m.renderCareLogRow(e, maxW))
	}
	return lines
}

func (m Model) renderCareLogTab() string {
	c := m.cat()
	title := StyleSectionTitle.Render(c.CareLogTitle())

	if len(m.careLog) == 0 {
		return m.renderSection(title + "\n" + StyleRowMuted.Render(c.NoCareEvents()))
	}

	maxW := m.innerTextWidth()
	bodyLines := m.careLogVisibleLines()
	entries := m.careLogEntryLines()
	total := len(entries)

	start := m.careLogOffset
	end := start + bodyLines
	if end > total {
		end = total
	}
	visible := entries[start:end]
	for len(visible) < bodyLines {
		visible = append(visible, "")
	}

	var parts []string
	parts = append(parts, title)
	parts = append(parts, m.renderCareLogHeader(maxW))
	if total > bodyLines {
		parts = append(parts, m.renderCareLogScrollIndicator(start, end, total))
	}
	parts = append(parts, visible...)

	return m.renderSection(strings.Join(parts, "\n"))
}

func (m Model) renderCareLogScrollIndicator(start, end, total int) string {
	c := m.cat()

	up := StyleScrollBar.Render("▲")
	if m.careLogOffset == 0 {
		up = StyleScrollIndicator.Render("▲")
	}
	down := StyleScrollBar.Render("▼")
	if end >= total {
		down = StyleScrollIndicator.Render("▼")
	}

	label := StyleScrollIndicator.Render(c.CareLogScrollRange(start+1, end, total))
	sep := fillSpaces(1)
	return up + sep + label + sep + down
}

func (m Model) handleCareLogKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	step := 1
	switch msg.String() {
	case "down", "j":
		m.careLogOffset += step
	case "up", "k":
		m.careLogOffset -= step
	case "pgdown", "ctrl+d":
		m.careLogOffset += m.careLogVisibleLines()
	case "pgup", "ctrl+u":
		m.careLogOffset -= m.careLogVisibleLines()
	default:
		return m, nil
	}
	m.clampCareLogOffset()
	return m, nil
}
