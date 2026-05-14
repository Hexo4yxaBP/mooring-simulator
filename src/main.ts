const canvas = document.getElementById('canvas') as HTMLCanvasElement;
const ctx = canvas.getContext('2d')!;

const SCALE = 20; // pixels per world-meter

const BOAT_SIZE = {
  monohull:  { w: 3.0, h: 10.0 },
  catamaran: { w: 6.5, h: 11.0 },
} as const;

type BoatType = keyof typeof BOAT_SIZE;

interface Boat {
  type: BoatType;
  x: number;
  y: number;
  heading: number;     // radians, 0 = bow up (north)
  rudderAngle: number; // radians, +ve = starboard, range ±35°
  throttlePort: number; // 0=full astern … 2=neutral … 4=full ahead
  throttleStbd: number; // catamaran only; mirrors throttlePort for monohull (unused)
  vx?: number;    // world-space m/s; initialised to 0 by physicsStep
  vy?: number;    // world-space m/s
  omega?: number; // angular velocity rad/s
}

interface RevertAnim {
  idx: number;
  fromX: number; fromY: number; fromHeading: number;
  toX:   number; toY:   number; toHeading:   number;
  startTime: number;
  duration:  number;
}

interface OBB {
  x: number; y: number;
  w: number; h: number;
  heading: number;
}

interface StaticCleat { wx: number; wy: number; kind: 'pier' | 'buoy'; }
interface MooringLine  { from: CleatRef; to: CleatRef; naturalLength: number; }
type CleatRef =
  | { kind: 'static'; idx: number }
  | { kind: 'moving'; boat: number; cleat: number };

const boats: Boat[] = [];
let activeBoatIdx: number | null = null;
let dragBoatIdx: number | null = null;
let dragOffsetX = 0;
let dragOffsetY = 0;
let isRotating = false;
let savedX = 0, savedY = 0, savedHeading = 0;
let revertAnim: RevertAnim | null = null;

let staticCleats: StaticCleat[] = [];
const mooringLines: MooringLine[] = [];
let pendingCleat: CleatRef | null = null;
let cursorX = 0, cursorY = 0;

let lastTime = 0;

let windAngle = Math.PI * 1.25; // radians, 0 = blowing toward north; default SW
let windKt    = 10;
let isDraggingWind = false;

type GameMode = 'setup' | 'play';
let gameMode: GameMode = 'setup';

const WIND_CX = 70; // canvas px from left edge of canvas
const WIND_CY = 70; // canvas px from top edge of canvas
const WIND_R  = 40; // arrow length in px

const PIER_DEPTH_M = 5;   // pier depth in world-meters
const PLANK_W_M    = 0.5; // each plank width in world-meters (≈ 1:10 aspect at 5m deep)

const CLEAT_R      = 5;          // cleat circle radius in pixels
const CLEAT_SW     = 1.5;        // cleat stroke width in pixels
const CLEAT_FILL   = '#FFC90E';
const MOVING_CLEAT = '#22B14C';  // boat cleats
const STATIC_CLEAT = '#ED1C24';  // pier / buoy cleats
const BUOY_DIST_M  = 15;         // 1.5 × monohull 10 m
const RUDDER_LEN   = 1.0 * SCALE; // visual rudder blade length in pixels

// Physics constants
const BOAT_MASS: Record<BoatType, number> = { monohull: 7000, catamaran: 12000 };
// Moment of inertia — rectangular approximation (1/12)×m×(w²+h²)
const BOAT_I: Record<BoatType, number>    = { monohull: 63600, catamaran: 163000 };
// Thrust per throttle index (0=full astern … 4=full ahead), Newtons
const THROTTLE_FORCE = [-10000, -4000, 0, 5000, 15000];
// Starboard-positive lateral propwalk force at stern, Newtons (right-handed screw)
const PROP_WALK_TABLE = [-1500, -500, 0, 250, 500];
const C_DRAG_FWD  = 3000;   // N/(m/s) longitudinal drag
const C_DRAG_LAT  = 80000;  // N/(m/s) lateral drag (~27× fwd — keel effect)
const C_DRAG_ROT  = 50000;  // N·m/(rad/s) rotational drag
const PROP_AFT_M   = 1.5;   // propeller m aft of boat centre (near keel exit)
const RUDDER_AFT_M = 4.0;   // rudder post m aft of boat centre (~85% from bow on 10 m hull)
const C_RUDDER    = 1400;   // N per (rad × m/s) — r≈15 m (1.5L) at full rudder
const WIND_K_FWD  = 5;      // N/(m/s)² bow-on wind coefficient
const WIND_K_LAT  = 27;     // N/(m/s)² beam-on wind coefficient
const MOORING_K   = 50000;  // N/m spring constant
const MOORING_C   = 10000;  // N·s/m damping coefficient
const COLLISION_K = 150000; // N/m penalty spring constant

function makeImage(src: string): HTMLImageElement {
  const img = new Image();
  img.src = src;
  return img;
}

const plankImage = makeImage('/pier/plank.svg');

const boatImages: Record<BoatType, HTMLImageElement> = {
  monohull:  makeImage('/boats/monohull.svg'),
  catamaran: makeImage('/boats/catamaran.svg'),
};

// Cleat positions in SVG path-space (same coordinate system as OUTLINE_PATHS paths).
// Points are on the hull outline: stern corners, widest-point midships, and near-bow sides.
// Monohull midship/bow computed from cubic bezier parametric solve (t≈0.295 / t≈0.875).
// Catamaran midship similarly; bow = exact hull tip coordinates from SVG.
const CLEAT_SVG: Record<BoatType, Array<[number, number]>> = {
  monohull: [
    [ 0,       0      ],  // port stern corner
    [11.786,   0      ],  // starboard stern corner
    [-1.085,  18.0    ],  // port midship (hull outline, midway between bow y=36 and stern y=0)
    [12.871,  18.0    ],  // starboard midship
    [ 3.585,  36.0    ],  // port bow (~1 m from tip, t≈0.875)
    [ 8.202,  36.0    ],  // starboard bow
  ],
  catamaran: [
    [ 0,       0      ],  // port stern (left hull outer)
    [23.565,   0      ],  // starboard stern (right hull outer)
    [-0.831,  20.0    ],  // port midship (left hull outline at exact hull midpoint)
    [24.396,  20.0    ],  // starboard midship (right hull outline at exact hull midpoint)
    [ 3.520,  40.0    ],  // left hull bow tip
    [20.045,  40.0    ],  // right hull bow tip
  ],
};

