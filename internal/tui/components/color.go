package components

import (
	"fmt"
	"strconv"
	"strings"
)

// HexToANSI converts a hex colour string (e.g. "#1292b4" or "1292b4") to the
// nearest ANSI 256-colour palette index.
//
// The 6×6×6 colour cube occupies indices 16–231. Each channel is quantised to
// one of six levels: 0, 95, 135, 175, 215, 255. The remaining 24 slots (0–15
// and 232–255) are system colours and greyscale.
func HexToANSI(hex string) (int, error) {
	hex = strings.TrimPrefix(hex, "#")
	if len(hex) != 6 {
		return 0, fmt.Errorf("hex colour must be 6 digits, got %q", hex)
	}

	r, err := strconv.ParseUint(hex[0:2], 16, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid red component: %w", err)
	}
	g, err := strconv.ParseUint(hex[2:4], 16, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid green component: %w", err)
	}
	b, err := strconv.ParseUint(hex[4:6], 16, 8)
	if err != nil {
		return 0, fmt.Errorf("invalid blue component: %w", err)
	}

	levels := []uint8{0, 95, 135, 175, 215, 255}

	ri := nearestLevel(uint8(r), levels)
	gi := nearestLevel(uint8(g), levels)
	bi := nearestLevel(uint8(b), levels)

	// Colour cube: 16 + 36*r + 6*g + b
	return 16 + 36*int(ri) + 6*int(gi) + int(bi), nil
}

// HexToANSIString is a convenience wrapper that returns the ANSI colour index
// as a string, suitable for direct use with lipgloss.Color.
func HexToANSIString(hex string) string {
	n, err := HexToANSI(hex)
	if err != nil {
		return "0"
	}
	return strconv.Itoa(n)
}

func nearestLevel(v uint8, levels []uint8) uint8 {
	bestIdx := 0
	bestDist := 999
	for i, l := range levels {
		d := int(v) - int(l)
		if d < 0 {
			d = -d
		}
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	return uint8(bestIdx)
}
