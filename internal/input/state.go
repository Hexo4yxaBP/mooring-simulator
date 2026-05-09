package input

// InputState captures all raw mouse/wheel input for one game tick.
// Populated by main.go from the Ebiten input APIs; no Ebiten import needed here.
type InputState struct {
	MouseX, MouseY   int
	LeftJustPressed  bool
	LeftHeld         bool
	LeftJustReleased bool
	RightJustPressed bool
	WheelDY          float64 // positive = scroll up
}
