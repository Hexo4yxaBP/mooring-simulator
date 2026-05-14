---
updated: 2026-05-14
supersedes: original implementation-tasks.md (was Go/WASM multi-package; actual codebase is TypeScript single-file)
---

# Implementation Tasks — Physics Simulation

*Decomposed from docs/Design/architecture.md, interface-spec.md, data-schemas.md.*
*All changes are in `src/main.ts` unless stated otherwise.*
*All tasks estimated 1–4 hours.*

---

## Group 0 — Prerequisites: State and Helpers

### TASK-001: Extend interfaces and add physics constants
**Depends on:** none
**Acceptance criteria:**
- [ ] `Boat` interface gains three optional fields: `vx?: number`, `vy?: number`, `omega?: number`
  (interface-spec.md §Boat)
- [ ] `MooringLine` interface gains one required field: `naturalLength: number`
  (interface-spec.md §MooringLine)
- [ ] `BOAT_MASS`, `BOAT_I` added as `Record<BoatType, number>` constants with values from interface-spec.md §Boat Physical Properties
- [ ] `THROTTLE_FORCE` added as `number[]` (5 values, indices 0..4) with values `[-10000,-4000,0,5000,15000]`
- [ ] `PROP_WALK_TABLE` added as `number[]` (5 values) with values `[-1500,-500,0,250,500]`
- [ ] `C_DRAG_FWD = 3000`, `C_DRAG_LAT = 80000`, `C_DRAG_ROT = 50000` added
- [ ] `C_RUDDER = 30000` added
- [ ] `WIND_K_FWD = 5`, `WIND_K_LAT = 27` added
- [ ] `MOORING_K = 50000`, `MOORING_C = 10000` added
- [ ] `COLLISION_K = 150000` added
- [ ] `npx tsc --noEmit` (or Vite dev server start) reports zero type errors
**Definition of done:** all criteria checked; `npx tsc --noEmit` passes

---

### TASK-002: Implement `getCleatWorld()` and update mooring line creation
**Depends on:** TASK-001
**Acceptance criteria:**
- [ ] `getCleatWorld(ref: CleatRef): [number, number]` function added near `getCleatCanvas`
- [ ] For `kind === 'static'`: returns `[staticCleats[ref.idx].wx, staticCleats[ref.idx].wy]` directly
- [ ] For `kind === 'moving'`: computes local canvas offset `(lx, ly)` identically to `getMovingCleatCanvas`, then returns `[boat.x + rlx/SCALE, boat.y - rly/SCALE]`
- [ ] Consistency check: for a boat at world `(0,0)` heading=0, `getCleatWorld({kind:'moving', boat:0, cleat:0})` returns approximately `(0, -5)` (port stern, ~5 m aft) — negative Y = south = stern for bow-north boat
- [ ] `computeNaturalLength(from: CleatRef, to: CleatRef): number` added: calls `getCleatWorld` for both endpoints, returns Euclidean distance
- [ ] The single `mooringLines.push({from: pendingCleat, to: hitCleat})` call updated to include `naturalLength: computeNaturalLength(pendingCleat, hitCleat)`
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; type check passes; function is reachable from physics step

---

### TASK-003: Add `dt` tracking and play-mode velocity reset
**Depends on:** TASK-001
**Acceptance criteria:**
- [ ] Module-level `let lastTime = 0` added (architecture.md §Render Loop dt Tracking)
- [ ] `render` function updated to accept `DOMHighResTimeStamp` parameter (or use `performance.now()` inline)
- [ ] `dt` computed as `Math.min((now - lastTime) / 1000, 0.1)` (seconds, clamped — constraint T2)
- [ ] `lastTime = now` assigned each frame after dt computation
- [ ] In `playBtn` click handler, when entering play mode (`gameMode` transitions to `'play'`), active boat's `vx`, `vy`, `omega` are set to 0 (architecture.md §Play Mode Toggle Reset)
- [ ] `requestAnimationFrame(render)` call passes the timestamp correctly (rAF already provides this)
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; opening browser console shows no errors; dt is non-zero each frame

---

## Group 1 — Physics Step Core

### TASK-010: `physicsStep()` skeleton + semi-implicit Euler integrator
**Depends on:** TASK-001, TASK-003
**Acceptance criteria:**
- [ ] `function physicsStep(dt: number): void` added to `src/main.ts`
- [ ] Function returns immediately if `activeBoatIdx === null`
- [ ] On entry, initialises `boat.vx`, `boat.vy`, `boat.omega` to 0 if `undefined`
- [ ] Local variables `let fx = 0, fy = 0, torque = 0` accumulate forces
- [ ] Bow and starboard unit vectors derived: `bowX = Math.sin(boat.heading)`, `bowY = Math.cos(boat.heading)`, `stbdX = Math.cos(boat.heading)`, `stbdY = -Math.sin(boat.heading)` (matches OBB SAT axis convention in existing `obbMTV()`)
- [ ] After all force sections (initially empty), semi-implicit Euler integration runs:
  `vx += (fx/mass)*dt`, `vy += (fy/mass)*dt`, `omega += (torque/I)*dt`, `x += vx*dt`, `y += vy*dt`, `heading += omega*dt`
