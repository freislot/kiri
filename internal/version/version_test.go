package version

import "testing"

func TestLine(t *testing.T) {
	if got := Line(); got != "kiri v0.1.0" {
		t.Fatalf("Line() = %q", got)
	}
}

func TestLabel(t *testing.T) {
	if got := Label(); got != "kiri (v0.1.0)" {
		t.Fatalf("Label() = %q", got)
	}
}
