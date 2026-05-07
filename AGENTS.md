# AGENTS.md — Mooring Simulator

Agent and contributor reference for this repository.

---

## Technology Stack

| Layer | Technology | Version |
|-------|-----------|---------|
| Language | Go | 1.22+ |
| Game engine / rendering | Ebiten v2 | `github.com/hajimehoshi/ebiten/v2` |
| Build target | WebAssembly (`GOOS=js GOARCH=wasm`) | Go stdlib WASM |
| Physics | Custom pure-Go (no CGO) | — |
| WASM runtime shim | `wasm_exec.js` from Go SDK | matches Go version |

**No CGO. No external physics library. No server.** The binary is a self-contained static WASM file.

---

## Build Commands

```bash
# Native build (for testing, not the final target)
go build ./...

# WASM production build
GOOS=js GOARCH=wasm go build -o main.wasm .

# Serve locally (after WASM build)
make serve
# or manually:
go run tools/serve.go   # http://localhost:8080
```

---

## Test Commands

```bash
# All unit + integration tests (native — no browser needed)
go test ./internal/...

# With race detector (required before merge)
go test -race ./internal/physics/... ./internal/sim/... ./internal/input/...

# With coverage report
go test -coverprofile=coverage.out ./internal/physics/... ./internal/sim/...
go tool cover -func=coverage.out

# Specific integration scenario
go test -run TestWorldStep ./internal/sim/...

# Smoke test (world init + 10 sim-seconds without panic)
go test -run TestWorldInit ./internal/sim/...
```

---

## Linting and Static Analysis

```bash
go vet ./...
staticcheck ./...          # install: go install honnef.co/go/tools/cmd/staticcheck@latest
govulncheck ./...          # install: go install golang.org/x/vuln/cmd/govulncheck@latest
```

All three must pass with zero findings before any task is marked done.

---

## Code Conventions

### Validation boundary
All user input (keyboard, mouse) is validated and range-clamped in `internal/input` before being passed as typed `Command` values to `internal/sim`. The `sim` package **trusts** incoming `Command` values as already valid. Do not add redundant validation inside `sim.ApplyCommand`.

### Error handling
- `sim.AddMooringLine` returns `error` for invalid boat ID or cleat. Caller (main.go) logs and discards the command.
- All other mutation methods are panic-free (invalid indices are no-ops).
- No error wrapping libraries — stdlib `errors.New` / `fmt.Errorf` only.

### Package dependency rule (hard constraint)
```
physics ← sim ← main → render, input, ui
```
`internal/physics` and `internal/sim` must **never** import Ebiten. Verify with:
```
go list -f '{{.Imports}}' ./internal/physics/... | grep ebiten  # must be empty
go list -f '{{.Imports}}' ./internal/sim/...    | grep ebiten  # must be empty
```

### No unsafe, no CGO
```bash
grep -r "unsafe" --include="*.go" .    # must return empty
grep -r '"C"' --include="*.go" .       # must return empty
```

### Coordinate system
- Physics world: Y-up, meters, CCW angles
- Screen space: Y-down, pixels
- Y-flip happens **only** in `internal/render/viewport.go` `WorldToScreen`. Nowhere else.

### Constants
All tunable physics constants (thrust, drag, spring stiffness, etc.) live in `internal/sim/constants.go`. Do not hard-code numeric values in force functions or `World.Step`.

### Commit conventions
```
<type>(<scope>): <short description>

type: feat | fix | test | refactor | docs | build
scope: physics | sim | render | input | ui | main | docs
```
Example: `feat(physics): add SAT collision penalty`

---

## Coverage Thresholds

| Package | Minimum |
|---------|---------|
| `internal/physics` | 90% |
| `internal/sim` | 80% |
| `internal/input` | 80% |
| `internal/render` | 60% (viewport only) |

---

## MVP Scope Boundary

Do not implement the following without a new task:
- Catamaran engine / prop walk (post-MVP)
- User-configurable dock shape (post-MVP)
- Inextensible mooring lines (post-MVP)
- User-editable boat parameters (post-MVP)

See docs/Design/architecture.md "MVP Scope Boundary" for full list.