- [ ] Mass and moment of inertia looked up from `BOAT_MASS[boat.type]` and `BOAT_I[boat.type]`
- [ ] `physicsStep(dt)` called from `render()` **before** draw calls, guarded: `if (gameMode === 'play' && activeBoatIdx !== null) physicsStep(dt);`
- [ ] With empty force sections: boat placed in water drifts (no forces, constant velocity) — position changes each frame by `vx*dt`, `vy*dt`
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; boat moves smoothly when given initial velocity in browser console (`boats[0].vx = 1`)

---

### TASK-011: Hydrodynamic drag
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] Forward body-frame speed computed: `vFwd = boat.vx * bowX + boat.vy * bowY`
- [ ] Lateral body-frame speed: `vLat = boat.vx * stbdX + boat.vy * stbdY`
- [ ] Drag forces added: `fx -= C_DRAG_FWD * vFwd * bowX + C_DRAG_LAT * vLat * stbdX`, `fy -= C_DRAG_FWD * vFwd * bowY + C_DRAG_LAT * vLat * stbdY`
- [ ] Rotational drag: `torque -= C_DRAG_ROT * boat.omega`
- [ ] **Behavioral test**: boat at `vx=0, vy=2` (moving north) with heading=0 (bow north) decelerates: after calling `physicsStep(0.1)` ten times, `boat.vy` is less than 2 (drag acts)
- [ ] Lateral drag >> forward drag: beam-on motion (`vLat=1`) produces ~26.7× more drag force magnitude than forward motion (`vFwd=1`) at same speed (`C_DRAG_LAT/C_DRAG_FWD = 80000/3000`)
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; in browser, boat given `vy=5` coasts to a stop within visible play area

---

### TASK-012: Engine thrust and propeller walk
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] **Monohull thrust**: `const thrustN = THROTTLE_FORCE[boat.throttlePort]`; `fx += thrustN * bowX; fy += thrustN * bowY`
- [ ] Thrust applied at stern — torque contribution: `(sternOffsetY * 0 - sternOffsetX * thrustN * bowY)` — thrust is along bow, stern is on bow axis, so cross product = 0 (no torque from straight thrust, verified algebraically)
- [ ] **Monohull prop walk**: `const pwN = PROP_WALK_TABLE[boat.throttlePort]`; lateral force at stern: `fx += pwN * stbdX; fy += pwN * stbdY`; stern offset from boat centre ≈ `−h/2` along bow (world: `sternWx = -(h/2)*bowX, sternWy = -(h/2)*bowY`); torque: `torque += sternWx * (pwN*stbdY) - sternWy * (pwN*stbdX)`
- [ ] **Catamaran thrust**: port engine = `THROTTLE_FORCE[boat.throttlePort]/2` at port stern; stbd engine = `THROTTLE_FORCE[boat.throttleStbd]/2` at stbd stern; thrust from each added to `fx/fy`; differential throttle generates yaw torque (port force applied at `sternWx - lateralArm*stbdX`, stbd force at `sternWx + lateralArm*stbdX` where `lateralArm = 2.04`)
- [ ] **Catamaran prop walk**: port engine: `+PROP_WALK_TABLE[boat.throttlePort]`; stbd engine: `-PROP_WALK_TABLE[boat.throttleStbd]` (contra-rotation); both applied at respective stern positions
- [ ] **Behavioral test (monohull)**: throttle=4 (full ahead), heading=0, no drag — after `physicsStep(1.0)`, `boat.vy` increases (boat moves north)
- [ ] **Behavioral test (monohull astern)**: throttle=0 (full astern), `boat.omega` changes sign (prop walk creates yaw)
- [ ] **Behavioral test (catamaran)**: equal throttle on both engines → zero net yaw torque from differential thrust
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; in browser, monohull accelerates to ~5 m/s at full throttle with drag balanced

---

