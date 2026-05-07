# Task Brief — Mooring Simulator

## 1. Task Type

**Greenfield application.** No existing codebase, build system, or tooling. Starting from zero.

Sub-classification: browser-based 2D real-time physics game / simulator.

## 2. Input Data

| Source | Location | What it provides |
|--------|----------|-----------------|
| Problem statement | `initial-problem.md` | Full feature spec, stack choice (Go) |
| CLAUDE.md | `CLAUDE.md` | Project status (empty repo), stack TBD beyond "Go-based" |
| README | `README.md` | Title only |

No API specs, schemas, existing code, or additional documentation exist.

## 3. Facts Extracted from Problem Statement

**Rendering:**
- Browser-based
- Top-down 2D view
- Schematic (simple geometric shapes — not realistic graphics)
- Elements to render: boats (hull outline + CoM marker), dock, water background, mooring lines

**Entities:**
- One or more boats (monohull sailboat + catamaran)
- A dock (fixed in space)
- Mooring lines / springs (dock point → boat cleat)

**Boat controls (per active/selected boat):**
- Rudder position (angle)
- Throttle: Neutral | Slow Forward | Full Forward | Slow Astern | Full Astern
- Prop walk parameter (tunable per boat)

**Mooring line placement:**
- Dock end: any point on dock
- Boat end: only at cleats — stern, midships, bow

**Physics forces to simulate:**
1. Wind (strength + direction) — global parameter
2. Engine thrust — function of throttle state
3. Prop walk — lateral force, tunable, stronger in reverse
4. Rudder — lateral force + torque, requires boat speed
5. Mooring line tension — spring/elastic, applied at cleat
6. Center of mass — determines torque arm for all forces
7. Contact/collision with other boats and with dock

**Multi-boat:**
- Multiple boats can be in the water simultaneously
- One boat is "active" (user-controlled) at a time

## 4. Phase 1 Questions

### Go + Browser Architecture
1. Pure client-side (Go → WASM, no server) or Go server + thin JS frontend?
2. Use an existing Go 2D game library (e.g., Ebiten which compiles to WASM), or raw WASM + Canvas API?

### Physics Fidelity
3. Target accuracy level — nautical-grade (real hydrodynamic coefficients) or tunable/gameplay feel?
4. Real-time continuous simulation or user can pause/step?

### Boats
5. How physically distinct should monohull vs catamaran be? (Catamaran typically has twin engines → two independent prop walks, wider hull, different CoM height)
6. Are boat dimensions/mass fixed defaults, or user-configurable per session?

### Dock
7. Dock shape — single straight pier, L-shape, T-shape, or multiple configurations?
8. Is the dock a hard collision boundary, or can boats pass through it (i.e., is collision detection required)?

### Mooring Lines
9. Are mooring lines always elastic (spring), or can they be inextensible (taut line, zero extension)?
10. Maximum number of mooring lines per session?

### Interaction / UI
11. How is a boat selected as "active"? Click on it?
12. How are mooring lines placed? Click dock point then click cleat?
13. Is there a scenario/level system, or is it pure sandbox?
14. Should wind affect monohull and catamaran differently (different sail/windage area)?

### Collisions
15. Collision response type: penalty spring (soft), impulse (hard), or simply prevent penetration?
