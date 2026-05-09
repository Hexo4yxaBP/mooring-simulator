package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

var (
	dockFillColor    = color.RGBA{0x8B, 0x7D, 0x6B, 0xff}
	dockOutlineColor = color.RGBA{0x5C, 0x4E, 0x3D, 0xff}
)

func drawDock(dst *ebiten.Image, dock *sim.Dock, vp Viewport) {
	if len(dock.Vertices) < 3 {
		return
	}
	pts := make([]physics.Vec2, len(dock.Vertices))
	for i, v := range dock.Vertices {
		sx, sy := vp.WorldToScreen(v)
		pts[i] = physics.Vec2{X: sx, Y: sy}
	}
	drawFilledPolygon(dst, pts, dockFillColor)
	drawPolygonOutline(dst, pts, 2, dockOutlineColor)
}
