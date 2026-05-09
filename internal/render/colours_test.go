package render

import (
	"image/color"
	"testing"
)

func TestLineTensionColor(t *testing.T) {
	cases := []struct {
		fraction float64
		want     color.RGBA
	}{
		{0, color.RGBA{0x88, 0x88, 0x88, 0xff}},    // slack
		{-1, color.RGBA{0x88, 0x88, 0x88, 0xff}},   // below zero = slack
		{0.1, color.RGBA{0x88, 0xcc, 0x88, 0xff}},  // green
		{0.5, color.RGBA{0xcc, 0xcc, 0x00, 0xff}},  // yellow
		{0.85, color.RGBA{0xcc, 0x44, 0x00, 0xff}}, // orange
		{1.0, color.RGBA{0xff, 0x00, 0x00, 0xff}},  // red
		{2.0, color.RGBA{0xff, 0x00, 0x00, 0xff}},  // above max = still red
	}
	for _, tc := range cases {
		got := lineTensionColor(tc.fraction)
		if got != tc.want {
			t.Errorf("lineTensionColor(%v) = %v want %v", tc.fraction, got, tc.want)
		}
	}
}
