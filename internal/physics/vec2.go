package physics

import "math"

// Vec2 is a 2D vector used for all physics calculations (SI units: meters, m/s, etc.).
type Vec2 struct {
	X, Y float64
}

func (v Vec2) Add(u Vec2) Vec2        { return Vec2{X: v.X + u.X, Y: v.Y + u.Y} }
func (v Vec2) Sub(u Vec2) Vec2        { return Vec2{X: v.X - u.X, Y: v.Y - u.Y} }
func (v Vec2) Scale(s float64) Vec2   { return Vec2{X: v.X * s, Y: v.Y * s} }
func (v Vec2) Dot(u Vec2) float64     { return v.X*u.X + v.Y*u.Y }
func (v Vec2) Cross(u Vec2) float64   { return v.X*u.Y - v.Y*u.X }
func (v Vec2) Len() float64           { return math.Sqrt(v.X*v.X + v.Y*v.Y) }
func (v Vec2) LenSq() float64         { return v.X*v.X + v.Y*v.Y }

// Normalize returns the unit vector. Panics on zero-length vector (caller error).
func (v Vec2) Normalize() Vec2 {
	l := v.Len()
	if l == 0 {
		panic("physics.Vec2.Normalize: zero-length vector")
	}
	return Vec2{X: v.X / l, Y: v.Y / l}
}

// Rotate rotates v by angle radians counter-clockwise.
func (v Vec2) Rotate(angle float64) Vec2 {
	c, s := math.Cos(angle), math.Sin(angle)
	return Vec2{X: v.X*c - v.Y*s, Y: v.X*s + v.Y*c}
}

// Perp returns a vector perpendicular to v (90° CCW).
func (v Vec2) Perp() Vec2 { return Vec2{X: -v.Y, Y: v.X} }
