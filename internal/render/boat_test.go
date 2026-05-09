package render

import (
	"math"
	"testing"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

// allDistinct returns true when no two points in pts are closer than eps.
func allDistinct(pts []physics.Vec2, eps float64) bool {
	for i := range pts {
		for j := i + 1; j < len(pts); j++ {
			d := pts[i].Sub(pts[j]).Len()
			if d < eps {
				return false
			}
		}
	}
	return true
}

func TestMonohullHullPtsNonDegenerate(t *testing.T) {
	vp := DefaultViewport()
	boat := sim.NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	pts := HullScreenPts(boat, vp)
	if len(pts) < 5 {
		t.Fatalf("expected at least 5 hull vertices, got %d", len(pts))
	}
	if !allDistinct(pts, 1.0) {
		t.Errorf("hull vertices are not all distinct (degenerate polygon): %v", pts)
	}
}

func TestMonohullHullPtsRotated(t *testing.T) {
	vp := DefaultViewport()
	boat := sim.NewMonohull(1, physics.Vec2{X: 0, Y: 0}, math.Pi/2)
	pts := HullScreenPts(boat, vp)
	if !allDistinct(pts, 1.0) {
		t.Errorf("rotated hull vertices are degenerate: %v", pts)
	}
}

func TestMonohullBowIsForward(t *testing.T) {
	// With heading=0, pts[0] is the bow tip (+X) → highest screen X.
	// pts[4] is the stern centre (−X) → lowest screen X.
	vp := DefaultViewport()
	boat := sim.NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	sp := HullScreenPts(boat, vp)
	bowX := sp[0].X
	// Find the minimum screen X among all points (that is the stern side).
	minX := sp[0].X
	for _, p := range sp[1:] {
		if p.X < minX {
			minX = p.X
		}
	}
	if bowX <= minX {
		t.Errorf("heading=0: bow screen X (%v) should be > minimum screen X (%v)", bowX, minX)
	}
}
