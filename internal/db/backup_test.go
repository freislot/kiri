package db

import (
	"database/sql"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"kiri/internal/i18n"

	_ "github.com/mattn/go-sqlite3"
)

func TestBackupTo(t *testing.T) {
	store := openTestStore(t)
	backupPath := filepath.Join(t.TempDir(), "backup.db")

	if err := store.BackupTo(backupPath); err != nil {
		t.Fatalf("backup: %v", err)
	}

	db, err := sql.Open("sqlite3", backupPath)
	if err != nil {
		t.Fatalf("open backup: %v", err)
	}
	defer db.Close()

	var count int
	if err := db.QueryRow(`SELECT COUNT(*) FROM plants`).Scan(&count); err != nil {
		t.Fatalf("query backup plants: %v", err)
	}
	if count == 0 {
		t.Fatal("backup should contain seeded plants")
	}

	settingsBackup := SettingsBackupPathFor(backupPath)
	data, err := os.ReadFile(settingsBackup)
	if err != nil {
		t.Fatalf("read settings backup: %v", err)
	}
	if !strings.Contains(string(data), "language") {
		t.Fatalf("settings backup = %q, want language setting", string(data))
	}

	if err := store.BackupTo(backupPath); err != nil {
		t.Fatalf("second backup: %v", err)
	}
}

func TestBackupTo_IncludesUpdatedSettings(t *testing.T) {
	store := openTestStore(t)
	if err := store.SetLanguage(i18n.EN); err != nil {
		t.Fatalf("set language: %v", err)
	}

	backupPath := filepath.Join(t.TempDir(), "backup.db")
	if err := store.BackupTo(backupPath); err != nil {
		t.Fatalf("backup: %v", err)
	}

	data, err := os.ReadFile(SettingsBackupPathFor(backupPath))
	if err != nil {
		t.Fatalf("read settings backup: %v", err)
	}
	if !strings.Contains(string(data), "language = 'en'") {
		t.Fatalf("settings backup = %q, want language = 'en'", string(data))
	}
}

func TestDefaultBackupPath(t *testing.T) {
	path, err := DefaultBackupPath()
	if err != nil {
		t.Fatalf("DefaultBackupPath: %v", err)
	}
	if path == "" || filepath.Ext(path) != ".bak" {
		t.Fatalf("unexpected backup path: %q", path)
	}
}

func TestDefaultSettingsBackupPath(t *testing.T) {
	path, err := DefaultSettingsBackupPath()
	if err != nil {
		t.Fatalf("DefaultSettingsBackupPath: %v", err)
	}
	if path == "" || filepath.Base(path) != "settings.toml.bak" {
		t.Fatalf("unexpected settings backup path: %q", path)
	}
}
