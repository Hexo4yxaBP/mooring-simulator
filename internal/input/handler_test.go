package input

import (
	"math"
	"testing"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

func makeWorld() *sim.World {
	dock := sim.NewBottomPier(physics.Vec2{X: 0, Y: -13}, 64, 8)
	b := sim.NewMonohull(1, physics.Vec2{X: 0, Y: 0}, 0)
	b.IsActive = true
	return &sim.World{
		Boats: []*sim.Boat{b},
		Dock:  dock,
		Wind:  sim.WindField{Speed: 5, Direction: 0},
		MooringPoints: []sim.MooringPoint{
			{ID: 1, Pos: physics.Vec2{X: 0, Y: -13}, Kind: sim.MooringPierCleat},
			{ID: 2, Pos: physics.Vec2{X: 0, Y: 4}, Kind: sim.MooringBuoy},
		},
	}
}

func makeHandler() *Handler {
	return &Handler{Scale: 20, OriginWorld: physics.Vec2{X: -32, Y: -18}, ScreenH: 720}
}

// worldToScreen is the inverse of screenToWorld, used in tests.
func (h *Handler) worldToScreen(p physics.Vec2) (float64, float64) {
	sx := (p.X - h.OriginWorld.X) * h.Scale
	sy := float64(h.ScreenH) - (p.Y-h.OriginWorld.Y)*h.Scale
	return sx, sy
}

// --- Throttle via scroll wheel ---

func TestScrollThrottleForward(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Throttle = sim.ThrottleNeutral
	h := makeHandler()
	var s InputState
	s.WheelDY = 1
	s.MouseX, s.MouseY = 640, 400
	cmds := h.Poll(s, w)
	found := false
	for _, c := range cmds {
		if c.Type == CmdSetThrottle {
			p := c.Payload.(ThrottlePayload)
			if p.State == sim.ThrottleSlowFwd {
				found = true
			}
		}
	}
	if !found {
		t.Error("scroll up from Neutral should emit ThrottleSlowFwd")
	}
}

func TestScrollThrottleAstern(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Throttle = sim.ThrottleNeutral
	h := makeHandler()
	var s InputState
	s.WheelDY = -1
	s.MouseX, s.MouseY = 640, 400
	cmds := h.Poll(s, w)
	found := false
	for _, c := range cmds {
		if c.Type == CmdSetThrottle {
			p := c.Payload.(ThrottlePayload)
			if p.State == sim.ThrottleSlowAstern {
				found = true
			}
		}
	}
	if !found {
		t.Error("scroll down from Neutral should emit ThrottleSlowAstern")
	}
}

func TestScrollThrottleAtFullFwd(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Throttle = sim.ThrottleFullFwd
	h := makeHandler()
	var s InputState
	s.WheelDY = 1
	s.MouseX, s.MouseY = 640, 400
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdSetThrottle {
			t.Error("scroll up at FullFwd should not emit throttle command")
		}
	}
}

func TestScrollThrottleAtFullAstern(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Throttle = sim.ThrottleFullAstern
	h := makeHandler()
	var s InputState
	s.WheelDY = -1
	s.MouseX, s.MouseY = 640, 400
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdSetThrottle {
			t.Error("scroll down at FullAstern should not emit throttle command")
		}
	}
}

// throttle transition coverage (all steps)

func TestScrollThrottleSlowFwdToFull(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Throttle = sim.ThrottleSlowFwd
	h := makeHandler()
	var s InputState
	s.WheelDY = 1
	s.MouseX, s.MouseY = 640, 400
	cmds := h.Poll(s, w)
	found := false
	for _, c := range cmds {
		if c.Type == CmdSetThrottle {
			if c.Payload.(ThrottlePayload).State == sim.ThrottleFullFwd {
				found = true
			}
		}
	}
	if !found {
		t.Error("scroll up from SlowFwd should emit FullFwd")
	}
}

func TestScrollThrottleSlowAsternToFull(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Throttle = sim.ThrottleSlowAstern
	h := makeHandler()
	var s InputState
	s.WheelDY = -1
	s.MouseX, s.MouseY = 640, 400
	cmds := h.Poll(s, w)
	found := false
	for _, c := range cmds {
		if c.Type == CmdSetThrottle {
			if c.Payload.(ThrottlePayload).State == sim.ThrottleFullAstern {
				found = true
			}
		}
	}
	if !found {
		t.Error("scroll down from SlowAstern should emit FullAstern")
	}
}

// --- Wind scroll in wind panel ---

