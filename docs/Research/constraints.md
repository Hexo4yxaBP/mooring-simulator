---
updated: 2026-05-14
supersedes: original constraints.md (was Go/WASM; actual codebase is TypeScript)
---

# Constraints — Physics Implementation

---

## Hard Constraints (from existing code + user requirements)

| # | Constraint | Source |
|---|-----------|--------|
| C1 | Must run in the browser with no server | Existing architecture |
| C2 | TypeScript + HTML5 Canvas 2D — no WebGL, no external physics library | Existing codebase |
| C3 | Physics only in Play mode (`gameMode === 'play'`) | User feature table |
| C4 | Only `boats[activeBoatIdx]` receives engine/wind/collision physics | User statement |
| C5 | Throttle: exactly 5 discrete states (0..4 integer) | Existing slider |
| C6 | Rudder range: ±35° (±0.6109 rad) | Existing slider |
| C7 | Mooring lines connect `CleatRef` pairs — endpoints cannot change | Existing data model |
| C8 | Pier OBB geometry is derived from `canvas.height` and `PIER_DEPTH_M` | `getPierOBB()` |
| C9 | `SCALE = 20` px/m — world units are meters | Hardcoded constant |
| C10 | `render()` drives the main loop via `requestAnimationFrame` | Existing architecture |

---

## Physics Constraints (from domain knowledge)

| # | Constraint | Rationale |
|---|-----------|-----------|
| P1 | Mooring lines: tension-only (no compression) | Ropes pull, never push |
| P2 | Rudder force requires non-zero boat speed (proportional to v²) | Lift device — no flow = no force |
| P3 | Hydrodynamic drag: lateral >> longitudinal | Hull is streamlined fore-aft |
| P4 | Prop walk: strongest in astern, weaker in ahead | Propeller slipstream asymmetry |
| P5 | Catamaran prop walk: port engine walks to port, starboard to starboard | Independent screws |
| P6 | Forces produce both linear and rotational effects: τ = r × F | Off-centre application creates torque |
| P7 | Pier boundary: boat cannot penetrate below waterline | Hard wall |
| P8 | Wind force proportional to v² (drag equation) — but simplified to linear(windKt) is acceptable | Gameplay fidelity |
| P9 | Fixed-seed, deterministic physics loop preferred (fixed Δt or scaled variable Δt) | Stability |

---

## Technical Constraints

| # | Constraint | Reason |
|---|-----------|--------|
| T1 | Variable timestep from `requestAnimationFrame` — must compute `dt = now − lastTime` | No game loop with fixed step |
| T2 | dt must be clamped (e.g. max 100 ms) to avoid spiral-of-death after tab backgrounding | rAF pauses when tab hidden |
| T3 | Bezier curves in `OUTLINE_PATHS` are NOT directly usable for collision math | Path2D has no API to sample points |
| T4 | Hull polygon must be pre-computed (sampled from bezier) — done once at startup | Sampling at 60 fps would be wasteful |
| T5 | OBB SAT (`obbMTV`) is already correct for pier/boat rectangular collision detection | Reuse for collision impulse force direction |
| T6 | `getCleatCanvas()` returns canvas pixels; need world-space equivalent for spring force vector | Existing gap — new helper needed |
| T7 | No external physics library — impulse / spring math must be implemented inline | Browser-only constraint |
| T8 | TypeScript strict mode (`noUnusedLocals`, `noUnusedParameters`) — all new state fields must be used | Existing tsconfig |

---

## What Can Be Decided Without User Clarification

| Decision | Default choice | Justification |
|----------|---------------|---------------|
| Physics integration method | Semi-implicit Euler (leapfrog variant) | Stable for spring systems, simple to implement |
| Mooring spring model | Hookean spring + linear damper: `F = k·x − c·v` | Standard, stable for gameplay |
| Wind force model | `F_wind = k_wind · windKt · windArea · (windDir − heading projection)` | Simplified drag; tunable constant |
| Mass model | Fixed constants per boat type (monohull ~7000 kg, catamaran ~12000 kg) | No user config planned |
| Moment of inertia | `I = (1/12) · m · (w² + h²)` (rectangular approximation) | Sufficient for gameplay |
| Hull collision for physics | Reuse OBB SAT — the bounding box is a good enough proxy for collision force direction | Bezier polygon would add complexity without meaningful gameplay gain |
| Prop walk model | Lateral impulse: `F_pw = k_pw · throttle_factor · sign(throttle)` on stern cleat position | Standard simplified model |
| Catamaran prop walk | Each engine independent: port thrust at port stern cleat, starboard at starboard stern cleat | User confirmed twin independent engines |
| Inactive boats | Static (no physics). Wind drift for inactive boats is out of scope for now. | User said active boat only |

---

## Known Gaps (must be addressed in design)

| Gap | Description |
|-----|-------------|
| G1 | `Boat` interface has no velocity/angular velocity fields | Must add `vx, vy, omega` |
| G2 | No `getCleatWorld()` function | Required for spring force vector computation |
| G3 | No `dt` tracking in render loop | Must add `lastTime` variable and compute `dt` |
| G4 | No thrust-to-force mapping table | Need constants: `THRUST_FORCE[0..4]` in Newtons |
| G5 | No damping / drag constants defined | Need `DRAG_LINEAR`, `DRAG_ANGULAR`, `DRAG_LATERAL` |
| G6 | No spring/damper constants for mooring lines | Need `MOORING_K` (N/m), `MOORING_C` (N·s/m) |
| G7 | No wind force coefficient | Need `WIND_FORCE_K` |
| G8 | No propeller walk coefficient | Need `PROP_WALK_K` |