### TASK-013: Rudder force
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] `vFwd` (forward speed) computed (may reuse from drag section or recompute)
- [ ] Rudder force: `const rudderF = C_RUDDER * boat.rudderAngle * vFwd`; applied as lateral force `fx += rudderF * stbdX; fy += rudderF * stbdY`
- [ ] Applied at stern position (same `sternWx, sternWy` as thrust); torque: `torque += sternWx * (rudderF*stbdY) - sternWy * (rudderF*stbdX)`
- [ ] Zero-speed guard: if `Math.abs(vFwd) < 0.05`, rudder force = 0 (constraint P5 — `docs/Research/constraints.md`)
- [ ] **Behavioral test**: heading=0, rudderAngle=+0.35 rad (20° stbd), vFwd=2 m/s → torque is non-zero and causes `boat.omega` to change sign (boat turns)
- [ ] **Behavioral test**: same setup but `vFwd=0` → `torque` unchanged (rudder adds nothing)
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; in browser, boat at speed turns when rudder applied

---

### TASK-014: Wind force
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] Wind speed converted: `const wsMs = windKt * 0.51444` (knots to m/s)
- [ ] Wind blow-toward vector: `windVx = -wsMs * Math.sin(windAngle)`, `windVy = -wsMs * Math.cos(windAngle)` (windAngle convention: 0 = blowing toward north — `docs/Research/task-brief.md` §Wind)
- [ ] Apparent wind components: `appFwd = windVx * bowX + windVy * bowY`, `appLat = windVx * stbdX + windVy * stbdY`
- [ ] Quadratic drag: `windFwdF = WIND_K_FWD * appFwd * Math.abs(appFwd)`, `windLatF = WIND_K_LAT * appLat * Math.abs(appLat)` (sign-preserving quadratic)
- [ ] Added to accumulator: `fx += windFwdF * bowX + windLatF * stbdX; fy += windFwdF * bowY + windLatF * stbdY`
- [ ] **Behavioral test**: heading=0 (bow north), wind blowing east (windAngle = π/2 → 270° compass): `appLat > 0` (wind on starboard side) → positive lateral wind force
- [ ] **Behavioral test**: 15 kt beam-on wind → lateral force magnitude ≈ `27 × (15×0.514)² ≈ 1600 N` (within 5%)
- [ ] **Behavioral test**: stationary boat in 20 kt beam-on wind drifts laterally (slowly — keel drag is high)
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; in browser, changing wind direction visibly affects boat heading/drift

---

### TASK-015: Mooring line spring forces
**Depends on:** TASK-002, TASK-010
**Acceptance criteria:**
- [ ] Loop iterates over all `mooringLines`
- [ ] Each line: determine if active boat is one endpoint (`line.from.kind==='moving' && line.from.boat===activeBoatIdx` or same for `line.to`)
- [ ] Skip lines not connected to active boat
- [ ] For connected lines: compute `[ax, ay] = getCleatWorld(movingEnd)` and `[bx, by] = getCleatWorld(otherEnd)`
- [ ] `dist = Math.sqrt((bx-ax)**2 + (by-ay)**2)`, `ext = dist - line.naturalLength`
- [ ] If `ext <= 0`: no force (slack — constraint P1)
- [ ] If `ext > 0`: direction `nx = (bx-ax)/dist, ny = (by-ay)/dist`; extension rate `dext = (boat.vx*nx + boat.vy*ny)`; `fMag = Math.max(0, MOORING_K * ext - MOORING_C * dext)`; add `fMag*nx` to `fx`, `fMag*ny` to `fy`
- [ ] Torque from line at cleat: `const rx = ax - boat.x, ry = ay - boat.y`; `torque += rx*(fMag*ny) - ry*(fMag*nx)`
- [ ] **Behavioral test**: boat 5 m beyond natural length of taut line → `fMag = MOORING_K * 5 = 250000 N` (large restoring force)
- [ ] **Behavioral test**: boat exactly at natural length → `ext=0`, no force
- [ ] **Behavioral test**: boat drifting away from anchor (`dext > 0`) → damping reduces force; boat approaching (`dext < 0`) → damping increases force (but `max(0,...)` prevents negative)
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; in browser, a moored boat in wind holds position near natural length

---

### TASK-016: Collision penalty forces
**Depends on:** TASK-010
**Acceptance criteria:**
- [ ] Build `activeOBB = boatToOBB(boats[activeBoatIdx])` each frame
- [ ] Build obstacles list: all other boats' OBBs + `getPierOBB()` — uses existing functions
- [ ] For each obstacle: `const mtv = obbMTV(activeOBB, obs)`; if non-null, apply penalty force: `fx += COLLISION_K * mtv[0]`, `fy += COLLISION_K * mtv[1]`
- [ ] 50% position correction applied simultaneously: `boat.x += mtv[0] * 0.5; boat.y += mtv[1] * 0.5` (prevents tunnelling)
- [ ] Rebuild `activeOBB` after position corrections (or loop corrections before forces — either approach documented in code)
- [ ] **Behavioral test**: boat positioned with bow touching pier (overlapping by 0.1 m) → after one physics step, boat's Y position has increased (pushed away from pier)
- [ ] **Behavioral test**: boat far from pier (no overlap) → `obbMTV` returns null, no collision force added
- [ ] **Behavioral test**: pier acts as hard wall — boat cannot pass through during normal gameplay (multiple frames at full throttle toward pier)
- [ ] `npx tsc --noEmit` passes
**Definition of done:** all criteria checked; in browser, boat does not penetrate pier or other boats at normal speeds

