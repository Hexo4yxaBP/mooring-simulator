# Plan: Mooring Simulator for Sailboats & Catamarans

## Progress Log

### Phase 1 ✅ COMPLETE
- ✅ Initialized Vite project with TypeScript, Rapier physics, Canvas rendering
- ✅ Created core game loop with fixed 60 Hz timestep physics + variable framerate rendering
- ✅ Built Canvas 2D renderer with boat/dock/water/mooring line visualization
- ✅ Created central state manager (Scene) for boats, mooring lines, wind
- ✅ Built functional UI with wind controls, boat selection, throttle/rudder controls
- ✅ Dev server running on http://localhost:5173/

### Phase 2 ✅ COMPLETE
- ✅ Defined sailboat and catamaran configs (mass, drag, wind profile, cleat positions)
- ✅ Created Rapier rigid body creation with compound colliders (hull + keel)
- ✅ Implemented buoyancy simulation
- ✅ Physics synchronization (Rapier ↔ Scene state)

### Phase 3 ✅ COMPLETE
- ✅ Wind force calculation (applies at center of wind pressure above COM)
- ✅ Rudder deflection force (velocity-squared model, -30° to +30°)
- ✅ Engine thrust (5 throttle states with direction-based forces)
- ✅ Prop walk asymmetric thrust (user-adjustable left/right direction)
- ✅ Water drag (linear + quadratic damping)
- ✅ All forces integrated into physics update loop

### Phase 4 🚀 IN PROGRESS (Mooring System)
- ✅ Mooring constraint creation (Rapier distance constraints with slack)
- ✅ Interactive mooring placement UI (click boat cleat → click dock)
- ✅ Mooring line display (taut vs slack visual feedback)
- ⏳ Need to test mooring constraints in physics loop

---

## TL;DR
Build a browser-based, mobile-optimized 2D mooring simulator supporting 1-3 boats with realistic physics. Architecture: Canvas 2D rendering + Rapier physics engine (for rigid bodies + constraints) + custom marine forces (wind, rudder, prop walk, buoyancy). Vanilla JS or lightweight framework (Vue.js recommended for simplicity). Fixed 60 Hz physics timestep, variable framerate rendering. MVP focuses on core simulation; only wind strength/direction exposed initially. Use browser localStorage for scene persistence.

**Key decision:** Hybrid physics approach — Rapier handles rigid body dynamics and mooring constraints, custom code applies marine-specific forces (wind, rudder, prop walk).

---

## Steps (Grouped by Phase)

### Phase 1: Project Setup & Infrastructure
1. **Initialize project structure**
   - Set up Vite or Webpack for bundling (for mobile scalability)
   - Install Rapier WASM physics engine (`@react-three/rapier` or direct Rapier)
   - Add TypeScript (optional but recommended for maintainability)
   - Configure Canvas 2D + Rapier integration layer
   - Set up mobile-responsive viewport (100vw, 100vh, no zoom)

2. **Create core game loop architecture**
   - Implement fixed-timestep physics loop (1/60 sec increments with accumulator pattern)
   - Implement variable framerate rendering loop (synced to physics)
   - Create central state manager for boats, wind, mooring lines, UI state
   - *Note: Use vanilla JS or Vue.js for light overhead*

3. **Set up Canvas 2D rendering pipeline**
   - Create `Renderer` class with methods: drawBoat(), drawDock(), drawWater(), drawMooringLines()
   - Implement camera/transform system for panning/zooming (manual 2D transforms)
   - Establish coordinate system: origin at water surface, +X = starboard, +Y = forward

### Phase 2: Boat Physics Foundation
4. **Define boat types (Sailboat vs Catamaran)**
   - **Sailboat model:**
     - Hull: compound shape (lower hull box + keel box + superstructure)
     - Mass: ~5000 kg, beam 2.5 m, LOA 10 m
     - Cleat positions: bow (+5m), midships (0m), stern (-4m) relative to center
     - Wind profile area: ~20 m² (height × beam estimate)
     - Drag coefficients: linear damping 0.3, angular damping 0.35
   
   - **Catamaran model:**
     - Hull: twin hulls (two parallel boxes)
     - Mass: ~7000 kg, beam 4 m (each hull 1.2m wide), LOA 11 m
     - Cleat positions: same relative locations
     - Wind profile area: ~18 m² (lower due to wider, lower profile)
     - Drag coefficients: linear damping 0.25, angular damping 0.25 (less rotational resistance)

