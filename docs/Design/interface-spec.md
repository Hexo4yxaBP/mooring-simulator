---
updated: 2026-05-14
supersedes: original interface-spec.md (was Go package contracts; actual codebase is TypeScript)
---

# Interface Specification — Physics Subsystem

*Every decision references a fact in `docs/Research/`.*

---

## New Functions Added to `src/main.ts`

### `physicsStep(dt: number): void`

Single entry point for the physics simulation. Called from `render()` each frame before draw calls.

**Preconditions**: `gameMode === 'play'` AND `activeBoatIdx !== null` (caller guards).  
**Input**: `dt` — elapsed seconds since last frame, clamped to `[0, 0.1]` by caller. Source: constraint T2 (`docs/Research/constraints.md`).  
**Side effects**: mutates `boats[activeBoatIdx].{x, y, heading, vx, vy, omega}` only. No other boat or line is mutated.  
**Reads**: `windAngle`, `windKt`, `activeBoatIdx`, `boats`, `mooringLines`, `staticCleats`.

**Force accumulation order** (each adds to running `fx, fy, torque`):
1. Hydrodynamic drag
2. Engine thrust
3. Propeller walk
4. Rudder force
5. Wind force
6. Mooring line spring forces
7. Collision penalty forces

**Integration**: semi-implicit Euler — velocity updated before position:
```
vx += (fx/mass) * dt
vy += (fy/mass) * dt
omega += (torque/I) * dt
x  += vx * dt
y  += vy * dt
heading += omega * dt
```
Source: architecture.md D4 (existing doc).

---

### `getCleatWorld(ref: CleatRef): [number, number]`

Returns `[wx, wy]` in world-space meters for any cleat reference.

For `kind === 'static'`: returns `[staticCleats[ref.idx].wx, staticCleats[ref.idx].wy]` directly.

For `kind === 'moving'`: applies the identical SVG→local-canvas transform as `getMovingCleatCanvas`
(lines ~332–344 in `src/main.ts`), then converts canvas offset to world:
```typescript
// lx, ly are local canvas-space offsets (Y-down, pixels)
// rlx, rly are rotated by boat.heading
wx = boat.x + rlx / SCALE
wy = boat.y - rly / SCALE   // flip Y: canvas-down → world-up
```
Source: gap G2, coordinate system contract — `docs/Research/constraints.md` and `interfaces.md`.

---

### `computeNaturalLength(from: CleatRef, to: CleatRef): number`

Returns the Euclidean distance (meters) between two cleats at call time.  
Called once when a mooring line is created, to populate `MooringLine.naturalLength`.

Source: domain-model.md §MooringLine — natural length = rest length at placement time.

---

## Modified Interfaces

### `Boat` — three new optional velocity fields

```typescript
interface Boat {
  // existing (unchanged)
  type: BoatType;
  x: number;
  y: number;
  heading: number;
  rudderAngle: number;
  throttlePort: number;
  throttleStbd: number;
  // new
  vx?: number;    // world-space m/s; default 0 if absent
  vy?: number;    // world-space m/s; default 0 if absent
  omega?: number; // angular velocity rad/s; default 0 if absent
}
```

Optional (`?`) so all existing `boats.push(...)` calls remain valid.
`physicsStep` initialises any absent field to 0 on first call.

Source: gap G1 — `docs/Research/constraints.md`. Interface lock constraint — `docs/Research/interfaces.md`.

### `MooringLine` — one new required field

```typescript
interface MooringLine {
  from: CleatRef;
  to: CleatRef;
  naturalLength: number; // world metres — computed at line creation time
}
```

The single `mooringLines.push(...)` call (in `mousedown`) must be updated to pass `naturalLength`.
`drawMooringLines` and `deleteBoat` iterate over lines but never read `naturalLength` — no change needed.

Source: domain-model.md §MooringLine, constraint P1 (tension-only rope) — `docs/Research/constraints.md`.

---

## Physics Constants

All declared as module-level `const` in `src/main.ts` alongside existing constants (SCALE, CLEAT_R, etc.).

### Boat Physical Properties

