package weather

import (
	"context"
	"testing"
	"time"
)

func TestPickPreviousDayPrecipitationSum_NoMatchingDay(t *testing.T) {
	loc := Location{Timezone: "UTC"}
	today := dayInLocation(loc, 0)
	_, known := pickPreviousDayPrecipitationSum(loc, []string{today}, []float64{3.2})
	if known {
		t.Fatalf("expected unknown precipitation when previous day is missing")
	}
}

func TestPickPreviousDayPrecipitationSum_MatchingDay(t *testing.T) {
	loc := Location{Timezone: "UTC"}
	prev := dayInLocation(loc, -1)
	today := dayInLocation(loc, 0)
	val, known := pickPreviousDayPrecipitationSum(loc, []string{prev, today}, []float64{8.5, 1.0})
	if !known {
		t.Fatalf("expected known precipitation for previous day")
	}
	if val != 8.5 {
		t.Fatalf("unexpected value: got %.1f want 8.5", val)
	}
}

func TestPickPreviousDayPrecipitationSum_EmptyInput(t *testing.T) {
	loc := Location{Timezone: "UTC"}
	_, known := pickPreviousDayPrecipitationSum(loc, nil, nil)
	if known {
		t.Fatal("empty input should be unknown")
	}
}

func TestDayInLocationUsesOffset(t *testing.T) {
	loc := Location{Timezone: "UTC"}
	t0, _ := time.Parse("2006-01-02", dayInLocation(loc, 0))
	tm1, _ := time.Parse("2006-01-02", dayInLocation(loc, -1))
	if !tm1.Before(t0) {
		t.Fatalf("expected previous day to be before today")
	}
}

func TestLocationDisplayName(t *testing.T) {
	loc := Location{Name: "Moscow", Admin1: "Moscow", Country: "Russia"}
	got := loc.DisplayName()
	want := "Moscow, Moscow, Russia"
	if got != want {
		t.Fatalf("DisplayName = %q, want %q", got, want)
	}

	minimal := Location{Name: "Berlin"}
	if minimal.DisplayName() != "Berlin" {
		t.Fatalf("minimal DisplayName = %q", minimal.DisplayName())
	}
}

func TestLocationMatches(t *testing.T) {
	a := Location{Latitude: 55.75, Longitude: 37.62}
	b := Location{Latitude: 55.76, Longitude: 37.63}
	if !a.Matches(b) {
		t.Fatal("nearby coords should match")
	}
	c := Location{Latitude: 52.52, Longitude: 13.40}
	if a.Matches(c) {
		t.Fatal("distant coords should not match")
	}
}

func TestLocalizeName(t *testing.T) {
	stored := Location{
		Name:      "Moscow",
		Latitude:  55.7558,
		Longitude: 37.6173,
		Timezone:  "Europe/Moscow",
	}
	items := []Location{
		{Name: "Москва", Admin1: "Москва", Country: "Россия", Latitude: 55.75, Longitude: 37.62},
		{Name: "Paris", Latitude: 48.85, Longitude: 2.35},
	}
	got := LocalizeName(items, stored)
	if got.Name != "Москва" {
		t.Fatalf("localized name = %q, want Москва", got.Name)
	}
	if got.Latitude != stored.Latitude || got.Longitude != stored.Longitude {
		t.Fatal("stored coordinates should be preserved")
	}
	if got.Timezone != stored.Timezone {
		t.Fatalf("timezone = %q, want %q", got.Timezone, stored.Timezone)
	}
}

func TestLocalizeName_FallbackToFirst(t *testing.T) {
	stored := Location{Latitude: 0, Longitude: 0}
	items := []Location{{Name: "Lisbon", Latitude: 38.72, Longitude: -9.14}}
	got := LocalizeName(items, stored)
	if got.Name != "Lisbon" {
		t.Fatalf("fallback name = %q", got.Name)
	}
}

func TestLocalizeName_EmptyItems(t *testing.T) {
	stored := Location{Name: "Test", Latitude: 1, Longitude: 2}
	got := LocalizeName(nil, stored)
	if got.Name != stored.Name {
		t.Fatalf("empty items should return stored: %+v", got)
	}
}

func TestSearchCities_ShortQuery(t *testing.T) {
	svc := NewOpenMeteo(nil)
	results, err := svc.SearchCities(context.Background(), "a", "en", 5)
	if err != nil {
		t.Fatalf("SearchCities: %v", err)
	}
	if results != nil {
		t.Fatalf("short query should return nil, got %v", results)
	}
}
