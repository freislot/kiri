package db

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"kiri/internal/model"

	_ "github.com/mattn/go-sqlite3"
)

type Store struct {
	db         *sql.DB
	dbPath     string
	settingsMu sync.Mutex
}

func DefaultDBPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config", "kiri", "data.db"), nil
}

func Open(path string) (*Store, error) {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return nil, err
	}
	db, err := sql.Open("sqlite3", path)
	if err != nil {
		return nil, err
	}
	if err := db.Ping(); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA journal_mode=WAL;`); err != nil {
		return nil, err
	}
	if _, err := db.Exec(`PRAGMA synchronous=NORMAL;`); err != nil {
		return nil, err
	}
	s := &Store{db: db, dbPath: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) Close() error {
	return s.db.Close()
}

func (s *Store) DBPath() string {
	return s.dbPath
}

func (s *Store) migrate() error {
	schema := `
CREATE TABLE IF NOT EXISTS schema_version (
	version INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS plants (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	location TEXT NOT NULL,
	base_interval_days INTEGER NOT NULL DEFAULT 7,
	is_outdoor INTEGER NOT NULL DEFAULT 0,
	water_level REAL NOT NULL DEFAULT 100.0,
	consecutive_postpones INTEGER NOT NULL DEFAULT 0,
	state TEXT NOT NULL DEFAULT 'normal',
	next_action TEXT NOT NULL DEFAULT '',
	last_updated_at TEXT NOT NULL,
	created_at TEXT NOT NULL
);

CREATE TABLE IF NOT EXISTS care_log (
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	plant_id INTEGER NOT NULL,
	event_type TEXT NOT NULL,
	message TEXT NOT NULL,
	created_at TEXT NOT NULL,
	FOREIGN KEY (plant_id) REFERENCES plants(id)
);
`
	if _, err := s.db.Exec(schema); err != nil {
		return fmt.Errorf("migrate: %w", err)
	}

	if err := s.migrateLegacySettings(); err != nil {
		return fmt.Errorf("migrate legacy settings: %w", err)
	}

	var count int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM plants`).Scan(&count); err != nil {
		return err
	}
	if count == 0 {
		return seed(s)
	}
	return nil
}

func scanPlant(row interface {
	Scan(dest ...any) error
}) (*model.Plant, error) {
	var p model.Plant
	var isOutdoor int
	var lastUpdated, created string
	if err := row.Scan(
		&p.ID, &p.Name, &p.Location, &p.BaseIntervalDays, &isOutdoor,
		&p.WaterLevel, &p.ConsecutivePostpones, &p.State, &p.NextAction,
		&lastUpdated, &created,
	); err != nil {
		return nil, err
	}
	p.IsOutdoor = isOutdoor == 1
	var err error
	p.LastUpdatedAt, err = time.Parse(time.RFC3339, lastUpdated)
	if err != nil {
		p.LastUpdatedAt = time.Now()
	}
	p.CreatedAt, err = time.Parse(time.RFC3339, created)
	if err != nil {
		p.CreatedAt = time.Now()
	}
	return &p, nil
}

const plantColumns = `id, name, location, base_interval_days, is_outdoor,
	water_level, consecutive_postpones, state, next_action, last_updated_at, created_at`

func (s *Store) ListPlants() ([]model.Plant, error) {
	rows, err := s.db.Query(`SELECT ` + plantColumns + ` FROM plants ORDER BY id DESC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var plants []model.Plant
	for rows.Next() {
		p, err := scanPlant(rows)
		if err != nil {
			return nil, err
		}
		plants = append(plants, *p)
	}
	return plants, rows.Err()
}

func (s *Store) GetPlant(id int64) (*model.Plant, error) {
	row := s.db.QueryRow(`SELECT `+plantColumns+` FROM plants WHERE id = ?`, id)
	return scanPlant(row)
}

func (s *Store) SavePlant(p *model.Plant) error {
	if p.LastUpdatedAt.IsZero() {
		p.LastUpdatedAt = time.Now()
	}
	if p.CreatedAt.IsZero() {
		p.CreatedAt = time.Now()
	}
	isOutdoor := 0
	if p.IsOutdoor {
		isOutdoor = 1
	}
	if p.ID == 0 {
		res, err := s.db.Exec(`
			INSERT INTO plants (name, location, base_interval_days, is_outdoor, water_level,
				consecutive_postpones, state, next_action, last_updated_at, created_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
			p.Name, p.Location, p.BaseIntervalDays, isOutdoor, p.WaterLevel,
			p.ConsecutivePostpones, p.State, p.NextAction,
			p.LastUpdatedAt.Format(time.RFC3339), p.CreatedAt.Format(time.RFC3339),
		)
		if err != nil {
			return err
		}
		id, err := res.LastInsertId()
		if err != nil {
			return err
		}
		p.ID = id
		return nil
	}
	_, err := s.db.Exec(`
		UPDATE plants SET name=?, location=?, base_interval_days=?, is_outdoor=?,
			water_level=?, consecutive_postpones=?, state=?, next_action=?, last_updated_at=?
		WHERE id=?`,
		p.Name, p.Location, p.BaseIntervalDays, isOutdoor, p.WaterLevel,
		p.ConsecutivePostpones, p.State, p.NextAction,
		p.LastUpdatedAt.Format(time.RFC3339), p.ID,
	)
	return err
}

func (s *Store) DeletePlant(id int64) error {
	if _, err := s.db.Exec(`DELETE FROM care_log WHERE plant_id = ?`, id); err != nil {
		return err
	}
	_, err := s.db.Exec(`DELETE FROM plants WHERE id = ?`, id)
	return err
}

func (s *Store) AddCareLog(plantID int64, eventType, message string) error {
	return s.addCareLogAt(plantID, eventType, message, time.Now())
}

func (s *Store) addCareLogAt(plantID int64, eventType, message string, at time.Time) error {
	_, err := s.db.Exec(`
		INSERT INTO care_log (plant_id, event_type, message, created_at)
		VALUES (?, ?, ?, ?)`,
		plantID, eventType, message, at.Format(time.RFC3339),
	)
	return err
}

func (s *Store) CareLogForPlant(plantID int64, limit int) ([]model.CareLogEntry, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.plant_id, p.name, c.event_type, c.message, c.created_at
		FROM care_log c
		JOIN plants p ON p.id = c.plant_id
		WHERE c.plant_id = ?
		ORDER BY c.created_at DESC
		LIMIT ?`, plantID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.CareLogEntry
	for rows.Next() {
		var e model.CareLogEntry
		var created string
		if err := rows.Scan(&e.ID, &e.PlantID, &e.PlantName, &e.EventType, &e.Message, &created); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, created)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}

func (s *Store) AllCareLog(limit int) ([]model.CareLogEntry, error) {
	rows, err := s.db.Query(`
		SELECT c.id, c.plant_id, p.name, c.event_type, c.message, c.created_at
		FROM care_log c
		JOIN plants p ON p.id = c.plant_id
		ORDER BY c.created_at DESC
		LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []model.CareLogEntry
	for rows.Next() {
		var e model.CareLogEntry
		var created string
		if err := rows.Scan(&e.ID, &e.PlantID, &e.PlantName, &e.EventType, &e.Message, &created); err != nil {
			return nil, err
		}
		e.CreatedAt, _ = time.Parse(time.RFC3339, created)
		entries = append(entries, e)
	}
	return entries, rows.Err()
}
