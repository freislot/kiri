package db

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func DefaultBackupPath() (string, error) {
	dbPath, err := DefaultDBPath()
	if err != nil {
		return "", err
	}
	return dbPath + ".bak", nil
}

func DefaultSettingsBackupPath() (string, error) {
	dbPath, err := DefaultDBPath()
	if err != nil {
		return "", err
	}
	return filepath.Join(filepath.Dir(dbPath), "settings.toml.bak"), nil
}

func SettingsBackupPathFor(dbBackupPath string) string {
	return filepath.Join(filepath.Dir(dbBackupPath), "settings.toml.bak")
}

func (s *Store) BackupTo(destPath string) error {
	if err := s.backupDatabaseTo(destPath); err != nil {
		return err
	}
	return s.backupSettingsTo(SettingsBackupPathFor(destPath))
}

func (s *Store) backupDatabaseTo(destPath string) error {
	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	if err := os.Remove(destPath); err != nil && !os.IsNotExist(err) {
		return err
	}
	abs, err := filepath.Abs(destPath)
	if err != nil {
		return err
	}
	escaped := strings.ReplaceAll(abs, "'", "''")
	_, err = s.db.Exec(fmt.Sprintf("VACUUM INTO '%s'", escaped))
	return err
}

func (s *Store) backupSettingsTo(destPath string) error {
	s.settingsMu.Lock()
	defer s.settingsMu.Unlock()

	src := s.settingsFilePath()
	if _, err := os.Stat(src); os.IsNotExist(err) {
		cfg, loadErr := s.loadSettingsLocked()
		if loadErr != nil && !os.IsNotExist(loadErr) {
			return loadErr
		}
		if writeErr := s.writeSettingsFile(cfg); writeErr != nil {
			return writeErr
		}
	}

	if err := os.MkdirAll(filepath.Dir(destPath), 0o755); err != nil {
		return err
	}
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.OpenFile(destPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	return err
}
