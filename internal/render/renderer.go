package render

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dvpalkin/mooring-simulator/internal/input"
	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

var (
	waterColor      = color.RGBA{0x1A, 0x6B, 0x8A, 0xff}
	panelBg         = color.RGBA{0x18, 0x18, 0x28, 0xE0}
	panelDivider    = color.RGBA{0x44, 0x44, 0x66, 0xff}
	mooringCleatClr = color.RGBA{0xFF, 0xCC, 0x00, 0xff} // yellow ring
	mooringCleatDot = color.RGBA{0xFF, 0x44, 0x22, 0xff} // red-orange centre
	mooringBuoyClr  = color.RGBA{0xFF, 0xCC, 0x00, 0xff}
	mooringBuoyDot  = color.RGBA{0xFF, 0x88, 0x00, 0xff} // orange centre
	ghostLineClr    = color.RGBA{0xFF, 0xFF, 0x44, 0xAA}
	ghostBoatClr    = color.RGBA{0xFF, 0xFF, 0xFF, 0x66}
)

const (
	panelW       = 100 // matches input.panelWidth
	iconMonoY    = 220 // centre Y of monohull icon in panel
	iconCatY     = 480 // centre Y of catamaran icon in panel
	iconSize     = 32  // icon radius in pixels
)

// Renderer draws the simulation world to an Ebiten image.
type Renderer struct {
	VP Viewport
}

// Draw renders one complete frame.
func (r *Renderer) Draw(dst *ebiten.Image, world *sim.World, handler *input.Handler) {
	dst.Fill(waterColor)
	drawDock(dst, &world.Dock, r.VP)
	drawMooringPoints(dst, world, handler, r.VP)
	drawMooringLines(dst, world, r.VP)
	for _, b := range world.Boats {
		drawBoat(dst, b, r.VP)
	}
	if handler.PlacementActive() {
		drawGhostLine(dst, handler, r.VP)
	}
	if handler.DraggingBoat() {
		drawGhostBoat(dst, handler, r.VP)
	}
	drawLeftPanel(dst, handler)
}

// drawMooringPoints renders pre-defined mooring points as concentric circles.
func drawMooringPoints(dst *ebiten.Image, world *sim.World, handler *input.Handler, vp Viewport) {
	for i := range world.MooringPoints {
		mp := &world.MooringPoints[i]
		sx, sy := vp.WorldToScreen(mp.Pos)
		outer, inner := mooringCleatClr, mooringCleatDot
		if mp.Kind == sim.MooringBuoy {
			outer = mooringBuoyClr
			inner = mooringBuoyDot
		}
		// Highlight when this point is selected in placement mode.
		if handler.PlacementActive() {
			p := handler.PendingDockPoint()
			if p.Sub(mp.Pos).Len() < 0.01 {
				outer = color.RGBA{0xFF, 0xFF, 0xFF, 0xff}
			}
		}
		drawCircle(dst, float32(sx), float32(sy), 7, outer)
		drawCircle(dst, float32(sx), float32(sy), 3.5, inner)
	}
}

// drawGhostLine draws the in-progress mooring line from selected point to cursor.
func drawGhostLine(dst *ebiten.Image, handler *input.Handler, vp Viewport) {
	dockPt := handler.PendingDockPoint()
	ax, ay := vp.WorldToScreen(dockPt)
	cx, cy := ebiten.CursorPosition()
	vector.StrokeLine(dst, float32(ax), float32(ay), float32(cx), float32(cy), 1.5, ghostLineClr, false)
	drawCircle(dst, float32(ax), float32(ay), 6, color.RGBA{0xFF, 0xFF, 0x00, 0xff})
}

// drawGhostBoat draws a semi-transparent boat silhouette at the cursor position.
func drawGhostBoat(dst *ebiten.Image, handler *input.Handler, vp Viewport) {
	cx, cy := ebiten.CursorPosition()
	worldPt := vp.ScreenToWorld(float64(cx), float64(cy))

	// Create a temporary boat for vertex calculation; heading π/2 = bow north.
	var b *sim.Boat
	switch handler.DraggingBoatType() {
	case sim.BoatMonohull:
		b = sim.NewMonohull(0, worldPt, math.Pi/2)
	case sim.BoatCatamaran:
		b = sim.NewCatamaran(0, worldPt, math.Pi/2)
	}
	if b == nil {
		return
	}

	switch b.Type {
	case sim.BoatMonohull:
		sp := worldToScreenPts(monohullWorldPts(b), vp)
		drawFilledPolygon(dst, sp, ghostBoatClr)
	case sim.BoatCatamaran:
		port, stbd, cross := catamaranWorldPts(b)
		for _, pts := range [][]physics.Vec2{port, stbd, cross} {
			sp := worldToScreenPts(pts, vp)
			drawFilledPolygon(dst, sp, ghostBoatClr)
		}
	}
}

// drawLeftPanel renders the boat-icon selection panel on the left edge.
func drawLeftPanel(dst *ebiten.Image, handler *input.Handler) {
	// Background
	vector.DrawFilledRect(dst, 0, 0, panelW, float32(dst.Bounds().Dy()), panelBg, false)
	// Divider line on right edge of panel
	vector.StrokeLine(dst, panelW, 0, panelW, float32(dst.Bounds().Dy()), 1, panelDivider, false)

	// Label
	ebitenutil.DebugPrintAt(dst, "BOATS", 26, 12)

	// Divider between label and icons
	vector.StrokeLine(dst, 5, 28, panelW-5, 28, 1, panelDivider, false)

	// Monohull icon and label
	DrawBoatIcon(dst, sim.BoatMonohull, panelW/2, iconMonoY, iconSize)
	ebitenutil.DebugPrintAt(dst, "Mono", 22, iconMonoY+iconSize+6)

	// Divider
	midY := float32((iconMonoY + iconCatY) / 2)
	vector.StrokeLine(dst, 5, midY, panelW-5, midY, 1, panelDivider, false)

	// Catamaran icon and label
	DrawBoatIcon(dst, sim.BoatCatamaran, panelW/2, iconCatY, iconSize)
	ebitenutil.DebugPrintAt(dst, "Cat", 27, iconCatY+iconSize+6)

	// Highlight the icon being dragged
	if handler.DraggingBoat() {
		var hy float32
		if handler.DraggingBoatType() == sim.BoatMonohull {
			hy = iconMonoY
		} else {
			hy = iconCatY
		}
		vector.StrokeCircle(dst, panelW/2, hy, iconSize+4, 2, color.RGBA{0xFF, 0xFF, 0x00, 0xff}, false)
	}
}
