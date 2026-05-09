package physics

// SpringForce computes the mooring line tension force on the cleat.
// Constraint P1: tension-only — returns zero if the line is slack (currentLen ≤ naturalLen).
// anchorPos: world-space dock attachment point.
// cleatPos: world-space boat cleat position.
// cleatVel: world-space velocity of the cleat point.
// Returns a force vector pointing from cleatPos toward anchorPos.
func SpringForce(anchorPos, cleatPos, cleatVel Vec2, naturalLen, stiffness, damping float64) Vec2 {
	diff := cleatPos.Sub(anchorPos)
	currentLen := diff.Len()
	if currentLen <= naturalLen {
		return Vec2{}
	}
	extension := currentLen - naturalLen

	lineDir := diff.Scale(1.0 / currentLen) // unit vector from anchor toward cleat
	extensionRate := cleatVel.Dot(lineDir)  // positive = cleat moving away from anchor

	fMag := stiffness*extension + damping*extensionRate
	if fMag < 0 {
		fMag = 0 // clamp: damping cannot create compression
	}

	return lineDir.Scale(-fMag) // force pulls cleat toward anchor
}
