# Interfaces — Existing Contracts

*What cannot be broken when adding physics.*

---

## Public Interfaces / Types

### `Boat` — must preserve all existing fields
```typescript
interface Boat {
  type: BoatType;
  x: number;
  y: number;
  heading: number;
  rudderAngle: number;
  throttlePort: number;
  throttleStbd: number;
}
```
**Why locked**: every field is read by `drawBoat`, `syncThrottleUI`, `syncRudderUI`, `boatToOBB`,
`getMovingCleatCanvas`, event handlers in `mousedown`/`drop`.
**Extension strategy**: add new optional fields (`vx?: number`, `vy?: number`, `omega?: number`),
defaulting to 0 when absent — no existing code breaks.

### `CleatRef` / `MooringLine` — structural, used everywhere
```typescript
type CleatRef =
  | { kind: 'static'; idx: number }
  | { kind: 'moving'; boat: number; cleat: number };
interface MooringLine { from: CleatRef; to: CleatRef; }
```
**Why locked**: `deleteBoat`, `hitTestCleats`, `cleatEq`, `canConnect`, `drawMooringLines`
all pattern-match on `kind`. Adding a new `CleatRef` variant requires updating all these.

### `OBB` — used by existing SAT collision
```typescript
interface OBB { x: number; y: number; w: number; h: number; heading: number; }
```
`obbMTV` and `boatToOBB` depend on this shape. Physics may reuse or extend collision
but must not change this interface.

---

## Coordinate System Contract

- World space: Y-up, meters. `boat.x/y` are world meters.
- `heading = 0` means bow pointing in the +Y world direction (north on screen = up).
- `windAngle = 0` means wind blowing toward north.
- These conventions are embedded in `worldToCanvas`, `canvasToWorld`, SAT axis formulas,
  drag handlers, and rotation handle math. **Must not change.**

---

## Function Contracts Locked by UI

| Function | Callers | What it must keep doing |
|----------|---------|------------------------|
| `syncControlPanels()` | `playBtn` click, `mousedown`, `deleteBoat` | Show/hide rudder+throttle panels |
| `syncRudderUI()` | `syncControlPanels` | Sync slider from `boat.rudderAngle` |
| `syncThrottleUI()` | `syncControlPanels` | Sync sliders from `boat.throttlePort/Stbd` |
| `worldToCanvas(wx,wy)` | `drawBoat`, `drawStaticCleats`, `getCleatCanvas` | Must remain pure, no side-effects |
| `canvasToWorld(cx,cy)` | `mousedown`, `mousemove`, `drop` | Inverse of `worldToCanvas` |
| `render()` | `requestAnimationFrame` self-loop | Must continue calling `requestAnimationFrame(render)` |
| `deleteBoat(idx)` | `endDrag` | Must keep mooring line index shift logic intact |
| `rebuildStaticCleats()` | `resize` | Must keep populating `staticCleats[]` |

---

## Rendering Contract

`drawBoat(boat, isActive)` reads `boat.x/y/heading/rudderAngle/type` on every frame.
Physics writes to `boat.x/y/heading` — no separation needed; same reference.

`revertAnim` only mutates `boat.x/y/heading` during its animation window.
Physics must **not run** while `revertAnim` is active for that boat index, or must be disabled
during revert entirely (easier: `revertAnim` is a setup-mode-only feature, so won't conflict
since physics only runs in play mode).

---

## Behaviour Contracts for Physics

| Contract | Source |
|----------|--------|
| Physics only runs when `gameMode === 'play'` | User feature matrix |
| Only the active boat (`activeBoatIdx`) is physics-driven | User statement |
| Rudder/throttle inputs come from existing sliders (no new inputs needed) | UI already built |
| `windAngle` / `windKt` are the only global wind inputs | Existing state |
| Mooring lines run between `CleatRef` endpoints (positions already computed by `getCleatCanvas`) | Existing code |
| Pier collision boundary is the existing OBB (infinite downward depth, top = waterline − 5 m) | `getPierOBB()` |
