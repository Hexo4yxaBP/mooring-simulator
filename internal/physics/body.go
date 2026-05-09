package physics

// RigidBody holds the kinematic and inertial state of a 2D rigid body.
// Position is the geometric center (not CoM). CoM = Position + ComOffset.Rotate(Heading).
type RigidBody struct {
	Position   Vec2
	Velocity   Vec2
	Heading    float64 // radians, CCW from +x axis
	AngularVel float64 // rad/s, CCW positive
	Mass       float64 // kg, must be > 0
	InertiaI   float64 // kg·m², about CoM, must be > 0
	ComOffset  Vec2    // CoM offset relative to geometric center, body-frame
}

// ComWorldPos returns the world-space position of the center of mass.
func (b *RigidBody) ComWorldPos() Vec2 {
	return b.Position.Add(b.ComOffset.Rotate(b.Heading))
}

// ForceAccumulator accumulates forces and torques for one simulation step.
type ForceAccumulator struct {
	Force  Vec2    // world-space, Newtons
	Torque float64 // N·m, CCW positive
}

// Apply adds force applied at worldPoint. The torque contribution is cross(r, force)
// where r = worldPoint − comWorldPos.
func (a *ForceAccumulator) Apply(force Vec2, worldPoint Vec2, comWorldPos Vec2) {
	a.Force = a.Force.Add(force)
	r := worldPoint.Sub(comWorldPos)
	a.Torque += r.Cross(force)
}

// Reset clears accumulated forces and torques.
func (a *ForceAccumulator) Reset() {
	a.Force = Vec2{}
	a.Torque = 0
}

// Integrate advances body one timestep using semi-implicit Euler.
// Velocity is updated before position to improve stability with spring forces.
func Integrate(body *RigidBody, acc ForceAccumulator, dt float64) {
	if dt == 0 {
		return
	}
	linAcc := acc.Force.Scale(1.0 / body.Mass)
	angAcc := acc.Torque / body.InertiaI

	body.Velocity = body.Velocity.Add(linAcc.Scale(dt))
	body.AngularVel += angAcc * dt

	body.Position = body.Position.Add(body.Velocity.Scale(dt))
	body.Heading += body.AngularVel * dt
}
