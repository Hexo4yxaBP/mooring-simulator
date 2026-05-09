package sim

import (
	"math"
	"testing"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

func TestNewBottomPier(t *testing.T) {
	d := NewBottomPier(physics.Vec2{X: 0, Y: -13}, 64, 8)
	if len(d.Vertices) != 4 {
		t.Errorf("bottom pier should have 4 vertices, got %d", len(d.Vertices))
	}
	// Face is at Y=-13; body extends to Y=-21.
	for _, v := range d.Vertices {
		if v.Y > -13+1e-9 && v.Y < -13-1e-9 {
			// each vertex is either on the face or on the back
		}
	}
	// Nearest point south of face should snap to face
	p := physics.Vec2{X: 0, Y: -12}
	nearest := d.NearestPointOnEdge(p)
	if math.Abs(nearest.Y-(-13)) > 1e-9 {
		t.Errorf("nearest Y should be -13 (face), got %v", nearest.Y)
	}
}

func TestNewStraightPier(t *testing.T) {
	d := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	if len(d.Vertices) != 4 {
		t.Errorf("Straight pier should have 4 vertices, got %d", len(d.Vertices))
	}
}

func TestNearestPointOnEdgeFront(t *testing.T) {
	// Dock: x=-25..25, y=30..35
	d := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	// Point directly south of dock face
	p := physics.Vec2{X: 5, Y: 25}
	nearest := d.NearestPointOnEdge(p)
	// Should snap to point on south face (y=30)
	if math.Abs(nearest.Y-30) > 1e-9 {
		t.Errorf("Nearest point Y should be 30 (south face), got %v", nearest.Y)
	}
	if math.Abs(nearest.X-5) > 1e-9 {
		t.Errorf("Nearest point X should be 5, got %v", nearest.X)
	}
}

func TestNearestPointOnEdgeBeside(t *testing.T) {
	// Dock: x=-25..25, y=30..35
	d := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	// Point to the right of the dock
	p := physics.Vec2{X: 30, Y: 32}
	nearest := d.NearestPointOnEdge(p)
	// Nearest on right edge (x=25)
	if math.Abs(nearest.X-25) > 1e-9 {
		t.Errorf("Nearest point X should be 25 (right edge), got %v", nearest.X)
	}
}

func TestNearestPointOnEdgeBehind(t *testing.T) {
	d := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	// Point above the dock (behind)
	p := physics.Vec2{X: 0, Y: 40}
	nearest := d.NearestPointOnEdge(p)
	// Should snap to north face (y=35)
	if math.Abs(nearest.Y-35) > 1e-9 {
		t.Errorf("Nearest point Y should be 35 (north face), got %v", nearest.Y)
	}
}

func TestNearestPointOnEdgeLiesOnEdge(t *testing.T) {
	d := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	points := []physics.Vec2{
		{X: 5, Y: 20},
		{X: -30, Y: 32},
		{X: 10, Y: 40},
		{X: -30, Y: 40},
	}
	for _, p := range points {
		nearest := d.NearestPointOnEdge(p)
		onEdge := false
		n := len(d.Vertices)
		for i := range d.Vertices {
			a := d.Vertices[i]
			b := d.Vertices[(i+1)%n]
			snap := nearestOnSegment(a, b, nearest)
			if snap.Sub(nearest).LenSq() < 1e-9 {
				onEdge = true
				break
			}
		}
		if !onEdge {
			t.Errorf("Nearest point %v for input %v does not lie on any edge", nearest, p)
		}
	}
}
