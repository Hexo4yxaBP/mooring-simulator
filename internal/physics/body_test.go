package physics

import (
	"math"
	"testing"
)

func TestForceAccumulatorApplyAtCoM(t *testing.T) {
	var acc ForceAccumulator
	com := Vec2{X: 1, Y: 1}
	force := Vec2{X: 5, Y: 0}
	acc.Apply(force, com, com) // applied AT CoM
	if acc.Force != force {
		t.Errorf("Force: got %v want %v", acc.Force, force)
	}
	if acc.Torque != 0 {
		t.Errorf("Torque at CoM should be 0, got %v", acc.Torque)
	}
}

func TestForceAccumulatorApplyOffset(t *testing.T) {
	var acc ForceAccumulator
	com := Vec2{X: 0, Y: 0}
	// Force +Y applied at point (1, 0) — should produce positive torque (CCW)
	acc.Apply(Vec2{X: 0, Y: 1}, Vec2{X: 1, Y: 0}, com)
	if acc.Torque <= 0 {
		t.Errorf("Expected positive torque, got %v", acc.Torque)
	}
}

func TestForceAccumulatorApplyNegativeTorque(t *testing.T) {
	var acc ForceAccumulator
	com := Vec2{X: 0, Y: 0}
	// Force -Y applied at point (1, 0) — negative torque (CW)
	acc.Apply(Vec2{X: 0, Y: -1}, Vec2{X: 1, Y: 0}, com)
	if acc.Torque >= 0 {
		t.Errorf("Expected negative torque, got %v", acc.Torque)
	}
}

func TestForceAccumulatorReset(t *testing.T) {
	acc := ForceAccumulator{Force: Vec2{X: 1, Y: 2}, Torque: 5}
	acc.Reset()
	if acc.Force != (Vec2{}) || acc.Torque != 0 {
		t.Error("Reset did not clear accumulator")
	}
}

func TestIntegrateForwardForce(t *testing.T) {
	body := RigidBody{Mass: 1, InertiaI: 1}
	acc := ForceAccumulator{Force: Vec2{X: 10, Y: 0}}
	Integrate(&body, acc, 1.0)
	// v += (10/1) * 1 = 10; p += 10 * 1 = 10
	if body.Velocity.X != 10 {
		t.Errorf("Velocity.X: got %v want 10", body.Velocity.X)
	}
	if body.Position.X != 10 {
		t.Errorf("Position.X: got %v want 10", body.Position.X)
	}
}

func TestIntegrateSemiImplicit(t *testing.T) {
	// Semi-implicit: velocity updated before position.
	// With initial vel=0, force=10, mass=1, dt=1:
	// v_new = 0 + 10*1 = 10; p_new = 0 + 10*1 = 10 (uses new velocity)
	body := RigidBody{Mass: 1, InertiaI: 1}
	acc := ForceAccumulator{Force: Vec2{X: 10, Y: 0}}
	Integrate(&body, acc, 1.0)
	if body.Position.X != 10 {
		t.Errorf("semi-implicit: position should use updated velocity, got %v", body.Position.X)
	}
}

func TestIntegrateCoasting(t *testing.T) {
	body := RigidBody{Mass: 1, InertiaI: 1, Velocity: Vec2{X: 3, Y: 0}}
	var acc ForceAccumulator
	Integrate(&body, acc, 1.0)
	if body.Velocity.X != 3 {
		t.Errorf("Velocity unchanged: got %v want 3", body.Velocity.X)
	}
	if body.Position.X != 3 {
		t.Errorf("Position advanced by v*dt: got %v want 3", body.Position.X)
	}
}

func TestIntegrateZeroDtNoMutation(t *testing.T) {
	body := RigidBody{Mass: 1, InertiaI: 1, Velocity: Vec2{X: 5, Y: 0}}
	orig := body
	Integrate(&body, ForceAccumulator{Force: Vec2{X: 10, Y: 0}}, 0)
	if body != orig {
		t.Error("dt=0 should not mutate body")
	}
}

func TestIntegrateRotation(t *testing.T) {
	body := RigidBody{Mass: 1, InertiaI: 1, AngularVel: 1.0}
	acc := ForceAccumulator{Torque: -2.0} // restoring torque opposing CCW rotation
	Integrate(&body, acc, 1.0)
	// angAcc = -2/1 = -2; ω_new = 1 + (-2)*1 = -1; heading = 0 + (-1)*1 = -1
	if body.AngularVel >= 1.0 {
		t.Errorf("AngularVel should decrease, got %v", body.AngularVel)
	}
}

func TestComWorldPos(t *testing.T) {
	body := RigidBody{
		Position:  Vec2{X: 0, Y: 0},
		Heading:   0,
		ComOffset: Vec2{X: 1, Y: 0},
	}
	got := body.ComWorldPos()
	want := Vec2{X: 1, Y: 0}
	if math.Abs(got.X-want.X) > 1e-9 || math.Abs(got.Y-want.Y) > 1e-9 {
		t.Errorf("ComWorldPos heading=0: got %v want %v", got, want)
	}

	body.Heading = math.Pi / 2
	got = body.ComWorldPos()
	// ComOffset (1,0) rotated by Pi/2 = (0, 1)
	if math.Abs(got.X-0) > 1e-9 || math.Abs(got.Y-1) > 1e-9 {
		t.Errorf("ComWorldPos heading=Pi/2: got %v want {0,1}", got)
	}
}
