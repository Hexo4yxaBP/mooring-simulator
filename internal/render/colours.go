package render

import "image/color"

// lineTensionColor returns the draw colour for a mooring line based on its tension fraction.
// fraction = currentTension / maxTension (where maxTension = NaturalLength * Stiffness * 0.5).
func lineTensionColor(fraction float64) color.RGBA {
	switch {
	case fraction <= 0:
		return color.RGBA{0x88, 0x88, 0x88, 0xff} // slack: grey
	case fraction < 0.3:
		return color.RGBA{0x88, 0xcc, 0x88, 0xff} // light tension: green
	case fraction < 0.7:
		return color.RGBA{0xcc, 0xcc, 0x00, 0xff} // moderate tension: yellow
	case fraction < 1.0:
		return color.RGBA{0xcc, 0x44, 0x00, 0xff} // high tension: orange
	default:
		return color.RGBA{0xff, 0x00, 0x00, 0xff} // overstressed: red
	}
}
