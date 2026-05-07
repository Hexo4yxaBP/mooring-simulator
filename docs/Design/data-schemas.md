# Data Schemas and Normalization

*This is an all-Go WASM application with no external API. "Normalization" here means transformations at package boundaries: raw input → semantic commands, physics state → render data, screen coordinates → world coordinates.*

*Sources: domain-model.md (entity fields), interface-spec.md (types), architecture.md (D1/D2).*

---

## Boundary 1 — User Input → Commands

Raw Ebiten input (pixel coordinates, key codes, mouse buttons) is mapped to typed game Commands.

### Screen Position → World Position

Used when user clicks to place a mooring line dock point (constraint C5) or selects a boat.

```
worldPos.X = screenX / viewport.Scale + viewport.OriginWorld.X
worldPos.Y = (viewport.ScreenH - screenY) / viewport.Scale + viewport.OriginWorld.Y
```

Y-axis is flipped at this boundary (screen Y-down → world Y-up, decision D2).

**Snap rule for dock attachment:** The raw world position from click is projected to the nearest point on the dock boundary polygon edge. This enforces that dock-end mooring points are always on the dock surface (constraint C5).

```
dockPoint = dock.NearestPointOnEdge(rawWorldPos)
```

**Snap rule for boat cleats:** User clicks within a `CleatHitRadius` (= 0.5 m world-space, ~10px at default scale) of a cleat dot. The closest cleat ID within radius is selected; clicks outside all radii are ignored (constraint C4).

### Throttle State Encoding

| User action | ThrottleState int | Label |
|-------------|------------------|-------|
| Key `4` / W from Full Fwd | 0 | Neutral |
| Key `3` / W once from Neutral | 1 | Slow Forward |
| Key `2` / W twice from Neutral | 2 | Full Forward |
| Key `1` / S once from Neutral | 3 | Slow Astern |
| Key `0` / S twice from Neutral | 4 | Full Astern |

Throttle cycles up (W) and down (S); clamped at edges (no wrap).

### Rudder Angle Encoding

Raw: keyboard held A/D. Each tick adds/subtracts `RudderStepPerTick = 0.0349 rad (2°)`.  
Clamped to `[-MaxRudderAngle, +MaxRudderAngle]` = `[-0.6109 rad, +0.6109 rad]` = `±35°`.  
Positive = starboard turn (right), matching domain-model.md Boat.rudderAngle definition.

### Wind Input

Wind speed: mouse-wheel scroll on panel → ± 0.5 m/s per notch, clamped `[0, 30]` m/s.  
Wind direction: click-drag on compass dial → angle in radians.  
Both validated on input; no physics constants change.

---

## Boundary 2 — Physics State → Render Data

The renderer reads `sim.World` directly (no intermediate DTO). Transforms:

### World Position → Screen Position

```
screenX = (worldPos.X - viewport.OriginWorld.X) * viewport.Scale
screenY = viewport.ScreenH - (worldPos.Y - viewport.OriginWorld.Y) * viewport.Scale
```

### Heading → Draw Rotation

Ebiten `DrawImageOptions.GeoM.Rotate(angle)` uses clockwise positive.  
Physics heading is CCW positive (math convention).  
Transform: `ebitenAngle = -heading` (negate for screen-space CW convention).

### Hull Polygon (world-space → screen-space)

Monohull is approximated as an oriented rectangle with bow taper. Vertices in body frame:

```
// Body frame, CoM at origin, bow = +x direction
halfLen  = HullLen / 2
halfBeam = HullBeam / 2
tapering = HullLen * 0.15  // bow taper offset

vertices_body = [
    {+halfLen - tapering, 0},         // bow tip (centerline)
    {+halfLen * 0.6, +halfBeam},      // bow starboard shoulder
    {-halfLen, +halfBeam * 0.8},      // stern starboard
    {-halfLen, -halfBeam * 0.8},      // stern port
    {+halfLen * 0.6, -halfBeam},      // bow port shoulder
]
```

Transform to world-space: rotate by heading, translate by CoM world position.  
Transform to screen-space: apply Viewport.WorldToScreen to each vertex.

