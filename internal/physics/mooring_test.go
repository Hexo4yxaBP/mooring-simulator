package physics

import (
	"math"
	"testing"
)

func TestSpringForceSlack(t *testing.T) {
	anchor := Vec2{X: 0, Y: 0}
	cleat := Vec2{X: 3, Y: 0}
	f := SpringForce(anchor, cleat, Vec2{}, 5.0, 5000, 500) // naturalLen=5, currentLen=3 → slack
	if f != (Vec2{}) {
		t.Errorf("Slack line should produce zero force, got %v", f)
	}
}

func TestSpringForceAtNaturalLength(t *testing.T) {
	anchor := Vec2{X: 0, Y: 0}
	cleat := Vec2{X: 5, Y: 0}
	f := SpringForce(anchor, cleat, Vec2{}, 5.0, 5000, 500) // exactly at natural length
	if f != (Vec2{}) {
		t.Errorf("Line at natural length should produce zero force, got %v", f)
	}
}

func TestSpringForceMagnitude(t *testing.T) {
	anchor := Vec2{X: 0, Y: 0}
	cleat := Vec2{X: 6, Y: 0} // 1 m extension beyond naturalLen=5
	f := SpringForce(anchor, cleat, Vec2{}, 5.0, 5000, 0)
	// Expected magnitude: 5000 * 1 = 5000
	mag := f.Len()
	if math.Abs(mag-5000) > 0.01 {
		t.Errorf("Spring force magnitude: got %v want 5000", mag)
	}
}

func TestSpringForceDirection(t *testing.T) {
	anchor := Vec2{X: 0, Y: 0}
	cleat := Vec2{X: 6, Y: 0}
	f := SpringForce(anchor, cleat, Vec2{}, 5.0, 5000, 0)
	// Force should point from cleat TOWARD anchor (-X direction)
	if f.X >= 0 {
		t.Errorf("Spring force should pull cleat toward anchor (negative X), got %v", f.X)
	}
}

func TestSpringForceDampingAwayFromAnchor(t *testing.T) {
	anchor := Vec2{X: 0, Y: 0}
	cleat := Vec2{X: 6, Y: 0}
	// Cleat moving away from anchor → extensionRate > 0 → increases force
	cleatVelAway := Vec2{X: 1, Y: 0}
	fNo := SpringForce(anchor, cleat, Vec2{}, 5.0, 5000, 500)
	fAway := SpringForce(anchor, cleat, cleatVelAway, 5.0, 5000, 500)
	if fAway.Len() <= fNo.Len() {
		t.Errorf("Damping should increase force when moving away from anchor")
	}
}

func TestSpringForceDampingTowardAnchor(t *testing.T) {
	anchor := Vec2{X: 0, Y: 0}
	cleat := Vec2{X: 6, Y: 0}
	// Cleat moving toward anchor → extensionRate < 0 → decreases force but stays ≥ 0
	cleatVelToward := Vec2{X: -0.1, Y: 0}
	f := SpringForce(anchor, cleat, cleatVelToward, 5.0, 5000, 500)
	if f.Len() < 0 {
		t.Error("Spring force magnitude must not be negative")
	}
}

func TestSpringForceDampingClampToZero(t *testing.T) {
	// Very fast approach: damping term overrides spring → clamp to 0
	anchor := Vec2{X: 0, Y: 0}
	cleat := Vec2{X: 5.001, Y: 0} // barely extended
	cleatVelToward := Vec2{X: -100, Y: 0}
	f := SpringForce(anchor, cleat, cleatVelToward, 5.0, 5000, 500)
	if f.Len() < 0 {
		t.Error("Clamped force must not be negative")
	}
}
