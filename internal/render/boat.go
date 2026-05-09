package render

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
	"github.com/dvpalkin/mooring-simulator/internal/sim"
)

var (
	hullWhite      = color.RGBA{0xF5, 0xF5, 0xF0, 0xff}
	hullOutlineClr = color.RGBA{0x0A, 0x0A, 0x0A, 0xff}
	cabinFill      = color.RGBA{0x28, 0x28, 0x32, 0xff}
	cabinBorder    = color.RGBA{0x0A, 0x0A, 0x0A, 0xff}
	activeOutline  = color.RGBA{0xFF, 0xFF, 0x00, 0xff}
	comDotClr      = color.RGBA{0xBB, 0x00, 0x00, 0xff}
	cleatOuter     = color.RGBA{0xCC, 0xFF, 0x00, 0xff} // chartreuse
	cleatInner     = color.RGBA{0x00, 0xAA, 0x00, 0xff} // green
)

// HullScreenPts returns the outer hull vertices of a monohull in screen space (exported for tests).
func HullScreenPts(b *sim.Boat, vp Viewport) []physics.Vec2 {
	return worldToScreenPts(monohullWorldPts(b), vp)
}

// monohullWorldPts returns an 8-point oval hull. pts[0] = bow tip, pts[4] = stern.
func monohullWorldPts(b *sim.Boat) []physics.Vec2 {
	hl := b.HullLen / 2
	hb := b.HullBeam / 2
	return bodyToWorld([]physics.Vec2{
		{X: +hl, Y: 0},                  // [0] bow tip — sharp point
		{X: +hl * 0.60, Y: +hb},        // [1] bow stbd (quick widening)
		{X: -hl * 0.10, Y: +hb},        // [2] max-beam stbd
		{X: -hl * 0.65, Y: +hb * 0.85}, // [3] stern stbd shoulder
		{X: -hl, Y: 0},                 // [4] stern — rounded
		{X: -hl * 0.65, Y: -hb * 0.85}, // [5] stern port shoulder
		{X: -hl * 0.10, Y: -hb},        // [6] max-beam port
		{X: +hl * 0.60, Y: -hb},        // [7] bow port
	}, b)
}

// monohullCabinWorldPts returns the 4 corners of the cabin rectangle.
func monohullCabinWorldPts(b *sim.Boat) []physics.Vec2 {
	hl := b.HullLen / 2
	hb := b.HullBeam / 2
	return bodyToWorld([]physics.Vec2{
		{X: +hl * 0.24, Y: +hb * 0.58},
		{X: -hl * 0.50, Y: +hb * 0.58},
		{X: -hl * 0.50, Y: -hb * 0.58},
		{X: +hl * 0.24, Y: -hb * 0.58},
	}, b)
}

// catConstants returns the shared dimensional constants for catamaran rendering.
func catConstants(b *sim.Boat) (hl, sp, hw, dw, df, db float64) {
	hl = b.HullLen / 2
	sp = b.HullBeam * 0.37 // hull centreline offset from boat centre
	hw = b.HullBeam * 0.13 // hull half-width
	dw = sp - hw           // deck half-width = inner hull edge offset
	df = hl * 0.70         // deck front X
	db = -hl * 0.57        // deck back X
	return
}

// catamaranWorldPts returns port hull, starboard hull, and deck polygons.
// Each hull polygon's inner edge routes through the deck corners, creating the
// V-strut silhouette visible at bow and stern.
func catamaranWorldPts(b *sim.Boat) (port, stbd, deck []physics.Vec2) {
	hl, sp, hw, dw, df, db := catConstants(b)

	mkHull := func(sign float64) []physics.Vec2 {
		y := sign * sp
		return bodyToWorld([]physics.Vec2{
			{X: +hl, Y: y},                         // bow tip (outer)
			{X: +hl * 0.55, Y: y + sign*hw},        // bow outer shoulder
			{X: -hl * 0.15, Y: y + sign*hw},        // outer max-beam
			{X: -hl * 0.65, Y: y + sign*hw*0.60},   // outer stern shoulder
			{X: -hl * 0.95, Y: y},                  // stern tip
			{X: -hl * 0.65, Y: y - sign*hw*0.60},   // inner stern shoulder
			{X: db, Y: sign * dw},                  // deck back corner → stern strut
			{X: df, Y: sign * dw},                  // deck fore corner → closing = bow strut
		}, b)
	}

	port = mkHull(+1)
	stbd = mkHull(-1)

	deck = bodyToWorld([]physics.Vec2{
		{X: df, Y: +dw},
		{X: db, Y: +dw},
		{X: db, Y: -dw},
		{X: df, Y: -dw},
	}, b)
	return
}

