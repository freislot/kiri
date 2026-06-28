package statuscli

import (
	"fmt"
	"io"
	"strings"
	"time"

	"kiri/internal/db"
	"kiri/internal/i18n"
	"kiri/internal/model"

	"github.com/mattn/go-runewidth"
)

const (
	barWidth     = 10
	colGap       = 2
	alreadyLevel = 80
)

func Print(w io.Writer, store *db.Store) error {
	plants, modelCfg, c, err := loadPlants(store)
	if err != nil {
		return err
	}

	now := time.Now()
	rows := make([][]string, 0, len(plants))
	for _, p := range plants {
		rows = append(rows, []string{
			c.CLIPlantName(p.Name, p.IsOutdoor),
			moistureBar(p.WaterLevel),
			wateringLabel(c, p, now, modelCfg),
		})
	}

	if len(rows) == 0 {
		_, err := fmt.Fprintln(w, c.CLINoPlants())
		return err
	}

	header := []string{c.CLIHeaderPlant(), c.CLIHeaderStatus(), c.CLIHeaderWatering()}
	widths := columnWidths(append([][]string{header}, rows...))
	var b strings.Builder
	b.WriteString(formatRow(header, widths))
	b.WriteByte('\n')
	for _, row := range rows {
		b.WriteString(formatRow(row, widths))
		b.WriteByte('\n')
	}
	_, err = io.WriteString(w, b.String())
	return err
}

func wateringLabel(c i18n.Catalog, p model.Plant, now time.Time, cfg model.ModelConfig) string {
	if p.State == model.StateWaterNow || p.WaterLevel <= 0 {
		return c.CLIWateringUrgent()
	}
	days := daysUntil(model.WateringDueDateWithConfig(p, now, cfg), now)
	if days <= 0 {
		return c.CLIWateringUrgent()
	}
	if p.WaterLevel >= alreadyLevel {
		return c.CLIWateringAlready()
	}
	return c.CLIWateringInDays(days)
}

func daysUntil(due, now time.Time) int {
	a := dateOnly(now)
	b := dateOnly(due)
	return int(b.Sub(a).Hours() / 24)
}

func dateOnly(t time.Time) time.Time {
	y, m, d := t.Date()
	return time.Date(y, m, d, 0, 0, 0, 0, t.Location())
}

func moistureBar(level float64) string {
	filled := int(level/100*float64(barWidth) + 0.5)
	if filled > barWidth {
		filled = barWidth
	}
	if filled < 0 {
		filled = 0
	}
	out := make([]rune, barWidth+2)
	out[0] = '['
	out[barWidth+1] = ']'
	for i := 0; i < barWidth; i++ {
		if i < filled {
			out[i+1] = '█'
		} else {
			out[i+1] = '░'
		}
	}
	return string(out)
}

func columnWidths(rows [][]string) [3]int {
	var widths [3]int
	for _, row := range rows {
		for i, cell := range row {
			if w := runewidth.StringWidth(cell); w > widths[i] {
				widths[i] = w
			}
		}
	}
	return widths
}

func formatRow(cells []string, widths [3]int) string {
	parts := make([]string, len(cells))
	for i, cell := range cells {
		parts[i] = padRight(cell, widths[i])
	}
	return strings.Join(parts, strings.Repeat(" ", colGap))
}

func padRight(s string, width int) string {
	if runewidth.StringWidth(s) >= width {
		return runewidth.Truncate(s, width, "")
	}
	return s + strings.Repeat(" ", width-runewidth.StringWidth(s))
}
