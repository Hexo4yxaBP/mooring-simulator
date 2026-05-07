# Implementation Tasks

*Decomposed from docs/Design/architecture.md, interface-spec.md, data-schemas.md.*
*Dependency rule: physics ← sim ← main → render, input, ui (architecture.md C4 Container).*

---

## Group 0 — Project Skeleton

### TASK-001: Initialize Go module and directory structure
**Depends on:** none
**Acceptance criteria:**
- [ ] `go.mod` exists with module path `github.com/dvpalkin/mooring-simulator` and Go 1.22+
- [ ] `go get github.com/hajimehoshi/ebiten/v2` succeeds; entry appears in `go.sum`
- [ ] Directories `internal/physics`, `internal/sim`, `internal/render`, `internal/input`, `internal/ui` exist
- [ ] `go build ./...` compiles with zero errors on native (non-WASM) target
**Definition of done:** all criteria checked + empty package stubs compile

---

### TASK-002: Create WASM build pipeline
**Depends on:** TASK-001
**Acceptance criteria:**
- [ ] `index.html` loads `wasm_exec.js` and `main.wasm` with correct MIME types
- [ ] `wasm_exec.js` is copied from `$(go env GOROOT)/misc/wasm/wasm_exec.js`
- [ ] `Makefile` target `build-wasm` runs `GOOS=js GOARCH=wasm go build -o main.wasm .` without error
- [ ] `Makefile` target `serve` starts a local HTTP server on port 8080 (e.g. `go run tools/serve.go`)
- [ ] Opening `http://localhost:8080` in a browser shows a blank page without JS console errors
**Definition of done:** all criteria checked; no unit test required (build artifact)

---

### TASK-003: Minimal Ebiten game loop
**Depends on:** TASK-001
**Acceptance criteria:**
- [ ] `main.go` defines a struct implementing `ebiten.Game` (`Update`, `Draw`, `Layout`)
- [ ] `Update()` returns nil; `Draw()` fills screen with a solid colour; `Layout()` returns 1280×720
- [ ] Native binary (`go run .`) opens a 1280×720 window with no crash for 5 seconds
- [ ] `go vet ./...` passes
**Definition of done:** all criteria checked; no unit test required (visual artifact)

---

## Group 1 — internal/physics

### TASK-010: Vec2 type and math helpers
**Depends on:** TASK-001
**Acceptance criteria:**
- [ ] `Vec2{X, Y float64}` struct defined in `internal/physics/vec2.go`
- [ ] `Add`, `Sub`, `Scale`, `Dot`, `Cross`, `Len`, `Normalize`, `Rotate` methods match signatures in interface-spec.md
- [ ] `Normalize(Vec2{0,0})` panics (caller-error contract from interface-spec.md)
- [ ] `Rotate(Vec2{1,0}, math.Pi/2)` returns approximately `Vec2{0,1}` (within 1e-9)
- [ ] `Cross(Vec2{1,0}, Vec2{0,1})` returns `1.0`
**Definition of done:** all criteria checked + unit tests in `internal/physics/vec2_test.go` cover all methods

---

### TASK-011: RigidBody and ForceAccumulator types
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `RigidBody` struct defined with all fields from interface-spec.md
- [ ] `ForceAccumulator.Apply(force, worldPoint, comWorldPos)` adds `force` to `acc.Force` and adds `cross(worldPoint-comWorldPos, force)` to `acc.Torque`
- [ ] `ForceAccumulator.Reset()` sets Force to zero vector and Torque to 0
- [ ] Applying a force at the CoM produces zero torque
- [ ] Applying a force offset from CoM produces non-zero torque with correct sign
**Definition of done:** all criteria checked + unit tests in `internal/physics/body_test.go`

---

### TASK-012: Integrate() — semi-implicit Euler
**Depends on:** TASK-011
**Acceptance criteria:**
- [ ] `Integrate(body, acc, dt)` signature matches interface-spec.md
- [ ] With only a forward force on a stationary body: after one step `body.Velocity.X > 0`, `body.Position.X > 0`
- [ ] Velocity is updated before position (semi-implicit: `v += a*dt` then `p += v*dt`)
- [ ] With zero accumulator and non-zero initial velocity, body coasts: position advances by `v*dt` each step
- [ ] With a restoring torque opposing angular velocity, angular velocity decreases each step
**Definition of done:** all criteria checked + unit tests in `internal/physics/body_test.go`

