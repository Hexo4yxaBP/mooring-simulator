package input

import (
	"math"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

// placementPhase tracks the current interaction state.
type placementPhase int

const (
	phaseIdle               placementPhase = iota
	phaseMooringPointChosen               // mooring point selected; awaiting cleat click
	phaseDraggingBoat                     // dragging boat icon from left panel
)

const (
	panelWidth        = 100   // px — left boat-icon panel width
	panelSplit        = 360   // px — Y split between monohull (above) and catamaran (below) icons
	mooringHitRadius  = 1.5   // world-space meters — mooring point click detection
	cleatHitRadius    = 3.5   // world-space meters — cleat click detection (covers bilateral visual dots)
	lineRemoveRadius  = 1.5   // world-space meters — right-click line removal
	rudderSensitivity = 0.003 // radians per pixel of horizontal drag
	windPanelMinX     = 1050  // screen-space px — wind panel bounds
	windPanelMinY     = 10
	windPanelMaxX     = 1270
	windPanelMaxY     = 180
	windScrollStep    = 0.5 // m/s per scroll notch
	maxWindSpeed      = 30.0
)

// Handler translates InputState into game Commands.
// It holds the viewport parameters to convert screen↔world coordinates.
type Handler struct {
	// Viewport parameters — set by main.go to match Renderer.VP.
	Scale       float64
	OriginWorld physics.Vec2
	ScreenH     int

	phase            placementPhase
	pendingMooringPt physics.Vec2
	draggingBoatType sim.BoatType

	prevMouseX int
}

// PlacementActive reports whether the user is mid-way through placing a mooring line.
func (h *Handler) PlacementActive() bool { return h.phase == phaseMooringPointChosen }

// DraggingBoat reports whether the user is dragging a boat from the left panel.
func (h *Handler) DraggingBoat() bool { return h.phase == phaseDraggingBoat }

// DraggingBoatType returns the type being dragged (valid only when DraggingBoat is true).
func (h *Handler) DraggingBoatType() sim.BoatType { return h.draggingBoatType }

// PendingDockPoint returns the selected mooring point position during line placement.
func (h *Handler) PendingDockPoint() physics.Vec2 { return h.pendingMooringPt }

// Poll reads InputState and world state, returns zero or more Commands for this tick.
func (h *Handler) Poll(state InputState, world *sim.World) []Command {
	var cmds []Command

	dx := state.MouseX - h.prevMouseX
	h.prevMouseX = state.MouseX

	active := world.ActiveBoat()

	// --- Wind panel scroll: adjust wind speed ---
	if state.WheelDY != 0 && inWindPanel(state.MouseX, state.MouseY) {
		newSpeed := clampF(world.Wind.Speed+state.WheelDY*windScrollStep, 0, maxWindSpeed)
		cmds = append(cmds, Command{CmdSetWind, WindPayload{newSpeed, normaliseAngle(world.Wind.Direction)}})
		return cmds
	}

	// --- Throttle scroll (outside wind panel) ---
	if state.WheelDY != 0 && active != nil {
		var next sim.ThrottleState
		if state.WheelDY > 0 {
			next = throttleForward(active.Throttle)
		} else {
			next = throttleAstern(active.Throttle)
		}
		if next != active.Throttle {
			cmds = append(cmds, Command{CmdSetThrottle, ThrottlePayload{next}})
		}
	}

	// --- Right click: remove nearest mooring line ---
	if state.RightJustPressed {
		worldPt := h.screenToWorld(state.MouseX, state.MouseY)
		cmds = append(cmds, h.handleRightClick(worldPt, world)...)
	}

	switch h.phase {
	case phaseIdle:
		if state.LeftJustPressed {
			if inPanel(state.MouseX) {
				h.phase = phaseDraggingBoat
				h.draggingBoatType = panelBoatType(state.MouseY)
			} else {
				worldPt := h.screenToWorld(state.MouseX, state.MouseY)
				// Check mooring point click
				for i := range world.MooringPoints {
					mp := &world.MooringPoints[i]
					if worldPt.Sub(mp.Pos).Len() <= mooringHitRadius {
						h.phase = phaseMooringPointChosen
						h.pendingMooringPt = mp.Pos
						return cmds
					}
				}
				// Check boat click → select
				for _, b := range world.Boats {
					if pointInRect(worldPt, b.HullVertices) {
						cmds = append(cmds, Command{CmdSelectBoat, SelectBoatPayload{b.ID}})
						return cmds
					}
				}
			}
		}
		// Rudder: left held (not the press frame itself, not in panel)
		if state.LeftHeld && !state.LeftJustPressed && !inPanel(state.MouseX) && dx != 0 && active != nil {
			newAngle := clampF(active.Rudder+float64(dx)*rudderSensitivity, -sim.MaxRudderAngle, sim.MaxRudderAngle)
			cmds = append(cmds, Command{CmdSetRudder, RudderPayload{newAngle}})
		}

	case phaseMooringPointChosen:
		if state.LeftJustPressed {
			worldPt := h.screenToWorld(state.MouseX, state.MouseY)
			if active != nil {
				for _, id := range []sim.CleatID{sim.CleatBow, sim.CleatMidships, sim.CleatStern} {
					cp := active.CleatWorldPos(id)
					if worldPt.Sub(cp).Len() <= cleatHitRadius {
						h.phase = phaseIdle
						cmds = append(cmds, Command{CmdAddMooringLine, AddLinePayload{
							DockPoint: h.pendingMooringPt,
							BoatID:    active.ID,
							Cleat:     id,
						}})
						return cmds
					}
				}
			}
			// Miss → cancel
			h.phase = phaseIdle
		}

	case phaseDraggingBoat:
		if state.LeftJustReleased {
			if !inPanel(state.MouseX) {
				worldPt := h.screenToWorld(state.MouseX, state.MouseY)
				cmds = append(cmds, Command{CmdAddBoat, AddBoatPayload{
					Type:     h.draggingBoatType,
					Position: worldPt,
					Heading:  math.Pi / 2, // default heading: bow pointing north
				}})
			}
			h.phase = phaseIdle
		}
	}

	return cmds
}

// handleRightClick finds the nearest mooring line and emits a removal command.
func (h *Handler) handleRightClick(worldPt physics.Vec2, world *sim.World) []Command {
	bestIdx := -1
	bestDist := math.MaxFloat64
	for i, line := range world.Lines {
		boat := findBoatByID(world.Boats, line.BoatID)
		if boat == nil {
			continue
		}
		cleatPos := boat.CleatWorldPos(line.Cleat)
		mid := line.DockPoint.Add(cleatPos).Scale(0.5)
		d := worldPt.Sub(mid).Len()
		if d < bestDist {
			bestDist = d
			bestIdx = i
		}
	}
	if bestIdx >= 0 && bestDist <= lineRemoveRadius {
		return []Command{{CmdRemoveMooringLine, bestIdx}}
	}
	return nil
}

// screenToWorld converts screen-space pixel coordinates to world-space meters.
func (h *Handler) screenToWorld(sx, sy int) physics.Vec2 {
	return physics.Vec2{
		X: float64(sx)/h.Scale + h.OriginWorld.X,
		Y: float64(h.ScreenH-sy)/h.Scale + h.OriginWorld.Y,
	}
}

func inPanel(screenX int) bool { return screenX < panelWidth }

func inWindPanel(x, y int) bool {
	return x >= windPanelMinX && x <= windPanelMaxX && y >= windPanelMinY && y <= windPanelMaxY
}

func panelBoatType(screenY int) sim.BoatType {
	if screenY < panelSplit {
		return sim.BoatMonohull
	}
	return sim.BoatCatamaran
}

func findBoatByID(boats []*sim.Boat, id int) *sim.Boat {
	for _, b := range boats {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// pointInRect tests if p is inside the axis-aligned bounding box of 4 corners.
func pointInRect(p physics.Vec2, verts [4]physics.Vec2) bool {
	minX, maxX := verts[0].X, verts[0].X
	minY, maxY := verts[0].Y, verts[0].Y
	for _, v := range verts[1:] {
		if v.X < minX {
			minX = v.X
		}
		if v.X > maxX {
			maxX = v.X
		}
		if v.Y < minY {
			minY = v.Y
		}
		if v.Y > maxY {
			maxY = v.Y
		}
	}
	return p.X >= minX && p.X <= maxX && p.Y >= minY && p.Y <= maxY
}

// throttleForward steps throttle one notch toward full-forward.
func throttleForward(t sim.ThrottleState) sim.ThrottleState {
	switch t {
	case sim.ThrottleFullAstern:
		return sim.ThrottleSlowAstern
	case sim.ThrottleSlowAstern:
		return sim.ThrottleNeutral
	case sim.ThrottleNeutral:
		return sim.ThrottleSlowFwd
	case sim.ThrottleSlowFwd:
		return sim.ThrottleFullFwd
	default:
		return t
	}
}

// throttleAstern steps throttle one notch toward full-astern.
func throttleAstern(t sim.ThrottleState) sim.ThrottleState {
	switch t {
	case sim.ThrottleFullFwd:
		return sim.ThrottleSlowFwd
	case sim.ThrottleSlowFwd:
		return sim.ThrottleNeutral
	case sim.ThrottleNeutral:
		return sim.ThrottleSlowAstern
	case sim.ThrottleSlowAstern:
		return sim.ThrottleFullAstern
	default:
		return t
	}
}

func clampF(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

func normaliseAngle(a float64) float64 {
	a = math.Mod(a, 2*math.Pi)
	if a < 0 {
		a += 2 * math.Pi
	}
	return a
}