Catamaran hull: two rectangles (no taper), spaced `HullBeam * 0.6` apart, connected by a cross-beam rectangle. Same rotation/translation.

### CoM Dot Position

```
comWorldPos = body.Position + rotate(body.ComOffset, body.Heading)
```

Drawn as filled circle, radius 4px, distinct color (white or red depending on boat active state).

### Mooring Line Color by Tension

| Tension / MaxTension | Color |
|---------------------|-------|
| 0 (slack) | `#888888` grey |
| 0–0.3 | `#88CC88` green |
| 0.3–0.7 | `#CCCC00` yellow |
| 0.7–1.0 | `#CC4400` orange |
| ≥ 1.0 (overstressed) | `#FF0000` red |

`MaxTension` = `line.NaturalLength * line.Stiffness * 0.5` (50% extension = reference max for display scaling).

---

## Boundary 3 — Simulation Constants (physics/constants.go)

All physics constants are tunable floats defined in one file. No SI lookup tables — values are chosen for gameplay feel.

### Thrust Table (Monohull, Newtons)

| ThrottleState | Value | Notes |
|---------------|-------|-------|
| Neutral | 0 | |
| SlowForward | +800 | ~10% of full |
| FullForward | +8000 | ~typical 30ft sailboat engine |
| SlowAstern | -600 | Astern less efficient |
| FullAstern | -5000 | |

### Prop Walk Table (fractional multiplier on boat.PropWalk)

| ThrottleState | Multiplier |
|---------------|-----------|
| Neutral | 0.0 |
| SlowForward | 0.15 |
| FullForward | 0.25 |
| SlowAstern | 0.8 |
| FullAstern | 1.0 |

Constraint P4: astern multipliers >> forward multipliers.

### Default Boat Parameters (Monohull)

| Parameter | Value |
|-----------|-------|
| Mass | 5000 kg |
| HullLen | 10 m |
| HullBeam | 3 m |
| MomentOfInertia | `mass * (len² + beam²) / 12` ≈ 44,750 kg·m² |
| ComOffset | `{-0.5, 0}` m (slightly aft of center) |
| Cleat Bow | `{+4.5, 0}` m (body frame) |
| Cleat Midships | `{0, 0}` m |
| Cleat Stern | `{-4.5, 0}` m |
| PropWalk default | 300 N (right-handed prop = walks to port in astern) |

### Default Boat Parameters (Catamaran, passive)

| Parameter | Value |
|-----------|-------|
| Mass | 8000 kg |
| HullLen | 12 m |
| HullBeam | 6 m (overall) |
| ComOffset | `{0, 0}` m (symmetric) |
| PropWalk | 0 (MVP: no engine) |

### Drag Coefficients (Monohull)

| Coefficient | Value | Ratio |
|-------------|-------|-------|
| LongDrag (fwd/aft) | 300 N·s/m | 1× |
| LatDrag (port/stbd) | 6000 N·s/m | 20× |
| RotDrag | 50000 N·m·s/rad | — |

Constraint P3: LatDrag/LongDrag ratio ≈ 20. Boats move easily forward, resist sideways motion.

### Wind Drag Coefficients (Monohull)

| Coefficient | Value |
|-------------|-------|
| LongWind | 15 m² (bow-on area) |
| LatWind | 80 m² (beam-on area, includes sail) |

### Mooring Line Defaults

| Parameter | Value |
|-----------|-------|
| Stiffness | 5000 N/m |
| Damping | 500 N·s/m |
| NaturalLength | 0.95 × initial distance at placement |

### Collision Penalty

| Parameter | Value |
|-----------|-------|
| PenaltyStiffness | 100000 N/m |

---

## Dropped Fields / Intentional Omissions (MVP)

| Field | Reason for omission |
|-------|-------------------|
| Catamaran engine controls | MVP scope decision — catamaran is passive (constraints.md D4/confirmed decisions) |
| Inextensible line mode | Requires constraint solver; deferred post-MVP |
| Boat parameter editor | Fixed defaults for MVP (constraints.md U10) |
| Audio | Not in spec; Ebiten audio deferred |
| Save/load state | Not in spec |
| Wind variation / gusts | Global constant wind only (constraints.md U11) |