---

### TASK-013: Wind, Thrust, PropWalk, Rudder force calculators
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `WindForce(windVel, heading, longCoeff, latCoeff)` — beam-on wind (`windVel` perpendicular to heading) returns force with larger lateral component than bow-on wind of same speed
- [ ] `ThrustForce(heading, thrustN)` with `thrustN > 0` returns vector in direction of `heading` (bow direction)
- [ ] `ThrustForce(heading, thrustN)` with `thrustN < 0` returns vector opposite to heading
- [ ] `PropWalkForce(heading, walkN > 0)` returns vector 90° port of heading
- [ ] `RudderForce(heading, rudderAngle, bodyVelFwd < 0.05, coeff)` returns `Vec2{0,0}` (constraint P5)
- [ ] `RudderForce(heading, rudderAngle > 0, bodyVelFwd > 0.05, coeff)` returns force with starboard component
**Definition of done:** all criteria checked + unit tests in `internal/physics/forces_test.go`

---

### TASK-014: HydroDragForce calculator
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `HydroDragForce(vel, heading, angularVel, fwdCoeff, latCoeff, rotCoeff)` matches interface-spec.md signature
- [ ] Linear drag force opposes velocity direction
- [ ] With `latCoeff = 20 * fwdCoeff`, a beam-on velocity produces 20× more drag force than an equal forward velocity
- [ ] Torque drag magnitude equals `rotCoeff * abs(angularVel)` and opposes sign of `angularVel`
- [ ] Zero velocity + zero angular velocity returns `(Vec2{0,0}, 0)`
**Definition of done:** all criteria checked + unit tests in `internal/physics/forces_test.go`

---

### TASK-015: SpringForce — mooring line tension
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `SpringForce(anchorPos, cleatPos, cleatVel, naturalLen, stiffness, damping)` returns `Vec2{0,0}` when `Len(cleatPos-anchorPos) <= naturalLen` (constraint P1, tension-only)
- [ ] With line stretched 1 m beyond natural length, stiffness 5000 N/m, zero velocity: `Len(force) == 5000`
- [ ] Force direction points from `cleatPos` toward `anchorPos` (pulling boat toward dock)
- [ ] Positive damping with cleat moving away from anchor increases force magnitude
- [ ] Positive damping with cleat moving toward anchor decreases force magnitude (but total remains ≥ 0)
**Definition of done:** all criteria checked + unit tests in `internal/physics/mooring_test.go`

---

### TASK-016: CollisionPenalty — SAT convex polygon overlap
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `CollisionPenalty(verticesA, verticesB, stiffness)` returns `Vec2{0,0}` for non-overlapping polygons
- [ ] Two identical squares centered at the same point return a non-zero force (full overlap)
- [ ] Two squares overlapping by 0.1 m: returned force magnitude equals `stiffness * 0.1`
- [ ] Force direction pushes polygon A out of polygon B (in direction of minimum penetration axis)
- [ ] Returned `contactPoint` lies on the boundary between the two polygons (within 0.01 m)
- [ ] Handles degenerate case: one polygon is a point (single vertex) without panic
**Definition of done:** all criteria checked + unit tests in `internal/physics/collision_test.go` including at least 5 test cases (no overlap, partial overlap, full overlap, rotated rectangle, edge-touching)

---

## Group 2 — internal/sim

### TASK-020: Enums, types, and validation helpers
**Depends on:** TASK-001
**Acceptance criteria:**
- [ ] `ThrottleState`, `CleatID`, `BoatType` enums defined in `internal/sim/types.go` with exactly the values from interface-spec.md
- [ ] `ThrottleState` has exactly 5 values: 0=Neutral, 1=SlowFwd, 2=FullFwd, 3=SlowAstern, 4=FullAstern (constraint C6)
- [ ] `CleatID` has exactly 3 values: 0=Bow, 1=Midships, 2=Stern (constraint C4)
- [ ] `IsValidThrottle(t ThrottleState) bool` returns true iff `t ∈ [0,4]`
- [ ] `IsValidCleat(c CleatID) bool` returns true iff `c ∈ [0,2]`
**Definition of done:** all criteria checked + unit tests in `internal/sim/types_test.go`

---