---

## Group 2 — Verification

### TASK-020: TypeScript type-check pass
**Depends on:** TASK-001 through TASK-016
**Acceptance criteria:**
- [ ] `npx tsc --noEmit` reports zero errors
- [ ] `tsconfig.json` settings respected: `strict: true`, `noUnusedLocals: true`, `noUnusedParameters: true`
- [ ] No `any` casts introduced (all new code uses proper types)
- [ ] No variables declared but unused (especially loop variables in mooring line iteration)
**Definition of done:** `npx tsc --noEmit` exits with code 0; zero errors, zero warnings

---

### TASK-021: Browser smoke tests (manual)
**Depends on:** TASK-020
**Acceptance criteria:**
- [ ] `make serve` (or `npx vite`) starts dev server without error
- [ ] Browser console shows no uncaught exceptions on page load
- [ ] **Smoke 1 — drag to stop**: Drop monohull, enter Play mode; observe boat is stationary
- [ ] **Smoke 2 — throttle**: Set throttle to Full Ahead; boat accelerates northward (bow up), reaches ~5 m/s terminal velocity
- [ ] **Smoke 3 — rudder**: At speed, set rudder to full starboard (35°); boat turns to starboard
- [ ] **Smoke 4 — prop walk**: Set Full Astern from stationary; boat develops angular velocity (turns without moving forward)
- [ ] **Smoke 5 — wind**: Set wind to 20 kt, stationary boat in beam-on wind drifts slowly sideways
- [ ] **Smoke 6 — mooring**: Draw mooring line stern-to-pier-cleat; in 20 kt wind, boat holds position near natural length
- [ ] **Smoke 7 — pier collision**: Motor at full ahead into pier; boat stops at pier face, does not penetrate
- [ ] **Smoke 8 — catamaran**: Drop catamaran, enter Play; set port throttle=4, stbd=2 (differential); boat turns to starboard
- [ ] **Smoke 9 — mode switch**: Switch back to Setup from Play; mooring lines remain; boat velocity carries over as zero (reset on mode enter)
- [ ] Rudder/throttle panels remain visible and functional throughout
**Definition of done:** all 9 smoke tests pass without visible artefacts (jitter, teleportation, NaN positions)

---

### TASK-022: Physics constant tuning
**Depends on:** TASK-021
**Acceptance criteria:**
- [ ] Terminal velocity at Full Ahead throttle is in range 4–8 kt (~2–4 m/s) — visible play area not escaped
- [ ] Beam-on 20 kt wind drift rate for stationary boat ≤ 0.1 m/s (keel drag dominates)
- [ ] Mooring line does not oscillate visibly — spring+damper is critically or over-damped
- [ ] Prop walk at Full Astern causes ≥ 2°/s yaw rate at zero forward speed
- [ ] Catamaran dual-engine differential thrust causes ≥ 3°/s yaw at max differential
- [ ] No physics instability (NaN, Infinity) after 5 minutes of continuous play with all forces active
- [ ] Constants updated in `src/main.ts` if any tuning was needed; document final values in inline comment block near constant declarations
**Definition of done:** all criteria checked; final constants noted

---

## Dependency Summary

```
TASK-001 (interfaces + constants)
  ├── TASK-002 (getCleatWorld + naturalLength)
  ├── TASK-003 (dt tracking)
  │
  TASK-003 → TASK-010 (physicsStep skeleton + Euler)
               ├── TASK-011 (drag)
               ├── TASK-012 (thrust + propwalk)   ← monohull + catamaran
               ├── TASK-013 (rudder)
               ├── TASK-014 (wind)
               ├── TASK-015 (mooring springs)     ← needs TASK-002
               └── TASK-016 (collision)
                          ↓
                    TASK-020 (tsc check)
                          ↓
                    TASK-021 (smoke tests)
                          ↓
                    TASK-022 (constant tuning)
```

All tasks in Group 1 (TASK-011..016) are independent of each other — they add separate force
sections to the same accumulator. They can be implemented in any order or in parallel within
a single editing session.
