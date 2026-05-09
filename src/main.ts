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

const boats: Boat[] = [];

function makeImage(src: string): HTMLImageElement {
  const img = new Image();
  img.src = src;
  return img;
}

const boatImages: Record<BoatType, HTMLImageElement> = {
  monohull:  makeImage('/boats/monohull.svg'),
  catamaran: makeImage('/boats/catamaran.svg'),
};

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

function drawBoat(boat: Boat): void {
  const img = boatImages[boat.type];
  if (!img.complete || img.naturalWidth === 0) return;

  const { w, h } = BOAT_SIZE[boat.type];
  const [cx, cy] = worldToCanvas(boat.x, boat.y);

  ctx.save();
  ctx.translate(cx, cy);
  ctx.rotate(boat.heading);
  ctx.drawImage(img, -(w * SCALE) / 2, -(h * SCALE) / 2, w * SCALE, h * SCALE);
  ctx.restore();

  ctx.beginPath();
  ctx.arc(cx, cy, 4, 0, Math.PI * 2);
  ctx.fillStyle = '#880015';
  ctx.fill();
}

function render(): void {
  drawWater();
  boats.forEach(drawBoat);
  requestAnimationFrame(render);
}

let dragType: BoatType | null = null;

document.querySelectorAll<HTMLElement>('.boat-icon').forEach(el => {
  el.addEventListener('dragstart', () => {
    dragType = el.dataset['type'] as BoatType;
  });
});

canvas.addEventListener('dragover', e => e.preventDefault());

canvas.addEventListener('drop', e => {
  e.preventDefault();
  if (!dragType) return;
  const rect = canvas.getBoundingClientRect();
  const [wx, wy] = canvasToWorld(e.clientX - rect.left, e.clientY - rect.top);
  boats.push({ type: dragType, x: wx, y: wy, heading: 0 });
  dragType = null;
});

window.addEventListener('resize', resize);

resize();
render();
