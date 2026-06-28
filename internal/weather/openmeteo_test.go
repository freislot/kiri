package weather

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

func TestSearchCities_HTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/search" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"results":[{"name":"Berlin","admin1":"Berlin","country":"Germany","latitude":52.52,"longitude":13.405,"timezone":"Europe/Berlin"}]}`))
	}))
	defer srv.Close()

	svc := NewOpenMeteo(srv.Client())
	svc.client.Transport = rewriteHostTransport{base: srv.URL, inner: srv.Client().Transport}

	results, err := svc.SearchCities(context.Background(), "Berlin", "en", 5)
	if err != nil {
		t.Fatalf("SearchCities: %v", err)
	}
	if len(results) != 1 || results[0].Name != "Berlin" {
		t.Fatalf("results = %+v", results)
	}
}

func TestCurrent_HTTP(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/forecast" {
			t.Fatalf("path = %s", r.URL.Path)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"current":{"temperature_2m":23.5,"precipitation":0.2,"weather_code":3},
			"daily":{"time":["2026-06-27","2026-06-28"],"precipitation_sum":[6.0,0.0]}
		}`))
	}))
	defer srv.Close()

	svc := NewOpenMeteo(srv.Client())
	svc.client.Transport = rewriteHostTransport{base: srv.URL, inner: srv.Client().Transport}

	loc := Location{Name: "Test", Latitude: 52.5, Longitude: 13.4, Timezone: "UTC"}
	cond, err := svc.Current(context.Background(), loc)
	if err != nil {
		t.Fatalf("Current: %v", err)
	}
	if cond.TemperatureC != 23.5 {
		t.Fatalf("temp = %.1f", cond.TemperatureC)
	}
	if cond.PrecipMM != 0.2 {
		t.Fatalf("precip = %.1f", cond.PrecipMM)
	}
	if cond.WeatherCode != 3 {
		t.Fatalf("code = %d", cond.WeatherCode)
	}
}

type rewriteHostTransport struct {
	base  string
	inner http.RoundTripper
}

func (t rewriteHostTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "geocoding-api.open-meteo.com" || req.URL.Host == "api.open-meteo.com" {
		base, err := url.Parse(t.base)
		if err != nil {
			return nil, err
		}
		cloned := req.Clone(req.Context())
		cloned.URL.Scheme = base.Scheme
		cloned.URL.Host = base.Host
		rt := t.inner
		if rt == nil {
			rt = http.DefaultTransport
		}
		return rt.RoundTrip(cloned)
	}
	rt := t.inner
	if rt == nil {
		rt = http.DefaultTransport
	}
	return rt.RoundTrip(req)
}