func TestScrollWindInPanel(t *testing.T) {
	w := makeWorld()
	w.Wind.Speed = 5
	h := makeHandler()
	var s InputState
	s.MouseX = (windPanelMinX + windPanelMaxX) / 2
	s.MouseY = (windPanelMinY + windPanelMaxY) / 2
	s.WheelDY = 1
	cmds := h.Poll(s, w)
	found := false
	for _, c := range cmds {
		if c.Type == CmdSetWind {
			if c.Payload.(WindPayload).Speed > 5 {
				found = true
			}
		}
	}
	if !found {
		t.Error("scroll up in wind panel should increase wind speed")
	}
}

func TestScrollWindNotThrottleInPanel(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Throttle = sim.ThrottleNeutral
	h := makeHandler()
	var s InputState
	s.MouseX = (windPanelMinX + windPanelMaxX) / 2
	s.MouseY = (windPanelMinY + windPanelMaxY) / 2
	s.WheelDY = 1
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdSetThrottle {
			t.Error("scroll in wind panel should not emit throttle command")
		}
	}
}

func TestScrollWindClampAtZero(t *testing.T) {
	w := makeWorld()
	w.Wind.Speed = 0
	h := makeHandler()
	var s InputState
	s.MouseX = (windPanelMinX + windPanelMaxX) / 2
	s.MouseY = (windPanelMinY + windPanelMaxY) / 2
	s.WheelDY = -1
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdSetWind {
			if c.Payload.(WindPayload).Speed < 0 {
				t.Errorf("wind speed should not go below 0: %v", c.Payload.(WindPayload).Speed)
			}
		}
	}
}

func TestScrollWindClampAtMax(t *testing.T) {
	w := makeWorld()
	w.Wind.Speed = maxWindSpeed
	h := makeHandler()
	var s InputState
	s.MouseX = (windPanelMinX + windPanelMaxX) / 2
	s.MouseY = (windPanelMinY + windPanelMaxY) / 2
	s.WheelDY = 1
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdSetWind {
			if c.Payload.(WindPayload).Speed > maxWindSpeed {
				t.Errorf("wind speed exceeded max: %v", c.Payload.(WindPayload).Speed)
			}
		}
	}
}

// --- Rudder via horizontal drag ---

func TestRudderDragRight(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Rudder = 0
	h := makeHandler()
	h.prevMouseX = 600
	var s InputState
	s.MouseX = 650
	s.MouseY = 400
	s.LeftHeld = true
	s.LeftJustPressed = false
	cmds := h.Poll(s, w)
	found := false
	for _, c := range cmds {
		if c.Type == CmdSetRudder {
			if c.Payload.(RudderPayload).Angle > 0 {
				found = true
			}
		}
	}
	if !found {
		t.Error("drag right with left held should produce positive rudder angle")
	}
}

func TestRudderDragLeft(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Rudder = 0
	h := makeHandler()
	h.prevMouseX = 600
	var s InputState
	s.MouseX = 550
	s.MouseY = 400
	s.LeftHeld = true
	s.LeftJustPressed = false
	cmds := h.Poll(s, w)
	found := false
	for _, c := range cmds {
		if c.Type == CmdSetRudder {
			if c.Payload.(RudderPayload).Angle < 0 {
				found = true
			}
		}
	}
	if !found {
		t.Error("drag left with left held should produce negative rudder angle")
	}
}

func TestRudderClampAtMax(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Rudder = sim.MaxRudderAngle - 0.001
	h := makeHandler()
	h.prevMouseX = 400
	var s InputState
	s.MouseX = 700 // large rightward drag
	s.MouseY = 400
	s.LeftHeld = true
	s.LeftJustPressed = false
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdSetRudder {
			p := c.Payload.(RudderPayload)
			if p.Angle > sim.MaxRudderAngle+1e-9 {
				t.Errorf("rudder angle %v exceeds MaxRudderAngle", p.Angle)
			}
		}
	}
}

func TestRudderNotInPanel(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Rudder = 0
	h := makeHandler()
	h.prevMouseX = 50
	var s InputState
	s.MouseX = 80 // inside panel (< panelWidth=100)
	s.MouseY = 400
	s.LeftHeld = true
	s.LeftJustPressed = false
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdSetRudder {
			t.Error("drag inside left panel should not emit rudder command")
		}
	}
}

