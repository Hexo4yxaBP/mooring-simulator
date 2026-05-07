# Technology Options — Go + Browser

*Source: `initial-problem.md` (stack: "Golang based"), standard Go ecosystem knowledge.*

---

## Constraint

The application must run in a browser. The stack is Go-based. This means Go must be the implementation language for at least the core simulation logic.

---

## Option A — Ebiten (Go 2D Game Library → WASM)

**What it is:** [Ebiten](https://ebitengine.org/) is a mature Go 2D game engine. It compiles to WebAssembly via `GOOS=js GOARCH=wasm`, running entirely client-side.

**Provides:**
- Fixed-timestep game loop (`Update` + `Draw` callbacks)
- 2D canvas rendering (sprites, shapes, text via `ebitenutil`)
- Input handling (keyboard, mouse, touch)
- Audio (optional)
- No server required

**Pros:**
- All logic in pure Go — single language
- No server infrastructure
- Game loop architecture fits a real-time physics simulator exactly
- Active community, good WASM support
- `ebiten/v2/vector` package for drawing lines/polygons (mooring lines, hulls)

**Cons:**
- WASM binary ~5–15 MB (acceptable for a dev/training tool)
- HTML UI overlays (forms, sliders) require JS interop or ebitenui library
- Less ergonomic for form-heavy parameter panels than HTML

**Build & run:**
```bash
GOOS=js GOARCH=wasm go build -o main.wasm .
# Serve with any static file server
```

---

## Option B — Raw Go WASM + Canvas API

**What it is:** Go compiled to WASM using `syscall/js`, calling the browser Canvas 2D API directly from Go.

**Provides:**
- Full control over rendering via JS Canvas API calls
- Can use HTML/CSS for UI panels freely

**Pros:**
- Lightweight (only standard library)
- HTML/CSS UI with no constraints

**Cons:**
- No game loop, input, or rendering helpers — must implement manually
- `syscall/js` calls have overhead vs native Go
- More boilerplate for the same result as Ebiten

---

## Option C — Go Server (WebSocket) + JavaScript Canvas

**What it is:** Go runs the physics simulation server-side. Browser connects via WebSocket; receives world state at each tick; renders with JS Canvas API.

**Provides:**
- Physics in pure Go (no WASM)
- Full HTML/CSS UI

**Pros:**
- Smaller client payload (just JS + small HTML)
- Go server can use full stdlib without WASM constraints
- Easier multiplayer / shared simulation later

**Cons:**
- Requires running a server (not a standalone static file)
- Network latency adds jitter to simulation feel
- Must build a JS rendering layer separately
- Two languages to maintain (Go + JS)

---

## Option D — Go Backend REST/HTTP + JS Frontend (SPA)

**What it is:** Go serves HTTP; browser polls or uses SSE for state; JS renders.

**Assessment:** Poor fit for real-time simulation — polling latency and CPU overhead.

---

## Recommendation (for Phase 2 decision)

| Criterion | Ebiten/WASM | Raw WASM | Go+WS+JS |
|-----------|------------|----------|----------|
| Single language (Go only) | ✓ | ✓ | Partial |
| No server required | ✓ | ✓ | ✗ |
| Game loop built-in | ✓ | ✗ | ✗ |
| HTML UI for params | Limited | ✓ | ✓ |
| Physics in Go | ✓ | ✓ | ✓ |
| Best fit for "game-like" sim | **✓** | — | — |

**Ebiten → WASM is the natural fit** for this problem given the Go constraint and real-time simulation requirement. [UNKNOWN — user confirmation needed]

---

## Go Modules / Packages Relevant to Physics

| Package | Purpose |
|---------|---------|
| `math` | Trigonometry for force rotation, angle math |
| `golang.org/x/exp` | Potentially useful numeric helpers |
| Custom `physics` package | Rigid body sim, force accumulation, integrator |
| Custom `world` package | Entity management |
| Custom `render` package | Ebiten draw calls |
| Custom `input` package | Mouse/keyboard event mapping |

---

## Physics Library Options

No mature Go rigid-body 2D physics library exists at the level of Box2D. Options:

1. **Custom implementation** — Write the 2D rigid-body integrator. Feasible given the limited entity count (few boats + dock) and 2D constraint. Likely the right call.
2. **Box2D Go bindings** — CGO bindings exist (e.g., `github.com/ByteArena/box2d`) but CGO does not compile to WASM.
3. **Chipmunk Go bindings** — Same CGO limitation.

**Conclusion:** Custom physics engine (no CGO dependency) is required for WASM target. Scope is manageable: ~5 entity types, 2D, ~7 force types.
