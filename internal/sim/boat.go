package sim

import (
	"math"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

// Boat represents a vessel in the simulation.
type Boat struct {
	ID       int
	Type     BoatType
	Body     physics.RigidBody
	HullLen  float64 // meters, bow-to-stern
	HullBeam float64 // meters, max width

	// Cleat offsets in body frame relative to CoM (bow=+x, port=+y).
	Cleats [3]physics.Vec2 // [CleatBow, CleatMidships, CleatStern]

	// Per-boat controls — only used for active monohull in MVP.
	Throttle ThrottleState
	PropWalk float64 // base lateral force (N); sign = prop handedness
	Rudder   float64 // radians, ±MaxRudderAngle

	IsActive bool

	// Drag and wind coefficients (set from defaults, readable by World.Step).
	LongDrag float64
	LatDrag  float64
	RotDrag  float64
	LongWind float64
	LatWind  float64

	// HullVertices: world-space rectangle corners used for collision.
	// Updated by World.Step each tick.
	HullVertices [4]physics.Vec2
}

// SetRudder sets the rudder angle clamped to ±MaxRudderAngle.
func (b *Boat) SetRudder(angle float64) {
	b.Rudder = math.Max(-MaxRudderAngle, math.Min(MaxRudderAngle, angle))
}

// CleatWorldPos returns the world-space position of the given cleat.
func (b *Boat) CleatWorldPos(id CleatID) physics.Vec2 {
	com := b.Body.ComWorldPos()
	return com.Add(b.Cleats[id].Rotate(b.Body.Heading))
}

// CleatVelocity returns the world-space velocity of the given cleat point.
// v_cleat = v_body + ω × r  where r = cleatPos − comPos.
func (b *Boat) CleatVelocity(id CleatID) physics.Vec2 {
	com := b.Body.ComWorldPos()
	cleat := b.CleatWorldPos(id)
	r := cleat.Sub(com)
	// 2D: ω × r = ω * (-r.Y, r.X)
	omega := b.Body.AngularVel
	tangential := physics.Vec2{X: -r.Y * omega, Y: r.X * omega}
	return b.Body.Velocity.Add(tangential)
}

// updateHullVertices recomputes world-space bounding rectangle corners from current state.
func (b *Boat) updateHullVertices() {
	com := b.Body.ComWorldPos()
	halfLen := b.HullLen / 2
	halfBeam := b.HullBeam / 2
	// Body-frame corners (CCW)
	corners := [4]physics.Vec2{
		{X: halfLen, Y: halfBeam},
		{X: -halfLen, Y: halfBeam},
		{X: -halfLen, Y: -halfBeam},
		{X: halfLen, Y: -halfBeam},
	}
	for i, c := range corners {
		b.HullVertices[i] = com.Add(c.Rotate(b.Body.Heading))
	}
}

// NewMonohull creates a monohull boat with default parameters.
func NewMonohull(id int, pos physics.Vec2, heading float64) *Boat {
	d := MonohullDefaults
	mass := d.Mass
	b := &Boat{
		ID:   id,
		Type: BoatMonohull,
		Body: physics.RigidBody{
			Position:   pos,
			Heading:    heading,
			Mass:       mass,
			InertiaI:   momentOfInertia(mass, d.HullLen, d.HullBeam),
			ComOffset:  physics.Vec2{X: d.ComOffX, Y: 0},
		},
		HullLen:  d.HullLen,
		HullBeam: d.HullBeam,
		PropWalk: d.PropWalk,
		LongDrag: d.LongDrag,
		LatDrag:  d.LatDrag,
		RotDrag:  d.RotDrag,
		LongWind: d.LongWind,
		LatWind:  d.LatWind,
	}
	b.Cleats[CleatBow] = physics.Vec2{X: d.HullLen/2 - 0.5, Y: 0}
	b.Cleats[CleatMidships] = physics.Vec2{X: 0, Y: 0}
	b.Cleats[CleatStern] = physics.Vec2{X: -(d.HullLen/2 - 0.5), Y: 0}
	b.updateHullVertices()
	return b
}

// NewCatamaran creates a passive catamaran with default parameters.
func NewCatamaran(id int, pos physics.Vec2, heading float64) *Boat {
	d := CatamaranDefaults
	mass := d.Mass
	b := &Boat{
		ID:   id,
		Type: BoatCatamaran,
		Body: physics.RigidBody{
			Position:  pos,
			Heading:   heading,
			Mass:      mass,
			InertiaI:  momentOfInertia(mass, d.HullLen, d.HullBeam),
			ComOffset: physics.Vec2{X: d.ComOffX, Y: 0},
		},
		HullLen:  d.HullLen,
		HullBeam: d.HullBeam,
		PropWalk: 0, // MVP: no engine
		LongDrag: d.LongDrag,
		LatDrag:  d.LatDrag,
		RotDrag:  d.RotDrag,
		LongWind: d.LongWind,
		LatWind:  d.LatWind,
	}
	b.Cleats[CleatBow] = physics.Vec2{X: d.HullLen/2 - 0.5, Y: 0}
	b.Cleats[CleatMidships] = physics.Vec2{X: 0, Y: 0}
	b.Cleats[CleatStern] = physics.Vec2{X: -(d.HullLen/2 - 0.5), Y: 0}
	b.updateHullVertices()
	return b
}