5. **Implement boat body creation in Rapier**
   - Create `BoatPhysicsBody` class
   - Use `RigidBodyDesc.dynamic()` with compound shapes (lower hull + keel + superstructure)
   - Set `mass_properties` based on boat type
   - Configure collision groups: boats can collide with dock & each other, NOT with mooring lines
   - Store metadata: boat type, cleat positions, center of mass offset

6. **Add water/buoyancy simulation**
   - Water level constant: y = 0
   - In each physics step: calculate submerged volume per boat (simplified: box intersection with waterline)
   - Apply upward buoyancy force: `displacement * water_density(1000 kg/m³) * gravity(9.81)`
   - Apply linear velocity-squared damping: `drag = 0.5 * water_density * drag_coeff * velocity²`
   - Result: boats naturally float and resist motion

### Phase 3: Marine Forces
7. **Implement wind force application**
   - Store wind: direction (0-360°) + magnitude (0-20 m/s)
   - For each boat: calculate force at "center of wind pressure" (estimated height above center of mass)
   - Wind force vector: `magnitude² * boat_profile_area * drag_coeff(0.7) * wind_direction_unit`
   - Apply at offset position (higher than center of mass) → creates heeling/yaw moment
   - UI: Sliders for wind magnitude + direction (MVP only)

8. **Implement rudder force & deflection**
   - Store boat's `rudder_angle` (-30° to +30°)
   - Rudder force (simplified, no lookup tables):
     - `rudder_force = rudder_angle_rad * boat_forward_velocity² * rudder_area(0.5m²) * lift_coeff(0.8)`
     - Direction: perpendicular to forward direction (causes side force)
   - Torque: `rudder_force * distance_aft_from_center_of_mass(~3 m)`
   - Input: UI slider for active boat's rudder angle (±30°)

9. **Implement prop walk effect**
   - Store boat's `throttle_state` (neutral, slow fwd, full fwd, slow astern, full astern)
   - Store boat's `prop_walk_direction` (left/right, user-adjustable from UI)
   - Map throttle to `thrust_magnitude` and `shaft_rpm_direction`
   - Prop walk angle: ±12° off thrust direction (determined by prop_walk_direction setting)
   - Side force from prop slip: `thrust * sin(prop_walk_angle) * prop_walk_magnitude(0.05-0.15)`
   - Apply force offset aft of center (~2-3m) → combined with rudder creates complex turning behavior
   - Input: UI buttons/slider for throttle state + UI toggle/selector for prop walk direction

10. **Implement engine thrust**
    - Map throttle states to forward force:
      - Neutral: 0 N
      - Slow forward: ~1000 N
      - Full forward: ~3000 N
      - Slow astern: ~800 N
      - Full astern: ~2500 N
    - Direction: forward along boat's heading (using boat's rotation matrix)
    - Duration: instant when throttle changes, persist until changed

### Phase 4: Mooring System
11. **Create mooring line constraints**
    - `MooringLine` class: dock_cleat_position → boat_cleat_position
    - Use Rapier's `DistanceConstraint`:
      - `min_length = 0` (allows slack)
      - `max_length = line_length_meters`
      - `stiffness = 0.1` (low, to prevent oscillation)
      - `damping = 0.5` (moderate, to absorb energy)
    - Only apply constraint when line is taut (distance >= max_length)
    - Storage: array of mooring lines per boat

12. **Implement interactive mooring placement**
    - UI: Click on boat cleat → select it (highlight)
    - UI: Click on dock → place mooring line from selected cleat to dock point
    - Display: Draw mooring lines as lines from dock to boat
    - Visual feedback: different colors for taut vs slack lines
    - Delete: Right-click mooring line to remove
    - Limit: max 3 mooring lines per boat (practical constraint)

### Phase 5: Visualization & Rendering
13. **Implement boat rendering**
    - Draw boat as polygon (simple hull silhouette) from top-down view
    - Include:
      - Hull outline (filled polygon, different colors per type)
      - Center of mass indicator (small dot or cross, offset from geometric center if applicable)
      - Bow indicator (arrow or different shade at bow)
      - Rudder indicator (small perpendicular line at stern, rotates with rudder angle)
    - Use boat's rotation angle from Rapier body

