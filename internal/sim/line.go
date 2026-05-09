package sim

import "github.com/dvpalkin/mooring-simulator/internal/physics"

// MooringLine connects a fixed dock point to a boat cleat via a Hookean spring.
// Constraint P1: tension-only — no compression forces.
type MooringLine struct {
	DockPoint     physics.Vec2
	BoatID        int
	Cleat         CleatID
	NaturalLength float64 // meters; set to 0.95 × initial distance at placement
	Stiffness     float64 // N/m
	Damping       float64 // N·s/m
}

// findBoat returns the boat with matching ID, or nil.
func findBoat(boats []*Boat, id int) *Boat {
	for _, b := range boats {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// CurrentLength returns the current Euclidean distance from DockPoint to the cleat.
func (l *MooringLine) CurrentLength(boats []*Boat) float64 {
	b := findBoat(boats, l.BoatID)
	if b == nil {
		return 0
	}
	return l.DockPoint.Sub(b.CleatWorldPos(l.Cleat)).Len()
}

// Tension returns the current tension force magnitude in Newtons.
// Returns 0 when the line is slack (constraint P1).
func (l *MooringLine) Tension(boats []*Boat) float64 {
	cur := l.CurrentLength(boats)
	if cur <= l.NaturalLength {
		return 0
	}
	return l.Stiffness * (cur - l.NaturalLength)
}