### TASK-021: Physics constants file
**Depends on:** TASK-020
**Acceptance criteria:**
- [ ] `internal/sim/constants.go` defines all numeric constants from data-schemas.md "Boundary 3"
- [ ] `ThrustTable [5]float64` entries match data-schemas.md Thrust Table (Neutral=0, SlowFwd=800, FullFwd=8000, SlowAstern=-600, FullAstern=-5000)
- [ ] `PropWalkMultiplier [5]float64` entries match data-schemas.md Prop Walk Table
- [ ] `MonohullDefaults` and `CatamaranDefaults` structs hold all default boat parameters from data-schemas.md
- [ ] `DefaultMooringStiffness`, `DefaultMooringDamping`, `CollisionPenaltyStiffness` constants defined
- [ ] File compiles; `go vet` passes
**Definition of done:** all criteria checked; no unit test (constants only)

---

### TASK-022: Boat struct, constructors, and helpers
**Depends on:** TASK-010, TASK-020, TASK-021
**Acceptance criteria:**
- [ ] `Boat` struct defined with all fields from interface-spec.md
- [ ] `NewMonohull(id int, pos physics.Vec2, heading float64) *Boat` returns boat with defaults from `MonohullDefaults` (mass=5000, len=10, beam=3, cleats at ±4.5 m)
- [ ] `NewCatamaran(id int, pos physics.Vec2, heading float64) *Boat` returns passive boat (PropWalk=0, IsActive=false)
- [ ] `CleatWorldPos(CleatBow)` for heading=0 returns a point forward of `body.Position`
- [ ] `CleatWorldPos(CleatStern)` for heading=0 returns a point aft of `body.Position`
- [ ] `CleatVelocity(id)` = `body.Velocity + angularVel × (cleatWorldPos - comWorldPos)` (vector cross product in 2D)
- [ ] Setting `Rudder` to a value outside `±MaxRudderAngle` via a setter clamps to the limit
**Definition of done:** all criteria checked + unit tests in `internal/sim/boat_test.go`

---

### TASK-023: Dock struct, straight-pier constructor, edge-snap
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `Dock.Vertices` is `[]physics.Vec2` with CCW winding
- [ ] `NewStraightPier(pos, width=50, depth=5)` returns a 4-vertex rectangle dock
- [ ] `NearestPointOnEdge(p)` — for point directly in front of dock face, returns point on dock face
- [ ] `NearestPointOnEdge(p)` — for point far from dock, returns the nearest vertex or edge point
- [ ] `NearestPointOnEdge(p)` — returned point has zero distance from any polygon edge (i.e. lies exactly on an edge segment)
**Definition of done:** all criteria checked + unit tests in `internal/sim/dock_test.go`

---

### TASK-024: MooringLine struct, length, and tension helpers
**Depends on:** TASK-010, TASK-022
**Acceptance criteria:**
- [ ] `MooringLine` struct fields match interface-spec.md
- [ ] `CurrentLength(boats)` returns correct Euclidean distance from `DockPoint` to cleat world position
- [ ] `Tension(boats)` returns 0 when `CurrentLength <= NaturalLength` (constraint P1)
- [ ] `Tension(boats)` returns `stiffness * (currentLen - naturalLen)` when `currentLen > naturalLen` (at zero velocity)
- [ ] `Tension(boats)` is always non-negative
**Definition of done:** all criteria checked + unit tests in `internal/sim/line_test.go`

---

### TASK-025: World.Step() — full force accumulation and integration loop
**Depends on:** TASK-012, TASK-013, TASK-014, TASK-015, TASK-016, TASK-022, TASK-023, TASK-024
**Acceptance criteria:**
- [ ] `World.Step(dt)` iterates over all boats, resets their force accumulators, and applies wind force to each boat
- [ ] For active monohull: thrust, prop walk, and rudder forces are applied each step using values from `ThrustTable` and `PropWalkMultiplier`
- [ ] Hydrodynamic drag is applied to every boat (including catamaran) each step
- [ ] Mooring line spring forces are applied at the correct cleat with correct torque arm
- [ ] Collision forces (boat↔dock and boat↔boat) are applied as penalty springs each step
- [ ] After calling `Integrate` for each boat: at Neutral throttle with no wind and no lines, a boat initially moving forward decelerates each step due to drag
- [ ] Boat at FullForward throttle with no drag (`latCoeff=fwdCoeff=0`) accelerates forward each step
- [ ] A moored boat (single stern line) in beam-on wind does not translate past the mooring line natural length (line holds it)
**Definition of done:** all criteria checked + integration test in `internal/sim/world_test.go`

