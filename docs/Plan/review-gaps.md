# Review: Implementation Gaps

*Completeness check against docs/Design/architecture.md, interface-spec.md, and access-control.md.*

---

## Missing Components

**None found.**

Checklist against architecture.md package structure:

| Component | Covered by tasks |
|-----------|-----------------|
| `internal/physics/vec2.go` | TASK-010 |
| `internal/physics/body.go` | TASK-011, TASK-012 |
| `internal/physics/forces.go` | TASK-013, TASK-014 |
| `internal/physics/mooring.go` | TASK-015 |
| `internal/physics/collision.go` | TASK-016 |
| `internal/sim/types.go` | TASK-020 |
| `internal/sim/constants.go` | TASK-021 |
| `internal/sim/boat.go` | TASK-022 |
| `internal/sim/dock.go` | TASK-023 |
| `internal/sim/world.go` (lines) | TASK-024 |
| `internal/sim/world.go` (Step) | TASK-025 |
| `internal/sim/world.go` (AddLine/Remove) | TASK-026 |
| `internal/sim/world.go` (SetActive) | TASK-027 |
| `internal/input/commands.go` | TASK-030 |
| `internal/input/handler.go` (keyboard) | TASK-031 |
| `internal/input/handler.go` (mouse) | TASK-032, TASK-033 |
| `internal/input/handler.go` (validation) | TASK-034 |
| `internal/render/viewport.go` | TASK-040 |
| `internal/render/dock.go` + background | TASK-041 |
| `internal/render/lines.go` | TASK-042 |
| `internal/render/boat.go` | TASK-043 |
| `internal/render/renderer.go` (highlight/ghost) | TASK-044 |
| `internal/ui/hud.go` | TASK-050 |
| `internal/ui/panels.go` (wind) | TASK-051 |
| `internal/ui/panels.go` (rudder) | TASK-052 |
| `internal/ui/hud.go` (tension hover) | TASK-053 |
| `main.go` | TASK-060, TASK-061 |
| `index.html`, `wasm_exec.js`, `Makefile` | TASK-002, TASK-062 |

---

## Tasks Without Tests

| Task | Has explicit test? | Notes |
|------|--------------------|-------|
| TASK-001 | No — build artifact | Verified by `go build` |
| TASK-002 | No — build artifact | Verified by `GOOS=js GOARCH=wasm go build` |
| TASK-003 | No — visual artifact | Manual smoke test |
| TASK-021 | No — constants only | Compilation verify only |
| TASK-030 | No — type definitions only | Compilation verify only |
| TASK-041..044 | No — render output | Manual visual verification |
| TASK-050..053 | No — render output | Manual visual verification |
| TASK-062 | No — deployment artifact | Manual browser test |
| All others | **Yes** — unit or integration test defined | ✓ |

All logically testable units have explicit test tasks. Pure build/render artifacts are acknowledged as visual-only.

---

## Cyclic Dependencies

Dependency graph checked manually:

```
physics (no deps)
  ↑
sim (depends on physics)
  ↑
input (depends on sim types)
  ↑
render (depends on physics for Vec2, sim for world state)
  ↑
ui (depends on sim for world state, input for Command types)
  ↑
main (depends on all above)
```

No cycles. Confirmed: `physics` ← `sim` ← `main`; `render` and `input` also point only toward `sim`/`physics`, not back from them.

**One potential issue to watch:** `internal/input/handler.go` imports `internal/sim` (for `CleatID`, `ThrottleState`, `BoatType` in payloads) and also `internal/render` (for `Viewport.ScreenToWorld` in mouse coordinate transform). This creates `input → render → physics` chain — not a cycle, but the `input` package must not import from `render` if `render` imports `input`. **Mitigation:** Move `Viewport.ScreenToWorld` so it's accessible to `input` without importing `render` — either inline the transform in `input.Handler` (it's 4 lines) or expose a standalone `WorldToScreen/ScreenToWorld` in `internal/physics` or a separate `internal/viewport` package. Recommended: **inline the viewport math in `input.Handler`**; the Handler struct holds a `Scale` and `OriginWorld` directly, avoiding the cross-dependency entirely. No new task needed — addressed in TASK-032 acceptance criteria which says Handler receives Viewport parameters.

---

## Security Requirements Without Tasks

From access-control.md:

| Requirement | Task |
|-------------|------|
| No unsafe pointer use | TASK-081 (grep check) |
| No CGO | TASK-081 (grep check) |
| No net/http in WASM binary | TASK-081 (build tag check) |
| HTTPS recommendation (deploy-time, not Go code) | TASK-062 (README note) |
| Input ranges validated before physics | TASK-034 (validation) + TASK-080 (audit) |

All security requirements from access-control.md have corresponding tasks. No gaps.

---

## Blocked Tasks

**No blocked tasks.** All design unknowns were resolved in Phase 2.

Previously blocked unknowns (now resolved):
- U1 (architecture) → WASM+Ebiten → unblocks all tasks
- U3 (physics fidelity) → tunable constants → TASK-021 unblocked
- U4 (catamaran) → passive MVP → TASK-022 unblocked
- U5 (dock shape) → straight pier → TASK-023 unblocked
- U6 (collision) → penalty spring → TASK-016/TASK-025 unblocked
- U7 (units) → SI → TASK-021 unblocked

---

## Risk Flags (Not Blockers)

| Risk | Mitigation | Task |
|------|-----------|------|
| SAT collision is the most complex single task (3-4h, custom algorithm) | Start with AABB (axis-aligned bounding box) as a v1; upgrade to OBB+SAT as v2 within TASK-016 | TASK-016 |
| `input → render` potential import cycle | Inline viewport math in `Handler` struct | TASK-032, TASK-040 |
| Ebiten on-canvas UI scope creep | MVP panels strictly defined; no new panels without a task | TASK-050..053 |
| Physics constants require tuning (feel vs accuracy) | Constants are in one file (`constants.go`); iterative tuning is TASK-021 + TASK-075 feedback loop | TASK-021, TASK-075 |