func TestRudderPayloadClamped(t *testing.T) {
	w := makeWorld()
	w.ActiveBoat().Rudder = sim.MaxRudderAngle - 0.001
	h := makeHandler()
	for i := 0; i < 20; i++ {
		h.prevMouseX = 0
		var s InputState
		s.MouseX = 100 // dx = +100 each iteration
		s.MouseY = 400
		s.LeftHeld = true
		s.LeftJustPressed = false
		cmds := h.Poll(s, w)
		for _, c := range cmds {
			if c.Type == CmdSetRudder {
				p := c.Payload.(RudderPayload)
				if math.Abs(p.Angle) > sim.MaxRudderAngle+1e-9 {
					t.Fatalf("iteration %d: rudder %v exceeds MaxRudderAngle", i, p.Angle)
				}
				w.ActiveBoat().SetRudder(p.Angle)
			}
		}
	}
}

// --- Mooring placement state machine ---

func TestMooringPointClickActivatesPlacement(t *testing.T) {
	w := makeWorld()
	h := makeHandler()
	mp := w.MooringPoints[0]
	sx, sy := h.worldToScreen(mp.Pos)
	var s InputState
	s.LeftJustPressed = true
	s.LeftHeld = true
	s.MouseX, s.MouseY = int(sx), int(sy)
	h.Poll(s, w)
	if !h.PlacementActive() {
		t.Error("click on mooring point should activate placement")
	}
}

func TestMooringTwoStep(t *testing.T) {
	w := makeWorld()
	h := makeHandler()

	// Step 1: click mooring point
	mp := w.MooringPoints[0]
	sx, sy := h.worldToScreen(mp.Pos)
	var s1 InputState
	s1.LeftJustPressed = true
	s1.LeftHeld = true
	s1.MouseX, s1.MouseY = int(sx), int(sy)
	h.Poll(s1, w)

	if !h.PlacementActive() {
		t.Fatal("placement should be active after mooring point click")
	}

	// Step 2: click bow cleat
	cleatPos := w.ActiveBoat().CleatWorldPos(sim.CleatBow)
	sx2, sy2 := h.worldToScreen(cleatPos)
	var s2 InputState
	s2.LeftJustPressed = true
	s2.LeftHeld = true
	s2.MouseX, s2.MouseY = int(sx2), int(sy2)
	cmds := h.Poll(s2, w)

	if h.PlacementActive() {
		t.Error("placement should end after cleat click")
	}
	found := false
	for _, c := range cmds {
		if c.Type == CmdAddMooringLine {
			found = true
		}
	}
	if !found {
		t.Error("cleat click should emit CmdAddMooringLine")
	}
}

func TestMooringMissCancel(t *testing.T) {
	w := makeWorld()
	h := makeHandler()
	h.phase = phaseMooringPointChosen
	h.pendingMooringPt = physics.Vec2{X: 0, Y: -13}

	var s InputState
	s.LeftJustPressed = true
	s.MouseX, s.MouseY = 200, 200 // far from any cleat
	cmds := h.Poll(s, w)

	if h.PlacementActive() {
		t.Error("miss click should cancel placement")
	}
	for _, c := range cmds {
		if c.Type == CmdAddMooringLine {
			t.Error("miss click should not emit CmdAddMooringLine")
		}
	}
}

func TestPendingMooringPointSet(t *testing.T) {
	w := makeWorld()
	h := makeHandler()
	mp := w.MooringPoints[1] // buoy at (0, 4)
	sx, sy := h.worldToScreen(mp.Pos)
	var s InputState
	s.LeftJustPressed = true
	s.MouseX, s.MouseY = int(sx), int(sy)
	h.Poll(s, w)
	if !h.PlacementActive() {
		t.Fatal("placement not active after mooring point click")
	}
	p := h.PendingDockPoint()
	if math.Abs(p.Y-4) > 0.1 {
		t.Errorf("PendingDockPoint Y should be ~4, got %v", p.Y)
	}
}

// --- Boat drag from left panel ---

func TestBoatDragPlaceMonohull(t *testing.T) {
	w := makeWorld()
	h := makeHandler()

	// Press in top half of panel (monohull)
	var s1 InputState
	s1.LeftJustPressed = true
	s1.LeftHeld = true
	s1.MouseX, s1.MouseY = 50, 180
	h.Poll(s1, w)

	if !h.DraggingBoat() {
		t.Fatal("pressing in panel should start boat drag")
	}
	if h.DraggingBoatType() != sim.BoatMonohull {
		t.Errorf("top-half panel → want Monohull, got %v", h.DraggingBoatType())
	}

	// Release in water
	var s2 InputState
	s2.LeftJustReleased = true
	s2.MouseX, s2.MouseY = 640, 300
	cmds := h.Poll(s2, w)

	if h.DraggingBoat() {
		t.Error("releasing should end drag")
	}
	found := false
	for _, c := range cmds {
		if c.Type == CmdAddBoat {
			if c.Payload.(AddBoatPayload).Type == sim.BoatMonohull {
				found = true
			}
		}
	}
	if !found {
		t.Error("releasing in water should emit CmdAddBoat with Monohull")
	}
}

