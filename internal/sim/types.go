package sim

// ThrottleState represents one of five discrete engine throttle positions.
// Constraint C6: exactly 5 states.
type ThrottleState int

const (
	ThrottleNeutral    ThrottleState = 0
	ThrottleSlowFwd    ThrottleState = 1
	ThrottleFullFwd    ThrottleState = 2
	ThrottleSlowAstern ThrottleState = 3
	ThrottleFullAstern ThrottleState = 4
)

var throttleLabels = [5]string{"Neutral", "Slow Fwd", "Full Fwd", "Slow Astern", "Full Astern"}

func (t ThrottleState) String() string {
	if !IsValidThrottle(t) {
		return "Unknown"
	}
	return throttleLabels[t]
}

// CleatID identifies one of three boat attachment points.
// Constraint C4: only Bow, Midships, or Stern.
type CleatID int

const (
	CleatBow      CleatID = 0
	CleatMidships CleatID = 1
	CleatStern    CleatID = 2
)

// BoatType distinguishes monohull from catamaran.
type BoatType int

const (
	BoatMonohull  BoatType = 0
	BoatCatamaran BoatType = 1
)

// IsValidThrottle reports whether t is a defined ThrottleState.
func IsValidThrottle(t ThrottleState) bool { return t >= ThrottleNeutral && t <= ThrottleFullAstern }

// IsValidCleat reports whether c is a defined CleatID.
func IsValidCleat(c CleatID) bool { return c >= CleatBow && c <= CleatStern }
