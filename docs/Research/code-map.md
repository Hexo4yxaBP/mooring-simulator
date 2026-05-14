# Code Map — Physics-Relevant Sections

*All code is in `src/main.ts`. Line numbers are approximate (single 983-line file).*

---

## State Variables (physics-relevant)

| Variable | Type | Location | Role |
|----------|------|----------|------|
| `boats` | `Boat[]` | line ~43 | All boat entities; position/heading mutated by drag today |
| `activeBoatIdx` | `number \| null` | line ~44 | Physics applied only to this boat |
| `windAngle` | `number` | line ~57 | Global wind direction (rad, 0=blowing N) |
| `windKt` | `number` | line ~58 | Global wind speed (knots) |
| `gameMode` | `'setup'\|'play'` | line ~61 | Physics runs only in 'play' |
| `mooringLines` | `MooringLine[]` | line ~53 | Line connections; no spring force yet |
| `staticCleats` | `StaticCleat[]` | line ~52 | Pier/buoy cleat world positions |
| `SCALE` | `20` | line ~4 | Pixels per world-meter |
| `BOAT_SIZE` | record | line ~6 | `{w,h}` in world-meters per type |

---

## Interfaces

### `Boat` (line ~13)
```typescript
interface Boat {
  type: 'monohull' | 'catamaran';
  x: number;           // world m — mutable by physics
  y: number;           // world m — mutable by physics
  heading: number;     // rad — mutable by physics
  rudderAngle: number; // rad ±35° — input to physics
  throttlePort: number; // 0..4 — input to physics
  throttleStbd: number; // 0..4 — catamaran only
}
// Physics needs to add: vx, vy (m/s), omega (rad/s)
```

### `MooringLine` / `CleatRef` (line ~38)
```typescript
interface MooringLine { from: CleatRef; to: CleatRef; }
type CleatRef =
  | { kind: 'static'; idx: number }       // index into staticCleats[]
  | { kind: 'moving'; boat: number; cleat: number }; // boat index + cleat index
```

### `OBB` (line ~31)
```typescript
interface OBB { x: number; y: number; w: number; h: number; heading: number; }
```

---

## Key Functions

### Coordinate Transforms (line ~248)
```typescript
worldToCanvas(wx, wy): [cx, cy]
canvasToWorld(cx, cy): [wx, wy]
```
Physics operates in world space; draw functions convert to canvas space.

### Collision (line ~169)
```typescript
obbMTV(a: OBB, b: OBB): [number, number] | null
// Returns world-space MTV to push `a` out of `b`.
// Axes derived from heading: bow=(sin h, cos h), beam=(cos h, -sin h).
```
Currently only called in drag validation (`findNearestValid`, `anyOverlap`).
**Not integrated into physics loop.** For physics, collision response force = MTV × spring constant, or impulse.

### Pier OBB (line ~158)
```typescript
getPierOBB(): OBB
// Center: (0, pierTopY − h/2) where pierTopY = −(ch/2)/S + PIER_DEPTH_M
// Width: canvas.width/S + 40 (overhangs screen)
// Height: PIER_DEPTH_M + 1000 (extends deep down → MTV always pushes boat upward)
```

### Cleat World Positions (line ~332)
```typescript
getMovingCleatCanvas(boat, cleat): [cx, cy]  // canvas space only
getCleatCanvas(ref): [cx, cy]                 // canvas space only
```
**No world-space equivalent exists.** Physics needs `getCleatWorld(ref): [wx, wy]`.
Formula: apply inverse of `worldToCanvas` to canvas position, OR compute directly from boat.x/y + local offset rotated by boat.heading.

