package sim

import (
	"errors"
	"math"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

// MooringPointKind categorises a pre-defined mooring attachment point.
type MooringPointKind int

const (
	MooringPierCleat MooringPointKind = iota // fixed cleat on the pier face
	MooringBuoy                              // free-standing buoy in the water
)

// MooringPoint is a pre-defined location in the world where mooring lines can be attached.
type MooringPoint struct {
	ID   int
	Pos  physics.Vec2
	Kind MooringPointKind
}

// World holds all simulation entities and advances the physics each tick.
// All input validation for commands is performed by the caller (main.go) before
// mutating world state — the sim package trusts that incoming values are valid.
type World struct {
	Boats         []*Boat
	Dock          Dock
	Lines         []MooringLine
	Wind          WindField
	MooringPoints []MooringPoint
	Time          float64 // seconds since simulation start
}

// Step advances the simulation by dt seconds.
// Forces applied per boat: wind, engine (active monohull only), hydro drag,
// mooring line springs, collision penalties. Then integrates each body.
func (w *World) Step(dt float64) {
	if dt <= 0 {
		return
	}

	// Update hull vertices from current poses before collision detection
	for _, b := range w.Boats {
		b.updateHullVertices()
	}

	dockVerts := w.Dock.Vertices

	for _, b := range w.Boats {
		var acc physics.ForceAccumulator
		com := b.Body.ComWorldPos()
		windVel := w.Wind.Velocity()

		// 1. Wind drag
		wf := physics.WindForce(windVel, b.Body.Heading, b.LongWind, b.LatWind)
		acc.Apply(wf, com, com) // applied at CoM — no additional torque

		// 2. Engine forces (active monohull only; catamaran has no engine in MVP)
		if b.IsActive && b.Type == BoatMonohull {
			thrustN := ThrustTable[b.Throttle]
			if thrustN != 0 {
				stern := b.CleatWorldPos(CleatStern)
				tf := physics.ThrustForce(b.Body.Heading, thrustN)
				acc.Apply(tf, stern, com)
			}

			propN := b.PropWalk * PropWalkMultiplier[b.Throttle]
			if propN != 0 {
				stern := b.CleatWorldPos(CleatStern)
				pf := physics.PropWalkForce(b.Body.Heading, propN)
				acc.Apply(pf, stern, com)
			}

			if b.Rudder != 0 {
				bowVec := physics.Vec2{X: math.Cos(b.Body.Heading), Y: math.Sin(b.Body.Heading)}
				bodyVelFwd := b.Body.Velocity.Dot(bowVec)
				stern := b.CleatWorldPos(CleatStern)
				rf := physics.RudderForce(b.Body.Heading, b.Rudder, bodyVelFwd, RudderCoeff)
				acc.Apply(rf, stern, com)
			}
		}

		// 3. Hydrodynamic drag
		dragF, dragT := physics.HydroDragForce(
			b.Body.Velocity, b.Body.Heading, b.Body.AngularVel,
			b.LongDrag, b.LatDrag, b.RotDrag,
		)
		acc.Apply(dragF, com, com)
		acc.Torque += dragT

		// 4. Mooring line spring forces
		for _, line := range w.Lines {
			if line.BoatID != b.ID {
				continue
			}
			cleatPos := b.CleatWorldPos(line.Cleat)
			cleatVel := b.CleatVelocity(line.Cleat)
			sf := physics.SpringForce(line.DockPoint, cleatPos, cleatVel,
				line.NaturalLength, line.Stiffness, line.Damping)
			acc.Apply(sf, cleatPos, com)
		}

		// 5. Collision: boat ↔ dock
		if len(dockVerts) >= 3 {
			hv := b.HullVertices[:]
			cf, cp := physics.CollisionPenalty(hv, dockVerts, CollisionPenaltyStiffness)
			if cf != (physics.Vec2{}) {
				acc.Apply(cf, cp, com)
			}
		}

		// 6. Collision: boat ↔ other boats
		for _, other := range w.Boats {
			if other.ID == b.ID {
				continue
			}
			hv := b.HullVertices[:]
			ov := other.HullVertices[:]
			cf, cp := physics.CollisionPenalty(hv, ov, CollisionPenaltyStiffness)
			if cf != (physics.Vec2{}) {
				acc.Apply(cf, cp, com)
			}
		}

		physics.Integrate(&b.Body, acc, dt)
	}

	w.Time += dt
}

// ActiveBoat returns the currently active boat or nil if none is active.
func (w *World) ActiveBoat() *Boat {
	for _, b := range w.Boats {
		if b.IsActive {
			return b
		}
	}
	return nil
}

// SetActiveBoat marks the boat with the given ID as active and all others inactive.
// No-op if no boat has that ID.
func (w *World) SetActiveBoat(id int) {
	found := false
	for _, b := range w.Boats {
		if b.ID == id {
			found = true
			break
		}
	}
	if !found {
		return
	}
	for _, b := range w.Boats {
		b.IsActive = b.ID == id
	}
}

// AddMooringLine validates and appends a mooring line.
// Returns error if the boat ID is unknown or the cleat ID is invalid.
func (w *World) AddMooringLine(line MooringLine) error {
	if findBoat(w.Boats, line.BoatID) == nil {
		return errors.New("sim: unknown boat ID")
	}
	if !IsValidCleat(line.Cleat) {
		return errors.New("sim: invalid cleat ID")
	}
	w.Lines = append(w.Lines, line)
	return nil
}

// RemoveMooringLine removes the line at index idx. No-op if idx is out of range.
func (w *World) RemoveMooringLine(idx int) {
	if idx < 0 || idx >= len(w.Lines) {
		return
	}
	w.Lines = append(w.Lines[:idx], w.Lines[idx+1:]...)
}

// WindField holds global wind parameters.
type WindField struct {
	Speed     float64 // m/s, ∈ [0, 30]
	Direction float64 // radians, wind-FROM direction (0 = from +x/east)
}

// Velocity returns the wind velocity vector in world space.
// Wind blows FROM direction → velocity points toward (direction + π).
func (wf WindField) Velocity() physics.Vec2 {
	return physics.Vec2{
		X: -wf.Speed * math.Cos(wf.Direction),
		Y: -wf.Speed * math.Sin(wf.Direction),
	}
}
