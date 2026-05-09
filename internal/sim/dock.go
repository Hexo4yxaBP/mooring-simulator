package sim

import (
	"math"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

// Dock is a static, immovable structure. Boats cannot pass through it.
// Vertices define the dock polygon in world-space with CCW winding.
type Dock struct {
	Vertices []physics.Vec2
}

// NewStraightPier creates a rectangular dock aligned with the x-axis.
// pos: center of the boat-facing (south) edge, width: pier length, depth: pier thickness.
func NewStraightPier(pos physics.Vec2, width, depth float64) Dock {
	half := width / 2
	// CCW winding viewed from above (Y-up)
	return Dock{
		Vertices: []physics.Vec2{
			{X: pos.X + half, Y: pos.Y},         // SE corner (boat side)
			{X: pos.X - half, Y: pos.Y},         // SW corner
			{X: pos.X - half, Y: pos.Y + depth}, // NW corner
			{X: pos.X + half, Y: pos.Y + depth}, // NE corner
		},
	}
}

// NearestPointOnEdge returns the world-space point on the dock boundary
// polygon edge closest to p. Used for snapping mooring line dock endpoints.
func (d *Dock) NearestPointOnEdge(p physics.Vec2) physics.Vec2 {
	best := physics.Vec2{}
	bestDist := math.MaxFloat64
	n := len(d.Vertices)
	for i := range d.Vertices {
		a := d.Vertices[i]
		b := d.Vertices[(i+1)%n]
		closest := nearestOnSegment(a, b, p)
		dist := closest.Sub(p).LenSq()
		if dist < bestDist {
			bestDist = dist
			best = closest
		}
	}
	return best
}

// NewBottomPier creates a rectangular pier whose boat-facing edge is at pos.Y
// and whose body extends south (toward -Y). Suitable for a pier rendered at screen bottom.
func NewBottomPier(pos physics.Vec2, width, depth float64) Dock {
	half := width / 2
	return Dock{
		Vertices: []physics.Vec2{
			{X: pos.X + half, Y: pos.Y},         // NE corner (boat side)
			{X: pos.X - half, Y: pos.Y},         // NW corner
			{X: pos.X - half, Y: pos.Y - depth}, // SW corner
			{X: pos.X + half, Y: pos.Y - depth}, // SE corner
		},
	}
}

// nearestOnSegment returns the point on segment [a, b] nearest to p.
func nearestOnSegment(a, b, p physics.Vec2) physics.Vec2 {
	ab := b.Sub(a)
	lenSq := ab.LenSq()
	if lenSq < 1e-18 {
		return a // degenerate segment
	}
	t := p.Sub(a).Dot(ab) / lenSq
	t = math.Max(0, math.Min(1, t))
	return a.Add(ab.Scale(t))
}
