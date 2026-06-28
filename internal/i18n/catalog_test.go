package i18n

import (
	"strings"
	"testing"
	"time"
)

func TestTranslateLogMessage(t *testing.T) {
	en := New(EN)
	ru := New(RU)

	if got := en.TranslateLogMessage("Полив нормальный (внесено 2л воды)"); got != "Normal watering (2L water added)" {
		t.Fatalf("RU→EN: %q", got)
	}
	if got := ru.TranslateLogMessage("Normal watering (2L water added)"); got != "Полив нормальный (внесено 2л воды)" {
		t.Fatalf("EN→RU: %q", got)
	}
	if got := en.TranslateLogMessage("unknown message"); got != "unknown message" {
		t.Fatalf("unknown should pass through: %q", got)
	}
}

func TestFormatDate(t *testing.T) {
	d := time.Date(2026, 6, 15, 12, 0, 0, 0, time.UTC)
	en := New(EN).FormatDate(d)
	if !strings.Contains(en, "June") || !strings.Contains(en, "15") {
		t.Fatalf("EN date = %q", en)
	}
	ru := New(RU).FormatDate(d)
	if ru != "15 июня" {
		t.Fatalf("RU date = %q, want 15 июня", ru)
	}
}

func TestFormatDateTime(t *testing.T) {
	d := time.Date(2026, 3, 5, 14, 30, 0, 0, time.UTC)
	en := New(EN).FormatDateTime(d)
	if !strings.Contains(en, "Mar") || !strings.Contains(en, "14:30") {
		t.Fatalf("EN datetime = %q", en)
	}
	ru := New(RU).FormatDateTime(d)
	if !strings.Contains(ru, "5 мар") || !strings.Contains(ru, "14:30") {
		t.Fatalf("RU datetime = %q", ru)
	}
}

func TestFormatLogLine(t *testing.T) {
	d := time.Date(2026, 1, 10, 0, 0, 0, 0, time.UTC)
	line := New(RU).FormatLogLine(d, "Полив нормальный (внесено 2л воды)")
	if !strings.Contains(line, "10 января") {
		t.Fatalf("log line = %q", line)
	}
	if !strings.Contains(line, "Полив нормальный") {
		t.Fatalf("log line should contain message: %q", line)
	}
}

func TestCLIWateringInDays_RussianPlural(t *testing.T) {
	ru := New(RU)
	cases := map[int]string{
		1:  "Через 1 день",
		2:  "Через 2 дня",
		4:  "Через 4 дня",
		5:  "Через 5 дней",
		11: "Через 11 дней",
		21: "Через 21 день",
		22: "Через 22 дня",
	}
	for days, want := range cases {
		if got := ru.CLIWateringInDays(days); got != want {
			t.Fatalf("days=%d: got %q, want %q", days, got, want)
		}
	}
}

func TestSettingsDescriptions_NonEmptyBothLangs(t *testing.T) {
	for _, lang := range []Lang{EN, RU} {
		c := New(lang)
		descs := []string{
			c.SettingsDescLanguage(),
			c.SettingsDescCity(),
			c.SettingsDescWeatherRefresh(),
			c.SettingsDescDefaultInterval(),
			c.SettingsDescFallbackTemp(),
			c.SettingsDescAutoBackup(),
			c.SettingsDescTransparent(),
			c.SettingsDescFastBoot(),
		}
		for i, got := range descs {
			if strings.TrimSpace(got) == "" {
				t.Fatalf("lang=%s desc[%d] is empty", lang, i)
			}
		}
	}
}

func TestLogRainShifted(t *testing.T) {
	en := New(EN).LogRainShifted(5, 16, 8.5)
	if !strings.Contains(en, "5") || !strings.Contains(en, "16") {
		t.Fatalf("EN rain log = %q", en)
	}
	ru := New(RU).LogRainShifted(5, 16, 8.5)
	if !strings.Contains(ru, "5") || !strings.Contains(ru, "16") {
		t.Fatalf("RU rain log = %q", ru)
	}
}

func TestRowStatusShiftedRain_IncludesFormattedDate(t *testing.T) {
	d := time.Date(2026, 7, 5, 0, 0, 0, 0, time.UTC)

	ru := New(RU).RowStatusShiftedRain(d)
	if ru != "Полив: 5 июля (дождь)" {
		t.Fatalf("RU status = %q", ru)
	}

	en := New(EN).RowStatusShiftedRain(d)
	if en != "Water: July 5 (rain)" {
		t.Fatalf("EN status = %q", en)
	}
}

func TestFooterHints_NoConfirmKey(t *testing.T) {
	for _, lang := range []Lang{EN, RU} {
		for _, h := range New(lang).FooterHints() {
			if h.Key == "o" {
				t.Fatalf("%s FooterHints still lists confirm key", lang)
			}
		}
	}
}
