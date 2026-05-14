# AGENTS.md — Mooring Simulator

Agent and contributor reference for this repository.

---

## Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Language | TypeScript | 5.x (strict mode) |
| Bundler | Vite | latest |
| Rendering | HTML5 Canvas 2D API | browser built-in |
| Physics | Custom inline (no library) | in `src/main.ts` |
| UI | HTML + CSS overlay panels | `index.html` |

**No external physics library. No server. No test runner (yet).**
The entire app is `src/main.ts` (~1100 lines after physics) + `index.html`.

---

## Build and Serve Commands

```bash
# Install dependencies (first time)
npm install

# Development server with hot reload
npx vite
# or:
make serve

# Production build (outputs to dist/)
npx vite build

# Type check only (no emit)
npx tsc --noEmit
```

Dev server runs at `http://localhost:5173` by default (Vite) or `http://localhost:8080` if using `make serve`.

---

## Test Commands

The project has no automated test runner. Verification happens at two levels:

```bash
# 1. Type check (required before every commit)
npx tsc --noEmit

# 2. Lint (if eslint is configured)
npx eslint src/main.ts

# 3. Production build (catches bundler errors)
npx vite build
```

Manual browser smoke tests are defined in `docs/Plan/implementation-tasks.md` §TASK-021.

---

## Code Conventions

### Single-file architecture
All application logic lives in `src/main.ts`. Physics constants, interfaces, helper functions,
event handlers, and the render loop are all in this file. Do not create additional `.ts` files
without a deliberate refactoring task.

### Coordinate system (critical)
- **World space**: Y-up, meters. Origin at canvas centre. `SCALE = 20 px/m`.
- **Canvas space**: Y-down, pixels. Origin at canvas top-left.
- Transforms: `worldToCanvas(wx, wy)` and `canvasToWorld(cx, cy)` (lines ~248–260 in `src/main.ts`).
- `boat.heading = 0` means bow pointing north (world +Y).
- Bow unit vector in world space: `(sin(heading), cos(heading))`.
- Starboard unit vector: `(cos(heading), -sin(heading))`.
- **These conventions are used everywhere including OBB SAT, physics forces, and draw code. Do not change them.**

### Validation boundary
- User inputs (sliders, click positions) are validated and clamped in the event handlers.
- Physics functions trust their inputs. Do not add re-validation inside `physicsStep`.

### Typing rules
- TypeScript strict mode is on: `"strict": true` in tsconfig.
- `noUnusedLocals: true`, `noUnusedParameters: true` — every declared variable and parameter must be used.
- No `any` types in new code. Use proper types or `unknown` with a type guard.
- Interface extensions use optional fields (`?`) to maintain backward compatibility with all existing `boats.push(...)` call sites.

### Physics constants
All tunable physics constants are declared as module-level `const` near the top of `src/main.ts`,
alongside `SCALE`, `CLEAT_R`, etc. Do not hard-code physics values (force magnitudes, drag
coefficients, mass) inside `physicsStep` or other functions.

### Naming
- Constants: `SCREAMING_SNAKE_CASE`
- Functions: `camelCase`
- Interfaces: `PascalCase`
- Event listener blocks: inline anonymous functions (existing convention)

### No comments on what the code does
Add a comment only when the WHY is non-obvious: a hidden constraint, a coordinate convention,
a derivation formula, or a workaround. Do not describe what a function does if its name and
types already convey that.

### Commit conventions
```
<type>(<scope>): <short description>

type: feat | fix | test | refactor | docs | build
scope: physics | render | ui | input | docs
```
Example: `feat(physics): add mooring line spring forces`

---

## Physics Implementation Reference

After physics is implemented, `physicsStep(dt)` is called in `render()` before draw calls:

```typescript
// In render():
const dt = Math.min((now - lastTime) / 1000, 0.1);
lastTime = now;
if (gameMode === 'play' && activeBoatIdx !== null) {
  physicsStep(dt);
}
```

Force accumulation order inside `physicsStep`:
1. Hydrodynamic drag
2. Engine thrust + propeller walk
3. Rudder force
4. Wind force
5. Mooring line spring forces
6. Collision penalty forces

Integration: semi-implicit Euler (velocity updated before position).

See `docs/Design/interface-spec.md` for all physics constants and their rationale.
See `docs/Design/architecture.md` for the full sequence diagram.

---

## MVP Scope Boundary

In-scope for current physics implementation:
- Full physics on active boat only
- Wind force, engine thrust, prop walk, rudder, mooring springs, OBB collision

Deferred (do not implement without a new task):
- Wind drift on inactive boats
- Inextensible (rigid) mooring lines
- Bezier polygon hull collision (hull-accurate vs OBB)
- User-configurable boat parameters

See `docs/Design/architecture.md` §MVP Scope Boundary.