```typescript
const BOAT_MASS: Record<BoatType, number> = {
  monohull:  7000,    // kg — typical 10 m monohull (domain-model.md)
  catamaran: 12000,   // kg — typical 11 m catamaran
};

const BOAT_I: Record<BoatType, number> = {
  monohull:  63600,   // kg·m² — (1/12)×7000×(3²+10²)
  catamaran: 163000,  // kg·m² — (1/12)×12000×(6.5²+11²)
};
```

Moment of inertia uses rectangular approximation — `docs/Research/constraints.md` §What Can Be Decided.

### Engine Thrust

```typescript
// Index = throttle integer (0=full astern … 4=full ahead). Values in Newtons.
const THROTTLE_FORCE = [-10000, -4000, 0, 5000, 15000];
```

Catamaran: each engine gets `THROTTLE_FORCE[throttle] / 2`.
Monohull: all thrust from single engine `THROTTLE_FORCE[throttlePort]`.
Applied along bow vector at stern position.

### Propeller Walk

```typescript
// Lateral force at stern in starboard direction (+stbd = positive). Newtons.
// Right-handed screw convention: astern pushes stern to port (negative).
const PROP_WALK_TABLE = [-1500, -500, 0, 250, 500];
```

- Monohull: net lateral = `PROP_WALK_TABLE[throttlePort]` (single right-handed prop).
- Catamaran: port engine (right-handed) = `+PROP_WALK_TABLE[throttlePort]`;
  starboard engine (left-handed, contra-rotating) = `−PROP_WALK_TABLE[throttleStbd]`.
  Net walk cancels when both throttles equal; amplifies on differential throttle.

Source: constraint P4, P5 — `docs/Research/constraints.md`.

### Hydrodynamic Drag (linear model)

```typescript
const C_DRAG_FWD = 3000;   // N/(m/s) — longitudinal drag
const C_DRAG_LAT = 80000;  // N/(m/s) — lateral drag (keel effect, ~27× fwd)
const C_DRAG_ROT = 50000;  // N·m/(rad/s) — rotational drag
```

Source: constraint P3 (lateral >> longitudinal) — `docs/Research/constraints.md`.

### Rudder

```typescript
const C_RUDDER = 30000; // N per (rad × m/s forward speed)
```

`F_rudder = C_RUDDER × rudderAngle × vForward`. Applied at stern, perpendicular to heading.
Zero contribution when `|vForward| < 0.05 m/s`. Source: constraint P2, P5.

### Wind

```typescript
const WIND_K_FWD = 5;   // N/(m/s)² — bow-on wind drag coefficient
const WIND_K_LAT = 27;  // N/(m/s)² — beam-on wind drag coefficient
```

Wind speed converted at use time: `ws_ms = windKt × 0.51444`.
Force is quadratic: `F = K × |vw_component| × vw_component`.
At 15 kt (7.7 m/s) beam-on: `27 × 7.7² ≈ 1600 N`.
Source: domain-model.md §Wind Force.

### Mooring Line Spring

```typescript
const MOORING_K = 50000; // N/m spring constant
const MOORING_C = 10000; // N·s/m damping coefficient
```

Tension formula: `F = max(0, MOORING_K × extension − MOORING_C × extensionRate)`.
Clamp ensures tension-only (constraint P1). Source: domain-model.md §MooringLine.

### Collision

```typescript
const COLLISION_K = 150000; // N/m penalty spring constant
```

Applied as `F = COLLISION_K × mtv` where `mtv` is the world-space MTV from `obbMTV()`.
50% position correction applied simultaneously to prevent tunnelling at low dt.
Source: constraint T5, decision in `docs/Research/constraints.md` §What Can Be Decided.

---

## Value Constraints

| Value | Allowed range | Enforcement |
|-------|--------------|-------------|
| `dt` | [0, 0.1] s | clamped in `render()` before call |
| `throttlePort`, `throttleStbd` | 0..4 integer | existing slider; THROTTLE_FORCE indexed directly |
| `rudderAngle` | [−0.6109, +0.6109] rad | existing slider |
| `windKt` | [0, 99] | existing input |
| `vx`, `vy` | unclamped | drag naturally limits terminal velocity |
| `omega` | unclamped | rotational drag naturally limits |
