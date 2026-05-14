---
updated: 2026-05-14
supersedes: original data-schemas.md (was Go boundary transforms; actual codebase is TypeScript)
---

# Data Schemas — Physics State Extension

---

## Boat State Extension

### New fields added to `Boat` interface

| Field | Type | Default | Unit | Initialised by |
|-------|------|---------|------|---------------|
| `vx` | `number?` | 0 | m/s world-X | `physicsStep` on first call |
| `vy` | `number?` | 0 | m/s world-Y | `physicsStep` on first call |
| `omega` | `number?` | 0 | rad/s | `physicsStep` on first call |

`vx`, `vy` use the same Y-up world coordinate system as `boat.x`, `boat.y`.
Source: gap G1 — `docs/Research/constraints.md`. Coordinate system — `docs/Research/interfaces.md`.

### Existing fields unchanged

| Field | Unit | Notes |
|-------|------|-------|
| `x`, `y` | world m | mutated each frame by `physicsStep` |
| `heading` | rad, 0=bow north | mutated each frame by `physicsStep` |
| `rudderAngle` | rad, ±35° | read each frame; written only by rudder slider |
| `throttlePort` | integer 0..4 | read each frame; written only by throttle slider |
| `throttleStbd` | integer 0..4 | catamaran only; mirrors throttlePort for monohull (unused) |

---

## MooringLine State Extension

### New field added to `MooringLine` interface

| Field | Type | Unit | Initialised by |
|-------|------|------|---------------|
| `naturalLength` | `number` | world m | `mooringLines.push(...)` call |

`naturalLength` = distance between the two cleats at the moment the line is connected.
Computed by `computeNaturalLength(from, to)` inline at push time.

This is a **required** field (not optional) — all `MooringLine` objects must carry it.
There is only one push site in the codebase (`mousedown` handler, line ~733), so the migration is minimal.

Lines with `extension <= 0` (slack) produce zero force. Source: constraint P1 — `docs/Research/constraints.md`.

---

## Coordinate Transform at Physics Boundary

Physics forces are computed in **world space** (Y-up, meters). Rendering works in **canvas space** (Y-down, pixels). The boundary transform is:

```
canvas x  =  world x  ×  SCALE  +  canvas.width/2
canvas y  = −world y  ×  SCALE  +  canvas.height/2
```

Inverse (canvas → world):
```
world x = (canvas x − canvas.width/2) / SCALE
world y = (canvas.height/2 − canvas y) / SCALE
```

All physics force vectors remain in world space. No conversion to canvas space during physics.
Source: `docs/Research/code-map.md` §Coordinate Transforms.

---

## Force Vector Schema

Forces within `physicsStep` are accumulated as `(fx: number, fy: number, torque: number)`.

| Symbol | World X component | World Y component |
|--------|------------------|------------------|
| Bow unit vector | `sin(heading)` | `cos(heading)` |
| Starboard unit vector | `cos(heading)` | `−sin(heading)` |

Source: OBB SAT axes in `obbMTV()` (`docs/Research/code-map.md` §Collision).

### Force application to torque

```
torque += (rx × fy) − (ry × fx)
```
where `(rx, ry)` = world-space vector from boat centre to point of application.
This is the Z-component of the 2D cross product `r × F`.

---

## Throttle Integer → Thrust Newtons Mapping

| Slider value | Meaning | `THROTTLE_FORCE[v]` |
|---|---|---|
| 0 | Full astern | −10000 N |
| 1 | Slow astern | −4000 N |
| 2 | Neutral | 0 N |
| 3 | Slow ahead | +5000 N |
| 4 | Full ahead | +15000 N |

Source: `docs/Research/task-brief.md` §Throttle mapping.

---

## Propeller Walk — Lateral Force Mapping

```
// Starboard-positive lateral force on stern. Newtons.
PROP_WALK_TABLE = [-1500, -500, 0, 250, 500]
//                 ^astern            ahead^
```

| Boat type | Port engine lateral | Starboard engine lateral |
|-----------|--------------------|-----------------------|
| Monohull | `PROP_WALK_TABLE[throttlePort]` | — |
| Catamaran | `+PROP_WALK_TABLE[throttlePort]` | `−PROP_WALK_TABLE[throttleStbd]` |

Catamaran contra-rotation: when both throttles equal, net walk = 0.
Applied at each hull's stern centerline — generates additional yaw torque.
Source: constraint P4, P5 — `docs/Research/constraints.md`.

---

## Engine Application Points (Stern Positions)

Applied force points are computed from `CLEAT_SVG` data at startup (or inline per frame):

### Monohull

Rudder/prop attachment = `(CLEAT_SVG.monohull[0][0] + CLEAT_SVG.monohull[1][0]) / 2` = x 5.893, y=0 in SVG space.

Local canvas offset:
- `lx = (5.893 + 1.364) × 4.134 − 30 ≈ 0 px` (centerline)
- `ly = (40.142 − 0) × 4.964 − 100 ≈ 99.3 px`

World offset from boat centre: `(0, −4.96)` m — stern, 5 m aft.

### Catamaran

Port engine at SVG `(3.520, 0)`, starboard at `(20.045, 0)`:

| Engine | Local canvas (lx, ly) px | World offset (wx, wy) m |
|--------|--------------------------|------------------------|
| Port | (−40.7, 108.6) | (−2.04, −5.43) |
| Starboard | (+40.7, 108.6) | (+2.04, −5.43) |

Source: `docs/Research/code-map.md` §Catamaran Engine Application Points.

For implementation efficiency, these offsets can be computed at runtime from existing helpers:
```typescript
const [cx, cy] = getMovingCleatCanvas(boat, /* port stern = cleat 0 */);
// then convert to world via canvasToWorld or the formula above
```
Or pre-compute once using the same transform used in `getCleatWorld`.

---

## Drag Model

Linear drag (F = C × v), computed in body frame then rotated to world:

```typescript
const vFwd = vx * bowX + vy * bowY;      // forward body-frame speed
const vLat = vx * stbdX + vy * stbdY;   // starboard body-frame speed
const dragFwd = C_DRAG_FWD * vFwd;       // opposes forward motion
const dragLat = C_DRAG_LAT * vLat;       // opposes lateral motion
// world-space drag force:
fx -= dragFwd * bowX + dragLat * stbdX;
fy -= dragFwd * bowY + dragLat * stbdY;
torque -= C_DRAG_ROT * omega;
```

Applied at centre of mass (no torque contribution beyond rotational drag).

---

## Omitted in This Iteration

| Feature | Reason |
|---------|--------|
| Inactive boat wind drift | User: "forces act only on the active boat" |
| Inextensible mooring lines | Requires constraint solver — deferred |
| Sail aerodynamics | Wind treated as bare-hull drag only |
| Variable boat mass | Fixed per type; no UI for tuning in scope |
| Wave effects | Not in task spec |
