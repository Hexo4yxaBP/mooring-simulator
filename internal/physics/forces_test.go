package physics

import (
	"math"
	"testing"
)

func TestWindForceBeamVsBow(t *testing.T) {
	speed := 10.0
	heading := 0.0
	longC, latC := 15.0, 80.0

	// Beam-on: wind from +y (south), wind velocity = (0, +speed)
	// In body frame: apparent = (0, +speed), force_lat = latC * speed
	beamWind := Vec2{X: 0, Y: speed}
	fb := WindForce(beamWind, heading, longC, latC)

	// Bow-on: wind from -x (west), wind velocity = (-speed, 0)
	// In body frame: apparent = (-speed, 0), force_long = longC * speed
	bowWind := Vec2{X: -speed, Y: 0}
	fw := WindForce(bowWind, heading, longC, latC)

	latMag := math.Abs(fb.Y)
	longMag := math.Abs(fw.X)
	if latMag <= longMag {
		t.Errorf("Beam-on force %v should exceed bow-on force %v (latC=%v > longC=%v)", latMag, longMag, latC, longC)
	}
}

func TestWindForceZeroWind(t *testing.T) {
	f := WindForce(Vec2{}, 0, 15, 80)
	if f != (Vec2{}) {
		t.Errorf("Zero wind should produce zero force, got %v", f)
	}
}

func TestThrustForceForward(t *testing.T) {
	f := ThrustForce(0, 100)
	if f.X <= 0 {
		t.Errorf("Forward thrust heading=0 should have positive X, got %v", f)
	}
	if math.Abs(f.Y) > 1e-9 {
		t.Errorf("Forward thrust heading=0 should have zero Y, got %v", f.Y)
	}
}

func TestThrustForceAstern(t *testing.T) {
	f := ThrustForce(0, -100)
	if f.X >= 0 {
		t.Errorf("Astern thrust heading=0 should have negative X, got %v", f)
	}
}

func TestThrustForceHeading90(t *testing.T) {
	f := ThrustForce(math.Pi/2, 100)
	if math.Abs(f.X) > 1e-9 {
		t.Errorf("Thrust heading=Pi/2 should have zero X, got %v", f.X)
	}
	if f.Y <= 0 {
		t.Errorf("Thrust heading=Pi/2 should have positive Y, got %v", f.Y)
	}
}

func TestPropWalkPortWalk(t *testing.T) {
	f := PropWalkForce(0, 300)
	// heading=0: port is +y direction
	if f.Y <= 0 {
		t.Errorf("Positive walkN with heading=0 should push to port (+Y), got %v", f)
	}
	if math.Abs(f.X) > 1e-9 {
		t.Errorf("PropWalk heading=0 should have zero X, got %v", f.X)
	}
}

func TestPropWalkZero(t *testing.T) {
	f := PropWalkForce(0, 0)
	if f != (Vec2{}) {
		t.Errorf("Zero walk should produce zero force, got %v", f)
	}
}

func TestRudderForceZeroSpeed(t *testing.T) {
	f := RudderForce(0, 0.3, 0.04, 5000) // bodyVelFwd=0.04 < 0.05 threshold
	if f != (Vec2{}) {
		t.Errorf("RudderForce below speed threshold should be zero, got %v", f)
	}
}

func TestRudderForceStarboard(t *testing.T) {
	// Positive rudder angle, positive fwd speed, heading=0
	// Starboard unit = (sin(0), -cos(0)) = (0, -1)
	// Force = coeff * rudderAngle * bodyVelFwd * (0,-1) → negative Y = starboard component
	f := RudderForce(0, 0.3, 2.0, 5000)
	// Starboard direction at heading=0 is -y (south)
	if f.Y >= 0 {
		t.Errorf("RudderForce stbd rudder fwd speed should have negative Y, got %v", f)
	}
}

func TestRudderForceExactThreshold(t *testing.T) {
	f := RudderForce(0, 0.3, 0.05, 5000) // exactly at threshold — should produce force
	if f == (Vec2{}) {
		t.Error("RudderForce at exactly 0.05 m/s should produce force")
	}
}

func TestHydroDragOpposesVelocity(t *testing.T) {
	fwd := Vec2{2, 0}
	drag, _ := HydroDragForce(fwd, 0, 0, 300, 6000, 50000)
	if drag.X >= 0 {
		t.Errorf("Drag should oppose forward velocity (negative X), got %v", drag.X)
	}
}

func TestHydroDragLateralRatio(t *testing.T) {
	v := 1.0
	fwdC, latC := 300.0, 6000.0
	// Forward velocity
	fDrag, _ := HydroDragForce(Vec2{X: v, Y: 0}, 0, 0, fwdC, latC, 0)
	// Lateral (beam-on) velocity at same speed
	lDrag, _ := HydroDragForce(Vec2{X: 0, Y: v}, 0, 0, fwdC, latC, 0)

	fMag := math.Abs(fDrag.X)
	lMag := math.Abs(lDrag.Y)
	ratio := lMag / fMag
	if math.Abs(ratio-20) > 0.01 {
		t.Errorf("Lateral/forward drag ratio: got %v want 20", ratio)
	}
}

func TestHydroDragTorqueOpposesRotation(t *testing.T) {
	_, torque := HydroDragForce(Vec2{}, 0, 1.0, 0, 0, 50000)
	if torque >= 0 {
		t.Errorf("Torque drag should oppose CCW rotation (negative), got %v", torque)
	}
}

func TestHydroDragZeroInputs(t *testing.T) {
	f, t2 := HydroDragForce(Vec2{}, 0, 0, 300, 6000, 50000)
	if f != (Vec2{}) || t2 != 0 {
		t.Errorf("Zero inputs should produce zero outputs, got f=%v t=%v", f, t2)
	}
}
