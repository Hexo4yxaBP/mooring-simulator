package render

import (
	"image"
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"

	"github.com/dvpalkin/mooring-simulator/internal/physics"
)

// emptyImage and emptySubImage are initialised lazily on first Draw call
// to avoid panicking when the Ebiten graphics context is not yet available (e.g. in tests).
var emptyImage *ebiten.Image
var emptySubImage *ebiten.Image

func lazySubImage() *ebiten.Image {
	if emptySubImage == nil {
		emptyImage = ebiten.NewImage(3, 3)
		emptyImage.Fill(color.White)
		emptySubImage = emptyImage.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
	}
	return emptySubImage
}

// drawFilledPolygon draws a filled convex polygon on dst.
func drawFilledPolygon(dst *ebiten.Image, pts []physics.Vec2, clr color.RGBA) {
	if len(pts) < 3 {
		return
	}
	var path vector.Path
	path.MoveTo(float32(pts[0].X), float32(pts[0].Y))
	for _, p := range pts[1:] {
		path.LineTo(float32(p.X), float32(p.Y))
	}
	path.Close()

	vs, is := path.AppendVerticesAndIndicesForFilling(nil, nil)
	cr := float32(clr.R) / 255
	cg := float32(clr.G) / 255
	cb := float32(clr.B) / 255
	ca := float32(clr.A) / 255
	for i := range vs {
		vs[i].SrcX = 1
		vs[i].SrcY = 1
		vs[i].ColorR = cr
		vs[i].ColorG = cg
		vs[i].ColorB = cb
		vs[i].ColorA = ca
	}
	dst.DrawTriangles(vs, is, lazySubImage(), nil)
}

// drawPolygonOutline strokes the edges of a polygon.
func drawPolygonOutline(dst *ebiten.Image, pts []physics.Vec2, strokeW float32, clr color.RGBA) {
	if len(pts) < 2 {
		return
	}
	n := len(pts)
	for i := range pts {
		a := pts[i]
		b := pts[(i+1)%n]
		vector.StrokeLine(dst,
			float32(a.X), float32(a.Y),
			float32(b.X), float32(b.Y),
			strokeW, clr, false)
	}
}

// drawCircle draws a filled circle at screen-space position (cx, cy).
func drawCircle(dst *ebiten.Image, cx, cy, r float32, clr color.RGBA) {
	vector.DrawFilledCircle(dst, cx, cy, r, clr, false)
}
