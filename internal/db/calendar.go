package db

import (
	"time"
)

func (s *Store) IsTaskDoneOnDate(plantID int64, day time.Time) (bool, error) {
	date := day.Format("2006-01-02")
	var count int
	err := s.db.QueryRow(`
		SELECT COUNT(*) FROM care_log
		WHERE plant_id = ?
		  AND event_type IN ('task_done', 'watered')
		  AND date(created_at) = ?`,
		plantID, date,
	).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (s *Store) MarkTaskDone(plantID int64, day time.Time, message string) error {
	at := time.Date(day.Year(), day.Month(), day.Day(), 12, 0, 0, 0, day.Location())
	return s.addCareLogAt(plantID, "task_done", message, at)
}

func (s *Store) UnmarkTaskDone(plantID int64, day time.Time) error {
	date := day.Format("2006-01-02")
	_, err := s.db.Exec(`
		DELETE FROM care_log
		WHERE plant_id = ?
		  AND event_type = 'task_done'
		  AND date(created_at) = ?`,
		plantID, date,
	)
	return err
}