// SVG group-transform metadata + hull outline paths (no centerlines).
// Coordinates are in the SVG group's local space; the canvas transform
// re-applies scale(1,-1) + translate to map them to screen pixels.
const OUTLINE_PATHS: Record<BoatType, {
  svgW: number; svgH: number; tx: number; ty: number; paths: Path2D[];
}> = {
  monohull: {
    svgW: 14.513544331456156,
    svgH: 40.284579300616784,
    tx: 1.363737,
    ty: 40.142290,
    paths: [new Path2D(
      'M 11.786070044516881 0 L 0 0 ' +
      'C -2.7798926079868256 16.291597809404944 -0.8692287120888675 29.70275818571594 5.893035022258441 40 ' +
      'C 12.655298756605749 29.70275818571594 14.565962652503705 16.291597809404944 11.786070044516881 0 Z',
    )],
  },
  catamaran: {
    svgW: 26.384432532508832,
    svgH: 40.51734181436292,
    tx: 1.409670,
    ty: 40.258671,
    paths: [
      new Path2D(
        'M 7.040310526102853 0 L 0 0 ' +
        'C -2.996143072606003 17.083082181569583 0.4224889809398942 30.154096359788877 3.5201552630514263 40 ' +
        'C 6.6178215451629585 30.15409635978887 10.036453598708857 17.083082181569583 7.040310526102853 0 Z',
      ),
      new Path2D(
        'M 16.524781473897146 0 L 23.565092 0 ' +
        'C 26.561235072606003 17.083082181569583 23.142603019060104 30.15409635978887 20.04493673694857 40 ' +
        'C 16.947270454837042 30.15409635978887 13.528638401291143 17.083082181569583 16.524781473897146 0 Z',
      ),
      new Path2D('M 7.345880 1.873169 L 16.219211 1.873169 L 17.236977 29.730728 L 6.328115 29.730728 Z'),
    ],
  },
};

function boatToOBB(boat: Boat): OBB {
  const { w, h } = BOAT_SIZE[boat.type];
  return { x: boat.x, y: boat.y, w, h, heading: boat.heading };
}

function getPierOBB(): OBB {
  const w = canvas.width / SCALE + 40;
  // Extend pier far below the screen so the bottom edge is never the closest exit;
  // the MTV will always push a boat upward through the waterline.
  const pierTopY = -(canvas.height / 2) / SCALE + PIER_DEPTH_M;
  const h = PIER_DEPTH_M + 1000;
  return { x: 0, y: pierTopY - h / 2, w, h, heading: 0 };
}

// SAT OBB test. Returns the minimum translation vector to push `a` out of `b`,
// or null if the boxes do not intersect.
function obbMTV(a: OBB, b: OBB): [number, number] | null {
  const ahw = a.w / 2, ahh = a.h / 2;
  const bhw = b.w / 2, bhh = b.h / 2;

  const ac = Math.cos(a.heading), asin = Math.sin(a.heading);
  const bc = Math.cos(b.heading), bsin = Math.sin(b.heading);
  const dx = b.x - a.x, dy = b.y - a.y;

  // World-space principal axes (bow-stern: (sin h, cos h), beam: (cos h, -sin h))
  // because ctx.rotate maps canvas-local (0,-1) [bow] to world (sin h, cos h) after Y-flip.
  const axes: Array<[number, number]> = [
    [asin,  ac], [ ac, -asin],
    [bsin,  bc], [ bc, -bsin],
  ];

  let minOverlap = Infinity;
  let mtx = 0, mty = 0;

  for (const [nx, ny] of axes) {
    const dn  = dx * nx + dy * ny;
    const eA  = ahh * Math.abs(asin * nx + ac   * ny) + ahw * Math.abs(ac   * nx - asin * ny);
    const eB  = bhh * Math.abs(bsin * nx + bc   * ny) + bhw * Math.abs(bc   * nx - bsin * ny);
    const ov  = eA + eB - Math.abs(dn);
    if (ov <= 0) return null;
    if (ov < minOverlap) {
      minOverlap = ov;
      const sign = dn >= 0 ? -1 : 1;
      mtx = sign * ov * nx;
      mty = sign * ov * ny;
    }
  }

  return [mtx, mty];
}

function anyOverlap(idx: number): boolean {
  const a = boatToOBB(boats[idx]);
  for (let i = 0; i < boats.length; i++) {
    if (i !== idx && obbMTV(a, boatToOBB(boats[i])) !== null) return true;
  }
  return obbMTV(a, getPierOBB()) !== null;
}

// Iteratively push `boats[idx]` out of all obstacles (boats + pier) using MTV.
// Returns the nearest valid {x, y, heading}, falling back to saved state if unresolvable.
function findNearestValid(idx: number): { x: number; y: number; heading: number } {
  let x = boats[idx].x;
  let y = boats[idx].y;
  const { w, h } = BOAT_SIZE[boats[idx].type];
  const heading = boats[idx].heading;

  const obstacles: OBB[] = [
    ...boats.filter((_, i) => i !== idx).map(boatToOBB),
    getPierOBB(),
  ];

  for (let iter = 0; iter < 20; iter++) {
    let moved = false;
    for (const obs of obstacles) {
      const mtv = obbMTV({ x, y, w, h, heading }, obs);
      if (mtv !== null) {
        x += mtv[0];
        y += mtv[1];
        moved = true;
      }
    }
    if (!moved) return { x, y, heading };
  }

  return { x: savedX, y: savedY, heading: savedHeading };
}

function resize(): void {
  const palette = document.getElementById('palette')!;
  canvas.width  = window.innerWidth  - palette.offsetWidth;
  canvas.height = window.innerHeight;
  rebuildStaticCleats();
}

function worldToCanvas(wx: number, wy: number): [number, number] {
  return [
    canvas.width  / 2 + wx * SCALE,
    canvas.height / 2 - wy * SCALE,
  ];
}

