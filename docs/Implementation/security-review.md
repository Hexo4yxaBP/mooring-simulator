# Security Review

_Reviewed after Phase 4 implementation. This is a client-only WASM game with no network API, no authentication, no persistent storage, and no user-supplied data beyond keyboard/mouse input._

---

## PASSED

- **No credential leakage**: `grep -rn "token|password|Bearer|api_key|secret|-----BEGIN" --include="*.go" .` — zero matches in production code.
- **No unsafe pointer use**: `grep -rn '"unsafe"' --include="*.go" .` — zero matches. Pure-Go physics with no `unsafe` package.
- **No CGO**: `grep -rn '"C"' --include="*.go" .` — zero matches. CGO-free; required for WASM target.
- **No network code**: Application is a pure client-side WASM game. No HTTP handlers, no sockets, no external requests.
- **No PII in logs**: The only output is `log.Fatal` on Ebiten startup failure (no user data).
- **Input validation at boundary**: All user input passes through `internal/input.Handler.Poll()` before reaching the sim. Rudder angle, wind speed, and throttle state are clamped/validated before command emission. The `applyCommand` dispatcher in `main.go` performs typed assertion from validated handler output only.
- **No integer overflow risk in line removal**: Line index is bounded by `len(world.Lines)`, checked in `RemoveMooringLine` before use.
- **Wind direction NaN safety**: `WindPayload.Direction` is normalised to `[0, 2π)` by `normaliseAngle()` before use. `math.Cos/Sin` on any finite float64 is always finite.
- **No SQL/XSS/injection surface**: Application has no database, no HTML template rendering, no user-generated string output.
- **Dependencies**: Only `github.com/hajimehoshi/ebiten/v2` — a well-maintained game library. No web framework or crypto library dependency.

## WARNINGS

- **Render coverage 21.1%**: The Ebiten drawing functions (`drawBoat`, `drawDock`, etc.) cannot be unit-tested without a GPU context. The pure-math portions (Viewport, lineTensionColor, hull vertex math) are 100% covered. This is a structural limitation, not a security risk.
- **No WASM content-security-policy**: `index.html` does not set CSP headers. Since the app is served by a development-only `tools/serve.go`, this is acceptable for local use. A production deployment should add `Content-Security-Policy: default-src 'self'`.

## CRITICAL

_(none)_

---

## SECURITY SCORE: 9/10

Deduction: -1 for missing CSP headers in `index.html` (relevant only if deployed publicly, not for local dev use).