---

### TASK-026: World.AddMooringLine and RemoveMooringLine
**Depends on:** TASK-025
**Acceptance criteria:**
- [ ] `AddMooringLine(line)` returns error if `line.BoatID` is not in `w.Boats`
- [ ] `AddMooringLine(line)` returns error if `line.Cleat` is not a valid CleatID
- [ ] `AddMooringLine(line)` succeeds for valid input; `len(w.Lines)` increases by 1
- [ ] `RemoveMooringLine(idx)` removes line at index `idx`; `len(w.Lines)` decreases by 1
- [ ] `RemoveMooringLine(-1)` and `RemoveMooringLine(len(lines))` are no-ops (no panic)
**Definition of done:** all criteria checked + unit tests in `internal/sim/world_test.go`

---

### TASK-027: World.SetActiveBoat and ActiveBoat
**Depends on:** TASK-022
**Acceptance criteria:**
- [ ] `SetActiveBoat(id)` sets `IsActive=true` on the boat with matching ID and `IsActive=false` on all others
- [ ] `SetActiveBoat(id)` with non-existent ID is a no-op; no boat becomes active
- [ ] `ActiveBoat()` returns the boat with `IsActive==true`, or nil if none
- [ ] After `SetActiveBoat(id)`, exactly one boat has `IsActive==true`
**Definition of done:** all criteria checked + unit tests in `internal/sim/world_test.go`

---

## Group 3 — internal/input

### TASK-030: Command types and payload structs
**Depends on:** TASK-020
**Acceptance criteria:**
- [ ] All `CommandType` constants defined in `internal/input/commands.go` (7 types from interface-spec.md)
- [ ] All payload structs (`ThrottlePayload`, `RudderPayload`, `WindPayload`, `SelectBoatPayload`, `AddLinePayload`, `AddBoatPayload`) defined with correct field types
- [ ] `Command{Type, Payload any}` struct defined
- [ ] File compiles; `go vet` passes
**Definition of done:** all criteria checked; no unit test (type definitions only)

---

### TASK-031: Handler — keyboard throttle and rudder mapping
**Depends on:** TASK-030, TASK-020
**Acceptance criteria:**
- [ ] Pressing `W` when throttle is `ThrottleNeutral` emits `CmdSetThrottle{SlowFwd}`
- [ ] Pressing `W` when throttle is `ThrottleFullFwd` emits nothing (already at max, no wrap)
- [ ] Pressing `S` when throttle is `ThrottleNeutral` emits `CmdSetThrottle{SlowAstern}`
- [ ] Pressing `S` when throttle is `ThrottleFullAstern` emits nothing
- [ ] Holding `A` emits `CmdSetRudder` with angle decremented by `RudderStepPerTick` each tick
- [ ] `CmdSetRudder` angle is clamped to `[-MaxRudderAngle, +MaxRudderAngle]` before being emitted
- [ ] Direct keys `0`–`4` emit the corresponding throttle state
**Definition of done:** all criteria checked + unit tests in `internal/input/handler_test.go` using a mock input state

---

### TASK-032: Handler — mouse boat selection and mooring line placement
**Depends on:** TASK-030, TASK-022, TASK-023, TASK-040
**Acceptance criteria:**
- [ ] Left-clicking within a boat's bounding polygon emits `CmdSelectBoat{BoatID}`
- [ ] Left-clicking on dock boundary (within 1 m world-space snap tolerance) enters "dock-point-selected" state; a second click within `CleatHitRadius` (0.5 m) of a cleat emits `CmdAddMooringLine`
- [ ] Second click outside all cleats cancels the placement (state resets to idle, no command emitted)
- [ ] Left-clicking with no boat/dock in range emits nothing
- [ ] `Handler` exposes `PlacementState` so the renderer can draw the ghost line preview
**Definition of done:** all criteria checked + unit tests in `internal/input/handler_test.go`

---

### TASK-033: Handler — mooring line removal and wind panel input
**Depends on:** TASK-031, TASK-030
**Acceptance criteria:**
- [ ] Right-clicking within 0.5 m world-space of a mooring line midpoint emits `CmdRemoveMooringLine{idx}`
- [ ] Mouse-wheel scroll-up on the wind speed panel area emits `CmdSetWind` with `Speed += 0.5`
- [ ] Mouse-wheel scroll-down emits `CmdSetWind` with `Speed -= 0.5`
- [ ] Wind speed emitted is clamped to `[0, 30]` m/s before command is emitted
**Definition of done:** all criteria checked + unit tests in `internal/input/handler_test.go`