// catamaranCockpitWorldPts returns the inner cockpit outline (drawn on the deck).
func catamaranCockpitWorldPts(b *sim.Boat) []physics.Vec2 {
	_, _, _, dw, df, db := catConstants(b)
	inset := dw * 0.25
	return bodyToWorld([]physics.Vec2{
		{X: df - inset, Y: +(dw - inset)},
		{X: db + inset, Y: +(dw - inset)},
		{X: db + inset, Y: -(dw - inset)},
		{X: df - inset, Y: -(dw - inset)},
	}, b)
}

func bodyToWorld(pts []physics.Vec2, b *sim.Boat) []physics.Vec2 {
	com := b.Body.ComWorldPos()
	out := make([]physics.Vec2, len(pts))
	for i, p := range pts {
		out[i] = com.Add(p.Rotate(b.Body.Heading))
	}
	return out
}

func worldToScreenPts(pts []physics.Vec2, vp Viewport) []physics.Vec2 {
	out := make([]physics.Vec2, len(pts))
	for i, p := range pts {
		sx, sy := vp.WorldToScreen(p)
		out[i] = physics.Vec2{X: sx, Y: sy}
	}
	return out
}

func drawBoat(dst *ebiten.Image, b *sim.Boat, vp Viewport) {
	switch b.Type {
	case sim.BoatMonohull:
		sp := worldToScreenPts(monohullWorldPts(b), vp)
		drawFilledPolygon(dst, sp, hullWhite)
		drawPolygonOutline(dst, sp, 4, hullOutlineClr)
		// Cabin
		cp := worldToScreenPts(monohullCabinWorldPts(b), vp)
		drawFilledPolygon(dst, cp, cabinFill)
		drawPolygonOutline(dst, cp, 2.5, cabinBorder)
		// Active selection ring drawn over outline
		if b.IsActive {
			drawPolygonOutline(dst, sp, 2, activeOutline)
		}

	case sim.BoatCatamaran:
		port, stbd, deck := catamaranWorldPts(b)
		// Hulls first (underneath deck)
		for _, pts := range [][]physics.Vec2{port, stbd} {
			sp := worldToScreenPts(pts, vp)
			drawFilledPolygon(dst, sp, hullWhite)
			drawPolygonOutline(dst, sp, 3.5, hullOutlineClr)
		}
		// Deck on top (covers hull inner sections)
		dsp := worldToScreenPts(deck, vp)
		drawFilledPolygon(dst, dsp, hullWhite)
		drawPolygonOutline(dst, dsp, 2.5, hullOutlineClr)
		// Inner cockpit rectangle
		csp := worldToScreenPts(catamaranCockpitWorldPts(b), vp)
		drawPolygonOutline(dst, csp, 1.5, hullOutlineClr)
		// Active selection ring on deck
		if b.IsActive {
			drawPolygonOutline(dst, dsp, 2, activeOutline)
		}
	}
	drawBoatMeta(dst, b, vp)
}

// drawBoatMeta draws the CoM dot and cleat indicators positioned on the hull outline.
func drawBoatMeta(dst *ebiten.Image, b *sim.Boat, vp Viewport) {
	// CoM: large red dot
	com := b.Body.ComWorldPos()
	cx, cy := vp.WorldToScreen(com)
	drawCircle(dst, float32(cx), float32(cy), 8, comDotClr)

	switch b.Type {
	case sim.BoatMonohull:
		hl := b.HullLen / 2
		hb := b.HullBeam / 2
		// Bow: single centre cleat at hull tip
		drawCleatAt(dst, b, vp, physics.Vec2{X: +hl, Y: 0})
		// Midships: bilateral, at max-beam points
		drawCleatAt(dst, b, vp, physics.Vec2{X: -hl * 0.10, Y: +hb})
		drawCleatAt(dst, b, vp, physics.Vec2{X: -hl * 0.10, Y: -hb})
		// Stern: bilateral, at stern shoulder
		drawCleatAt(dst, b, vp, physics.Vec2{X: -hl * 0.65, Y: +hb * 0.85})
		drawCleatAt(dst, b, vp, physics.Vec2{X: -hl * 0.65, Y: -hb * 0.85})

	case sim.BoatCatamaran:
		hl, sp, hw, _, _, _ := catConstants(b)
		// Bow: on each hull bow tip
		drawCleatAt(dst, b, vp, physics.Vec2{X: +hl, Y: +sp})
		drawCleatAt(dst, b, vp, physics.Vec2{X: +hl, Y: -sp})
		// Midships: outer hull edges
		drawCleatAt(dst, b, vp, physics.Vec2{X: 0, Y: +(sp + hw)})
		drawCleatAt(dst, b, vp, physics.Vec2{X: 0, Y: -(sp + hw)})
		// Stern: on each hull stern tip
		drawCleatAt(dst, b, vp, physics.Vec2{X: -hl * 0.95, Y: +sp})
		drawCleatAt(dst, b, vp, physics.Vec2{X: -hl * 0.95, Y: -sp})
	}
}

