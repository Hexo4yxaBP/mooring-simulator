# Constraints — Mooring Simulator

*Sources: `initial-problem.md`, Go ecosystem characteristics, 2D physics conventions.*

---

## Hard Constraints (from spec)

| # | Constraint | Source |
|---|-----------|--------|
| C1 | Must run in a browser | spec |
| C2 | Golang-based implementation | spec |
| C3 | Top-down 2D visualization | spec |
| C4 | Mooring line boat-end: only at stern, midships, or bow cleats | spec |
| C5 | Mooring line dock-end: any point on dock | spec |
| C6 | Throttle: exactly 5 discrete states (Neutral, Slow Fwd, Full Fwd, Slow Astern, Full Astern) | spec |
| C7 | Forces to include: wind, rudder, CoM, dock/boat collision, prop walk, engine thrust | spec |
| C8 | Boat hull must show center of mass marker | spec |
| C9 | Support multiple simultaneous boats | spec |
| C10 | Support both monohull sailboat and catamaran | spec |

---

## Technical Constraints

| # | Constraint | Reason |
|---|-----------|--------|
| T1 | No CGO — physics engine must be pure Go | CGO incompatible with `GOOS=js GOARCH=wasm` (WASM target) |
| T2 | WASM binary size: minimise unused imports | Large binaries slow first-load |
| T3 | Collision detection must cover convex hull shapes | Boats and dock are non-trivial polygons |
| T4 | Fixed timestep simulation loop required | Ensures deterministic physics regardless of frame rate |
| T5 | `syscall/js` or Ebiten are the only ways Go renders to browser canvas | Standard library has no browser drawing API |

---

## Physics Constraints

| # | Constraint | Reason |
|---|-----------|--------|
| P1 | Mooring lines: tension-only (no compression) | Real rope/line can only pull, not push |
| P2 | Boat cannot pass through dock — collision boundary enforced | Spec requires "contact with dock" force |
| P3 | Hydrodynamic lateral drag >> longitudinal drag | Sailboat hull is streamlined bow-to-stern, not broadside |
| P4 | Prop walk strongest in astern, weaker in forward | Hydrodynamic reality — propeller slipstream vs hull |
| P5 | Rudder force requires non-zero boat speed | Rudder is a lift device — no water flow = no force |
| P6 | Forces must create both linear and rotational effects (torque = r × F) | Off-center force application must rotate the boat |

---

## Unknowns Requiring User Clarification

| ID | Question | Impact on Design |
|----|---------|-----------------|
| U1 | ~~WASM client-only vs Go server + JS frontend?~~ | **RESOLVED: Ebiten → WASM, no server** |
| U2 | ~~Ebiten vs raw WASM?~~ | **RESOLVED: Ebiten** |
| U3 | Physics accuracy level (naval-grade vs tunable)? | Determines whether to use real hydrodynamic formulas or simplified constants |
| U4 | ~~Catamaran twin-engine prop walk model?~~ | **RESOLVED: Twin engines (future). MVP: catamaran = non-active boat, no engine physics** |
| U5 | ~~Dock shape?~~ | **RESOLVED: Straight pier for MVP; architecture must support configurable shape later** |
| U6 | Collision response (penalty spring vs impulse)? | Determines physics solver approach |
| U7 | Measurement units (meters/SI vs pixels vs nautical)? | Affects all physics constants |
| U8 | Real-time continuous vs pausable/step simulation? | Affects game loop and UI design |
| U9 | Mooring lines: elastic spring only, or also inextensible (taut) option? | Affects constraint solver |
| U10 | Boat parameters (mass, dimensions) fixed defaults or user-configurable? | Affects UI requirements |
| U11 | Is wind global only, or can there be local wind variations? | Affects wind force model |
| U12 | How many mooring lines maximum per session? | Affects performance budgeting |

---

## Confirmed Decisions (2026-05-07)

| Decision | Choice | Notes |
|----------|--------|-------|
| Runtime | **Ebiten → WASM** | Client-only, no server |
| Catamaran MVP scope | **Non-active boat only** | No engine/prop walk for catamaran in MVP; drifts under wind/lines only |
| Catamaran full scope | Twin independent engines | Defer post-MVP |
| Dock MVP | **Straight pier** | Fixed geometry in MVP |
| Dock future | User-configurable shape | Design for extensibility now |

---

## What Can Be Decided Without Clarification

- Physics integration method: **semi-implicit Euler** (simple, stable for spring systems) or RK4 (more accurate). Either works; semi-implicit Euler is standard for games.
- Coordinate system: screen-space (y down) for rendering, with world-space (y up) for physics, transformed at draw time. Standard pattern.
- Force accumulation pattern: per-tick accumulator cleared each frame.
- Cleat positions are fixed offsets in body frame, defined per boat type.
- Mooring line spring model: Hookean spring + linear damper (standard).
- Wind model: constant global vector field (simplest, matches spec).