func TestBoatDragPlaceCatamaran(t *testing.T) {
	w := makeWorld()
	h := makeHandler()

	// Press in bottom half of panel (catamaran)
	var s1 InputState
	s1.LeftJustPressed = true
	s1.LeftHeld = true
	s1.MouseX, s1.MouseY = 50, 420
	h.Poll(s1, w)

	if h.DraggingBoatType() != sim.BoatCatamaran {
		t.Errorf("bottom-half panel → want Catamaran, got %v", h.DraggingBoatType())
	}

	var s2 InputState
	s2.LeftJustReleased = true
	s2.MouseX, s2.MouseY = 640, 300
	cmds := h.Poll(s2, w)

	found := false
	for _, c := range cmds {
		if c.Type == CmdAddBoat {
			if c.Payload.(AddBoatPayload).Type == sim.BoatCatamaran {
				found = true
			}
		}
	}
	if !found {
		t.Error("catamaran panel drag should emit CmdAddBoat with Catamaran")
	}
}

func TestBoatDragCancelInPanel(t *testing.T) {
	w := makeWorld()
	h := makeHandler()

	var s1 InputState
	s1.LeftJustPressed = true
	s1.LeftHeld = true
	s1.MouseX, s1.MouseY = 50, 180
	h.Poll(s1, w)

	// Release back in panel
	var s2 InputState
	s2.LeftJustReleased = true
	s2.MouseX, s2.MouseY = 50, 200
	cmds := h.Poll(s2, w)

	for _, c := range cmds {
		if c.Type == CmdAddBoat {
			t.Error("releasing in panel should not emit CmdAddBoat")
		}
	}
	if h.DraggingBoat() {
		t.Error("releasing in panel should end drag phase")
	}
}

// --- Right-click line removal ---

func TestRightClickRemovesLine(t *testing.T) {
	w := makeWorld()
	b := w.ActiveBoat()
	sternPos := b.CleatWorldPos(sim.CleatStern)
	dockPt := physics.Vec2{X: sternPos.X, Y: -13}
	_ = w.AddMooringLine(sim.MooringLine{
		DockPoint:     dockPt,
		BoatID:        b.ID,
		Cleat:         sim.CleatStern,
		NaturalLength: 10,
		Stiffness:     sim.DefaultMooringStiffness,
		Damping:       sim.DefaultMooringDamping,
	})

	h := makeHandler()
	mid := dockPt.Add(sternPos).Scale(0.5)
	sx, sy := h.worldToScreen(mid)
	var s InputState
	s.RightJustPressed = true
	s.MouseX, s.MouseY = int(sx), int(sy)
	cmds := h.Poll(s, w)

	found := false
	for _, c := range cmds {
		if c.Type == CmdRemoveMooringLine {
			found = true
		}
	}
	if !found {
		t.Error("right-click on line midpoint should emit CmdRemoveMooringLine")
	}
}

func TestRightClickNoLineEmitsNothing(t *testing.T) {
	w := makeWorld()
	h := makeHandler()
	var s InputState
	s.RightJustPressed = true
	s.MouseX, s.MouseY = 200, 200
	cmds := h.Poll(s, w)
	for _, c := range cmds {
		if c.Type == CmdRemoveMooringLine {
			t.Error("right-click far from any line should not emit CmdRemoveMooringLine")
		}
	}
}

// --- Boat selection ---

func TestClickSelectBoat(t *testing.T) {
	w := makeWorld()
	h := makeHandler()
	b := w.Boats[0]
	comPos := b.Body.ComWorldPos()
	sx, sy := h.worldToScreen(comPos)
	var s InputState
	s.LeftJustPressed = true
	s.LeftHeld = true
	s.MouseX, s.MouseY = int(sx), int(sy)
	cmds := h.Poll(s, w)

	found := false
	for _, c := range cmds {
		if c.Type == CmdSelectBoat {
			found = true
		}
	}
	if !found {
		t.Error("clicking on boat should emit CmdSelectBoat")
	}
}
