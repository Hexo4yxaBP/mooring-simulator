---
updated: 2026-05-14
context: Physics subsystem "access control" = physics value bounds control.
         Traditional auth/session security is not applicable (client-side game, no server,
         no user accounts — see original access-control.md §Security Context).
---

# Physics Bounds Control

*In a client-side single-player game there is no authentication boundary. The relevant "access
control" question for a physics simulation is: what prevents unbounded growth of state values
that would cause numerical instability, invisible boats, or browser hangs?*

---

## Option A — Natural Bounds via Physics Constants (No Explicit Clamps)

**Strategy:** Choose drag coefficients large enough that terminal velocity is reached well within
the visible play area. Let physics self-limit without runtime clamps.

**Terminal velocity analysis:**

At full-ahead throttle (15000 N), the boat accelerates until drag equals thrust:
```
C_DRAG_FWD × v_terminal = THROTTLE_FORCE[4]
3000 × v_terminal = 15000
v_terminal ≈ 5 m/s ≈ 9.7 kt
```

At that speed, 1 second of travel = 5 m = 100 px — well within the visible canvas.

Lateral terminal velocity (e.g. from wind beam-on at 15 kt):
```
C_DRAG_LAT × v_lat = WIND_K_LAT × (7.7 m/s)²
80000 × v_lat ≈ 1600
v_lat ≈ 0.02 m/s — near-zero
```

Keel drag makes lateral drift negligible. Physics will not produce runaway lateral motion.

**Angular terminal velocity** (from prop walk at stern, moment arm ~5 m):
```
torque_propwalk = 1500 N × 5 m = 7500 N·m
C_DRAG_ROT × ω_terminal = 7500
50000 × ω_terminal = 7500
ω_terminal ≈ 0.15 rad/s → full rotation in ~42 s
```

Reasonable.

**Pros:**
- Zero implementation overhead — no clamp logic
- Constants can be tuned together: lowering drag and raising thrust both increase terminal velocity proportionally
- Correct physics behaviour: boat decelerates naturally, no step-function cutoffs

**Cons:**
- If constants are mistuned, no safety net (boat could escape the visible area)
- `dt` spike from tab-switch (mitigated by 0.1 s clamp — see architecture.md D2) could momentarily overshoot terminal velocity

**Audit:** None needed — there is no security impact.

---

## Option B — Explicit Velocity Clamps After Integration

**Strategy:** After each `physicsStep` integration, clamp:
```typescript
const MAX_V = 10;    // m/s linear speed
const MAX_ω = 1.0;  // rad/s angular speed

const speed = Math.sqrt(boat.vx**2 + boat.vy**2);
if (speed > MAX_V) {
  const s = MAX_V / speed;
  boat.vx *= s;
  boat.vy *= s;
}
boat.omega = Math.max(-MAX_ω, Math.min(MAX_ω, boat.omega));
```

**Pros:**
- Hard guarantee: boat never goes faster than visible play requires
- Defensive against future constant changes that accidentally miscalibrate drag

**Cons:**
- Adds 6 lines of code and two tunable constants
- Creates a discontinuity: near the clamp boundary, adding more throttle has zero effect (unphysical feel)
- The speed discontinuity is noticeable on rudder response: at `MAX_V` constant speed, reducing throttle causes instant deceleration feel

---

## Recommendation: **Option A (natural bounds)**

With the constants defined in `interface-spec.md`, the physics self-limits at physically plausible speeds. The 0.1 s `dt` clamp (architecture.md) is the only defensive measure needed.

**Why not B:** The velocity clamp introduces a non-physical "wall" at the speed limit that players will notice, particularly when changing rudder at full speed. The cost of debugging a miscalibrated constant is low (tweak one number); the cost of removing a clamp that causes visible artefacts is hidden design debt.

**Monitoring:** If constants are later adjusted and the boat escapes the view area, a simple `boat.x` / `boat.y` world-bounds check at the start of `physicsStep` can reset velocity (not position) to zero — a softer recovery than a hard clamp.

---

## Browser Security Boundary (informational)

The WASM/TypeScript sandbox is the relevant security layer. The physics code:
- Cannot access the filesystem, network, or DOM outside the `canvas` element
- Cannot persist state between sessions (no localStorage writes)
- Runs in the same browser origin sandbox as any JS bundle

This is unchanged from the original `access-control.md` §WASM Security Notes — no new attack surface is introduced by adding physics.
