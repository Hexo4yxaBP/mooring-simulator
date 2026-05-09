package main

import (
	"log"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"github.com/dvpalkin/mooring-simulator/internal/input"
	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/render"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
	"github.com/dvpalkin/mooring-simulator/internal/ui"
)

const (
	screenW = 1280
	screenH = 720
	dt      = 1.0 / 60.0
)

// Game implements ebiten.Game.
type Game struct {
	world    *sim.World
	handler  input.Handler
	renderer render.Renderer
}

func newGame() *Game {
	world := defaultWorld()
	vp := render.DefaultViewport()
	return &Game{
		world: world,
		handler: input.Handler{
			Scale:       vp.Scale,
			OriginWorld: vp.OriginWorld,
			ScreenH:     vp.ScreenH,
		},
		renderer: render.Renderer{VP: vp},
	}
}

func (g *Game) Update() error {
	state := pollInput()
	cmds := g.handler.Poll(state, g.world)
	for _, cmd := range cmds {
		applyCommand(g.world, cmd)
	}
	g.world.Step(dt)
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	g.renderer.Draw(screen, g.world, &g.handler)
	ui.DrawHUD(screen, g.world)
	ui.DrawWindPanel(screen, g.world.Wind)
	ui.DrawRudderIndicator(screen, g.world, screenW, screenH)
	cx, cy := ebiten.CursorPosition()
	ui.DrawTensionTooltip(screen, g.world, cx, cy, screenW)
}

func (g *Game) Layout(_, _ int) (int, int) {
	return screenW, screenH
}

// pollInput reads Ebiten mouse/wheel state and returns an InputState for this tick.
func pollInput() input.InputState {
	var s input.InputState
	s.MouseX, s.MouseY = ebiten.CursorPosition()
	s.LeftJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft)
	s.LeftHeld = ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft)
	s.LeftJustReleased = inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft)
	s.RightJustPressed = inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight)
	_, wy := ebiten.Wheel()
	s.WheelDY = wy
	return s
}

// applyCommand mutates world state based on a validated command.
func applyCommand(world *sim.World, cmd input.Command) {
	switch cmd.Type {
	case input.CmdSetThrottle:
		p := cmd.Payload.(input.ThrottlePayload)
		if active := world.ActiveBoat(); active != nil {
			active.Throttle = p.State
		}
	case input.CmdSetRudder:
		p := cmd.Payload.(input.RudderPayload)
		if active := world.ActiveBoat(); active != nil {
			active.SetRudder(p.Angle)
		}
	case input.CmdSetWind:
		p := cmd.Payload.(input.WindPayload)
		world.Wind.Speed = p.Speed
		world.Wind.Direction = p.Direction
	case input.CmdSelectBoat:
		p := cmd.Payload.(input.SelectBoatPayload)
		world.SetActiveBoat(p.BoatID)
	case input.CmdAddMooringLine:
		p := cmd.Payload.(input.AddLinePayload)
		boat := worldBoat(world, p.BoatID)
		if boat == nil {
			return
		}
		cleatPos := boat.CleatWorldPos(p.Cleat)
		dist := p.DockPoint.Sub(cleatPos).Len()
		line := sim.MooringLine{
			DockPoint:     p.DockPoint,
			BoatID:        p.BoatID,
			Cleat:         p.Cleat,
			NaturalLength: dist * 0.95,
			Stiffness:     sim.DefaultMooringStiffness,
			Damping:       sim.DefaultMooringDamping,
		}
		_ = world.AddMooringLine(line)
	case input.CmdRemoveMooringLine:
		idx := cmd.Payload.(int)
		world.RemoveMooringLine(idx)
	case input.CmdAddBoat:
		p := cmd.Payload.(input.AddBoatPayload)
		id := len(world.Boats) + 1
		var b *sim.Boat
		switch p.Type {
		case sim.BoatMonohull:
			b = sim.NewMonohull(id, p.Position, p.Heading)
		case sim.BoatCatamaran:
			b = sim.NewCatamaran(id, p.Position, p.Heading)
		}
		if b != nil {
			// First boat placed becomes active.
			if len(world.Boats) == 0 {
				b.IsActive = true
			}
			world.Boats = append(world.Boats, b)
		}
	}
}

func worldBoat(world *sim.World, id int) *sim.Boat {
	for _, b := range world.Boats {
		if b.ID == id {
			return b
		}
	}
	return nil
}

// defaultWorld creates the scenario: pier at screen bottom, mooring points placed, no boats.
// Visible world at default viewport (scale=20, 1280×720): X ∈ [-32,+32], Y ∈ [-18,+18].
// Pier face at world Y=-13 → screenY=620 (100 px from bottom).
func defaultWorld() *sim.World {
	// Pier: face at Y=-13, body extends to Y=-21 (off-screen).
	dock := sim.NewBottomPier(physics.Vec2{X: 0, Y: -13}, 64, 8)

	// Mooring points: 7 pier cleats + 7 water buoys at X = -24..+24 step 8.
	mooringXs := []float64{-24, -16, -8, 0, 8, 16, 24}
	var pts []sim.MooringPoint
	id := 1
	for _, x := range mooringXs {
		pts = append(pts, sim.MooringPoint{
			ID:   id,
			Pos:  physics.Vec2{X: x, Y: -13},
			Kind: sim.MooringPierCleat,
		})
		id++
	}
	for _, x := range mooringXs {
		pts = append(pts, sim.MooringPoint{
			ID:   id,
			Pos:  physics.Vec2{X: x, Y: 4},
			Kind: sim.MooringBuoy,
		})
		id++
	}

	return &sim.World{
		Dock:          dock,
		MooringPoints: pts,
		Wind:          sim.WindField{Speed: 3.0, Direction: math.Pi / 2},
	}
}

func main() {
	ebiten.SetWindowSize(screenW, screenH)
	ebiten.SetWindowTitle("Mooring Simulator")
	if err := ebiten.RunGame(newGame()); err != nil {
		log.Fatal(err)
	}
}