14. **Implement dock rendering**
    - Static dock structure: rect or polygon
    - Mark cleat positions (small circles at standard locations)
    - Cleat labels: "Bow", "Midships", "Stern"

15. **Implement water rendering**
    - Simple: gradient or solid color background
    - Optional: ripple effect at boat's current position (low-cost canvas effect)
    - Waterline at y=0

16. **Implement mooring line rendering**
    - Draw lines from dock cleats to boat attachment points
    - Color coding: green (slack), red (taut, pulling)
    - Thickness variation for visual clarity
    - Annotate with line length on hover (optional)

17. **Create telemetry display**
    - Overlay text with:
      - Active boat name + type
      - Position (X, Y)
      - Heading angle (degrees)
      - Velocity magnitude + direction
      - Wind (current direction + magnitude)
      - Active boat's: rudder angle, throttle state, prop walk effect
    - Update every render frame

### Phase 6: UI Controls & Interaction
18. **Implement boat selection**
    - Click boat → select (highlight border color change)
    - Only one active boat at a time (color highlight)
    - Keyboard shortcut (1, 2, 3) to switch between boats
    - Display which boat is active in telemetry

19. **Implement control sliders/buttons**
    - Wind direction slider (0-360°) + magnitude slider (0-20 m/s)
    - For active boat:
      - Rudder angle slider (-30° to +30°, visual feedback on boat)
      - Throttle buttons (or slider) for 5 states
      - Prop walk direction toggle (left / right) + magnitude adjustment if applicable
    - Pause/Resume button (freezes physics)
    - Reset button (returns boats to initial positions)

20. **Implement boat placement UI**
    - Initial state: click on water to place new boat
    - Type selector: dropdown (Sailboat / Catamaran)
    - Limit to 3 boats total
    - Display boat count

21. **Implement mooring line UI**
    - Mode toggle: "Place mooring" / "Normal mode"
    - In place mode: click boat cleat → click dock location → line placed
    - Visual feedback: highlight valid cleats and dock areas
    - List mooring lines with delete option

### Phase 7: Save/Load System
22. **Implement localStorage persistence**
    - On scene change: auto-save boat positions, rotations, mooring lines, wind state
    - On load: restore previous scene
    - Add "Clear saved game" button in UI
    - Data format: JSON with boats[] array, mooring_lines[] array, wind state

### Phase 8: Polish & Testing
23. **Performance optimization**
    - Profile Canvas rendering (should be <2ms per frame for 1-3 boats)
    - Tune physics timestep substeps if needed (may need multiple steps for stability)
    - Test on mobile devices (iOS/Android browsers)
    - Ensure responsive layout adapts to landscape/portrait

24. **Physics tuning & validation**
    - Manual testing: verify boats float correctly and respond to forces
    - Test wind at various speeds/directions
    - Test rudder deflection: should cause rotation + side slip
    - Test prop walk: should be noticeable in tight turning maneuvers
    - Test mooring line behavior: slack/taut transitions should be smooth
    - Collision response: boats should not overlap dock or each other

