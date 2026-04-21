import { MagneticGrid } from "./canvas/MagneticGrid";
import { WebGLNoise } from "./canvas/WebGLNoise";

let gridInstances: MagneticGrid[] = [];
let noiseInstance: WebGLNoise | null = null;

function initCanvas() {
  const baseCanvas = document.getElementById("magnetic-canvas-base") as HTMLCanvasElement;
  const waveCanvas = document.getElementById("magnetic-canvas-wave") as HTMLCanvasElement;
  const noiseCanvas = document.getElementById("webgl-noise-canvas") as HTMLCanvasElement;
  
  if (baseCanvas && gridInstances.length === 0) {
    gridInstances.push(new MagneticGrid(baseCanvas, 0.25));
  }
  if (waveCanvas && gridInstances.length === 1) {
    gridInstances.push(new MagneticGrid(waveCanvas, 0.9));
  }
  if (noiseCanvas && !noiseInstance) {
    noiseInstance = new WebGLNoise(noiseCanvas);
  }
}

function initCursor() {
  const cursor = document.getElementById("custom-cursor");
  if (!cursor) return;

  let targetX = -100;
  let targetY = -100;

  window.addEventListener("mousemove", (e) => {
    targetX = e.clientX;
    targetY = e.clientY;
    if (cursor.style.opacity !== "1") {
      cursor.style.opacity = "1";
    }
  }, { passive: true });

  function renderCursor() {
    cursor!.style.transform = `translate(calc(${targetX}px - 50%), calc(${targetY}px - 50%))`;
    requestAnimationFrame(renderCursor);
  }
  requestAnimationFrame(renderCursor);
}

document.addEventListener("DOMContentLoaded", () => {
  initCanvas();
  initCursor();
});

document.addEventListener("htmx:afterSwap", () => {
  gridInstances.forEach(g => g.destroy());
  gridInstances = [];
  if (noiseInstance) {
    noiseInstance.destroy();
    noiseInstance = null;
  }
  initCanvas();
  initCursor();
});
