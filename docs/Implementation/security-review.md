---
date: 2026-05-14
scope: physics subsystem added to src/main.ts
---

# Security Review — Physics Implementation

## PASSED

- **No credential leakage** — grep for `token|password|Bearer|api_key|secret|-----BEGIN` in `src/` returns zero matches. Physics code introduces no secrets.
- **No new network calls** — `physicsStep` reads and writes only module-level state; no `fetch`, `XMLHttpRequest`, or WebSocket calls added.
- **No new DOM access** — physics code does not touch the DOM beyond the existing `canvas` element.
- **No localStorage writes** — physics state (`vx`, `vy`, `omega`) is in-memory only; not persisted.
- **`dt` clamped to [0, 0.1]** — `const dt = Math.min((now - lastTime) / 1000, 0.1)` prevents spiral-of-death from tab-backgrounding spikes.
- **Division-by-zero guard** — `if (dist < 1e-6) continue` in mooring spring force prevents NaN from zero-length lines.
- **Input boundary enforcement** — `rudderAngle` and `throttlePort/Stbd` are validated at the slider event handlers (existing code, unchanged). `physicsStep` trusts these pre-validated values per the validation-boundary convention in AGENTS.md.
- **No `any` types** — `npx tsc --noEmit` passes with `strict: true`, `noUnusedLocals: true`, `noUnusedParameters: true`. No type-safety escapes introduced.
- **Terminal velocity self-limited** — drag constants ensure ~5 m/s at full ahead; no explicit clamp needed (verified analytically in access-control.md §Option A).

## WARNINGS

- **`MOORING_K / MOORING_C` oscillation risk** — spring constant 50000 N/m with damper 10000 N·s/m may produce underdamped oscillation for heavy boats at large extensions. TASK-022 tuning will verify; increase `MOORING_C` if visible jitter occurs.
- **Single-frame position correction** — the 50% OBB position correction in collision penalty is applied without rebuilding the OBB for subsequent obstacles in the same frame. At normal speeds this is safe; at extreme velocities a boat could partially tunnel through a thin obstacle. The 0.1 s `dt` clamp makes this unreachable in practice.

## CRITICAL

None.

## SECURITY SCORE: 9/10
