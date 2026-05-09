package ui

import (
	"fmt"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"

	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

// DrawHUD renders the top-left boat status panel.
func DrawHUD(dst *ebiten.Image, world *sim.World) {
	active := world.ActiveBoat()
	if active == nil {
		ebitenutil.DebugPrintAt(dst, "No active boat", 10, 10)
		return
	}

	speed := active.Body.Velocity.Len()
	headingDeg := math.Mod(active.Body.Heading*180/math.Pi, 360)
	if headingDeg < 0 {
		headingDeg += 360
	}

	line0 := fmt.Sprintf("Speed:   %.1f m/s", speed)
	line1 := fmt.Sprintf("Heading: %.0f°", headingDeg)
	line2 := fmt.Sprintf("Throttle: %s", throttleLabel(active.Throttle))
	line3 := fmt.Sprintf("Rudder:  %.0f°", active.Rudder*180/math.Pi)

	ebitenutil.DebugPrintAt(dst, line0, 10, 10)
	ebitenutil.DebugPrintAt(dst, line1, 10, 26)
	ebitenutil.DebugPrintAt(dst, line2, 10, 42)
	ebitenutil.DebugPrintAt(dst, line3, 10, 58)
}

func throttleLabel(t sim.ThrottleState) string {
	switch t {
	case sim.ThrottleNeutral:
		return "Neutral"
	case sim.ThrottleSlowFwd:
		return "Slow Fwd"
	case sim.ThrottleFullFwd:
		return "Full Fwd"
	case sim.ThrottleSlowAstern:
		return "Slow Astern"
	case sim.ThrottleFullAstern:
		return "Full Astern"
	default:
		return "?"
	}
}
