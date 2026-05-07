# Interface Specification — Internal Package Contracts

*This app has no HTTP API. Interfaces are Go package boundaries. Each section defines one package's public surface: exported types, method signatures, and constraints.*

*Sources: domain-model.md (entity fields, force models), constraints.md (C4/C6, P1–P6), architecture.md (D8/D9).*

---

## Package: `internal/physics`

Pure math. No Ebiten. No game state. Stateless functions + value types only.

### Vec2

```go
type Vec2 struct {
    X, Y float64  // meters (or any consistent unit)
}

func (v Vec2) Add(u Vec2) Vec2
func (v Vec2) Sub(u Vec2) Vec2
func (v Vec2) Scale(s float64) Vec2
func (v Vec2) Dot(u Vec2) float64
func (v Vec2) Cross(u Vec2) float64   // scalar cross: v.X*u.Y - v.Y*u.X
func (v Vec2) Len() float64
func (v Vec2) Normalize() Vec2        // panics if len==0; caller must guard
func (v Vec2) Rotate(angle float64) Vec2  // CCW rotation by angle radians
```

**Constraint:** Normalize on zero vector is caller error — no silent fallback.

---

### RigidBody

```go
type RigidBody struct {
    Position   Vec2
    Velocity   Vec2
    Heading    float64   // radians, CCW from +x
    AngularVel float64   // rad/s, CCW positive
    Mass       float64   // kg, must be > 0
    InertiaI   float64   // kg·m², must be > 0
    ComOffset  Vec2      // CoM relative to geometric center, body-frame meters
}
```

---

### ForceAccumulator

```go
type ForceAccumulator struct {
    Force  Vec2    // world-space, Newtons
    Torque float64 // N·m, CCW positive
}

func (a *ForceAccumulator) Apply(force Vec2, worldPoint Vec2, comWorldPos Vec2)
// Adds force to Force; adds cross(worldPoint - comWorldPos, force) to Torque.
// worldPoint: where force is applied (e.g. cleat position)
// comWorldPos: center of mass in world space

func (a *ForceAccumulator) Reset()
```

---

### Integrate

```go
// Integrate advances body by one timestep using semi-implicit Euler.
// dt must be > 0. Caller owns synchronization if called concurrently.
func Integrate(body *RigidBody, acc ForceAccumulator, dt float64)
```

Implementation:
```
a = acc.Force / body.Mass
α = acc.Torque / body.InertiaI
body.Velocity    += a * dt          // update velocity first (semi-implicit)
body.AngularVel  += α * dt
body.Position    += body.Velocity * dt
body.Heading     += body.AngularVel * dt
```

---

### Force Calculators (stateless functions)

All return a `(force Vec2, applicationPoint Vec2)` pair so the caller passes to ForceAccumulator.Apply().

```go
// WindForce returns wind force on a hull in world space.
// windVel: wind velocity vector (world-space, m/s)
// heading: boat heading (radians)
// longCoeff, latCoeff: drag coefficients (tunable, m² equivalent area)
func WindForce(windVel Vec2, heading float64, longCoeff, latCoeff float64) Vec2
// Applied at hull geometric center (caller passes comWorldPos as applicationPoint)

// ThrustForce returns engine thrust vector.
// heading: boat heading; thrustN: signed Newtons (+fwd, -astern)
func ThrustForce(heading float64, thrustN float64) Vec2
// Applied at stern (caller computes stern world position)

// PropWalkForce returns lateral prop-walk force vector.
// heading: boat heading; walkN: signed Newtons (+ = port, - = stbd)
func PropWalkForce(heading float64, walkN float64) Vec2
// Applied at stern

// RudderForce returns rudder lateral force vector.
// heading: boat heading; rudderAngle: radians (+stbd); bodyVelFwd: m/s along bow axis
// coeff: rudder force coefficient
// Constraint P5: if abs(bodyVelFwd) < 0.05 m/s, returns zero vector
func RudderForce(heading float64, rudderAngle float64, bodyVelFwd float64, coeff float64) Vec2
// Applied at rudder position (caller computes rudder world position, near stern)

// HydroDragForce returns hydrodynamic drag forces in world space.
// Returns (linearDrag Vec2, torqueDrag float64) — torqueDrag applied directly to accumulator
// Constraint P3: latCoeff >> fwdCoeff
func HydroDragForce(vel Vec2, heading float64, angularVel float64,
    fwdCoeff, latCoeff, rotCoeff float64) (Vec2, float64)
```

---

### SpringForce

