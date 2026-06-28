package tui

import (
	"fmt"
	"strings"
	"time"

	"kiri/internal/model"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/mattn/go-runewidth"
)

const (
	calColWidth  = 7 // inner column width for each day cell
	calIconWidth = 2 // terminal cells reserved for day-status icon slot
	calGridWidth = 7*calColWidth + 8 // manual grid: corners + cols + column seps

	calIconUrgent = "❗"
	calIconDue    = "💧"
	calIconEmpty  = "  "
)

type calCellMode int

const (
	calCellMulti   calCellMode = iota // day + icon on two lines
	calCellCompact                    // day + icon on one line (6-week months / short terminal)
)

type calFocus int

const (
	calFocusGrid calFocus = iota
	calFocusTasks
)

type calDayCell struct {
	day      int
	selected bool
	today    bool
	hasDue   bool
	urgent   bool
	icon     string
}

type CalendarTask struct {
	PlantID   int64
	PlantName string
	Date      time.Time
	Urgent    bool
	Done      bool
	DoneByLog bool
}

func (m Model) calSelectedDate() time.Time {
	day := m.calSelectedDay
	if day < 1 {
		day = 1
	}
	return time.Date(m.calYear, time.Month(m.calMonth), day, 0, 0, 0, 0, time.Local)
}

func (m *Model) moveCalendarDay(delta int) {
	d := m.calSelectedDate().AddDate(0, 0, delta)
	m.calYear = d.Year()
	m.calMonth = int(d.Month())
	m.calSelectedDay = d.Day()
	m.taskCursor = 0
}

func (m Model) calendarRowInnerBudget() int {
	w := m.headerTextWidth() - 1 - 2*sectionHBorder
	if w < 40 {
		return 40
	}
	return w
}

func (m Model) calendarPanelWidths() (calW, sideW int) {
	budget := m.calendarRowInnerBudget()
	minCal := calGridWidth + sectionHPad
	const sideMin = 24

	calW = budget * 62 / 100
	sideW = budget - calW

	if calW < minCal {
		calW = minCal
		sideW = budget - calW
	}
	if sideW < sideMin {
		sideW = sideMin
		calW = budget - sideW
		if calW < minCal && budget >= minCal+sideMin {
			calW = minCal
			sideW = budget - calW
		}
	}
	if calW < 1 {
		calW = 1
	}
	if sideW < 1 {
		sideW = 1
	}
	return calW, sideW
}

func (m Model) calendarPanelWidth() int {
	calW, _ := m.calendarPanelWidths()
	return calW
}

func (m Model) tasksForDate(day time.Time) []CalendarTask {
	now := time.Now()
	d := dateOnly(day)
	var tasks []CalendarTask
	for _, p := range m.plants {
		due, urgent := model.IsDueOnDateWithConfig(p, d, now, m.effectiveModelCfg())
		if !due {
			continue
		}
		done, doneByLog := m.taskDoneStateForDate(p.ID, d)
		tasks = append(tasks, CalendarTask{
			PlantID:   p.ID,
			PlantName: p.Name,
			Date:      d,
			Urgent:    urgent,
			Done:      done,
			DoneByLog: doneByLog,
		})
	}
	return tasks
}

func (m Model) taskDoneStateForDate(plantID int64, day time.Time) (done bool, doneByTaskLog bool) {
	targetDay := dateOnly(day)
	for _, e := range m.careLog {
		if e.PlantID != plantID {
			continue
		}
		if !dateOnly(e.CreatedAt).Equal(targetDay) {
			continue
		}
		switch e.EventType {
		case "task_done":
			return true, true
		case "watered":
			return true, false
		}
	}
	return false, false
}

func dateOnly(t time.Time) time.Time {
	y, mo, d := t.Date()
	return time.Date(y, mo, d, 0, 0, 0, 0, t.Location())
}

func (m Model) dayStatusIcon(day time.Time) string {
	now := time.Now()
	hasUrgent := false
	hasDue := false
	for _, p := range m.plants {
		due, urgent := model.IsDueOnDateWithConfig(p, day, now, m.effectiveModelCfg())
		if !due {
			continue
		}
		if urgent {
			hasUrgent = true
		} else {
			hasDue = true
		}
	}
	return calDayIcon(hasUrgent, hasDue)
}

func calDayIcon(urgent, hasDue bool) string {
	if urgent {
		return calIconUrgent
	}
	if hasDue {
		return calIconDue
	}
	return calIconEmpty
}

func calIconSlot(icon string) string {
	if icon == "" {
		icon = calIconEmpty
	}
	w := runewidth.StringWidth(icon)
	if w > calIconWidth {
		return runewidth.Truncate(icon, calIconWidth, "")
	}
	if w < calIconWidth {
		return icon + strings.Repeat(" ", calIconWidth-w)
	}
	return icon
}