---

### TASK-034: Input validation — range clamping and type safety
**Depends on:** TASK-031, TASK-032, TASK-033
**Acceptance criteria:**
- [ ] `RudderPayload.Angle` is always in `[-0.6109, +0.6109]` radians after handler processing
- [ ] `WindPayload.Speed` is always in `[0, 30]` after handler processing
- [ ] `WindPayload.Direction` is normalised to `[0, 2π)` after handler processing
- [ ] `AddLinePayload.Cleat` is always a valid `CleatID` (0–2); invalid click positions produce no command
- [ ] `ThrottlePayload.State` is always in `[0, 4]`; no command is emitted for out-of-range throttle
**Definition of done:** all criteria checked + unit tests in `internal/input/handler_test.go`; all tests pass under `-race`

---

## Group 4 — internal/render

### TASK-040: Viewport — world↔screen coordinate transforms
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `Viewport{Scale: 20, ScreenW: 1280, ScreenH: 720}`: `WorldToScreen(Vec2{0,0})` returns `(0, 720)` (world origin at bottom-left of screen)
- [ ] `WorldToScreen` and `ScreenToWorld` are exact inverses: `ScreenToWorld(WorldToScreen(p)) == p` for any `p`
- [ ] `WorldToScreen` applies Y-axis flip: higher world Y maps to lower screen Y (data-schemas.md D2)
- [ ] `WorldToScreen(Vec2{64, 36})` at scale 20 returns `(1280, 0)` (top-right corner with default 1280×720 screen)
**Definition of done:** all criteria checked + unit tests in `internal/render/viewport_test.go`

---

### TASK-041: Draw water background and dock
**Depends on:** TASK-040, TASK-023
**Acceptance criteria:**
- [ ] `Renderer.Draw()` fills the entire screen with the water colour (`#1A6B8A` or similar blue) before drawing anything else
- [ ] Dock polygon is drawn as a filled grey rectangle at the correct screen-space position
- [ ] Dock position in screen-space matches `Viewport.WorldToScreen(vertex)` for each vertex
- [ ] No Ebiten import in `internal/physics` or `internal/sim` packages (`go list -f '{{.Imports}}' ./internal/physics/...` contains no Ebiten import)
**Definition of done:** all criteria checked; visual verification required — screenshot reviewed manually; no automated test for pixel output

---

### TASK-042: Draw mooring lines with tension colour coding
**Depends on:** TASK-040, TASK-024
**Acceptance criteria:**
- [ ] Each `MooringLine` is drawn as a line segment from `DockPoint` to cleat world position
- [ ] Slack line (`tension == 0`) is drawn in grey `#888888`
- [ ] Line at 30% of MaxTension is drawn green `#88CC88`
- [ ] Line at 70% of MaxTension is drawn yellow `#CCCC00`
- [ ] Line at 100%+ of MaxTension is drawn red `#FF0000`
- [ ] `MaxTension` computation matches data-schemas.md formula: `NaturalLength * Stiffness * 0.5`
**Definition of done:** all criteria checked; tension colour thresholds verified with unit test on the colour-lookup helper function in `internal/render/lines_test.go`

---

### TASK-043: Draw boat hulls, CoM, and cleats
**Depends on:** TASK-040, TASK-022
**Acceptance criteria:**
- [ ] Monohull is drawn as a 5-vertex tapered polygon matching vertices from data-schemas.md hull polygon definition
- [ ] Catamaran is drawn as two rectangles plus a cross-beam rectangle
- [ ] CoM dot is drawn as a filled circle at `comWorldPos`, radius 4px
- [ ] Cleat dots (3 per boat) are drawn as small circles at their world-space positions transformed to screen-space
- [ ] Hull polygon vertices are correctly rotated by `boat.Body.Heading` and translated by CoM world position before viewport transform
**Definition of done:** all criteria checked; visual verification required; unit test confirms that hull vertex transform produces a non-degenerate polygon (all vertices distinct) in `internal/render/boat_test.go`

---