```go
// SpringForce computes mooring line tension force.
// anchorPos: dock attachment point (world-space)
// cleatPos: boat cleat position (world-space)
// cleatVel: velocity of cleat point (world-space) = body.Velocity + ω × r
// naturalLen: rest length (meters); stiffness: N/m; damping: N·s/m
// Constraint P1: returns zero if currentLen <= naturalLen (tension-only)
func SpringForce(anchorPos, cleatPos, cleatVel Vec2,
    naturalLen, stiffness, damping float64) Vec2
```

---

### Collision

```go
// CollisionPenalty detects overlap between two convex polygons and returns
// a penalty spring force pushing polyA out of polyB.
// verticesA, verticesB: polygon vertices in world-space, CCW winding
// penaltyStiffness: N/m (should be >> mooring stiffness to prevent visible penetration)
// Returns (force on A, contact point) — caller applies equal/opposite to B if B is dynamic.
// Returns zero force if no overlap.
func CollisionPenalty(verticesA, verticesB []Vec2, penaltyStiffness float64) (Vec2, Vec2)
```

Internally uses SAT (Separating Axis Theorem). Handles degenerate case (zero-overlap) gracefully.

---

## Package: `internal/sim`

Game entity state. No Ebiten. Owns physics step.

### Enums and Constants

```go
type ThrottleState int
const (
    ThrottleNeutral    ThrottleState = 0
    ThrottleSlowFwd    ThrottleState = 1
    ThrottleFullFwd    ThrottleState = 2
    ThrottleSlowAstern ThrottleState = 3
    ThrottleFullAstern ThrottleState = 4
)
// Constraint C6: exactly 5 states

type CleatID int
const (
    CleatBow      CleatID = 0
    CleatMidships CleatID = 1
    CleatStern    CleatID = 2
)
// Constraint C4: only these 3 attachment points on boat

type BoatType int
const (
    BoatMonohull  BoatType = 0
    BoatCatamaran BoatType = 1
)
```

---

### Boat

```go
type Boat struct {
    ID       int               // unique, assigned at creation, immutable
    Type     BoatType
    Body     physics.RigidBody
    HullLen  float64           // meters, bow-to-stern
    HullBeam float64           // meters, max width
    // Cleat offsets in body frame (bow=fwd, stern=aft), relative to CoM
    Cleats   [3]physics.Vec2   // index: CleatBow, CleatMidships, CleatStern

    // Active-boat controls (read by sim.World.Step; ignored for catamaran MVP)
    Throttle  ThrottleState
    PropWalk  float64  // N per unit; sign = handedness. Default: see constants.go
    Rudder    float64  // radians, ±MaxRudderAngle
    IsActive  bool

    // Derived / cached (updated each step, not user-set)
    HullVertices [4]physics.Vec2  // world-space corners of hull bounding box
}

// CleatWorldPos returns the world-space position of a cleat.
func (b *Boat) CleatWorldPos(id CleatID) physics.Vec2

// CleatVelocity returns the world-space velocity of a cleat point (body vel + ω × r).
func (b *Boat) CleatVelocity(id CleatID) physics.Vec2

// Constraints:
//   HullLen  ∈ (0, 100] meters
//   HullBeam ∈ (0, 20]  meters
//   Rudder   ∈ [-MaxRudderAngle, +MaxRudderAngle] (clamped on set, default ±35°)
//   Mass     > 0 (set via Body.Mass)
```

---

### Dock

```go
type Dock struct {
    Vertices []physics.Vec2  // polygon, world-space, CCW winding, min 4 points
    // For straight pier MVP: exactly 4 vertices (axis-aligned or rotated rectangle)
}

// NewStraightPier creates a dock aligned along the top edge of the world.
// pos: center of pier face (the boat-facing edge), width: pier length, depth: pier thickness
func NewStraightPier(pos physics.Vec2, width, depth float64) Dock

// NearestPointOnEdge returns the world-space point on the dock boundary
// closest to p. Used for dock-end mooring line placement (constraint C5).
func (d *Dock) NearestPointOnEdge(p physics.Vec2) physics.Vec2
```

---

### MooringLine

```go
type MooringLine struct {
    DockPoint     physics.Vec2  // world-space, on dock boundary
    BoatID        int
    Cleat         CleatID
    NaturalLength float64   // meters; set to 0.95 * initialDist at placement
    Stiffness     float64   // N/m; default from constants.go
    Damping       float64   // N·s/m; default from constants.go
}

// CurrentLength computes the present line length given boat state.
func (l *MooringLine) CurrentLength(boats []*Boat) float64

// Tension returns the current tension force magnitude (0 if slack).
// Constraint P1: non-negative always.
func (l *MooringLine) Tension(boats []*Boat) float64
```

---

### WindField

```go
type WindField struct {
    Speed     float64  // m/s, ∈ [0, 30]
    Direction float64  // radians, wind-FROM direction (0 = from +x / east)
}

func (w WindField) Velocity() physics.Vec2  // wind velocity vector (world-space, m/s)
```