25. **User experience polish**
    - Smooth transitions between UI modes
    - Clear visual feedback for all interactions
    - Prevent invalid operations (e.g., placing mooring at boat cleat that's already occupied)
    - Zoom/pan controls for viewing full scene
    - Tooltips for all controls

---

## Relevant Files
(To be created during implementation)
- `src/main.ts` — Entry point, game loop initialization
- `src/physics/boat.ts` — Boat body creation, properties per type
- `src/physics/forces.ts` — Wind, rudder, prop walk, buoyancy calculations
- `src/physics/mooring.ts` — Mooring constraint management
- `src/physics/world.ts` — Rapier world setup, fixed timestep loop
- `src/rendering/renderer.ts` — Canvas 2D drawing (boats, dock, water, mooring lines, telemetry)
- `src/rendering/camera.ts` — Pan/zoom/transform handling
- `src/ui/controls.ts` — Slider/button input handling
- `src/ui/boat-placement.ts` — Boat creation UI
- `src/ui/mooring-placement.ts` — Mooring line placement UI
- `src/state/scene.ts` — Central state manager (boats, wind, mooring lines)
- `src/persistence/storage.ts` — localStorage save/load
- `index.html` — Minimal HTML with canvas element
- `package.json` — Dependencies (Rapier, build tooling)

---

## Verification
1. **Physics accuracy** — Create a boat, apply wind, verify it drifts perpendicular to wind direction + rotates
2. **Rudder control** — Apply rudder angle at various boat speeds, verify yaw rate increases with speed
3. **Mooring constraints** — Place 2 mooring lines on same boat, apply wind, verify boat settles to equilibrium without oscillating
4. **Collision response** — Push boat toward dock, verify it stops (no overlap) and can be pulled back by mooring tension
5. **Mobile responsiveness** — Test on tablet in portrait/landscape, verify UI elements are accessible and simulation remains smooth (60+ fps)
6. **Save/load** — Place boats with mooring lines, refresh page, verify scene is restored exactly
7. **Wind direction** — Apply wind from each cardinal direction, verify force direction matches UI input
8. **Throttle response** — Switch throttle states, verify forward/astern thrust direction matches boat heading

---

## Decisions & Scope
- **Decision 1:** Canvas 2D over WebGL — Simpler for MVP, sufficient for 1-3 boats, mobile-friendly
- **Decision 2:** Rapier physics engine — Superior rope/constraint support compared to Cannon.js, WebAssembly performance
- **Decision 3:** Hybrid physics — Rapier handles rigid bodies; custom code handles marine-specific forces (wind, rudder, prop walk)
- **Decision 4:** MVP scope — Only wind strength/direction exposed in UI initially; boat mass, drag coefficients, mooring stiffness are code constants
- **Decision 5:** No 3D rotation — Keep boats flat on water (no heel angle calculation); rely on damping for stability
- **Decision 6:** Simplified rudder model — Linear force approximation instead of lift coefficient lookup tables (can be added later)
- **Decision 7:** Catamaran physics — Use slightly modified sailboat model (lower damping for less rotational resistance); distinct visual appearance

**Deliberately excluded from MVP:**
- Sails & aerodynamic coefficients (simplifies wind model)
- Water current/tide effects
- Realistic hull hydrodynamic tables (ITTC coefficients)
- Metacenter/heeling physics (rely on damping)
- Multiple collision layers/groups beyond boat/dock
- Undo/redo for mooring placement
- Multiplayer or scenario mode
- Advanced graphics (water surface animation, boat wake)

**Future enhancements (if time permits):**
1. Add water current UI slider
2. Add boat mass/drag coefficient tuning (debug mode)
3. Add scenario builder (named setups, objectives)
4. Add realistic wind gust simulation
5. Migrate rendering to Three.js with WebGL for dynamic water effects
6. Network multiplayer via WebSocket

---

## Further Considerations

**1. Physics engine library choice — should we also evaluate lightweight alternatives?**
   - **Recommendation:** Stick with Rapier. It's the only major WASM physics engine with good rope constraint support. Cannon.js is older and doesn't handle this as elegantly. Building custom physics from scratch would take 2-3x longer.

**2. Framework choice — vanilla JS vs. Vue.js vs. React?**
   - **Recommendation:** **Vanilla JS with Lit (lightweight web components)** OR **Vue.js 3**.
     - Vanilla JS: Fastest, most control, smallest bundle (good for mobile)
     - Vue.js: Better component organization, less boilerplate, slight bundle size overhead (~50KB gzipped)
     - React: Overkill for this (100KB+ gzipped), slower initial load on mobile
     - **Suggest:** Start with Vanilla JS + Lit for minimal overhead; migrate to Vue if UI complexity grows.

**3. How detailed should the boat geometry be?**
   - **Recommendation:** Keep it simple for MVP. 5-point polygon hull (bow, stern, port/starboard) is enough. Don't create detailed 3D meshes.
   - If later you want accurate ballast distribution → parameterize the compound shape's center of mass offset.

**4. Prop walk implementation — should we simulate actual shaft torque or simplified side force?**
   - **Recommendation:** Use simplified side force for MVP (Step 9). It's 80% of the effect with 20% of the complexity. Shaft torque physics can be added later if needed.

**5. Should collision detection between mooring lines and dock be enabled?**
   - **Recommendation:** NO. Keep mooring lines as constraints only (no collision shapes). Prevents performance issues and complex interactions.