function canvasToWorld(cx: number, cy: number): [number, number] {
  return [
    (cx - canvas.width  / 2) / SCALE,
    -(cy - canvas.height / 2) / SCALE,
  ];
}

function drawWater(): void {
  ctx.fillStyle = '#3a7d9c';
  ctx.fillRect(0, 0, canvas.width, canvas.height);

  const gridPx = 10 * SCALE;
  ctx.strokeStyle = 'rgba(255, 255, 255, 0.08)';
  ctx.lineWidth = 1;
  const ox = (canvas.width  / 2) % gridPx;
  const oy = (canvas.height / 2) % gridPx;

  for (let x = ox; x < canvas.width; x += gridPx) {
    ctx.beginPath();
    ctx.moveTo(x, 0);
    ctx.lineTo(x, canvas.height);
    ctx.stroke();
  }
  for (let y = oy; y < canvas.height; y += gridPx) {
    ctx.beginPath();
    ctx.moveTo(0, y);
    ctx.lineTo(canvas.width, y);
    ctx.stroke();
  }
}

function drawPier(): void {
  const depthPx  = PIER_DEPTH_M * SCALE;
  const plankPxW = PLANK_W_M * SCALE;
  const plankPxH = depthPx;
  const y = canvas.height - depthPx;

  if (plankImage.complete && plankImage.naturalWidth !== 0) {
    const count = Math.ceil(canvas.width / plankPxW) + 1;
    for (let i = 0; i < count; i++) {
      ctx.drawImage(plankImage, i * plankPxW, y, plankPxW, plankPxH);
    }
  } else {
    ctx.fillStyle = '#B9A999';
    ctx.fillRect(0, y, canvas.width, depthPx);
  }

  // Waterline beam
  ctx.fillStyle = '#6B5744';
  ctx.fillRect(0, y, canvas.width, 4);
}

function drawCleat(cx: number, cy: number, stroke: string): void {
  ctx.beginPath();
  ctx.arc(cx, cy, CLEAT_R, 0, Math.PI * 2);
  ctx.fillStyle   = CLEAT_FILL;
  ctx.strokeStyle = stroke;
  ctx.lineWidth   = CLEAT_SW;
  ctx.fill();
  ctx.stroke();
}

function rebuildStaticCleats(): void {
  staticCleats = [];
  const plankPxW = PLANK_W_M * SCALE;
  const pierTopY = -(canvas.height / 2) / SCALE + PIER_DEPTH_M;
  const total = Math.ceil(canvas.width / plankPxW) + 1;
  for (let n = 3; n < total; n += 4) {
    const wx = ((n + 0.5) * plankPxW - canvas.width / 2) / SCALE;
    staticCleats.push({ wx, wy: pierTopY - 1.0, kind: 'pier' });
  }
  for (let n = 3; n < total; n += 8) {
    const wx = ((n + 0.5) * plankPxW - canvas.width / 2) / SCALE;
    staticCleats.push({ wx, wy: pierTopY + BUOY_DIST_M, kind: 'buoy' });
  }
}

function getMovingCleatCanvas(boat: Boat, cleat: number): [number, number] {
  const { w, h } = BOAT_SIZE[boat.type];
  const meta = OUTLINE_PATHS[boat.type];
  const scaleX = (w * SCALE) / meta.svgW;
  const scaleY = (h * SCALE) / meta.svgH;
  const [sx, sy] = CLEAT_SVG[boat.type][cleat];
  const lx = (sx + meta.tx) * scaleX - (w * SCALE) / 2;
  const ly = (meta.ty - sy) * scaleY - (h * SCALE) / 2;
  const [bcx, bcy] = worldToCanvas(boat.x, boat.y);
  const cos = Math.cos(boat.heading);
  const sin = Math.sin(boat.heading);
  return [bcx + lx * cos - ly * sin, bcy + lx * sin + ly * cos];
}

function getCleatWorld(ref: CleatRef): [number, number] {
  if (ref.kind === 'static') {
    const sc = staticCleats[ref.idx];
    return [sc.wx, sc.wy];
  }
  const boat = boats[ref.boat];
  const { w, h } = BOAT_SIZE[boat.type];
  const meta = OUTLINE_PATHS[boat.type];
  const scaleX = (w * SCALE) / meta.svgW;
  const scaleY = (h * SCALE) / meta.svgH;
  const [sx, sy] = CLEAT_SVG[boat.type][ref.cleat];
  const lx = (sx + meta.tx) * scaleX - (w * SCALE) / 2;
  const ly = (meta.ty - sy) * scaleY - (h * SCALE) / 2;
  const cos = Math.cos(boat.heading), sin = Math.sin(boat.heading);
  const rlx = lx * cos - ly * sin;
  const rly = lx * sin + ly * cos;
  return [boat.x + rlx / SCALE, boat.y - rly / SCALE];
}

function computeNaturalLength(from: CleatRef, to: CleatRef): number {
  const [ax, ay] = getCleatWorld(from);
  const [bx, by] = getCleatWorld(to);
  return Math.sqrt((bx - ax) ** 2 + (by - ay) ** 2);
}

function getCleatCanvas(ref: CleatRef): [number, number] {
  if (ref.kind === 'static') {
    const sc = staticCleats[ref.idx];
    return worldToCanvas(sc.wx, sc.wy);
  }
  return getMovingCleatCanvas(boats[ref.boat], ref.cleat);
}

function hitTestCleats(mx: number, my: number): CleatRef | null {
  const R = CLEAT_R + 4;
  for (let b = 0; b < boats.length; b++) {
    for (let c = 0; c < CLEAT_SVG[boats[b].type].length; c++) {
      const [cx, cy] = getMovingCleatCanvas(boats[b], c);
      if ((mx - cx) ** 2 + (my - cy) ** 2 <= R * R) return { kind: 'moving', boat: b, cleat: c };
    }
  }
  for (let i = 0; i < staticCleats.length; i++) {
    const [cx, cy] = worldToCanvas(staticCleats[i].wx, staticCleats[i].wy);
    if ((mx - cx) ** 2 + (my - cy) ** 2 <= R * R) return { kind: 'static', idx: i };
  }
  return null;
}

