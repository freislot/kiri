package model

import "testing"

func TestDefaultConfigMatchesLegacyConstants(t *testing.T) {
	c := DefaultConfig()
	if c.FallbackTempC != 22.0 {
		t.Fatalf("FallbackTempC = %v, want 22", c.FallbackTempC)
	}
	if c.WaterSoonPercent != 20 {
		t.Fatalf("WaterSoonPercent = %v, want 20", c.WaterSoonPercent)
	}
	if c.PostponeBoostPercent != 15 {
		t.Fatalf("PostponeBoostPercent = %v, want 15", c.PostponeBoostPercent)
	}
	if c.PostponeSuggestAfter != 2 {
		t.Fatalf("PostponeSuggestAfter = %v, want 2", c.PostponeSuggestAfter)
	}
	if c.Season.Summer != 1.5 || c.Season.Winter != 0.5 {
		t.Fatalf("season coeffs = %+v", c.Season)
	}
	if c.Rain.LightMM != 1.0 || c.Rain.HeavyMM != 10.0 || c.Rain.BaseShiftHours != 24.0 {
		t.Fatalf("rain coeffs = %+v", c.Rain)
	}
	if c.Temp.CoolThresholdC != 20 || c.Temp.CoolFactor != 0.8 || c.Temp.HotThresholdC != 27 || c.Temp.HotSlope != 0.05 {
		t.Fatalf("temp coeffs = %+v", c.Temp)
	}
}

func TestNormalizeModelConfigPartialOverride(t *testing.T) {
	c := NormalizeModelConfig(ModelConfig{
		FallbackTempC: 18,
		Season:        SeasonCoeffs{Summer: 2.0},
	})
	if c.FallbackTempC != 18 {
		t.Fatalf("FallbackTempC = %v, want 18", c.FallbackTempC)
	}
	if c.Season.Summer != 2.0 {
		t.Fatalf("Season.Summer = %v, want 2.0", c.Season.Summer)
	}
	if c.Season.Winter != 0.5 {
		t.Fatalf("Season.Winter should fall back to default, got %v", c.Season.Winter)
	}
}

func TestNormalizeModelConfigInvalidFallback(t *testing.T) {
	c := NormalizeModelConfig(ModelConfig{FallbackTempC: 99})
	if c.FallbackTempC != DefaultConfig().FallbackTempC {
		t.Fatalf("invalid fallback should reset to default, got %v", c.FallbackTempC)
	}
}
