---
updated: 2026-05-14
supersedes: original review-gaps.md (was Go multi-package; actual codebase is TypeScript)
---

# Review: Implementation Gaps

*Completeness check against docs/Design/architecture.md, interface-spec.md, data-schemas.md,
and access-control.md.*

---

## Missing Components

**None found.**

Checklist against architecture.md §MVP Scope Boundary:

| Component | Covered by task |
|-----------|----------------|
| Boat interface extension (vx/vy/omega) | TASK-001 |
| MooringLine interface extension (naturalLength) | TASK-001 |
| All physics constants block | TASK-001 |
| `getCleatWorld()` helper | TASK-002 |
| `computeNaturalLength()` helper | TASK-002 |
| `mooringLines.push()` update | TASK-002 |
| `dt` tracking in render loop | TASK-003 |
| Velocity reset on play mode entry | TASK-003 |
| `physicsStep()` skeleton + Euler integrator | TASK-010 |
| Hydrodynamic drag | TASK-011 |
| Engine thrust (mono + cat) | TASK-012 |
| Propeller walk (mono + cat, contra-rotating) | TASK-012 |
| Rudder force (zero-speed guard) | TASK-013 |
| Wind force (quadratic) | TASK-014 |
| Mooring line spring+damper (tension-only) | TASK-015 |
| Collision penalty (OBB reuse, 50% correction) | TASK-016 |
| TypeScript type check gate | TASK-020 |
| Browser smoke tests (9 scenarios) | TASK-021 |
| Physics constant tuning | TASK-022 |

---

## Tasks Without Tests

All Group 0–1 tasks include behavioral test criteria in their acceptance criteria.
Type-check (TASK-020) provides the automated verification gate.

| Task | Test type | Verification |
|------|-----------|-------------|
| TASK-001 | Compile-time | `npx tsc --noEmit` |
| TASK-002 | Compile-time + consistency check | Type check + console assertion |
| TASK-003 | Compile-time + observable | Type check + browser console |
| TASK-010 | Behavioral | Console `boats[0].vx = 1` observation |
| TASK-011 | Behavioral | Deceleration observable in browser |
| TASK-012 | Behavioral | Acceleration + yaw observable |
| TASK-013 | Behavioral | Turn at speed observable |
| TASK-014 | Behavioral | Lateral drift observable |
| TASK-015 | Behavioral | Line-hold observable |
| TASK-016 | Behavioral | Pier wall observable |
| TASK-020 | Automated | `tsc --noEmit` exit code 0 |
| TASK-021 | Manual (9 smoke tests) | Checklist in implementation-tasks.md |
| TASK-022 | Manual + stability | 5-minute run + constant review |

No task is entirely untested. TASK-021 (manual smoke) is the acceptance gate for the entire feature.

---

## Cyclic Dependencies

Dependency graph is a DAG:

```
TASK-001
  ↓       ↓
TASK-002  TASK-003
           ↓
         TASK-010 (physicsStep skeleton)
           ↓
    ┌──────┼──────┬──────┬──────────┐
  TASK-011 TASK-012 TASK-013 TASK-014
                                    TASK-015 (needs TASK-002)
                                    TASK-016
                                        ↓
                                    TASK-020
                                        ↓
                                    TASK-021
                                        ↓
                                    TASK-022
```

No cycles. TASK-011..016 are all independent additions to the same force accumulator.
TASK-015 has an additional dependency on TASK-002 (needs `getCleatWorld`).

---

## Security Requirements Without Tasks

From access-control.md §Option A (Natural Bounds via Physics Constants):

| Requirement | Task |
|-------------|------|
| `dt` clamped to [0, 0.1] to prevent dt spike | TASK-003 |
| Terminal velocity self-limited by drag constants | TASK-022 (tuning verifies bounds) |
| `dist < 1e-6` guard in spring force (prevent div/0) | TASK-015 acceptance criterion |
| `throttlePort/Stbd` index in [0,4] — existing slider enforces | Existing code (no new task) |
| `rudderAngle` in [±35°] — existing slider enforces | Existing code (no new task) |
| No new network calls, no localStorage, no DOM access beyond canvas | Implicit (no new APIs used) |

All security requirements from access-control.md are covered. No gaps.

---

## Blocked Tasks

**No blocked tasks.**

All design unknowns were resolved in Phase 2:

| Former unknown | Resolution | Unblocks |
|----------------|-----------|---------|
| Variable vs fixed timestep | Variable dt, clamped 0.1 s | TASK-003, TASK-010 |
| Hull collision type | OBB SAT reuse | TASK-016 |
| Mooring spring model | Hookean + linear damper | TASK-015 |
| Physics accuracy level | Tunable gameplay constants | TASK-001, TASK-022 |
| Catamaran engine model | Twin contra-rotating engines | TASK-012 |
| Active-only physics scope | Active boat only | TASK-010 (early return) |

---

## Risk Flags (Not Blockers)

| Risk | Impact | Mitigation | Task |
|------|--------|-----------|------|
| Mooring spring `MOORING_K/MOORING_C` ratio causes oscillation | Visible jitter on mooring lines | TASK-022 tuning; increase C if needed | TASK-022 |
| Catamaran lateral arm constant (2.04 m) is hard-coded | Wrong torque if SVG geometry changes | Inline comment with derivation formula; compute dynamically if SVG changes | TASK-012 |
| `physicsStep` in one 983-line file adds 100–150 lines | `src/main.ts` becomes large (~1100 lines) | Acceptable for now; Vite module split is a separate refactor task | — |
| Console assertions not repeatable | Behavioral tests require manual setup each time | Document in TASK-021 checklist; consider adding Vitest post-MVP | TASK-021 |

---

## Stale Documentation Warning

The following files describe a Go/WASM architecture that was **never implemented**. They should
not be referenced for implementation:

- `AGENTS.md` — Go toolchain commands (updated in Phase 3)
- `CLAUDE.md` — mentions "Go 1.24, Ebiten" (should be updated separately)
- `docs/Research/tech-options.md` — Go option analysis
- `docs/Design/access-control.md` (original) — superseded

All `docs/Research/` and `docs/Design/` files have been updated to reflect the TypeScript
codebase as of Phase 2 completion.
