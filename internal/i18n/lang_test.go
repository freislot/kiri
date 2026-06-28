package i18n

import "testing"

func TestParse(t *testing.T) {
	tests := []struct {
		in   string
		want Lang
	}{
		{"en", EN},
		{"EN", EN},
		{" ru ", RU},
		{"", RU},
		{"de", RU},
	}
	for _, tt := range tests {
		if got := Parse(tt.in); got != tt.want {
			t.Fatalf("Parse(%q) = %s, want %s", tt.in, got, tt.want)
		}
	}
}

func TestLangToggle(t *testing.T) {
	if EN.Toggle() != RU {
		t.Fatal("EN.Toggle should be RU")
	}
	if RU.Toggle() != EN {
		t.Fatal("RU.Toggle should be EN")
	}
}
