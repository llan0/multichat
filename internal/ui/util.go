package ui

import (
	"fmt"
	"image/color"
)

func parseHexColor(hex string) color.Color {
	if len(hex) == 0 {
		return color.White
	}

	if hex[0] == '#' {
		hex = hex[1:]
	}

	if len(hex) != 6 {
		return color.White
	}

	var r, g, b uint8
	_, err := fmt.Sscanf(hex, "%02x%02x%02x", &r, &g, &b)
	if err != nil {
		return color.White
	}

	return color.RGBA{R: r, G: g, B: b, A: 255}
}
