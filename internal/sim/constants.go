package sim

import "math"

// ThrustTable maps ThrottleState → engine thrust in Newtons (positive = forward).
var ThrustTable = [5]float64{0, 800, 8000, -600, -5000}

// PropWalkMultiplier maps ThrottleState → fraction of boat.PropWalk base force.
// Constraint P4: astern multipliers >> forward multipliers.
var PropWalkMultiplier = [5]float64{0.0, 0.15, 0.25, 0.8, 1.0}

// MaxRudderAngle is the maximum rudder deflection in radians (±35°).
const MaxRudderAngle = 35.0 * math.Pi / 180.0

// RudderStepPerTick is the rudder angle change per Update() tick when A/D held.
const RudderStepPerTick = 2.0 * math.Pi / 180.0

// RudderCoeff is the rudder lift force coefficient (N per rad per m/s).
const RudderCoeff = 800.0

// boatDefaults holds default parameters for each boat type.
type boatDefaults struct {
	Mass      float64
	HullLen   float64
	HullBeam  float64
	ComOffX   float64 // body-frame x offset of CoM from geometric center
	PropWalk  float64
	LongDrag  float64
	LatDrag   float64
	RotDrag   float64
	LongWind  float64
	LatWind   float64
}

// MonohullDefaults holds the default parameters for a monohull sailboat.
// Drag values are tuned for gameplay: LongDrag/Mass time-constant ≈ 2.8 s so
// a 2 m/s coast decays below 0.1 m/s within 10 s (TASK-075 acceptance criterion).
// LatDrag maintains the 20:1 lateral-to-longitudinal ratio (constraint P3).
var MonohullDefaults = boatDefaults{
	Mass:     5000,
	HullLen:  10,
	HullBeam: 3,
	ComOffX:  -0.5, // CoM slightly aft of geometric center
	PropWalk: 300,
	LongDrag: 1800,
	LatDrag:  36000,
	RotDrag:  50000,
	LongWind: 15,
	LatWind:  80,
}

// CatamaranDefaults holds the default parameters for a passive catamaran.
var CatamaranDefaults = boatDefaults{
	Mass:     8000,
	HullLen:  12,
	HullBeam: 6,
	ComOffX:  0,
	PropWalk: 0, // MVP: no engine
	LongDrag: 400,
	LatDrag:  8000,
	RotDrag:  80000,
	LongWind: 20,
	LatWind:  120,
}

// DefaultMooringStiffness is the default spring constant for new mooring lines (N/m).
// Tuned so that a single stern line in 10 m/s beam wind holds the boat within 2 m
// lateral displacement (TASK-075 acceptance criterion).
const DefaultMooringStiffness = 8000.0

// DefaultMooringDamping is the default damping for new mooring lines (N·s/m).
const DefaultMooringDamping = 500.0

// CollisionPenaltyStiffness is the penalty spring constant for body contacts (N/m).
const CollisionPenaltyStiffness = 100000.0

// momentOfInertia computes I for a uniform rectangular body.
func momentOfInertia(mass, length, beam float64) float64 {
	return mass * (length*length + beam*beam) / 12.0
}
