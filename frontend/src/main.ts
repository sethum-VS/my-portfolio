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

// Full page load (direct navigation to /home)
document.addEventListener("DOMContentLoaded", initCanvas);

// HTMX swap (splash → home transition via outerHTML swap)
document.addEventListener("htmx:afterSwap", () => {
  // Destroy previous instances if any
  gridInstances.forEach(g => g.destroy());
  gridInstances = [];
  initCanvas();
});