// drawCleatAt draws one cleat indicator at the given body-frame offset from CoM.
func drawCleatAt(dst *ebiten.Image, b *sim.Boat, vp Viewport, bodyOffset physics.Vec2) {
	com := b.Body.ComWorldPos()
	worldPos := com.Add(bodyOffset.Rotate(b.Body.Heading))
	sx, sy := vp.WorldToScreen(worldPos)
	drawCircle(dst, float32(sx), float32(sy), 5, cleatOuter)
	drawCircle(dst, float32(sx), float32(sy), 2.5, cleatInner)
}

// DrawBoatIcon draws a top-down boat silhouette for the left panel.
// cx, cy: centre; size: half-height in pixels; bow points upward (screen).
func DrawBoatIcon(dst *ebiten.Image, boatType sim.BoatType, cx, cy, size float32) {
	switch boatType {
	case sim.BoatMonohull:
		drawMonohullIcon(dst, cx, cy, size)
	case sim.BoatCatamaran:
		drawCatamaranIcon(dst, cx, cy, size)
	}
}

func drawMonohullIcon(dst *ebiten.Image, cx, cy, size float32) {
	// Matches monohullWorldPts proportions; bow points up (-Y screen).
	pts := []physics.Vec2{
		{X: float64(cx), Y: float64(cy - size)},                     // bow tip
		{X: float64(cx + size*0.50), Y: float64(cy - size*0.35)},   // bow stbd
		{X: float64(cx + size*0.50), Y: float64(cy + size*0.10)},   // max-beam stbd
		{X: float64(cx + size*0.42), Y: float64(cy + size*0.60)},   // stern stbd shoulder
		{X: float64(cx), Y: float64(cy + size*0.78)},                // stern centre
		{X: float64(cx - size*0.42), Y: float64(cy + size*0.60)},   // stern port shoulder
		{X: float64(cx - size*0.50), Y: float64(cy + size*0.10)},   // max-beam port
		{X: float64(cx - size*0.50), Y: float64(cy - size*0.35)},   // bow port
	}
	drawFilledPolygon(dst, pts, hullWhite)
	drawPolygonOutline(dst, pts, 3, hullOutlineClr)
	cabin := []physics.Vec2{
		{X: float64(cx + size*0.26), Y: float64(cy - size*0.05)},
		{X: float64(cx - size*0.28), Y: float64(cy - size*0.05)},
		{X: float64(cx - size*0.28), Y: float64(cy + size*0.48)},
		{X: float64(cx + size*0.26), Y: float64(cy + size*0.48)},
	}
	drawFilledPolygon(dst, cabin, cabinFill)
	drawPolygonOutline(dst, cabin, 1.5, cabinBorder)
}

func drawCatamaranIcon(dst *ebiten.Image, cx, cy, size float32) {
	spF := size * 0.38 // hull centreline offset from icon centre
	hwF := size * 0.16 // hull half-width
	dwF := spF - hwF   // deck half-width (inner hull edge)

	for _, sign := range []float32{-1, +1} {
		hx := cx + sign*spF    // hull centreline X
		outer := hx + sign*hwF // outer hull edge X
		inner := cx + sign*dwF // inner hull edge X (deck edge)

		hull := []physics.Vec2{
			{X: float64(hx), Y: float64(cy - size)},                             // bow tip
			{X: float64(outer), Y: float64(cy - size*0.55)},                     // outer bow shoulder
			{X: float64(outer), Y: float64(cy + size*0.28)},                     // outer max-beam
			{X: float64(cx + sign*(spF+hwF*0.60)), Y: float64(cy + size*0.65)},  // outer stern shoulder
			{X: float64(hx), Y: float64(cy + size*0.82)},                        // stern tip
			{X: float64(cx + sign*(spF-hwF*0.60)), Y: float64(cy + size*0.65)},  // inner stern shoulder
			{X: float64(inner), Y: float64(cy + size*0.52)},                     // deck back corner
			{X: float64(inner), Y: float64(cy - size*0.68)},                     // deck fore corner
		}
		drawFilledPolygon(dst, hull, hullWhite)
		drawPolygonOutline(dst, hull, 2.5, hullOutlineClr)
	}

	// Deck bridge
	deck := []physics.Vec2{
		{X: float64(cx + dwF), Y: float64(cy - size*0.68)},
		{X: float64(cx - dwF), Y: float64(cy - size*0.68)},
		{X: float64(cx - dwF), Y: float64(cy + size*0.52)},
		{X: float64(cx + dwF), Y: float64(cy + size*0.52)},
	}
	drawFilledPolygon(dst, deck, hullWhite)
	drawPolygonOutline(dst, deck, 2, hullOutlineClr)
}