### TASK-044: Active boat highlight and placement ghost line
**Depends on:** TASK-043, TASK-042
**Acceptance criteria:**
- [ ] Active boat is drawn with a distinct bright outline (different colour from inactive boats)
- [ ] When `Handler.PlacementState == DockPointSelected`, a ghost line is drawn from the selected dock point to the current mouse cursor position in screen-space
- [ ] Ghost line is rendered in a semi-transparent or dashed style visually distinct from real mooring lines
- [ ] Ghost line disappears when placement is completed or cancelled
**Definition of done:** all criteria checked; visual verification required

---

## Group 5 — internal/ui

### TASK-050: HUD active boat info panel
**Depends on:** TASK-027, TASK-020
**Acceptance criteria:**
- [ ] Top-left panel shows: boat speed (m/s, 1 decimal place), heading (degrees, 0 decimal), throttle state label
- [ ] When no boat is active, panel shows "No active boat"
- [ ] Throttle label matches: `Neutral`, `Slow Fwd`, `Full Fwd`, `Slow Astern`, `Full Astern`
- [ ] Panel does not emit any Commands (read-only display)
**Definition of done:** all criteria checked; visual verification required

---

### TASK-051: Wind control panel
**Depends on:** TASK-030
**Acceptance criteria:**
- [ ] Top-right panel shows current wind speed (m/s, 1 decimal) and direction as a compass arrow
- [ ] Arrow rotates to match `WindField.Direction`
- [ ] Panel occupies a defined screen-space rectangle used by `Handler.Poll()` to scope mouse-wheel events (architecture.md "Wind Model")
- [ ] Panel emits `CmdSetWind` when mouse-wheel is scrolled within the panel bounds
**Definition of done:** all criteria checked; visual verification required

---

### TASK-052: Rudder visual indicator
**Depends on:** TASK-027
**Acceptance criteria:**
- [ ] Bottom-centre indicator shows a rudder symbol rotated by `activeBoat.Rudder` angle
- [ ] Indicator is centred at zero when `Rudder == 0`
- [ ] Indicator reaches maximum visual deflection at `±MaxRudderAngle`
- [ ] When no active boat, indicator shows at zero position
**Definition of done:** all criteria checked; visual verification required

---

### TASK-053: Mooring line tension hover readout
**Depends on:** TASK-042, TASK-024
**Acceptance criteria:**
- [ ] When mouse cursor is within 10px of any mooring line midpoint in screen-space, a tooltip shows tension in Newtons (format: `"123 N"`)
- [ ] Tooltip disappears when cursor moves away
- [ ] Multiple lines in proximity: tooltip shows the nearest line's tension
**Definition of done:** all criteria checked; visual verification required

---

## Group 6 — main.go and Deployment

### TASK-060: Wire main.go — full game loop
**Depends on:** TASK-025, TASK-026, TASK-027, TASK-031, TASK-032, TASK-033, TASK-034, TASK-044, TASK-050, TASK-051, TASK-052, TASK-053
**Acceptance criteria:**
- [ ] `main.go` `Update()` calls `handler.Poll()`, then `world.ApplyCommands()`, then `world.Step(1.0/60.0)`
- [ ] `main.go` `Draw()` calls `renderer.Draw(screen, world)` then `hud.Draw(screen, world)` and applies resulting commands to world
- [ ] No package-level mutable state outside of the `Game` struct
- [ ] `go build ./...` (native) and `GOOS=js GOARCH=wasm go build -o main.wasm .` both succeed
**Definition of done:** all criteria checked; full manual smoke test (see TASK-061)

---

### TASK-061: World initialisation — default scenario
**Depends on:** TASK-022, TASK-023, TASK-027
**Acceptance criteria:**
- [ ] On startup: one monohull boat placed 10 m from dock face, heading toward dock, set as active
- [ ] On startup: one catamaran placed 20 m from dock face, heading 90° from dock, not active
- [ ] Dock is a 50 m wide, 5 m deep straight pier centred at world `{0, 30}`
- [ ] Wind initialised to `{Speed: 3.0, Direction: π/2}` (light breeze from south)
- [ ] No mooring lines on startup
- [ ] Smoke test: run 600 simulation steps (`world.Step(1/60)` × 600 = 10 sim-seconds) without panic
**Definition of done:** all criteria checked + smoke test passes as a Go test in `internal/sim/world_test.go`

---