func padCenterTerm(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := runewidth.StringWidth(s)
	if w >= width {
		return runewidth.Truncate(s, width, "")
	}
	pad := width - w
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}


func (m Model) renderCalendarTab() string {
	left := m.renderCalendarPanel()
	right := m.renderCalendarSidebar()
	targetH := lipgloss.Height(left)
	rightW := lipgloss.Width(right)
	if lipgloss.Height(right) < targetH || rightW > 0 {
		right = styleBg(
			lipgloss.NewStyle().Width(rightW).Height(targetH),
			colorCardBg,
		).Render(right)
	}
	gap := fillStyle().Width(1).Height(targetH).Render("")
	return lipgloss.JoinHorizontal(lipgloss.Top, left, gap, right)
}

func (m Model) renderCalendarPanel() string {
	c := m.cat()
	panelW := m.calendarPanelWidth()

	header := StyleSectionTitle.Render(c.CalendarMonthTitle(m.calYear, time.Month(m.calMonth)))
	grid := m.renderCalendarTable()

	content := header + "\n" + grid

	return StyleSection.
		Width(panelW).
		Render(content)
}

func (m Model) calWeekCount() int {
	first := time.Date(m.calYear, time.Month(m.calMonth), 1, 0, 0, 0, 0, time.Local)
	daysInMonth := time.Date(m.calYear, time.Month(m.calMonth)+1, 0, 0, 0, 0, 0, time.Local).Day()
	startOffset := int(first.Weekday())
	return (startOffset + daysInMonth + 6) / 7
}

func (m Model) calTabAvailLines() int {
	const overhead = 10
	h := m.termHeight() - overhead
	if h < 8 {
		return 8
	}
	return h
}

func (m Model) calCellMode() calCellMode {
	weeks := m.calWeekCount()
	avail := m.calTabAvailLines()

	if weeks >= 6 {
		return calCellCompact
	}

	multiNeed := 1 + 1 + weeks*3
	if multiNeed <= avail {
		return calCellMulti
	}
	return calCellCompact
}

func (m Model) buildCalendarGrid() [][]calDayCell {
	first := time.Date(m.calYear, time.Month(m.calMonth), 1, 0, 0, 0, 0, time.Local)
	daysInMonth := time.Date(m.calYear, time.Month(m.calMonth)+1, 0, 0, 0, 0, 0, time.Local).Day()
	startOffset := int(first.Weekday())

	weeks := m.calWeekCount()

	grid := make([][]calDayCell, weeks)
	for r := range grid {
		grid[r] = make([]calDayCell, 7)
	}

	cell := 0
	for r := 0; r < weeks; r++ {
		for col := 0; col < 7; col++ {
			if cell < startOffset || cell-startOffset >= daysInMonth {
				grid[r][col] = calDayCell{}
			} else {
				day := cell - startOffset + 1
				d := time.Date(m.calYear, time.Month(m.calMonth), day, 0, 0, 0, 0, time.Local)
				hasDue, urgent := false, false
				now := time.Now()
				for _, p := range m.plants {
					due, isUrgent := model.IsDueOnDateWithConfig(p, d, now, m.effectiveModelCfg())
					if due {
						hasDue = true
						if isUrgent {
							urgent = true
						}
					}
				}
				grid[r][col] = calDayCell{
					day:      day,
					selected: day == m.calSelectedDay,
					today:    dateOnly(d).Equal(dateOnly(time.Now())),
					hasDue:   hasDue,
					urgent:   urgent,
					icon:     calDayIcon(urgent, hasDue),
				}
			}
			cell++
		}
	}
	return grid
}

func (m Model) renderCalendarTable() string {
	c := m.cat()
	grid := m.buildCalendarGrid()
	gridFocused := m.calFocus == calFocusGrid
	mode := m.calCellMode()
	return m.renderCalendarGrid(grid, c.CalendarWeekdays(), mode, gridFocused)
}

