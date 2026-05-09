package render

import "github.com/dvpalkin/mooring-simulator/internal/physics"

// Viewport maps between physics world-space (meters, Y-up) and screen-space (pixels, Y-down).
type Viewport struct {
	OriginWorld physics.Vec2 // world-space point at screen top-left
	Scale       float64      // pixels per meter; default 20
	ScreenW     int
	ScreenH     int
}

// DefaultViewport returns a viewport centred on the world origin for a 1280×720 screen.
func DefaultViewport() Viewport {
	const scale = 20.0
	return Viewport{
		OriginWorld: physics.Vec2{X: -1280.0 / (2 * scale), Y: -720.0 / (2 * scale)},
		Scale:       scale,
		ScreenW:     1280,
		ScreenH:     720,
	}
}

// WorldToScreen converts a world-space point to screen-space (float64 for sub-pixel precision).
// Higher world Y maps to lower screen Y (Y-flip at this boundary per architecture D2).
func (v Viewport) WorldToScreen(p physics.Vec2) (x, y float64) {
	x = (p.X - v.OriginWorld.X) * v.Scale
	y = float64(v.ScreenH) - (p.Y-v.OriginWorld.Y)*v.Scale
	return
}

// ScreenToWorld converts screen-space integer coordinates to world-space.
func (v Viewport) ScreenToWorld(sx, sy float64) physics.Vec2 {
	return physics.Vec2{
		X: sx/v.Scale + v.OriginWorld.X,
		Y: (float64(v.ScreenH)-sy)/v.Scale + v.OriginWorld.Y,
	}
}