function cleatEq(a: CleatRef, b: CleatRef): boolean {
  if (a.kind !== b.kind) return false;
  if (a.kind === 'static' && b.kind === 'static') return a.idx === b.idx;
  if (a.kind === 'moving' && b.kind === 'moving') return a.boat === b.boat && a.cleat === b.cleat;
  return false;
}

function canConnect(a: CleatRef, b: CleatRef): boolean {
  if (cleatEq(a, b)) return false;
  if (a.kind === 'static' && b.kind === 'static') return false;
  if (a.kind === 'moving' && b.kind === 'moving' && a.boat === b.boat) return false;
  return true;
}

function drawStaticCleats(): void {
  for (let i = 0; i < staticCleats.length; i++) {
    const [cx, cy] = worldToCanvas(staticCleats[i].wx, staticCleats[i].wy);
    drawCleat(cx, cy, STATIC_CLEAT);
  }
}

function drawBoat(boat: Boat, isActive: boolean): void {
  const img = boatImages[boat.type];
  if (!img.complete || img.naturalWidth === 0) return;

  const { w, h } = BOAT_SIZE[boat.type];
  const [cx, cy] = worldToCanvas(boat.x, boat.y);
  const meta   = OUTLINE_PATHS[boat.type];
  const scaleX = (w * SCALE) / meta.svgW;
  const scaleY = (h * SCALE) / meta.svgH;

  ctx.save();
  ctx.translate(cx, cy);
  ctx.rotate(boat.heading);
  ctx.drawImage(img, -(w * SCALE) / 2, -(h * SCALE) / 2, w * SCALE, h * SCALE);

  for (const [sx, sy] of CLEAT_SVG[boat.type]) {
    drawCleat(
      (sx + meta.tx) * scaleX - (w * SCALE) / 2,
      (meta.ty - sy) * scaleY - (h * SCALE) / 2,
      MOVING_CLEAT,
    );
  }

  // Propeller(s) and rudder — positions match physics application points
  const propAftPx   = PROP_AFT_M   * SCALE; // local canvas pixels aft of centre
  const rudderAftPx = RUDDER_AFT_M * SCALE;

  const propPositions: Array<[number, number]> =
    boat.type === 'monohull'
      ? [[0, propAftPx]]
      : [[-2.04 * SCALE, propAftPx], [2.04 * SCALE, propAftPx]];

  for (const [px, py] of propPositions) {
    ctx.beginPath();
    ctx.arc(px, py, 4, 0, Math.PI * 2);
    ctx.fillStyle = '#555555';
    ctx.fill();
    ctx.strokeStyle = '#000000';
    ctx.lineWidth = 1.5;
    ctx.stroke();
  }

  // Rudder blade — centreline, aft of propeller (monohull only)
  if (boat.type === 'monohull') {
    ctx.strokeStyle = '#000000';
    ctx.lineWidth = 3;
    ctx.beginPath();
    ctx.moveTo(0, rudderAftPx);
    ctx.lineTo(
      RUDDER_LEN * Math.sin(boat.rudderAngle),
      rudderAftPx + RUDDER_LEN * Math.cos(boat.rudderAngle),
    );
    ctx.stroke();
  }

  if (isActive) {
    ctx.save();
    ctx.transform(
      scaleX, 0, 0, -scaleY,
      meta.tx * scaleX - (w * SCALE) / 2,
      meta.ty * scaleY - (h * SCALE) / 2,
    );
    ctx.strokeStyle = '#FFD700';
    ctx.lineWidth = 2 / Math.sqrt(scaleX * scaleY);
    for (const p of meta.paths) ctx.stroke(p);
    ctx.restore();

    // Bow and stern rotation handles (setup mode only)
    if (gameMode === 'setup') {
      const hw = (h * SCALE) / 2;
      ctx.strokeStyle = '#FFD700';
      ctx.fillStyle = 'rgba(255, 215, 0, 0.3)';
      ctx.lineWidth = 2;
      for (const hy of [-hw, hw]) {
        ctx.beginPath();
        ctx.arc(0, hy, 8, 0, Math.PI * 2);
        ctx.fill();
        ctx.stroke();
      }
    }
  }

  ctx.restore();

  ctx.beginPath();
  ctx.arc(cx, cy, 4, 0, Math.PI * 2);
  ctx.fillStyle = '#880015';
  ctx.fill();
}

function drawMooringLines(): void {
  ctx.save();
  ctx.lineWidth = 2;

  ctx.strokeStyle = '#4A3728';
  ctx.setLineDash([]);
  for (const line of mooringLines) {
    const [ax, ay] = getCleatCanvas(line.from);
    const [bx, by] = getCleatCanvas(line.to);
    ctx.beginPath();
    ctx.moveTo(ax, ay);
    ctx.lineTo(bx, by);
    ctx.stroke();
  }

  if (pendingCleat !== null) {
    const [ax, ay] = getCleatCanvas(pendingCleat);
    ctx.setLineDash([6, 4]);
    ctx.strokeStyle = 'rgba(74, 55, 40, 0.6)';
    ctx.beginPath();
    ctx.moveTo(ax, ay);
    ctx.lineTo(cursorX, cursorY);
    ctx.stroke();
    ctx.setLineDash([]);

    ctx.beginPath();
    ctx.arc(ax, ay, CLEAT_R + 4, 0, Math.PI * 2);
    ctx.strokeStyle = '#FFFFFF';
    ctx.lineWidth = 2;
    ctx.stroke();
  }

  ctx.restore();
}

function getBinPos(): [number, number] {
  return [canvas.width - 42, 52];
}

function drawBin(): void {
  const [cx, cy] = getBinPos();
  const hot = dragBoatIdx !== null;
  const col = hot ? '#e05555' : '#4a6b8a';
  const bg  = hot ? 'rgba(224,85,85,0.18)' : 'rgba(255,255,255,0.04)';

  ctx.save();
  ctx.strokeStyle = col;
  ctx.fillStyle   = bg;
  ctx.lineWidth   = 2;
  ctx.lineJoin    = 'round';
  ctx.lineCap     = 'round';

  const bw = 22, bh = 26;

  // Body
  ctx.beginPath();
  ctx.roundRect(cx - bw / 2, cy - bh / 2 + 7, bw, bh, 3);
  ctx.fill();
  ctx.stroke();

  // Lid
  ctx.beginPath();
  ctx.roundRect(cx - bw / 2 - 3, cy - bh / 2 + 2, bw + 6, 6, 2);
  ctx.fill();
  ctx.stroke();

  // Handle
  ctx.beginPath();
  ctx.roundRect(cx - 5, cy - bh / 2 - 4, 10, 6, 2);
  ctx.stroke();

  // Slats inside body
  ctx.lineWidth = 1.5;
  for (let i = -1; i <= 1; i++) {
    ctx.beginPath();
    ctx.moveTo(cx + i * 7, cy - bh / 2 + 12);
    ctx.lineTo(cx + i * 7, cy + bh / 2 + 3);
    ctx.stroke();
  }

  ctx.restore();
}