---

### World

```go
type World struct {
    Boats []*Boat
    Dock  Dock
    Lines []MooringLine
    Wind  WindField
    Time  float64  // simulation seconds since start
}

// Step advances the simulation by dt seconds.
// Applies all forces to all boats, detects and resolves collisions, integrates.
// dt should be 1/60 or configured fixed timestep; must be > 0.
func (w *World) Step(dt float64)

// ApplyCommand mutates world state (throttle, rudder, wind, active boat, mooring lines).
func (w *World) ApplyCommand(cmd input.Command)

// ActiveBoat returns the currently active boat, or nil if none.
func (w *World) ActiveBoat() *Boat

// SetActiveBoat changes the active boat by ID. No-op if ID not found.
func (w *World) SetActiveBoat(id int)

// AddMooringLine validates and appends a line. Returns error if BoatID invalid,
// CleatID out of range, or dock point is not on dock boundary.
func (w *World) AddMooringLine(line MooringLine) error

// RemoveMooringLine removes line by index. Index out of range = no-op.
func (w *World) RemoveMooringLine(idx int)
```

---

## Package: `internal/input`

Maps Ebiten input state → typed Commands. No physics, no rendering.

### Command

```go
type CommandType int
const (
    CmdSetThrottle    CommandType = iota  // payload: ThrottlePayload
    CmdSetRudder                          // payload: RudderPayload
    CmdSetWind                            // payload: WindPayload
    CmdSelectBoat                         // payload: SelectBoatPayload
    CmdAddMooringLine                     // payload: AddLinePayload
    CmdRemoveMooringLine                  // payload: int (line index)
    CmdAddBoat                            // payload: AddBoatPayload
)

type Command struct {
    Type    CommandType
    Payload any
}

type ThrottlePayload  struct{ State sim.ThrottleState }
type RudderPayload    struct{ Angle float64 }            // radians, clamped externally
type WindPayload      struct{ Speed, Direction float64 } // m/s, radians
type SelectBoatPayload struct{ BoatID int }
type AddLinePayload   struct {
    DockPoint physics.Vec2
    BoatID    int
    Cleat     sim.CleatID
}
type AddBoatPayload   struct {
    Type     sim.BoatType
    Position physics.Vec2
    Heading  float64
}
```

### Handler

```go
type Handler struct{ /* ebiten state */ }

// Poll reads current Ebiten input state and returns zero or more Commands.
// Called once per Update() tick. Returns nil slice (not error) on no input.
func (h *Handler) Poll(mouseX, mouseY int, world *sim.World) []Command
```

**Key mappings (defaults):**

| Input | Command |
|-------|---------|
| `W` / `S` | Throttle up/down (cycles through 5 states) |
| `A` / `D` | Rudder left/right (±5° per tick, held = continuous) |
| `0`..`4` numpad | Set throttle directly |
| Click on boat | SelectBoat |
| Click on dock edge then boat cleat | AddMooringLine (two-step interaction) |
| Right-click on mooring line | RemoveMooringLine |
| Mouse wheel on wind panel | SetWind speed |

---

## Package: `internal/render`

Reads sim.World, produces Ebiten draw calls. No state mutation.

### Viewport

```go
type Viewport struct {
    OriginWorld physics.Vec2  // world-space point at screen top-left
    Scale       float64       // pixels per meter; default 20
    ScreenW     int
    ScreenH     int
}

func (v Viewport) WorldToScreen(p physics.Vec2) (x, y float64)
func (v Viewport) ScreenToWorld(x, y float64) physics.Vec2
```

### Renderer

```go
type Renderer struct {
    VP Viewport
}

// Draw renders the full world state onto screen.
func (r *Renderer) Draw(screen *ebiten.Image, world *sim.World)
```

Draw order (painter's algorithm, back to front):
1. Water background (solid fill)
2. Dock (grey filled polygon)
3. Mooring lines (color by tension: slack=grey, taut=yellow, max=red)
4. Boats (hull polygon + CoM dot + cleat dots)
5. Active boat highlight (bright outline)
6. Placement preview (ghost line during mooring line placement)

---

## Package: `internal/ui`

On-canvas panels rendered via Ebiten. Reads world for display; emits Commands via returned slice.

```go
type HUD struct{ /* panel layout state */ }

// Draw renders all UI panels and returns any Commands triggered by user interaction.
func (h *HUD) Draw(screen *ebiten.Image, world *sim.World) []Command
```

**Panels (MVP):**
- Top-left: active boat info (speed, heading, throttle indicator)
- Top-right: wind controls (speed + direction dial)
- Bottom: rudder visual indicator
- Hover: mooring line tension readout on mouse-over
