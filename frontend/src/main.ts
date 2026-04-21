import { MagneticGrid } from "./canvas/MagneticGrid";

let gridInstances: MagneticGrid[] = [];

function initCanvas() {
  const baseCanvas = document.getElementById("magnetic-canvas-base") as HTMLCanvasElement;
  const waveCanvas = document.getElementById("magnetic-canvas-wave") as HTMLCanvasElement;
  
  if (baseCanvas && gridInstances.length === 0) {
    gridInstances.push(new MagneticGrid(baseCanvas, 0.25));
  }
  if (waveCanvas && gridInstances.length === 1) {
    gridInstances.push(new MagneticGrid(waveCanvas, 0.9));
  }
}

function initCursor() {
  const cursor = document.getElementById("custom-cursor");
  if (!cursor) return;

  // Use requestAnimationFrame for lag-free performance matching hardware pointer
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

// Full page load (direct navigation to /home)
document.addEventListener("DOMContentLoaded", () => {
  initCanvas();
  initCursor();
});

// HTMX swap (splash → home transition via outerHTML swap)
document.addEventListener("htmx:afterSwap", () => {
  // Destroy previous instances if any
  gridInstances.forEach(g => g.destroy());
  gridInstances = [];
  initCanvas();
  initCursor();
});
