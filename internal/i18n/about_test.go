package i18n

import (
	"testing"

	"kiri/internal/version"
)

func TestAboutInfoLines(t *testing.T) {
	en := New(EN)
	enLines := en.AboutInfoLines()
	if len(enLines) != 13 {
		t.Fatalf("lines count = %d, want 13", len(enLines))
	}
	if enLines[0] != version.Label() {
		t.Fatalf("version line = %q", enLines[0])
	}
	if enLines[3] != en.AboutTagline() {
		t.Fatalf("tagline = %q", enLines[3])
	}
	wantENContact := []string{
		"Author:  Pavel Antonov",
		"Contact: freislot@gmail.com",
		"GitHub:  github.com/freislot/kiri",
	}
	for i, want := range wantENContact {
		if enLines[10+i] != want {
			t.Fatalf("EN contact[%d] = %q, want %q", i, enLines[10+i], want)
		}
	}

	ru := New(RU)
	ruLines := ru.AboutInfoLines()
	if ruLines[3] != "Следит за вашими растениями и подсказывает, когда их поливать." {
		t.Fatalf("RU tagline = %q", ruLines[3])
	}
	wantRUContact := []string{
		"Автор:    Павел Антонов",
		"Контакты: freislot@gmail.com",
		"GitHub:   github.com/freislot/kiri",
	}
	for i, want := range wantRUContact {
		if ruLines[10+i] != want {
			t.Fatalf("RU contact[%d] = %q, want %q", i, ruLines[10+i], want)
		}
	}
	if ruLines[2] != "" {
		t.Fatalf("expected blank line after subtitle, got %q", ruLines[2])
	}
}