func (m Model) renderCalendarGrid(grid [][]calDayCell, weekdays []string, mode calCellMode, gridFocused bool) string {
	border := lipgloss.RoundedBorder()
	borderStyle := lipgloss.NewStyle().Foreground(colorBorderFaint)
	if !skipBackgrounds() {
		borderStyle = borderStyle.Background(colorCardBg)
	}

	linesPerWeek := 1
	if mode == calCellMulti {
		linesPerWeek = 2
	}

	var out strings.Builder
	out.WriteString(calGridTopBorder(border, borderStyle))
	out.WriteByte('\n')

	header := make([]string, 7)
	for i, wd := range weekdays {
		header[i] = StyleCalWeekday.Width(calColWidth).MaxWidth(calColWidth).
			Render(truncateWidth(wd, calColWidth))
	}
	out.WriteString(calGridRow(header, border, borderStyle))
	out.WriteByte('\n')
	out.WriteString(calGridHeaderSep(border, borderStyle))
	out.WriteByte('\n')

	for r, week := range grid {
		for line := 0; line < linesPerWeek; line++ {
			cells := make([]string, 7)
			for col, cell := range week {
				cells[col] = m.renderCalendarCellLine(cell, line, mode, gridFocused)
			}
			out.WriteString(calGridRow(cells, border, borderStyle))
			out.WriteByte('\n')
		}
		if r < len(grid)-1 {
			out.WriteString(calGridRowSep(border, borderStyle))
			out.WriteByte('\n')
		}
	}

	out.WriteString(calGridBottomBorder(border, borderStyle))
	return out.String()
}

func calGridTopBorder(b lipgloss.Border, bs lipgloss.Style) string {
	return calGridHBorder(b.TopLeft, b.Top, b.TopRight, b.MiddleTop, bs)
}

func calGridBottomBorder(b lipgloss.Border, bs lipgloss.Style) string {
	return calGridHBorder(b.BottomLeft, b.Bottom, b.BottomRight, b.MiddleBottom, bs)
}

func calGridHeaderSep(b lipgloss.Border, bs lipgloss.Style) string {
	return calGridHBorder(b.MiddleLeft, b.Top, b.MiddleRight, b.Middle, bs)
}

func calGridRowSep(b lipgloss.Border, bs lipgloss.Style) string {
	return calGridHBorder(b.MiddleLeft, b.Bottom, b.MiddleRight, b.Middle, bs)
}

func calGridHBorder(left, mid, right, sep string, bs lipgloss.Style) string {
	var s strings.Builder
	s.WriteString(bs.Render(left))
	for col := 0; col < 7; col++ {
		s.WriteString(bs.Render(strings.Repeat(mid, calColWidth)))
		if col < 6 {
			s.WriteString(bs.Render(sep))
		}
	}
	s.WriteString(bs.Render(right))
	return s.String()
}

func calGridRow(cells []string, b lipgloss.Border, bs lipgloss.Style) string {
	var s strings.Builder
	s.WriteString(bs.Render(b.Left))
	for i, cell := range cells {
		s.WriteString(cell)
		if i < len(cells)-1 {
			s.WriteString(bs.Render(b.Left))
		}
	}
	s.WriteString(bs.Render(b.Right))
	return s.String()
}

func (m Model) renderCalendarCellLine(cell calDayCell, line int, mode calCellMode, gridFocused bool) string {
	text := m.formatCalendarCellLine(cell, line, mode)
	style := m.calendarCellStyle(cell, gridFocused, mode)
	return calPadCell(style.Render(text), calColWidth, cellBgColor(cell, gridFocused))
}

func cellBgColor(cell calDayCell, gridFocused bool) lipgloss.Color {
	if cell.day == 0 || skipBackgrounds() {
		return colorCardBg
	}
	switch {
	case cell.selected && gridFocused:
		return colorAqua
	case cell.selected:
		return lipgloss.Color("#3a4a42")
	case cell.today:
		return lipgloss.Color("#2d353b")
	case cell.urgent:
		return colorCalUrgent
	case cell.hasDue:
		return colorCalDue
	default:
		return colorCardBg
	}
}

func calPadCell(styled string, w int, bg lipgloss.Color) string {
	if pad := w - styledTermWidth(styled); pad > 0 {
		styled += lipgloss.NewStyle().Background(bg).Render(strings.Repeat(" ", pad))
	}
	return styled
}

func (m Model) formatCalendarCellLine(cell calDayCell, line int, mode calCellMode) string {
	if cell.day == 0 {
		return padCenterTerm(" ", calColWidth)
	}
	icon := calIconSlot(cell.icon)
	if mode == calCellCompact {
		if line != 0 {
			return padCenterTerm(" ", calColWidth)
		}
		return padCenterTerm(fmt.Sprintf("%2d%s", cell.day, icon), calColWidth)
	}
	if line == 0 {
		return padCenterTerm(fmt.Sprintf("%2d", cell.day), calColWidth)
	}
	return padCenterTerm(icon, calColWidth)
}

func (m Model) formatCalendarCellContent(cell calDayCell, mode calCellMode) string {
	if mode == calCellCompact {
		return m.formatCalendarCellLine(cell, 0, mode)
	}
	return m.formatCalendarCellLine(cell, 0, mode) + "\n" + m.formatCalendarCellLine(cell, 1, mode)
}

