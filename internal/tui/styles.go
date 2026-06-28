package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

var (
	colorBg           = lipgloss.Color("#272e33") // terminal background
	colorCardBg       = lipgloss.Color("#272e33") // card / panel background
	colorFg           = lipgloss.Color("#d3c6aa") // primary text
	colorMuted        = lipgloss.Color("#7c8377") // muted labels
	colorGreen        = lipgloss.Color("#a7c080") // good moisture / accent
	colorAqua         = lipgloss.Color("#83c092") // rain / auto water
	colorWarning      = lipgloss.Color("#dbbc7f") // water soon
	colorCritical     = lipgloss.Color("#e67e80") // urgent
	colorBorder       = lipgloss.Color("#414b50") // inactive borders
	colorBorderFaint  = lipgloss.Color("#353d42") // subtle borders / grid lines
	colorBorderGhost  = lipgloss.Color("#414b50") // unfocused card borders
	colorCardBgDim    = lipgloss.Color("#272e33") // unfocused card background
	colorFgDim        = lipgloss.Color("#5a6158") // unfocused primary text
	colorGreenDim     = lipgloss.Color("#6d7a5a") // unfocused accent
	colorAquaDim      = lipgloss.Color("#5a7a6e") // unfocused rain accent
	colorWarningDim   = lipgloss.Color("#8a7a56") // unfocused warning
	colorCriticalDim  = lipgloss.Color("#9a6365") // unfocused urgent
	colorBackdrop     = lipgloss.Color("#181c1f") // overlay dim
	colorTaskRowFocus = lipgloss.Color("#343f44") // focused calendar task row
	colorCalDue       = lipgloss.Color("#2d3a30") // calendar day with watering due
	colorCalUrgent    = lipgloss.Color("#3a2d2d") // calendar day with urgent watering

	gradTop    = lipgloss.Color("#a7c080")
	gradRight  = lipgloss.Color("#95b172")
	gradBottom = lipgloss.Color("#83c092")
	gradLeft   = lipgloss.Color("#8fa18a")

	StyleApp                 lipgloss.Style
	StyleFrame               lipgloss.Style
	StyleWeather             lipgloss.Style
	StyleTabActive           lipgloss.Style
	StyleTabInactive         lipgloss.Style
	StyleToggleActive        lipgloss.Style
	StyleToggleInactive      lipgloss.Style
	StyleSection             lipgloss.Style
	StyleModal               lipgloss.Style
	StyleModalBackup         lipgloss.Style
	StyleBackdrop            lipgloss.Style
	StyleSectionTitle        lipgloss.Style
	StyleFooterKey           lipgloss.Style
	StyleFooterAction        lipgloss.Style
	StyleFooterSep           lipgloss.Style
	StyleRowNameSelected     lipgloss.Style
	StyleRowText             lipgloss.Style
	StyleRowMuted            lipgloss.Style
	StyleTaskUnchecked       lipgloss.Style
	StyleTaskChecked         lipgloss.Style
	StyleTaskRowFocused      lipgloss.Style
	StyleTaskRowFocusedMuted lipgloss.Style
	StyleStatusUrgent        lipgloss.Style
	StyleDetailBullet        lipgloss.Style
	StyleStatusMsg           lipgloss.Style
	StyleScrollIndicator     lipgloss.Style
	StyleScrollBar           lipgloss.Style
	StyleCardFocused         lipgloss.Style
	StyleCardUrgentFocused   lipgloss.Style
	StyleCardInactive        lipgloss.Style
	StyleCardUrgentInactive  lipgloss.Style
	StyleCalWeekday          lipgloss.Style
	StyleCalCellNormal       lipgloss.Style
	StyleCalCellDue          lipgloss.Style
	StyleCalCellUrgent       lipgloss.Style
	StyleCalCellToday        lipgloss.Style
	StyleCalCellSelected     lipgloss.Style
	StyleCalCellSelectedDim  lipgloss.Style
	StyleCalCellEmpty        lipgloss.Style
)

func moistureColor(level float64) lipgloss.Color {
	switch {
	case level >= 80:
		return colorGreen
	case level >= 30:
		return colorWarning
	default:
		return colorCritical
	}
}

func cardBg(focused bool) lipgloss.Color {
	if focused {
		return colorCardBg
	}
	return colorCardBgDim
}

func cardTextStyle(fg lipgloss.Color, bg lipgloss.Color, bold bool) lipgloss.Style {
	s := lipgloss.NewStyle().Foreground(fg)
	if !skipBackgrounds() {
		s = s.Background(bg)
	}
	if bold {
		s = s.Bold(true)
	}
	return s
}

func cardMoistureColor(level float64, focused bool) lipgloss.Color {
	if focused {
		return moistureColor(level)
	}
	switch {
	case level >= 80:
		return colorGreenDim
	case level >= 30:
		return colorWarningDim
	default:
		return colorCriticalDim
	}
}

func cardBorderStyle(focused, urgent bool) lipgloss.Style {
	switch {
	case focused && urgent:
		return StyleCardUrgentFocused
	case focused:
		return StyleCardFocused
	case urgent:
		return StyleCardUrgentInactive
	default:
		return StyleCardInactive
	}
}

func cardNameStyle(focused bool) lipgloss.Style {
	bg := cardBg(focused)
	if focused {
		return cardTextStyle(colorGreen, bg, true)
	}
	return cardTextStyle(colorFgDim, bg, false)
}

func cardLocationStyle(focused bool) lipgloss.Style {
	bg := cardBg(focused)
	if focused {
		return cardTextStyle(colorMuted, bg, false)
	}
	return cardTextStyle(colorFgDim, bg, false).Faint(true)
}

func cardActionStyle(focused bool) lipgloss.Style {
	bg := cardBg(focused)
	if focused {
		return cardTextStyle(colorFg, bg, false)
	}
	return cardTextStyle(colorFgDim, bg, false)
}

func cardActionUrgentStyle(focused bool) lipgloss.Style {
	bg := cardBg(focused)
	if focused {
		return cardTextStyle(colorCritical, bg, true)
	}
	return cardTextStyle(colorCriticalDim, bg, false)
}

func cardActionRainStyle(focused bool) lipgloss.Style {
	bg := cardBg(focused)
	if focused {
		return cardTextStyle(colorAqua, bg, false)
	}
	return cardTextStyle(colorAquaDim, bg, false)
}

func styledBarOnBg(level float64, width int, bg lipgloss.Color, focused bool) string {
	bar := progressBar(level, width)
	s := lipgloss.NewStyle().Foreground(cardMoistureColor(level, focused))
	if !skipBackgrounds() {
		s = s.Background(bg)
	}
	return s.Render(bar)
}

func styledPercentOnBg(level float64, bg lipgloss.Color, focused bool) string {
	s := lipgloss.NewStyle().Foreground(cardMoistureColor(level, focused))
	if !skipBackgrounds() {
		s = s.Background(bg)
	}
	return s.Render(fmtPercent(level))
}

func fmtPercent(level float64) string {
	pct := int(level + 0.5)
	if pct < 0 {
		pct = 0
	}
	if pct > 100 {
		pct = 100
	}
	return fmt.Sprintf("%d%%", pct)
}

func progressBar(level float64, width int) string {
	filled := int(level / 100 * float64(width))
	if filled > width {
		filled = width
	}
	if filled < 0 {
		filled = 0
	}
	out := make([]rune, width)
	for i := 0; i < width; i++ {
		if i < filled {
			out[i] = '▰'
		} else {
			out[i] = '▱'
		}
	}
	return string(out)
}
