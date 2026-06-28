package tui

import (
	"strings"
	"unicode/utf8"

	"kiri/internal/i18n"

	"github.com/charmbracelet/lipgloss"
	"github.com/charmbracelet/x/ansi"
)

const (
	appHPad        = 2 // StyleApp Padding(0, 1)
	frameHBorder   = 2
	frameHPad      = 4 // StyleFrame Padding(1, 2)
	sectionHBorder = 2
	sectionHPad    = 4 // StyleSection Padding(1, 2)
)

func (m Model) termWidth() int {
	if m.width > 0 {
		return m.width
	}
	return 80
}

func (m Model) headerTextWidth() int {
	w := m.termWidth() - appHPad - frameHBorder - frameHPad
	if w < 24 {
		return 24
	}
	return w
}

func (m Model) sectionBoxWidth() int {
	w := m.headerTextWidth() - sectionHBorder
	if w < 20 {
		return 20
	}
	return w
}

func (m Model) innerTextWidth() int {
	w := m.sectionBoxWidth() - sectionHPad
	if w < 16 {
		return 16
	}
	return w
}

func padCenter(s string, width int) string {
	if width <= 0 {
		return ""
	}
	w := lipgloss.Width(s)
	if w >= width {
		return truncateWidth(s, width)
	}
	pad := width - w
	left := pad / 2
	return strings.Repeat(" ", left) + s + strings.Repeat(" ", pad-left)
}

func truncateWidth(s string, max int) string {
	if max <= 0 {
		return ""
	}
	if lipgloss.Width(s) <= max {
		return s
	}
	if max == 1 {
		return "…"
	}
	for len(s) > 0 {
		_, size := utf8.DecodeLastRuneInString(s)
		if size == 0 {
			break
		}
		s = s[:len(s)-size]
		if lipgloss.Width(s+"…") <= max {
			return s + "…"
		}
	}
	return "…"
}

func wrapWidth(s string, maxW int) string {
	if maxW <= 0 {
		return s
	}
	if lipgloss.Width(s) <= maxW {
		return s
	}

	var lines []string
	var line string
	for _, word := range strings.Fields(s) {
		candidate := word
		if line != "" {
			candidate = line + " " + word
		}
		if lipgloss.Width(candidate) <= maxW {
			line = candidate
			continue
		}
		if line != "" {
			lines = append(lines, line)
			line = ""
		}
		if lipgloss.Width(word) <= maxW {
			line = word
			continue
		}
		chunks := breakLongWord(word, maxW)
		lines = append(lines, chunks[:len(chunks)-1]...)
		line = chunks[len(chunks)-1]
	}
	if line != "" {
		lines = append(lines, line)
	}
	return strings.Join(lines, "\n")
}

func breakLongWord(word string, maxW int) []string {
	var lines []string
	var current string
	for _, r := range word {
		ch := string(r)
		next := current + ch
		if lipgloss.Width(next) > maxW && current != "" {
			lines = append(lines, current)
			current = ch
		} else {
			current = next
		}
	}
	if current != "" {
		lines = append(lines, current)
	}
	if len(lines) == 0 {
		return []string{""}
	}
	return lines
}

func wrapBlock(s string, maxW int) string {
	parts := strings.Split(s, "\n")
	out := make([]string, 0, len(parts))
	for _, part := range parts {
		if part == "" {
			out = append(out, "")
			continue
		}
		out = append(out, strings.Split(wrapWidth(part, maxW), "\n")...)
	}
	return strings.Join(out, "\n")
}

func footerLineStyled(text string, rowW int) string {
	if text == "" {
		if skipBackgrounds() {
			return strings.Repeat(" ", rowW)
		}
		return fillStyle().Width(rowW).Render("")
	}
	line := text
	if pad := rowW - lipgloss.Width(line); pad > 0 {
		line += fillSpaces(pad)
	}
	return line
}

func footerHintGroup(h i18n.FooterHint) string {
	return StyleFooterKey.Render(h.Key) + StyleFooterAction.Render(" "+h.Label)
}

type FooterRegion struct {
	X0, X1 int
	Key    string
}

func footerHintLayout(hints []i18n.FooterHint, rowW int) (line string, regions []FooterRegion) {
	if rowW < 1 {
		rowW = 1
	}
	sep := StyleFooterSep.Render(" │ ")
	sepW := lipgloss.Width(sep)

	var cur strings.Builder
	curW := 0

	for _, h := range hints {
		group := footerHintGroup(h)
		gw := lipgloss.Width(group)
		need := gw
		if curW > 0 {
			need += sepW
		}
		if curW > 0 && curW+need > rowW {
			break
		}
		if curW > 0 {
			cur.WriteString(sep)
			curW += sepW
		}
		if curW+gw > rowW {
			group = truncateWidth(group, rowW-curW)
			gw = lipgloss.Width(group)
		}
		if clickKey, ok := footerHintClickKey(h); ok {
			regions = append(regions, FooterRegion{X0: curW, X1: curW + gw, Key: clickKey})
		}
		cur.WriteString(group)
		curW += gw
		if curW >= rowW {
			break
		}
	}

	line = cur.String()
	if lipgloss.Width(line) > rowW {
		line = truncateWidth(line, rowW)
	}
	return footerLineStyled(line, rowW), regions
}

func renderFooterHints(hints []i18n.FooterHint, rowW int) string {
	line, _ := footerHintLayout(hints, rowW)
	return line
}

func styleWrapped(style lipgloss.Style, s string, maxW int) string {
	return styleWrappedMax(style, s, maxW, 0)
}

