package render

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

func drawMooringLines(dst *ebiten.Image, world *sim.World, vp Viewport) {
	for i := range world.Lines {
		line := &world.Lines[i]

		boat := lineBoat(world.Boats, line.BoatID)
		if boat == nil {
			continue
		}

		ax, ay := vp.WorldToScreen(line.DockPoint)
		cleatPos := boat.CleatWorldPos(line.Cleat)
		bx, by := vp.WorldToScreen(cleatPos)

		maxTension := line.NaturalLength * line.Stiffness * 0.5
		var fraction float64
		if maxTension > 0 {
			fraction = line.Tension(world.Boats) / maxTension
		}
		clr := lineTensionColor(fraction)

		vector.StrokeLine(dst,
			float32(ax), float32(ay),
			float32(bx), float32(by),
			2, clr, false)

		drawCircle(dst, float32(ax), float32(ay), 3, clr)
	}
}

func lineBoat(boats []*sim.Boat, id int) *sim.Boat {
	for _, b := range boats {
		if b.ID == id {
			return b
		}
	}
	return nil
}
