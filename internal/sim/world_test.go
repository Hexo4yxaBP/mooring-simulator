package sim

import (
	"math"
	"testing"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

// defaultWorld creates the standard startup scenario.
func defaultWorld() *World {
	dock := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	monohull := NewMonohull(1, physics.Vec2{X: 0, Y: 20}, math.Pi/2) // facing north toward dock
	monohull.IsActive = true
	catamaran := NewCatamaran(2, physics.Vec2{X: 0, Y: 10}, 0)
	return &World{
		Boats: []*Boat{monohull, catamaran},
		Dock:  dock,
		Wind:  WindField{Speed: 3.0, Direction: math.Pi / 2},
	}
}

func TestWorldInitSmoke(t *testing.T) {
	w := defaultWorld()
	for i := 0; i < 600; i++ {
		w.Step(1.0 / 60.0)
	}
	for _, b := range w.Boats {
		if math.IsNaN(b.Body.Position.X) || math.IsNaN(b.Body.Position.Y) {
			t.Errorf("Boat %d position became NaN after 600 steps", b.ID)
		}
		if math.IsNaN(b.Body.Velocity.X) || math.IsNaN(b.Body.Velocity.Y) {
			t.Errorf("Boat %d velocity became NaN after 600 steps", b.ID)
		}
	}
}

func TestBoatCoastsToStop(t *testing.T) {
	w := &World{
		Dock: NewStraightPier(physics.Vec2{X: 0, Y: 1000}, 50, 5), // far away
	}
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	b.Body.Velocity = physics.Vec2{X: 2, Y: 0}
	b.IsActive = true
	w.Boats = []*Boat{b}

	for i := 0; i < 600; i++ { // 10 seconds
		w.Step(1.0 / 60.0)
	}
	speed := b.Body.Velocity.Len()
	if speed >= 0.1 {
		t.Errorf("Boat should coast to near-stop, speed after 10s: %v", speed)
	}
}

func TestThrottleForwardAccelerates(t *testing.T) {
	w := &World{
		Dock: NewStraightPier(physics.Vec2{X: 0, Y: 1000}, 50, 5),
	}
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	b.IsActive = true
	b.Throttle = ThrottleFullFwd
	// Zero drag for this test — we just want to verify thrust creates acceleration
	b.LongDrag = 0
	b.LatDrag = 0
	b.RotDrag = 0
	w.Boats = []*Boat{b}
	w.Wind = WindField{} // no wind

	for i := 0; i < 300; i++ { // 5 seconds
		w.Step(1.0 / 60.0)
	}
	speed := b.Body.Velocity.Len()
	if speed < 1.0 {
		t.Errorf("Full forward throttle should accelerate to >1 m/s in 5s, got %v", speed)
	}
}

func TestPropWalkInAstern(t *testing.T) {
	w := &World{
		Dock: NewStraightPier(physics.Vec2{X: 0, Y: 1000}, 50, 5),
	}
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0) // heading east
	b.IsActive = true
	b.Throttle = ThrottleFullAstern
	b.PropWalk = 300 // positive = port walk
	b.LongDrag = 0
	b.LatDrag = 0
	b.RotDrag = 0
	w.Boats = []*Boat{b}
	w.Wind = WindField{}

	startY := b.Body.Position.Y
	for i := 0; i < 180; i++ { // 3 seconds
		w.Step(1.0 / 60.0)
	}
	// Positive PropWalk with astern → port side force → +Y displacement at heading=0
	displacement := b.Body.Position.Y - startY
	if displacement <= 0 {
		t.Errorf("Prop walk in astern should displace to port (+Y), got %v", displacement)
	}
}

func TestCollisionPushesBoatAway(t *testing.T) {
	dock := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	// Place boat so its hull overlaps the dock south face (y=30)
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 29}, math.Pi/2)
	b.LongDrag = 0
	b.LatDrag = 0
	b.RotDrag = 0
	w := &World{
		Boats: []*Boat{b},
		Dock:  dock,
	}
	startY := b.Body.Position.Y
	for i := 0; i < 6; i++ { // 0.1 seconds
		w.Step(1.0 / 60.0)
	}
	if b.Body.Position.Y >= startY {
		t.Errorf("Collision penalty should push boat south (away from dock), startY=%v endY=%v",
			startY, b.Body.Position.Y)
	}
}

