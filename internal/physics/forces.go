package physics

import "math"

// WindForce computes wind drag force on a hull in world space.
// windVel is the wind velocity vector (world-space, m/s).
// longCoeff and latCoeff are drag area coefficients (m²) for bow-on and beam-on exposure.
// Returns a world-space force vector. Applied at hull geometric center (no torque).
func WindForce(windVel Vec2, heading float64, longCoeff, latCoeff float64) Vec2 {
	if windVel.LenSq() == 0 {
		return Vec2{}
	}
	apparent := windVel.Rotate(-heading)
	fb := Vec2{X: longCoeff * apparent.X, Y: latCoeff * apparent.Y}
	return fb.Rotate(heading)
}

// ThrustForce returns the engine thrust vector in world space.
// thrustN is signed: positive = forward, negative = astern.
func ThrustForce(heading float64, thrustN float64) Vec2 {
	return Vec2{X: math.Cos(heading), Y: math.Sin(heading)}.Scale(thrustN)
}

// PropWalkForce returns the lateral prop-walk force in world space.
// walkN is signed: positive = port-side push, negative = starboard-side push.
// Port direction = heading + π/2.
func PropWalkForce(heading float64, walkN float64) Vec2 {
	// Port unit vector: 90° CCW from bow
	return Vec2{X: -math.Sin(heading), Y: math.Cos(heading)}.Scale(walkN)
}

// RudderForce returns the rudder lateral force in world space.
// rudderAngle: positive = starboard deflection.
// bodyVelFwd: speed along bow axis (m/s). Returns zero if |bodyVelFwd| < 0.05 (constraint P5).
// Positive rudderAngle + positive speed → starboard component.
func RudderForce(heading float64, rudderAngle float64, bodyVelFwd float64, coeff float64) Vec2 {
	if math.Abs(bodyVelFwd) < 0.05 {
		return Vec2{}
	}
	// Starboard unit vector: 90° CW from bow = heading - π/2
	stbd := Vec2{X: math.Sin(heading), Y: -math.Cos(heading)}
	return stbd.Scale(coeff * rudderAngle * bodyVelFwd)
}

// HydroDragForce returns hydrodynamic drag forces opposing motion.
// fwdCoeff: longitudinal drag (N·s/m), latCoeff: lateral drag (N·s/m), rotCoeff: rotational drag (N·m·s/rad).
// Returns (linearDrag Vec2, torqueDrag float64). Apply linear drag at CoM; add torqueDrag directly.
// Constraint P3: latCoeff >> fwdCoeff for a keel boat.
func HydroDragForce(vel Vec2, heading float64, angularVel float64, fwdCoeff, latCoeff, rotCoeff float64) (Vec2, float64) {
	velBody := vel.Rotate(-heading)
	dragBody := Vec2{X: -fwdCoeff * velBody.X, Y: -latCoeff * velBody.Y}
	torqueDrag := -rotCoeff * angularVel
	return dragBody.Rotate(heading), torqueDrag
}