function deleteBoat(idx: number): void {
  // Remove mooring lines connected to this boat
  for (let i = mooringLines.length - 1; i >= 0; i--) {
    const { from, to } = mooringLines[i];
    if ((from.kind === 'moving' && from.boat === idx) ||
        (to.kind  === 'moving' && to.boat   === idx)) {
      mooringLines.splice(i, 1);
    }
  }
  // Shift down boat indices in remaining mooring lines
  for (const line of mooringLines) {
    if (line.from.kind === 'moving' && line.from.boat > idx)
      line.from = { kind: 'moving', boat: line.from.boat - 1, cleat: line.from.cleat };
    if (line.to.kind === 'moving' && line.to.boat > idx)
      line.to   = { kind: 'moving', boat: line.to.boat   - 1, cleat: line.to.cleat   };
  }
  // Fix pending cleat reference
  if (pendingCleat !== null && pendingCleat.kind === 'moving') {
    if (pendingCleat.boat === idx) pendingCleat = null;
    else if (pendingCleat.boat > idx)
      pendingCleat = { kind: 'moving', boat: pendingCleat.boat - 1, cleat: pendingCleat.cleat };
  }
  // Fix active boat index
  if (activeBoatIdx === idx) {
    activeBoatIdx = null;
    syncControlPanels();
  } else if (activeBoatIdx !== null && activeBoatIdx > idx) {
    activeBoatIdx--;
  }
  boats.splice(idx, 1);
}

function drawWindArrow(): void {
  const sin = Math.sin(windAngle);
  const cos = Math.cos(windAngle);
  const tipX  =  WIND_CX + WIND_R * sin;
  const tipY  =  WIND_CY - WIND_R * cos;
  const tailX =  WIND_CX - WIND_R * 0.35 * sin;
  const tailY =  WIND_CY + WIND_R * 0.35 * cos;

  // Background circle
  ctx.beginPath();
  ctx.arc(WIND_CX, WIND_CY, WIND_R + 10, 0, Math.PI * 2);
  ctx.fillStyle = 'rgba(15, 28, 43, 0.75)';
  ctx.fill();
  ctx.strokeStyle = '#2d4a6b';
  ctx.lineWidth = 1;
  ctx.stroke();

  // Shaft
  ctx.beginPath();
  ctx.moveTo(tailX, tailY);
  ctx.lineTo(tipX, tipY);
  ctx.strokeStyle = '#7ab5d4';
  ctx.lineWidth = 2.5;
  ctx.stroke();

  // Arrowhead (filled triangle, tip at arrow tip)
  ctx.save();
  ctx.translate(tipX, tipY);
  ctx.rotate(windAngle);
  ctx.beginPath();
  ctx.moveTo(0, 0);
  ctx.lineTo(-7, 13);
  ctx.lineTo( 7, 13);
  ctx.closePath();
  ctx.fillStyle = '#7ab5d4';
  ctx.fill();
  ctx.restore();

  // Drag handle at tip (setup mode only)
  if (gameMode === 'setup') {
    ctx.beginPath();
    ctx.arc(tipX, tipY, 6, 0, Math.PI * 2);
    ctx.fillStyle = 'rgba(122, 181, 212, 0.25)';
    ctx.strokeStyle = '#7ab5d4';
    ctx.lineWidth = 1.5;
    ctx.fill();
    ctx.stroke();
  }

  // Wind speed label
  ctx.save();
  ctx.font = 'bold 12px system-ui, sans-serif';
  ctx.textAlign = 'center';
  ctx.fillStyle = '#7ab5d4';
  ctx.fillText(`${windKt} kt`, WIND_CX, WIND_CY + WIND_R + 22);
  ctx.restore();
}

function endDrag(): void {
  isDraggingWind = false;
  const idx = dragBoatIdx;

  // Recycle bin drop — must clear dragBoatIdx first so drawBin stops highlighting
  if (idx !== null) {
    const [bx, by] = getBinPos();
    if ((cursorX - bx) ** 2 + (cursorY - by) ** 2 <= 36 * 36) {
      dragBoatIdx = null;
      isRotating  = false;
      deleteBoat(idx);
      return;
    }
  }

  if (idx !== null && anyOverlap(idx)) {
    const target = findNearestValid(idx);
    revertAnim = {
      idx,
      fromX: boats[idx].x, fromY: boats[idx].y, fromHeading: boats[idx].heading,
      toX: target.x, toY: target.y, toHeading: target.heading,
      startTime: performance.now(),
      duration: 450,
    };
  }
  dragBoatIdx = null;
  isRotating  = false;
}