### TASK-062: Static file deployment setup
**Depends on:** TASK-002
**Acceptance criteria:**
- [ ] `index.html` serves the app correctly with `Content-Type: application/wasm` for `.wasm` files
- [ ] A simple Go HTTP server (`tools/serve.go`) serves files with correct MIME type
- [ ] `README.md` updated with: build command, serve command, and browser URL
- [ ] Opening the app in Chrome/Firefox runs without console errors
**Definition of done:** all criteria checked; no automated test (deployment artifact)

---

## Group 7 — Tests

### TASK-070: physics package unit tests — Vec2, RigidBody, Integrate
**Depends on:** TASK-010, TASK-011, TASK-012
**Acceptance criteria:**
- [ ] `go test ./internal/physics/...` passes with zero failures
- [ ] Test coverage for `internal/physics` ≥ 90% (`go test -cover`)
- [ ] All tests pass under `-race` flag
- [ ] Edge cases covered: zero-length vector operations, zero `dt` passed to Integrate (should not mutate), negative mass handling documented as undefined
**Definition of done:** all criteria checked

---

### TASK-071: Force calculator unit tests
**Depends on:** TASK-013, TASK-014
**Acceptance criteria:**
- [ ] All 5 force functions (`WindForce`, `ThrustForce`, `PropWalkForce`, `RudderForce`, `HydroDragForce`) have table-driven tests covering: zero input, nominal input, edge inputs
- [ ] `RudderForce` test explicitly verifies the P5 threshold (`bodyVelFwd < 0.05`)
- [ ] `HydroDragForce` test verifies 20:1 lateral:longitudinal drag ratio
- [ ] `go test ./internal/physics/...` passes; coverage ≥ 90%
**Definition of done:** all criteria checked

---

### TASK-072: SpringForce unit tests
**Depends on:** TASK-015
**Acceptance criteria:**
- [ ] Tension-only constraint P1 verified: slack line returns zero force
- [ ] Correct magnitude at 1 m extension verified
- [ ] Damping contribution verified with non-zero `cleatVel`
- [ ] Force direction verified to point anchor→cleat (pulling toward anchor)
**Definition of done:** all criteria checked; `go test ./internal/physics/...` passes

---

### TASK-073: CollisionPenalty unit tests
**Depends on:** TASK-016
**Acceptance criteria:**
- [ ] At least 5 test cases: non-overlapping, partial overlap x-axis, partial overlap y-axis, fully contained, two rotated rectangles overlapping
- [ ] Non-overlapping case returns exactly `Vec2{0,0}` for force
- [ ] Penetration depth in partial-overlap cases matches stiffness formula within 1%
- [ ] Degenerate case (single-vertex polygon) does not panic
- [ ] `go test -count=100` (fuzz-style repeat) passes without panic
**Definition of done:** all criteria checked; coverage of `collision.go` ≥ 85%

---

### TASK-074: sim package unit tests — Boat, Dock, MooringLine
**Depends on:** TASK-022, TASK-023, TASK-024
**Acceptance criteria:**
- [ ] `CleatWorldPos` and `CleatVelocity` verified for heading=0, heading=π/2, heading=π
- [ ] `NearestPointOnEdge` verified for point ahead of dock, beside dock, behind dock, at corner
- [ ] `MooringLine.Tension` verified: slack=0, extended returns correct magnitude
- [ ] `go test ./internal/sim/...` passes; coverage ≥ 80%
**Definition of done:** all criteria checked

---

### TASK-075: World.Step integration test
**Depends on:** TASK-025, TASK-061
**Acceptance criteria:**
- [ ] Test "boat coasts to stop": boat starts with velocity 2 m/s forward, zero wind, no lines; after 10 sim-seconds speed < 0.1 m/s (drag brings it to rest)
- [ ] Test "mooring holds boat": boat moored at stern, beam-on wind 10 m/s; after 30 sim-seconds boat has not moved more than 2 m laterally (line absorbs wind)
- [ ] Test "throttle forward": full-forward monohull, no drag override; after 5 sim-seconds speed > 1 m/s
- [ ] Test "prop walk in astern": full-astern monohull with positive PropWalk; after 3 sim-seconds lateral displacement > 0 (boat walks to port)
- [ ] Test "collision with dock": boat placed 0.5 m inside dock polygon; after 0.1 sim-seconds boat has moved away from dock (penalty spring pushes it out)
**Definition of done:** all criteria checked; `go test -run TestWorldStep ./internal/sim/...` passes

