package tui

import (
	"io"
	"os"

	"github.com/muesli/termenv"
)

var (
	terminalTrueColor = true
	colorProfileFor   = func(w io.Writer) termenv.Profile {
		return termenv.NewOutput(w).EnvColorProfile()
	}
)

func InitTerminalDisplay(w io.Writer) {
	if w == nil {
		w = os.Stdout
	}
	terminalTrueColor = colorProfileFor(w) == termenv.TrueColor
	initStyles()
}

func skipBackgrounds() bool {
	return transparentMode || !terminalTrueColor
}

func TerminalSupportsTrueColor() bool {
	return terminalTrueColor
}
