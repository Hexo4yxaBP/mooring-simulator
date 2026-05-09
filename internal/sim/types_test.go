package sim

import "testing"

func TestThrottleValues(t *testing.T) {
	if ThrottleNeutral != 0 || ThrottleSlowFwd != 1 || ThrottleFullFwd != 2 ||
		ThrottleSlowAstern != 3 || ThrottleFullAstern != 4 {
		t.Error("ThrottleState enum values out of spec")
	}
}

func TestIsValidThrottle(t *testing.T) {
	for i := ThrottleState(0); i <= ThrottleFullAstern; i++ {
		if !IsValidThrottle(i) {
			t.Errorf("IsValidThrottle(%d) should be true", i)
		}
	}
	if IsValidThrottle(-1) || IsValidThrottle(5) {
		t.Error("Out-of-range throttle should be invalid")
	}
}

func TestCleatValues(t *testing.T) {
	if CleatBow != 0 || CleatMidships != 1 || CleatStern != 2 {
		t.Error("CleatID enum values out of spec")
	}
}

func TestIsValidCleat(t *testing.T) {
	for i := CleatID(0); i <= CleatStern; i++ {
		if !IsValidCleat(i) {
			t.Errorf("IsValidCleat(%d) should be true", i)
		}
	}
	if IsValidCleat(-1) || IsValidCleat(3) {
		t.Error("Out-of-range cleat should be invalid")
	}
}

func TestThrottleLabels(t *testing.T) {
	cases := map[ThrottleState]string{
		ThrottleNeutral:    "Neutral",
		ThrottleSlowFwd:    "Slow Fwd",
		ThrottleFullFwd:    "Full Fwd",
		ThrottleSlowAstern: "Slow Astern",
		ThrottleFullAstern: "Full Astern",
	}
	for state, want := range cases {
		if got := state.String(); got != want {
			t.Errorf("ThrottleState(%d).String() = %q want %q", state, got, want)
		}
	}
}