func TestMooringHoldsBoat(t *testing.T) {
	dock := NewStraightPier(physics.Vec2{X: 0, Y: 30}, 50, 5)
	b := NewMonohull(1, physics.Vec2{X: 0, Y: 20}, math.Pi/2)
	b.IsActive = true
	sternPos := b.CleatWorldPos(CleatStern)
	dockPt := physics.Vec2{X: sternPos.X, Y: 30} // dock point above stern
	initDist := physics.Vec2{}.Add(dockPt).Sub(sternPos).Len()
	line := MooringLine{
		DockPoint:     dockPt,
		BoatID:        1,
		Cleat:         CleatStern,
		NaturalLength: initDist * 0.95,
		Stiffness:     DefaultMooringStiffness,
		Damping:       DefaultMooringDamping,
	}
	w := &World{
		Boats: []*Boat{b},
		Dock:  dock,
		Lines: []MooringLine{line},
		Wind:  WindField{Speed: 10, Direction: 0}, // beam-on wind from east
	}
	startX := b.Body.Position.X
	for i := 0; i < 1800; i++ { // 30 seconds
		w.Step(1.0 / 60.0)
	}
	lateralDisp := math.Abs(b.Body.Position.X - startX)
	if lateralDisp > 2.0 {
		t.Errorf("Moored boat should not drift >2 m laterally in beam wind, got %v m", lateralDisp)
	}
}

func TestSetActiveBoat(t *testing.T) {
	w := defaultWorld()
	w.SetActiveBoat(2)
	if w.Boats[0].IsActive {
		t.Error("Boat 1 should not be active after SetActiveBoat(2)")
	}
	if !w.Boats[1].IsActive {
		t.Error("Boat 2 should be active after SetActiveBoat(2)")
	}
}

func TestSetActiveBoatNotFound(t *testing.T) {
	w := defaultWorld()
	w.Boats[0].IsActive = true
	w.SetActiveBoat(999) // unknown ID
	if !w.Boats[0].IsActive {
		t.Error("SetActiveBoat with unknown ID should be no-op, boat 1 should remain active")
	}
}

func TestActiveBoatNone(t *testing.T) {
	w := &World{}
	if w.ActiveBoat() != nil {
		t.Error("ActiveBoat with no boats should return nil")
	}
}

func TestAddMooringLineValid(t *testing.T) {
	w := defaultWorld()
	err := w.AddMooringLine(MooringLine{
		BoatID:        1,
		Cleat:         CleatBow,
		NaturalLength: 5,
		Stiffness:     5000,
		Damping:       500,
	})
	if err != nil {
		t.Errorf("Valid AddMooringLine returned error: %v", err)
	}
	if len(w.Lines) != 1 {
		t.Errorf("Expected 1 line, got %d", len(w.Lines))
	}
}

func TestAddMooringLineInvalidBoat(t *testing.T) {
	w := defaultWorld()
	err := w.AddMooringLine(MooringLine{BoatID: 999, Cleat: CleatBow})
	if err == nil {
		t.Error("AddMooringLine with unknown boat ID should return error")
	}
}

func TestAddMooringLineInvalidCleat(t *testing.T) {
	w := defaultWorld()
	err := w.AddMooringLine(MooringLine{BoatID: 1, Cleat: CleatID(99)})
	if err == nil {
		t.Error("AddMooringLine with invalid cleat should return error")
	}
}

func TestRemoveMooringLine(t *testing.T) {
	w := defaultWorld()
	w.Lines = []MooringLine{{BoatID: 1, Cleat: CleatBow}, {BoatID: 1, Cleat: CleatStern}}
	w.RemoveMooringLine(0)
	if len(w.Lines) != 1 {
		t.Errorf("Expected 1 line after removal, got %d", len(w.Lines))
	}
}

func TestRemoveMooringLineOutOfRange(t *testing.T) {
	w := defaultWorld()
	w.RemoveMooringLine(-1)  // no panic
	w.RemoveMooringLine(100) // no panic
}
