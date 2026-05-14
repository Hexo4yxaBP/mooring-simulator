---
updated: 2026-05-14
supersedes: original test-strategy.md (was Go test runner; actual codebase is TypeScript, no test runner configured)
---

# Test Strategy — Physics Simulation

*Sources: architecture.md (D1–D8), interface-spec.md (value constraints), implementation-tasks.md (TASK-020..022).*

---

## Testing Philosophy

The physics subsystem is added to a single-file TypeScript app (`src/main.ts`). There is no
test runner configured. The testing strategy has three layers, in order of automation:

1. **Type check** — `npx tsc --noEmit` catches interface mismatches, missing fields, wrong types.
   This is the only automated test gate.
2. **Behavioral assertions** — each task in Group 1 defines in-code assertions as `console.assert`
   or inline spot checks that can be run from the browser DevTools console.
3. **Manual smoke tests** — TASK-021 defines 9 observable behaviors the developer verifies
   by hand in the browser. These are the acceptance gate for the full feature.

No test framework (Jest, Vitest, etc.) is introduced by this plan. Adding one is a separate task.

---

## Layer 1 — TypeScript Type Check (automated)

**Command:**
```
npx tsc --noEmit
```

**What it checks:**
- `Boat.vx/vy/omega` optional fields are properly initialised before use
- `MooringLine.naturalLength` is provided at every `push()` call site
- `getCleatWorld()` return type `[number, number]` is used correctly
- No `any` casts that would hide type errors
- `THROTTLE_FORCE[boat.throttlePort]` indexing is in-bounds (TypeScript cannot verify this at compile time, but value constraints in interface-spec.md guard it at runtime)
- Strict mode (`noUnusedLocals`, `noUnusedParameters`) enforced

**Coverage:** 100% of new interfaces and function signatures are type-checked at compile time.

**Threshold:** zero errors required before TASK-021 begins. One compiler error = task blocked.

---

## Layer 2 — Console Behavioral Assertions

Each force function has a documented "behavioral test" in its acceptance criteria. These are
verified by opening `http://localhost:5173` (Vite dev server), dropping a monohull in play mode,
then running assertions in the browser DevTools console.

### Assertion Template

```javascript
// Hydrodynamic drag test (TASK-011)
boats[0].vx = 0;
boats[0].vy = 2;
boats[0].vx = 0;
boats[0].omega = 0;
// call physicsStep manually is not possible (it's module-private),
// but the assertion is: after ~1 second of play, vy decreases from 2 toward 0
// verified by watching boats[0].vy in console
```

Since `physicsStep` is module-scoped (not exported), behavioral assertions are verified by
observation rather than direct function calls. The key behaviors are:

| Behavior | Observable indicator |
|----------|---------------------|
| Drag decelerates boat | `boats[0].vy` decreases over time |
| Thrust accelerates | Speed at full throttle reaches ~2-4 m/s |
| Prop walk yaws | `boats[0].omega` non-zero at full astern |
| Rudder turns boat | Heading changes at speed with rudder applied |
| Wind drifts boat | Position changes slowly in wind direction |
| Line holds boat | Position oscillates near natural length, not beyond |
| Pier stops boat | Boat heading toward pier decelerates and stops |

---

## Layer 3 — Manual Smoke Tests (TASK-021)

Nine smoke tests defined in TASK-021, each covering one physics feature.
These are the **acceptance gate** — all nine must pass before implementation is considered done.

### Smoke Test Checklist

| # | Feature | Pass signal | Fail signal |
|---|---------|-------------|-------------|
| 1 | Stationary in play mode | Boat does not drift | Boat moves without input |
| 2 | Throttle / terminal velocity | Boat reaches ~5 m/s, holds | Boat accelerates forever or barely moves |
| 3 | Rudder at speed | Heading changes smoothly | No turn, or violent spin |
| 4 | Prop walk (astern) | Yaw with near-zero forward speed | No rotation, or rotation is wrong direction |
| 5 | Wind drift | Very slow lateral drift, boat does not rotate significantly | Fast drift or no drift at all |
| 6 | Mooring holds boat | Boat stays within ~2 m of attach point | Line passes through, or oscillates wildly |
| 7 | Pier collision | Boat decelerates at pier face | Boat passes through pier |
| 8 | Catamaran differential throttle | Boat turns toward lower-throttle engine side | No turn, or turn toward wrong side |
| 9 | Mode switch | Velocity reset to zero on re-entry | Boat teleports or keeps phantom velocity |

---

## What Cannot Be Tested Without a Test Runner

| Concern | Risk | Mitigation |
|---------|------|-----------|
| Spring constant stability at edge dt | Oscillation if `MOORING_K/MOORING_C` ratio is large | TASK-022 5-minute stability test |
| NaN propagation from division by zero | `dist < 1e-6` guard in TASK-015 | Type check + guard in code |
| Throttle index out of bounds | `boat.throttlePort` outside 0..4 | Existing slider clamps; documented in interface-spec.md |
| Catamaran lateral arm constant hard-coded | Wrong torque if SVG changes | Commented with derivation; recalculate if SVG changes |

---

## Recommended Future Step: Add Vitest

After physics is implemented and smoke-tested, adding [Vitest](https://vitest.dev/) would allow:
- Direct unit tests on `physicsStep` by factoring out pure force functions
- Automated regression on constant changes
- Coverage measurement

This is out of scope for the current task but would be the natural next evolution.
Estimated setup effort: 2 hours to configure + extract pure functions.

---

## Build and Check Commands

```bash
# Type check (required before each commit)
npx tsc --noEmit

# Dev server (for smoke tests)
make serve
# or:
npx vite

# Production build (Vite bundles TypeScript)
npx vite build

# Lint (if eslint is configured)
npx eslint src/main.ts
```

---

## Coverage Threshold

| Layer | Target |
|-------|--------|
| Type check (`tsc --noEmit`) | 0 errors (hard gate) |
| Behavioral assertions | All 7 table entries observable |
| Manual smoke tests | All 9 smoke tests pass |
| Physics stability (5-min run) | Zero NaN/Infinity in `boat.x/y` |
