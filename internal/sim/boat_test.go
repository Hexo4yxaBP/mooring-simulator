package sim

import (
	"math"
	"testing"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

func TestNewMonohullDefaults(t *testing.T) {
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	if b.Body.Mass != MonohullDefaults.Mass {
		t.Errorf("Mass: got %v want %v", b.Body.Mass, MonohullDefaults.Mass)
	}
	if b.HullLen != MonohullDefaults.HullLen {
		t.Errorf("HullLen: got %v want %v", b.HullLen, MonohullDefaults.HullLen)
	}
	if b.PropWalk != MonohullDefaults.PropWalk {
		t.Errorf("PropWalk: got %v want %v", b.PropWalk, MonohullDefaults.PropWalk)
	}
}

func TestNewCatamaranPassive(t *testing.T) {
	c := NewCatamaran(2, physics.Vec2{X: 0, Y: 0}, 0)
	if c.PropWalk != 0 {
		t.Errorf("Catamaran MVP PropWalk should be 0, got %v", c.PropWalk)
	}
	if c.IsActive {
		t.Error("Catamaran should not be active on creation")
	}
}

func TestCleatWorldPosHeading0(t *testing.T) {
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	bow := b.CleatWorldPos(CleatBow)
	stern := b.CleatWorldPos(CleatStern)
	if bow.X <= 0 {
		t.Errorf("Bow cleat should be forward (+X) at heading=0, got %v", bow.X)
	}
	if stern.X >= 0 {
		t.Errorf("Stern cleat should be aft (-X) at heading=0, got %v", stern.X)
	}
}

func TestCleatWorldPosHeadingPiOver2(t *testing.T) {
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, math.Pi/2)
	bow := b.CleatWorldPos(CleatBow)
	// Heading Pi/2: bow is in +Y direction
	if bow.Y <= 0 {
		t.Errorf("Bow cleat heading=Pi/2 should have positive Y, got %v", bow.Y)
	}
}

func TestCleatWorldPosHeadingPi(t *testing.T) {
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, math.Pi)
	bow := b.CleatWorldPos(CleatBow)
	// Heading Pi: bow is in -X direction
	if bow.X >= 0 {
		t.Errorf("Bow cleat heading=Pi should have negative X, got %v", bow.X)
	}
}

func TestCleatVelocityNoRotation(t *testing.T) {
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	b.Body.Velocity = physics.Vec2{X: 2, Y: 0}
	b.Body.AngularVel = 0
	vel := b.CleatVelocity(CleatBow)
	// No rotation: cleat velocity = body velocity
	if math.Abs(vel.X-2) > 1e-9 || math.Abs(vel.Y) > 1e-9 {
		t.Errorf("CleatVelocity no rotation: got %v want {2,0}", vel)
	}
}

func TestCleatVelocityWithRotation(t *testing.T) {
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	b.Body.Velocity = physics.Vec2{X: 0, Y: 0}
	b.Body.AngularVel = 1.0 // CCW
	// Bow cleat at (+4.5, 0) from CoM: ω × r = 1 * (-0, +4.5) = (0, 4.5)
	vel := b.CleatVelocity(CleatBow)
	if vel.Y <= 0 {
		t.Errorf("Bow cleat with CCW rotation should have positive Y velocity, got %v", vel)
	}
}

func TestSetRudderClamping(t *testing.T) {
	b := NewMonohull(1, physics.Vec2{}, 0)
	b.SetRudder(10) // far beyond MaxRudderAngle
	if b.Rudder != MaxRudderAngle {
		t.Errorf("Rudder clamped to +Max: got %v want %v", b.Rudder, MaxRudderAngle)
	}
	b.SetRudder(-10)
	if b.Rudder != -MaxRudderAngle {
		t.Errorf("Rudder clamped to -Max: got %v want %v", b.Rudder, -MaxRudderAngle)
	}
}
