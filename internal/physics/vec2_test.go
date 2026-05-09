package physics

import (
	"math"
	"testing"
)

func TestVec2Add(t *testing.T) {
	got := Vec2{X: 1, Y: 2}.Add(Vec2{X: 3, Y: 4})
	want := Vec2{X: 4, Y: 6}
	if got != want {
		t.Errorf("Add: got %v want %v", got, want)
	}
}

func TestVec2Sub(t *testing.T) {
	got := Vec2{X: 5, Y: 3}.Sub(Vec2{X: 2, Y: 1})
	want := Vec2{X: 3, Y: 2}
	if got != want {
		t.Errorf("Sub: got %v want %v", got, want)
	}
}

func TestVec2Scale(t *testing.T) {
	got := Vec2{X: 2, Y: 3}.Scale(2)
	want := Vec2{X: 4, Y: 6}
	if got != want {
		t.Errorf("Scale: got %v want %v", got, want)
	}
}

func TestVec2Dot(t *testing.T) {
	got := Vec2{X: 1, Y: 2}.Dot(Vec2{X: 3, Y: 4})
	if got != 11 {
		t.Errorf("Dot: got %v want 11", got)
	}
}

func TestVec2Cross(t *testing.T) {
	got := Vec2{X: 1, Y: 0}.Cross(Vec2{X: 0, Y: 1})
	if got != 1.0 {
		t.Errorf("Cross: got %v want 1.0", got)
	}
	got2 := Vec2{X: 0, Y: 1}.Cross(Vec2{X: 1, Y: 0})
	if got2 != -1.0 {
		t.Errorf("Cross reverse: got %v want -1.0", got2)
	}
}

func TestVec2Len(t *testing.T) {
	got := Vec2{X: 3, Y: 4}.Len()
	if math.Abs(got-5) > 1e-9 {
		t.Errorf("Len: got %v want 5", got)
	}
}

func TestVec2Normalize(t *testing.T) {
	got := Vec2{X: 3, Y: 0}.Normalize()
	want := Vec2{X: 1, Y: 0}
	if got != want {
		t.Errorf("Normalize: got %v want %v", got, want)
	}
}

func TestVec2NormalizeZeroPanics(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Error("Normalize(zero) did not panic")
		}
	}()
	Vec2{0, 0}.Normalize()
}

func TestVec2Rotate(t *testing.T) {
	got := Vec2{X: 1, Y: 0}.Rotate(math.Pi / 2)
	if math.Abs(got.X-0) > 1e-9 || math.Abs(got.Y-1) > 1e-9 {
		t.Errorf("Rotate Pi/2: got %v want {0,1}", got)
	}

	got2 := Vec2{X: 1, Y: 0}.Rotate(math.Pi)
	if math.Abs(got2.X-(-1)) > 1e-9 || math.Abs(got2.Y-0) > 1e-9 {
		t.Errorf("Rotate Pi: got %v want {-1,0}", got2)
	}
}

func TestVec2RotateIdentity(t *testing.T) {
	v := Vec2{X: 3, Y: 4}
	got := v.Rotate(0)
	if math.Abs(got.X-v.X) > 1e-9 || math.Abs(got.Y-v.Y) > 1e-9 {
		t.Errorf("Rotate 0: got %v want %v", got, v)
	}
}

func TestVec2Perp(t *testing.T) {
	got := Vec2{X: 1, Y: 0}.Perp()
	want := Vec2{X: 0, Y: 1}
	if got != want {
		t.Errorf("Perp: got %v want %v", got, want)
	}
}

func approxEq(a, b, eps float64) bool {
	return math.Abs(a-b) <= eps
}
