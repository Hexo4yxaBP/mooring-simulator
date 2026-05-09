package physics

import (
	"math"
	"testing"
)

func square(cx, cy, half float64) []Vec2 {
	return []Vec2{
		{X: cx + half, Y: cy + half},
		{X: cx - half, Y: cy + half},
		{X: cx - half, Y: cy - half},
		{X: cx + half, Y: cy - half},
	}
}

func TestCollisionNoOverlap(t *testing.T) {
	a := square(0, 0, 1)
	b := square(5, 0, 1) // gap of 3 between them
	f, _ := CollisionPenalty(a, b, 1000)
	if f != (Vec2{}) {
		t.Errorf("Non-overlapping polygons should produce zero force, got %v", f)
	}
}

func TestCollisionPartialOverlapX(t *testing.T) {
	a := square(0, 0, 1) // x: [-1,1]
	b := square(1.5, 0, 1) // x: [0.5, 2.5] → overlap 0.5
	f, _ := CollisionPenalty(a, b, 1000)
	if f == (Vec2{}) {
		t.Error("Overlapping polygons should produce non-zero force")
	}
	// Force should push A in -X direction (away from B)
	if f.X >= 0 {
		t.Errorf("A should be pushed in -X direction, got force %v", f)
	}
	// Magnitude ≈ stiffness * overlap = 1000 * 0.5 = 500
	if math.Abs(f.Len()-500) > 10 {
		t.Errorf("Force magnitude: got %v want ~500", f.Len())
	}
}

func TestCollisionPartialOverlapY(t *testing.T) {
	a := square(0, 0, 1)
	b := square(0, 1.5, 1) // y: [0.5, 2.5] → overlap 0.5 on Y axis
	f, _ := CollisionPenalty(a, b, 1000)
	if f == (Vec2{}) {
		t.Error("Y-axis overlap should produce non-zero force")
	}
	// Force should push A in -Y direction
	if f.Y >= 0 {
		t.Errorf("A should be pushed in -Y direction, got force %v", f)
	}
}

func TestCollisionFullOverlap(t *testing.T) {
	a := square(0, 0, 1)
	b := square(0, 0, 1) // identical squares
	f, _ := CollisionPenalty(a, b, 1000)
	if f == (Vec2{}) {
		t.Error("Fully overlapping polygons should produce non-zero force")
	}
}

func TestCollisionDegenerate(t *testing.T) {
	single := []Vec2{{X: 0, Y: 0}} // single vertex
	b := square(0, 0, 1)
	// Should not panic
	f, cp := CollisionPenalty(single, b, 1000)
	_ = f
	_ = cp
}

func TestCollisionRotatedRectangles(t *testing.T) {
	// A: square at origin
	a := square(0, 0, 1)
	// B: square rotated 45° centered at (1.2, 0)
	half := 1.0
	cx, cy := 1.2, 0.0
	angle := math.Pi / 4
	cos45, sin45 := math.Cos(angle), math.Sin(angle)
	corners := []Vec2{
		{X: half, Y: half}, {X: -half, Y: half}, {X: -half, Y: -half}, {X: half, Y: -half},
	}
	b := make([]Vec2, 4)
	for i, c := range corners {
		rx := c.X*cos45 - c.Y*sin45 + cx
		ry := c.X*sin45 + c.Y*cos45 + cy
		b[i] = Vec2{X: rx, Y: ry}
	}
	// These may or may not overlap — just verify no panic
	f, _ := CollisionPenalty(a, b, 1000)
	_ = f
}

func TestCollisionEdgeTouching(t *testing.T) {
	a := square(0, 0, 1) // x: [-1,1]
	b := square(2, 0, 1) // x: [1,3] → touching at x=1, overlap=0
	f, _ := CollisionPenalty(a, b, 1000)
	// Edge-touching (overlap=0) → no force
	if f.Len() > 1e-9 {
		t.Errorf("Edge-touching should produce no force, got %v", f)
	}
}

func TestCollisionRepeatStability(t *testing.T) {
	a := square(0, 0, 1)
	b := square(1.5, 0, 1)
	for i := 0; i < 100; i++ {
		f, _ := CollisionPenalty(a, b, 1000)
		if math.IsNaN(f.X) || math.IsNaN(f.Y) {
			t.Fatal("NaN in collision force")
		}
	}
}