function physicsStep(dt: number): void {
  if (activeBoatIdx === null) return;
  const boat = boats[activeBoatIdx];
  let vx    = boat.vx    ?? 0;
  let vy    = boat.vy    ?? 0;
  let omega = boat.omega ?? 0;

  const mass = BOAT_MASS[boat.type];
  const I    = BOAT_I[boat.type];

  // Bow and starboard unit vectors (matches OBB SAT axis convention)
  const bowX  = Math.sin(boat.heading), bowY  = Math.cos(boat.heading);
  const stbdX = Math.cos(boat.heading), stbdY = -Math.sin(boat.heading);

  let fx = 0, fy = 0, torque = 0;

  // --- Hydrodynamic drag ---
  const vFwd = vx * bowX  + vy * bowY;
  const vLat = vx * stbdX + vy * stbdY;
  fx -= C_DRAG_FWD * vFwd * bowX  + C_DRAG_LAT * vLat * stbdX;
  fy -= C_DRAG_FWD * vFwd * bowY  + C_DRAG_LAT * vLat * stbdY;
  torque -= C_DRAG_ROT * omega;

  // --- Engine thrust + propeller walk ---
  // Propeller position: PROP_AFT_M aft of centre along bow axis
  const propWx = -PROP_AFT_M * bowX;
  const propWy = -PROP_AFT_M * bowY;

  if (boat.type === 'monohull') {
    const thrustN = THROTTLE_FORCE[boat.throttlePort];
    fx += thrustN * bowX;
    fy += thrustN * bowY;
    // Propeller walk — lateral force at prop position
    const pwN = PROP_WALK_TABLE[boat.throttlePort];
    fx += pwN * stbdX;
    fy += pwN * stbdY;
    torque += propWy * (pwN * stbdX) - propWx * (pwN * stbdY);
  } else {
    // Catamaran — lateral arm between centreline and each hull engine (m)
    const lateralArm = 2.04;
    const portPropWx = propWx - lateralArm * stbdX;
    const portPropWy = propWy - lateralArm * stbdY;
    const stbdPropWx = propWx + lateralArm * stbdX;
    const stbdPropWy = propWy + lateralArm * stbdY;

    const thrustP = THROTTLE_FORCE[boat.throttlePort]  / 2;
    const thrustS = THROTTLE_FORCE[boat.throttleStbd] / 2;
    fx += (thrustP + thrustS) * bowX;
    fy += (thrustP + thrustS) * bowY;
    // Differential thrust torque
    torque += portPropWy * (thrustP * bowX) - portPropWx * (thrustP * bowY);
    torque += stbdPropWy * (thrustS * bowX) - stbdPropWx * (thrustS * bowY);

    // Contra-rotating propwalk: port right-handed (+table), stbd left-handed (-table)
    const pwP =  PROP_WALK_TABLE[boat.throttlePort];
    const pwS = -PROP_WALK_TABLE[boat.throttleStbd];
    fx += (pwP + pwS) * stbdX;
    fy += (pwP + pwS) * stbdY;
    torque += portPropWy * (pwP * stbdX) - portPropWx * (pwP * stbdY);
    torque += stbdPropWy * (pwS * stbdX) - stbdPropWx * (pwS * stbdY);
  }

  // --- Rudder force ---
  if (Math.abs(vFwd) >= 0.05) {
    const rudderWx = -RUDDER_AFT_M * bowX;
    const rudderWy = -RUDDER_AFT_M * bowY;
    const rudderF = C_RUDDER * boat.rudderAngle * vFwd;
    fx -= rudderF * stbdX;
    fy -= rudderF * stbdY;
    torque += rudderWx * (rudderF * stbdY) - rudderWy * (rudderF * stbdX);
  }

  // --- Wind force (quadratic) ---
  const wsMs = windKt * 0.51444;
  // Wind blow-toward vector: windAngle 0 = blowing toward north
  const windVx = -wsMs * Math.sin(windAngle);
  const windVy = -wsMs * Math.cos(windAngle);
  const appFwd = windVx * bowX  + windVy * bowY;
  const appLat = windVx * stbdX + windVy * stbdY;
  const wFwdF  = WIND_K_FWD * appFwd * Math.abs(appFwd);
  const wLatF  = WIND_K_LAT * appLat * Math.abs(appLat);
  fx += wFwdF * bowX + wLatF * stbdX;
  fy += wFwdF * bowY + wLatF * stbdY;

  // --- Mooring line spring forces ---
  for (const line of mooringLines) {
    const fromIsActive = line.from.kind === 'moving' && line.from.boat === activeBoatIdx;
    const toIsActive   = line.to.kind   === 'moving' && line.to.boat   === activeBoatIdx;
    if (!fromIsActive && !toIsActive) continue;

    const movingEnd = fromIsActive ? line.from : line.to;
    const otherEnd  = fromIsActive ? line.to   : line.from;
    const [ax, ay] = getCleatWorld(movingEnd);
    const [bx, by] = getCleatWorld(otherEnd);
    const dist = Math.sqrt((bx - ax) ** 2 + (by - ay) ** 2);
    if (dist < 1e-6) continue;
    const ext = dist - line.naturalLength;
    if (ext <= 0) continue;

    const nx = (bx - ax) / dist, ny = (by - ay) / dist;
    const dext = vx * nx + vy * ny;
    const fMag = Math.max(0, MOORING_K * ext - MOORING_C * dext);
    fx += fMag * nx;
    fy += fMag * ny;
    // Torque from line force applied at cleat position relative to boat centre
    const rx = ax - boat.x, ry = ay - boat.y;
    torque += ry * (fMag * nx) - rx * (fMag * ny);
  }

  // --- Collision penalty forces ---
  const activeOBB = boatToOBB(boat);
  const obstacles: OBB[] = [
    ...boats.filter((_, i) => i !== activeBoatIdx).map(boatToOBB),
    getPierOBB(),
  ];
  for (const obs of obstacles) {
    const mtv = obbMTV(activeOBB, obs);
    if (mtv === null) continue;
    fx += COLLISION_K * mtv[0];
    fy += COLLISION_K * mtv[1];
    // 50% position correction to prevent tunnelling
    boat.x += mtv[0] * 0.5;
    boat.y += mtv[1] * 0.5;
  }

  // Semi-implicit Euler integration (velocity before position)
  vx    += (fx / mass) * dt;
  vy    += (fy / mass) * dt;
  omega += (torque / I) * dt;
  boat.vx = vx;
  boat.vy = vy;
  boat.omega = omega;
  boat.x += vx * dt;
  boat.y += vy * dt;
  boat.heading += omega * dt;
}

