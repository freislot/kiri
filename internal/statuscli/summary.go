package statuscli

import (
	"fmt"
	"io"
	"time"

	"kiri/internal/db"
)

func PrintSummary(w io.Writer, store *db.Store) error {
	plants, modelCfg, c, err := loadPlants(store)
	if err != nil {
		return err
	}

	urgent := countNeedsWatering(plants, time.Now(), modelCfg)
	line := c.CLISummaryOK()
	if urgent > 0 {
		line = c.CLISummaryNeedsWater(urgent)
	}
	_, err = fmt.Fprintln(w, line)
	return err
}
