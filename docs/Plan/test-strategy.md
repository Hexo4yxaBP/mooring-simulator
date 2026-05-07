# Test Strategy — Mooring Simulator

*Sources: architecture.md (D9: physics/sim must not import Ebiten, enabling native testing), interface-spec.md (contracts per package), implementation-tasks.md (TASK-070..082).*

---

## Testing Philosophy

The app separates physics and simulation logic from rendering (architecture D9). This means `internal/physics` and `internal/sim` can be fully tested with plain `go test` on any platform — no browser, no WASM, no display required. Only `internal/render` and `internal/ui` require Ebiten, and those are tested by visual inspection only.

---

## Test Pyramid

```
              [Manual/Visual]
           TASK-041..044, 050..053
          (render + UI — Ebiten only)

        [Integration Tests]
      TASK-075 (World.Step scenarios)
      TASK-061 (world init smoke test)

      [Unit Tests]
  physics: TASK-070..073
  sim:     TASK-074
  input:   TASK-076..077
```

---

## Unit Tests

### Packages covered
- `internal/physics` — Vec2, RigidBody, ForceAccumulator, Integrate, all force functions, SpringForce, CollisionPenalty
- `internal/sim` — enums, Boat helpers, Dock, MooringLine
- `internal/input` — Handler state machine, all key/mouse mappings, validation
- `internal/render` — Viewport transforms only (no draw calls)

### Coverage targets

| Package | Minimum coverage |
|---------|-----------------|
| `internal/physics` | ≥ 90% |
| `internal/sim` | ≥ 80% |
| `internal/input` | ≥ 80% |
| `internal/render` | ≥ 60% (viewport only; draw functions excluded) |

### Test style
- Table-driven tests (`[]struct{ name, input, expected }`) for all function with multiple cases
- One test file per source file: `vec2.go` → `vec2_test.go`
- No mocks needed (all tested units are pure functions or value types with no external dependencies)

---

## Integration Tests

### World.Step scenarios (TASK-075)
File: `internal/sim/world_test.go`

These tests run the full simulation loop (physics + sim, no render) for N steps and assert on macro outcomes. They validate that the force model produces physically plausible behaviour without requiring tuned constants to be exact.

| Test name | Steps | Assert |
|-----------|-------|--------|
| `TestBoatCoastsToStop` | 600 (10s) | Final speed < 0.1 m/s |
| `TestMooringHoldsBoat` | 1800 (30s) | Lateral displacement < 2 m |
| `TestThrottleForwardAccelerates` | 300 (5s) | Speed > 1 m/s |
| `TestPropWalkInAstern` | 180 (3s) | Lateral displacement > 0 (port) |
| `TestCollisionPushesBoatAway` | 6 (0.1s) | Boat outside dock polygon |

### Smoke test (TASK-061)
Run 600 steps of the default world with no input. Assert: no panic, no NaN in any boat position or velocity.

---

## Race Detector

All non-render tests must pass with `-race`:
```
go test -race ./internal/physics/... ./internal/sim/... ./internal/input/...
```

The game loop runs on a single goroutine (Ebiten model), so there is no concurrency in production code. The race detector run validates no accidental goroutine is spawned.

---

## WASM Build Verification

WASM compilation is not a unit test but is a required CI check:
```
GOOS=js GOARCH=wasm go build -o main.wasm .
```
This must pass after every change to catch WASM-incompatible imports (CGO, net, os/file).

---

## Test Commands

```bash
# Run all native unit + integration tests
go test ./internal/...

# With race detector
go test -race ./internal/physics/... ./internal/sim/... ./internal/input/...

# With coverage
go test -coverprofile=coverage.out ./internal/physics/... ./internal/sim/...
go tool cover -html=coverage.out

# Specific integration test
go test -run TestWorldStep ./internal/sim/...

# Static analysis
go vet ./...
staticcheck ./...

# Vulnerability scan
govulncheck ./...

# WASM build check
GOOS=js GOARCH=wasm go build -o /dev/null .   # Unix
# Windows: GOOS=js GOARCH=wasm go build -o NUL .
```

---

## What Is NOT Tested Automatically

| Component | Reason | How verified |
|-----------|--------|-------------|
| `internal/render` draw calls | Requires Ebiten display context | Manual visual check |
| `internal/ui` panels | Requires Ebiten display context | Manual visual check |
| WASM browser execution | Requires browser + JS runtime | Manual: open `http://localhost:8080` |
| Mooring line colour thresholds | Colour lookup helper is testable; pixel output is not | Unit test on colour-lookup function only |

---

## CI Environment

No CI server is configured yet. When added, the following must all pass on push:

1. `go build ./...` (native)
2. `GOOS=js GOARCH=wasm go build -o main.wasm .`
3. `go test -race ./internal/...`
4. `go vet ./...`
5. `govulncheck ./...` (non-blocking on informational findings, blocking on high severity)

No sandbox, secrets, or environment variables required — the entire application is deterministic pure-Go with no external dependencies at runtime.

---

## Coverage Enforcement

Recommended: add to `Makefile`:
```makefile
test-coverage:
	go test -coverprofile=coverage.out ./internal/physics/... ./internal/sim/...
	go tool cover -func=coverage.out | grep -E "total|physics|sim"
	@go tool cover -func=coverage.out | tail -1 | awk '{if ($$3+0 < 80) {print "Coverage below 80%"; exit 1}}'
```
