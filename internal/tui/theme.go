package tui

import (
	"strings"

	"github.com/charmbracelet/lipgloss"
)

var transparentMode bool

func init() {
	initStyles()
}

func ApplyTransparentMode(on bool) {
	transparentMode = on
	initStyles()
}

func fillStyle() lipgloss.Style {
	if skipBackgrounds() {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Background(colorCardBg)
}

func fillSpaces(n int) string {
	if n <= 0 {
		return ""
	}
	return fillStyle().Render(strings.Repeat(" ", n))
}

func styleBg(s lipgloss.Style, c lipgloss.Color) lipgloss.Style {
	if skipBackgrounds() {
		return s
	}
	return s.Background(c)
}

func styleBorderBg(s lipgloss.Style, c lipgloss.Color) lipgloss.Style {
	if skipBackgrounds() {
		return s
	}
	return s.BorderBackground(c)
}

func initStyles() {
	StyleApp = styleBg(lipgloss.NewStyle().Foreground(colorFg).Padding(0, 1), colorBg)

	StyleFrame = styleBorderBg(
		styleBg(
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(0, 2),
			colorCardBg,
		),
		colorCardBg,
	)

	StyleWeather = styleBg(lipgloss.NewStyle().Foreground(colorFg), colorCardBg)
	StyleTabActive = styleBg(lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Underline(true), colorCardBg)
	StyleTabInactive = styleBg(lipgloss.NewStyle().Foreground(colorMuted), colorCardBg)
	StyleToggleActive = styleBg(lipgloss.NewStyle().Foreground(colorGreen).Bold(true), colorCardBg)
	StyleToggleInactive = styleBg(lipgloss.NewStyle().Foreground(colorMuted), colorCardBg)

	StyleSection = styleBorderBg(
		styleBg(
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorder).
				Padding(1, 2),
			colorCardBg,
		),
		colorCardBg,
	)

	StyleModal = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(colorGreen).
		BorderBackground(colorCardBg).
		Background(colorCardBg).
		Padding(1, 2)

	StyleModalBackup = lipgloss.NewStyle().
		Border(lipgloss.RoundedBorder()).
		BorderForeground(gradTop, gradRight, gradBottom, gradLeft).
		BorderBackground(colorCardBg).
		Background(colorCardBg).
		Padding(1, 2)

	if skipBackgrounds() {
		StyleBackdrop = lipgloss.NewStyle().Foreground(colorMuted)
	} else {
		StyleBackdrop = lipgloss.NewStyle().Foreground(colorMuted).Background(colorBackdrop)
	}

	StyleSectionTitle = styleBg(lipgloss.NewStyle().Foreground(colorGreen).Bold(true), colorCardBg)
	StyleFooterKey = styleBg(lipgloss.NewStyle().Foreground(colorGreen).Bold(true), colorCardBg)
	StyleFooterAction = styleBg(lipgloss.NewStyle().Foreground(colorFgDim), colorCardBg)
	StyleFooterSep = styleBg(lipgloss.NewStyle().Foreground(colorBorderFaint), colorCardBg)
	StyleRowNameSelected = styleBg(lipgloss.NewStyle().Foreground(colorAqua).Bold(true), colorCardBg)
	StyleRowText = styleBg(lipgloss.NewStyle().Foreground(colorFg), colorCardBg)
	StyleRowMuted = styleBg(lipgloss.NewStyle().Foreground(colorMuted), colorCardBg)
	StyleTaskUnchecked = styleBg(lipgloss.NewStyle().Foreground(colorMuted), colorCardBg)
	StyleTaskChecked = styleBg(lipgloss.NewStyle().Foreground(colorGreen).Bold(true), colorCardBg)

	if skipBackgrounds() {
		StyleTaskRowFocused = lipgloss.NewStyle().Foreground(colorFg).Bold(true).Underline(true)
		StyleTaskRowFocusedMuted = lipgloss.NewStyle().Foreground(colorMuted).Underline(true)
	} else {
		StyleTaskRowFocused = lipgloss.NewStyle().Foreground(colorFg).Background(colorTaskRowFocus)
		StyleTaskRowFocusedMuted = lipgloss.NewStyle().Foreground(colorMuted).Background(colorTaskRowFocus)
	}

	StyleStatusUrgent = styleBg(lipgloss.NewStyle().Foreground(colorCritical).Bold(true), colorCardBg)
	StyleDetailBullet = styleBg(lipgloss.NewStyle().Foreground(colorFg), colorCardBg)
	StyleStatusMsg = styleBg(lipgloss.NewStyle().Foreground(colorWarning).Italic(true), colorCardBg)
	StyleScrollIndicator = styleBg(lipgloss.NewStyle().Foreground(colorMuted).Italic(true), colorCardBg)
	StyleScrollBar = styleBg(lipgloss.NewStyle().Foreground(colorGreen), colorCardBg)

	StyleCardFocused = styleBorderBg(
		styleBg(
			lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(gradTop, gradRight, gradBottom, gradLeft).
				Padding(0, 1),
			colorCardBg,
		),
		colorCardBg,
	)

	StyleCardUrgentFocused = styleBorderBg(
		styleBg(
			lipgloss.NewStyle().
				Border(lipgloss.ThickBorder()).
				BorderForeground(colorCritical).
				Padding(0, 1),
			colorCardBg,
		),
		colorCardBg,
	)

	StyleCardInactive = styleBorderBg(
		styleBg(
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorBorderGhost).
				Padding(0, 1),
			colorCardBgDim,
		),
		colorCardBgDim,
	)

	StyleCardUrgentInactive = styleBorderBg(
		styleBg(
			lipgloss.NewStyle().
				Border(lipgloss.RoundedBorder()).
				BorderForeground(colorCriticalDim).
				Padding(0, 1),
			colorCardBgDim,
		),
		colorCardBgDim,
	)

	StyleCalWeekday = styleBg(
		lipgloss.NewStyle().Foreground(colorMuted).Bold(true).Align(lipgloss.Center).Padding(0, 1),
		colorCardBg,
	)
	StyleCalCellNormal = styleBg(
		lipgloss.NewStyle().Foreground(colorFg).Align(lipgloss.Center).Padding(0, 1),
		colorCardBg,
	)
	StyleCalCellEmpty = styleBg(
		lipgloss.NewStyle().Align(lipgloss.Center).Padding(0, 1),
		colorCardBg,
	)

	if skipBackgrounds() {
		StyleCalCellDue = lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellUrgent = lipgloss.NewStyle().Foreground(colorCritical).Bold(true).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellToday = lipgloss.NewStyle().Foreground(colorGreen).Bold(true).Underline(true).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellSelected = lipgloss.NewStyle().Foreground(colorAqua).Bold(true).Reverse(true).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellSelectedDim = lipgloss.NewStyle().Foreground(colorFg).Bold(true).Underline(true).Align(lipgloss.Center).Padding(0, 1)
	} else {
		StyleCalCellDue = lipgloss.NewStyle().Background(colorCalDue).Foreground(colorGreen).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellUrgent = lipgloss.NewStyle().Background(colorCalUrgent).Foreground(colorCritical).Bold(true).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellToday = lipgloss.NewStyle().Background(lipgloss.Color("#2d353b")).Foreground(colorGreen).Bold(true).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellSelected = lipgloss.NewStyle().Background(colorAqua).Foreground(colorBg).Bold(true).Align(lipgloss.Center).Padding(0, 1)
		StyleCalCellSelectedDim = lipgloss.NewStyle().Background(lipgloss.Color("#3a4a42")).Foreground(colorFg).Bold(true).Align(lipgloss.Center).Padding(0, 1)
	}
}
