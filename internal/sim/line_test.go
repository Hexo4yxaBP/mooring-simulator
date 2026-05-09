package sim

import (
	"math"
	"testing"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

func boatAt(x, y float64) *Boat {
	return NewMonohull(1, physics.Vec2{X: x, Y: y}, 0)
}

func TestMooringLineCurrentLength(t *testing.T) {
	b := boatAt(0, 0)
	// Bow cleat is ~4.5 m forward of CoM. CoM is at (-0.5, 0) from geometric center.
	// With geometric center at (0,0): CoM at (-0.5, 0), bow cleat at (-0.5 + 4.5, 0) = (4.0, 0).
	bowPos := b.CleatWorldPos(CleatBow)
	dockPt := physics.Vec2{X: bowPos.X + 3, Y: bowPos.Y} // 3 m away
	line := MooringLine{
		DockPoint: dockPt,
		BoatID:    1,
		Cleat:     CleatBow,
	}
	got := line.CurrentLength([]*Boat{b})
	if math.Abs(got-3) > 1e-9 {
		t.Errorf("CurrentLength: got %v want 3", got)
	}
}

func TestMooringLineTensionSlack(t *testing.T) {
	b := boatAt(0, 0)
	bowPos := b.CleatWorldPos(CleatBow)
	dockPt := physics.Vec2{X: bowPos.X + 3, Y: bowPos.Y}
	line := MooringLine{
		DockPoint:     dockPt,
		BoatID:        1,
		Cleat:         CleatBow,
		NaturalLength: 5, // naturalLen > currentLen → slack
		Stiffness:     5000,
	}
	if line.Tension([]*Boat{b}) != 0 {
		t.Error("Slack line tension should be 0")
	}
}

func TestMooringLineTensionExtended(t *testing.T) {
	b := boatAt(0, 0)
	bowPos := b.CleatWorldPos(CleatBow)
	// Place dock 6 m away from bow cleat
	dockPt := physics.Vec2{X: bowPos.X + 6, Y: bowPos.Y}
	line := MooringLine{
		DockPoint:     dockPt,
		BoatID:        1,
		Cleat:         CleatBow,
		NaturalLength: 5,
		Stiffness:     5000,
	}
	got := line.Tension([]*Boat{b})
	want := 5000.0 * 1.0 // stiffness * extension = 5000 * 1
	if math.Abs(got-want) > 1e-6 {
		t.Errorf("Tension: got %v want %v", got, want)
	}
}

func TestMooringLineTensionNonNegative(t *testing.T) {
	b := boatAt(0, 0)
	bowPos := b.CleatWorldPos(CleatBow)
	dockPt := physics.Vec2{X: bowPos.X + 3, Y: bowPos.Y}
	line := MooringLine{
		DockPoint:     dockPt,
		BoatID:        1,
		Cleat:         CleatBow,
		NaturalLength: 10, // very slack
		Stiffness:     5000,
	}
	if line.Tension([]*Boat{b}) < 0 {
		t.Error("Tension must never be negative")
	}
}