func styleWrappedMax(style lipgloss.Style, s string, maxW, maxLines int) string {
	wrapped := wrapBlock(s, maxW)
	lines := strings.Split(wrapped, "\n")
	if maxLines > 0 && len(lines) > maxLines {
		lines = lines[:maxLines]
		lines[maxLines-1] = truncateWidth(lines[maxLines-1]+"…", maxW)
	}
	for i, line := range lines {
		if line != "" {
			lines[i] = style.Render(line)
		}
	}
	return strings.Join(lines, "\n")
}

func clampBlock(s string, maxW int) string {
	lines := strings.Split(s, "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = truncateWidth(line, maxW)
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderSection(content string) string {
	return StyleSection.Width(m.sectionBoxWidth()).Render(clampBlock(content, m.innerTextWidth()))
}

const (
	cardBorderV   = 2                         // top + bottom border
	cardBodyMin   = 4                         // name(1) + location(1) + bar(1) + action(1)
	cardBodyMax   = 5                         // name wrapped to 2 lines
	cardHeight    = cardBodyMin + cardBorderV // 6 — nominal/min rendered height
	cardMaxHeight = cardBodyMax + cardBorderV // 7 — worst-case rendered height
)

const maxPlantCols = 3

const cardBorderH = 2

const cardHPad = 2

const uiOverhead = 19

func (m Model) plantGridWidth() int {
	return m.sectionBoxWidth() + cardBorderH
}

func (m Model) plantCardSlotWidths() []int {
	total := m.plantGridWidth()
	base := total / maxPlantCols
	rem := total % maxPlantCols
	slots := make([]int, maxPlantCols)
	for i := range slots {
		slots[i] = base
		if i < rem {
			slots[i]++
		}
	}
	return slots
}

func (m Model) plantCardContentWidth(slot int) int {
	slots := m.plantCardSlotWidths()
	if slot < 0 || slot >= len(slots) {
		slot = 0
	}
	content := slots[slot] - cardBorderH
	if content < 1 {
		content = 1
	}
	return content
}

func (m Model) modalBoxWidth() int {
	w := m.termWidth() - 8
	if w > 54 {
		w = 54
	}
	if w < 36 {
		w = 36
	}
	return w
}

func (m Model) modalTextWidth() int {
	w := m.modalBoxWidth() - 6
	if w < 28 {
		w = 28
	}
	return w
}

func backdropLine(width int) string {
	return StyleBackdrop.Render(strings.Repeat(" ", width))
}

func overlayCompose(base, popup string, width, height int) string {
	if width < 1 {
		width = 80
	}
	if height < 1 {
		height = 24
	}

	canvas := lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, base)
	rows := strings.Split(canvas, "\n")
	for len(rows) < height {
		rows = append(rows, strings.Repeat(" ", width))
	}
	rows = rows[:height]

	popupRows := strings.Split(popup, "\n")
	popupW := 0
	for _, line := range popupRows {
		if w := ansi.StringWidth(line); w > popupW {
			popupW = w
		}
	}
	popupH := len(popupRows)
	startY := (height - popupH) / 2
	startX := (width - popupW) / 2
	if startX < 0 {
		startX = 0
	}

	for i, line := range popupRows {
		y := startY + i
		if y < 0 || y >= height {
			break
		}
		rows[y] = overlayLine(rows[y], line, startX)
	}

	return strings.Join(rows, "\n")
}

func overlayAt(base, popup string, startX, startY, width, height int) string {
	if width < 1 {
		width = 80
	}
	if height < 1 {
		height = 24
	}
	if popup == "" {
		return base
	}

	canvas := lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, base)
	rows := strings.Split(canvas, "\n")
	for len(rows) < height {
		rows = append(rows, strings.Repeat(" ", width))
	}
	rows = rows[:height]

	if startX < 0 {
		startX = 0
	}

	for i, line := range strings.Split(popup, "\n") {
		y := startY + i
		if y < 0 || y >= height {
			break
		}
		rows[y] = overlayLine(rows[y], line, startX)
	}

	return strings.Join(rows, "\n")
}

func overlayCenter(base, popup string, width, height int) string {
	if width < 1 {
		width = 80
	}
	if height < 1 {
		height = 24
	}

	var canvas string
	if skipBackgrounds() {
		canvas = lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top, base)
	} else {
		canvas = lipgloss.Place(width, height, lipgloss.Left, lipgloss.Top,
			StyleBackdrop.Render(base),
			lipgloss.WithWhitespaceBackground(colorBackdrop),
		)
	}

	rows := strings.Split(canvas, "\n")
	for len(rows) < height {
		rows = append(rows, backdropLine(width))
	}
	rows = rows[:height]

	popupRows := strings.Split(popup, "\n")
	popupW := 0
	for _, line := range popupRows {
		if w := ansi.StringWidth(line); w > popupW {
			popupW = w
		}
	}
	popupH := len(popupRows)
	startY := (height - popupH) / 2
	startX := (width - popupW) / 2
	if startX < 0 {
		startX = 0
	}

	for i, line := range popupRows {
		y := startY + i
		if y < 0 || y >= height {
			break
		}
		rows[y] = overlayLine(rows[y], line, startX)
	}

	return strings.Join(rows, "\n")
}

func overlayLine(background, foreground string, x int) string {
	bgW := ansi.StringWidth(background)
	if x >= bgW {
		return background
	}
	left := ansi.Cut(background, 0, x)
	fgW := ansi.StringWidth(foreground)
	rightStart := x + fgW
	var right string
	if rightStart < bgW {
		right = ansi.Cut(background, rightStart, bgW)
	}
	return left + foreground + right
}
