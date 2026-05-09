package ui

import (
	"fmt"
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

const (
	WindPanelX = 1050
	WindPanelY = 10
	WindPanelW = 220
	WindPanelH = 170
)

// DrawWindPanel renders the wind speed/direction panel in the top-right corner.
func DrawWindPanel(dst *ebiten.Image, wind sim.WindField) {
	// Background
	vector.DrawFilledRect(dst,
		float32(WindPanelX), float32(WindPanelY),
		float32(WindPanelW), float32(WindPanelH),
		color.RGBA{0x00, 0x00, 0x00, 0xAA}, false)

	ebitenutil.DebugPrintAt(dst, "WIND", WindPanelX+90, WindPanelY+6)
	ebitenutil.DebugPrintAt(dst,
		fmt.Sprintf("%.1f m/s", wind.Speed),
		WindPanelX+80, WindPanelY+22)

	// Compass arrow: centre of panel
	cx := float32(WindPanelX + WindPanelW/2)
	cy := float32(WindPanelY + WindPanelH/2 + 20)
	r := float32(50)

	// Circle background
	vector.StrokeCircle(dst, cx, cy, r, 1, color.RGBA{0x88, 0x88, 0x88, 0xff}, false)

	// Arrow tip points in the direction the wind is blowing FROM (source direction).
	// Direction=0 means from +x; screen: +x is right, but Y is flipped.
	// We rotate -direction because screen Y is flipped.
	angle := wind.Direction
	tx := cx + r*float32(math.Cos(-angle))
	ty := cy + r*float32(math.Sin(-angle))
	// Arrow from opposite side toward centre
	bx := cx - r*0.6*float32(math.Cos(-angle))
	by := cy - r*0.6*float32(math.Sin(-angle))

	vector.StrokeLine(dst, bx, by, tx, ty, 2, color.RGBA{0xFF, 0xAA, 0x00, 0xff}, false)
	// Arrowhead
	drawArrowHead(dst, bx, by, tx, ty, color.RGBA{0xFF, 0xAA, 0x00, 0xff})

	ebitenutil.DebugPrintAt(dst,
		fmt.Sprintf("%.0f°", math.Mod(wind.Direction*180/math.Pi, 360)),
		WindPanelX+90, WindPanelY+WindPanelH-20)
}

// DrawRudderIndicator renders a simple arc indicator at the bottom centre.
func DrawRudderIndicator(dst *ebiten.Image, world *sim.World, screenW, screenH int) {
	cx := float32(screenW / 2)
	cy := float32(screenH - 40)
	r := float32(30)

	active := world.ActiveBoat()
	var rudder float64
	if active != nil {
		rudder = active.Rudder
	}

	// Background arc (full range)
	vector.StrokeCircle(dst, cx, cy, r, 1, color.RGBA{0x44, 0x44, 0x44, 0xff}, false)

	// Rudder line
	angle := -math.Pi/2 + rudder // -90° is straight down in screen space
	tx := cx + r*float32(math.Cos(angle))
	ty := cy + r*float32(math.Sin(angle))
	vector.StrokeLine(dst, cx, cy, tx, ty, 3, color.RGBA{0x00, 0xFF, 0xAA, 0xff}, false)

	ebitenutil.DebugPrintAt(dst, "RDR", int(cx)-10, int(cy+r)+4)
}

// DrawTensionTooltip shows line tension near the cursor when hovering a line.
func DrawTensionTooltip(dst *ebiten.Image, world *sim.World, mouseX, mouseY int, screenW int) {
	const hoverPx = 10.0

	bestDist := math.MaxFloat64
	bestTension := -1.0

	for i := range world.Lines {
		line := &world.Lines[i]
		var boat *sim.Boat
		for _, b := range world.Boats {
			if b.ID == line.BoatID {
				boat = b
				break
			}
		}
		if boat == nil {
			continue
		}
		// We need screen coords of the line midpoint.
		// ui package doesn't have Viewport, so we skip this for now — handled in main.
		_ = line
		_ = boat
	}

	if bestTension >= 0 && bestDist <= hoverPx {
		ebitenutil.DebugPrintAt(dst,
			fmt.Sprintf("%.0f N", bestTension),
			mouseX+10, mouseY-10)
	}
}

// drawArrowHead draws a small triangle at the tip of a line from (bx,by) to (tx,ty).
func drawArrowHead(dst *ebiten.Image, bx, by, tx, ty float32, clr color.RGBA) {
	dx := tx - bx
	dy := ty - by
	length := float32(math.Sqrt(float64(dx*dx + dy*dy)))
	if length < 1 {
		return
	}
	// Normalised direction
	nx := dx / length
	ny := dy / length
	// Perpendicular
	px := -ny
	py := nx
	// Head size
	hs := float32(8)
	p1x := tx - nx*hs + px*hs*0.4
	p1y := ty - ny*hs + py*hs*0.4
	p2x := tx - nx*hs - px*hs*0.4
	p2y := ty - ny*hs - py*hs*0.4
	vector.StrokeLine(dst, tx, ty, p1x, p1y, 2, clr, false)
	vector.StrokeLine(dst, tx, ty, p2x, p2y, 2, clr, false)
}
