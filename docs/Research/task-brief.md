---
updated: 2026-05-14
supersedes: original task-brief.md (was Go/WASM; actual codebase is TypeScript)
---

# Task Brief — Physics Simulation

## 1. Task Type

**Adding functionality to existing codebase.**

Sub-type: game physics integration — adding a continuous force-based simulation loop to a
browser canvas simulator that currently has no physics (boats are positioned by drag-and-drop only).

---

## 2. Input Data

| Source | Location | What it provides |
|--------|----------|-----------------|
| Application source | `src/main.ts` | Entire app — ~983 lines, single file |
| Boat SVG assets | `public/boats/monohull.svg`, `catamaran.svg` | Hull outline shapes |
| User task statement | This conversation | Physics forces required |
| HTML entry point | `index.html` | UI panels: rudder slider, throttle sliders, wind input |

No external API specs, test suite, or build system beyond Vite.

---

## 3. Facts About the Current System

### Coordinate system
- **World space**: Y-up, meters, origin at canvas centre. `SCALE = 20 px/m`.
- **Canvas space**: Y-down pixels, origin top-left.
- Conversion: `worldToCanvas(wx, wy) → [cw/2 + wx·S, ch/2 − wy·S]`
- `boat.x`, `boat.y` are world-space positions in meters.
- `boat.heading`: radians, `0 = bow pointing north (up)`.

### Boat state (no physics yet)
```typescript
interface Boat {
  type: 'monohull' | 'catamaran';
  x: number;          // world m
  y: number;          // world m
  heading: number;    // rad, 0=north
  rudderAngle: number; // rad, ±35°
  throttlePort: number; // 0..4  (0=full astern, 2=neutral, 4=full ahead)
  throttleStbd: number; // catamaran only; same range
}
```
**Missing for physics**: `vx`, `vy` (m/s world-space velocity), `omega` (rad/s angular velocity).

### Simulation loop
`render()` calls `requestAnimationFrame(render)` — **variable timestep, no `dt` tracking**.
No physics update step exists. Boat position/heading is only mutated by drag handlers.

### Wind
- `windAngle: number` — radians, 0 = blowing toward north.
- `windKt: number` — speed in knots (0–99), global.
- Both are already available to any new physics code.

### Throttle mapping
| Slider value | Meaning |
|---|---|
| 0 | Full astern |
| 1 | Slow astern |
| 2 | Neutral |
| 3 | Slow ahead |
| 4 | Full ahead |

Thrust fractions to derive: [-1, -0.5, 0, +0.5, +1] (implementation decision, not yet in code).

### Mooring lines
- Stored as `MooringLine[]` — pairs of `CleatRef` (static index or boat+cleat index).
- Currently **zero-force**: lines are drawn but apply no spring tension.
- `getCleatCanvas(ref)` returns canvas-space `[cx, cy]`. World-space equivalent needed for physics.
- All cleats are points (no length/elasticity constants defined).

### Hull geometry
Both hulls are **cubic bezier curves** stored as `Path2D` objects in `OUTLINE_PATHS`.
Canvas uses `ctx.stroke(path)` within a `translate + rotate + scale(1,−1)` transform.
For physics, hulls must be sampled to polygon vertices.

Monohull: 1 closed bezier path, symmetric, ~12.5 m long × 3 m wide (SVG units map to BOAT_SIZE).
Catamaran: 2 hull paths + 1 deck rect, ~11 m long × 6.5 m wide.

Cleat positions are defined separately in `CLEAT_SVG` (SVG path space, 6 cleats per boat).

### Collision (current)
OBB SAT (`obbMTV()`) used during drag-and-drop only. Returns MTV vector in world space.
**Not called during physics update.** Pier is an infinite-depth OBB below the waterline.

### Play mode
`gameMode: 'setup' | 'play'` — physics must only run when `gameMode === 'play'`.
`activeBoatIdx: number | null` — the one user-controlled boat.

---

## 4. Phase 1 Questions

| # | Question | Why it matters |
|---|----------|----------------|
| Q1 | Physics accuracy: naval-grade coefficients vs tunable gameplay constants? | Determines whether to model hydrodynamic lift/drag curves or use hand-tuned values |
| Q2 | Does wind affect inactive (non-active) boats too, or active boat only? | Scope of per-frame update loop |
| Q3 | Should mooring line tension prevent hull penetration into the pier, or is pier collision separate? | Force superposition vs constraint priority |
| Q4 | Mooring line model: simple Hookean spring, or spring + damper, or inextensible? | Determines tension formula and stability requirement |
| Q5 | Mass model: fixed constants per boat type, or user-configurable? | Affects acceleration calculation |
| Q6 | Catamaran prop walk: independent port/starboard engines each produce their own walk, or averaged? | Torque model for catamaran |
| Q7 | Collision response: penalty spring (continuous force) or impulse (instantaneous)? | Determines solver architecture |
| Q8 | Hull collision: use existing OBB SAT or upgrade to bezier polygon approximation? | Accuracy vs implementation complexity |
