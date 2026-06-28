package tui

import (
	"path/filepath"
	"strings"
	"testing"

	"kiri/internal/db"
	"kiri/internal/i18n"

	tea "github.com/charmbracelet/bubbletea"
)

func TestHandleSettingsKey_SpaceTogglesAutoBackup(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	m := Model{
		store:       store,
		tab:         TabSettings,
		lang:        i18n.EN,
		cursor:      settingsRowAutoBackup,
		autoBackup:  false,
		splashPhase: splashPhaseDone,
	}

	next, cmd := m.handleSettingsKey(tea.KeyMsg{Type: tea.KeySpace})
	if cmd != nil {
		t.Fatalf("unexpected cmd: %v", cmd)
	}
	got := next.(Model)
	if !got.autoBackup {
		t.Fatal("space should enable auto-backup")
	}
}

func TestHandleSettingsKey_SpaceStringMatchesKeySpace(t *testing.T) {
	if (tea.KeyMsg{Type: tea.KeySpace}).String() == "space" {
		t.Fatal("bubbletea KeySpace.String() is not \"space\"; use \" \" in switches")
	}
}

func TestHandleSettingsKey_ShiftDefaultInterval(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	m := Model{
		store:               store,
		tab:                 TabSettings,
		lang:                i18n.EN,
		cursor:              settingsRowDefaultInterval,
		defaultIntervalDays: 7,
		splashPhase:         splashPhaseDone,
	}

	next, cmd := m.handleSettingsKey(tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'l'}})
	if cmd != nil {
		t.Fatalf("unexpected cmd: %v", cmd)
	}
	got := next.(Model)
	if got.defaultIntervalDays != 8 {
		t.Fatalf("default interval = %d, want 8", got.defaultIntervalDays)
	}
	cfg, err := store.LoadSettings()
	if err != nil {
		t.Fatalf("load settings: %v", err)
	}
	if cfg.DefaultIntervalDays != 8 {
		t.Fatalf("stored default interval = %d, want 8", cfg.DefaultIntervalDays)
	}
}

func TestRenderSettingsTab_ShowsDatabaseAndBackupPaths(t *testing.T) {
	dbPath := filepath.Join(t.TempDir(), "data.db")
	store, err := db.Open(dbPath)
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	m := Model{
		store:  store,
		lang:   i18n.RU,
		width:  120,
		height: 40,
	}

	view := m.renderSettingsTab()
	if !strings.Contains(view, "База данных") {
		t.Fatalf("settings tab = %q, want database label", view)
	}
	if !strings.Contains(view, dbPath) {
		t.Fatalf("settings tab = %q, want database path %q", view, dbPath)
	}
	if !strings.Contains(view, "Бэкап") {
		t.Fatalf("settings tab = %q, want backup label", view)
	}
	if !strings.Contains(view, dbPath+".bak") {
		t.Fatalf("settings tab = %q, want backup path %q", view, dbPath+".bak")
	}
	settingsPath := filepath.Join(filepath.Dir(dbPath), "settings.toml")
	if !strings.Contains(view, "Конфиг") {
		t.Fatalf("settings tab = %q, want config label", view)
	}
	if !strings.Contains(view, settingsPath) {
		t.Fatalf("settings tab = %q, want config path %q", view, settingsPath)
	}
	if !strings.Contains(view, "Бэкап конфига") {
		t.Fatalf("settings tab = %q, want config backup label", view)
	}
	if !strings.Contains(view, settingsPath+".bak") {
		t.Fatalf("settings tab = %q, want config backup path %q", view, settingsPath+".bak")
	}
}

func TestRenderSettingsTab_ShowsFocusedOptionDescription(t *testing.T) {
	store, err := db.Open(filepath.Join(t.TempDir(), "data.db"))
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	defer store.Close()

	m := Model{
		store:  store,
		lang:   i18n.RU,
		width:  120,
		height: 40,
		cursor: settingsRowLanguage,
	}
	view := m.renderSettingsTab()
	if !strings.Contains(view, i18n.New(i18n.RU).SettingsDescLanguage()) {
		t.Fatalf("settings tab = %q, want language description", view)
	}

	m.cursor = settingsRowAutoBackup
	view = m.renderSettingsTab()
	if !strings.Contains(view, i18n.New(i18n.RU).SettingsDescAutoBackup()) {
		t.Fatalf("settings tab = %q, want auto-backup description", view)
	}
	if strings.Contains(view, i18n.New(i18n.RU).SettingsDescLanguage()) {
		t.Fatal("settings tab should not show previous option description")
	}
}
