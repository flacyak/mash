package components

import "testing"

func TestHexToANSI(t *testing.T) {
	tests := []struct {
		hex  string
		want int
	}{
		{"#1292b4", 31},  // R=18 G=146 B=180 → (0,2,3) → 16+0+12+3=31
		{"#000000", 16},  // (0,0,0)
		{"#ffffff", 231}, // (5,5,5) → 16+180+30+5
		{"#ff0000", 196}, // (5,0,0) → 16+180+0+0
		{"#00ff00", 46},  // (0,5,0) → 16+0+30+0
		{"#0000ff", 21},  // (0,0,5) → 16+0+0+5
		{"#57c7ff", 81},  // R=87→95(1) G=199→215(4) B=255→255(5)
	}

	for _, tt := range tests {
		got, err := HexToANSI(tt.hex)
		if err != nil {
			t.Errorf("HexToANSI(%q) error: %v", tt.hex, err)
			continue
		}
		if got != tt.want {
			t.Errorf("HexToANSI(%q) = %d, want %d", tt.hex, got, tt.want)
		}
	}
}

func TestHexToANSIString(t *testing.T) {
	if s := HexToANSIString("#1292b4"); s != "31" {
		t.Errorf("HexToANSIString(#1292b4) = %q, want %q", s, "31")
	}
}

func TestHexToANSIInvalid(t *testing.T) {
	_, err := HexToANSI("xyz")
	if err == nil {
		t.Error("expected error for invalid hex")
	}
}
