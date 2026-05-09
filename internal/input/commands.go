package input

import (
	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

// CommandType identifies the kind of game command produced by user input.
type CommandType int

const (
	CmdSetThrottle     CommandType = iota // payload: ThrottlePayload
	CmdSetRudder                          // payload: RudderPayload
	CmdSetWind                            // payload: WindPayload
	CmdSelectBoat                         // payload: SelectBoatPayload
	CmdAddMooringLine                     // payload: AddLinePayload
	CmdRemoveMooringLine                  // payload: int (line index)
	CmdAddBoat                            // payload: AddBoatPayload
)

// Command carries a typed user action from the input handler to the game loop.
type Command struct {
	Type    CommandType
	Payload any
}

type ThrottlePayload struct{ State sim.ThrottleState }
type RudderPayload struct{ Angle float64 } // radians, pre-clamped
type WindPayload struct {
	Speed     float64 // m/s, [0, 30]
	Direction float64 // radians, [0, 2π)
}
type SelectBoatPayload struct{ BoatID int }
type AddLinePayload struct {
	DockPoint physics.Vec2
	BoatID    int
	Cleat     sim.CleatID
}
type AddBoatPayload struct {
	Type     sim.BoatType
	Position physics.Vec2
	Heading  float64
}