function render(now: number): void {
  const dt = Math.min((now - lastTime) / 1000, 0.1);
  lastTime = now;
  if (gameMode === 'play' && activeBoatIdx !== null) physicsStep(dt);

  // Advance revert animation (only when the boat is not being dragged)
  if (revertAnim !== null && revertAnim.idx !== dragBoatIdx) {
    const t    = Math.min((performance.now() - revertAnim.startTime) / revertAnim.duration, 1);
    const ease = 1 - (1 - t) ** 3; // cubic ease-out
    const ra   = revertAnim;
    boats[ra.idx].x = ra.fromX + (ra.toX - ra.fromX) * ease;
    boats[ra.idx].y = ra.fromY + (ra.toY - ra.fromY) * ease;
    let da = ra.toHeading - ra.fromHeading;
    while (da >  Math.PI) da -= 2 * Math.PI;
    while (da < -Math.PI) da += 2 * Math.PI;
    boats[ra.idx].heading = ra.fromHeading + da * ease;
    if (t >= 1) revertAnim = null;
  }

  drawWater();
  drawPier();
  drawStaticCleats();
  boats.forEach((boat, i) => drawBoat(boat, i === activeBoatIdx));
  drawMooringLines();
  drawBin();
  drawWindArrow();
  requestAnimationFrame(render);
}

let dragType: BoatType | null = null;

document.querySelectorAll<HTMLElement>('.boat-icon').forEach(el => {
  el.addEventListener('dragstart', () => {
    if (gameMode === 'play') return;
    dragType = el.dataset['type'] as BoatType;
  });
});

canvas.addEventListener('mousedown', e => {
  if (e.button !== 0) return;
  const rect = canvas.getBoundingClientRect();
  const mx = e.clientX - rect.left;
  const my = e.clientY - rect.top;
  const [wx, wy] = canvasToWorld(mx, my);

  // Wind arrow tip drag (setup only)
  if (gameMode === 'setup') {
    const wtx = WIND_CX + WIND_R * Math.sin(windAngle);
    const wty = WIND_CY - WIND_R * Math.cos(windAngle);
    if ((mx - wtx) ** 2 + (my - wty) ** 2 <= 12 * 12) {
      isDraggingWind = true;
      e.preventDefault();
      return;
    }
  }

  // Cleat interaction: takes priority over boat selection
  const hitCleat = hitTestCleats(mx, my);
  if (hitCleat !== null) {
    if (pendingCleat === null) {
      pendingCleat = hitCleat;
    } else if (cleatEq(pendingCleat, hitCleat)) {
      pendingCleat = null;
    } else if (canConnect(pendingCleat, hitCleat)) {
      const isDup = mooringLines.some(
        l => (cleatEq(l.from, pendingCleat!) && cleatEq(l.to, hitCleat)) ||
             (cleatEq(l.from, hitCleat) && cleatEq(l.to, pendingCleat!)),
      );
      if (!isDup) mooringLines.push({ from: pendingCleat, to: hitCleat, naturalLength: computeNaturalLength(pendingCleat, hitCleat) });
      pendingCleat = null;
    } else {
      pendingCleat = null;
    }
    e.preventDefault();
    return;
  }

  // Clicking away from any cleat cancels a pending mooring line
  if (pendingCleat !== null) {
    pendingCleat = null;
    e.preventDefault();
    return;
  }

  // Everything below is setup-only
  if (gameMode === 'play') return;

  // Check bow/stern handles of the active boat first → rotation
  if (activeBoatIdx !== null) {
    const boat = boats[activeBoatIdx];
    const { h } = BOAT_SIZE[boat.type];
    const [cx, cy] = worldToCanvas(boat.x, boat.y);
    const hw = (h * SCALE) / 2;
    const sin = Math.sin(boat.heading);
    const cos = Math.cos(boat.heading);
    const bowCx   = cx + hw * sin;
    const bowCy   = cy - hw * cos;
    const sternCx = cx - hw * sin;
    const sternCy = cy + hw * cos;
    const R = 14;
    if (
      (mx - bowCx)   ** 2 + (my - bowCy)   ** 2 <= R * R ||
      (mx - sternCx) ** 2 + (my - sternCy) ** 2 <= R * R
    ) {
      savedX = boat.x; savedY = boat.y; savedHeading = boat.heading;
      if (revertAnim?.idx === activeBoatIdx) revertAnim = null;
      dragBoatIdx = activeBoatIdx;
      isRotating  = true;
      e.preventDefault();
      return;
    }
  }

  // Check boat bodies → select + move
  dragBoatIdx  = null;
  isRotating   = false;
  activeBoatIdx = null;
  for (let i = boats.length - 1; i >= 0; i--) {
    const { x, y, type } = boats[i];
    const { w, h } = BOAT_SIZE[type];
    if ((wx - x) ** 2 + (wy - y) ** 2 <= (Math.max(w, h) / 2) ** 2) {
      savedX = x; savedY = y; savedHeading = boats[i].heading;
      if (revertAnim?.idx === i) revertAnim = null;
      activeBoatIdx = i;
      dragBoatIdx   = i;
      dragOffsetX   = wx - x;
      dragOffsetY   = wy - y;
      syncControlPanels();
      e.preventDefault();
      return;
    }
  }
  syncControlPanels();
});

canvas.addEventListener('mousemove', e => {
  const rect = canvas.getBoundingClientRect();
  cursorX = e.clientX - rect.left;
  cursorY = e.clientY - rect.top;
  if (isDraggingWind) {
    windAngle = Math.atan2(cursorX - WIND_CX, -(cursorY - WIND_CY));
    return;
  }
  const idx = dragBoatIdx;
  if (idx === null) return;
  if (isRotating) {
    const mx = e.clientX - rect.left;
    const my = e.clientY - rect.top;
    const [cx, cy] = worldToCanvas(boats[idx].x, boats[idx].y);
    boats[idx].heading = Math.atan2(mx - cx, -(my - cy));
  } else {
    const [wx, wy] = canvasToWorld(e.clientX - rect.left, e.clientY - rect.top);
    boats[idx].x = wx - dragOffsetX;
    boats[idx].y = wy - dragOffsetY;
  }
});

canvas.addEventListener('mouseup',    endDrag);
canvas.addEventListener('mouseleave', endDrag);

canvas.addEventListener('dblclick', e => {
  const rect = canvas.getBoundingClientRect();
  const mx = e.clientX - rect.left;
  const my = e.clientY - rect.top;
  const THRESH = 6;
  for (let i = mooringLines.length - 1; i >= 0; i--) {
    const [ax, ay] = getCleatCanvas(mooringLines[i].from);
    const [bx, by] = getCleatCanvas(mooringLines[i].to);
    const dx = bx - ax, dy = by - ay;
    const lenSq = dx * dx + dy * dy;
    const t = lenSq > 0 ? Math.max(0, Math.min(1, ((mx - ax) * dx + (my - ay) * dy) / lenSq)) : 0;
    const px = ax + t * dx - mx;
    const py = ay + t * dy - my;
    if (px * px + py * py <= THRESH * THRESH) {
      mooringLines.splice(i, 1);
      e.preventDefault();
      return;
    }
  }
});