func (m Model) calendarCellStyle(cell calDayCell, gridFocused bool, _ calCellMode) lipgloss.Style {
	if cell.day == 0 {
		return StyleCalCellEmpty.Padding(0, 0)
	}
	var s lipgloss.Style
	switch {
	case cell.selected && gridFocused:
		s = StyleCalCellSelected
	case cell.selected:
		s = StyleCalCellSelectedDim
	case cell.today:
		s = StyleCalCellToday
	case cell.urgent:
		s = StyleCalCellUrgent.Foreground(colorWarning)
	case cell.hasDue:
		s = StyleCalCellDue
	default:
		s = StyleCalCellNormal
	}
	return s.Padding(0, 0)
}

func (m Model) renderCalendarSidebar() string {
	_, panelW := m.calendarPanelWidths()
	textW := panelW - sectionHPad

	content := m.renderSelectedDayTasks(textW)

	return StyleSection.
		Width(panelW).
		Render(clampBlock(content, textW))
}

func (m Model) renderSelectedDayTasks(maxW int) string {
	c := m.cat()
	day := m.calSelectedDate()
	tasks := m.tasksForDate(day)

	title := StyleSectionTitle.Render(c.CalendarSelectedDayTitle(day))
	if len(tasks) == 0 {
		return title + "\n" + StyleRowMuted.Render(c.CalendarNoTasks())
	}

	var lines []string
	for i, t := range tasks {
		line := m.renderTaskLine(t, i == m.taskCursor && m.calFocus == calFocusTasks, maxW)
		lines = append(lines, line)
	}
	return title + "\n" + strings.Join(lines, "\n")
}

func (m Model) renderTaskLine(t CalendarTask, focused bool, maxW int) string {
	c := m.cat()
	sep := fillSpaces(1)

	var box string
	if t.Done {
		if focused && !skipBackgrounds() {
			box = lipgloss.NewStyle().Foreground(colorGreen).Background(colorTaskRowFocus).Bold(true).Render("● ")
		} else {
			box = StyleTaskChecked.Render("● ")
		}
	} else if focused {
		box = StyleTaskRowFocused.Render("○ ")
	} else {
		box = StyleTaskUnchecked.Render("○ ")
	}

	label := c.CalendarTaskWater(t.PlantName)
	var labelPart string
	switch {
	case focused && t.Done:
		labelPart = StyleTaskRowFocusedMuted.Render(label)
	case focused:
		labelPart = StyleTaskRowFocused.Render(label)
	case t.Done:
		labelPart = StyleRowMuted.Render(label)
	default:
		labelPart = StyleRowText.Render(label)
	}

	line := box + sep + labelPart
	if t.Urgent && !t.Done {
		urgentStyle := lipgloss.NewStyle().Foreground(colorCritical).Bold(true)
		if !skipBackgrounds() {
			bg := colorCardBg
			if focused {
				bg = colorTaskRowFocus
			}
			urgentStyle = urgentStyle.Background(bg)
		}
		line += sep + urgentStyle.Render("!")
	}

	if focused {
		if w := lipgloss.Width(line); w < maxW {
			line += fillSpaces(maxW - w)
		}
	}
	return truncateWidth(line, maxW)
}

func (m *Model) toggleSelectedTask() tea.Cmd {
	tasks := m.tasksForDate(m.calSelectedDate())
	if len(tasks) == 0 || m.taskCursor < 0 || m.taskCursor >= len(tasks) {
		return nil
	}
	task := tasks[m.taskCursor]
	c := m.cat()

	if task.Done {
		if !task.DoneByLog {
			m.status = c.CalendarTaskAlreadyWatered(task.PlantName)
			return nil
		}
		if err := m.store.UnmarkTaskDone(task.PlantID, task.Date); err != nil {
			m.err = err
			return nil
		}
		m.status = c.CalendarTaskUnchecked(task.PlantName)
		return loadCareLog(m.store)
	}

	if err := m.store.MarkTaskDone(task.PlantID, task.Date, c.LogTaskCompleted(task.PlantName)); err != nil {
		m.err = err
		return nil
	}

	p, err := m.store.GetPlant(task.PlantID)
	if err != nil {
		m.err = err
		return nil
	}
	model.WaterPlant(p)
	if err := m.store.SavePlant(p); err != nil {
		m.err = err
		return nil
	}
	for i := range m.plants {
		if m.plants[i].ID == p.ID {
			m.plants[i] = *p
			break
		}
	}

	m.status = c.CalendarTaskCompleted(task.PlantName)
	return tea.Batch(loadCareLog(m.store), m.reloadPlantCareLog())
}
