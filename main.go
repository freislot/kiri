package main

import (
	"flag"
	"fmt"
	"os"

	"kiri/internal/db"
	"kiri/internal/statuscli"
	"kiri/internal/tui"
	"kiri/internal/version"
	"kiri/internal/weather"

	tea "github.com/charmbracelet/bubbletea"
)

func main() {
	showVersion := flag.Bool("version", false, "print version")
	showStatus := flag.Bool("status", false, "print status table")
	showSummary := flag.Bool("summary", false, "print one-line status summary")
	showJSON := flag.Bool("json", false, "print data as JSON")
	transparent := flag.Bool("transparent", false, "disable UI background colors (transparent mode)")
	flag.Parse()

	if *showVersion {
		fmt.Println(version.Line())
		return
	}

	dbPath, err := db.DefaultDBPath()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config path error: %v\n", err)
		os.Exit(1)
	}

	store, err := db.Open(dbPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "database error: %v\n", err)
		os.Exit(1)
	}
	defer store.Close()

	if *showStatus {
		if err := statuscli.Print(os.Stdout, store); err != nil {
			fmt.Fprintf(os.Stderr, "status error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *showSummary {
		if err := statuscli.PrintSummary(os.Stdout, store); err != nil {
			fmt.Fprintf(os.Stderr, "summary error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *showJSON {
		if err := statuscli.PrintJSON(os.Stdout, store); err != nil {
			fmt.Fprintf(os.Stderr, "json error: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *transparent {
		if err := store.SetTransparentMode(true); err != nil {
			fmt.Fprintf(os.Stderr, "transparent mode setting error: %v\n", err)
			os.Exit(1)
		}
	}

	if enabled, err := store.GetAutoBackup(); err == nil && enabled {
		backupPath, err := db.DefaultBackupPath()
		if err != nil {
			fmt.Fprintf(os.Stderr, "backup path error: %v\n", err)
		} else if err := store.BackupTo(backupPath); err != nil {
			fmt.Fprintf(os.Stderr, "auto-backup failed: %v\n", err)
		}
	}

	wx := weather.NewOpenMeteo(nil)
	m := tui.New(store, wx, tui.Options{Transparent: *transparent})

	p := tea.NewProgram(m, tea.WithAltScreen(), tea.WithMouseCellMotion())
	if _, err := p.Run(); err != nil {
		fmt.Fprintf(os.Stderr, "runtime error: %v\n", err)
		os.Exit(1)
	}
}
