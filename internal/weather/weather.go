package weather

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

const (
	geocodingURL = "https://geocoding-api.open-meteo.com/v1/search"
	forecastURL  = "https://api.open-meteo.com/v1/forecast"
)

type Location struct {
	Name      string  `json:"name"`
	Admin1    string  `json:"admin1"`
	Country   string  `json:"country"`
	Latitude  float64 `json:"latitude"`
	Longitude float64 `json:"longitude"`
	Timezone  string  `json:"timezone"`
}

func (l Location) DisplayName() string {
	parts := []string{l.Name}
	if l.Admin1 != "" {
		parts = append(parts, l.Admin1)
	}
	if l.Country != "" {
		parts = append(parts, l.Country)
	}
	return strings.Join(parts, ", ")
}

func (l Location) Matches(other Location) bool {
	const eps = 0.05
	return math.Abs(l.Latitude-other.Latitude) < eps && math.Abs(l.Longitude-other.Longitude) < eps
}

func LocalizeName(items []Location, target Location) Location {
	for _, item := range items {
		if item.Matches(target) {
			return mergeLocation(target, item)
		}
	}
	if len(items) > 0 {
		return mergeLocation(target, items[0])
	}
	return target
}

func mergeLocation(stored, localized Location) Location {
	out := localized
	out.Latitude = stored.Latitude
	out.Longitude = stored.Longitude
	if stored.Timezone != "" {
		out.Timezone = stored.Timezone
	}
	return out
}

type Conditions struct {
	TemperatureC       float64
	PrecipMM           float64
	PrevDayPrecipMM    float64
	PrevDayPrecipKnown bool
	WeatherCode        int
}

type Service interface {
	SearchCities(ctx context.Context, query, language string, count int) ([]Location, error)
	Current(ctx context.Context, loc Location) (Conditions, error)
}

type OpenMeteoService struct {
	client *http.Client
}

func NewOpenMeteo(client *http.Client) *OpenMeteoService {
	if client == nil {
		client = &http.Client{Timeout: 8 * time.Second}
	}
	return &OpenMeteoService{client: client}
}

func (s *OpenMeteoService) SearchCities(ctx context.Context, query, language string, count int) ([]Location, error) {
	query = strings.TrimSpace(query)
	if len(query) < 2 {
		return nil, nil
	}
	if count < 1 {
		count = 8
	}
	if count > 100 {
		count = 100
	}
	if language == "" {
		language = "en"
	}
	u, err := url.Parse(geocodingURL)
	if err != nil {
		return nil, err
	}
	q := u.Query()
	q.Set("name", query)
	q.Set("count", strconv.Itoa(count))
	q.Set("language", strings.ToLower(language))
	q.Set("format", "json")
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("geocoding request failed: %s", resp.Status)
	}

	var payload struct {
		Results []Location `json:"results"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return nil, err
	}
	return payload.Results, nil
}

func (s *OpenMeteoService) Current(ctx context.Context, loc Location) (Conditions, error) {
	u, err := url.Parse(forecastURL)
	if err != nil {
		return Conditions{}, err
	}
	q := u.Query()
	q.Set("latitude", strconv.FormatFloat(loc.Latitude, 'f', 4, 64))
	q.Set("longitude", strconv.FormatFloat(loc.Longitude, 'f', 4, 64))
	q.Set("current", "temperature_2m,precipitation,weather_code")
	q.Set("daily", "precipitation_sum")
	q.Set("past_days", "1")
	q.Set("forecast_days", "1")
	if loc.Timezone != "" {
		q.Set("timezone", loc.Timezone)
	} else {
		q.Set("timezone", "auto")
	}
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return Conditions{}, err
	}
	resp, err := s.client.Do(req)
	if err != nil {
		return Conditions{}, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return Conditions{}, fmt.Errorf("forecast request failed: %s", resp.Status)
	}

	var payload struct {
		Current struct {
			Temperature2M float64 `json:"temperature_2m"`
			Precipitation float64 `json:"precipitation"`
			WeatherCode   int     `json:"weather_code"`
		} `json:"current"`
		Daily struct {
			Time             []string  `json:"time"`
			PrecipitationSum []float64 `json:"precipitation_sum"`
		} `json:"daily"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
		return Conditions{}, err
	}

	prevDayPrecipMM, prevDayKnown := pickPreviousDayPrecipitationSum(loc, payload.Daily.Time, payload.Daily.PrecipitationSum)
	return Conditions{
		TemperatureC:       payload.Current.Temperature2M,
		PrecipMM:           payload.Current.Precipitation,
		PrevDayPrecipMM:    prevDayPrecipMM,
		PrevDayPrecipKnown: prevDayKnown,
		WeatherCode:        payload.Current.WeatherCode,
	}, nil
}

func pickPreviousDayPrecipitationSum(loc Location, days []string, sums []float64) (float64, bool) {
	if len(days) == 0 || len(sums) == 0 {
		return 0, false
	}
	n := len(days)
	if len(sums) < n {
		n = len(sums)
	}
	if n == 0 {
		return 0, false
	}
	prevDay := dayInLocation(loc, -1)
	for i := 0; i < n; i++ {
		if days[i] == prevDay {
			return sums[i], true
		}
	}
	return 0, false
}

func dayInLocation(loc Location, dayOffset int) string {
	if loc.Timezone != "" {
		if tz, err := time.LoadLocation(loc.Timezone); err == nil {
			return time.Now().In(tz).AddDate(0, 0, dayOffset).Format("2006-01-02")
		}
	}
	return time.Now().AddDate(0, 0, dayOffset).Format("2006-01-02")
}
