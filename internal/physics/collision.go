package physics

import "math"

// CollisionPenalty detects overlap between two convex polygons using the
// Separating Axis Theorem (SAT) and returns a penalty spring force that pushes
// polygon A out of polygon B, plus the approximate contact point.
//
// Returns (Vec2{}, Vec2{}) if there is no overlap or if either polygon has fewer
// than 2 vertices (degenerate, no panic).
func CollisionPenalty(verticesA, verticesB []Vec2, penaltyStiffness float64) (Vec2, Vec2) {
	if len(verticesA) < 2 || len(verticesB) < 2 {
		return Vec2{}, Vec2{}
	}

	minOverlap := math.MaxFloat64
	minAxis := Vec2{}

	// Test axes from edges of A
	for i := range verticesA {
		axis, ok := edgeNormal(verticesA[i], verticesA[(i+1)%len(verticesA)])
		if !ok {
			continue
		}
		overlap := penetrationOnAxis(verticesA, verticesB, axis)
		if overlap <= 0 {
			return Vec2{}, Vec2{} // separating axis found
		}
		if overlap < minOverlap {
			minOverlap = overlap
			minAxis = axis
		}
	}

	// Test axes from edges of B
	for i := range verticesB {
		axis, ok := edgeNormal(verticesB[i], verticesB[(i+1)%len(verticesB)])
		if !ok {
			continue
		}
		overlap := penetrationOnAxis(verticesA, verticesB, axis)
		if overlap <= 0 {
			return Vec2{}, Vec2{}
		}
		if overlap < minOverlap {
			minOverlap = overlap
			minAxis = axis
		}
	}

	// Ensure axis points from B toward A (pushes A away from B)
	cA := centroid(verticesA)
	cB := centroid(verticesB)
	if minAxis.Dot(cA.Sub(cB)) < 0 {
		minAxis = minAxis.Scale(-1)
	}

	force := minAxis.Scale(penaltyStiffness * minOverlap)
	contact := midpointBetween(verticesA, verticesB)
	return force, contact
}

// edgeNormal returns the outward-facing unit normal for edge (a→b).
// Returns (zero, false) if the edge is degenerate (zero length).
func edgeNormal(a, b Vec2) (Vec2, bool) {
	edge := b.Sub(a)
	if edge.LenSq() < 1e-18 {
		return Vec2{}, false
	}
	return Vec2{X: -edge.Y, Y: edge.X}.Normalize(), true
}

// projectPolygon returns the min and max projection of a polygon onto axis.
func projectPolygon(verts []Vec2, axis Vec2) (min, max float64) {
	min = verts[0].Dot(axis)
	max = min
	for _, v := range verts[1:] {
		p := v.Dot(axis)
		if p < min {
			min = p
		}
		if p > max {
			max = p
		}
	}
	return
}

// penetrationOnAxis returns the overlap of two polygons projected onto axis.
// Returns ≤ 0 if they are separated on this axis.
func penetrationOnAxis(a, b []Vec2, axis Vec2) float64 {
	minA, maxA := projectPolygon(a, axis)
	minB, maxB := projectPolygon(b, axis)
	// Overlap = min of the two possible overlaps
	o1 := maxA - minB
	o2 := maxB - minA
	if o1 <= 0 || o2 <= 0 {
		return 0
	}
	if o1 < o2 {
		return o1
	}
	return o2
}

// centroid returns the average of polygon vertices.
func centroid(verts []Vec2) Vec2 {
	c := Vec2{}
	for _, v := range verts {
		c = c.Add(v)
	}
	return c.Scale(1.0 / float64(len(verts)))
}

// midpointBetween returns the midpoint between the two polygon centroids.
func midpointBetween(a, b []Vec2) Vec2 {
	return centroid(a).Add(centroid(b)).Scale(0.5)
}