### Hull Transform (line ~390 in `drawBoat`)
```typescript
ctx.save();
ctx.translate(cx, cy);     // boat canvas centre
ctx.rotate(boat.heading);  // canvas rotation (Y-down, so positive = CW)
ctx.drawImage(img, -(w*S)/2, -(h*S)/2, w*S, h*S);
// Then hull outline:
ctx.transform(scaleX, 0, 0, -scaleY, tx*scaleX-(w*S)/2, ty*scaleY-(h*S)/2);
ctx.stroke(meta.paths[i]);
```
The `scale(1, -1)` in the hull transform flips SVG Y (up) to canvas Y (down).
**For physics polygon sampling**: apply the same transform chain to `Path2D` sample points → canvas coords → `canvasToWorld` → world polygon vertices.

### Render Loop (line ~669)
```typescript
function render(): void {
  // advance revertAnim
  drawWater(); drawPier(); drawStaticCleats();
  boats.forEach((boat, i) => drawBoat(boat, i === activeBoatIdx));
  drawMooringLines(); drawBin(); drawWindArrow();
  requestAnimationFrame(render);
}
```
**Physics update must be inserted before the draw calls, with `dt` computed from `performance.now()`.**

---

## Hull Geometry Data

### `OUTLINE_PATHS` (line ~118)
```
monohull:  svgW=14.51, svgH=40.28, tx=1.364, ty=40.142
  path: M 11.786 0 L 0 0 C -2.780,16.292 -0.869,29.703 5.893,40 C 12.655,29.703 14.566,16.292 11.786,0 Z
catamaran: svgW=26.38, svgH=40.52, tx=1.410, ty=40.259
  path[0] (left hull):  M 7.040 0 L 0 0 C -2.996,17.083 0.422,30.154 3.520,40 C 6.618,30.154 10.036,17.083 7.040,0 Z
  path[1] (right hull): M 16.525 0 L 23.565 0 C 26.561,17.083 23.143,30.154 20.045,40 C 16.947,30.154 13.529,17.083 16.525,0 Z
  path[2] (deck rect):  M 7.346,1.873 L 16.219,1.873 L 17.237,29.731 L 6.328,29.731 Z
```

Both hulls map via `scaleX = (w·SCALE)/svgW`, `scaleY = (h·SCALE)/svgH` and flip `scale(1,−1)`.

Bezier control points (raw, for sampling):
- Monohull port side: `C −2.780,16.292 −0.869,29.703 5.893,40` (cubic, from stern to bow)
- Monohull starboard side: `C 12.655,29.703 14.566,16.292 11.786,0` (cubic, from bow to stern)

### `CLEAT_SVG` (line ~96)
SVG-path-space positions; same coordinate system as `OUTLINE_PATHS`.

Monohull (6 cleats):
- [0] `(0, 0)` — port stern
- [1] `(11.786, 0)` — starboard stern
- [2] `(−1.085, 18)` — port midship
- [3] `(12.871, 18)` — starboard midship
- [4] `(3.585, 36)` — port bow
- [5] `(8.202, 36)` — starboard bow

Catamaran (6 cleats):
- [0] `(0, 0)` — port stern (left outer)
- [1] `(23.565, 0)` — starboard stern (right outer)
- [2] `(−0.831, 20)` — port midship
- [3] `(24.396, 20)` — starboard midship
- [4] `(3.520, 40)` — left hull bow tip
- [5] `(20.045, 40)` — right hull bow tip

**Local canvas position** of cleat `[sx, sy]` relative to boat centre:
```
lx = (sx + tx) * scaleX − (w·S)/2
ly = (ty − sy) * scaleY − (h·S)/2
```
**World position** = rotate `(lx, ly)` by `heading`, add boat canvas centre, convert to world.

---

## Extension Points

- `Boat` interface — add `vx`, `vy`, `omega` fields here.
- `render()` — insert `physicsStep(dt)` call before draw calls.
- `mooringLines` — `MooringLine` is unchanged; spring force computed from cleat world positions.
- `gameMode` guard — physics update conditional on `gameMode === 'play'`.
- `activeBoatIdx` — identifies which boat receives physics; other boats static (or wind-drift only).
