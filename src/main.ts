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
  heading: number; // radians, 0 = bow up (north)
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

const boats: Boat[] = [];
let activeBoatIdx: number | null = null;
let dragBoatIdx: number | null = null;
let dragOffsetX = 0;
let dragOffsetY = 0;
let isRotating = false;
let savedX = 0, savedY = 0, savedHeading = 0;
let revertAnim: RevertAnim | null = null;

const PIER_DEPTH_M = 5;   // pier depth in world-meters
const PLANK_W_M    = 0.5; // each plank width in world-meters (≈ 1:10 aspect at 5m deep)

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

  const axes: Array<[number, number]> = [
    [ ac,   asin], [-asin,  ac],
    [ bc,   bsin], [-bsin,  bc],
  ];

  let minOverlap = Infinity;
  let mtx = 0, mty = 0;

  for (const [nx, ny] of axes) {
    const dn  = dx * nx + dy * ny;
    const eA  = ahw * Math.abs(ac * nx + asin * ny) + ahh * Math.abs(-asin * nx + ac * ny);
    const eB  = bhw * Math.abs(bc * nx + bsin * ny) + bhh * Math.abs(-bsin * nx + bc * ny);
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

function drawBoat(boat: Boat, isActive: boolean): void {
  const img = boatImages[boat.type];
  if (!img.complete || img.naturalWidth === 0) return;

  const { w, h } = BOAT_SIZE[boat.type];
  const [cx, cy] = worldToCanvas(boat.x, boat.y);

  ctx.save();
  ctx.translate(cx, cy);
  ctx.rotate(boat.heading);
  ctx.drawImage(img, -(w * SCALE) / 2, -(h * SCALE) / 2, w * SCALE, h * SCALE);

  if (isActive) {
    const meta = OUTLINE_PATHS[boat.type];
    const scaleX = (w * SCALE) / meta.svgW;
    const scaleY = (h * SCALE) / meta.svgH;
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

    // Bow and stern rotation handles
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

  ctx.restore();

  ctx.beginPath();
  ctx.arc(cx, cy, 4, 0, Math.PI * 2);
  ctx.fillStyle = '#880015';
  ctx.fill();
}

function endDrag(): void {
  const idx = dragBoatIdx;
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

function render(): void {
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
  boats.forEach((boat, i) => drawBoat(boat, i === activeBoatIdx));
  requestAnimationFrame(render);
}

let dragType: BoatType | null = null;

document.querySelectorAll<HTMLElement>('.boat-icon').forEach(el => {
  el.addEventListener('dragstart', () => {
    dragType = el.dataset['type'] as BoatType;
  });
});

canvas.addEventListener('mousedown', e => {
  if (e.button !== 0) return;
  const rect = canvas.getBoundingClientRect();
  const mx = e.clientX - rect.left;
  const my = e.clientY - rect.top;
  const [wx, wy] = canvasToWorld(mx, my);

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
      e.preventDefault();
      return;
    }
  }
});

canvas.addEventListener('mousemove', e => {
  const idx = dragBoatIdx;
  if (idx === null) return;
  const rect = canvas.getBoundingClientRect();
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

canvas.addEventListener('dragover', e => e.preventDefault());

canvas.addEventListener('drop', e => {
  e.preventDefault();
  if (!dragType) return;
  const rect = canvas.getBoundingClientRect();
  const [wx, wy] = canvasToWorld(e.clientX - rect.left, e.clientY - rect.top);
  boats.push({ type: dragType, x: wx, y: wy, heading: 0 });
  if (anyOverlap(boats.length - 1)) boats.pop();
  dragType = null;
});

window.addEventListener('resize', resize);

resize();
render();