canvas.addEventListener('dragover', e => e.preventDefault());

canvas.addEventListener('drop', e => {
  e.preventDefault();
  if (!dragType || gameMode === 'play') return;
  const rect = canvas.getBoundingClientRect();
  const [wx, wy] = canvasToWorld(e.clientX - rect.left, e.clientY - rect.top);
  boats.push({ type: dragType, x: wx, y: wy, heading: 0, rudderAngle: 0, throttlePort: 2, throttleStbd: 2 });
  if (anyOverlap(boats.length - 1)) boats.pop();
  dragType = null;
});

const rudderPanel    = document.getElementById('rudder-panel')!   as HTMLElement;
const rudderSlider   = document.getElementById('rudder-slider')!  as HTMLInputElement;
const rudderDeg      = document.getElementById('rudder-deg')!     as HTMLElement;
const throttlePanel = document.getElementById('throttle-panel')! as HTMLElement;
const tSliderP      = document.getElementById('tslider-p')!      as HTMLInputElement;
const tSliderS      = document.getElementById('tslider-s')!      as HTMLInputElement;
const tHdrP         = document.getElementById('t-hdr-p')!        as HTMLElement;
const tColS         = document.getElementById('t-col-s')!        as HTMLElement;

function updateThrottleTrack(slider: HTMLInputElement): void {
  const v = Number(slider.value);
  const pos = (v / 4) * 100;
  const lo = v >= 2 ? 50 : pos;
  const hi = v >= 2 ? pos : 50;
  const t = '#2d4a6b', f = '#7ab5d4';
  slider.style.background =
    `linear-gradient(to right,${t} ${lo}%,${f} ${lo}%,${f} ${hi}%,${t} ${hi}%)`;
}

function syncThrottleUI(): void {
  if (activeBoatIdx === null || gameMode !== 'play') {
    throttlePanel.style.display = 'none';
    return;
  }
  throttlePanel.style.display = '';
  const boat = boats[activeBoatIdx];
  const isCat = boat.type === 'catamaran';
  tHdrP.style.visibility = isCat ? 'visible' : 'hidden';
  tColS.style.display     = isCat ? ''        : 'none';
  tSliderP.value = String(boat.throttlePort);
  tSliderS.value = String(boat.throttleStbd);
  updateThrottleTrack(tSliderP);
  updateThrottleTrack(tSliderS);
}

tSliderP.addEventListener('input', () => {
  if (activeBoatIdx === null) return;
  boats[activeBoatIdx].throttlePort = Number(tSliderP.value);
  updateThrottleTrack(tSliderP);
});

tSliderS.addEventListener('input', () => {
  if (activeBoatIdx === null) return;
  boats[activeBoatIdx].throttleStbd = Number(tSliderS.value);
  updateThrottleTrack(tSliderS);
});

function updateRudderTrack(): void {
  const v = Number(rudderSlider.value);
  const lo = v >= 0 ? 50 : 50 + (v / 70 * 100);
  const hi = v >= 0 ? 50 + (v / 70 * 100) : 50;
  const t = '#2d4a6b', f = '#7ab5d4';
  rudderSlider.style.background =
    `linear-gradient(to right,${t} ${lo}%,${f} ${lo}%,${f} ${hi}%,${t} ${hi}%)`;
}

function syncRudderUI(): void {
  if (activeBoatIdx === null || gameMode !== 'play') {
    rudderPanel.style.display = 'none';
    return;
  }
  rudderPanel.style.display = '';
  const deg = Math.round(boats[activeBoatIdx].rudderAngle * 180 / Math.PI);
  rudderSlider.value    = String(deg);
  rudderDeg.textContent = String(deg);
  updateRudderTrack();
}

rudderSlider.addEventListener('input', () => {
  if (activeBoatIdx === null) return;
  boats[activeBoatIdx].rudderAngle = Number(rudderSlider.value) * Math.PI / 180;
  rudderDeg.textContent = rudderSlider.value;
  updateRudderTrack();
});

rudderPanel.addEventListener('dblclick', () => {
  if (activeBoatIdx === null) return;
  boats[activeBoatIdx].rudderAngle = 0;
  rudderDeg.textContent = '0';
  rudderSlider.value = '0';
  updateRudderTrack();
});

const windKtInput = document.getElementById('wind-kt-input')! as HTMLInputElement;
windKtInput.addEventListener('input', () => {
  const v = parseInt(windKtInput.value, 10);
  if (!isNaN(v)) windKt = Math.max(0, Math.min(99, v));
});

const playBtn = document.getElementById('play-btn')! as HTMLButtonElement;

function syncPlayBtn(): void {
  if (gameMode === 'play') {
    playBtn.textContent = '■';
    playBtn.title       = 'Back to Setup';
    playBtn.className   = 'is-stop';
    playBtn.disabled    = false;
  } else {
    playBtn.textContent = '▶';
    playBtn.title       = activeBoatIdx === null ? 'Select a boat to enable Play mode' : 'Enter Play mode';
    playBtn.className   = 'is-play';
    playBtn.disabled    = activeBoatIdx === null;
  }
  const isPlay = gameMode === 'play';
  document.querySelectorAll<HTMLElement>('.boat-icon').forEach(el => {
    el.classList.toggle('disabled', isPlay);
  });
}

function syncControlPanels(): void {
  syncRudderUI();
  syncThrottleUI();
  syncPlayBtn();
}

playBtn.addEventListener('click', () => {
  gameMode = gameMode === 'setup' ? 'play' : 'setup';
  windKtInput.disabled = gameMode === 'play';
  if (gameMode === 'play' && activeBoatIdx !== null) {
    const b = boats[activeBoatIdx];
    b.vx = 0; b.vy = 0; b.omega = 0;
  }
  syncControlPanels();
});

window.addEventListener('resize', resize);

resize();
requestAnimationFrame(render);