---

### TASK-076: Input handler unit tests
**Depends on:** TASK-031, TASK-032, TASK-033, TASK-034
**Acceptance criteria:**
- [ ] Table-driven tests for all 7 key mappings in TASK-031
- [ ] Two-step mooring placement state machine: dock-click → cleat-click → command; dock-click → miss-click → cancel
- [ ] Wind panel scroll clamping verified: scroll down from 0 stays at 0; scroll up from 30 stays at 30
- [ ] Rudder angle clamp verified: repeated A presses cap at `+MaxRudderAngle`
- [ ] All handler tests pass under `-race`
**Definition of done:** all criteria checked; coverage of `internal/input` ≥ 80%

---

### TASK-077: Viewport transform unit tests
**Depends on:** TASK-040
**Acceptance criteria:**
- [ ] Round-trip test: `ScreenToWorld(WorldToScreen(p)) == p` for 10 random points
- [ ] Y-flip verified: `WorldToScreen(Vec2{0, 10})` has lower screen Y than `WorldToScreen(Vec2{0, 0})`
- [ ] Origin mapping verified per TASK-040 acceptance criteria
- [ ] Scale change test: doubling `Scale` halves the world extent visible on screen
**Definition of done:** all criteria checked; `go test ./internal/render/...` passes

---

## Group 8 — Security and Quality

### TASK-080: Input validation audit
**Depends on:** TASK-034, TASK-076
**Acceptance criteria:**
- [ ] Grep confirms all `CommandType` payloads are validated before `ApplyCommand` is called: `grep -r "ApplyCommand" --include="*.go"` shows only calls in `main.go` where commands come from validated handler
- [ ] No raw `Payload.(type)` assertion in `sim` package without a preceding validation step
- [ ] `WindPayload.Direction` cannot cause NaN in physics (e.g. `math.Cos(direction)` is always finite)
- [ ] No integer overflow possible in mooring line index (line count bounded by session, not user-controlled integer injection)
**Definition of done:** all criteria checked; documented in a brief comment block in `internal/sim/world.go`

---

### TASK-081: Security review — unsafe/CGO/network check
**Depends on:** TASK-060
**Acceptance criteria:**
- [ ] `grep -r "unsafe" --include="*.go" .` returns zero results (no unsafe pointer use)
- [ ] `grep -r "\"C\"" --include="*.go" .` returns zero results (no CGO)
- [ ] `grep -r "net/http" --include="*.go" .` returns only `tools/serve.go` (serve tool only, not WASM binary)
- [ ] `go build -tags netgo GOOS=js GOARCH=wasm .` succeeds (confirms no net import in WASM binary)
- [ ] `govulncheck ./...` runs without high-severity findings (install: `go install golang.org/x/vuln/cmd/govulncheck@latest`)
**Definition of done:** all criteria checked; findings (if any) documented in review-gaps.md

---

### TASK-082: Static analysis pass
**Depends on:** TASK-060
**Acceptance criteria:**
- [ ] `go vet ./...` returns zero warnings
- [ ] `staticcheck ./...` returns zero warnings (install: `go install honnef.co/go/tools/cmd/staticcheck@latest`)
- [ ] `go test -race ./internal/...` passes (all non-render packages, since render requires Ebiten display)
- [ ] Build tag `//go:build js` is present in `main.go` only for WASM-specific entry point if needed
**Definition of done:** all criteria checked

---

## Dependency Summary (linear build order)

```
TASK-001
  ↓
TASK-010 → TASK-011 → TASK-012
  ↓             ↓
TASK-013    TASK-020 → TASK-021
TASK-014         ↓
TASK-015    TASK-022 → TASK-024 → TASK-025 → TASK-026
TASK-016         ↓                    ↓          ↓
                TASK-023           TASK-027   TASK-030
                                              → TASK-031
                                              → TASK-032
                                              → TASK-033 → TASK-034
TASK-040 → TASK-041 → TASK-042 → TASK-044
              ↓         ↓
           TASK-043   TASK-053
TASK-050 → TASK-051 → TASK-052
All above → TASK-060 → TASK-061 → TASK-062
All above → TASK-070..077
TASK-034, TASK-076 → TASK-080
TASK-060 → TASK-081
TASK-060 → TASK-082
```
