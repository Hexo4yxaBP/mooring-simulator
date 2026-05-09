package render

import (
	"math"
	"math/rand"
	"testing"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

func TestWorldToScreenOrigin(t *testing.T) {
	vp := Viewport{Scale: 20, ScreenW: 1280, ScreenH: 720, OriginWorld: physics.Vec2{X: 0, Y: 0}}
	x, y := vp.WorldToScreen(physics.Vec2{X: 0, Y: 0})
	if x != 0 || y != 720 {
		t.Errorf("Origin: got (%v,%v) want (0,720)", x, y)
	}
}

func TestWorldToScreenTopRight(t *testing.T) {
	vp := Viewport{Scale: 20, ScreenW: 1280, ScreenH: 720, OriginWorld: physics.Vec2{X: 0, Y: 0}}
	x, y := vp.WorldToScreen(physics.Vec2{X: 64, Y: 36})
	if math.Abs(x-1280) > 1e-9 || math.Abs(y-0) > 1e-9 {
		t.Errorf("Top-right: got (%v,%v) want (1280,0)", x, y)
	}
}

func TestYAxisFlip(t *testing.T) {
	vp := Viewport{Scale: 20, ScreenW: 1280, ScreenH: 720, OriginWorld: physics.Vec2{X: 0, Y: 0}}
	_, y10 := vp.WorldToScreen(physics.Vec2{X: 0, Y: 10})
	_, y0 := vp.WorldToScreen(physics.Vec2{X: 0, Y: 0})
	if y10 >= y0 {
		t.Errorf("Y-flip: higher world Y should map to lower screen Y; y10=%v y0=%v", y10, y0)
	}
}

func TestRoundTrip(t *testing.T) {
	vp := Viewport{Scale: 20, ScreenW: 1280, ScreenH: 720, OriginWorld: physics.Vec2{X: -32, Y: -18}}
	rng := rand.New(rand.NewSource(42))
	for i := 0; i < 10; i++ {
		wx := rng.Float64()*100 - 50
		wy := rng.Float64()*100 - 50
		p := physics.Vec2{X: wx, Y: wy}
		sx, sy := vp.WorldToScreen(p)
		back := vp.ScreenToWorld(sx, sy)
		if math.Abs(back.X-p.X) > 1e-9 || math.Abs(back.Y-p.Y) > 1e-9 {
			t.Errorf("RoundTrip: input %v → screen (%v,%v) → %v", p, sx, sy, back)
		}
	}
}

func TestScaleHalves(t *testing.T) {
	// Doubling scale should halve the world extent visible on screen
	vp1 := Viewport{Scale: 20, ScreenW: 1280, ScreenH: 720, OriginWorld: physics.Vec2{X: 0, Y: 0}}
	vp2 := Viewport{Scale: 40, ScreenW: 1280, ScreenH: 720, OriginWorld: physics.Vec2{X: 0, Y: 0}}
	// World point at screen edge with scale 20
	edge1 := vp1.ScreenToWorld(1280, 0)
	edge2 := vp2.ScreenToWorld(1280, 0)
	// scale=40 shows half the world extent
	if math.Abs(edge2.X-edge1.X/2) > 1e-9 {
		t.Errorf("Doubling scale should halve world X extent: edge1=%v edge2=%v", edge1.X, edge2.X)
	}
}
